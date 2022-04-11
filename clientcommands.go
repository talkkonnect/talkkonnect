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
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
)

func FatalCleanUp(message string) {
	term.Close()
	fmt.Println(message)
	time.Sleep(5 * time.Second)
	fmt.Println("Talkkonnect Terminated Abnormally with the Error(s) As Described Perviously, Ignore any GPIO errors if you are not using Single Board Computer.")
	os.Exit(1)
}

func CleanUp() {

	if Config.Global.Hardware.TargetBoard == "rpi" {
		t := time.Now()
		if LCDEnabled {
			LcdText = [4]string{"talkkonnect stopped", t.Format("02-01-2006 15:04:05"), "Please Visit", "www.talkkonnect.com"}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 1, "talkkonnect stopped")
			oledDisplay(false, 1, 1, t.Format("02-01-2006 15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
		GPIOOutAll("led/relay", "off")
		MyLedStripGPIOOffAll()
	}

	term.Close()
	fmt.Println("SIGHUP Termination of Program Requested by User...shutting down talkkonnect")
	os.Exit(0)
}

func (b *Talkkonnect) Connect() {
	IsConnected = false
	IsPlayStream = false
	NowStreaming = false
	KillHeartBeat = false
	var err error

	_, err = gumble.DialWithDialer(new(net.Dialer), b.Address, b.Config, &b.TLSConfig)

	if err != nil {
		log.Printf("error: Connection Error %v  connecting to %v failed, attempting again...", err, b.Address)
		log.Println("debug: In the Connect Function & Trying With Username ", Username)
		b.ReConnect()
	} else {
		b.OpenStream()
	}
}

func (b *Talkkonnect) ReConnect() {
	IsConnected = false
	IsPlayStream = false
	NowStreaming = false

	if b.Client != nil {
		log.Println("info: Attempting Reconnection With Server")
		b.Client.Disconnect()
	}

	if ConnectAttempts < 3 {
		ConnectAttempts++
		b.Connect()
	} else {
		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"Failed to Connect!", "nil", "nil", "nil"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 2, 1, "Failed to Connect!")
			}
		}
		FatalCleanUp("Unable to Connect to mumble server, Giving Up!")
	}
}

func (b *Talkkonnect) TransmitStart() {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()
	LastSpeaker = ""
	if Config.Global.Software.Settings.SimplexWithMute {
		err := volume.Mute(Config.Global.Software.Settings.OutputDevice)
		if err != nil {
			log.Println("error: Unable to Mute ", err)
		} else {
			log.Println("info: Speaker Muted ")
		}
	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		for _, sound := range Config.Global.Software.Sounds.Sound {
			if sound.Enabled {
				if sound.Event == "stream" {
					if s, err := strconv.ParseFloat(sound.Volume, 32); err == nil {
						b.playIntoStream(sound.File, float32(s))
					}
				}
			}
		}
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {
		// use groutine so no need to wait for local screen cause it causes delay
		go GPIOOutPin("transmit", "on")
		go MyLedStripTransmitLEDOn()
		go txScreen()
	}

	b.IsTransmitting = true

	if pstream.State() == gumbleffmpeg.StatePlaying {
		pstream.Stop()
	}

	b.StartSource()

}

func (b *Talkkonnect) TransmitStop(withBeep bool) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	if Config.Global.Hardware.TargetBoard == "rpi" {
		GPIOOutPin("transmit", "off")
		MyLedStripTransmitLEDOff()
		if LCDEnabled {
			LcdText[0] = b.Name // b.Address
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 0, 1, b.Name) //b.Address
		}
	}

	b.IsTransmitting = false
	b.StopSource()

	if Config.Global.Software.Settings.SimplexWithMute {
		err := volume.Unmute(Config.Global.Software.Settings.OutputDevice)
		if err != nil {
			log.Println("error: Unable to Unmute ", err)
		} else {
			log.Println("info: Speaker UnMuted ")
		}
	}
}

