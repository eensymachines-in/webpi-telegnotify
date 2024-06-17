package models

import (
	"strconv"
	"strings"
	"time"

	"github.com/eensymachines-in/patio/aquacfg"
)

var (
	EMOJI_warning, _   = strconv.ParseInt(strings.TrimPrefix("\\U26A0", "\\U"), 16, 32)
	EMOJI_grinface, _  = strconv.ParseInt(strings.TrimPrefix("\\U1F600", "\\U"), 16, 32)
	EMOJI_rofl, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F923", "\\U"), 16, 32)
	EMOJI_redcross, _  = strconv.ParseInt(strings.TrimPrefix("\\U274C", "\\U"), 16, 32)
	EMOJI_redqs, _     = strconv.ParseInt(strings.TrimPrefix("\\U2194 	", "\\U"), 16, 32)
	EMOJI_bikini, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F459", "\\U"), 16, 32)
	EMOJI_greentick, _ = strconv.ParseInt(strings.TrimPrefix("\\U2705", "\\U"), 16, 32)
	EMOJI_clover, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F340", "\\U"), 16, 32)
	EMOJI_meat, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F357", "\\U"), 16, 32)
	EMOJI_robot, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F916", "\\U"), 16, 32)
	EMOJI_copyrt, _    = strconv.ParseInt(strings.TrimPrefix("\\U00A9", "\\U"), 16, 32)
	EMOJI_banana, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F34C", "\\U"), 16, 32)
	EMOJI_garlic, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F9C4", "\\U"), 16, 32)
	EMOJI_email, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F4E7", "\\U"), 16, 32)
	EMOJI_badge, _     = strconv.ParseInt(strings.TrimPrefix("\\U1FAAA", "\\U"), 16, 32)
	EMOJI_sheild, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F6E1", "\\U"), 16, 32)
	EMOJI_recycle, _   = strconv.ParseInt(strings.TrimPrefix("\\U267B", "\\U"), 16, 32)
	EMOJI_wilted, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F940", "\\U"), 16, 32)
	EMOJI_rupee, _     = strconv.ParseInt(strings.TrimPrefix("\\U20B9", "\\U"), 16, 32)
	EMOJI_clock, _     = strconv.ParseInt(strings.TrimPrefix("\\U1F55C", "\\U"), 16, 32)
	EMOJI_free, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F193", "\\U"), 16, 32)
	EMOJI_runner, _    = strconv.ParseInt(strings.TrimPrefix("\\U1F3C3", "\\U"), 16, 32)
	EMOJI_up, _        = strconv.ParseInt(strings.TrimPrefix("\\U1F53C", "\\U"), 16, 32)
	EMOJI_down, _      = strconv.ParseInt(strings.TrimPrefix("\\U1F53D", "\\U"), 16, 32)
)

type GPIOConnection uint8 // sensor or actuator
type GPIOPinState uint8   // since we are dealing with digital pin only 3 possible outcomes - high, low, floating

/*
denotes the connection type to the gpio
motor : actuator
humidity sensor: sensor
For most of the purposes the actuator would need to report if the gpio is high or low
*/
const (
	SENSOR GPIOConnection = iota
	ACTUATOR
)

const (
	DIGIPIN_LOW GPIOPinState = iota
	DIGIPIN_FLOAT
	DIGIPIN_HIGH
)

var (
	/* Notification: generic body of the notification to be sent
	specific notification is an attachment to this generic on
	*/
	Notification = func(name, mac string, dt time.Time) DeviceNotifcn {
		return &anyNotification{
			DeviceName:   name,
			DeviceMac:    mac,
			CurrDate:     dt,
			Notification: nil, // cfgchange, gpio status, vitalstats
		}
	}
	/*
		CfgChange: is the specific notification denoting the change in the device configuration  */
	CfgChange = func(schd *aquacfg.Schedule) TelegNotification {
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
	GpioStatus = func(pins ...*Pinstat) TelegNotification {
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
	VitalStats = func(aqpsrv, cfgwatchsrv, online, vmstat, uptime string) TelegNotification {
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

type BotMessage struct {
	ChatID    string `json:"chat_id"`
	Txt       string `json:"text"`
	ParseMode string `json:"parse_mode"` // markdown or html - message then can be parse accordigly
}

// DeviceNotifcn : Any struct that can beconverted to BotText as a message that can be dispatched to Telegram
type DeviceNotifcn interface {
	SetNotification(internal TelegNotification) DeviceNotifcn // sets the internal specific notification
	SetGrpId(gid string) TelegNotification                    // sets the telegram group Id of the notificationSetGrpId(gid string) TelegNotification // sets the telegram group Id of the notification
	ToBotMessage(mode string) *BotMessage
}

type TelegNotification interface { // interface to deal with telegram related behaviour of notification
	ToMessageTxt() (string, error) // any object to text messages with emojis
}
