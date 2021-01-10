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
	"bufio"
	"bytes"
	"fmt"
	_ "github.com/talkkonnect/gumble/opus"
	"github.com/talkkonnect/volume-go"
	"log"
	"runtime"
	"strconv"
	"time"
)

func (b *Talkkonnect) commandKeyDel() {
	log.Println("debug: Delete Key Pressed Menu and Session Information Requested")

	if TTSEnabled && TTSDisplayMenu {
		err := PlayWavLocal(TTSDisplayMenuFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSDisplayMenuFilenameAndPath) Returned Error: ", err)
		}

	}

	b.talkkonnectMenu("\u001b[44;1m") // add blue background to banner reference https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#background-colors
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) commandKeyF1() {
	log.Println("debug: F1 pressed Channel Up (+) Requested")
	b.ChannelUp()
}

func (b *Talkkonnect) commandKeyF2() {
	log.Println("debug: F2 pressed Channel Down (-) Requested")
	b.ChannelDown()
}

func (b *Talkkonnect) commandKeyF3(subCommand string) {
	log.Println("debug: ", TTSMuteUnMuteSpeakerFilenameAndPath)

	//any other subcommand besides mute and unmute will get the current status of mute from volume.go
	origMuted, err := volume.GetMuted(OutputDevice)

	if err != nil {
		log.Println("error: get muted failed: %+v", err)
	}

	//force mute
	if subCommand == "mute" {
		origMuted = false
	}

	//force unmute
	if subCommand == "unmute" {
		origMuted = true
	}

	if origMuted {
		err := volume.Unmute(OutputDevice)

		if err != nil {
			log.Println("error: unmute failed: %+v", err)
		}

		log.Println("debug: F3 pressed Mute/Unmute Speaker Requested Now UnMuted")
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("error: PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath) Returned Error: ", err)
			}

		}
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Unmuted")
			}

		}
	} else {
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("error: PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath) Returned Error: ", err)
			}

		}
		err = volume.Mute(OutputDevice)
		if err != nil {
			log.Println("error: Mute failed: %+v", err)
		}

		log.Println("debug: F3 pressed Mute/Unmute Speaker Requested Now Muted")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Muted"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Muted")
			}

		}
	}

}

func (b *Talkkonnect) commandKeyF4() {
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("error: Unable to get current volume: %+v", err)
	}

	log.Println("debug: F4 pressed Volume Level Requested")
	log.Println("info: Volume Level is at", origVolume, "%")

	if TTSEnabled && TTSCurrentVolumeLevel {
		err := PlayWavLocal(TTSCurrentVolumeLevelFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSCurrentVolumeLevelFilenameAndPath) Returned Error: ", err)
		}

	}
	if TargetBoard == "rpi" {
		if LCDEnabled == true {
			LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(origVolume)}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(origVolume))
		}

	}
}

func (b *Talkkonnect) commandKeyF5() {
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("warn: unable to get original volume: %+v", err)
	}

	if origVolume < 100 {
		err := volume.IncreaseVolume(+1, OutputDevice)
		if err != nil {
			log.Println("warn: F5 Increase Volume Failed! ", err)
		}

		log.Println("debug: F5 pressed Volume UP (+)")
		log.Println("info: Volume UP (+) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Volume + " + strconv.Itoa(origVolume)}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(origVolume))
			}
		}
	} else {
		log.Println("debug: F5 Increase Volume")
		log.Println("info: Already at Maximum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Max Vol"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Max Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeUp {
		err := PlayWavLocal(TTSDigitalVolumeUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSDigitalVolumeUpFilenameAndPath) Returned Error: ", err)
		}

	}

}

func (b *Talkkonnect) commandKeyF6() {
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("error: unable to get original volume: %+v", err)
	}

	if origVolume > 0 {
		origVolume--
		err := volume.IncreaseVolume(-1, OutputDevice)
		if err != nil {
			log.Println("error: F6 Decrease Volume Failed! ", err)
		}

		log.Println("info: F6 pressed Volume Down (-)")
		log.Println("debug: Volume Down (-) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Volume - " + strconv.Itoa(origVolume)}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Volume -")
			}

		}
	} else {
		log.Println("debug: F6 Increase Volume Already")
		log.Println("info: Already at Minimum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Min Vol"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Min Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeDown {
		err := PlayWavLocal(TTSDigitalVolumeDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSDigitalVolumeDownFilenameAndPath) Returned Error: ", err)
		}

	}

}