func (b *Talkkonnect) ChangeChannel(ChannelName string) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	channel := b.Client.Channels.Find(ChannelName)
	if channel != nil {

		b.Client.Self.Move(channel)

		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText[1] = "Joined " + ChannelName
				LcdText[2] = Username[AccountIndex]
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 0, 1, "Joined "+ChannelName)
				oledDisplay(false, 1, 1, Username[AccountIndex])
			}
		}

		log.Println("info: Joined Channel Name: ", channel.Name, " ID ", channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		log.Println("warn: Unable to Find Channel Name: ", ChannelName)
		prevChannelID = 0
	}
}

func (b *Talkkonnect) ParticipantLEDUpdate(verbose bool) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	var participantCount = len(b.Client.Self.Channel.Users)

	var eventSound = EventSoundStruct{}

	eventSound = findEventSound("joinedchannel")
	if eventSound.Enabled {
		if participantCount > prevParticipantCount {
			if v, err := strconv.Atoi(eventSound.Volume); err == nil {
				localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)
			}
		}
	}
	eventSound = findEventSound("leftchannel")
	if eventSound.Enabled {
		if participantCount < prevParticipantCount {
			if v, err := strconv.Atoi(eventSound.Volume); err == nil {
				localMediaPlayer(eventSound.FileName, v, eventSound.Blocking, 0, 1)
			}
		}
	}

	if participantCount > 1 {
		for _, tts := range Config.Global.Software.TTS.Sound {
			if tts.Action == "participants" {
				if tts.Enabled {
					tempStatus := Config.Global.Software.TTSMessages.TTSTone.ToneEnabled
					Config.Global.Software.TTSMessages.TTSTone.ToneEnabled = false
					b.Speak("There Are Currently "+strconv.Itoa(participantCount)+" Users in The Channel "+b.Client.Self.Channel.Name, "local", 1, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
					Config.Global.Software.TTSMessages.TTSTone.ToneEnabled = tempStatus
				}
			}
		}
	}
	prevParticipantCount = len(b.Client.Self.Channel.Users)

	if verbose {
		log.Println("info: Current Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")
		b.ListUsers()
		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText[0] = b.Name //b.Address
				LcdText[1] = "(" + strconv.Itoa(participantCount) + ")" + b.Client.Self.Channel.Name
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 0, 1, b.Name) //b.Address
				oledDisplay(false, 1, 1, "("+strconv.Itoa(participantCount)+")"+b.Client.Self.Channel.Name)
				oledDisplay(false, 6, 1, "Please Visit")
				oledDisplay(false, 7, 1, "www.talkkonnect.com")
			}

		}
	}

	if participantCount > 1 {
		if Config.Global.Hardware.TargetBoard == "rpi" {
			GPIOOutPin("participants", "on")
		}
	} else {

		if verbose {
			for _, tts := range Config.Global.Software.TTS.Sound {
				if tts.Action == "participants" {
					if tts.Enabled {
						b.Speak("You are Currently Alone in The Channel "+b.Client.Self.Channel.Name, "local", 1, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
					}
				}
			}

			log.Println("info: Channel ", b.Client.Self.Channel.Name, " has no other participants")
		}

		prevParticipantCount = len(b.Client.Self.Channel.Users)

		if Config.Global.Hardware.TargetBoard == "rpi" {
			GPIOOutPin("participants", "off")
			if LCDEnabled && verbose {
				LcdText = [4]string{b.Name, "(0)" + b.Client.Self.Channel.Name, "", "nil"} //b.Address
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled && verbose {
				oledDisplay(false, 0, 1, b.Name) //b.Address
				oledDisplay(false, 1, 1, "(0)"+b.Client.Self.Channel.Name)
			}
		}
	}
}

func (b *Talkkonnect) ListUsers() {
	if !(IsConnected) {
		return
	}

	item := 0
	for _, usr := range b.Client.Users {
		if usr.Channel.ID == b.Client.Self.Channel.ID {
			item++
			log.Println(fmt.Sprintf("info: %d. User %#v is online. [%v]", item, usr.Name, usr.Comment))
		}
	}
}

func (b *Talkkonnect) ListChannels(verbose bool) {

	if !(IsConnected) {
		return
	}

	ChannelsList = make([]ChannelsListStruct, len(b.Client.Channels))
	counter := 0
	ChannelIDs := []int{}

	for _, ch := range b.Client.Channels {
		ChannelIDs = append(ChannelIDs, int(ch.ID))
		counter++
	}

	sort.Ints(ChannelIDs)

	counter = 0
	for _, cid := range ChannelIDs {
		ChannelsList[counter].chanIndex = counter
		ChannelsList[counter].chanID = cid
		b.findChannelDetailsByID(uint32(cid), counter)
		counter++
	}
	if verbose {
		for i := 0; i < len(b.Client.Channels); i++ {
			log.Println("debug: ", ChannelsList[i])
		}
	}
}

func (b *Talkkonnect) ChannelUp() {
	if !(IsConnected) {
		return
	}
	ChannelAction = "channelup"
	TTSEvent("channelup")
	Channel := b.Client.Channels.Find()
	currentIndex := b.findChannelIndex(b.Client.Self.Channel.ID)
	NextIndex := currentIndex

	if currentIndex+1 < len(b.Client.Channels) {
		Channel = b.Client.Channels.Find(ChannelsList[currentIndex+1].chanName)
		NextIndex = currentIndex + 1
	}

	if ChannelsList[NextIndex].chanenterPermissions {
		if Channel != nil {
			b.Client.Self.Move(Channel)
		}
	} else {
		for i := NextIndex + 1; i <= len(b.Client.Channels); i++ {
			//special handling for when highest channel has no token
			if i == len(b.Client.Channels) {
				Channel = b.Client.Channels.Find()
				b.Client.Self.Move(Channel)
				return
			} else {
				Channel = b.Client.Channels.Find(ChannelsList[i].chanName)
			}
			if ChannelsList[i].chanenterPermissions {
				b.Client.Self.Move(Channel)
				break
			}
		}
	}
}

func (b *Talkkonnect) ChannelDown() {
	if !(IsConnected) {
		return
	}
	ChannelAction = "channeldown"
	TTSEvent("channeldown")
	Channel := b.Client.Channels.Find(ChannelsList[len(b.Client.Channels)-1].chanName)
	currentIndex := b.findChannelIndex(b.Client.Self.Channel.ID)
	NextIndex := currentIndex

	if currentIndex == 0 {
		//special handling of max channel without token
		Channel = b.Client.Channels.Find(ChannelsList[len(b.Client.Channels)-1].chanName)
		NextIndex = len(b.Client.Channels) - 1
	}

	if currentIndex == 1 {
		Channel = b.Client.Channels.Find()
		NextIndex = currentIndex - 1
	}

	if currentIndex > 1 {
		Channel = b.Client.Channels.Find(ChannelsList[currentIndex-1].chanName)
		NextIndex = currentIndex - 1
	}

	if ChannelsList[NextIndex].chanenterPermissions {
		if Channel != nil {
			b.Client.Self.Move(Channel)
		}
	} else {
		for i := NextIndex - 1; i >= 0; i-- {
			Channel = b.Client.Channels.Find(ChannelsList[i].chanName)
			if ChannelsList[i].chanenterPermissions {
				b.Client.Self.Move(Channel)
				break
			}
		}
	}
}

func (b *Talkkonnect) Scan() {
	if !(IsConnected) {
		return
	}

	log.Println("alert: New Scan Not Implemented Yet")
}

func (b *Talkkonnect) SendMessage(textmessage string, PRecursive bool) {
	if !(IsConnected) {
		return
	}
	b.Client.Self.Channel.Send(textmessage, PRecursive)
}

func (b *Talkkonnect) SetComment(comment string) {
	if IsConnected {
		b.BackLightTimer()
		b.Client.Self.SetComment(comment)
		t := time.Now()
		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText[2] = "Status at " + t.Format("15:04:05")
				time.Sleep(500 * time.Millisecond)
				LcdText[3] = b.Client.Self.Comment
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 1, 1, "Status at "+t.Format("15:04:05"))
				oledDisplay(false, 4, 1, b.Client.Self.Comment)
			}
		}
	}
}

