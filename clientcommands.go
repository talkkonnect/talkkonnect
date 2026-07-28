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
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/volume-go"
)

var (
	prevChannelID uint32
	maxchannelid  uint32
	shutdownExitCode int32 // set to 1 on abnormal termination
	mainLoopRunning  atomic.Bool
)

// FatalCleanUp requests shutdown. While the main ClientStart loop is running, this cancels MasterCtx so Init can run performCleanup and return. During early initialization, it performs cleanup and exits the process.
func FatalCleanUp(message string) {
	log.Println("alert: " + message)
	log.Println("alert: Talkkonnect Terminated Abnormally with the Error(s) As Described Above, Ignore any GPIO errors if you are not using Single Board Computer.")
	atomic.StoreInt32(&shutdownExitCode, 1)
	if mainLoopRunning.Load() {
		if appTalkkonnect != nil && appTalkkonnect.masterCancel != nil {
			appTalkkonnect.masterCancel()
		}
		return
	}
	log.Println("info: Fatal error before main loop — exiting process after cleanup")
	time.Sleep(2 * time.Second)
	performCleanup(false)
	os.Exit(int(atomic.LoadInt32(&shutdownExitCode)))
}

// performCleanup releases hardware and lifecycle without exiting (used when returning from Init after ClientStart).
func performCleanup(withShutdown bool) {
	internetRadioShutdownKill()
	StopOpusTrafficRecording()

	if appTalkkonnect != nil {
		appTalkkonnect.shutdownDaemonLifecycle()
	}

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
		hwSyncGPIOOutAll("led/relay", "off")
	}

	fmt.Println("SIGHUP Termination of Program Requested by User...shutting down talkkonnect")
	bottomCLIDisableLayout()
	if withShutdown {
		time.Sleep(5 * time.Second)
		syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
	}
}

// CleanUp stops the process immediately after cleanup (menu quit, GPIO shutdown, duplicate instance).
func CleanUp(withShutdown bool) {
	performCleanup(withShutdown)
	os.Exit(0)
}

