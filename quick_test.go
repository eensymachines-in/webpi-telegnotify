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
	baseurl := "http://localhost:8080/api/devices/b8:27:eb:a5:be:48/notifications"
	t.Run("cfg change", func(t *testing.T) {
		url := fmt.Sprintf("%s/?typ=cfgchange", baseurl)
		not := models.Notification("Test aquaponics configuration", "b8:27:eb:a5:be:48", time.Now(), models.CfgChange(&aquacfg.Schedule{
			Config:   1,
			TickAt:   "11:30",
			PulseGap: 100,
			Interval: 500,
		}))

		byt, err := json.Marshal(not)
		assert.Nil(t, err, "Unexpected error when marshaling bot message")
		payload := bytes.NewBuffer(byt)
		req, err := http.NewRequest("POST", url, payload)
		assert.Nil(t, err, "Unexpected error when forming the request")
		resp, err := cl.Do(req)
		assert.Nil(t, err, "unexpected error when executing the request, do you have access to the server ?")
		assert.Equal(t, resp.Status, 200, "Unepxected response code from server")
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

// func TestTelegSendMessage(t *testing.T) {
// 	cl := &http.Client{Timeout: 5 * time.Second}
// 	url := fmt.Sprintf("%s%s/sendMessage", os.Getenv("BOT_BASEURL"), os.Getenv("BOT_TOK"))

// 	msg, _ := models.Notification("Aquaponics pump-III", "D5-5C-1C-04-81-29", time.Now(), models.CfgChange(&aquacfg.Schedule{
// 		Config:   2,
// 		TickAt:   "14:00",
// 		PulseGap: 100,
// 		Interval: 500,
// 	})).ToMessageTxt()
// 	bm := BotMessage{ChatID: GRP_ID, Txt: msg, ParseMode: "markdown"}
// 	// byt, _ := json.Marshal(map[string]string{"chat_id": GRP_ID, "text": "This is hi from inside the bot"})
// 	byt, _ := json.Marshal(bm)

// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
// 	req.Header.Set("Content-Type", "application/json")
// 	assert.Nil(t, err, "failed to create new request")
// 	assert.NotNil(t, req, "Unexpected nil request object")

// 	resp, err := cl.Do(req)
// 	assert.Nil(t, err, "Unexpected error when sending request")
// 	assert.NotNil(t, resp, "Unexpected nil response")

// 	byt, err = io.ReadAll(resp.Body)
// 	defer resp.Body.Close()
// 	assert.Nil(t, err, "Unexpected error when reading response body")
// 	result := map[string]interface{}{}
// 	err = json.Unmarshal(byt, &result)
// 	assert.Nil(t, err, "Unexpected error when unmarshalling")
// 	t.Log(result)
// }