func (b *Talkkonnect) BackLightTimer() {
	BackLightTime = *BackLightTimePtr

	if Config.Global.Hardware.TargetBoard != "rpi" || (!LCDBackLightTimerEnabled && !OLEDEnabled && !LCDEnabled) {
		return
	}

	if LCDEnabled {
		GPIOOutPin("backlight", "on")
	}

	if OLEDEnabled {
		Oled.DisplayOn()
	}

	BackLightTime.Reset(time.Duration(LCDBackLightTimeout) * time.Second)
}

func (b *Talkkonnect) TxLockTimer() {
	if Config.Global.Hardware.PanicFunction.TxLockEnabled {
		TxLockTicker := time.NewTicker(time.Duration(Config.Global.Hardware.PanicFunction.TxLockTimeOutSecs) * time.Second)
		log.Println("info: TX Locked for ", Config.Global.Hardware.PanicFunction.TxLockTimeOutSecs, " seconds")
		b.TransmitStop(false)
		b.TransmitStart()

		go func() {
			<-TxLockTicker.C
			b.TransmitStop(true)
			log.Println("info: TX UnLocked After ", Config.Global.Hardware.PanicFunction.TxLockTimeOutSecs, " seconds")
		}()
	}
}

func (b *Talkkonnect) pingServers() {
	currentconn := " Not Connected "
	for i := 0; i < len(Server); i++ {
		resp, err := gumble.Ping(Server[i], time.Second*1, time.Second*5)

		if b.Address == Server[i] {
			currentconn = " ** Connected ** "
		} else {
			currentconn = ""
		}

		log.Println("info: Server # ", i+1, "["+Name[i]+"]"+currentconn)

		if err != nil {
			log.Println(fmt.Sprintf("error: Ping Error %q", err))
			continue
		}

		major, minor, patch := resp.Version.SemanticVersion()

		log.Println("info: Server Address:         ", resp.Address)
		log.Println("info: Server Ping:            ", resp.Ping)
		log.Println("info: Server Version:         ", major, ".", minor, ".", patch)
		log.Println("info: Server Users:           ", resp.ConnectedUsers, "/", resp.MaximumUsers)
		log.Println("info: Server Maximum Bitrate: ", resp.MaximumBitrate)
	}
}

