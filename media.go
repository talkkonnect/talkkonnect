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
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

func localPlaybackDeviceName() string {
	return Config.Global.Software.Settings.LocalPlaybackDevice
}

func applyLocalPlaybackDevice(cmd *exec.Cmd) {
	dev := localPlaybackDeviceName()
	if dev == "" {
		return
	}
	cmd.Env = append(os.Environ(), "AUDIODEV="+dev)
}

func aplayLocal(fileNameWithPath string) {
	var player string
	var CmdArguments = []string{}

	if path, err := exec.LookPath("aplay"); err == nil {
		player = path
		if dev := localPlaybackDeviceName(); dev != "" {
			CmdArguments = []string{"-D", dev, fileNameWithPath, "-q", "-N"}
		} else {
			CmdArguments = []string{fileNameWithPath, "-q", "-N"}
		}
	} else if path, err := exec.LookPath("paplay"); err == nil {
		CmdArguments = []string{fileNameWithPath}
		player = path
	} else {
		return
	}

	log.Printf("debug: player %v CmdArguments %v", player, CmdArguments)

	cmd := exec.Command(player, CmdArguments...)
	applyLocalPlaybackDevice(cmd)

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
	applyLocalPlaybackDevice(cmd)

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
		applyLocalPlaybackDevice(cmd)
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

var announcementPlayMu sync.Mutex

func multimediaDelaySeconds(d time.Duration) time.Duration {
	if d <= 0 {
		return 0
	}
	return d * time.Second
}

func defaultPlaybackVolume(volume int) int {
	if volume <= 0 {
		return 50
	}
	return volume
}

func defaultStreamVolume(volume float32) float32 {
	if volume <= 0 {
		return 50
	}
	return volume
}

func mediaSourceLoop(loop int) int {
	if loop <= 0 {
		return 1
	}
	return loop
}

func multimediaFilePlayable(path string) bool {
	if len(strings.TrimSpace(path)) == 0 {
		return false
	}
	return FileExists(path) || checkRegex("(http|https|rtsp)", path)
}

func findMultimediaProfileIndex(mediaID string) int {
	mediaID = strings.TrimSpace(mediaID)
	for i, profile := range Config.Global.Multimedia.ID {
		if profile.Enabled && strings.EqualFold(profile.Value, mediaID) {
			return i
		}
	}
	return -1
}

func (b *Talkkonnect) cmdAnnouncement(mediaID string) {
	go b.playAnnouncementMedia(mediaID)
}

func (b *Talkkonnect) playAnnouncementMedia(mediaID string) {
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		log.Println("error: announcement media id is empty")
		return
	}

	idx := findMultimediaProfileIndex(mediaID)
	if idx < 0 {
		log.Printf("error: multimedia profile %q not found or disabled", mediaID)
		return
	}

	announcementPlayMu.Lock()
	defer announcementPlayMu.Unlock()

	profile := Config.Global.Multimedia.ID[idx]
	log.Printf("info: playing multimedia announcement profile %q", mediaID)

	if profile.Params.Voicetarget {
		log.Println("warn: voicetarget playback for multimedia is not implemented yet")
	}

	if profile.Params.Localplay {
		b.playMultimediaLocal(idx)
	}
	if profile.Params.Playintostream {
		b.playMultimediaIntoStream(idx)
	}
}

func (b *Talkkonnect) multimediaApplyDelay(idx int, pre bool) {
	profile := Config.Global.Multimedia.ID[idx]
	if pre {
		if profile.Params.Predelay.Enabled {
			if delay := multimediaDelaySeconds(profile.Params.Predelay.Value); delay > 0 {
				time.Sleep(delay)
			}
		}
		return
	}
	if profile.Params.Postdelay.Enabled {
		if delay := multimediaDelaySeconds(profile.Params.Postdelay.Value); delay > 0 {
			time.Sleep(delay)
		}
	}
}

func (b *Talkkonnect) playMultimediaLocal(idx int) {
	profile := Config.Global.Multimedia.ID[idx]

	if profile.Params.GPIO.Enabled {
		GPIOOutPin(profile.Params.GPIO.Name, "on")
	}
	defer func() {
		if profile.Params.GPIO.Enabled {
			GPIOOutPin(profile.Params.GPIO.Name, "off")
		}
	}()

	b.multimediaApplyDelay(idx, true)

	if profile.Params.Announcementtone.Enabled && multimediaFilePlayable(profile.Params.Announcementtone.File) {
		vol := defaultPlaybackVolume(profile.Params.Announcementtone.Volume)
		localMediaPlayer(profile.Params.Announcementtone.File, vol, profile.Params.Announcementtone.Blocking, 0, 1)
	}

	for _, source := range profile.Media.Source {
		if !source.Enabled || !multimediaFilePlayable(source.File) {
			continue
		}
		log.Printf("debug: local multimedia playing %q file %q", source.Name, source.File)
		localMediaPlayer(source.File, defaultPlaybackVolume(source.Volume), source.Blocking, source.Duration, mediaSourceLoop(source.Loop))
	}

	b.multimediaApplyDelay(idx, false)
	log.Printf("info: finished local multimedia announcement profile %q", profile.Value)
}

func (b *Talkkonnect) playFileIntoMumbleStream(filepath string, vol float32) {
	if !multimediaFilePlayable(filepath) {
		log.Printf("warn: cannot play into stream, file missing or unsupported: %s", filepath)
		return
	}

	b.BackLightTimer()
	if b.IsTransmitting {
		log.Println("alert: talkkonnect was already transmitting; stopping TX before announcement stream playback")
		b.TransmitStop(false)
	}

	GPIOOutPin("transmit", "on")
	b.splayIntoStream(filepath, defaultStreamVolume(vol))
	GPIOOutPin("transmit", "off")
}

func (b *Talkkonnect) playMultimediaIntoStream(idx int) {
	profile := Config.Global.Multimedia.ID[idx]
	streamVol := defaultStreamVolume(profile.Params.Streamvolume)

	b.multimediaApplyDelay(idx, true)

	if profile.Params.Announcementtone.Enabled && multimediaFilePlayable(profile.Params.Announcementtone.File) {
		b.playFileIntoMumbleStream(profile.Params.Announcementtone.File, streamVol)
	}

	for _, source := range profile.Media.Source {
		if !source.Enabled || !multimediaFilePlayable(source.File) {
			continue
		}
		log.Printf("debug: stream multimedia playing %q file %q", source.Name, source.File)
		loops := mediaSourceLoop(source.Loop)
		for i := 0; i < loops; i++ {
			b.playFileIntoMumbleStream(source.File, streamVol)
		}
	}

	b.multimediaApplyDelay(idx, false)
	log.Printf("info: finished stream multimedia announcement profile %q", profile.Value)
}

func (b *Talkkonnect) announcementSchedules() {
	for _, profile := range Config.Global.Multimedia.ID {
		if !profile.Enabled || !profile.Schedule.Enabled || profile.Schedule.IntervalSecs <= 0 {
			continue
		}
		mediaID := profile.Value
		intervalSecs := profile.Schedule.IntervalSecs
		go func(id string, secs int) {
			log.Printf("info: multimedia schedule started for profile %q every %v seconds", id, secs)
			ticker := time.NewTicker(time.Duration(secs) * time.Second)
			for range ticker.C {
				b.playAnnouncementMedia(id)
			}
		}(mediaID, intervalSecs)
	}
}

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
