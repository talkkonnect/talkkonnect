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
	"github.com/kennygrant/sanitize"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	_ "github.com/talkkonnect/gumble/opus"
	htgotts "github.com/talkkonnect/htgo-tts"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

func reset() {
	term.Sync()
}

func esc(str string) string {
	return sanitize.HTML(str)
}

func cleanstring(str string) string {
	return sanitize.Name(str)
}


func (b *Talkkonnect) CleanUp() {
	log.Println("warn: SIGHUP Termination of Program Requested...shutting down...bye!")

	if TargetBoard == "rpi" {
		t := time.Now()
		if LCDEnabled == true {
			LcdText = [4]string{"talkkonnect stopped", t.Format("02-01-2006 15:04:05"), "Please Visit", "www.talkkonnect.com"}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 1, "talkkonnect stopped")
			oledDisplay(false, 1, 1, t.Format("02-01-2006 15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
		b.LEDOffAll()
	}

	//b.Client.Disconnect()
	c := exec.Command("reset")
	c.Stdout = os.Stdout
	c.Run()
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
		//go func() {
		if !ServerHop {
			ConnectAttempts++
			b.Connect()
		}
		//}()
	} else {
		log.Println("alert: Unable to connect, giving up")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"Failed to Connect!", "nil", "nil", "nil"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 2, 1, "Failed to Connect!")
			}
		}
		log.Fatal("Exiting talkkonnect! ...... bye!\n")
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
		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if TargetBoard == "rpi" {
		b.LEDOn(b.TransmitLED)
		if LCDEnabled == true {
			LcdText[0] = "Online/TX"
			LcdText[3] = "TX at " + t.Format("15:04:05")
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(false, 0, 1, "Online/TX")
			oledDisplay(false, 3, 1, "TX at "+t.Format("15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit       ")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
	}

	b.IsTransmitting = true

	if RepeaterToneEnabled {
		b.RepeaterTone(RepeaterToneFilenameAndPath, RepeaterToneVolume)
	}

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
		b.LEDOff(b.TransmitLED)

		if LCDEnabled == true {
			LcdText[0] = b.Address
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
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
			if LCDEnabled == true {
				LcdText[1] = "Joined " + ChannelName
				LcdText[2] = Username[AccountIndex]
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
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

	if participantCount != prevParticipantCount {
		if EventSoundEnabled {
			err := PlayWavLocal(EventSoundFilenameAndPath, 100)
			if err != nil {
				log.Println("error: PlayWavLocal(EventSoundFilenameAndPath) Returned Error: ", err)
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
				if LCDEnabled == true {
					LcdText[0] = b.Address
					LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
					go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
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
			b.LEDOn(b.ParticipantsLED)
			b.LEDOn(b.OnlineLED)
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

				b.LEDOff(b.ParticipantsLED)

				if LCDEnabled == true {
					LcdText = [4]string{b.Address, "Alone in " + b.Client.Self.Channel.Name, "", "nil"}
					go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
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
		err := PlayWavLocal(TTSChannelUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelUp"

	b.ListChannels(false)

	// Set Upper Boundary
	if b.Client.Self.Channel.ID == maxchannelid {
		log.Println("error: Can't Increment Channel Maximum Channel Reached")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText[2] = "Max Chan Reached"
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
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

					if LCDEnabled == true {
						go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled == true {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}
				break
			}
		}
	}
	return
}

func (b *Talkkonnect) ChannelDown() {
	if !(IsConnected) {
		return
	}

	if prevChannelID == 0 {
		prevChannelID = b.Client.Self.Channel.ID
	}

	if TTSEnabled && TTSChannelDown {
		err := PlayWavLocal(TTSChannelDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
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

			if LCDEnabled == true {
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
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

					if LCDEnabled == true {
						go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled == true {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}

				break
			}
		}
	}
	return
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
	return
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
			if LCDEnabled == true {
				LcdText[2] = "Status at " + t.Format("15:04:05")
				time.Sleep(500 * time.Millisecond)
				LcdText[3] = b.Client.Self.Comment
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 1, 1, "Status at "+t.Format("15:04:05"))
				oledDisplay(false, 4, 1, b.Client.Self.Comment)
			}
		}
	}
}

func (b *Talkkonnect) BackLightTimer() {
	BackLightTime = *BackLightTimePtr

	if TargetBoard != "rpi" || (LCDBackLightTimerEnabled == false && OLEDEnabled == false && LCDEnabled == false) {
		return
	}

	if LCDEnabled == true {
		b.LEDOn(b.BackLightLED)
	}

	if OLEDEnabled == true {
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
			log.Println(fmt.Sprintf("error: Ping Error ", err))
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