func (b *Talkkonnect) repeatTx() {
	if Config.Global.Software.Settings.RepeatTXTimes == 0 || Config.Global.Software.Settings.RepeatTXDelay == 0 {
		return
	}
	for i := 0; i < Config.Global.Software.Settings.RepeatTXTimes; i++ {
		b.TransmitStart()
		b.IsTransmitting = true
		time.Sleep(Config.Global.Software.Settings.RepeatTXDelay * time.Second)
		b.TransmitStop(true)
		b.IsTransmitting = false
		time.Sleep(Config.Global.Software.Settings.RepeatTXDelay * time.Second)
		if i > 0 {
			log.Println("info: TX Cycle ", i)
			if isrepeattx {
				log.Println("info: Repeat Tx Loop Text Forcefully Stopped")
			}
		}

		if isrepeattx {
			break
		}
	}
}

func (b *Talkkonnect) cmdSendVoiceTargets(targetID uint32) {

	GenericCounter = 0
	for _, account := range Config.Accounts.Account {
		if account.Default {
			for _, vtvalue := range account.Voicetargets.ID {

				if GenericCounter == AccountIndex {

					if vtvalue.Value == targetID {
						log.Println("debug: Account Index ", GenericCounter, vtvalue)
						log.Printf("debug: User Requested VT-ID %v\n", vtvalue.Value)

						for _, vtuser := range vtvalue.Users.User {
							b.VoiceTargetUserSet(vtvalue.Value, vtuser)
						}

						for _, vtchannel := range vtvalue.Channels.Channel {
							b.VoiceTargetChannelSet(vtvalue.Value, vtchannel.Name, vtchannel.Recursive, vtchannel.Links, vtchannel.Group)
						}
					}
				}
			}
			GenericCounter++
		}
	}
}

