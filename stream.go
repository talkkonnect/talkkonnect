package talkkonnect

import (
	"encoding/binary"
	"errors"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/go-openal/openal"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/gumble/gumble"
	"log"
	"time"
)

var (
	ErrState         = errors.New("gumbleopenal: invalid state")
	lastspeaker      = "Nil"
	lcdtext          = [4]string{"nil", "nil", "nil", ""} //global variable declaration for 4 lines of LCD
	BackLightLED     = gpio.NewOutput(uint(BackLightLEDPin), false)
	VoiceActivityLED = gpio.NewOutput(VoiceActivityLEDPin, false)
	now              = time.Now()
	LastTime         = now.Unix()
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
	log.Println("alert: Destroy Source")
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

func (s *Stream) StartSourceFile() error {
	log.Println("alert: Start Source File")
	if s.sourceStop != nil {
		return ErrState
	}
	s.deviceSource.CaptureStart()
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	return nil
}

func (s *Stream) StartSource() error {
	log.Println("alert: Start Source")
	if s.sourceStop != nil {
		return ErrState
	}
	s.deviceSource.CaptureStart()
	s.sourceStop = make(chan bool)
	go s.sourceRoutine()
	return nil
}

func (s *Stream) StopSource() error {
	log.Println("alert: Stop Source File")
	if s.sourceStop == nil {
		return ErrState
	}
	close(s.sourceStop)
	s.sourceStop = nil
	s.deviceSource.CaptureStop()

	// on stop source adjust from 100ms to 25 ms helps reduce recovery time and helps with audio chopping (suvir kumar)
	time.Sleep(25 * time.Millisecond)

	s.deviceSource.CaptureCloseDevice()
	s.deviceSource = nil

	s.deviceSource = openal.CaptureOpenDevice("", gumble.AudioSampleRate, openal.FormatMono16, uint32(s.sourceFrameSize))

	return nil
}

func (s *Stream) OnAudioStream(e *gumble.AudioStreamEvent) {

	if TargetBoard == "rpi" {
		LEDOffFunc(VoiceActivityLED)
		LEDOffFunc(BackLightLED)
	}

	timertalkled := time.NewTimer(time.Millisecond * 200)

	var watchpin = true

	go func() {
		for watchpin {
			<-timertalkled.C
			if TargetBoard == "rpi" {
				LEDOffFunc(VoiceActivityLED)
			}
			lastspeaker = "Nil"
		}
	}()

	go func() {
		source := openal.NewSource()
		emptyBufs := openal.NewBuffers(8)
		reclaim := func() {
			if n := source.BuffersProcessed(); n > 0 {
				reclaimedBufs := make(openal.Buffers, n)
				source.UnqueueBuffers(reclaimedBufs)
				emptyBufs = append(emptyBufs, reclaimedBufs...)
			}
		}
		var raw [gumble.AudioMaximumFrameSize * 2]byte

		for packet := range e.C {
			samples := len(packet.AudioBuffer)
			if TargetBoard == "rpi" {
				LEDOnFunc(VoiceActivityLED)
				LEDOnFunc(BackLightLED)
			}

			timertalkled.Reset(time.Second)
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
				now = time.Now()
				if LastTime != now.Unix() {
					log.Println("alert: Source State is", source.State())
					now = time.Now()
					LastTime = now.Unix()
				}

				source.Play()
				if lastspeaker != e.User.Name {
					log.Println("Speaking:", e.User.Name)
					lastspeaker = e.User.Name
					t := time.Now()
					if TargetBoard == "rpi" {
						lcdtext = [4]string{"nil", "", "", e.User.Name + " " + t.Format("15:04:05")}
						go hd44780.LcdDisplay(lcdtext, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
						BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)
					}
				}
			}
		}
		watchpin = false
		reclaim()
		//emptyBufs.Delete()
		//source.Delete()
	}()
}

func (s *Stream) sourceRoutine() {
	interval := s.client.Config.AudioInterval
	frameSize := s.client.Config.AudioFrameSize()

	if frameSize != s.sourceFrameSize {
		log.Println("alert: FrameSize Error!")
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
			log.Println("alert: Ticker Stop!")
			return
		case <-ticker.C:
			//this is for encofing (transmitting)
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
