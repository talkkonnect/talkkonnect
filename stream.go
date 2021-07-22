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
 * stream.go part of mumble openal client modified to work with talkkonnect
 */

package talkkonnect

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"time"

	"github.com/talkkonnect/go-openal/openal"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
)

var (
	errState         = errors.New("gumbleopenal: invalid state")
	lcdtext          = [4]string{"nil", "nil", "nil", ""}
	BackLightLED     = gpio.NewOutput(uint(LCDBackLightLEDPin), false)
	VoiceActivityLED = gpio.NewOutput(VoiceActivityLEDPin, false)
	now              = time.Now()
	LastTime         = now.Unix()
	debuglevel       = 2
	emptyBufs        = openal.NewBuffers(16)
	StreamCounter    = 0
	TimerTalked      = time.NewTicker(time.Millisecond * 200)
	RXLEDStatus      = false
)

type Stream struct {
	client *gumble.Client
	link   gumble.Detacher

	deviceSource    *openal.CaptureDevice
	sourceFrameSize int
	sourceStop      chan bool

	deviceSink  *openal.Device
	contextSink *openal.Context
}

func New(client *gumble.Client) (*Stream, error) {
	s := &Stream{
		client:          client,
		sourceFrameSize: client.Config.AudioFrameSize(),
	}
	s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.sourceFrameSize))

	s.deviceSink = openal.OpenDevice("")

	s.contextSink = s.deviceSink.CreateContext()

	s.contextSink.Activate()

	s.link = client.Config.AttachAudio(s)

	return s, nil
}

func (s *Stream) Destroy() {
	if debuglevel >= 3 {
		log.Println("debug: Destroy Stream Source")
	}
	s.link.Detach()
	if s.deviceSource != nil {
		s.StopSource()
		s.deviceSource.CaptureCloseDevice()
		s.deviceSource = nil
	}
	if s.deviceSink != nil {
		s.contextSink.Destroy()
		s.deviceSink.CloseDevice()
		s.contextSink = nil
		s.deviceSink = nil
	}
}

func (s *Stream) StartSource() error {
	if debuglevel >= 3 {
		log.Println("debug: Start Stream Source")
	}
	if s.sourceStop != nil {
		return errState
	}

	if IncommingBeepSoundEnabled {
		s.playIntoStream(IncommingBeepSoundFilenameAndPath, IncommingBeepSoundVolume)
	}

	s.deviceSource.CaptureStart()
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	return nil
}

func (s *Stream) StopSource() error {
	if debuglevel >= 3 {
		log.Println("debug: Stop Source File")
	}
	if s.sourceStop == nil {
		return errState
	}
	close(s.sourceStop)
	s.sourceStop = nil
	s.deviceSource.CaptureStop()
	s.deviceSource.CaptureCloseDevice()

	if RogerBeepSoundEnabled {
		log.Println("debug: Rogerbeep Playing")
		s.playIntoStream(RogerBeepSoundFilenameAndPath, RogerBeepSoundVolume)
	}

	s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.sourceFrameSize))

	return nil
}

func (s *Stream) OnAudioStream(e *gumble.AudioStreamEvent) {

	if TargetBoard == "rpi" && LCDEnabled {
		LEDOffFunc(BackLightLED)
	}

	StreamCounter++

	TalkedTicker := time.NewTicker(200 * time.Millisecond)
	Talking := make(chan bool)

	if StreamCounter == 1 {
		go func() {
			for {
				select {
				case <-Talking:
					TalkedTicker.Reset(200 * time.Millisecond)
					if !RXLEDStatus {
						RXLEDStatus = true
						LEDOnFunc(VoiceActivityLED)
						log.Println("info: Speaking->", *e.LastSpeaker)
						t := time.Now()
						if TargetBoard == "rpi" {
							if LCDEnabled {
								LEDOnFunc(BackLightLED)
								lcdtext = [4]string{"nil", "", "", *e.LastSpeaker + " " + t.Format("15:04:05")}
								LcdDisplay(lcdtext, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
								BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)
							}

							if OLEDEnabled {
								Oled.DisplayOn()
								go oledDisplay(false, 3, 1, *e.LastSpeaker+" "+t.Format("15:04:05"))
								BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)
							}
						}
					}
				case <-TalkedTicker.C:
					RXLEDStatus = false
					LEDOffFunc(VoiceActivityLED)
					TalkedTicker.Stop()

				}
			}
		}()
	}

	go func() {
		if StreamCounter > 1 {
			StreamCounter--
			return
		}

		source = openal.NewSource()
		emptyBufs = openal.NewBuffers(16)

		reclaim := func() {
			if n := source.BuffersProcessed(); n > 0 {
				reclaimedBufs := make(openal.Buffers, n)
				source.UnqueueBuffers(reclaimedBufs)
				emptyBufs = append(emptyBufs, reclaimedBufs...)
			}
		}

		var raw [gumble.AudioMaximumFrameSize * 2]byte

		for packet := range e.C {
			Talking <- true

			if TargetBoard == "rpi" && LCDEnabled {
				LEDOnFunc(BackLightLED)
			}

			if CancellableStream && NowStreaming {
				pstream.Stop()
			}

			samples := len(packet.AudioBuffer)

			if samples > cap(raw) {
				continue
			}
			for i, value := range packet.AudioBuffer {
				binary.LittleEndian.PutUint16(raw[i*2:], uint16(value))
			}
			reclaim()
			if len(emptyBufs) == 0 {
				emptyBufs = openal.NewBuffers(16)
				continue
			}
			last := len(emptyBufs) - 1
			buffer := emptyBufs[last]
			emptyBufs = emptyBufs[:last]
			buffer.SetData(openal.FormatMono16, raw[:samples*2], gumble.AudioSampleRate)
			source.QueueBuffer(buffer)
			if source.State() != openal.Playing {
				source.Play()
			}
			Talking <- false
		}
		reclaim()
		emptyBufs.Delete()
		source.Delete()
	}()
}