func (b *Talkkonnect) TransmitStart() {
	if !(IsConnected) {
		return
	}
	if b.IsTransmitting {
		log.Println("debug: Ignoring PTT start while still transmitting or playing roger beep")
		return
	}

	internetRadioNotifyVoiceOrTX()

	b.BackLightTimer()
	LastSpeaker = ""
	ClearUILastMessage()
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
		GPIOOutPin("transmit", "on")
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

	b.StopSource()
	b.IsTransmitting = false

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

	b.ScanStopForChannelChange()

	b.BackLightTimer()
	channelPath := strings.Split(ChannelName, ",")
	channel := b.Client.Channels.Find(channelPath...)
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
		sshRemoteReplyF("Not connected to Mumble; cannot list online users.\n")
		return
	}

	item := 0
	for _, usr := range b.Client.Users {
		if usr.Channel.ID == b.Client.Self.Channel.ID {
			item++
			log.Printf("info: %d. User %#v is online. [%v]\n", item, usr.Name, usr.Comment)
			sshRemoteReplyF("%d. User %#v is online. [%v]\n", item, usr.Name, usr.Comment)
		}
	}
	if item == 0 {
		sshRemoteReplyF("No other users in the current channel.\n")
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
		sshRemoteReplyF("Not connected to Mumble; channel change ignored.\n")
		return
	}

	b.ScanStopForChannelChange()

	ChannelAction = "channelup"
	TTSEvent("channelup")

	currentChannelIndex := b.findChannelIndex(b.Client.Self.Channel.ID)

	//handling of roll over when max channel reached
	if b.Client.Self.Channel == TopChannel {
		log.Println("debug: Maximum Channel Reached Rolling Over")
		for i := 0; i <= len(b.Client.Channels)-1; i++ {
			if chanName, found := AccessableChannelMap[ChannelsList[i].chanID]; found {
				log.Printf("info: Moving to Accessable Channel (ID %v Name %v)\n", ChannelsList[i].chanID, chanName)
				sshRemoteReplyF("Moving to channel ID %v name %v\n", ChannelsList[i].chanID, chanName)
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
			sshRemoteReplyF("Moving to channel ID %v name %v\n", ChannelsList[i].chanID, chanName)
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
		sshRemoteReplyF("Not connected to Mumble; channel change ignored.\n")
		return
	}
	b.ScanStopForChannelChange()

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
			sshRemoteReplyF("Moving to channel ID %v name %v\n", ChannelsList[i].chanID, chanName)
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
		sshRemoteReplyF("Server #%d [%s]%s\n", i+1, Name[i], currentconn)

		if err != nil {
			log.Printf("error: Ping Error %q\n", err)
			sshRemoteReplyF("Ping error: %v\n", err)
			continue
		}

		major, minor, patch := resp.Version.SemanticVersion()

		log.Println("info: Server Address:         ", resp.Address)
		log.Println("info: Server Ping:            ", resp.Ping)
		log.Println("info: Server Version:         ", major, ".", minor, ".", patch)
		log.Println("info: Server Users:           ", resp.ConnectedUsers, "/", resp.MaximumUsers)
		log.Println("info: Server Maximum Bitrate: ", resp.MaximumBitrate)
		sshRemoteReplyF("  Address: %v\n", resp.Address)
		sshRemoteReplyF("  Ping:    %v\n", resp.Ping)
		sshRemoteReplyF("  Version: %v.%v.%v\n", major, minor, patch)
		sshRemoteReplyF("  Users:   %v / %v\n", resp.ConnectedUsers, resp.MaximumUsers)
		sshRemoteReplyF("  Max BR:  %v\n", resp.MaximumBitrate)
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
		RecordUIVoiceTarget(TargetID, "user", TargetUser)
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
		RecordUIVoiceTarget(targetID, "channel", vChannel.Name)
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
	RecordUIVoiceTarget(targetID, "channel", vChannel.Name)
	if targetID > 0 {
		GPIOOutPin("voicetarget", "on")
	}
}

// remoteAPIWhisperTargetID is the voice target slot ad-hoc whisper requests use.
// Valid Mumble target IDs are 1..30 and talkkonnect.xml normally configures the
// low ones under <voicetargets>, so the top slot is reserved for a whisper aimed
// at a user picked at run time — a dashboard click, say — which carries no
// configured ID of its own and must not overwrite a configured target.
const remoteAPIWhisperTargetID uint32 = 30

// cmdJoinChannel joins a channel by name instead of stepping through the tree
// with channelup / channeldown.
func (b *Talkkonnect) cmdJoinChannel(channelName string) {
	channelName = strings.TrimSpace(channelName)
	log.Printf("debug: Join Channel %v Requested\n", channelName)

	if !IsConnected || b.Client == nil {
		sshRemoteReplyF("Not connected to Mumble; cannot join a channel.\n")
		return
	}
	if channelName == "" {
		sshRemoteReplyF("Join channel requires a channel name.\n")
		return
	}
	// ChangeChannel only logs a warning when the name is unknown, so look it up
	// first to give the caller a real answer.
	if b.Client.Channels.Find(strings.Split(channelName, ",")...) == nil {
		log.Println("warn: Unable to Find Channel Name: ", channelName)
		sshRemoteReplyF("Channel %s not found on this server.\n", channelName)
		return
	}

	b.ChangeChannel(channelName)
	sshRemoteReplyF("Joining channel %s.\n", channelName)
}

// cmdWhisperUser points transmitted audio at one online user. Unlike
// voicetargetset, which replays a target configured in talkkonnect.xml, the user
// is named by the caller, so a dashboard can whisper to whoever it likes.
func (b *Talkkonnect) cmdWhisperUser(userName string) {
	userName = strings.TrimSpace(userName)
	log.Printf("debug: Whisper To User %v Requested\n", userName)

	if !IsConnected || b.Client == nil {
		sshRemoteReplyF("Not connected to Mumble; cannot set a whisper target.\n")
		return
	}
	if userName == "" {
		sshRemoteReplyF("Whisper requires a user name.\n")
		return
	}
	if b.Client.Users.Find(userName) == nil {
		sshRemoteReplyF("User %s is not online; whisper target unchanged.\n", userName)
		return
	}

	b.VoiceTargetUserSet(remoteAPIWhisperTargetID, userName)
	sshRemoteReplyF("Whispering to %s.\n", userName)
}

// cmdWhisperClear drops the whisper target so transmitted audio goes back to the
// joined channel.
func (b *Talkkonnect) cmdWhisperClear() {
	log.Println("debug: Clear Whisper Target Requested")

	if !IsConnected || b.Client == nil {
		sshRemoteReplyF("Not connected to Mumble; no whisper target to clear.\n")
		return
	}

	// Registering the slot with no entries releases it on the server, and a nil
	// client target makes gumble send audio on target 0 — normal talking.
	b.Client.Send(&gumble.VoiceTarget{ID: remoteAPIWhisperTargetID})
	b.Client.VoiceTarget = nil
	ClearUIVoiceTarget()
	GPIOOutPin("voicetarget", "off")
	b.sevenSegment("voicetarget", "0")
	log.Println("info: Whisper Target Cleared, Broadcasting To Channel")
	sshRemoteReplyF("Whisper target cleared, broadcasting to channel.\n")
}

// cmdSetRXVolume sets the speaker volume to an absolute percentage, which is what
// a dragged slider needs; volumerxup / volumerxdown only step by the configured
// amount. No TTS announcement here on purpose — a drag lands many values in a row
// and every one of them would queue up a spoken message.
func (b *Talkkonnect) cmdSetRXVolume(value int) {
	log.Printf("debug: Set RX Volume To %v%% Requested\n", value)

	if value < 0 || value > 100 {
		log.Println("error: RX Volume Out Of Range ", value)
		sshRemoteReplyF("Volume %d%% is out of range, expected 0 to 100.\n", value)
		return
	}

	if err := volume.SetVolume(value, Config.Global.Software.Settings.OutputVolControlDevice); err != nil {
		log.Println("error: Set RX Volume Failed! ", err)
		sshRemoteReplyF("Unable to set volume: %v\n", err)
		return
	}

	current, err := volume.GetVolume(Config.Global.Software.Settings.OutputVolControlDevice)
	if err != nil {
		current = value
	}
	log.Println("info: RX Volume Now At ", current, "%")
	sshRemoteReplyF("Volume now at %d%%\n", current)

	if Config.Global.Hardware.TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(current)}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 6, OLEDStartColumn, "Volume "+strconv.Itoa(current))
		}
		b.sevenSegment("localvolume", strconv.Itoa(current))
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
