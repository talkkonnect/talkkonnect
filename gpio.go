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
	"log"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"github.com/talkkonnect/gpio"
)

func (b *Talkkonnect) initGPIO() {

	if TargetBoard != "rpi" {
		return
	}

	if err := rpio.Open(); err != nil {
		log.Println("error: GPIO Error, ", err)
		b.GPIOEnabled = false
		return
	}
	b.GPIOEnabled = true

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

	if StreamButtonPin > 0 {
		StreamButtonPinPullUp := rpio.Pin(StreamButtonPin)
		StreamButtonPinPullUp.PullUp()
	}

	if TxButtonPin > 0 || TxTogglePin > 0 || UpButtonPin > 0 || DownButtonPin > 0 || PanicButtonPin > 0 || CommentButtonPin > 0 {
		rpio.Close()
	}

	if TxButtonPin > 0 {
		TxButton = gpio.NewInput(TxButtonPin)

		go func() {
			for {
				if IsConnected {

					time.Sleep(200 * time.Millisecond)
					currentState, err := TxButton.Read()

					if currentState != TxButtonState && err == nil {
						TxButtonState = currentState

						if b.Stream != nil {
							if TxButtonState == 1 {
								if isTx {
									isTx = false
									b.TransmitStop(true)
									time.Sleep(250 * time.Millisecond)
									if TxCounter {
										txcounter++
										log.Println("debug: Tx Button Count ", txcounter)
									}
								}

							} else {
								log.Println("debug: Tx Button is pressed")
								if !isTx {
									isTx = true
									b.TransmitStart()
									time.Sleep(250 * time.Millisecond)
								}
							}
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()

	}

	if TxTogglePin > 0 {
		TxToggle = gpio.NewInput(TxTogglePin)
		go func() {
			var prevState uint = 1
			for {
				if IsConnected {

					currentState, err := TxToggle.Read()
					time.Sleep(150 * time.Millisecond)

					if err != nil {
						log.Println("error: Error Opening TXToggle Pin")
						break
					}

					if currentState != prevState {
						isTx = !isTx
						if isTx {
							b.TransmitStop(true)
							log.Println("debug: Toggle Stopped Transmitting")
							for {
								currentState, err := TxToggle.Read()
								time.Sleep(150 * time.Millisecond)
								if currentState == 1 && err == nil {
									break
								}
							}
							prevState = 1
							time.Sleep(200 * time.Millisecond)
						}

						if !isTx {
							b.TransmitStart()
							for {
								currentState, err := TxToggle.Read()
								time.Sleep(150 * time.Millisecond)
								if currentState == 1 && err == nil {
									break
								}
							}
							prevState = 1
							log.Println("debug: Toggle Started Transmitting")
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if UpButtonPin > 0 {
		UpButton = gpio.NewInput(UpButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := UpButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != UpButtonState && err == nil {
						UpButtonState = currentState

						if UpButtonState == 1 {
							log.Println("debug: UP Button is released")
						} else {
							log.Println("debug: UP Button is pressed")
							b.ChannelUp()
							time.Sleep(200 * time.Millisecond)
						}

					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if DownButtonPin > 0 {
		DownButton = gpio.NewInput(DownButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := DownButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != DownButtonState && err == nil {
						DownButtonState = currentState

						if DownButtonState == 1 {
							log.Println("debug: Ch Down Button is released")
						} else {
							log.Println("debug: Ch Down Button is pressed")
							b.ChannelDown()
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if PanicButtonPin > 0 {

		PanicButton = gpio.NewInput(PanicButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := PanicButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != PanicButtonState && err == nil {
						PanicButtonState = currentState

						if PanicButtonState == 1 {
							log.Println("debug: Panic Button is released")
						} else {
							log.Println("debug: Panic Button is pressed")
							b.cmdPanicSimulation()
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if CommentButtonPin > 0 {

		CommentButton = gpio.NewInput(CommentButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := CommentButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != CommentButtonState && err == nil {
						CommentButtonState = currentState

						if CommentButtonState == 1 {
							log.Println("debug: Comment Button State 1 setting comment to State 1 Message")
							b.SetComment(CommentMessageOff)
						} else {
							log.Println("debug: Comment Button State 2 setting comment to State 2 Message")
							b.SetComment(CommentMessageOn)
						}
						time.Sleep(200 * time.Millisecond)
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()

	}

	if StreamButtonPin > 0 {

		StreamButton = gpio.NewInput(StreamButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := StreamButton.Read()
					time.Sleep(200 * time.Millisecond)

					if currentState != StreamButtonState && err == nil {
						StreamButtonState = currentState

						if StreamButtonState == 1 {
							log.Println("debug: Stream Button is released")
						} else {
							log.Println("debug: Stream Button is pressed")
							b.cmdPlayback()
							time.Sleep(200 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if OnlineLEDPin > 0 {
		OnlineLED = gpio.NewOutput(OnlineLEDPin, false)
	}

	if ParticipantsLEDPin > 0 {
		ParticipantsLED = gpio.NewOutput(ParticipantsLEDPin, false)
	}

	if TransmitLEDPin > 0 {
		TransmitLED = gpio.NewOutput(TransmitLEDPin, false)
	}

	if HeartBeatLEDPin > 0 {
		HeartBeatLED = gpio.NewOutput(HeartBeatLEDPin, false)
	}

	if AttentionLEDPin > 0 {
		AttentionLED = gpio.NewOutput(AttentionLEDPin, false)
	}

	if LCDBackLightLEDPin > 0 {
		BackLightLED = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
	}

	if VoiceActivityLEDPin > 0 {
		VoiceActivityLED = gpio.NewOutput(VoiceActivityLEDPin, false)
	}

	if VoiceTargetLEDPin > 0 {
		VoiceTargetLED = gpio.NewOutput(VoiceTargetLEDPin, false)
	}

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

func LEDOffAll() {
	if TargetBoard != "rpi" {
		return
	}
	log.Println("debug: Turning Off All LEDS!")

	if OnlineLEDPin > 0 {
		LEDOffFunc(OnlineLED)
	}
	if ParticipantsLEDPin > 0 {
		LEDOffFunc(ParticipantsLED)
	}
	if TransmitLEDPin > 0 {
		LEDOffFunc(TransmitLED)
	}
	if HeartBeatLEDPin > 0 {
		LEDOffFunc(HeartBeatLED)
	}
	if AttentionLEDPin > 0 {
		LEDOffFunc(AttentionLED)
	}
	if LCDBackLightLEDPin > 0 {
		LEDOffFunc(BackLightLED)
	}
	if VoiceActivityLEDPin > 0 {
		LEDOffFunc(VoiceActivityLED)
	}

	if VoiceTargetLEDPin > 0 {
		LEDOffFunc(VoiceTargetLED)
	}

}

func MyLedStripLEDOffAll() {
	MyLedStrip.ledCtrl(SOnlineLED, OffCol)
	MyLedStrip.ledCtrl(SParticipantsLED, OffCol)
	MyLedStrip.ledCtrl(STransmitLED, OffCol)
}

func MyLedStripOnlineLEDOn() {
	MyLedStrip.ledCtrl(SOnlineLED, OnlineCol)
}

func MyLedStripOnlineLEDOff() {
	MyLedStrip.ledCtrl(SOnlineLED, OffCol)
}

func MyLedStripParticipantsLEDOn() {
	MyLedStrip.ledCtrl(SParticipantsLED, ParticipantsCol)
}

func MyLedStripParticipantsLEDOff() {
	MyLedStrip.ledCtrl(SParticipantsLED, OffCol)
}

func MyLedStripTransmitLEDOn() {
	MyLedStrip.ledCtrl(STransmitLED, TransmitCol)
}

func MyLedStripTransmitLEDOff() {
	MyLedStrip.ledCtrl(STransmitLED, OffCol)
}

func relayCommand(relayNo int, command string) {
	// all relays (0)
	if relayNo == 0 {
		for i := 1; i <= int(TotalRelays); i++ {
			if command == "on" {
				log.Println("info: Relay ", i, "On")
				gpio.NewOutput(RelayPins[i], false)

			}
			if command == "off" {
				log.Println("info: Relay ", i, "Off")
				gpio.NewOutput(RelayPins[i], true)
			}
			if command == "pulse" {
				log.Println("info: Relay ", i, "Pulse")
				gpio.NewOutput(RelayPins[i], false)
				time.Sleep(RelayPulseMills * time.Millisecond)
				gpio.NewOutput(RelayPins[i], true)
			}
		}
		return
	}

	//specific relay (Number Between 1 and TotalRelays)
	if relayNo >= 0 && relayNo <= int(TotalRelays) {
		if command == "on" {
			log.Println("info: Relay ", relayNo, "On")
			gpio.NewOutput(RelayPins[relayNo], false)
		}
		if command == "off" {
			log.Println("info: Relay ", relayNo, "Off")
			gpio.NewOutput(RelayPins[relayNo], true)
		}
		if command == "pulse" {
			log.Println("info: Relay ", relayNo, "Pulse")
			gpio.NewOutput(RelayPins[relayNo], false)
			time.Sleep(RelayPulseMills * time.Millisecond)
			gpio.NewOutput(RelayPins[relayNo], true)
		}
	}
}
