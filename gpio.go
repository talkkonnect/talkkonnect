package talkkonnect

import (
	"github.com/stianeikeland/go-rpio"
	"github.com/talkkonnect/gpio"
	"log"
	"time"
)

func (b *Talkkonnect) initGPIO() {
	if err := rpio.Open(); err != nil {
		log.Println("error: GPIO Error, ", err)
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
						log.Println("info: TX Button is released")
						b.TransmitStop()
					} else {
						log.Println("info: TX Button is pressed")
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
					log.Println("info: UP Button is released")
				} else {
					log.Println("info: UP Button is pressed")
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
					log.Println("info: Ch Down Button is released")
				} else {
					log.Println("info: Ch Down Button is pressed")
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