func (b *Talkkonnect) commandKeyF7() {
	log.Println("debug: F7 pressed Channel List Requested")

	if TTSEnabled && TTSListServerChannels {
		err := PlayWavLocal(TTSListServerChannelsFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSListServerChannelsFilenameAndPath) Returned Error: ", err)
		}

	}

	b.ListChannels(true)
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) commandKeyF8() {
	log.Println("debug: F8 pressed TX Mode Requested (Start Transmitting)")
	log.Println("info: Start Transmitting")

	if TTSEnabled && TTSStartTransmitting {
		err := PlayWavLocal(TTSStartTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSStartTransmittingFilenameAndPath) Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if !b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStart()
	} else {
		log.Println("error: Already in Transmitting Mode")
	}
}

func (b *Talkkonnect) commandKeyF9() {
	log.Println("debug: F9 pressed RX Mode Request (Stop Transmitting)")
	log.Println("info: Stop Transmitting")

	if TTSEnabled && TTSStopTransmitting {
		err := PlayWavLocal(TTSStopTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: Play Wav Local Module Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStop(true)
	} else {
		log.Println("info: Not Already Transmitting")
	}
}

func (b *Talkkonnect) commandKeyF10() {
	log.Println("debug: F10 pressed Online User(s) in Current Channel Requested")
	log.Println("info: F10 Online User(s) in Current Channel")

	if TTSEnabled && TTSListOnlineUsers {
		err := PlayWavLocal(TTSListOnlineUsersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSListOnlineUsersFilenameAndPath) Returned Error: ", err)
		}

	}

	log.Println(fmt.Sprintf("info: Channel %#v Has %d Online User(s)", b.Client.Self.Channel.Name, len(b.Client.Self.Channel.Users)))
	b.ListUsers()
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) commandKeyF11() {
	log.Println("debug: F11 pressed Start/Stop Chimes Stream into Current Channel Requested")
	log.Println("info: Stream into Current Channel")

	b.BackLightTimer()

	if TTSEnabled && TTSPlayChimes {
		err := PlayWavLocal(TTSPlayChimesFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSPlayChimesFilenameAndPath) Returned Error: ", err)

		}

	}

	if b.IsTransmitting {
		log.Println("alert: talkkonnect was already transmitting will now stop transmitting and start the stream")
		b.TransmitStop(false)
	}

	IsPlayStream = !IsPlayStream
	NowStreaming = IsPlayStream

	if IsPlayStream {
		b.SendMessage(fmt.Sprintf("%s Streaming", b.Username), false)
	}

	go b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)

}

func (b *Talkkonnect) commandKeyF12() {
	log.Println("debug: F12 pressed")
	log.Println("info: GPS details requested")

	if TTSEnabled && TTSRequestGpsPosition {
		err := PlayWavLocal(TTSRequestGpsPositionFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSRequestGpsPositionFilenameAndPath) Returned Error: ", err)
		}

	}

	var i int = 0
	var tries int = 10

	for i = 0; i < tries; i++ {
		goodGPSRead, err := getGpsPosition(true)

		if err != nil {
			log.Println("error: GPS Function Returned Error Message", err)
			break
		}

		if goodGPSRead == true {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
	}

}

func (b *Talkkonnect) commandKeyCtrlC() {
	log.Println("debug: Ctrl-C Terminate Program Requested")
	duration := time.Since(StartTime)
	log.Printf("info: Talkkonnect Now Running For %v \n", secondsToHuman(int(duration.Seconds())))

	if TTSEnabled && TTSQuitTalkkonnect {
		err := PlayWavLocal(TTSQuitTalkkonnectFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSQuitTalkkonnectFilenameAndPath) Returned Error: ", err)
		}

	}
	ServerHop = true
	b.CleanUp()
}

func (b *Talkkonnect) commandKeyCtrlD() {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, true)
	var debug bytes.Buffer
	debug.WriteString(string(buf[0:stackSize]))
	scanner := bufio.NewScanner(&debug)
	var line int = 1
	log.Println("debug: Pressed Ctrl-D")
	log.Println("info: Stack Dump Requested")
	for scanner.Scan() {
		log.Printf("debug: line: %d %s", line, scanner.Text())
		line++
	}
}

