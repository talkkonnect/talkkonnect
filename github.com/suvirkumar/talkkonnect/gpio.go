package talkkonnect

import (
	"github.com/fatih/color"
	"github.com/stianeikeland/go-rpio"
	"github.com/suvirkumar/gpio"
	"time"
)

func (b *Talkkonnect) initGPIO() {
	if err := rpio.Open(); err != nil {
		color.Red(time.Now().Format(time.Stamp) + " Error    : GPIO Error, %v\n", err)
		b.GPIOEnabled = false
		return
	} else {
		b.GPIOEnabled = true
	}

	TxButtonPinPullUp := rpio.Pin(TxButtonPin)
	TxButtonPinPullUp.PullUp()

	UpButtonPinPullUp := rpio.Pin(UpButtonPin)
	UpButtonPinPullUp.PullUp()

	DownButtonPinPullUp := rpio.Pin(DownButtonPin)
	DownButtonPinPullUp.PullUp()

	rpio.Close()

	b.TxButton = gpio.NewInput(TxButtonPin)
	go func() {
		for {
			currentState, err := b.TxButton.Read()

			if currentState != b.TxButtonState && err == nil {
				b.TxButtonState = currentState

				if b.Stream != nil {
					if b.TxButtonState == 1 {
						color.Yellow(time.Now().Format(time.Stamp) + " Event   : TX Button is released\n")
						b.TransmitStop()
					} else {
						color.Yellow(time.Now().Format(time.Stamp) + " Event   : TX Button is keyed\n")
						b.TransmitStart()
					}
				}

			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	b.UpButton = gpio.NewInput(UpButtonPin)
	go func() {
		for {
			currentState, err := b.UpButton.Read()

			if currentState != b.UpButtonState && err == nil {
				b.UpButtonState = currentState

				if b.UpButtonState == 1 {
					color.Yellow(time.Now().Format(time.Stamp) + " Event   : UP Button is released\n")
				} else {
					color.Yellow(time.Now().Format(time.Stamp) + " Event   : Up Button is pressed\n")
					b.ChannelUp()
				}

			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	b.DownButton = gpio.NewInput(DownButtonPin)
	go func() {
		for {
			currentState, err := b.DownButton.Read()

			if currentState != b.DownButtonState && err == nil {
				b.DownButtonState = currentState

				if b.DownButtonState == 1 {
					color.Yellow(time.Now().Format(time.Stamp) + " Event   : Down Button is released\n")
				} else {
					color.Yellow(time.Now().Format(time.Stamp) + " Event   : Down Button is pressed\n")
					b.ChannelDown()
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	// then we can do our gpio stuff
	b.OnlineLED = gpio.NewOutput(OnlineLEDPin, false)
	b.ParticipantsLED = gpio.NewOutput(ParticipantsLEDPin, false)
	b.TransmitLED = gpio.NewOutput(TransmitLEDPin, false)
}

func (b *Talkkonnect) LEDOn(LED gpio.Pin) {
	if b.GPIOEnabled == false {
		return
	}

	LED.High()
}

func (b *Talkkonnect) LEDOff(LED gpio.Pin) {
	if b.GPIOEnabled == false {
		return
	}

	LED.Low()
}

func (b *Talkkonnect) LEDOffAll() {
	if b.GPIOEnabled == false {
		return
	}

	b.LEDOff(b.OnlineLED)
	b.LEDOff(b.ParticipantsLED)
	b.LEDOff(b.TransmitLED)
}
