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
	"encoding/binary"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/talkkonnect/go-openal/openal"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
)

var (
	errState     = errors.New("gumbleopenal: invalid state")
	lcdtext      = [4]string{"nil", "nil", "nil", ""}
	now          = time.Now()
	TotalStreams int
	NeedToKill   int
)

// MumbleDuplex - listenera and outgoing
type MumbleDuplex struct{}

type Stream struct {
	client *gumble.Client
	link   gumble.Detacher

	deviceSource    *openal.CaptureDevice
	sourceFrameSize int
	sourceStop      chan bool

	deviceSink  *openal.Device
	contextSink *openal.Context
}

func (b *Talkkonnect) New(client *gumble.Client) (*Stream, error) {
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

func (b *Talkkonnect) Destroy() {
	b.Stream.link.Detach()
	if b.Stream.deviceSource != nil {
		b.Stream.deviceSource.CaptureStop()
		b.Stream.deviceSource.CaptureCloseDevice()
		b.Stream.deviceSource = nil
	}
	if b.Stream.deviceSink != nil {
		b.Stream.contextSink.Destroy()
		b.Stream.deviceSink.CloseDevice()
		b.Stream.contextSink = nil
		b.Stream.deviceSink = nil
	}
}

func (b *Talkkonnect) StartSource() error {
	var eventSound EventSoundStruct = findEventSound("incommingbeep")
	if eventSound.Enabled {
		if v, err := strconv.ParseFloat(eventSound.Volume, 32); err == nil {
			time.Sleep(300 * time.Millisecond)
			log.Println("alert: Playing Incomming into Stream")
			b.splayIntoStream(eventSound.FileName, float32(v))
		}
	}
	b.Stream.deviceSource.CaptureStart()
	b.Stream.sourceStop = make(chan bool)
	go b.sourceRoutine()
	return nil
}

func (b *Talkkonnect) StopSource() error {
	if b.Stream.sourceStop == nil {
		return errState
	}
	close(b.Stream.sourceStop)
	b.Stream.sourceStop = nil
	b.Stream.deviceSource.CaptureStop()
	b.Stream.deviceSource.CaptureCloseDevice()

	var eventSound EventSoundStruct = findEventSound("rogerbeep")
	if eventSound.Enabled {
		GPIOOutPin("transmit", "on")
		//MyLedStripTransmitLEDOn()
		log.Println("debug: Rogerbeep Playing")
		if v, err := strconv.ParseFloat(eventSound.Volume, 32); err == nil {
			b.splayIntoStream(eventSound.FileName, float32(v))
		}
		GPIOOutPin("transmit", "off")
		//MyLedStripTransmitLEDOff()
	}

	b.Stream.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(b.Stream.sourceFrameSize))

	return nil
}

