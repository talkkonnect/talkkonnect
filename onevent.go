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

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	if IsConnected {
		GPIOOutPin("online", "on")
		MyLedStripOnlineLEDOn()
		return
	}

	IsConnected = true
	GPIOOutPin("online", "on")
	MyLedStripOnlineLEDOn()
	b.BackLightTimer()
	b.Client = e.Client
	ConnectAttempts = 1

	//serialize tokens send one by one from slice to server
	if len(Tokens[AccountIndex]) > 0 {
		ATokens := make(gumble.AccessTokens, len(Tokens[AccountIndex]))
		for i, value := range Tokens[AccountIndex] {
			ATokens[i] = value
		}

		b.Client.Send(ATokens)
	}

	log.Printf("debug: Connected to %s Address %s on attempt %d index [%d]\n ", b.Name, b.Client.Conn.RemoteAddr(), b.ConnectAttempts, AccountIndex)
	if e.WelcomeMessage != nil {
		var tmessage string = fmt.Sprintf("%v", esc(*e.WelcomeMessage))
		log.Println("info: Welcome message: ")
		for _, line := range strings.Split(strings.TrimSuffix(tmessage, "\n"), "\n") {
			log.Println("info: ", line)
		}
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {

		GPIOOutPin("online", "on")

		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 0, "") // clear the screen
		}

		b.ParticipantLEDUpdate(true)
	}

	if b.ChannelName != "" {
		b.ChangeChannel(b.ChannelName)
		prevChannelID = b.Client.Self.Channel.ID
	}
}

func (b *Talkkonnect) OnDisconnect(e *gumble.DisconnectEvent) {
	b.BackLightTimer()

	var reason string

	switch e.Type {
	case gumble.DisconnectError:
		reason = "connection error"
	}

	IsConnected = false
	GPIOOutPin("online", "off")
	MyLedStripOnlineLEDOff()
	if Config.Global.Hardware.TargetBoard == "rpi" {
		GPIOOutAll("led/relay", "off")
		MyLedStripGPIOOffAll()
	}

	log.Println("alert: Attempting Reconnect in 5 seconds...")
	log.Println("alert: Connection to ", b.Address, "disconnected")
	log.Println("alert: Disconnection Reason ", reason)

	time.Sleep(5 * time.Second)
	b.ReConnect()
}

func (b *Talkkonnect) OnTextMessage(e *gumble.TextMessageEvent) {
	b.BackLightTimer()

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

	log.Println(fmt.Sprintf("info: Message ("+strconv.Itoa(len(tmessage))+") from %v %v\n", sender, tmessage))

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
			oledDisplay(false, 2, 1, "Msg From "+sender)
			if len(tmessage) <= 21 {
				oledDisplay(false, 3, 1, tmessage)
				oledDisplay(false, 4, 1, "")
				oledDisplay(false, 5, 1, "")
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
				return
			}
			if len(tmessage) <= 42 {
				oledDisplay(false, 3, 1, tmessage[0:21])
				oledDisplay(false, 4, 1, tmessage[21:])
				oledDisplay(false, 5, 1, "")
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
				return
			}
			if len(tmessage) <= 63 {
				oledDisplay(false, 3, 1, tmessage[0:21])
				oledDisplay(false, 4, 1, tmessage[21:42])
				oledDisplay(false, 5, 1, tmessage[42:])
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
				return
			}
			if len(tmessage) <= 84 {
				oledDisplay(false, 3, 1, tmessage[0:21])
				oledDisplay(false, 4, 1, tmessage[21:42])
				oledDisplay(false, 5, 1, tmessage[42:63])
				oledDisplay(false, 6, 1, tmessage[63:])
				oledDisplay(false, 7, 1, "")
				return
			}
			if len(tmessage) <= 105 {
				oledDisplay(false, 3, 1, tmessage[0:20])
				oledDisplay(false, 4, 1, tmessage[21:44])
				oledDisplay(false, 5, 1, tmessage[42:63])
				oledDisplay(false, 6, 1, tmessage[63:84])
				oledDisplay(false, 7, 1, tmessage[84:])
				return
			}
		}
	}

}

