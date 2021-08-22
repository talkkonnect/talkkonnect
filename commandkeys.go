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
	"log"
	"runtime"
	"strconv"
	"time"

	"github.com/talkkonnect/volume-go"
)

func (b *Talkkonnect) cmdDisplayMenu() {
	log.Println("debug: Delete Key Pressed Menu and Session Information Requested")

	if TTSEnabled && TTSDisplayMenu {
		err := playWavLocal(TTSDisplayMenuFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSDisplayMenuFilenameAndPath) Returned Error: ", err)
		}

	}

	b.talkkonnectMenu("\u001b[44;1m") // add blue background to banner reference https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#background-colors
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) cmdChannelUp() {
	log.Println("debug: F1 pressed Channel Up (+) Requested")
	b.ChannelUp()
}

func (b *Talkkonnect) cmdChannelDown() {
	log.Println("debug: F2 pressed Channel Down (-) Requested")
	b.ChannelDown()
}

func (b *Talkkonnect) cmdMuteUnmute(subCommand string) {

	log.Println("debug: F3 pressed Mute/Unmute Speaker Requested")
	if TTSEnabled && TTSMuteUnMuteSpeaker {
		err := playWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath) Returned Error: ", err)
		}
	}

	OrigMuted, err := volume.GetMuted(OutputDevice)

	if err != nil {
		log.Println("error: Unable to get current Muted/Unmuted State ", err)
	} else {
		if OrigMuted {
			log.Println("debug: Originally Device is Muted")
		} else {
			log.Println("debug: Originally Device is Unmuted")
		}
	}

	if subCommand == "toggle" {
		if OrigMuted {
			err := volume.Unmute(OutputDevice)
			if err != nil {
				log.Println("error: Unmuting Failed", err)
				return
			}
			log.Println("info: Output Device Unmuted")
			if TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 6, 1, "Unmuted")
				}
			}
			return
		} else {
			err = volume.Mute(OutputDevice)
			if err != nil {
				log.Println("error: Muting Failed", err)
			}
			log.Println("info: Output Device Muted")
			if TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText = [4]string{"nil", "nil", "nil", "Muted"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 6, 1, "Muted")
				}
			}
			return
		}
	}

	//force mute
	if subCommand == "mute" {
		err = volume.Mute(OutputDevice)
		if err != nil {
			log.Println("error: Muting Failed ", err)
			return
		}
		log.Println("info: Output Device Muted")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"nil", "nil", "nil", "Muted"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)

				if OLEDEnabled {
					oledDisplay(false, 6, 1, "Muted")
				}
			}
			return
		}

		//force unmute
		if subCommand == "unmute" {
			err := volume.Unmute(OutputDevice)
			if err != nil {
				log.Println("error: Unmute Failed ", err)
				return
			}
			log.Println("info: Output Device Unmuted")
			if TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 6, 1, "Unmuted")
				}
			}
			return
		}

	}
}
func (b *Talkkonnect) cmdCurrentVolume() {
	OrigVolume, err := volume.GetVolume(OutputDeviceShort)
	if err != nil {
		log.Printf("error: Unable to get current volume: %+v\n", err)
	}

	log.Println("debug: F4 pressed Volume Level Requested")
	log.Println("info: Volume Level is at", OrigVolume, "%")

	if TTSEnabled && TTSCurrentVolumeLevel {
		err := playWavLocal(TTSCurrentVolumeLevelFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSCurrentVolumeLevelFilenameAndPath) Returned Error: ", err)
		}

	}
	if TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(OrigVolume)}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(OrigVolume))
		}

	}
}

func (b *Talkkonnect) cmdVolumeUp() {
	origVolume, err := volume.GetVolume(OutputDeviceShort)
	if err != nil {
		log.Printf("warn: unable to get original volume: %+v\n", err)
	}

	if origVolume < 100 {
		err := volume.IncreaseVolume(+1, OutputDeviceShort)
		if err != nil {
			log.Println("warn: F5 Increase Volume Failed! ", err)
		}

		log.Println("debug: F5 pressed Volume UP (+)")
		log.Println("info: Volume UP (+) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"nil", "nil", "nil", "Volume + " + strconv.Itoa(origVolume)}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(origVolume))
			}
		}
	} else {
		log.Println("debug: F5 Increase Volume")
		log.Println("info: Already at Maximum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"nil", "nil", "nil", "Max Vol"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 6, 1, "Max Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeUp {
		err := playWavLocal(TTSDigitalVolumeUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSDigitalVolumeUpFilenameAndPath) Returned Error: ", err)
		}

	}

}