func (b *Talkkonnect) VoiceTargetUserSet(TargetID uint32, TargetUser string) {
	if len(TargetUser) == 0 && TargetID == 0 {
		TargetUser = b.Client.Self.Name
	}

	vtUser := b.Client.Users.Find(TargetUser)
	if (vtUser != nil) && (TargetID <= 31) {
		vtarget := &gumble.VoiceTarget{}
		vtarget.ID = TargetID
		vtarget.AddUser(vtUser)
		b.Client.VoiceTarget = vtarget
		if TargetID > 0 {
			log.Printf("debug: Added User %v to VT ID %v\n", TargetUser, TargetID)
			b.sevenSegment("voicetarget", strconv.Itoa(int(TargetID)))
			GPIOOutPin("voicetarget", "on")
		} else {
			//b.VoiceTarget.Clear()
			GPIOOutPin("voicetarget", "off")
			log.Println("debug: Cleared Voice Targets")
			b.sevenSegment("voicetarget", strconv.Itoa(int(TargetID)))
		}
		b.Client.Send(vtarget)
	} else {
		log.Printf("error: Cannot Add User %v to VT ID %v\n", TargetUser, TargetID)
	}

}

func (b *Talkkonnect) VoiceTargetChannelSet(targetID uint32, targetChannelName string, recursive bool, links bool, group string) {
	if len(targetChannelName) == 0 {
		return
	}

	vtarget := &gumble.VoiceTarget{}
	vtarget.ID = targetID
	vChannel := b.Client.Channels.Find(targetChannelName)

	//find root channel name workarround
	var RootChannelName string
	var RootChannel *gumble.Channel
	for _, v := range b.Client.Channels {
		if v.ID == 0 {
			RootChannelName = v.Name
			RootChannel = v
		}
	}

	if targetChannelName == RootChannelName {
		vChannel = RootChannel
		vtarget.AddChannel(vChannel, recursive, links, group)
		b.Client.VoiceTarget = vtarget
		b.Client.Send(vtarget)
		log.Printf("debug: Shouting to Root Channel %v to VT ID %v with recursive %v links %v group %v\n", vChannel.Name, targetID, recursive, links, group)
		GPIOOutPin("voicetarget", "off")
		b.sevenSegment("voicetarget", strconv.Itoa(int(targetID)))
		return
	}

	if vChannel == nil {
		log.Printf("error: Child Channel %v Not Found!\n", targetChannelName)
		return
	}

	vtarget.AddChannel(vChannel, recursive, links, group)
	b.Client.VoiceTarget = vtarget
	b.Client.Send(vtarget)
	log.Printf("debug: Shouting to Child Channel %v to VT ID %v with recursive %v links %v group %v\n", vChannel.Name, targetID, recursive, links, group)
	b.sevenSegment("voicetarget", strconv.Itoa(int(targetID)))
	if targetID > 0 {
		GPIOOutPin("voicetarget", "on")
	}
}

func (b *Talkkonnect) findChannelIndex(currentChannelID uint32) int {
	index := 0
	for _, ch := range ChannelsList {
		if ch.chanID == int(currentChannelID) {
			return index
		}
		index++
	}
	return 0
}

func (b *Talkkonnect) findChannelDetailsByID(ChannelID uint32, index int) {
	for _, ch := range b.Client.Channels {
		if ch.ID == ChannelID {
			ChannelsList[index].chanName = ch.Name
			ChannelsList[index].chanParent = ch.Parent
			ChannelsList[index].chanUsers = ch.Users
			ChannelsList[index].chanenterPermissions = true
		}
	}
}
