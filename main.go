package main

/* An endpoint for all the devices to post their notifications.
1. Change in the configurations
2. Status of GPIO and thus the actuators and sensors connected to it
3. Vital stats of the device - status of the services, temp, cpu usage percentage  */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/eensymachines-in/errx/httperr"
	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/eensymachines-in/webpi-telegnotify/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	/* this is to be populated from the environment*/
	environVars = []string{
		"DEVICEREG_URL",
		"BOT_BASEURL",
		"BOT_TOK",
		"BOT_UNAME",
	}
)

type BotMessage struct {
	ChatID    string `json:"chat_id"`
	Txt       string `json:"text"`
	ParseMode string `json:"parse_mode"` // markdown or html - message then can be parse accordigly
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
		PadLevelText:  true,
	})
	log.SetReportCaller(false)
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel) // default is info

	val := os.Getenv("FLOG")
	if val == "1" {
		f, err := os.Open(os.Getenv("LOGF")) // file for logs
		if err != nil {
			log.SetOutput(os.Stdout) // error in opening log file
			log.Warn("Failed to open log file, log output set to stdout")
		}
		log.SetOutput(f) // log output set to file direction
		log.Infof("log output is set to file: %s", os.Getenv("LOGF"))

	} else {
		log.SetOutput(os.Stdout)
		log.Info("log output to stdout")
	}

	val = os.Getenv("SILENT")
	if val == "1" {
		log.SetLevel(log.ErrorLevel) // for production
	} else {
		log.SetLevel(log.DebugLevel) // for development
	}
	/* Reading from the environment, one or more variables missing this shall panic here */
	/* ----------- */
	missingEnviron := false // denotes that one or more environment vars was mising
	for _, name := range environVars {
		if os.Getenv(name) == "" {
			sync.OnceFunc(func() {
				missingEnviron = true
			})
			log.Errorf("Missing environment variable : %s", name)
		}
	}
	if missingEnviron {
		log.Fatal("One or more environment variables is missing, cannot continue")
	}
}

/*
	FetchDeviceDetails : to know the details of device specifically the telegram group to which notifications are to be sent

makes a simple http call to the devicereg u-service, incase the call fails this shall abort any further calls to handlers
NOTE: for extensions in the future there has to be a fallback group that the notifications should be logged to. Or perhaps we can think of loggging
all errors in a group , notifications on a grop, logs to a group as well.
*/
func FetchDeviceDetails(c *gin.Context) {
	cl := &http.Client{
		Timeout: 3 * time.Second,
	}
	// TODO: somehow we need a well formed url
	url := fmt.Sprintf("%s/%s", os.Getenv("DEVICEREG_URL"), c.Param("devid"))
	log.WithFields(log.Fields{
		"url": url,
	}).Debug("Url ready")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("failed to form new http request, check url and then try again %s", err)
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
			"stack": "FetchDeviceDetails/NewRequest",
		}))
		return
	}
	resp, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("failed request to ge device details %d:%s", resp.StatusCode, err)
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
			"stack": "FetchDeviceDetails/Do",
		}))
		return
	}
	log.WithFields(log.Fields{
		"status_code": resp.StatusCode,
	}).Debug("Fetched the device details..")
	byt, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("error reading response body %s", err)
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "FetchDeviceDetails/io.ReadAll",
		}))
		return
	}
	result := struct {
		GrpID string `json:"telggrpid"`
	}{}
	err = json.Unmarshal(byt, &result)
	if err != nil {
		err = fmt.Errorf("error unmarshalling response body %s", err)
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "FetchDeviceDetails/json.Unmarshal",
		}))
		return
	}
	log.WithFields(log.Fields{
		"grp_id": result.GrpID,
	}).Debug("Group id the notification is posted to")
	c.Set("GRP_ID", result.GrpID)
	c.Next() // downstream handlers to take care of this

}
func HndlDeviceNotifics(c *gin.Context) {
	typOfNotify := c.Query("typ")
	var not models.DeviceNotifcn
	/* Figuring out the type of notificaiton and making the object accordingly*/
	if typOfNotify == "" {
		// incase when the hhandler does not know the query params to determine which type of notification
		httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("not enough query params in the request")), log.WithFields(log.Fields{
			"typ": typOfNotify,
		}))
		return
	} else if typOfNotify == "cfgchange" {
		log.Debug("Notifying a configuration change")
		not = models.Notification("", "", time.Now(), models.CfgChange(&aquacfg.Schedule{})) // onto which the payload would be unmarshalled

	} else if typOfNotify == "gpiostat" {
		log.Debug("Gpio status notification")
		not = models.Notification("", "", time.Now(), models.GpioStatus(&models.Pinstat{})) // onto which the payload would be unmarshalled
	} else if typOfNotify == "vitals" {
		log.Debug("Vitals status notification")
		not = models.Notification("", "", time.Now(), models.VitalStats("", "", "", "", "")) // onto which the payload would be unmarshalled
	}
	/* Reading the request body and that is agnostic of which notification it is */
	byt, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics",
		}))
		return
	}
	err = json.Unmarshal(byt, &not)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics",
		}))
		return
	}
	/* Preparing the notification to be sent across to telegram */
	msg, _ := not.ToMessageTxt()
	log.WithFields(log.Fields{
		"msg_txt": msg,
	}).Debug("Notification message text")
	grpId, _ := c.Get("GRP_ID") // from the previous handler we have the telegram grp id that we need to post the notification to
	bm := BotMessage{ChatID: grpId.(string), Txt: msg, ParseMode: "markdown"}
	byt, _ = json.Marshal(bm)
	url := fmt.Sprintf("%s%s/sendMessage", os.Getenv("BOT_BASEURL"), os.Getenv("BOT_TOK"))
	log.WithFields(log.Fields{
		"telegram_url": url,
	}).Debug("Telegram post url ready..")

	/* Sending the notificaiton  */
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(fmt.Errorf("failed to form new post request %s", err)), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics/typ=cfgchange",
		}))
		return
	}
	req.Header.Set("Content-Type", "application/json")
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(fmt.Errorf("failed to post notification message to telegram server %s", err)), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics/typ=cfgchange",
		}))
		return
	}
	log.Debug("Telegram message posted..")
	c.AbortWithStatusJSON(http.StatusOK, gin.H{})
}
func main() {
	log.Info("Starting webapi devicenotification..")
	defer log.Warn("closing the webapi application")

	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	/* Pinging the server  */
	r.GET("/ping", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusOK, gin.H{
			"msg": "If you can read this message then the teleg notificaiton server is running",
		})
	})
	// the device to telegram group notiifcation is stored on the devicereg database
	// the device registers the same when booting up thhe first time  ..
	notifics := r.Group("/api/devices/:devid/notifications")
	/*
		?typ=cfgchange : if the device would want to notify the change in the configuration
		?typ=gpiostat : if the device wants to report the current state of the GPI
		?typ=vitals : deivce uses this to notify vital stats
	*/
	notifics.POST("", FetchDeviceDetails, HndlDeviceNotifics)

	log.Fatal(r.Run(":8080"))
}
