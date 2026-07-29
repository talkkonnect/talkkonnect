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
	localMediaPlayerWithOffset(fileNameWithPath, playbackvolume, blocking, duration, 0, loop)
}

// localMediaPlayerWithOffset plays a file on the local speaker, optionally seeking
// offset seconds into it and stopping after duration seconds. Options are passed
// ahead of the url because that is the order ffplay documents.
func localMediaPlayerWithOffset(fileNameWithPath string, playbackvolume int, blocking bool, duration float32, offset float32, loop int) {

	loop = mediaSourceLoop(loop)

	CmdArguments := []string{"-nodisp", "-autoexit", "-volume", strconv.Itoa(playbackvolume), "-loop", strconv.Itoa(loop)}

	if offset > 0 {
		CmdArguments = append(CmdArguments, "-ss", fmt.Sprintf("%.1f", offset))
	}

	if duration > 0 {
		CmdArguments = append(CmdArguments, "-t", fmt.Sprintf("%.1f", duration))
	}

	CmdArguments = append(CmdArguments, fileNameWithPath)

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

// maxMediaSourceLoops caps how often one media source may repeat. ffplay treats
// -loop 0 as infinite, so a missing or silly loop count must never reach it
// verbatim; the same ceiling is applied to into-stream playback for consistency.
const maxMediaSourceLoops = 3

func mediaSourceLoop(loop int) int {
	if loop <= 0 {
		return 1
	}
	if loop > maxMediaSourceLoops {
		log.Printf("warn: loop count %v exceeds the maximum of %v, clamping", loop, maxMediaSourceLoops)
		return maxMediaSourceLoops
	}
	return loop
}

// multimediaVoicetargetEnabled reads the <voicetarget> element text. The tag is
// carried as a string so the historic empty form (<voicetarget/>) does not fail
// the whole config parse.
func multimediaVoicetargetEnabled(value string) bool {
	enabled, err := strconv.ParseBool(strings.TrimSpace(value))
	return err == nil && enabled
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

	// The profile GPIO drives an external amplifier or attention light, so it has
	// to span the whole announcement rather than just the local speaker leg.
	if profile.Params.GPIO.Enabled {
		GPIOOutPin(profile.Params.GPIO.Name, "on")
		defer GPIOOutPin(profile.Params.GPIO.Name, "off")
	}

	if profile.Params.Localplay {
		b.playMultimediaLocal(idx)
	}
	if profile.Params.Playintostream {
		b.playMultimediaIntoStream(idx)
	}
}

// multimediaApplyVoiceTarget points an into-stream announcement at a voice target
// or at the plain channel, and returns a func that restores the previous routing.
// Local speaker playback has no routing, so this only applies to stream playback.
func (b *Talkkonnect) multimediaApplyVoiceTarget(idx int) func() {
	profile := Config.Global.Multimedia.ID[idx]

	if b.Client == nil {
		return func() {}
	}

	previous := b.Client.VoiceTarget
	previousUI := uiVoiceTargetSnapshot()
	restore := func() { b.Client.VoiceTarget = previous }

	if !multimediaVoicetargetEnabled(profile.Params.Voicetarget.Value) {
		if previous != nil {
			log.Printf("debug: multimedia profile %q has voicetarget disabled, announcing to the current channel", profile.Value)
			b.Client.VoiceTarget = nil
		}
		return restore
	}

	targetID := profile.Params.Voicetarget.ID

	if targetID == 0 {
		if previous == nil {
			log.Printf("warn: multimedia profile %q wants a voicetarget but none is active and no id is configured, announcing to the current channel", profile.Value)
		} else {
			log.Printf("info: announcing multimedia profile %q to the already active voicetarget id %v", profile.Value, previous.ID)
		}
		return restore
	}

	log.Printf("info: announcing multimedia profile %q to voicetarget id %v", profile.Value, targetID)
	b.cmdSendVoiceTargets(targetID)

	// cmdSendVoiceTargets is best effort: it silently does nothing when the id is
	// missing from <voicetargets>, and VoiceTargetUserSet / VoiceTargetChannelSet
	// bail out when the named user or channel is not on the server. Either way the
	// target is left unset, so fall back to the channel rather than announcing at
	// whatever target happened to be selected before.
	if b.Client.VoiceTarget == nil || b.Client.VoiceTarget.ID != targetID {
		log.Printf("warn: voicetarget id %v could not be applied, announcing to the current channel. Check that the id exists under <voicetargets> for the active account and that its user or channel is present on the server", targetID)
		b.Client.VoiceTarget = nil
	}

	// cmdSendVoiceTargets also drives the voicetarget LED and the /uistatus
	// telemetry, so both are handed back with the target itself.
	return func() {
		b.Client.VoiceTarget = previous
		restoreUIVoiceTarget(previousUI)
		if previous == nil {
			GPIOOutPin("voicetarget", "off")
		}
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
		localMediaPlayerWithOffset(source.File, defaultPlaybackVolume(source.Volume), source.Blocking, source.Duration, source.Offset, mediaSourceLoop(source.Loop))
	}

	b.multimediaApplyDelay(idx, false)
	log.Printf("info: finished local multimedia announcement profile %q", profile.Value)
}

func (b *Talkkonnect) playFileIntoMumbleStream(filepath string, vol float32) {
	b.playFileIntoMumbleStreamWithOptions(filepath, vol, 0, 0)
}

// playFileIntoMumbleStreamWithOptions plays a file into the mumble stream, seeking
// offset seconds in and stopping after duration seconds when either is non-zero.
func (b *Talkkonnect) playFileIntoMumbleStreamWithOptions(filepath string, vol float32, offset float32, duration float32) {
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
	b.splayIntoStreamWithOptions(filepath, defaultStreamVolume(vol), offset, duration)
	GPIOOutPin("transmit", "off")
}

// multimediaStreamVolume prefers the per-source volume and falls back to the
// profile streamvolume, matching how local playback treats the same attribute.
func multimediaStreamVolume(sourceVolume int, streamVolume float32) float32 {
	if sourceVolume > 0 {
		return float32(sourceVolume)
	}
	return streamVolume
}

func (b *Talkkonnect) playMultimediaIntoStream(idx int) {
	profile := Config.Global.Multimedia.ID[idx]
	streamVol := defaultStreamVolume(profile.Params.Streamvolume)

	restoreVoiceTarget := b.multimediaApplyVoiceTarget(idx)
	defer restoreVoiceTarget()

	b.multimediaApplyDelay(idx, true)

	if profile.Params.Announcementtone.Enabled && multimediaFilePlayable(profile.Params.Announcementtone.File) {
		b.playFileIntoMumbleStream(profile.Params.Announcementtone.File, multimediaStreamVolume(profile.Params.Announcementtone.Volume, streamVol))
	}

	for _, source := range profile.Media.Source {
		if !source.Enabled || !multimediaFilePlayable(source.File) {
			continue
		}
		log.Printf("debug: stream multimedia playing %q file %q", source.Name, source.File)
		loops := mediaSourceLoop(source.Loop)
		for i := 0; i < loops; i++ {
			b.playFileIntoMumbleStreamWithOptions(source.File, multimediaStreamVolume(source.Volume, streamVol), source.Offset, source.Duration)
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