func (b *Talkkonnect) commandKeyCtrlE() {
	log.Println("debug: Ctrl-E Pressed")
	log.Println("info: Send Email Requested")

	var i int = 0
	var tries int = 10

	for i = 0; i < tries; i++ {
		goodGPSRead, err := getGpsPosition(false)

		if err != nil {
			log.Println("error: GPS Function Returned Error Message", err)
			break
		}

		if goodGPSRead == true {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
		return
	}

	if TTSEnabled && TTSSendEmail {
		err := PlayWavLocal(TTSSendEmailFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSSendEmailFilenameAndPath) Returned Error: ", err)
		}

	}

	if EmailEnabled {

		emailMessage := fmt.Sprintf(EmailMessage + "\n")
		emailMessage = emailMessage + fmt.Sprintf("Ident: %s \n", b.Ident)
		emailMessage = emailMessage + fmt.Sprintf("Mumble Username: %s \n", b.Username)

		if EmailGpsDateTime {
			emailMessage = emailMessage + fmt.Sprintf("Date "+GPSDate+" UTC Time "+GPSTime+"\n")
		}

		if EmailGpsLatLong {
			emailMessage = emailMessage + fmt.Sprintf("Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)+" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64)+"\n")
		}

		if EmailGoogleMapsURL {
			emailMessage = emailMessage + "http://www.google.com/maps/place/" + strconv.FormatFloat(GPSLatitude, 'f', 6, 64) + "," + strconv.FormatFloat(GPSLongitude, 'f', 6, 64)
		}

		err := sendviagmail(EmailUsername, EmailPassword, EmailReceiver, EmailSubject, emailMessage)
		if err != nil {
			log.Println("error: Error from Email Module: ", err)
		}
	} else {
		log.Println("warning: Sending Email Disabled in Config")
	}
}

