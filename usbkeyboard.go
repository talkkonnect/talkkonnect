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
 * Library golang-evdev Copyright (c) 2016 Georgi Valkov. All rights reserved. See Copyright Message in Library source code.
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * usbkeyboard.go -> function in talkkonnect for reading from external usb keyboard
 */

package talkkonnect

import (
	"log"
	"strconv"
	"strings"

	evdev "github.com/gvalkov/golang-evdev"
)

func (b *Talkkonnect) USBKeyboard() {

	device, err := evdev.Open(Config.Global.Hardware.USBKeyboard.USBKeyboardPath)
	if err != nil {
		log.Printf("error: Unable to open USB Keyboard input device: %s\nError: %v It will now Be Disabled\n", Config.Global.Hardware.USBKeyboard.USBKeyboardPath, err)
		return
	}

	var keyPrevStateDown bool

	for {
		events, err := device.Read()
		if err != nil {
			log.Printf("error: Unable to Read Event From USB Keyboard error %v\n", err)
			return
		}

		for _, ev := range events {
			switch ev.Type {
			case evdev.EV_KEY:
				ke := evdev.NewKeyEvent(&ev)

				if ke.State == evdev.KeyDown {
					keyPrevStateDown = true
					if _, ok := USBKeyMap[rune(ke.Scancode)]; ok {
						switch strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) {
						case "soundinterfacepttkey":
							b.TransmitStart()
						}
					}
				}

				// Functions that we allow Repeating Keys Defined Here
				if ke.State == evdev.KeyHold {
					keyPrevStateDown = false
					if _, ok := USBKeyMap[rune(ke.Scancode)]; ok {
						switch strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) {
						case "channelup":
							playIOMedia("usbchannelup")
							b.cmdChannelUp()
						case "channeldown":
							playIOMedia("usbchanneldown")
							b.cmdChannelDown()
						case "volumeup":
							playIOMedia("usbvolup")
							b.cmdVolumeRXUp()
						case "volumedown":
							playIOMedia("usbvoldown")
							b.cmdVolumeRXDown()
						case "volumetxup":
							playIOMedia("usbvolup")
							b.cmdVolumeTXUp()
						case "volumetxdown":
							playIOMedia("usbvoldown")
							b.cmdVolumeTXDown()
						case "pttkey":
							if !b.IsTransmitting {
								b.TransmitStart()
							}
						}
					} else {
						if ke.Scancode != uint16(Config.Global.Hardware.USBKeyboard.NumlockScanID) {
							log.Println("error: Key Not Mapped ASC ", ke.Scancode)
						}
					}
					continue
				}

				if ke.State == evdev.KeyUp {
					if strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) == "pttkey" {
						if b.IsTransmitting {
							b.TransmitStop(false)
						}
					}
				}

				//Key Up & Down One Shot
				if keyPrevStateDown && ke.State == evdev.KeyUp {
					keyPrevStateDown = false
					if _, ok := USBKeyMap[rune(ke.Scancode)]; ok {
						switch strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) {
						case "channelup":
							playIOMedia("usbchannelup")
							b.cmdChannelUp()
						case "channeldown":
							playIOMedia("usbchanneldown")
							b.cmdChannelDown()
						case "serverup":
							playIOMedia("usbserverup")
							b.cmdConnNextServer()
						case "serverdown":
							playIOMedia("usbpreviousserver")
							b.cmdConnPreviousServer()
						case "mute":
							playIOMedia("usbmute")
							b.cmdMuteUnmute("mute")
						case "unmute":
							b.cmdMuteUnmute("unmute")
							playIOMedia("usbunmute")
						case "mute-toggle":
							playIOMedia("usbmutetoggle")
							b.cmdMuteUnmute("toggle")
							playIOMedia("usbmutetoggle")
						case "stream-toggle":
							playIOMedia("usbstreamtoggle")
							b.cmdPlayback()
						case "currentrxvolume":
							playIOMedia("usbcurrentrxvol")
							b.cmdCurrentRXVolume()
						case "volumerxup":
							playIOMedia("usbvolup")
							b.cmdVolumeRXUp()
						case "volumerxdown":
							playIOMedia("usbvoldown")
							b.cmdVolumeRXDown()
						case "currenttxvolume":
							playIOMedia("usbcurrenttxvol")
							b.cmdCurrentTXVolume()
						case "volumetxup":
							playIOMedia("usbvolup")
							b.cmdVolumeTXUp()
						case "volumetxdown":
							playIOMedia("usbvoldown")
							b.cmdVolumeTXDown()
						case "setcomment":
							if USBKeyMap[rune(ke.Scancode)].ParamName == "setcomment" {
								log.Println("info: Set Commment ", USBKeyMap[rune(ke.Scancode)].ParamValue)
								playIOMedia("usbsetcomment")
								b.Client.Self.SetComment(USBKeyMap[rune(ke.Scancode)].ParamValue)
							}
						case "transmitstart":
							playIOMedia("usbstarttx")
							b.cmdStartTransmitting()
						case "transmitstop":
							playIOMedia("usbstoptx")
							b.cmdStopTransmitting()
						case "record":
							playIOMedia("usbrecord")
							b.cmdAudioTrafficRecord()
							b.cmdAudioMicRecord()
						case "voicetargetset":
							voicetarget, err := strconv.Atoi(USBKeyMap[rune(ke.Scancode)].ParamValue)
							if err != nil {
								log.Println("error: Target is Non-Numeric Value")
							} else {
								playIOMedia("usbvoicetarget")
								b.cmdSendVoiceTargets(uint32(voicetarget))
							}
						case "mqttpubpayloadset":
							if USBKeyMap[rune(ke.Scancode)].ParamName == "payloadvalue" {
								playIOMedia("usbmqttpubpayloadset")
								MQTTPublish(USBKeyMap[rune(ke.Scancode)].ParamValue)
							}
						case "changechannel":
							if USBKeyMap[rune(ke.Scancode)].ParamName == "channelname" {
								playIOMedia("changechannel")
								b.ChangeChannel(USBKeyMap[rune(ke.Scancode)].ParamValue)
							}
						case "repeatertoneplay":
							playIOMedia("iorepeatertone")
							b.cmdPlayRepeaterTone()
						case "listentochannelon":
							playIOMedia("usbstartlisten")
							b.listeningToChannels("start")
						case "listentochanneloff":
							playIOMedia("usbstopliosten")
							b.listeningToChannels("stop")
						case "soundinterfacepttkey":
							b.TransmitStop(false)
						case "gpioinput":
							GPIOInputPinControl(USBKeyMap[rune(ke.Scancode)].ParamName, USBKeyMap[rune(ke.Scancode)].ParamValue)
						case "gpiooutput":
							GPIOOutputPinControl(USBKeyMap[rune(ke.Scancode)].ParamName, USBKeyMap[rune(ke.Scancode)].ParamValue)
						default:
							log.Println("error: Command Not Defined ", strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command))
						}
					} else {
						if ke.Scancode != uint16(Config.Global.Hardware.USBKeyboard.NumlockScanID) {
							log.Println("error: Key Not Mapped ASC ", ke.Scancode)
						}
					}
				}
			}
		}
	}
}
