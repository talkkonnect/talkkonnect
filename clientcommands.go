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
	"strconv"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	htgotts "github.com/talkkonnect/htgo-tts"
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

func (b *Talkkonnect) CleanUp() {

	if TargetBoard == "rpi" {
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
		if !LedStripEnabled {
			b.LEDOffAll()
		} else {
			MyLedStripLEDOffAll()
		}
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
		if !ServerHop {
			log.Println("debug: In the Connect Function & Trying With Username ", Username)
			log.Println("debug: DEBUG Serverhop  Not Set Reconnecting!!")
			b.ReConnect()
		}
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
		if !ServerHop {
			ConnectAttempts++
			b.Connect()
		}
	} else {
		if TargetBoard == "rpi" {
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
	t := time.Now()

	if SimplexWithMute {
		err := volume.Mute(OutputDevice)
		if err != nil {
			log.Println("error: Unable to Mute ", err)
		} else {
			log.Println("info: Speaker Muted ")
		}
	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false
		time.Sleep(100 * time.Millisecond)
		b.playIntoStream(StreamSoundFilenameAndPath, StreamSoundVolume)
	}

	if TargetBoard == "rpi" {
		if !LedStripEnabled {
			b.LEDOn(b.TransmitLED)
		} else {
			MyLedStripTransmitLEDOn()
		}
		if LCDEnabled {
			LcdText[0] = "Online/TX"
			LcdText[3] = "TX at " + t.Format("15:04:05")
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(false, 0, 1, "Online/TX")
			oledDisplay(false, 3, 1, "TX at "+t.Format("15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit       ")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
	}

	b.IsTransmitting = true

	if pstream.State() == gumbleffmpeg.StatePlaying {
		pstream.Stop()
	}

	b.Stream.StartSource()

}

func (b *Talkkonnect) TransmitStop(withBeep bool) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	if TargetBoard == "rpi" {
		if !LedStripEnabled {
			b.LEDOff(b.TransmitLED)
		} else {
			MyLedStripTransmitLEDOff()
		}
		if LCDEnabled {
			LcdText[0] = b.Address
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 0, 1, b.Address)
		}
	}

	b.IsTransmitting = false
	b.Stream.StopSource()

	if SimplexWithMute {
		err := volume.Unmute(OutputDevice)
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

		if TargetBoard == "rpi" {
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

	if EventSoundEnabled {
		if participantCount > prevParticipantCount {
			err := playWavLocal(EventJoinedSoundFilenameAndPath, 100)
			if err != nil {
				log.Println("error: playWavLocal(EventJoinedSoundFilenameAndPath) Returned Error: ", err)
			}
		}
		if participantCount < prevParticipantCount {
			err := playWavLocal(EventLeftSoundFilenameAndPath, 100)
			if err != nil {
				log.Println("error: playWavLocal(EventLeftSoundFilenameAndPath) Returned Error: ", err)
			}
		}
	}

	if participantCount > 1 && participantCount != prevParticipantCount {

		if TTSEnabled && TTSParticipants {
			speech := htgotts.Speech{Folder: "audio", Language: "en"}
			speech.Speak("There Are Currently " + strconv.Itoa(participantCount) + " Users in The Channel " + b.Client.Self.Channel.Name)
		}

		prevParticipantCount = participantCount

		if verbose {
			log.Println("info: Current Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")
			b.ListUsers()
			if TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText[0] = b.Address
					LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 0, 1, b.Address)
					oledDisplay(false, 1, 1, b.Client.Self.Channel.Name+" ("+strconv.Itoa(participantCount)+" Users)")
					oledDisplay(false, 6, 1, "Please Visit")
					oledDisplay(false, 7, 1, "www.talkkonnect.com")
				}

			}
		}
	}

	if participantCount > 1 {
		if TargetBoard == "rpi" {
			if !LedStripEnabled {
				b.LEDOn(b.ParticipantsLED)
				b.LEDOn(b.OnlineLED)
			} else {
				MyLedStripParticipantsLEDOn()
				MyLedStripOnlineLEDOn()
			}
		}

	} else {

		if verbose {
			if TTSEnabled && TTSParticipants {
				speech := htgotts.Speech{Folder: "audio", Language: "en"}
				speech.Speak("You are Currently Alone in The Channel " + b.Client.Self.Channel.Name)
			}
			log.Println("info: Channel ", b.Client.Self.Channel.Name, " has no other participants")

			prevParticipantCount = 0

			if TargetBoard == "rpi" {
				if !LedStripEnabled {
					b.LEDOff(b.ParticipantsLED)
				} else {
					MyLedStripParticipantsLEDOff()
				}
				if LCDEnabled {
					LcdText = [4]string{b.Address, "Alone in " + b.Client.Self.Channel.Name, "", "nil"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 0, 1, b.Address)
					oledDisplay(false, 1, 1, "Alone in "+b.Client.Self.Channel.Name)
				}
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

	var records = int(len(b.Client.Channels))
	channelsList := make([]ChannelsListStruct, len(b.Client.Channels))
	counter := 0

	for _, ch := range b.Client.Channels {
		channelsList[counter].chanID = ch.ID
		channelsList[counter].chanName = ch.Name
		channelsList[counter].chanParent = ch.Parent
		channelsList[counter].chanUsers = len(ch.Users)

		if ch.ID > maxchannelid {
			maxchannelid = ch.ID
		}

		counter++
	}

	for i := 0; i < int(records); i++ {
		if channelsList[i].chanID == 0 || channelsList[i].chanParent.ID == 0 {
			if verbose {
				log.Println(fmt.Sprintf("info: Parent -> ID=%2d | Name=%-12v (%v) Users | ", channelsList[i].chanID, channelsList[i].chanName, channelsList[i].chanUsers))
			}
		} else {
			if verbose {
				log.Println(fmt.Sprintf("info: Child  -> ID=%2d | Name=%-12v (%v) Users | PID =%2d | PName=%-12s", channelsList[i].chanID, channelsList[i].chanName, channelsList[i].chanUsers, channelsList[i].chanParent.ID, channelsList[i].chanParent.Name))
			}
		}
	}

}

func (b *Talkkonnect) ChannelUp() {
	if !(IsConnected) {
		return
	}

	if prevChannelID == 0 {
		prevChannelID = b.Client.Self.Channel.ID
	}

	if TTSEnabled && TTSChannelUp {
		err := playWavLocal(TTSChannelUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelUp"

	b.ListChannels(false)

	// Set Upper Boundary
	if b.Client.Self.Channel.ID == maxchannelid {
		log.Println("error: Can't Increment Channel Maximum Channel Reached")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText[2] = "Max Chan Reached"
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 1, 1, "Max Chan Reached")
			}

		}
		return
	}

	// Implement Seek Up Avoiding any null channels
	if prevChannelID < maxchannelid {

		prevChannelID++

		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {

			channel := b.Client.Channels[i]

			if channel != nil {
				b.Client.Self.Move(channel)
				//displaychannel
				time.Sleep(500 * time.Millisecond)
				if TargetBoard == "rpi" {

					if len(b.Client.Self.Channel.Users) == 1 {
						LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
					} else {
						LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
					}

					if LCDEnabled {
						LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}
				break
			}
		}
	}
}

func (b *Talkkonnect) ChannelDown() {
	if !(IsConnected) {
		return
	}

	if prevChannelID == 0 {
		prevChannelID = b.Client.Self.Channel.ID
	}

	if TTSEnabled && TTSChannelDown {
		err := playWavLocal(TTSChannelDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelDown"
	b.ListChannels(false)

	// Set Lower Boundary
	if int(b.Client.Self.Channel.ID) == 0 {
		log.Println("error: Can't Decrement Channel Root Channel Reached")
		channel := b.Client.Channels[uint32(AccountIndex)]
		b.Client.Self.Move(channel)
		//displaychannel
		time.Sleep(500 * time.Millisecond)
		if TargetBoard == "rpi" {

			if len(b.Client.Self.Channel.Users) == 1 {
				LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
			} else {
				LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
			}

			if LCDEnabled {
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 1, 1, LcdText[1])
			}
		}

		return
	}

	// Implement Seek Down Avoiding any null channels
	if int(prevChannelID) > 0 {

		prevChannelID--

		for i := uint32(prevChannelID); uint32(i) < maxchannelid; i-- {
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				//displaychannel
				time.Sleep(500 * time.Millisecond)
				if TargetBoard == "rpi" {

					if len(b.Client.Self.Channel.Users) == 1 {
						LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
					} else {
						LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
					}

					if LCDEnabled {
						LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}

				break
			}
		}
	}
}

func (b *Talkkonnect) Scan() {
	if !(IsConnected) {
		return
	}

	b.ListChannels(false)

	if b.Client.Self.Channel.ID+1 > maxchannelid {
		prevChannelID = 0
		channel := b.Client.Channels[prevChannelID]
		b.Client.Self.Move(channel)
		return
	}

	if prevChannelID < maxchannelid {
		prevChannelID++

		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				time.Sleep(1000 * time.Millisecond)
				if len(b.Client.Self.Channel.Users) == 1 {
					b.Scan()
					break
				} else {

					log.Println("info: Found Someone Online Stopped Scan on Channel ", b.Client.Self.Channel.Name)
					return
				}
			}
		}
	}
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
		if TargetBoard == "rpi" {
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

	if TargetBoard != "rpi" || (!LCDBackLightTimerEnabled && !OLEDEnabled && !LCDEnabled) {
		return
	}

	if LCDEnabled {
		b.LEDOn(b.BackLightLED)
	}

	if OLEDEnabled {
		Oled.DisplayOn()
	}

	BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)
}

func (b *Talkkonnect) TxLockTimer() {
	if PTxLockEnabled {
		TxLockTicker := time.NewTicker(time.Duration(PTxlockTimeOutSecs) * time.Second)
		log.Println("info: TX Locked for ", PTxlockTimeOutSecs, " seconds")
		b.TransmitStop(false)
		b.TransmitStart()

		go func() {
			<-TxLockTicker.C
			b.TransmitStop(true)
			log.Println("info: TX UnLocked After ", PTxlockTimeOutSecs, " seconds")
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
	for i := 0; i < 100; i++ {
		b.TransmitStart()
		b.IsTransmitting = true
		time.Sleep(1 * time.Second)
		b.TransmitStop(true)
		b.IsTransmitting = false
		time.Sleep(1 * time.Second)
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
	for _, account := range Document.Accounts.Account {
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

func (b *Talkkonnect) VoiceTargetUserSet(targetID uint32, targetUser string) {
	vtUser := b.Client.Users.Find(targetUser)
	if (vtUser != nil) && (targetID <= 31) {
		vtarget := &gumble.VoiceTarget{}
		vtarget.ID = targetID
		vtarget.AddUser(vtUser)
		b.Client.VoiceTarget = vtarget
		b.Client.Send(vtarget)
		log.Printf("debug: Added User %v to VT ID %v\n", targetUser, targetID)
	} else {
		log.Printf("error: Cannot Add User %v to VT ID %v\n", targetUser, targetID)
	}

}

func (b *Talkkonnect) VoiceTargetChannelSet(targetID uint32, targetChannel string, recursive bool, links bool, group string) {
	vtChannel := b.Client.Self.Channel
	if vtChannel != nil {
		vtarget := &gumble.VoiceTarget{}
		vtarget.ID = targetID
		vtarget.AddChannel(vtChannel, recursive, links, group)
		b.Client.VoiceTarget = vtarget
		b.Client.Send(vtarget)
		log.Printf("debug: Shouting to Channel %v to VT ID %v with recursive %v links %v group %v\n", vtChannel.Name, targetID, recursive, links, group)
	} else {
		log.Printf("error: Target Channel %v Not Found\n", targetChannel)
	}
}