func (b *Talkkonnect) commandKeyCtrlF() {
	log.Println("debug: Ctrl-F Pressed")
	log.Println("info: Previous Server Requested")

	if TTSEnabled && TTSPreviousServer {
		err := PlayWavLocal(TTSPreviousServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSPreviousServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {

		if AccountIndex > 0 {
			AccountIndex--
		} else {
			AccountIndex = AccountCount - 1
		}

		modifyXMLTagServerHopping(ConfigXMLFile, "test.xml", AccountIndex)
	}

}

func (b *Talkkonnect) commandKeyCtrlL() {
	reset()
	log.Println("debug: Ctrl-L Pressed Cleared Screen")
	if TargetBoard == "rpi" {
		if LCDEnabled == true {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}

		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 0, "") // clear the screen
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlO() {
	log.Println("debug: Ctrl-O Pressed")
	log.Println("info: Ping Servers")

	if TTSEnabled && TTSPingServers {
		err := PlayWavLocal(TTSPingServersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: PlayWavLocal(TTSPingServersFilenameAndPath) Returned Error: ", err)
		}

	}

	b.pingServers()
}

func (b *Talkkonnect) commandKeyCtrlN() {
	log.Println("debug: Ctrl-N Pressed")
	log.Println("info: Next Server Requested")

	if TTSEnabled && TTSNextServer {
		err := PlayWavLocal(TTSNextServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: PlayWavLocal(TTSNextServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {
		if AccountIndex < AccountCount-1 {
			AccountIndex++
		} else {
			AccountIndex = 0
		}

		modifyXMLTagServerHopping(ConfigXMLFile, "test.xml", AccountIndex)
	}

}

func (b *Talkkonnect) commandKeyCtrlI() {
	log.Println("debug: Ctrl-I Pressed")
	log.Println("info: Traffic Recording Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "traffic" {
		log.Println("warn: Traffic Recording Not Enabled")
	}

	if AudioRecordEnabled == true {
		if AudioRecordMode == "traffic" {
			if AudioRecordFromOutput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordTraffic()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Traffic Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Traffic Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("info: Traffic Recording is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlJ() {
	log.Println("debug: Ctrl-J Pressed")
	log.Println("info: Ambient (Mic) Recording Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "ambient" {
		log.Println("warn: Ambient (Mic) Recording Not Enabled")
	}

	if AudioRecordEnabled == true {
		if AudioRecordMode == "ambient" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordAmbient()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Mic Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Mic Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("error: Ambient (Mic) Recording is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlK() {
	log.Println("debug: Ctrl-K Pressed")
	log.Println("info: Recording (Traffic and Mic) Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "combo" {
		log.Println("warn: Combo Recording (Traffic and Mic) Not Enabled")
	}

	if AudioRecordEnabled == true {
		if AudioRecordMode == "combo" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordCombo()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Combo Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Combo Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("error: Combo Recording (Traffic and Mic) is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlP() {
	if !(IsConnected) {
		return
	}
	b.BackLightTimer()
	log.Println("debug: Ctrl-P Pressed")
	log.Println("info: Panic Button Start/Stop Simulation Requested")

	if TTSEnabled && TTSPanicSimulation {
		err := PlayWavLocal(TTSPanicSimulationFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSPanicSimulationFilenameAndPath) Returned Error: ", err)
		}

	}

	if PEnabled {

		if b.IsTransmitting {
			b.TransmitStop(false)
		} else {
			b.IsTransmitting = true
			b.SendMessage(PMessage, PRecursive)

		}

		if PSendIdent {
			b.SendMessage(fmt.Sprintf("My Username is %s and Ident is %s", b.Username, b.Ident), PRecursive)
		}

		if PSendGpsLocation && GpsEnabled {

			var i int = 0
			var tries int = 10

			for i = 0; i < tries; i++ {
				goodGPSRead, err := getGpsPosition(false)

				if err != nil {
					log.Println("error: GPS Function Returned Error Message", err)
					break
				}

				if goodGPSRead == true {
					break
				}
			}

			if i == tries {
				log.Println("warn: Could Not Get a Good GPS Read")
			}

			if goodGPSRead == true && i != tries {
				log.Println("info: Sending GPS Info My Message")
				gpsMessage := "My GPS Coordinates are " + fmt.Sprintf(" Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)) + fmt.Sprintf(" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64))
				b.SendMessage(gpsMessage, PRecursive)
			}

			IsPlayStream = true
			b.playIntoStream(PFilenameAndPath, PVolume)
			if TargetBoard == "rpi" {
				if LCDEnabled == true {
					LcdText = [4]string{"nil", "nil", "nil", "Panic Message Sent!"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
					oledDisplay(false, 6, 1, "Panic Message Sent!")
				}
			}
			if PTxLockEnabled && PTxlockTimeOutSecs > 0 {
				b.TxLockTimer()
			}

		} else {
			log.Println("warn: Panic Function Disabled in Config")
		}
		IsPlayStream = false
		b.IsTransmitting = false
		b.LEDOff(b.TransmitLED)
	}
}

func (b *Talkkonnect) commandKeyCtrlR() {
	log.Println("debug: Ctrl-R Pressed")
	log.Println("info: Repeat TX Test Requested")
	isrepeattx = !isrepeattx
	go b.repeatTx()
}

func (b *Talkkonnect) commandKeyCtrlS() {
	log.Println("debug: Ctrl-S Pressed")
	log.Println("info: Scanning Channels")

	if TTSEnabled && TTSScan {
		err := PlayWavLocal(TTSScanFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSScanFilenameAndPath) Returned Error: ", err)
		}

	}

	b.Scan()
}

func (b *Talkkonnect) commandKeyCtrlT() {
	log.Println("debug: Ctrl-T Pressed")
	log.Println("info: Thanks and Acknowledgements Screen Request ")
	talkkonnectAcknowledgements("\u001b[44;1m") // add blue background to banner reference https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#background-colors
}

func (b *Talkkonnect) commandKeyCtrlV() {
	log.Println("debug: Ctrl-V Pressed")
	log.Println("info: Ctrl-V Version Request")
	log.Printf("info: Talkkonnect Version %v Released %v\n", talkkonnectVersion, talkkonnectReleased)
}

func (b *Talkkonnect) commandKeyCtrlU() {
	log.Println("debug: Ctrl-U Pressed")
	log.Println("info: Talkkonnect Uptime Request ")
	duration := time.Since(StartTime)
	log.Printf("info: Talkkonnect Now Running For %v \n", secondsToHuman(int(duration.Seconds())))
}

func (b *Talkkonnect) commandKeyCtrlX() {
	log.Println("debug: Ctrl-X Pressed")
	log.Println("info: Print XML Config " + ConfigXMLFile)

	if TTSEnabled && TTSPrintXmlConfig {
		err := PlayWavLocal(TTSPrintXmlConfigFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: PlayWavLocal(TTSPrintXmlConfigFilenameAndPath) Returned Error: ", err)
		}

	}

	printxmlconfig()
}
