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

func TestTelegSendMessage(t *testing.T) {
	cl := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s%s/sendMessage", os.Getenv("BOT_BASEURL"), os.Getenv("BOT_TOK"))

	msg, _ := models.Notification("Aquaponics pump-III", "D5-5C-1C-04-81-29", time.Now(), models.CfgChange(&aquacfg.Schedule{
		Config:   2,
		TickAt:   "14:00",
		PulseGap: 100,
		Interval: 500,
	})).ToMessageTxt()
	bm := BotMessage{ChatID: GRP_ID, Txt: msg, ParseMode: "markdown"}
	// byt, _ := json.Marshal(map[string]string{"chat_id": GRP_ID, "text": "This is hi from inside the bot"})
	byt, _ := json.Marshal(bm)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(byt))
	req.Header.Set("Content-Type", "application/json")
	assert.Nil(t, err, "failed to create new request")
	assert.NotNil(t, req, "Unexpected nil request object")

	resp, err := cl.Do(req)
	assert.Nil(t, err, "Unexpected error when sending request")
	assert.NotNil(t, resp, "Unexpected nil response")

	byt, err = io.ReadAll(resp.Body)
	defer resp.Body.Close()
	assert.Nil(t, err, "Unexpected error when reading response body")
	result := map[string]interface{}{}
	err = json.Unmarshal(byt, &result)
	assert.Nil(t, err, "Unexpected error when unmarshalling")
	t.Log(result)
}


