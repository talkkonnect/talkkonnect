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
	"fmt"
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
	errState = errors.New("gumbleopenal: invalid state")
	// errNoCaptureDevice is returned when transmit is requested but no microphone
	// could be opened. talkkonnect keeps running receive-only in that case.
	errNoCaptureDevice = errors.New("gumbleopenal: no capture device available")
	lcdtext            = [4]string{"nil", "nil", "nil", ""}
	now                = time.Now()
	TotalStreams       int
	NeedToKill         int

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

func openalInputDeviceName() string {
	return Config.Global.Software.Settings.OpenALInputDevice
}

func openalOutputDeviceName() string {
	return Config.Global.Software.Settings.OpenALOutputDevice
}

func openalDeviceLabel(name string) string {
	if name == "" {
		return "(default)"
	}
	return name
}

func openCaptureDevice(name string, frameSize int) (*openal.CaptureDevice, error) {
	device, err := openal.CaptureOpenDeviceChecked(name, gumble.AudioSampleRate, openal.FormatMono16, uint32(frameSize))
	if err != nil {
		log.Printf("error: OpenAL capture device %q failed: %v", openalDeviceLabel(name), err)
		log.Printf("info: Available OpenAL capture devices: %v", openal.CaptureDevices())
		return nil, fmt.Errorf("open capture device %q: %w", openalDeviceLabel(name), err)
	}
	opened := openal.GetDeviceString(&device.Device, openal.CaptureDeviceSpecifier)
	log.Printf("info: OpenAL capture device opened: %q (requested %q)", opened, openalDeviceLabel(name))
	return device, nil
}

func openPlaybackDevice(name string) (*openal.Device, *openal.Context, error) {
	device, err := openal.OpenDeviceChecked(name)
	if err != nil {
		log.Printf("error: OpenAL playback device %q failed: %v", openalDeviceLabel(name), err)
		log.Printf("info: Available OpenAL playback devices: %v", openal.PlaybackDevices())
		return nil, nil, fmt.Errorf("open playback device %q: %w", openalDeviceLabel(name), err)
	}
	contextSink := device.CreateContext()
	if !contextSink.Activate() {
		device.CloseDevice()
		return nil, nil, errors.New("failed to activate OpenAL playback context")
	}
	opened := openal.GetDeviceString(device, openal.AllDevicesSpecifier)
	if opened == "" {
		opened = openal.GetDeviceString(device, openal.DeviceSpecifier)
	}
	log.Printf("info: OpenAL playback device opened: %q (requested %q)", opened, openalDeviceLabel(name))
	return device, contextSink, nil
}

