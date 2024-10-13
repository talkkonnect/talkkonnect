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
 *
 */

package talkkonnect

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

var prevParticipantCount = 1
var participantsLED bool = false

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	if Config.Global.Hardware.TargetBoard == "rpi" {
		GPIOOutPin("online", "on")
		//MyLedStripOnlineLEDOn()
	}

	IsConnected = true
	//MyLedStripOnlineLEDOn()
	b.Client = e.Client
	ConnectAttempts = 1

	//serialize tokens send one by one from slice to server
	if len(Tokens[AccountIndex]) > 0 {
		ATokens := make(gumble.AccessTokens, len(Tokens[AccountIndex]))
		copy(ATokens[:], Tokens[AccountIndex])
		b.Client.Send(ATokens)
	}

	log.Printf("info: Connected to %s Address %s on attempt %d index [%d] ", b.Name, b.Client.Conn.RemoteAddr(), b.ConnectAttempts, AccountIndex)

	if e.WelcomeMessage != nil {
		var tmessage string = fmt.Sprintf("%v", esc(*e.WelcomeMessage))
		for _, line := range strings.Split(strings.TrimSpace(tmessage), "\n") {
			log.Println("info: ", strings.TrimSpace(line))
		}
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			LcdText[0] = b.Name //b.Address
			LcdText[1] = "(" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + ")" + b.Client.Self.Channel.Name
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			LCDIsDark = false
			oledDisplay(true, 0, 0, "")                    // clear the screen
			oledDisplay(false, 0, OLEDStartColumn, b.Name) //b.Address
			oledDisplay(false, 1, OLEDStartColumn, "("+strconv.Itoa(len(b.Client.Self.Channel.Users))+")"+b.Client.Self.Channel.Name)
			oledDisplay(false, 6, OLEDStartColumn, "Please Visit")
			oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
		}
	}

	if b.ChannelName != "" {
		b.ChangeChannel(b.ChannelName)
		prevChannelID = b.Client.Self.Channel.ID
	}
}

func (b *Talkkonnect) OnDisconnect(e *gumble.DisconnectEvent) {
	var reason string

	if Config.Global.Hardware.TargetBoard == "rpi" {
		b.BackLightTimer()
		GPIOOutPin("online", "off")
		GPIOOutAll("led/relay", "off")
	}

	//1 DisconnectError DisconnectType = iota + 1
	if e.Type.Has(1) {
		reason = "[error]"
	}

	//2 DisconnectKicked
	if e.Type.Has(2) {
		reason = "[kicked]"
	}

	//4 DisconnectBanned
	if e.Type.Has(3) {
		reason = "[banned]"
	}

	//8 DisconnectUser
	if e.Type.Has(4) {
		reason = "[user]"
	}

	IsConnected = false
	//MyLedStripOnlineLEDOff()

	FatalCleanUp("Connection to Mumble Server " + b.Address + " Lost Reason " + reason)

}

