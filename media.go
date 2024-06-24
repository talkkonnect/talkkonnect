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
 * Code Copied from https://www.socketloop.com/tutorials/golang-convert-seconds-to-human-readable-time-format-example
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
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"
)

func aplayLocal(fileNameWithPath string) {
	var player string
	var CmdArguments = []string{}

	if path, err := exec.LookPath("aplay"); err == nil {
		CmdArguments = []string{fileNameWithPath, "-q", "-N"}
		player = path
	} else if path, err := exec.LookPath("paplay"); err == nil {
		CmdArguments = []string{fileNameWithPath}
		player = path
	} else {
		return
	}

	log.Printf("debug: player %v CmdArguments %v", player, CmdArguments)

	cmd := exec.Command(player, CmdArguments...)

	_, err := cmd.CombinedOutput()

	if err != nil {
		return
	}
}

func localMediaPlayer(fileNameWithPath string, playbackvolume int, blocking bool, duration float32, loop int) {

	if loop == 0 || loop > 3 {
		log.Println("warn: Infinite Loop or more than 3 loops not allowed")
		return
	}

	CmdArguments := []string{fileNameWithPath, "-volume", strconv.Itoa(playbackvolume), "-autoexit", "-loop", strconv.Itoa(loop), "-autoexit", "-nodisp"}

	if duration > 0 {
		CmdArguments = []string{fileNameWithPath, "-volume", strconv.Itoa(playbackvolume), "-autoexit", "-t", fmt.Sprintf("%.1f", duration), "-loop", strconv.Itoa(loop), "-autoexit", "-nodisp"}
	}

	cmd := exec.Command("/usr/bin/ffplay", CmdArguments...)

	WaitForFFPlay := make(chan struct{})
	go func() {
		cmd.Run()
		if blocking {
			WaitForFFPlay <- struct{}{} // signal that the routine has completed
		}
	}()
	if blocking {
		<-WaitForFFPlay
	}
}

func (b *Talkkonnect) PlayTone(toneFreq int, toneDuration float32, destination string, withRXLED bool) {

	toneFilePath := "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/soundfiles/repeatertones/"
	toneFileName := toneFilePath + "sine_" + strconv.Itoa(toneFreq) + "_" + strconv.FormatFloat(float64(toneDuration), 'f', -1, 64) + ".wav"

	if !FileExists(toneFileName) {
		cmdArguments := []string{"-f", "lavfi", "-i", "sine=frequency=" + strconv.Itoa(toneFreq) + ":duration=" + fmt.Sprintf("%f", toneDuration), toneFileName}

		cmd := exec.Command("/usr/bin/ffmpeg", cmdArguments...)
		err := cmd.Run()
		if err != nil {
			log.Println("error: ffmpeg error cannot generate tone file", err)
			return
		} else {
			log.Printf("info: Generated Tone File %v Successfully\n", toneFileName)
		}
	}

	if destination != "intostream" {

		cmdArguments := []string{toneFileName, "-autoexit", "-nodisp"}
		cmd := exec.Command("/usr/bin/ffplay", cmdArguments...)
		var out bytes.Buffer
		cmd.Stdout = &out

		if withRXLED {
			GPIOOutPin("voiceactivity", "on")
		}
		err := cmd.Run()
		if err != nil {
			log.Println("error: ffplay error ", err)
			if withRXLED {
				GPIOOutPin("voiceactivity", "off")
			}
			return
		}
		if withRXLED {
			GPIOOutPin("voiceactivity", "off")
		}

		log.Printf("info: Played Tone at Frequency %v Hz With Duration of %v Seconds Locally\n", toneFreq, toneDuration)
	} else {
		GPIOOutPin("transmit", "on")
		//MyLedStripTransmitLEDOn()
		log.Println("debug: Repeater Tone Playing")
		b.splayIntoStream(toneFileName, 50)
		GPIOOutPin("transmit", "off")
		log.Printf("info: Played Tone at Frequency %v Hz With Duration of %v Seconds Into Stream\n", toneFreq, toneDuration)
	}

}

