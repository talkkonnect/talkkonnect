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

	device, err := evdev.Open(USBKeyboardPath)
	if err != nil {
		log.Printf("error: Unable to open USB Keyboard input device: %s\nError: %v It will now Be Disabled\n", USBKeyboardPath, err)
		return
	}

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

				if ke.State != evdev.KeyDown {
					continue
				}

				if _, ok := USBKeyMap[rune(ke.Scancode)]; ok {
					switch strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) {
					case "channelup":
						b.cmdChannelUp()
					case "channeldown":
						b.cmdChannelDown()
					case "serverup":
						b.cmdConnNextServer()
					case "serverdown":
						b.cmdConnPreviousServer()
					case "mute":
						b.cmdMuteUnmute("mute")
					case "unmute":
						b.cmdMuteUnmute("unmute")
					case "mute-toggle":
						b.cmdMuteUnmute("toggle")
					case "stream-toggle":
						b.cmdPlayback()
					case "volumeup":
						b.cmdVolumeUp()
					case "volumedown":
						b.cmdVolumeDown()
					case "setcomment":
						if USBKeyMap[rune(ke.Scancode)].ParamName == "setcomment" {
							log.Println("info: Set Commment ", USBKeyMap[rune(ke.Scancode)].ParamValue)
							b.Client.Self.SetComment(USBKeyMap[rune(ke.Scancode)].ParamValue)
						}
					case "transmitstart":
						b.cmdStartTransmitting()
					case "transmitstop":
						b.cmdStopTransmitting()
					case "record":
						b.cmdAudioTrafficRecord()
						b.cmdAudioMicRecord()
					case "voicetargetset":
						voicetarget, err := strconv.Atoi(USBKeyMap[rune(ke.Scancode)].ParamValue)
						if err != nil {
							log.Println("error: Target is Non-Numeric Value")
						} else {
							b.cmdSendVoiceTargets(uint32(voicetarget))
						}
					default:
						log.Println("Command Not Defined ", strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command))
					}
				} else {
					if ke.Scancode != uint16(NumlockScanID) {
						log.Println("error: Key Not Mapped ASC ", ke.Scancode)
					}
				}
			}
		}
	}
}
