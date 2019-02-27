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
 * playsound.go -> talkkonnect function to play sound locally and into mumble stream
 */

package talkkonnect

import (
	"errors"
	"fmt"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/volume-go"
	"log"
	"os/exec"
	"time"
)

var stream *gumbleffmpeg.Stream

func (b *Talkkonnect) PlayIntoStream(filepath string, vol float32) {

	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
		log.Println(fmt.Sprintf("info: Streaming Stopped", filepath))
		stream.Stop()
		b.LEDOff(b.TransmitLED)
		return
	}

	if ChimesSoundEnabled {
		if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
			time.Sleep(100 * time.Millisecond)
			return
		}

		b.LEDOn(b.TransmitLED)

		time.Sleep(100 * time.Millisecond)
		stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)

		if err := stream.Play(); err != nil {
			log.Println(fmt.Sprintf("alert: Can't play %s error %s", filepath, err))
		} else {
			log.Println(fmt.Sprintf("info: File %s Playing!", filepath))
			stream.Wait()
			stream.Stop()
			b.LEDOff(b.TransmitLED)
		}
	} else {
		log.Println(fmt.Sprintf("alert: Sound Disabled by Config"))
	}
	return
}

func (b *Talkkonnect) RogerBeep(filepath string, vol float32) error {
	if RogerBeepSoundEnabled {
		if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)
		if err := stream.Play(); err != nil {
			return errors.New(fmt.Sprintf("alert: Can't Play Roger beep File %s error %s", filepath, err))
		} else {
			log.Println("info: Roger Beep File " + filepath + " Playing!")
		}
	} else {
		log.Println(fmt.Sprintf("alert: Roger Beep Sound Disabled by Config"))
	}
	return nil
}

func PlayWavLocal(filepath string, playbackvolume int) error {
	origVolume, _ = volume.GetVolume(OutputDevice)
	cmd := exec.Command("/usr/bin/aplay", filepath)
	err := volume.SetVolume(playbackvolume, OutputDevice)
	if err != nil {
		return errors.New(fmt.Sprintf("alert: set volume failed: %+v", err))
	}
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("alert: cmd.Run() for aplay failed with %s\n", err))
	}
	err = volume.SetVolume(origVolume, OutputDevice)
	if err != nil {
		return errors.New(fmt.Sprintf("alert: set volume failed: %+v", err))
	}
	return nil
}
