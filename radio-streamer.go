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
 * The Initial Developer of the Original Code is Junsheng Cheng from https://github.com/talkkonnect/goradio
 *
 * Code modified for talkkonnect by
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
 * radio-stream.go -- streaming online radio listener function for talkkonnect
 *
 */

package talkkonnect

import (
	"io"
	"log"
	"os"
	"os/exec"
)

type RadioStation struct {
	name       string
	stream_url string
	volume     string
}

type Dj struct {
	player          RadioPlayer
	stations        []RadioStation
	current_station int
}

var stations = load_stations()
var status_chan = make(chan string)
var pipe_chan = make(chan io.ReadCloser)
var ffmpeg = FFmpeg{player_name: "ffmpeg", is_playing: false, pipe_chan: pipe_chan}
var player = Dj{player: &ffmpeg, stations: stations, current_station: -1}

func checker(err error) {
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
}

func Exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func (dj *Dj) Play(station int) {
	if 0 <= station && station < len(dj.stations) && dj.current_station != station {
		if dj.current_station >= 0 {
			dj.player.Close()
		}

		dj.current_station = station
		dj.player.Play(dj.stations[dj.current_station].stream_url, dj.stations[dj.current_station].volume)
	}
}

func (dj *Dj) Stop() {
	if dj.current_station >= 0 {
		dj.player.Close()
		dj.current_station = -1
	}
}

type RadioPlayer interface {
	Play(stream_url string, volume string)
	Close()
}

type FFmpeg struct {
	player_name string
	is_playing  bool
	stream_url  string
	command     *exec.Cmd
	in          io.WriteCloser
	out         io.ReadCloser
	pipe_chan   chan io.ReadCloser
}

func (player *FFmpeg) Play(stream_url string, volume string) {

	if !player.is_playing {
		var err error
		CmdArguments := []string{"-i", stream_url, "-filter:a", "volume=" + volume, "-f", "alsa", "default"}
		player.command = exec.Command("/usr/bin/ffmpeg", CmdArguments...)

		player.in, err = player.command.StdinPipe()
		checker(err)
		player.out, err = player.command.StdoutPipe()
		checker(err)

		err = player.command.Start()
		checker(err)

		player.is_playing = true
		player.stream_url = stream_url
		go func() {
			player.pipe_chan <- player.out
		}()
	}
}

func (player *FFmpeg) Close() {
	if player.is_playing {
		player.is_playing = false
		player.in.Write([]byte("q"))
		player.in.Close()
		player.out.Close()
		player.command = nil
		player.stream_url = ""
	}
}

func load_stations() []RadioStation {
	var stations []RadioStation
	stations = append(stations, RadioStation{"comedy", "https://listen.181fm.com/181-comedy_128k.mp3", "0.05"})
	stations = append(stations, RadioStation{"WBEZ 91.5", "http://stream.wbez.org/wbez128.mp3", "0.05"})
	stations = append(stations, RadioStation{"WGN", "http://provisioning.streamtheworld.com/pls/WGNPLUSAM.pls", "0.05"})
	return stations
}
