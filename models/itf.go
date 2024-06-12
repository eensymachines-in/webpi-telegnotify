package models

import (
	"strconv"
	"strings"
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

// DeviceNotifcn : Any struct that can beconverted to BotText as a message that can be dispatched to Telegram
type DeviceNotifcn interface {
	ToMessageTxt() (string, error) // any object to text messages with emojis
}