func (s *Stream) OnAudioStream(e *gumble.AudioStreamEvent) {
	TotalStreams++
	if _, userexists := StreamTracker[e.User.UserID]; userexists {
		log.Printf("debug: Stale GoRoutine Detected For UserID=%v UserName=%v Session=%v AudioStreamChannel=%v", e.User.UserID, e.User.Name, e.User.Session, e.C)
		NeedToKill++
	}
	StreamTracker[e.User.UserID] = streamTrackerStruct{UserID: e.User.UserID, UserName: e.User.Name, UserSession: e.User.Session, C: e.C}
	goStreamStats()

	go func() {
		source := openal.NewSource()
		emptyBufs := openal.NewBuffers(24)
		reclaim := func() {
			if n := source.BuffersProcessed(); n > 0 {
				reclaimedBufs := make(openal.Buffers, n)
				source.UnqueueBuffers(reclaimedBufs)
				emptyBufs = append(emptyBufs, reclaimedBufs...)
			}
		}
		var raw [gumble.AudioMaximumFrameSize * 2]byte
		for packet := range e.C {
			TalkedTicker.Reset(Config.Global.Hardware.VoiceActivityTimermsecs * time.Millisecond)
			if Config.Global.Software.IgnoreUser.IgnoreUserEnabled {
				if len(Config.Global.Software.IgnoreUser.IgnoreUserRegex) > 0 {
					if checkRegex(Config.Global.Software.IgnoreUser.IgnoreUserRegex, e.User.Name) {
						continue
					}
				}
			}

			if Config.Global.Software.Settings.CancellableStream && NowStreaming {
				IsPlayStream = !IsPlayStream
				NowStreaming = IsPlayStream
				pstream.Stop()
			}
			Talking <- talkingStruct{true, e.User.Name, e.User.Channel.Name}
			samples := len(packet.AudioBuffer)
			if samples > cap(raw) {
				continue
			}
			for i, value := range packet.AudioBuffer {
				binary.LittleEndian.PutUint16(raw[i*2:], uint16(value))
			}
			reclaim()
			if len(emptyBufs) == 0 {
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
			Talking <- talkingStruct{false, e.User.Name, e.User.Channel.Name}
		}
		reclaim()
		emptyBufs.Delete()
		source.Delete()
	}()
}

func (b *Talkkonnect) sourceRoutine() {
	interval := b.Stream.client.Config.AudioInterval
	frameSize := b.Stream.client.Config.AudioFrameSize()

	if frameSize != b.Stream.sourceFrameSize {
		log.Println("error: FrameSize Error!")
		b.Stream.deviceSource.CaptureCloseDevice()
		b.Stream.sourceFrameSize = frameSize
		b.Stream.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(b.Stream.sourceFrameSize))
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	stop := b.Stream.sourceStop

	outgoing := b.Stream.client.AudioOutgoing()
	defer close(outgoing)

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			//this is for encoding (transmitting)
			buff := b.Stream.deviceSource.CaptureSamples(uint32(frameSize))
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

func (b *Talkkonnect) playIntoStream(filepath string, vol float32) {
	if !IsPlayStream {
		log.Printf("info: File %s Stopped!", filepath)
		pstream.Stop()
		GPIOOutPin("transmit", "off")
		//MyLedStripTransmitLEDOff()
		return
	}

	var eventSound EventSoundStruct = findEventSound("stream")
	if eventSound.Enabled {
		if pstream != nil && pstream.State() == gumbleffmpeg.StatePlaying {
			pstream.Stop()
			return
		}

		GPIOOutPin("transmit", "on")
		//MyLedStripTransmitLEDOn()

		IsPlayStream = true
		pstream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol/100)
		if err := pstream.Play(); err != nil {
			log.Printf("error: Can't play %s error %s", filepath, err)
		} else {
			log.Printf("info: File %s Playing!", filepath)
			pstream.Wait()
			pstream.Stop()
			GPIOOutPin("transmit", "off")
			//MyLedStripTransmitLEDOff()
		}
	} else {
		log.Println("warn: Sound Disabled by Config")
	}
}

func (b *Talkkonnect) splayIntoStream(filepath string, vol float32) {
	pstream = gumbleffmpeg.New(b.Stream.client, gumbleffmpeg.SourceFile(filepath), vol/100)
	if err := pstream.Play(); err != nil {
		log.Printf("error: Can't play %s error %s", filepath, err)
	} else {
		log.Printf("info: File %s Playing!\n", filepath)
		pstream.Wait()
		pstream.Stop()
	}
}

func (b *Talkkonnect) OpenStream() {
	if stream, err := b.New(b.Client); err != nil {

		if Config.Global.Hardware.TargetBoard == "rpi" {
			if LCDEnabled {
				LcdText = [4]string{"Stream Error!", "nil", "nil", "nil"}
				LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled {
				oledDisplay(false, 2, OLEDStartColumn, "Stream Error!!")
			}

		}
		FatalCleanUp("Stream Open Error " + err.Error())
	} else {
		b.Stream = stream
	}
}

func (b *Talkkonnect) ResetStream() {
	b.Stream.contextSink.Destroy()
	time.Sleep(50 * time.Millisecond)
	b.OpenStream()
}

func goStreamStats() {
	log.Println("debug: Active Streams")
	for item, value := range StreamTracker {
		log.Printf("debug: Item=%v UserID=%v UserName=%v Session=%v AudioStreamChannel=%v", item, value.UserID, value.UserName, value.UserSession, value.C)
	}
	log.Printf("debug: Total GoRoutines Open=%v, Total GoRoutines Wasted=%v \n", TotalStreams, NeedToKill)
}
