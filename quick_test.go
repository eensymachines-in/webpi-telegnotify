package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
	"github.com/eensymachines-in/webpi-telegnotify/models"
	"github.com/stretchr/testify/assert"
)

func TestApi(t *testing.T) {
	cl := &http.Client{Timeout: 5 * time.Second}
	/* NOTE: we test on github actions and when running inside the container loopback url is the same as http:localost */
	baseurl := "http://localhost:8080/api/devices/b8:27:eb:a5:be:48/notifications"
	type payload struct {
		Dttm         time.Time                `json:"dttm"`
		Notification models.TelegNotification `json:"notification"`
	}
	t.Run("invalid_query_param", func(t *testing.T) {
		url := fmt.Sprintf("%s/?typ=invalidparam", baseurl)
		pl := payload{
			Dttm: time.Now(),
			Notification: models.CfgChange(&aquacfg.Schedule{
				Config:   1,
				TickAt:   "11:30",
				PulseGap: 100,
				Interval: 500,
			}),
		}
		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Unepxected response code from server")
	})

	t.Run("missing_device_reg", func(t *testing.T) {
		// Whhen the payload itself is nil
		url := fmt.Sprintf("%s/?typ=cfgchange", "http://localhost:8080/api/devices/b8:27:eb:a5:be:48/notifications")
		req, err := http.NewRequest("POST", url, nil)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusBadRequest, http.StatusBadRequest, resp.StatusCode, "Unepxected response code from server")
	})

	t.Run("missing_mac", func(t *testing.T) {
		// case of missing mac id
		url := fmt.Sprintf("%s/?typ=cfgchange", "http://localhost:8080/api/devices/b1:34:fg:a4:vg:50/notifications")
		pl := payload{
			Dttm:         time.Now(),
			Notification: nil,
		}
		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Unepxected response code from server")
	})
	t.Run("invalid_payload", func(t *testing.T) {
		// case of invalid payload
		url := fmt.Sprintf("%s/?typ=cfgchange", baseurl)
		pl := payload{
			Dttm:         time.Now(),
			Notification: nil,
		}
		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Unepxected response code from server")
	})
	// c5:8f:65:59:cb:fe - see this test data in the database, the telegrpid is empty
	t.Run("no_teleg_grp", func(t *testing.T) {
		// this is thhe case when the mac though is found the device registration is not complete
		// telegram group id is not found - this will have to be 404
		url := fmt.Sprintf("%s/?typ=cfgchange", "http://localhost:8080/api/devices/c5:8f:65:59:cb:fe/notifications")
		pl := payload{
			Dttm: time.Now(),
			Notification: models.CfgChange(&aquacfg.Schedule{
				Config:   1,
				TickAt:   "11:30",
				PulseGap: 100,
				Interval: 500,
			}),
		}
		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Unepxected response code from server")
	})

	// NOTE: there isnt a possibility of a bad configuration - since the update is tightly guarded by cfgwatch
	t.Run("cfg_change", func(t *testing.T) {
		url := fmt.Sprintf("%s/?typ=cfgchange", baseurl)
		pl := payload{
			Dttm: time.Now(),
			Notification: models.CfgChange(&aquacfg.Schedule{
				Config:   1,
				TickAt:   "11:30",
				PulseGap: 100,
				Interval: 500,
			}),
		}

		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Unepxected response code from server")
	})

	t.Run("GPIO_report_status", func(t *testing.T) {
		url := fmt.Sprintf("%s/?typ=gpiostat", baseurl)
		pl := payload{
			Dttm:         time.Now(),
			Notification: models.GpioStatus(models.PinStatus("Motor control relay", models.ACTUATOR, 33, models.DIGIPIN_HIGH)),
		}
		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Unepxected response code from server")
	})

	t.Run("vital_status", func(t *testing.T) {
		url := fmt.Sprintf("%s/?typ=vitals", baseurl)
		pl := payload{
			Dttm:         time.Now(),
			Notification: models.VitalStats("active", "active", "HTTP/2 200", "16 7", "4d 8h"),
		}

		byt, err := json.Marshal(pl)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		buff := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, buff)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, http.StatusOK, resp.StatusCode, "Unepxected response code from server")
	})
}

func TestTelegGetMe(t *testing.T) {
	cl := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s%s/getMe", os.Getenv("BOT_BASEURL"), os.Getenv("BOT_TOK"))
	req, err := http.NewRequest("GET", url, nil)
	assert.Nil(t, err, "failed to create new request")
	assert.NotNil(t, req, "Unexpected nil request object")
	resp, err := cl.Do(req)
	assert.Nil(t, err, "Unexpected error when sending request")
	assert.NotNil(t, resp, "Unexpected nil response")
	byt, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(t, err, "Unexpected error when reading response body")
	result := map[string]interface{}{}
	err = json.Unmarshal(byt, &result)
	assert.Nil(t, err, "Unexpected error when unmarshalling")
	t.Log(result)
}
