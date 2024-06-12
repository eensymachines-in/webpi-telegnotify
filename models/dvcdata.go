package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
)

var (
	/* Notification: generic body of the notification to be sent
	specific notification is an attachment to this generic on
	*/
	Notification = func(name, mac string, dt time.Time, specific DeviceNotifcn) DeviceNotifcn {
		return &anyNotification{
			DeviceName:   name,
			DeviceMac:    mac,
			CurrDate:     dt,
			Notification: specific, // cfgchange, gpio status, vitalstats
		}
	}
	/*
		CfgChange: is the specific notification denoting the change in the device configuration  */
	CfgChange = func(schd *aquacfg.Schedule) DeviceNotifcn {
		return &cfgChangeNotification{
			New: schd,
		}
	}
	/*
		Pinstatus : is the state opf the pins and gets encapsulated in the GPIO status
		ConnName	: name of the connection ex: Aquaponics pump -I
		ConnType	: Type of the connection, analogue, digital
		ConnPin		: number id of the pin
		PinState	: state of the pin high or low */
	PinStatus = func(name string, typ GPIOConnection, pin int, state GPIOPinState) *Pinstat {
		return &Pinstat{
			ConnName: name,
			ConnType: typ,
			ConnPin:  pin,
			PinState: state,
		}
	}
	/*
		GpioStatus: overall gpio status , collates the pin status at any given point in tim
	*/
	// BUG: right in the signature of the interface I have used a private variable call - this needs to change
	// Pinstat is not accessible to outside this package
	GpioStatus = func(pins ...*Pinstat) DeviceNotifcn {
		return &gpioStatus{
			AllPins: pins,
		}
	}
	/*
		VitalStats: constructor function for vital stats of the device
		All the parameters are bash command outputs that go into making the object
		Arguments as string are then converted to appropriate boolean flags

		aqpsrv: echo -n $(systemctl is-active service)
		service active status : "active" ==true, else false

		online: echo -n $(curl -Is https://www.google.com | head -n 1)
		online  == HTTP/2 200 then true else false.

		vmstat: echo -n $(vmstat | awk '{print $12" "$13}' | tail -1)
		gives space separated usage in % for user &system ex: 16 7

		uptime: echo -n $(uptime | awk '{print $3" "$4" "$5}'| rev | cut -c 2- | rev)
		gives the uptime of the cpu

		dt = time.Now()
		local time on the device
	*/
	VitalStats = func(aqpsrv, cfgwatchsrv, online, vmstat, uptime string) DeviceNotifcn {
		service_status := func(status string) bool {
			return status == "active"
		}
		return &vitalStats{
			AquaponeSrv: service_status(aqpsrv),
			CfgwatchSrv: service_status(cfgwatchsrv),
			Online: func(cmdop string) bool {
				return cmdop == "HTTP/2 200"
			}(online),
			FreeCPU: func(vmstatop string) int {
				bits := strings.Split(vmstatop, " ")
				if len(bits) != 2 {
					return -1 // error condition
				}
				usage := 0
				for _, b := range bits {
					val, err := strconv.ParseInt(b, 10, 32)
					if err != nil {
						return -1 //error condition
					}
					usage += int(val)
				}
				return 100 - usage
			}(vmstat),
			CPUUpTime: uptime,
		}
	}
)

/*
DeviceData: positively identifies the device. The service as such is ignostic to which device are we looking at
Plus with any other notification this has to be included. Device data now includes the date as well.
*/
type anyNotification struct {
	DeviceName   string        `json:"device_name"` // name of the device
	DeviceMac    string        `json:"device_mac"`  // mac id of the device
	CurrDate     time.Time     // current date on the device
	Notification DeviceNotifcn `json:"notification"` // specific notification - gpiostatus/cfgchng/vital stats
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