/*
func playAnnouncementMedia(id int) {

	for _, multimedia := range Config.Global.Multimedia.ID {
		apiid, err := strconv.Atoi(multimedia.Value)
		if apiid == id && err == nil {
			if multimedia.Params.Localplay {
				if multimedia.Params.GPIO.Enabled {
					GPIOOutPin(multimedia.Params.GPIO.Name, "on")
				}
				if multimedia.Params.Predelay.Enabled && multimedia.Params.Predelay.Value > 0 {
					time.Sleep(multimedia.Params.Predelay.Value * time.Second)
				}
				if multimedia.Params.Announcementtone.Enabled && FileExists(multimedia.Params.Announcementtone.File) {
					localMediaPlayer(multimedia.Params.Announcementtone.File, multimedia.Params.Announcementtone.Volume, multimedia.Params.Announcementtone.Blocking, 0, 1) //todo replace 1 with volume from xmlconfig
				}
				for _, source := range multimedia.Media.Source {
					if source.Enabled {
						log.Printf("debug: Playing %v filename %v\n", source.Name, source.File)
						localMediaPlayer(source.File, source.Volume, multimedia.Params.Announcementtone.Blocking, source.Duration, source.Loop)
					}
				}
				if multimedia.Params.Postdelay.Enabled && multimedia.Params.Postdelay.Value > 0 {
					time.Sleep(multimedia.Params.Postdelay.Value * time.Second)
				}
				if multimedia.Params.GPIO.Enabled {
					GPIOOutPin(multimedia.Params.GPIO.Name, "off")
				}
			}
			if multimedia.Params.Playintostream {
				log.Println("alert: todo play into stream not implemented yet")
			}

			if multimedia.Params.Voicetarget {
				log.Println("alert: todo play to voice targets not implemented yet")
			}
		}
	}
}
*/

func findEventSound(findEventSound string) EventSoundStruct {
	for _, sound := range Config.Global.Software.Sounds.Sound {
		if sound.Enabled && sound.Event == findEventSound {
			return EventSoundStruct{sound.Enabled, sound.File, sound.Volume, sound.Blocking}
		}
	}
	return EventSoundStruct{false, "", "0", false}
}

func findInputEventSoundFile(findInputEventSound string) InputEventSoundFileStruct {
	for _, sound := range Config.Global.Software.Sounds.Input.Sound {
		if sound.Event == findInputEventSound {
			if sound.Enabled {
				return InputEventSoundFileStruct{sound.Event, sound.File, sound.Enabled}
			}
		}
	}
	return InputEventSoundFileStruct{findInputEventSound, "", false}
}

func playIOMedia(inputEvent string) {
	if Config.Global.Software.Sounds.Input.Enabled {
		var inputEventSoundFile InputEventSoundFileStruct = findInputEventSoundFile(inputEvent)
		if inputEventSoundFile.Enabled {
			go aplayLocal(inputEventSoundFile.File)
		}
	}
}

func (b *Talkkonnect) beaconPlay() {
	BeaconTime = *BeaconTimePtr
	if !Config.Global.Software.Beacon.Enabled {
		BeaconTime.Stop()
		return
	}

	go func() {
		BeaconTime = time.NewTicker(time.Duration(Config.Global.Software.Beacon.BeaconTimerSecs) * time.Second)
		for range BeaconTime.C {
			if Config.Global.Software.Beacon.Playintostream {
				IsPlayStream = true
				b.playIntoStream(Config.Global.Software.Beacon.BeaconFileAndPath, Config.Global.Software.Beacon.BeaconVolumeIntoStream)
				IsPlayStream = false
				log.Println("info: Beacon Enabled and Timed Out Auto Played File ", Config.Global.Software.Beacon.BeaconFileAndPath, " Into Stream")
			}
			if Config.Global.Software.Beacon.LocalPlay {
				if Config.Global.Software.Beacon.GPIOEnabled {
					GPIOOutPin(Config.Global.Software.Beacon.GPIOName, "on")
				}
				log.Printf("info: Local/RF Beacon Playing %v with volume %v", Config.Global.Software.Beacon.BeaconFileAndPath, Config.Global.Software.Beacon.LocalVolume)
				localMediaPlayer(Config.Global.Software.Beacon.BeaconFileAndPath, Config.Global.Software.Beacon.LocalVolume, true, 0, 1)
				if Config.Global.Software.Beacon.GPIOEnabled {
					GPIOOutPin(Config.Global.Software.Beacon.GPIOName, "off")
				}
			}
		}
	}()
}