func (b *Talkkonnect) cmdVolumeDown() {
	origVolume, err := volume.GetVolume(OutputDeviceShort)
	if err != nil {
		log.Printf("error: unable to get original volume: %+v\n", err)
	}

	if origVolume > 0 {
		origVolume--
		err := volume.IncreaseVolume(-1, OutputDeviceShort)
		if err != nil {
			log.Println("error: F6 Decrease Volume Failed! ", err)
		}

		log.Println("info: F6 pressed Volume Down (-)")
		log.Println("info: Volume Down (-) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"nil", "nil", "nil", "Volume - " + strconv.Itoa(origVolume)}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 6, 1, "Volume -")
			}

		}
	} else {
		log.Println("debug: F6 Increase Volume Already")
		log.Println("info: Already at Minimum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"nil", "nil", "nil", "Min Vol"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 6, 1, "Min Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeDown {
		err := playWavLocal(TTSDigitalVolumeDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSDigitalVolumeDownFilenameAndPath) Returned Error: ", err)
		}

	}

}

func (b *Talkkonnect) cmdListServerChannels() {
	log.Println("debug: F7 pressed Channel List Requested")

	if TTSEnabled && TTSListServerChannels {
		err := playWavLocal(TTSListServerChannelsFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSListServerChannelsFilenameAndPath) Returned Error: ", err)
		}

	}

	b.ListChannels(true)
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) cmdStartTransmitting() {
	log.Println("debug: F8 pressed TX Mode Requested (Start Transmitting)")
	log.Println("info: Start Transmitting")

	if TTSEnabled && TTSStartTransmitting {
		err := playWavLocal(TTSStartTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSStartTransmittingFilenameAndPath) Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(StreamSoundFilenameAndPath, StreamSoundVolume)
	}

	if !b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStart()
	} else {
		log.Println("error: Already in Transmitting Mode")
	}
}

func (b *Talkkonnect) cmdStopTransmitting() {
	log.Println("debug: F9 pressed RX Mode Request (Stop Transmitting)")
	log.Println("info: Stop Transmitting")

	if TTSEnabled && TTSStopTransmitting {
		err := playWavLocal(TTSStopTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: Play Wav Local Module Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(StreamSoundFilenameAndPath, StreamSoundVolume)
	}

	if b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStop(true)
	} else {
		log.Println("info: Not Already Transmitting")
	}
}

