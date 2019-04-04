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
	"time"
)

var ledpin = 0
var loglevel = 3

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

	TxTogglePinPullUp := rpio.Pin(TxTogglePin)
	TxTogglePinPullUp.PullUp()

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
	go func() {
		for {
			if b.IsConnected {

				currentState, err := b.TxButton.Read()
				time.Sleep(200 * time.Millisecond)

				if currentState != b.TxButtonState && err == nil {
					b.TxButtonState = currentState

					if b.Stream != nil {
						if b.TxButtonState == 1 {
							log.Println("info: Tx Button is released")
							if isTx {
								b.TransmitStop(true)
								time.Sleep(200 * time.Millisecond)
								isTx = false
								if loglevel > 2 {
									txcounter++
									log.Println("info: Tx Button Count ", txcounter)
								}
							}

						} else {
							log.Println("info: Tx Button is pressed")
							if !isTx {
								b.TransmitStart()
								time.Sleep(200 * time.Millisecond)
								isTx = true
							}
						}
					}
				}
			} else {
				_, err := b.TxButton.Read()
				if err != nil && b.IsConnected {
					log.Println("warn: Error Reading TxButton")
				}
			}
		}
	}()

	b.TxToggle = gpio.NewInput(TxTogglePin)
	go func() {
		var prevState uint = 1
		for {
			if b.IsConnected {

				currentState, err := b.TxToggle.Read()
				time.Sleep(150 * time.Millisecond)

				if err != nil {
					log.Println("warn: Error Opening TXToggle Pin")
					break
				}

				if currentState != prevState {
					isTx = !isTx
					if isTx {
						b.TransmitStop(true)
						log.Println("info: Toggle Stopped Transmitting")
						for {
							currentState, err := b.TxToggle.Read()
                                			time.Sleep(150 * time.Millisecond)
							if currentState == 1 && err == nil {
								break
							}
						}
						prevState = 1
						time.Sleep(200 * time.Millisecond)
					}

					if isTx == false {
						b.TransmitStart()
						for {
							currentState, err := b.TxToggle.Read()
                                			time.Sleep(150 * time.Millisecond)
							if currentState == 1 && err ==nil {
								break
							}
						}
						prevState = 1
						log.Println("info: Toggle Started Transmitting")
						time.Sleep(200 * time.Millisecond)
					}
				}
			} else {
				_, err := b.TxToggle.Read()
				if err != nil && b.IsConnected {
						log.Println("warn: Error Reading TxToggle Button")
				}
			}
		}
	}()

	b.UpButton = gpio.NewInput(UpButtonPin)
	go func() {
		for {
			if b.IsConnected {

				currentState, err := b.UpButton.Read()
				time.Sleep(200 * time.Millisecond)

				if currentState != b.UpButtonState && err == nil {
					b.UpButtonState = currentState

					if b.UpButtonState == 1 {
						log.Println("info: UP Button is released")
					} else {
						log.Println("info: UP Button is pressed")
						b.ChannelUp()
						time.Sleep(200 * time.Millisecond)
					}

				}
			} else {
				_, err := b.UpButton.Read()
				if err != nil && b.IsConnected {
						log.Println("warn: Error Reading TxToggle Button")
				}
			}
		}
	}()

	b.DownButton = gpio.NewInput(DownButtonPin)
	go func() {
		for {
			if b.IsConnected {

				currentState, err := b.DownButton.Read()
				time.Sleep(200 * time.Millisecond)

				if currentState != b.DownButtonState && err == nil {
					b.DownButtonState = currentState

					if b.DownButtonState == 1 {
						log.Println("info: Ch Down Button is released")
					} else {
						log.Println("info: Ch Down Button is pressed")
						b.ChannelDown()
						time.Sleep(200 * time.Millisecond)
					}
				}
			} else {
				_, err := b.DownButton.Read()
				if err != nil && b.IsConnected {
					log.Println("warn: Error Reading Down Button")
				}
			}
		}
	}()

	b.PanicButton = gpio.NewInput(PanicButtonPin)
	go func() {
		for {
			if b.IsConnected {

				currentState, err := b.PanicButton.Read()
				time.Sleep(200 * time.Millisecond)

				if currentState != b.PanicButtonState && err == nil {
					b.PanicButtonState = currentState

					if b.PanicButtonState == 1 {
						log.Println("info: Panic Button is released")
					} else {
						log.Println("info: Panic Button is pressed")
						b.commandKeyCtrlP()
						time.Sleep(200 * time.Millisecond)
					}
				}
			} else {
				_, err := b.PanicButton.Read()
				if err != nil && b.IsConnected {
					log.Println("warn: Error Reading Panic Button ", err)
				}
			}
		}
	}()

	b.CommentButton = gpio.NewInput(CommentButtonPin)
	go func() {
		for {
			if b.IsConnected {

				currentState, err := b.CommentButton.Read()
				time.Sleep(200 * time.Millisecond)

				if currentState != b.CommentButtonState && err == nil {
					b.CommentButtonState = currentState

					if b.CommentButtonState == 1 {
						log.Println("info: Comment Button State 1 setting comment to State 1 Message")
						b.SetComment(CommentMessageOff)
					} else {
						log.Println("info: Comment Button State 2 setting comment to State 2 Message")
						b.SetComment(CommentMessageOn)
					}
					time.Sleep(200 * time.Millisecond)
				}
			} else {
				_, err := b.CommentButton.Read()
				if err != nil && b.IsConnected {
					log.Println("warn: Error Reading Comment Button ", err)
				}
			}
		}
	}()

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
