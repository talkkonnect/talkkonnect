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
	"github.com/talkkonnect/volume-go"
	"os/exec"
)

func PlayWavLocal(filepath string, playbackvolume int) error {
	origVolume, _ = volume.GetVolume(OutputDevice)
	var player string

	if path, err := exec.LookPath("aplay"); err == nil {
		player = path
	} else if path, err := exec.LookPath("paplay"); err == nil {
		player = path
	} else {
		return errors.New("Failed to find either aplay or paplay in PATH")
	}

	cmd := exec.Command(player, filepath)
	err := volume.SetVolume(playbackvolume, OutputDevice)

	if err != nil {
		return errors.New(fmt.Sprintf("error: set volume failed: %+v", err))
	}
	_, err = cmd.CombinedOutput()

	if err != nil {
		return errors.New(fmt.Sprintf("error: cmd.Run() for %s failed with %s\n", player, err))
	}
	err = volume.SetVolume(origVolume, OutputDevice)

	if err != nil {
		return errors.New(fmt.Sprintf("error: set volume failed: %+v", err))
	}
	return nil
}
