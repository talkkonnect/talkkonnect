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
	"os"
	"sort"
	"strconv"
	"syscall"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
)

var (
	prevChannelID uint32
	maxchannelid  uint32
)

func FatalCleanUp(message string) {
	log.Println("alert: " + message)
	log.Println("alert: Talkkonnect Terminated Abnormally with the Error(s) As Described Above, Ignore any GPIO errors if you are not using Single Board Computer.")
	log.Println("info: This Screen will close in 5 seconds")
	time.Sleep(5 * time.Second)
	term.Close()
	os.Exit(1)
}

func CleanUp(withShutdown bool) {

	if Config.Global.Hardware.TargetBoard == "rpi" {
		t := time.Now()
		if LCDEnabled {
			LcdText = [4]string{"talkkonnect stopped", t.Format("02-01-2006 15:04:05"), "Please Visit", "www.talkkonnect.com"}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, OLEDStartColumn, "talkkonnect stopped")
			oledDisplay(false, 1, OLEDStartColumn, t.Format("02-01-2006 15:04:05"))
			oledDisplay(false, 1, OLEDStartColumn, "version "+talkkonnectVersion)
			oledDisplay(false, 3, OLEDStartColumn, "Report Any Bugs To")
			oledDisplay(false, 4, OLEDStartColumn, "https://github.com/")
			oledDisplay(false, 5, OLEDStartColumn, "talkkonnect")
			oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
		}
		GPIOOutAll("led/relay", "off")
		//		MyLedStripGPIOOffAll()
	}

	term.Close()
	fmt.Println("SIGHUP Termination of Program Requested by User...shutting down talkkonnect")
	if withShutdown {
		time.Sleep(5 * time.Second)
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}
	os.Exit(0)
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
		//go MyLedStripTransmitLEDOn()
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
		//MyLedStripTransmitLEDOff()
		if LCDEnabled {
			LcdText[0] = "Online/RX" // b.Name
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 0, OLEDStartColumn, "Online/RX") //b.Name
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
	//fixthis change so that it checks accessablechannellist map first before changing to the requested channel
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
				oledDisplay(false, 0, OLEDStartColumn, "Joined "+ChannelName)
				oledDisplay(false, 1, OLEDStartColumn, Username[AccountIndex])
			}
		}

		log.Println("info: Joined Channel Name: ", channel.Name, " ID ", channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		log.Println("warn: Unable to Find Channel Name: ", ChannelName)
		prevChannelID = 0
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
			log.Printf("info: %d. User %#v is online. [%v]\n", item, usr.Name, usr.Comment)
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

	currentChannelIndex := b.findChannelIndex(b.Client.Self.Channel.ID)

	//handling of roll over when max channel reached
	if b.Client.Self.Channel == TopChannel {
		log.Println("debug: Maximum Channel Reached Rolling Over")
		for i := 0; i <= len(b.Client.Channels)-1; i++ {
			if chanName, found := AccessableChannelMap[ChannelsList[i].chanID]; found {
				log.Printf("info: Moving to Accessable Channel (ID %v Name %v)\n", ChannelsList[i].chanID, chanName)
				channel := b.Client.Channels.Find(chanName)
				if channel != nil {
					b.Client.Self.Move(channel)
					break
				} else {
					b.Client.Self.Move(RootChannel)
					break
				}
			}
		}
		return
	}

	//handling of connecting to next channel in accessable channel index
	for i := currentChannelIndex + 1; i <= len(b.Client.Channels)-1; i++ {
		if chanName, found := AccessableChannelMap[ChannelsList[i].chanID]; found {
			log.Printf("info: Moving to Accessable Channel (ID %v Name %v)\n", ChannelsList[i].chanID, chanName)
			channel := b.Client.Channels.Find(chanName)
			b.Client.Self.Move(channel)
			break
		} else {
			log.Println("alert: Skipping Unaccessable Channel!")
		}
	}
}

func (b *Talkkonnect) ChannelDown() {
	if !(IsConnected) {
		return
	}
	ChannelAction = "channeldown"
	TTSEvent("channeldown")

	currentChannelIndex := b.findChannelIndex(b.Client.Self.Channel.ID)

	if currentChannelIndex == 0 {
		log.Println("debug: Root Channel Reached Rolling Over")
		if TopChannel != nil {
			b.Client.Self.Move(TopChannel)
			return
		} else {
			log.Println("alert: Skipping Unaccessable Channel!")
		}
	}

	//handling of connecting to previous channel in accessable channel index
	for i := currentChannelIndex - 1; i >= 0; i-- {
		if chanName, found := AccessableChannelMap[ChannelsList[i].chanID]; found {
			log.Printf("info: Moving to Accessable Channel (ID %v Name %v)\n", ChannelsList[i].chanID, chanName)
			channel := b.Client.Channels.Find(chanName)
			if channel != nil {
				b.Client.Self.Move(channel)
				break
			} else {
				b.Client.Self.Move(RootChannel)
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

func (b *Talkkonnect) AddListeningChannelID(channelid []uint32) {
	if IsConnected {
		b.Client.Self.AddListeningChannel(channelid)
	}
}

func (b *Talkkonnect) RemoveListeningChannelID(channelid []uint32) {
	if IsConnected {
		b.Client.Self.RemoveListeningChannel(channelid)
	}
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
				oledDisplay(false, 1, OLEDStartColumn, "Status at "+t.Format("15:04:05"))
				oledDisplay(false, 4, OLEDStartColumn, b.Client.Self.Comment)
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
			log.Printf("error: Ping Error %q\n", err)
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
			b.VoiceTarget.Clear()
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
		}
	}
}

func (b *Talkkonnect) listeningToChannels(command string) {
	if !(IsConnected) {
		return
	}

	ListeningChannelNames := []string{}
	ListeningChannelIDs := []uint32{}

	for _, ChannelNames := range Config.Accounts.Account[AccountIndex].Listentochannels.ChannelNames {
		channel := b.Client.Channels.Find(ChannelNames)
		if channel != nil {
			ListeningChannelNames = append(ListeningChannelNames, channel.Name)
			ListeningChannelIDs = append(ListeningChannelIDs, channel.ID)
		}
	}

	if command == "start" {
		log.Printf("debug: Adding Channels %v With IDs %v For Listening\n", ListeningChannelNames, ListeningChannelIDs)
		b.AddListeningChannelID(ListeningChannelIDs)
		return
	}

	if command == "stop" {
		log.Printf("debug: Removing Channels %v With IDs %v For Listening\n", ListeningChannelNames, ListeningChannelIDs)
		b.RemoveListeningChannelID(ListeningChannelIDs)
	}
}

func (b *Talkkonnect) cmdListeningStart() {
	if !(IsConnected) {
		return
	}
	b.listeningToChannels("start")
}

func (b *Talkkonnect) cmdListeningStop() {
	if !(IsConnected) {
		return
	}
	b.listeningToChannels("stop")
}