func (b *Talkkonnect) OnTextMessage(e *gumble.TextMessageEvent) {
	if Config.Global.Hardware.TargetBoard == "rpi" {
		b.BackLightTimer()
	}

	var eventSound EventSoundStruct = findEventSound("message")
	if eventSound.Enabled {
		if v, err := strconv.Atoi(eventSound.Volume); err == nil {
			localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)
		}
	}

	if len(cleanstring(e.Message)) > 105 {
		log.Println("warn: Message Too Long to Be Displayed on Screen")
		tmessage = strings.TrimSpace(cleanstring(e.Message)[:105])
	} else {
		tmessage = strings.TrimSpace(cleanstring(e.Message))
	}

	var sender string

	if e.Sender != nil {
		sender = strings.TrimSpace(cleanstring(e.Sender.Name))
		log.Println("info: Sender Name is ", sender)
	} else {
		sender = ""
	}

	for _, tts := range Config.Global.Software.TTS.Sound {
		if tts.Action == "message" {
			if tts.Enabled {
				voiceMessage := fmt.Sprintf("Message from %v %v\n", sender, cleanstring(e.Message))
				if Config.Global.Software.TTSMessages.TTSMessageFromTag {
					b.TTSPlayerMessage(voiceMessage, Config.Global.Software.TTSMessages.LocalPlay, Config.Global.Software.TTSMessages.PlayIntoStream)
				} else {
					b.TTSPlayerMessage(cleanstring(e.Message), Config.Global.Software.TTSMessages.LocalPlay, Config.Global.Software.TTSMessages.PlayIntoStream)
				}
			}
		}
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText[0] = "Msg From " + sender
			LcdText[1] = tmessage
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 2, OLEDStartColumn, "Msg From "+sender)
			if len(tmessage) <= 21 {
				oledDisplay(false, 3, OLEDStartColumn, tmessage)
				oledDisplay(false, 4, OLEDStartColumn, "")
				oledDisplay(false, 5, OLEDStartColumn, "")
				oledDisplay(false, 6, OLEDStartColumn, "")
				oledDisplay(false, 7, OLEDStartColumn, "")
				return
			}
			if len(tmessage) <= 42 {
				oledDisplay(false, 3, OLEDStartColumn, tmessage[0:21])
				oledDisplay(false, 4, OLEDStartColumn, tmessage[21:])
				oledDisplay(false, 5, OLEDStartColumn, "")
				oledDisplay(false, 6, OLEDStartColumn, "")
				oledDisplay(false, 7, OLEDStartColumn, "")
				return
			}
			if len(tmessage) <= 63 {
				oledDisplay(false, 3, OLEDStartColumn, tmessage[0:21])
				oledDisplay(false, 4, OLEDStartColumn, tmessage[21:42])
				oledDisplay(false, 5, OLEDStartColumn, tmessage[42:])
				oledDisplay(false, 6, OLEDStartColumn, "")
				oledDisplay(false, 7, OLEDStartColumn, "")
				return
			}
			if len(tmessage) <= 84 {
				oledDisplay(false, 3, OLEDStartColumn, tmessage[0:21])
				oledDisplay(false, 4, OLEDStartColumn, tmessage[21:42])
				oledDisplay(false, 5, OLEDStartColumn, tmessage[42:63])
				oledDisplay(false, 6, OLEDStartColumn, tmessage[63:])
				oledDisplay(false, 7, OLEDStartColumn, "")
				return
			}
			if len(tmessage) <= 105 {
				oledDisplay(false, 3, OLEDStartColumn, tmessage[0:20])
				oledDisplay(false, 4, OLEDStartColumn, tmessage[21:44])
				oledDisplay(false, 5, OLEDStartColumn, tmessage[42:63])
				oledDisplay(false, 6, OLEDStartColumn, tmessage[63:84])
				oledDisplay(false, 7, OLEDStartColumn, tmessage[84:])
				return
			}
		}
	}

}

func (b *Talkkonnect) OnUserChange(e *gumble.UserChangeEvent) {
	b.BackLightTimer()

	var info string = ""
	var shortInfo string = ""

	//1 UserChangeConnected
	if e.Type.Has(1) {
		info = info + "[connected]"
		shortInfo = "Conn " + time.Now().Format("15:04:05")
	}

	//2 UserChangeDisconnected
	if e.Type.Has(2) {
		info = info + "[disconnected]"
		shortInfo = "Disconn " + time.Now().Format("15:04:05")
	}

	//4 UserChangeKicked
	if e.Type.Has(4) {
		info = info + "[kicked]"
		shortInfo = "Kick " + time.Now().Format("15:04:05")
	}

	//8 UserChangeBanned
	if e.Type.Has(8) {
		info = info + "[banned]"
		shortInfo = "Ban " + time.Now().Format("15:04:05")
	}

	//16 UserChangeRegistered
	if e.Type.Has(16) {
		info = info + "[registered]"
	}

	//32 UserChangeUnregistered
	if e.Type.Has(32) {
		info = info + "[unregistered]"
	}

	//64 UserChangeName
	if e.Type.Has(64) {
		info = info + "[changed name]"
	}

	//128 UserChangeChannel
	if e.Type.Has(128) {
		if OLEDEnabled {
			oledDisplay(false, 1, OLEDStartColumn, "("+strconv.Itoa(len(b.Client.Self.Channel.Users))+")"+b.Client.Self.Channel.Name)
		}
		info = info + "[changed channel]"
		if !e.Type.Has(1) {
			shortInfo = "Chg Chan " + time.Now().Format("15:04:05")
		}
	}

	//256 UserChangeComment
	if e.Type.Has(256) {
		info = info + "[changed comment]"
	}

	//512 UserChangeAudio
	if e.Type.Has(512) {
		info = info + "[changed audio]"
	}

	//1024 UserChangeTexture
	if e.Type.Has(1024) {
		info = info + "[changed texture]"
	}

	//2048 UserChangePrioritySpeaker
	if e.Type.Has(2048) {
		info = info + "[changed priority speaker]"
	}

	//4096 UserChangeRecording
	if e.Type.Has(4096) {
		info = info + "[change recording]"
	}

	//8192 UserChangeStats
	if e.Type.Has(8192) {
		info = info + "[change stats]"
	}

	if e.Type.Has(1) || e.Type.Has(2) || e.Type.Has(128) {
		if len(b.Client.Self.Channel.Users) > 1 && (len(b.Client.Self.Channel.Users) != prevParticipantCount) {
			GPIOOutPin("participants", "on")
			participantsLED = true
			b.BackLightTimer()
		}
		if len(b.Client.Self.Channel.Users) == 1 && participantsLED {
			GPIOOutPin("participants", "off")
			participantsLED = false
			b.BackLightTimer()
		}
		if len(b.Client.Self.Channel.Users) != prevParticipantCount {
			var toSpeakEvent string = ""
			var eventSound = EventSoundStruct{}

			b.BackLightTimer()
			if Config.Global.Hardware.TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText[0] = "Online/RX" //b.Name
					LcdText[1] = "(" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + ")" + b.Client.Self.Channel.Name
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 0, OLEDStartColumn, "Online/RX") //b.Name
					oledDisplay(false, 1, OLEDStartColumn, "("+strconv.Itoa(len(b.Client.Self.Channel.Users))+")"+b.Client.Self.Channel.Name)
					oledDisplay(false, 6, OLEDStartColumn, "Please Visit")
					oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
				}
			}

			if e.Type.Has(2) {
				eventSound = findEventSound("leftchannel")
				if eventSound.Enabled {
					if v, err := strconv.Atoi(eventSound.Volume); err == nil {
						localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)

					}
				}
				toSpeakEvent = cleanstring(e.User.Name) + " Has Disconnected "
			}

			if e.Type.Has(128) {
				if len(b.Client.Self.Channel.Users) < prevParticipantCount {
					eventSound = findEventSound("leftchannel")
					if eventSound.Enabled {
						if v, err := strconv.Atoi(eventSound.Volume); err == nil {
							localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)

						}
					}
					toSpeakEvent = cleanstring(e.User.Name) + " Has Left Channel "
				} else {
					eventSound = findEventSound("joinedchannel")
					if eventSound.Enabled {
						if v, err := strconv.Atoi(eventSound.Volume); err == nil {
							localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)

						}
					}
					toSpeakEvent = cleanstring(e.User.Name) + " Has Joined Channel "
				}
			}

			for _, tts := range Config.Global.Software.TTS.Sound {
				if tts.Action == "leftjoinedchannel" {
					if tts.Enabled {
						if v, err := strconv.Atoi(eventSound.Volume); err == nil {
							localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)
						}
					}
				}

				if tts.Action == "participants" {
					if tts.Enabled {
						b.Speak(toSpeakEvent, "local", Config.Global.Software.TTS.Volumelevel, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
					}
				}
			}
			prevParticipantCount = len(b.Client.Self.Channel.Users)
		}

		if b.Client.Self.Channel.Name == e.User.Channel.Name {
			b.BackLightTimer()
			if e.User.Name != b.Client.Self.Name {
				log.Printf("info: This Channel %v User %v, type bin=%v, type char info=%v\n", e.User.Channel.Name, e.User.Name, e.Type, info)
				go joinedLeftScreen(e.User.Name, shortInfo)
				return
			}
		}
	}
}

