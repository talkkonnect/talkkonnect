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

	"github.com/stianeikeland/go-rpio"
	"github.com/talkkonnect/go-mcp23017"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/max7219"
)

//Variables for Input Buttons/Switches
var (
	TxButtonUsed  bool
	TxButton      gpio.Pin
	TxButtonPin   uint
	TxButtonState uint

	TxToggleUsed  bool
	TxToggle      gpio.Pin
	TxTogglePin   uint
	TxToggleState uint

	UpButtonUsed  bool
	UpButton      gpio.Pin
	UpButtonPin   uint
	UpButtonState uint

	DownButtonUsed  bool
	DownButton      gpio.Pin
	DownButtonPin   uint
	DownButtonState uint

	PanicUsed        bool
	PanicButton      gpio.Pin
	PanicButtonPin   uint
	PanicButtonState uint

	StreamToggleUsed  bool
	StreamButton      gpio.Pin
	StreamButtonPin   uint
	StreamButtonState uint

	CommentUsed        bool
	CommentButton      gpio.Pin
	CommentButtonPin   uint
	CommentButtonState uint

	RotaryUsed bool
	RotaryA    gpio.Pin
	RotaryB    gpio.Pin
	RotaryAPin uint
	RotaryBPin uint

	VolUpButtonUsed  bool
	VolUpButton      gpio.Pin
	VolUpButtonPin   uint
	VolUpButtonState uint

	VolDownButtonUsed  bool
	VolDownButton      gpio.Pin
	VolDownButtonPin   uint
	VolDownButtonState uint

	TrackingUsed        bool
	TrackingButton      gpio.Pin
	TrackingButtonPin   uint
	TrackingButtonState uint

	MQTT0ButtonUsed  bool
	MQTT0Button      gpio.Pin
	MQTT0ButtonPin   uint
	MQTT0ButtonState uint

	MQTT1ButtonUsed  bool
	MQTT1Button      gpio.Pin
	MQTT1ButtonPin   uint
	MQTT1ButtonState uint

	NextServerButtonUsed  bool
	NextServerButton      gpio.Pin
	NextServerButtonPin   uint
	NextServerButtonState uint
)

var D [8]*mcp23017.Device

