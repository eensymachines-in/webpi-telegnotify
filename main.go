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

type DeviceDetails struct {
	GrpID string `json:"telggrpid"`
	Name  string `json:"name"`
	Mac   string `json:"mac"`
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

// FetchDeviceDetails : gets the details of a registered device relevant to notifications.
// device name
// mac id
// telegram grp id where the notification is destined to.
// This will construct a base notification and then send it downstream to handle for specific notification
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
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			err = fmt.Errorf("no device for mac id found %s: %s", c.Param("devid"), err)
			httperr.HttpErrOrOkDispatch(c, httperr.ErrResourceNotFound(err), log.WithFields(log.Fields{
				"stack":  "FetchDeviceDetails/Do",
				"mac_id": c.Param("devid"),
			}))
			return
		} else {
			err = fmt.Errorf("failed request to ge device details %d:%s", resp.StatusCode, err)
			httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(err), log.WithFields(log.Fields{
				"stack": "FetchDeviceDetails/Do",
			}))
			return
		}
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
	// Device details - name, mac, telegram grp that notificaiton is destined to, all are in this
	// such device details are sought from devicereg service on the cloud.
	result := DeviceDetails{}
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
		"name":   result.Name,
		"macid":  result.Mac,
	}).Debug("Group id the notification is posted to")
	// TODO: basic device notification can be made here, since the device details are available
	// But for now we havent go any means of pushing the specific notificaiton attributes besides the constructor
	if result.GrpID == "" { // the device is registered with no telegram group - this is missing context param
		err = fmt.Errorf("device registration is incomplete, has no telegram id")
		httperr.HttpErrOrOkDispatch(c, httperr.ErrResourceNotFound(err), log.WithFields(log.Fields{
			"stack": "FetchDeviceDetails/result.GrpID",
		}))
		return
	}
	not := models.Notification(result.Name, result.Mac, time.Now()).SetGrpId(result.GrpID)
	c.Set("notification", not) //sending the device details.
	c.Next()                   // downstream handlers to take care of this

}
func HndlDeviceNotifics(c *gin.Context) {
	/* base notification == *anyNotification.
	base details of the notification. - except the time and the specific notification that we receive from the device below */
	val, _ := c.Get("notification") // from the previous handler
	not := val.(models.DeviceNotifcn)
	if not == nil { // safety -incase you forgot to attach the handler in the router itself
		httperr.HttpErrOrOkDispatch(c, httperr.ErrResourceNotFound(fmt.Errorf("device for the mac id wasnt found")), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics",
		}))
		return
	}
	/* --------------------  Getting the specific notification
	Please see this is an empty default specific notification object*/
	var specificNot models.TelegNotification // specific notification
	typOfNotify := c.Query("typ")

	if typOfNotify == "" {
		// incase when the hhandler does not know the query params to determine which type of notification
		httperr.HttpErrOrOkDispatch(c, httperr.ErrContxParamMissing(fmt.Errorf("not enough query params in the request")), log.WithFields(log.Fields{
			"typ": typOfNotify,
		}))
		return
	} else if typOfNotify == "cfgchange" {
		log.Debug("Notifying a configuration change")
		specificNot = models.CfgChange(&aquacfg.Schedule{})
	} else if typOfNotify == "gpiostat" {
		log.Debug("Gpio status notification")
		specificNot = models.GpioStatus(&models.Pinstat{})
	} else if typOfNotify == "vitals" {
		log.Debug("Vitals status notification")
		specificNot = models.VitalStats("", "", "", "", "") // onto which the payload would be unmarshalled
	}
	not.SetNotification(specificNot) // notification object is complete

	/* -------------------- Reading the payload, details of the notification from the device, time */
	byt, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics",
		}))
		return
	}
	err = json.Unmarshal(byt, not)
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrBinding(err), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics",
		}))
		return
	}
	/* Convert from notification to BotMessage and prepare to send across to telegram  */
	msg, _ := not.(models.TelegNotification).ToMessageTxt()
	log.WithFields(log.Fields{
		"msg_txt": msg,
	}).Debug("Notification message text")
	bm := not.ToBotMessage("markdown") // *BotMessage
	byt, _ = json.Marshal(bm)
	url := fmt.Sprintf("%s%s/sendMessage", os.Getenv("BOT_BASEURL"), os.Getenv("BOT_TOK"))
	log.WithFields(log.Fields{
		"telegram_url": url,
	}).Debug("Telegram post url ready..")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
	if err != nil {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(fmt.Errorf("failed to form new post request %s", err)), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics/typ=cfgchange",
		}))
		return
	}
	req.Header.Set("Content-Type", "application/json") // never forget this
	cl := &http.Client{Timeout: 5 * time.Second}
	resp, err := cl.Do(req) // Sends the notification
	if err != nil || resp.StatusCode != http.StatusOK {
		httperr.HttpErrOrOkDispatch(c, httperr.ErrGatewayConnect(fmt.Errorf("failed to post notification message to telegram server %s", err)), log.WithFields(log.Fields{
			"stack": "HndlDeviceNotifics/typ=cfgchange",
		}))
		return
	}
	/* Done! we are ready to return 200 ok*/
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
