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
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/talkkonnect/gumble/gumbleffmpeg"
)

func aplayLocal(filepath string, playbackvolume int) error {
	var player string

	if path, err := exec.LookPath("aplay"); err == nil {
		player = path
	} else if path, err := exec.LookPath("paplay"); err == nil {
		player = path
	} else {
		return errors.New("failed to find either aplay or paplay in path")
	}

	log.Println("info: debug player ", player)
	log.Println("info: debug filepath ", filepath)
	cmd := exec.Command(player, filepath)

	_, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error: cmd.Run() for %s failed with %s", player, err)
	}

	return nil
}

func localMediaPlayer(fileNameWithPath string, playbackvolume float32, duration float32, loop int) {

	if loop == 0 || loop > 3 {
		log.Println("warn: Infinite Loop or more than 3 loops not allowed")
		return
	}

	CmdArguments := []string{fileNameWithPath, "-af", "volume=" + fmt.Sprintf("%1.1f", playbackvolume), "-autoexit", "-loop", strconv.Itoa(loop), "-autoexit", "-nodisp"}

	if duration > 0 {
		CmdArguments = []string{fileNameWithPath, "-af", "volume=" + fmt.Sprintf("%1.1f", playbackvolume), "-autoexit", "-t", fmt.Sprintf("%.1f", duration), "-loop", strconv.Itoa(loop), "-autoexit", "-nodisp"}
	}

	cmd := exec.Command("/usr/bin/ffplay", CmdArguments...)
	cmd.Run()

}

func (b *Talkkonnect) playIntoStream(filepath string, vol float32) {
	if !IsPlayStream {
		log.Println(fmt.Sprintf("info: File %s Stopped!", filepath))
		pstream.Stop()
		LEDOffFunc(TransmitLED)
		return
	}

	if StreamSoundEnabled && IsPlayStream {
		if pstream != nil && pstream.State() == gumbleffmpeg.StatePlaying {
			pstream.Stop()
			return
		}

		LEDOnFunc(TransmitLED)

		IsPlayStream = true
		pstream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)
		if err := pstream.Play(); err != nil {
			log.Println(fmt.Sprintf("error: Can't play %s error %s", filepath, err))
		} else {
			log.Println(fmt.Sprintf("info: File %s Playing!", filepath))
			pstream.Wait()
			pstream.Stop()
			LEDOffFunc(TransmitLED)
		}
	} else {
		log.Println("warn: Sound Disabled by Config")
	}
}

func (b *Talkkonnect) PlayTone(toneFreq int, toneDuration int, destination string, withRXLED bool) {

	if destination == "local" {

		cmdArguments := []string{"-f", "lavfi", "-i", "sine=frequency=" + strconv.Itoa(toneFreq) + ":duration=" + strconv.Itoa(toneDuration), "-autoexit", "-nodisp"}
		cmd := exec.Command("/usr/bin/ffplay", cmdArguments...)
		var out bytes.Buffer
		cmd.Stdout = &out

		if withRXLED {
			LEDOnFunc(VoiceActivityLED)
		}
		err := cmd.Run()
		if err != nil {
			log.Println("error: ffplay error ", err)
			if withRXLED {
				LEDOffFunc(VoiceActivityLED)
			}
			return
		}
		if withRXLED {
			LEDOffFunc(VoiceActivityLED)
		}

		log.Println("info: Played Tone at Frequency " + strconv.Itoa(RepeaterToneFrequencyHz) + " Hz With Duration of " + strconv.Itoa(RepeaterToneDurationSec) + " Seconds For Opening Repeater")

	}

}