func (b *Talkkonnect) initGPIO() {

	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}

	if err := rpio.Open(); err != nil {
		log.Println("error: GPIO Error, ", err)
		b.GPIOEnabled = false
		return
	}
	b.GPIOEnabled = true

	// Handle GPIO Expander Pins As Outputs if Enabled
	if Config.Global.Hardware.IO.GPIOExpander.Enabled {
		for _, gpioExpander := range Config.Global.Hardware.IO.GPIOExpander.Chip {
			if Config.Global.Hardware.IO.GPIOExpander.Chip[gpioExpander.ID].Enabled {
				log.Printf("debug: Setting up MCP23017 GPIO Expander on IC2 Bus %v Device No %v\n", gpioExpander.I2Cbus, gpioExpander.MCP23017Device)
				var err error
				D[gpioExpander.MCP23017Device], err = mcp23017.Open(gpioExpander.I2Cbus, gpioExpander.MCP23017Device)
				if err != nil {
					// log.Println("error: Unable To Setup Expander GPIO Chip On I2C Bus " + strconv.Itoa(int(gpioExpander.I2Cbus)) + " Device " + strconv.Itoa(int(gpioExpander.MCP23017Device)) + " With " + err.Error())
					return
				}
				for y := 0; y < 16; y++ {
					if Config.Global.Hardware.IO.Pins.Pin[y].Enabled && Config.Global.Hardware.IO.Pins.Pin[y].Direction == "output" && Config.Global.Hardware.IO.Pins.Pin[y].Type == "mcp23017" {
						log.Printf("debug: Pin %v Enabled as Output\n", y)
						err := D[gpioExpander.MCP23017Device].PinMode(uint8(y), mcp23017.OUTPUT)
						if err != nil {
							log.Printf("error: Cannot Set Pin %v as Output With Error %v\n", y, err)
						}
					}
					if Config.Global.Hardware.IO.Pins.Pin[y].Enabled && Config.Global.Hardware.IO.Pins.Pin[y].Direction == "input" && Config.Global.Hardware.IO.Pins.Pin[y].Type == "mcp23017" {
						log.Printf("debug: Pin %v Enabled as Input\n", y)
						err := D[gpioExpander.MCP23017Device].PinMode(uint8(y), mcp23017.INPUT)
						if err != nil {
							log.Printf("error: Cannot Set Pin %v as Input With Error %v\n", y, err)
						}
					}
				}
			}
		}
	}

	//handle inputs on RPI GPIO
	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Enabled && io.Direction == "input" && io.Type == "gpio" {
			if io.Name == "txptt" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				TxButtonPinPullUp := rpio.Pin(io.PinNo)
				TxButtonPinPullUp.PullUp()
				TxButtonUsed = true
				TxButtonPin = io.PinNo
			}
			if io.Name == "txtoggle" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				TxTogglePinPullUp := rpio.Pin(io.PinNo)
				TxTogglePinPullUp.PullUp()
				TxToggleUsed = true
				TxTogglePin = io.PinNo
			}
			if io.Name == "channelup" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				ChannelUpPinPullUp := rpio.Pin(io.PinNo)
				ChannelUpPinPullUp.PullUp()
				UpButtonUsed = true
				UpButtonPin = io.PinNo
			}
			if io.Name == "channeldown" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				ChannelDownPinPullUp := rpio.Pin(io.PinNo)
				ChannelDownPinPullUp.PullUp()
				DownButtonUsed = true
				DownButtonPin = io.PinNo
			}
			if io.Name == "panic" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				PanicPinPullUp := rpio.Pin(io.PinNo)
				PanicPinPullUp.PullUp()
				PanicUsed = true
				PanicButtonPin = io.PinNo
			}
			if io.Name == "streamtoggle" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				StreamTogglePinPullUp := rpio.Pin(io.PinNo)
				StreamTogglePinPullUp.PullUp()
				StreamToggleUsed = true
				StreamButtonPin = io.PinNo
			}
			if io.Name == "comment" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				CommentPinPullUp := rpio.Pin(io.PinNo)
				CommentPinPullUp.PullUp()
				CommentUsed = true
				CommentButtonPin = io.PinNo
			}
			if io.Name == "rotarya" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				RotaryAPinPullUp := rpio.Pin(io.PinNo)
				RotaryAPinPullUp.PullUp()
				RotaryUsed = true
				RotaryAPin = io.PinNo
			}
			if io.Name == "rotaryb" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				RotaryBPinPullUp := rpio.Pin(io.PinNo)
				RotaryBPinPullUp.PullUp()
				RotaryUsed = true
				RotaryBPin = io.PinNo
			}
			if io.Name == "volup" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				VolUpPinPullUp := rpio.Pin(io.PinNo)
				VolUpPinPullUp.PullUp()
				VolUpButtonUsed = true
				VolUpButtonPin = io.PinNo
			}
			if io.Name == "voldown" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				VolDownPinPullUp := rpio.Pin(io.PinNo)
				VolDownPinPullUp.PullUp()
				VolDownButtonUsed = true
				VolDownButtonPin = io.PinNo
			}
			if io.Name == "tracking" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				TrackingPinPullUp := rpio.Pin(io.PinNo)
				TrackingPinPullUp.PullUp()
				TrackingUsed = true
				TrackingButtonPin = io.PinNo
			}
			if io.Name == "mqtt0" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				MQTT0PinPullUp := rpio.Pin(io.PinNo)
				MQTT0PinPullUp.PullUp()
				MQTT0ButtonUsed = true
				MQTT0ButtonPin = io.PinNo
			}
			if io.Name == "mqtt1" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				MQTT1PinPullUp := rpio.Pin(io.PinNo)
				MQTT1PinPullUp.PullUp()
				MQTT1ButtonUsed = true
				MQTT1ButtonPin = io.PinNo
			}
			if io.Name == "nextserver" && io.PinNo > 0 {
				log.Printf("debug: GPIO Setup Input Device %v Name %v PinNo %v", io.Device, io.Name, io.PinNo)
				NextServerPinPullUp := rpio.Pin(io.PinNo)
				NextServerPinPullUp.PullUp()
				NextServerButtonUsed = true
				NextServerButtonPin = io.PinNo
			}
		}
	}

	if TxButtonUsed || TxToggleUsed || UpButtonUsed || DownButtonUsed || PanicUsed || StreamToggleUsed || CommentUsed || RotaryUsed || VolUpButtonUsed || VolDownButtonUsed || TrackingUsed || MQTT0ButtonUsed || MQTT1ButtonUsed || NextServerButtonUsed {
		rpio.Close()
	}

	if TxButtonUsed {
		TxButton = gpio.NewInput(TxButtonPin)
		go func() {
			for {
				if IsConnected {
					time.Sleep(150 * time.Millisecond)
					currentState, err := TxButton.Read()
					if currentState != TxButtonState && err == nil {
						TxButtonState = currentState
						if b.Stream != nil {
							if TxButtonState == 1 {
								if isTx {
									isTx = false
									b.TransmitStop(true)
									playIOMedia("iotxpttstop")
									if Config.Global.Software.Settings.TxCounter {
										txcounter++
										log.Println("debug: Tx Button Count ", txcounter)
									}
								}
							} else {
								log.Println("debug: Tx Button is pressed")
								if !isTx {
									isTx = true
									playIOMedia("iotxpttstart")
								} else {
									time.Sleep(150 * time.Millisecond)
								}
								b.TransmitStart()
							}
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if TxToggleUsed {
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
							playIOMedia("iotxtogglestop")
							for {
								currentState, err := TxToggle.Read()
								time.Sleep(150 * time.Millisecond)
								if currentState == 1 && err == nil {
									break
								}
							}
							prevState = 1
							time.Sleep(150 * time.Millisecond)
						}
						if !isTx {
							if Config.Global.Software.Sounds.Input.Enabled {
								var inputEventSoundFile InputEventSoundFileStruct = findInputEventSoundFile("txtogglestart")
								if inputEventSoundFile.Enabled {
									go aplayLocal(inputEventSoundFile.File)
								}
							}
							playIOMedia("txtogglestart")
							b.TransmitStart()
							log.Println("debug: Toggle Started Transmitting")
							for {
								currentState, err := TxToggle.Read()
								time.Sleep(150 * time.Millisecond)
								if currentState == 1 && err == nil {
									break
								}
							}
							prevState = 1
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if UpButtonUsed {
		UpButton = gpio.NewInput(UpButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := UpButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != UpButtonState && err == nil {
						UpButtonState = currentState
						if UpButtonState == 1 {
							log.Println("debug: UP Button is released")
						} else {
							log.Println("debug: UP Button is pressed")
							playIOMedia("iochannelup")
							b.ChannelUp()
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if DownButtonUsed {
		DownButton = gpio.NewInput(DownButtonPin)
		go func() {
			for {
				if IsConnected {

					currentState, err := DownButton.Read()
					time.Sleep(150 * time.Millisecond)

					if currentState != DownButtonState && err == nil {
						DownButtonState = currentState

						if DownButtonState == 1 {
							log.Println("debug: Ch Down Button is released")
						} else {
							log.Println("debug: Ch Down Button is pressed")
							playIOMedia("iochanneldown")
							b.ChannelDown()
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if PanicUsed {
		PanicButton = gpio.NewInput(PanicButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := PanicButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != PanicButtonState && err == nil {
						PanicButtonState = currentState

						if PanicButtonState == 1 {
							log.Println("debug: Panic Button is released")
						} else {
							log.Println("debug: Panic Button is pressed")
							playIOMedia("iopanic")
							b.cmdPanicSimulation()
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if CommentUsed {
		CommentButton = gpio.NewInput(CommentButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := CommentButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != CommentButtonState && err == nil {
						CommentButtonState = currentState
						if CommentButtonState == 1 {
							playIOMedia("iocommenton")
							log.Println("debug: Comment Button State 1 setting comment to State 1 Message ", Config.Global.Hardware.Comment.CommentMessageOff)
							b.SetComment(Config.Global.Hardware.Comment.CommentMessageOff)
						} else {
							playIOMedia("iocommentoff")
							log.Println("debug: Comment Button State 2 setting comment to State 2 Message ", Config.Global.Hardware.Comment.CommentMessageOn)
							b.SetComment(Config.Global.Hardware.Comment.CommentMessageOn)
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()

	}

	if StreamToggleUsed {
		StreamButton = gpio.NewInput(StreamButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := StreamButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != StreamButtonState && err == nil {
						StreamButtonState = currentState
						if StreamButtonState == 1 {
							log.Println("debug: Stream Button is released")
						} else {
							playIOMedia("iostreamtoggle")
							log.Println("debug: Stream Button is pressed")
							b.cmdPlayback()
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if RotaryUsed {
		RotaryA = gpio.NewInput(RotaryAPin)
		RotaryB = gpio.NewInput(RotaryBPin)
		go func() {
			var LastStateA uint = 0
			for {
				if IsConnected {
					time.Sleep(5 * time.Millisecond)
					currentStateA, err0 := RotaryA.Read()
					if currentStateA != LastStateA && err0 == nil {
						currentStateB, err1 := RotaryB.Read()
						if currentStateB != currentStateA && err1 == nil {
							playIOMedia("iorotarycw")
							log.Println("debug: Rotating Clockwise")
							b.cmdChannelUp()
						} else {
							log.Println("debug: Rotating CounterClockwise")
							playIOMedia("iorotaryccw")
							b.cmdChannelDown()
						}
					}
					LastStateA = currentStateA
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VolUpButtonUsed {
		VolUpButton = gpio.NewInput(VolUpButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := VolUpButton.Read()
					time.Sleep(150 * time.Millisecond)

					if currentState != VolUpButtonState && err == nil {
						VolUpButtonState = currentState

						if VolUpButtonState == 1 {
							log.Println("debug: Vol UP Button is released")
						} else {
							log.Println("debug: Vol UP Button is pressed")
							playIOMedia("iovolup")
							b.cmdVolumeUp()
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if VolDownButtonUsed {
		VolDownButton = gpio.NewInput(VolDownButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := VolDownButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != VolDownButtonState && err == nil {
						VolDownButtonState = currentState
						if VolDownButtonState == 1 {
							log.Println("debug: Vol Down Button is released")
						} else {
							log.Println("debug: Vol Down Button is pressed")
							playIOMedia("iovoldown")
							b.cmdVolumeDown()
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if TrackingUsed {
		TrackingButton = gpio.NewInput(TrackingButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := TrackingButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != TrackingButtonState && err == nil {
						TrackingButtonState = currentState
						if TrackingButtonState == 1 {
							playIOMedia("iotrackingon")
							log.Println("debug: Tracking Button State 1 setting comment to on  ")
							// place holder to start tracking timer
						} else {
							playIOMedia("iotrackingoff")
							log.Println("debug: Tracking Button State 1 setting comment to off ")
							// place holder to start tracking timer
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MQTT0ButtonUsed {
		MQTT0Button = gpio.NewInput(MQTT0ButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := MQTT0Button.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != MQTT0ButtonState && err == nil {
						MQTT0ButtonState = currentState
						if MQTT0ButtonState == 1 {
							log.Println("debug: MQTT0 Button is released")
						} else {
							log.Println("debug: MQTT0 Button is pressed")
							playIOMedia("iomqtt0")
							//mqtt button0 send command placeholder
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if MQTT1ButtonUsed {
		MQTT1Button = gpio.NewInput(MQTT1ButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := MQTT1Button.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != MQTT1ButtonState && err == nil {
						MQTT1ButtonState = currentState
						if MQTT1ButtonState == 1 {
							log.Println("debug: MQTT1 Button is released")
						} else {
							log.Println("debug: MQTT1 Button is pressed")
							playIOMedia("iomqtt1")
							//mqtt button1 send command placeholder
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}

	if NextServerButtonUsed {
		NextServerButton = gpio.NewInput(NextServerButtonPin)
		go func() {
			for {
				if IsConnected {
					currentState, err := NextServerButton.Read()
					time.Sleep(150 * time.Millisecond)
					if currentState != NextServerButtonState && err == nil {
						NextServerButtonState = currentState
						if NextServerButtonState == 1 {
							log.Println("debug: NextServer Button is released")
						} else {
							log.Println("debug: NextServer Button is pressed")
							playIOMedia("iocnextserver")
							b.ChannelUp()
							time.Sleep(150 * time.Millisecond)
						}
					}
				} else {
					time.Sleep(1 * time.Second)
				}
			}
		}()
	}
}

func GPIOOutPin(name string, command string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}

	for _, io := range Config.Global.Hardware.IO.Pins.Pin {

		if io.Enabled && io.Direction == "output" && io.Name == name {
			if command == "on" {
				switch io.Type {
				case "gpio":
					log.Printf("debug: Turning On %v at pin %v Output GPIO\n", io.Name, io.PinNo)
					gpio.NewOutput(io.PinNo, true)
				case "mcp23017":
					log.Printf("debug: Turning On %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					err := D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.LOW)
					if err != nil {
						log.Printf("error: Error Turning On %v at pin %v Output mcp23017 with error %v\n", io.Name, io.PinNo, err)
					}
				default:
					log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
				}
				break
			}

			if command == "off" {
				switch io.Type {
				case "gpio":
					log.Printf("debug: Turning Off %v at pin %v Output GPIO\n", io.Name, io.PinNo)
					gpio.NewOutput(io.PinNo, false)
				case "mcp23017":
					log.Printf("debug: Turning Off %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					err := D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.HIGH)
					if err != nil {
						log.Printf("error: Error Turning Off %v at pin %v Output mcp23017 with error %v\n", io.Name, io.PinNo, err)
					}
				default:
					log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
				}
				break
			}

			if command == "pulse" {
				switch io.Type {
				case "gpio":
					log.Printf("debug: Pulsing %v at pin %v Output GPIO\n", io.Name, io.PinNo)
					gpio.NewOutput(io.PinNo, false)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Leading * time.Millisecond)
					gpio.NewOutput(io.PinNo, true)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Pulse * time.Millisecond)
					gpio.NewOutput(io.PinNo, false)
					time.Sleep(Config.Global.Hardware.IO.Pulse.Trailing * time.Millisecond)
				case "mcp23017":
					log.Printf("debug: Pulsing %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					err := D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.HIGH)
					if err != nil {
						log.Printf("error: Error Turning Off %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					}
					time.Sleep(Config.Global.Hardware.IO.Pulse.Leading * time.Millisecond)
					err = D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.LOW)
					if err != nil {
						log.Printf("error: Error Turning On %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					}
					time.Sleep(Config.Global.Hardware.IO.Pulse.Pulse * time.Millisecond)
					err = D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.HIGH)
					if err != nil {
						log.Printf("error: Error Turning Off %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
					}
					time.Sleep(Config.Global.Hardware.IO.Pulse.Trailing * time.Millisecond)
				default:
					log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
				}
				break
			}
		}
	}
}

func GPIOOutAll(name string, command string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}

	for _, io := range Config.Global.Hardware.IO.Pins.Pin {
		if io.Enabled && io.Direction == "output" && io.Device == "led/relay" {
			switch io.Type {
			case "gpio":
				if command == "on" {
					log.Printf("debug: Turning On %v Output GPIO\n", io.Name)
					gpio.NewOutput(io.PinNo, true)
				}
				if command == "off" {
					log.Printf("debug: Turning Off %v Output GPIO\n", io.Name)
					gpio.NewOutput(io.PinNo, false)
				}
			case "mcp23017":
				if command == "on" {
					log.Printf("debug: Turning On %v Output mcp23017\n", io.Name)
					if D[io.ID] != nil {
						err := D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.LOW)
						if err != nil {
							log.Printf("error: Error Turning On %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
						}
					}
				}
				if command == "off" {
					log.Printf("debug: Turning Off %v Output mcp23017\n", io.Name)
					if D[io.ID] != nil {
						err := D[io.ID].DigitalWrite(uint8(io.PinNo), mcp23017.HIGH)
						if err != nil {
							log.Printf("error: Error Turning Off %v at pin %v Output mcp23017\n", io.Name, io.PinNo)
						}
					}
				}
			default:
				log.Println("error: GPIO Types Currently Supported are gpio or mcp23017 only!")
			}
		}
	}
}

func MyLedStripGPIOOffAll() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning Off All LEDStrip LEDs")
		MyLedStrip.ledCtrl(SOnlineLED, OffCol)
		MyLedStrip.ledCtrl(SVoiceActivityLED, OffCol)
		MyLedStrip.ledCtrl(STransmitLED, OffCol)
	}
}

func MyLedStripOnlineLEDOn() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning On LEDStrip Online LED")
		MyLedStrip.ledCtrl(SOnlineLED, OnlineCol)
	}
}

func MyLedStripOnlineLEDOff() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning Off LEDStrip Online LED")
		MyLedStrip.ledCtrl(SOnlineLED, OffCol)

	}
}

func MyLedStripVoiceActivityLEDOn() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning On LEDStrip VoiceActivity LED")
		MyLedStrip.ledCtrl(SVoiceActivityLED, VoiceActivityCol)
	}
}

func MyLedStripVoiceActivityLEDOff() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning Off LEDStrip VoiceActivity LED")
		MyLedStrip.ledCtrl(SVoiceActivityLED, OffCol)

	}
}

func MyLedStripTransmitLEDOn() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning On LEDStrip Transmit LED")
		MyLedStrip.ledCtrl(STransmitLED, TransmitCol)

	}
}
func MyLedStripTransmitLEDOff() {
	if Config.Global.Hardware.LedStripEnabled {
		log.Println("debug: Turning Off LEDStrip Transmit LED")
		MyLedStrip.ledCtrl(STransmitLED, OffCol)

	}
}

func Max7219(max7219Cascaded int, spiBus int, spiDevice int, brightness byte, toDisplay string) {
	if Config.Global.Hardware.IO.Max7219.Enabled {
		mtx := max7219.NewMatrix(max7219Cascaded)
		err := mtx.Open(spiBus, spiDevice, brightness)
		if err != nil {
			log.Fatal(err)

		}
		mtx.Device.SevenSegmentDisplay(toDisplay)
		defer mtx.Close()
	}
}