func (s *Stream) sourceRoutine() {
	interval := s.client.Config.AudioInterval
	frameSize := s.client.Config.AudioFrameSize()

	if frameSize != s.sourceFrameSize {
		log.Println("error: FrameSize Error!")
		s.deviceSource.CaptureCloseDevice()
		s.sourceFrameSize = frameSize
		s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.sourceFrameSize))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	stop := s.sourceStop

	outgoing := s.client.AudioOutgoing()
	defer close(outgoing)

	for {
		select {
		case <-stop:
			if debuglevel >= 3 {
				log.Println("debug: Ticker Stop!")
			}
			return
		case <-ticker.C:
			//this is for encoding (transmitting)
			buff := s.deviceSource.CaptureSamples(uint32(frameSize))
			if len(buff) != frameSize*2 {
				continue
			}
			int16Buffer := make([]int16, frameSize)
			for i := range int16Buffer {
				int16Buffer[i] = int16(binary.LittleEndian.Uint16(buff[i*2 : (i+1)*2]))
			}
			outgoing <- gumble.AudioBuffer(int16Buffer)
		}
	}
}

func (s *Stream) playIntoStream(filepath string, vol float32) {
	pstream = gumbleffmpeg.New(s.client, gumbleffmpeg.SourceFile(filepath), vol)
	if err := pstream.Play(); err != nil {
		log.Println(fmt.Sprintf("error: Can't play %s error %s", filepath, err))
	} else {
		log.Println(fmt.Sprintf("info: File %s Playing!", filepath))
		pstream.Wait()
		pstream.Stop()
	}
}

func (b *Talkkonnect) playIntoStream(filepath string, vol float32) {
	if !IsPlayStream {
		log.Println(fmt.Sprintf("info: File %s Stopped!", filepath))
		pstream.Stop()
		b.LEDOff(b.TransmitLED)
		return
	}

	if StreamSoundEnabled && IsPlayStream {
		if pstream != nil && pstream.State() == gumbleffmpeg.StatePlaying {
			pstream.Stop()
			return
		}

		b.LEDOn(b.TransmitLED)

		IsPlayStream = true
		pstream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)
		if err := pstream.Play(); err != nil {
			log.Println(fmt.Sprintf("error: Can't play %s error %s", filepath, err))
		} else {
			log.Println(fmt.Sprintf("info: File %s Playing!", filepath))
			pstream.Wait()
			pstream.Stop()
			b.LEDOff(b.TransmitLED)
		}
	} else {
		log.Println("warn: Sound Disabled by Config")
	}
	return
}

func (b *Talkkonnect) RepeaterTone() {

	if RepeaterToneEnabled {

		cmdArguments := []string{"-f", "lavfi", "-i", "sine=frequency=" + strconv.Itoa(RepeaterToneFrequencyHz) + ":duration=" + strconv.Itoa(RepeaterToneDurationSec), "-autoexit", "-nodisp"}

		cmd := exec.Command("/usr/bin/ffplay", cmdArguments...)

		var out bytes.Buffer

		LEDOnFunc(VoiceActivityLED)
		cmd.Stdout = &out
		err := cmd.Run()
		LEDOffFunc(VoiceActivityLED)

		if err != nil {
			log.Println("error: ffplay error ", err)
		} else {
			log.Println("info: Played Tone at Frequency " + strconv.Itoa(RepeaterToneFrequencyHz) + " Hz With Duration of " + strconv.Itoa(RepeaterToneDurationSec) + " Seconds For Opening Repeater")
		}

	} else {
		log.Println("warn: Repeater Tone Disabled by Config")
	}

}

func (b *Talkkonnect) OpenStream() {
	if ServerHop {
		log.Println("debug: Server Hop Requested Will Now Destroy Old Server Stream")
		b.Stream.Destroy()
		var participantCount = len(b.Client.Self.Channel.Users)

		log.Println("info: Current Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")
		b.ListUsers()
		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText[0] = b.Address
				LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 0, 1, b.Address)
				oledDisplay(false, 1, 1, b.Client.Self.Channel.Name+" ("+strconv.Itoa(participantCount)+" Users)")
				oledDisplay(false, 6, 1, "Please Visit")
				oledDisplay(false, 7, 1, "www.talkkonnect.com")
			}

		}
	}

	if stream, err := New(b.Client); err != nil {

		if TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"Stream Error!", "nil", "nil", "nil"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 2, 1, "Stream Error!!")
			}

		}
		FatalCleanUp("Stream Open Error " + err.Error())
	} else {
		b.Stream = stream
	}
}

func (b *Talkkonnect) ResetStream() {
	b.Stream.Destroy()
	time.Sleep(50 * time.Millisecond)
	b.OpenStream()
}
