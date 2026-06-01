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
	"context"
	"encoding/binary"
	"errors"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
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

	rxBufferDropCount   uint64
	rxBufferDropLogMu   sync.Mutex
	rxBufferDropLastLog time.Time
)

// MumbleDuplex - listenera and outgoing
type MumbleDuplex struct{}

type Stream struct {
	client *gumble.Client
	link   gumble.Detacher

	deviceSource    *openal.CaptureDevice
	sourceFrameSize int
	sourceStop      chan bool
	sourceWG        sync.WaitGroup

	deviceSink  *openal.Device
	contextSink *openal.Context

	// connCtx is the connection-level context (child of daemon MasterCtx); stream goroutines derive from it.
	connCtx context.Context
}

func (b *Talkkonnect) New(client *gumble.Client) (*Stream, error) {
	connParent := b.ConnCtx
	if connParent == nil {
		connParent = b.MasterCtx
	}
	if connParent == nil {
		connParent = context.Background()
	}
	s := &Stream{
		client:          client,
		sourceFrameSize: client.Config.AudioFrameSize(),
		connCtx:         connParent,
	}
	s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.sourceFrameSize))

	s.deviceSink = openal.OpenDevice("")

	s.contextSink = s.deviceSink.CreateContext()

	s.contextSink.Activate()

	s.link = client.Config.AttachAudio(s)

	return s, nil
}

func (b *Talkkonnect) Destroy() {
	if b.Stream == nil {
		return
	}
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
	// Ensure device is valid before starting capture
	if b.Stream.deviceSource == nil {
		log.Println("error: Audio device is nil, cannot start capture")
		return errState
	}
	b.Stream.deviceSource.CaptureStart()
	b.Stream.sourceStop = make(chan bool)
	b.Stream.sourceWG.Add(1)
	SafeGo(func() { b.sourceRoutine() })
	return nil
}

func (b *Talkkonnect) StopSource() error {
	if b.Stream.sourceStop == nil {
		return errState
	}
	close(b.Stream.sourceStop)
	b.Stream.sourceStop = nil
	b.Stream.deviceSource.CaptureStop()
	// Wait until mic AudioOutgoing is closed (Mumble terminator) before opening another for roger beep.
	b.Stream.sourceWG.Wait()
	// Device remains open for next transmission - only stop capture
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

	return nil
}

func (s *Stream) OnAudioStream(e *gumble.AudioStreamEvent) {
	TotalStreams++
	connParent := s.connCtx
	if connParent == nil {
		connParent = context.Background()
	}
	streamCtx, streamCancel := context.WithCancel(connParent)

	// Track streams by Session, not UserID: unregistered users all share UserID=0 and
	// would otherwise collide, cancel the wrong goroutine, and block gumble's unbuffered
	// audio channel (stopping all incoming audio).
	session := e.User.Session
	streamTrackerMu.Lock()
	if prev, ok := StreamTracker[session]; ok {
		log.Printf("debug: Stale GoRoutine Detected For UserID=%v UserName=%v Session=%v AudioStreamChannel=%v", e.User.UserID, e.User.Name, session, e.C)
		NeedToKill++
		if prev.Cancel != nil {
			prev.Cancel()
		}
		// Keep draining the superseded channel so gumble's handler never blocks on send.
		if prev.C != nil {
			go drainAudioStream(prev.C)
		}
	}
	StreamTracker[session] = streamTrackerStruct{
		UserID:      e.User.UserID,
		UserName:    e.User.Name,
		UserSession: session,
		C:           e.C,
		Cancel:      streamCancel,
	}
	streamTrackerMu.Unlock()

	goStreamStats()

	SafeGo(func() {
		defer streamCancel()
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

		cleanup := func() {
			reclaim()
			emptyBufs.Delete()
			source.Delete()
			streamTrackerMu.Lock()
			if ent, ok := StreamTracker[session]; ok && ent.C == e.C {
				delete(StreamTracker, session)
			}
			streamTrackerMu.Unlock()
		}
		defer cleanup()

		for {
			select {
			case <-streamCtx.Done():
				return
			case packet, ok := <-e.C:
				if !ok {
					return
				}
				internetRadioNotifyVoiceOrTX()
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
					logRxBufferDrop(e.User.Name)
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
		}
	})
}

func (b *Talkkonnect) sourceRoutine() {
	defer b.Stream.sourceWG.Done()

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

	var connDone <-chan struct{}
	if b.Stream.connCtx != nil {
		connDone = b.Stream.connCtx.Done()
	}

	outgoing := b.Stream.client.AudioOutgoing()
	defer close(outgoing)

	for {
		select {
		case <-stop:
			return
		case <-connDone:
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

// drainAudioStream discards packets on a superseded channel so gumble's unbuffered
// ch <- send in handleAudio does not stall the entire connection.
func drainAudioStream(ch <-chan *gumble.AudioPacket) {
	for range ch {
	}
}

// logRxBufferDrop emits a rate-limited warning when the OpenAL RX buffer pool is exhausted.
func logRxBufferDrop(userName string) {
	atomic.AddUint64(&rxBufferDropCount, 1)
	rxBufferDropLogMu.Lock()
	defer rxBufferDropLogMu.Unlock()
	if time.Since(rxBufferDropLastLog) < 5*time.Second {
		return
	}
	dropped := atomic.SwapUint64(&rxBufferDropCount, 0)
	rxBufferDropLastLog = time.Now()
	log.Printf("warn: RX audio buffer pool exhausted, dropped %v packet(s) (user=%v)", dropped, userName)
}

func goStreamStats() {
	log.Println("debug: Active Streams")
	streamTrackerMu.Lock()
	for item, value := range StreamTracker {
		log.Printf("debug: Item=%v UserID=%v UserName=%v Session=%v AudioStreamChannel=%v", item, value.UserID, value.UserName, value.UserSession, value.C)
	}
	streamTrackerMu.Unlock()
	log.Printf("debug: Total GoRoutines Open=%v, Total GoRoutines Wasted=%v \n", TotalStreams, NeedToKill)
}