// LogOpenALDevices logs available OpenAL devices when troubleshooting audio routing.
func LogOpenALDevices() {
	log.Printf("info: OpenAL capture devices: %v", openal.CaptureDevices())
	log.Printf("info: OpenAL playback devices: %v", openal.PlaybackDevices())
	log.Printf("info: Using openalinputdevice=%q", openalDeviceLabel(openalInputDeviceName()))
	log.Printf("info: Using openaloutputdevice=%q", openalDeviceLabel(openalOutputDeviceName()))
	log.Printf("info: Using localplaybackdevice=%q", openalDeviceLabel(Config.Global.Software.Settings.LocalPlaybackDevice))
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

	// A missing microphone is not fatal: talkkonnect stays up receive-only and
	// retries opening the capture device on the next transmit request.
	var err error
	s.deviceSource, err = openCaptureDevice(openalInputDeviceName(), s.sourceFrameSize)
	if err != nil {
		log.Printf("warn: No audio capture device, continuing in receive-only mode (transmit disabled until a microphone is available): %v", err)
		s.deviceSource = nil
	}

	s.deviceSink, s.contextSink, err = openPlaybackDevice(openalOutputDeviceName())
	if err != nil {
		if s.deviceSource != nil {
			s.deviceSource.CaptureCloseDevice()
			s.deviceSource = nil
		}
		return nil, err
	}

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

// ensureCaptureDevice opens the microphone if it is not open yet. It is called on
// every transmit so a microphone plugged in after startup starts working without
// restarting talkkonnect.
func (s *Stream) ensureCaptureDevice() error {
	if s.deviceSource != nil {
		return nil
	}
	deviceSource, err := openCaptureDevice(openalInputDeviceName(), s.sourceFrameSize)
	if err != nil {
		return err
	}
	log.Println("info: Audio capture device now available, transmit enabled")
	s.deviceSource = deviceSource
	return nil
}

func (b *Talkkonnect) StartSource() error {
	// Check the microphone before the beep so a receive-only unit does not
	// transmit a beep it cannot follow with voice.
	if err := b.Stream.ensureCaptureDevice(); err != nil {
		log.Printf("error: No audio capture device, cannot transmit: %v", err)
		return errNoCaptureDevice
	}

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
	if b.Stream.deviceSource != nil {
		b.Stream.deviceSource.CaptureStop()
	}
	// Wait until sourceRoutine exits (roger tail + Mumble terminator on same AudioOutgoing).
	b.Stream.sourceWG.Wait()
	// Device remains open for next transmission - only stop capture
	if rogerBeepNeedsFallback() {
		eventSound := findEventSound("rogerbeep")
		GPIOOutPin("transmit", "on")
		log.Println("debug: Rogerbeep Playing (ffmpeg fallback)")
		if v, err := strconv.ParseFloat(eventSound.Volume, 32); err == nil {
			b.splayIntoStream(eventSound.FileName, float32(v))
		}
		GPIOOutPin("transmit", "off")
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
				if ScanIsRunning() {
					noteScanVoiceActivity(e.User.Channel.Name)
				}
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
				RecordRXAudioLevel(packet.AudioBuffer)
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
		if b.Stream.deviceSource != nil {
			b.Stream.deviceSource.CaptureCloseDevice()
			b.Stream.deviceSource = nil
		}
		b.Stream.sourceFrameSize = frameSize
		deviceSource, err := openCaptureDevice(openalInputDeviceName(), b.Stream.sourceFrameSize)
		if err != nil {
			log.Printf("error: Failed to reopen OpenAL capture device: %v", err)
			return
		}
		b.Stream.deviceSource = deviceSource
	}

	if b.Stream.deviceSource == nil {
		log.Println("error: No audio capture device, transmit aborted")
		return
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
			playRogerBeepTail(b.Stream.client, outgoing)
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
			RecordTXAudioLevel(int16Buffer)
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
	b.splayIntoStreamWithOptions(filepath, vol, 0, 0)
}

// splayIntoStreamWithOptions plays a file into the mumble stream. offsetSecs seeks
// into the file before playback starts and durationSecs cuts playback short;
// gumbleffmpeg only understands the offset, so the duration is enforced here with
// a timer that stops the stream.
func (b *Talkkonnect) splayIntoStreamWithOptions(filepath string, vol float32, offsetSecs float32, durationSecs float32) {
	pstream = gumbleffmpeg.New(b.Stream.client, gumbleffmpeg.SourceFile(filepath), vol/100)
	if offsetSecs > 0 {
		pstream.Offset = time.Duration(float64(offsetSecs) * float64(time.Second))
	}

	// Keep a local handle so the timer below cannot stop a later playback that has
	// meanwhile replaced the package level pstream.
	stream := pstream

	if err := stream.Play(); err != nil {
		log.Printf("error: Can't play %s error %s", filepath, err)
		return
	}

	log.Printf("info: File %s Playing!\n", filepath)

	if durationSecs > 0 {
		timer := time.AfterFunc(time.Duration(float64(durationSecs)*float64(time.Second)), func() {
			log.Printf("debug: stopping %s after the configured duration of %.1f seconds", filepath, durationSecs)
			stream.Stop()
		})
		defer timer.Stop()
	}

	stream.Wait()
	stream.Stop()
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
		FatalCleanUp("Stream Open Error (playback) " + err.Error())
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
