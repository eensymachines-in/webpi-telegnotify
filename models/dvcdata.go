package models

import (
	"fmt"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
)

/*
DeviceData: positively identifies the device. The service as such is ignostic to which device are we looking at
Plus with any other notification this has to be included. Device data now includes the date as well.
*/
type anyNotification struct {
	DeviceName   string            // name of the device as it appears in the registration
	DeviceMac    string            // mac id of the device,
	TelegGrpID   string            // telegram grp id - destination of the notification
	CurrDate     time.Time         `json:"dttm"` // time on the device the notification was generated
	Notification TelegNotification `json:"notification"`
}

func (dd *anyNotification) ToMessageTxt() (string, error) {
	notifcn, err := dd.Notification.ToMessageTxt()
	if err != nil {
		result := fmt.Sprintf("*%s*\n_%s_\n%s\n----\n There was an error reading the device notification", dd.DeviceName, dd.DeviceMac, dd.CurrDate.Local().Format(time.RFC822))
		return result, nil
	}
	result := fmt.Sprintf("*%s*\n_%s_\n%s\n----\n%s", dd.DeviceName, dd.DeviceMac, dd.CurrDate.Local().Format(time.RFC822), notifcn)
	return result, nil
}
func (dd *anyNotification) SetGrpId(gid string) TelegNotification {
	dd.TelegGrpID = gid
	return dd
}
func (dd *anyNotification) SetNotification(internal TelegNotification) DeviceNotifcn {
	dd.Notification = internal
	return dd
}

func (dd *anyNotification) ToBotMessage(mode string) *BotMessage {
	if dd.Notification == nil { // incase the specific notification is nil there is no point in calling toMessageTxt()
		return nil
	}
	msg, _ := dd.ToMessageTxt()
	return &BotMessage{ChatID: dd.TelegGrpID, Txt: msg, ParseMode: mode}
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++ */

type cfgChangeNotification struct {
	New *aquacfg.Schedule `json:"new"` // new schedule just applied
}

func (ccn *cfgChangeNotification) ToMessageTxt() (string, error) {
	result := "New configuration applied.."
	if ccn.New != nil {
		result = fmt.Sprintf("%s\nConfig: %d\nTickAt: %s\nInterval: %d\nPulsegap: %d", result, ccn.New.Config, ccn.New.TickAt, ccn.New.Interval, ccn.New.PulseGap)
	} else {
		result = fmt.Sprintf("%s\nThere was an issue getting the new configuration", result)
	}
	return result, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++ */
/* PinStatus: contains a single pin status - slice of pin status go into gpio status */
type Pinstat struct {
	ConnName string         `json:"conn_name"` // name of the connection ex: Pump-I, Pump-II, Lights
	ConnType GPIOConnection `json:"conn_type"` // Actuattion /Sensory node
	ConnPin  int            `json:"conn_pin"`  // numeral identification of the pin
	PinState GPIOPinState   `json:"pin_state"` // Pin state - 0,1,float
}

type gpioStatus struct {
	AllPins []*Pinstat `json:"all_pins"` // since there are multiple pins reported in a notification
}

/* With the device details on the top this can print status of each pin name and sattus if high or low */
func (gps *gpioStatus) ToMessageTxt() (string, error) {
	result := ""
	for _, p := range gps.AllPins {
		state := EMOJI_redcross
		if p.PinState == DIGIPIN_HIGH {
			state = EMOJI_up
		} else if p.PinState == DIGIPIN_LOW {
			state = EMOJI_down
		}
		result = fmt.Sprintf("%s%s:\t\t%c\n", result, p.ConnName, state)
	}
	return result, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++ */

/* VitalStatsData: is the object that eventually gets converted to text message in bot send */
type vitalStats struct {
	AquaponeSrv bool   `json:"aquapone_service"` // indicates if the systemctl unit is working
	CfgwatchSrv bool   `json:"cfgwatch_service"` // indicates if the systemctl unit is working
	Online      bool   `json:"online"`           // indicates if the device is online, on internet
	FreeCPU     int    `json:"free_cpu"`         // indicates percentage of CPU that is free
	CPUUpTime   string `json:"cpu_uptime"`       // indicates the cpu up time from uptime command
}

func (vs *vitalStats) ToMessageTxt() (string, error) {
	// default result if everything else fails
	result := ""
	if vs.AquaponeSrv {
		result = fmt.Sprintf("%s\n%c\tAquapone.service", result, EMOJI_runner)
	} else {
		result = fmt.Sprintf("%s\n%c\tAquapone.service", result, EMOJI_redcross)
	}
	if vs.CfgwatchSrv {
		result = fmt.Sprintf("%s\n%c\tCfgwatch.service", result, EMOJI_runner)
	} else {
		result = fmt.Sprintf("%s\n%c\tCfgwatch.service", result, EMOJI_redcross)
	}
	if vs.Online {
		result = fmt.Sprintf("%s\n%c\tDevice online", result, EMOJI_greentick)
	} else {
		result = fmt.Sprintf("%s\n%c\tDevice offline", result, EMOJI_redcross)
	}
	if vs.FreeCPU > -1 {
		result = fmt.Sprintf("%s\n%c\tCPU free: %d%%", result, EMOJI_free, vs.FreeCPU)
	} else {
		result = fmt.Sprintf("%s\n%c\tCPU free: %c", result, EMOJI_free, EMOJI_redqs)
	}
	result = fmt.Sprintf("%s\n%c\tCPU up since: %s", result, EMOJI_clock, vs.CPUUpTime)
	return result, nil
}

/* ++++++++++++++++++++++++++++++++++++++++++++++++ */