func (b *Talkkonnect) OnUserChange(e *gumble.UserChangeEvent) {
	b.BackLightTimer()

	var info string

	switch e.Type {
	case gumble.UserChangeConnected:
		info = "conn"
	case gumble.UserChangeDisconnected:
		info = "disconnected!"
	case gumble.UserChangeKicked:
		info = "kicked"
	case gumble.UserChangeBanned:
		info = "banned"
	case gumble.UserChangeRegistered:
		info = "registered"
	case gumble.UserChangeUnregistered:
		info = "unregistered"
	case gumble.UserChangeName:
		info = "chg name"
	case gumble.UserChangeChannel:
		info = "chg channel"
		log.Println("info:", cleanstring(e.User.Name), " Changed Channel to ", e.User.Channel.Name)
		LcdText[2] = cleanstring(e.User.Name) + "->" + e.User.Channel.Name
		LcdText[3] = ""
		b.sevenSegment("mumblechannel", strconv.Itoa(int(b.Client.Self.Channel.ID)))
	case gumble.UserChangeComment:
		info = "chg comment"
	case gumble.UserChangeAudio:
		info = "chg audio"
	case gumble.UserChangePrioritySpeaker:
		info = "is priority"
	case gumble.UserChangeRecording:
		info = "chg rec status"
	case gumble.UserChangeStats:
		info = "chg stats"
		if info != "chg channel" {

			if info != "" {
				log.Println("info: User ", cleanstring(e.User.Name), " ", info, "Event type=", e.Type, " channel=", e.User.Channel.Name)

				for _, tts := range Config.Global.Software.TTS.Sound {
					if tts.Action == "participants" {
						if tts.Enabled {
							b.Speak("User "+cleanstring(e.User.Name)+info+"Has Changed to "+e.User.Channel.Name, "local", 1, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
						}
					}
				}
			}

		} else {
			log.Println("info: User ", cleanstring(e.User.Name), " Event type=", e.Type, " channel=", e.User.Channel.Name)
		}

		LcdText[2] = cleanstring(e.User.Name) + " " + info //+strconv.Atoi(string(e.Type))

	}

	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	var info string

	switch e.Type {
	case gumble.PermissionDeniedOther:
		info = e.String
	case gumble.PermissionDeniedPermission:
		info = "insufficient permissions"
		LcdText[2] = "insufficient perms"

		// Set Upper Boundary
		if prevButtonPress == "ChannelUp" && b.Client.Self.Channel.ID == maxchannelid {
			log.Println("error: Can't Increment Channel Maximum Channel Reached")
		}

		// Set Lower Boundary
		if prevButtonPress == "ChannelDown" && currentChannelID == 0 {
			log.Println("error: Can't Increment Channel Minimum Channel Reached")
		}

		// Implement Seek Up Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelUp" && b.Client.Self.Channel.ID+1 < maxchannelid {
			prevChannelID++
			b.ChannelUp()
			LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
		}

		// Implement Seek Dwn Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelDown" && int(b.Client.Self.Channel.ID) > 0 {
			prevChannelID--
			b.ChannelDown()
			LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
		}

		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 1, 1, LcdText[1])
				oledDisplay(false, 2, 1, LcdText[2])
			}
		}

	case gumble.PermissionDeniedSuperUser:
		info = "cannot modify SuperUser"
	case gumble.PermissionDeniedInvalidChannelName:
		info = "invalid channel name"
	case gumble.PermissionDeniedTextTooLong:
		info = "text too long"
	case gumble.PermissionDeniedTemporaryChannel:
		info = "temporary channel"
	case gumble.PermissionDeniedMissingCertificate:
		info = "missing certificate"
	case gumble.PermissionDeniedInvalidUserName:
		info = "invalid user name"
	case gumble.PermissionDeniedChannelFull:
		info = "channel full"
	case gumble.PermissionDeniedNestingLimit:
		info = "nesting limit"
	}

	LcdText[2] = info

	log.Printf("error: Permission denied %v to Join Channel %v\n", info, e.Channel.Name)
}

func (b *Talkkonnect) OnChannelChange(e *gumble.ChannelChangeEvent) {
	log.Println("debug: Channel Change Event Detected")
}

func (b *Talkkonnect) OnUserList(e *gumble.UserListEvent) {
	log.Println("debug: On User List Event Detected")
}

func (b *Talkkonnect) OnACL(e *gumble.ACLEvent) {
	log.Println("debug: On ACL Event Detected")
}

func (b *Talkkonnect) OnBanList(e *gumble.BanListEvent) {
	log.Println("debug: OnBanList Event Detected")
}

func (b *Talkkonnect) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
	log.Println("debug: OnContextActionChange Event Detected")
}

func (b *Talkkonnect) OnServerConfig(e *gumble.ServerConfigEvent) {
	log.Println("debug: OnServerConfigs Event Detected")
}
