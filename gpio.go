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
	"os"
	"os/exec"
	"time"
	hd44780 "github.com/talkkonnect/go-hd44780"
)

var ledpin = 0
var connectFailCounter int = 0

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

	if TxButtonPin > 0 {
		TxButtonPinPullUp := rpio.Pin(TxButtonPin)
		TxButtonPinPullUp.PullUp()
	}

	if TxTogglePin > 0 {
		TxTogglePinPullUp := rpio.Pin(TxTogglePin)
		TxTogglePinPullUp.PullUp()
	}

	if UpButtonPin > 0 {
		UpButtonPinPullUp := rpio.Pin(UpButtonPin)
		UpButtonPinPullUp.PullUp()
	}

	if DownButtonPin > 0 {
		DownButtonPinPullUp := rpio.Pin(DownButtonPin)
		DownButtonPinPullUp.PullUp()
	}

	if PanicButtonPin > 0 {
		PanicButtonPinPullUp := rpio.Pin(PanicButtonPin)
		PanicButtonPinPullUp.PullUp()
	}

	if CommentButtonPin > 0 {
		CommentButtonPinPullUp := rpio.Pin(CommentButtonPin)
		CommentButtonPinPullUp.PullUp()
	}

	if ChimesButtonPin > 0 {
		ChimesButtonPinPullUp := rpio.Pin(ChimesButtonPin)
		ChimesButtonPinPullUp.PullUp()
	}

	if TxButtonPin > 0 || TxTogglePin > 0 || UpButtonPin > 0 || DownButtonPin > 0 || PanicButtonPin > 0 || CommentButtonPin > 0 {
		rpio.Close()
	}

	if TxButtonPin > 0 {
		b.TxButton = gpio.NewInput(TxButtonPin)

		go func() {
			for {
				if b.IsConnected {

					time.Sleep(200 * time.Millisecond)
					currentState, err := b.TxButton.Read()

					if currentState != b.TxButtonState && err == nil {
						b.TxButtonState = currentState

						if b.Stream != nil {
							if b.TxButtonState == 1 {
								if isTx {
									isTx = false
									b.TransmitStop(true)
									time.Sleep(750 * time.Millisecond)
									if TxCounter {
										txcounter++
										log.Println("info: Tx Button Count ", txcounter)
									}
								}

							} else {
								log.Println("info: Tx Button is pressed")
								if !isTx {
									isTx = true
									b.TransmitStart()
									time.Sleep(750 * time.Millisecond)
								}
							}
						}
					}
				} else {
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()

	}

	if TxTogglePin > 0 {
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
								if currentState == 1 && err == nil {
									break
								}
							}
							prevState = 1
							log.Println("info: Toggle Started Transmitting")
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}

	if UpButtonPin > 0 {
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
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}

	if DownButtonPin > 0 {
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
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}

	if PanicButtonPin > 0 {

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
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}

	if CommentButtonPin > 0 {

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
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()

	}

	if ChimesButtonPin > 0 {

		b.ChimesButton = gpio.NewInput(ChimesButtonPin)
		go func() {
			for {
				if b.IsConnected {

					currentState, err := b.ChimesButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != b.ChimesButtonState && err == nil {
						b.ChimesButtonState = currentState

						if b.ChimesButtonState == 1 {
							log.Println("info: Chimes Button is released")
						} else {
							log.Println("info: Chimes Button is pressed")
							b.commandKeyF11()
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					ConnectErrorMessage()
					time.Sleep(2 * time.Second)
				}
			}
		}()
	}

	if OnlineLEDPin > 0 {
		b.OnlineLED = gpio.NewOutput(OnlineLEDPin, false)
	}

	if ParticipantsLEDPin > 0 {
		b.ParticipantsLED = gpio.NewOutput(ParticipantsLEDPin, false)
	}

	if TransmitLEDPin > 0 {
		b.TransmitLED = gpio.NewOutput(TransmitLEDPin, false)
	}

	if HeartBeatLEDPin > 0 {
		b.HeartBeatLED = gpio.NewOutput(HeartBeatLEDPin, false)
	}

	if LCDBackLightLEDPin > 0 {
		b.BackLightLED = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
		BackLightLED = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
	}

	if VoiceActivityLEDPin > 0 {
		VoiceActivityLED = gpio.NewOutput(VoiceActivityLEDPin, false)
	}
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

	if OnlineLEDPin > 0 {
		b.LEDOff(b.OnlineLED)
	}
	if ParticipantsLEDPin > 0 {
		b.LEDOff(b.ParticipantsLED)
	}
	if TransmitLEDPin > 0 {
		b.LEDOff(b.TransmitLED)
	}
	if HeartBeatLEDPin > 0 {
		b.LEDOff(b.HeartBeatLED)
	}
	if LCDBackLightLEDPin > 0 {
		LEDOffFunc(b.BackLightLED)
	}
	if VoiceActivityLEDPin > 0 {
		LEDOffFunc(b.VoiceActivityLED)
	}
}

func ConnectErrorMessage() {
	connectFailCounter++
	if connectFailCounter == 5 || connectFailCounter == 30 || connectFailCounter == 90 {
		log.Println("warn: Cannot Connect to Server, Retrying")
		return
	}

	if connectFailCounter == 100 {
		log.Println("warn: Cannot Connect to Server! Giving Up")
		log.Println("warn: Shutting Down talkkonnect due to multiple connection to server failures")
		t := time.Now()
               if LCDEnabled == true {
                        LcdText = [4]string{"Disconnected At", t.Format("02-01-2006 15:04:05"), "Multiple Reconnect", "Failures!!"}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
                }
		if OLEDEnabled == true {
                        oledDisplay(true, 0, 1, "Disconnected At")
                        oledDisplay(false, 1, 1, t.Format("02-01-2006 15:04:05"))
                        oledDisplay(false, 6, 1, "Multiple Reconnect")
                        oledDisplay(false, 7, 1, "Failures!!")
  		}

		time.Sleep(2 * time.Second)
		c := exec.Command("reset")
		c.Stdout = os.Stdout
		c.Run()
		os.Exit(0)
	}
}