func (b *Talkkonnect) cmdListOnlineUsers() {
	log.Println("debug: F10 pressed Online User(s) in Current Channel Requested")
	log.Println("info: F10 Online User(s) in Current Channel")

	if TTSEnabled && TTSListOnlineUsers {
		err := playWavLocal(TTSListOnlineUsersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSListOnlineUsersFilenameAndPath) Returned Error: ", err)
		}

	}

	log.Println(fmt.Sprintf("info: Channel %#v Has %d Online User(s)", b.Client.Self.Channel.Name, len(b.Client.Self.Channel.Users)))
	b.ListUsers()
	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) cmdPlayback() {
	log.Println("debug: F11 pressed Start/Stop Stream Stream into Current Channel Requested")
	log.Println("info: Stream into Current Channel")

	b.BackLightTimer()

	if TTSEnabled && TTSPlayStream {
		err := playWavLocal(TTSPlayStreamFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSPlayStreamFilenameAndPath) Returned Error: ", err)

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

	go b.playIntoStream(StreamSoundFilenameAndPath, StreamSoundVolume)

}

func (b *Talkkonnect) cmdGPSPosition() {
	log.Println("debug: F12 pressed")
	log.Println("info: GPS details requested")

	if TTSEnabled && TTSRequestGpsPosition {
		err := playWavLocal(TTSRequestGpsPositionFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSRequestGpsPositionFilenameAndPath) Returned Error: ", err)
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

		if goodGPSRead {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
	}

}

func (b *Talkkonnect) cmdQuitTalkkonnect() {
	log.Println("debug: Ctrl-C Terminate Program Requested")
	duration := time.Since(StartTime)
	log.Printf("info: Talkkonnect Now Running For %v \n", secondsToHuman(int(duration.Seconds())))

	if TTSEnabled && TTSQuitTalkkonnect {
		err := playWavLocal(TTSQuitTalkkonnectFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSQuitTalkkonnectFilenameAndPath) Returned Error: ", err)
		}

	}
	ServerHop = true
	b.CleanUp()
}

func (b *Talkkonnect) cmdDebugStacktrace() {
	buf := make([]byte, 1<<16)
	stackSize := runtime.Stack(buf, true)
	var debug bytes.Buffer
	debug.WriteString(string(buf[0:stackSize]))
	scanner := bufio.NewScanner(&debug)
	var line int = 1
	log.Println("debug: Pressed Ctrl-D")
	log.Println("info: Stack Dump Requested")
	for scanner.Scan() {
		log.Printf("debug: line: %d %s\n", line, scanner.Text())
		line++
	}
}

func (b *Talkkonnect) cmdSendEmail() {
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

		if goodGPSRead {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
		return
	}

	if TTSEnabled && TTSSendEmail {
		err := playWavLocal(TTSSendEmailFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: playWavLocal(TTSSendEmailFilenameAndPath) Returned Error: ", err)
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

func (b *Talkkonnect) cmdConnPreviousServer() {
	log.Println("debug: Ctrl-F Pressed")
	log.Println("info: Previous Server Requested")

	if TTSEnabled && TTSPreviousServer {
		err := playWavLocal(TTSPreviousServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSPreviousServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {

		if AccountIndex > 0 {
			AccountIndex--
		} else {
			AccountIndex = AccountCount - 1
		}

		modifyXMLTagServerHopping(ConfigXMLFile, AccountIndex)
	}

}

func (b *Talkkonnect) cmdClearScreen() {
	reset()
	log.Println("debug: Ctrl-L Pressed Cleared Screen")
	if TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}

		if OLEDEnabled {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 0, "")
		}
	}
}

func (b *Talkkonnect) cmdPingServers() {
	log.Println("debug: Ctrl-O Pressed")
	log.Println("info: Ping Servers")

	if TTSEnabled && TTSPingServers {
		err := playWavLocal(TTSPingServersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: playWavLocal(TTSPingServersFilenameAndPath) Returned Error: ", err)
		}

	}

	b.pingServers()
}

func (b *Talkkonnect) cmdConnNextServer() {
	log.Println("debug: Ctrl-N Pressed")
	log.Println("info: Next Server Requested Killing This Session, talkkonnect should be restarted by systemd")

	if TTSEnabled && TTSNextServer {
		err := playWavLocal(TTSNextServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: playWavLocal(TTSNextServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {
		if AccountIndex < AccountCount-1 {
			AccountIndex++
		} else {
			AccountIndex = 0
		}

		modifyXMLTagServerHopping(ConfigXMLFile, AccountIndex)
	}

}

func (b *Talkkonnect) cmdAudioTrafficRecord() {
	log.Println("debug: Ctrl-I Pressed")
	log.Println("info: Traffic Recording Requested")
	if !AudioRecordEnabled {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "traffic" {
		log.Println("warn: Traffic Recording Not Enabled")
	}

	if AudioRecordEnabled {
		if AudioRecordMode == "traffic" {
			if AudioRecordFromOutput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordTraffic()
					if TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "Traffic Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
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

func (b *Talkkonnect) cmdAudioMicRecord() {
	log.Println("debug: Ctrl-J Pressed")
	log.Println("info: Ambient (Mic) Recording Requested")
	if !AudioRecordEnabled {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "ambient" {
		log.Println("warn: Ambient (Mic) Recording Not Enabled")
	}

	if AudioRecordEnabled {
		if AudioRecordMode == "ambient" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordAmbient()
					if TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "Mic Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
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

func (b *Talkkonnect) cmdAudioMicTrafficRecord() {
	log.Println("debug: Ctrl-K Pressed")
	log.Println("info: Recording (Traffic and Mic) Requested")
	if !AudioRecordEnabled {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "combo" {
		log.Println("warn: Combo Recording (Traffic and Mic) Not Enabled")
	}

	if AudioRecordEnabled {
		if AudioRecordMode == "combo" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					go AudioRecordCombo()
					if TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "Combo Audio Rec ->", "nil"} // 4 or 3
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
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

func (b *Talkkonnect) cmdPanicSimulation() {
	if !(IsConnected) {
		return
	}
	b.BackLightTimer()
	log.Println("debug: Ctrl-P Pressed")
	log.Println("info: Panic Button Start/Stop Simulation Requested")

	if TTSEnabled && TTSPanicSimulation {
		err := playWavLocal(TTSPanicSimulationFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSPanicSimulationFilenameAndPath) Returned Error: ", err)
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

		if PSendGpsLocation {

			var i int = 0
			var tries int = 10

			for i = 0; i < tries; i++ {
				goodGPSRead, err := getGpsPosition(false)

				if err != nil {
					log.Println("error: GPS Function Returned Error Message", err)
					break
				}

				if goodGPSRead {
					break
				}
			}

			if i == tries {
				log.Println("warn: Could Not Get a Good GPS Read")
			}

			if goodGPSRead && i != tries {
				log.Println("info: Sending GPS Info My Message")
				gpsMessage := "My GPS Coordinates are " + fmt.Sprintf(" Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)) + fmt.Sprintf(" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64))
				b.SendMessage(gpsMessage, PRecursive)
			}

			IsPlayStream = true
			b.playIntoStream(PFilenameAndPath, PVolume)
			if TargetBoard == "rpi" {
				if LCDEnabled {
					LcdText = [4]string{"nil", "nil", "nil", "Panic Message Sent!"}
					LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled {
					oledDisplay(false, 6, 1, "Panic Message Sent!")
				}
			}
			if PTxLockEnabled && PTxlockTimeOutSecs > 0 {
				b.TxLockTimer()
			}

			// New. Send email after Panic Event //
			if PMailEnabled {
				b.cmdSendEmail()
				log.Println("info: Sending Panic Alert Email To Predefined Email Address")
			}
			//

			// New. Record ambient audio on Panic Event if recording is enabled
			if AudioRecordEnabled {
				log.Println("info: Running sox for Audio Recording...")
				AudioRecordAmbient()
			}
			//

		} else {
			log.Println("warn: Panic Function Disabled in Config")
		}
		IsPlayStream = false
		b.IsTransmitting = false

		if PLowProfile {
			b.LEDOffAll()
			log.Println("info: Low Profile Lights Option is Enabled. Turning All Leds Off During Panic Event")
			if LCDEnabled {
				log.Println("info: Low Profile Lights is Enabled. Turning Off Display During Panic Event")
				LcdText = [4]string{"", "", "", ""}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(true, 0, 0, "")
			}
		}
	}
}

func (b *Talkkonnect) cmdRepeatTxLoop() {
	log.Println("debug: Ctrl-R Pressed")
	log.Println("info: Repeat TX Test Requested")
	isrepeattx = !isrepeattx
	go b.repeatTx()
}

func (b *Talkkonnect) cmdScanChannels() {
	log.Println("debug: Ctrl-S Pressed")
	log.Println("info: Scanning Channels")

	if TTSEnabled && TTSScan {
		err := playWavLocal(TTSScanFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSScanFilenameAndPath) Returned Error: ", err)
		}

	}

	b.Scan()
}

func (b *Talkkonnect) cmdThanks() {
	log.Println("debug: Ctrl-T Pressed")
	log.Println("info: Thanks and Acknowledgements Screen Request ")
	talkkonnectAcknowledgements("\u001b[44;1m") // add blue background to banner reference https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#background-colors
}

func (b *Talkkonnect) cmdShowUptime() {
	log.Println("debug: Ctrl-V Pressed")
	log.Println("info: Ctrl-V Version Request")
	log.Printf("info: Talkkonnect Version %v Released %v\n", talkkonnectVersion, talkkonnectReleased)
}

func (b *Talkkonnect) cmdDisplayVersion() {
	log.Println("debug: Ctrl-U Pressed")
	log.Println("info: Talkkonnect Uptime Request ")
	duration := time.Since(StartTime)
	log.Printf("info: Talkkonnect Now Running For %v \n", secondsToHuman(int(duration.Seconds())))
}

func (b *Talkkonnect) cmdDumpXMLConfig() {
	log.Println("debug: Ctrl-X Pressed")
	log.Println("info: Print XML Config " + ConfigXMLFile)

	if TTSEnabled && TTSPrintXmlConfig {
		err := playWavLocal(TTSPrintXmlConfigFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("error: playWavLocal(TTSPrintXmlConfigFilenameAndPath) Returned Error: ", err)
		}

	}

	printxmlconfig()
}

func (b *Talkkonnect) cmdPlayRepeaterTone() {
	log.Println("debug: Ctrl-G Pressed")
	log.Println("info: Play Repeater Tone on Speaker and Simulate RX Signal")

	b.BackLightTimer()
	b.RepeaterTone()
}
