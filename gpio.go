/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * talkkonnect is the based on talkiepi and barnard by Daniel Chote and Tim Cooper
 *
 * The Initial Developer of the Original Code is
 * Suvir Kumar <suvir@talkkonnect.com>
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * gpio.go talkkonnects function to connect to SBC GPIO
 */

package talkkonnect

import (
	"github.com/stianeikeland/go-rpio"
	"github.com/talkkonnect/gpio"
	"log"
	"strconv"
	"time"
)

var ledpin = 0
var txcounter = 0
var  loglevel = 3

func (b *Talkkonnect) initGPIO() {
	if TargetBoard != "rpi" {
		return
	}

	if err := rpio.Open(); err != nil {
		log.Println("alert: GPIO Error, ", err)
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

	PanicButtonPinPullUp := rpio.Pin(PanicButtonPin)
	PanicButtonPinPullUp.PullUp()

	CommentButtonPinPullUp := rpio.Pin(CommentButtonPin)
	CommentButtonPinPullUp.PullUp()

	rpio.Close()

	b.TxButton = gpio.NewInput(TxButtonPin)

	TxTimeOutTimer := time.NewTimer(1 * time.Millisecond)

	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			currentState, err := b.TxButton.Read()

			if currentState != b.TxButtonState && err == nil {
				b.TxButtonState = currentState

				if b.Stream != nil {
					if b.TxButtonState == 1 {
						log.Println("info: Tx Button is released")
						b.TransmitStop(true)
						time.Sleep(200 * time.Millisecond)
						if loglevel > 2 {
							txcounter++
							log.Println("info: Tx Button Count ", txcounter)
						}
						if TxTimeOutEnabled {
							TxTimeOutTimer.Stop()
						}

					} else {
						log.Println("info: Tx Button is pressed")
						b.TransmitStart()
						time.Sleep(200 * time.Millisecond)

						if TxTimeOutEnabled {
							log.Println("warn: Starting Tx Timeout Timer Now")
							TxTimeOutTimer = time.NewTimer(time.Duration(TxTimeOutSecs) * time.Second)

							go func() {
								for {
									select {
									case <-TxTimeOutTimer.C:
										TxTimeOutTimer.Stop()
										b.TransmitStop(false)
										log.Println("warn: Tx Timed out After ", strconv.Itoa(TxTimeOutSecs), " Seconds.")
									}
								}
							}()
						}
					}

				}

			}

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

	b.PanicButton = gpio.NewInput(PanicButtonPin)
	go func() {
		for {
			currentState, err := b.PanicButton.Read()

			if currentState != b.PanicButtonState && err == nil {
				b.PanicButtonState = currentState

				if b.PanicButtonState == 1 {
					log.Println("info: Panic Button is released")
				} else {
					log.Println("info: Panic Button is pressed")
					b.commandKeyCtrlP()
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	b.CommentButton = gpio.NewInput(CommentButtonPin)
	go func() {
		for {
			currentState, err := b.CommentButton.Read()

			if currentState != b.CommentButtonState && err == nil {
				b.CommentButtonState = currentState

				if b.CommentButtonState == 1 {
					log.Println("info: Comment Button State 1 setting comment to State 1 Message")
					b.SetComment(CommentMessageOff)
				} else {
					log.Println("info: Comment Button State 2 setting comment to State 2 Message")
					b.SetComment(CommentMessageOn)
				}
			}

			time.Sleep(500 * time.Millisecond)
		}
	}()

	// then we can do our gpio stuff
	b.OnlineLED = gpio.NewOutput(OnlineLEDPin, false)
	b.ParticipantsLED = gpio.NewOutput(ParticipantsLEDPin, false)
	b.TransmitLED = gpio.NewOutput(TransmitLEDPin, false)
	b.HeartBeatLED = gpio.NewOutput(HeartBeatLEDPin, false)
	b.BackLightLED = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
	BackLightLED = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
	VoiceActivityLED = gpio.NewOutput(VoiceActivityLEDPin, false)
}

func (b *Talkkonnect) LEDOn(LED gpio.Pin) {
	if TargetBoard != "rpi" {
		return
	}

	LED.High()
}

func (b *Talkkonnect) LEDOff(LED gpio.Pin) {
	if TargetBoard != "rpi" {
		return
	}
	LED.Low()
}

func LEDOnFunc(LED gpio.Pin) {
	if TargetBoard != "rpi" {
		return
	}
	LED.High()
}

func LEDOffFunc(LED gpio.Pin) {
	if TargetBoard != "rpi" {
		return
	}
	LED.Low()
}

func (b *Talkkonnect) LEDOffAll() {
	if TargetBoard != "rpi" {
		return
	}
	log.Println("warn: Turning Off All LEDS!")
	b.LEDOff(b.OnlineLED)
	b.LEDOff(b.ParticipantsLED)
	b.LEDOff(b.TransmitLED)
	b.LEDOff(b.HeartBeatLED)
	LEDOffFunc(b.BackLightLED)
	LEDOffFunc(b.VoiceActivityLED)
}
