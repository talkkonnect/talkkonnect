package hd44780

import (
	"github.com/stianeikeland/go-rpio"
)

func initPin(pin int) (p rpio.Pin) {
	p = rpio.Pin(pin)
	rpio.PinMode(p, rpio.Output)
	return
}