func (b *Talkkonnect) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {

	switch e.Type {
	case gumble.PermissionDeniedPermission:
		log.Printf("warn: Permission Denied For Channel ID %v Channel Name %v\n", e.Channel.ID, e.Channel.Name)
	case gumble.PermissionDeniedSuperUser:
		log.Println("cannot modify SuperUser")
	case gumble.PermissionDeniedInvalidChannelName:
		log.Println("invalid channel name")
	case gumble.PermissionDeniedTextTooLong:
		log.Println("text too long")
	case gumble.PermissionDeniedTemporaryChannel:
		log.Println("temporary channel")
	case gumble.PermissionDeniedMissingCertificate:
		log.Println("missing certificate")
	case gumble.PermissionDeniedInvalidUserName:
		log.Println("invalid user name")
	case gumble.PermissionDeniedChannelFull:
		log.Println("channel full")
	case gumble.PermissionDeniedNestingLimit:
		log.Println("nesting limit")
	case gumble.PermissionDeniedOther:
		log.Println("other")
	}
}

func (b *Talkkonnect) OnChannelChange(e *gumble.ChannelChangeEvent) {
	if e.Channel.IsRoot() {
		RootChannel = e.Channel
	}
	e.Channel.RequestPermission()
	if e.Channel.Permission() != nil {
		if *e.Channel.Permission()&gumble.PermissionEnter != 0 {
			AccessableChannelMap[int(e.Channel.ID)] = e.Channel.Name
			if e.Channel.ID > TopChannelID {
				TopChannel = e.Channel
				TopChannelID = e.Channel.ID
			}
		}
	}
}

func (b *Talkkonnect) OnUserList(e *gumble.UserListEvent) {
	log.Println("alert: On User List Event Detected")
}

func (b *Talkkonnect) OnACL(e *gumble.ACLEvent) {
	log.Println("alert: On ACL Event Detected")
}

func (b *Talkkonnect) OnBanList(e *gumble.BanListEvent) {
	log.Println("alert: OnBanList Event Detected")
}

func (b *Talkkonnect) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	log.Println("alert: OnContextActionChange Event Detected")
}

func (b *Talkkonnect) OnServerConfig(e *gumble.ServerConfigEvent) {
	//placeholder
}
