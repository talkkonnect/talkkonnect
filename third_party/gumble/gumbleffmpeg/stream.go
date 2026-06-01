package gumbleffmpeg // import "github.com/talkkonnect/gumble/gumbleffmpeg"

import (
	"encoding/binary"
	"errors"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"github.com/talkkonnect/gumble/gumble"
)

// State represents the state of a Stream.
type State int32

// Valid states of Stream.
const (
	StateInitial State = iota + 1
	StatePlaying
	StatePaused
	StateStopped
)

// Stream is an audio stream that encodes media through ffmpeg and sends it to
// the server.
//
// A stream can only be used once; it cannot be started after it is stopped.
type Stream struct {
	// Command to execute to play the file. Defaults to "ffmpeg".
	Command string
	// Playback volume (can be changed while the source is playing).
	Volume float32
	// Audio source (cannot be changed after stream starts).
	Source Source
	// Starting offset.
	Offset time.Duration

	client  *gumble.Client
	cmd     *exec.Cmd
	pipe    io.ReadCloser
	pause   chan struct{}
	elapsed int64

	state State

	l  sync.Mutex
	wg sync.WaitGroup
}

// New returns a new Stream for the given gumble Client and Source.
func New(client *gumble.Client, source Source,vol float32) *Stream {
	return &Stream{
		client:  client,
		Volume:  vol,
		Source:  source,
		Command: "ffmpeg",
		pause:   make(chan struct{}),
		state:   StateInitial,
	}
}

// Play begins playing
func (s *Stream) Play() error {
	s.l.Lock()
	defer s.l.Unlock()

	switch s.state {
	case StatePaused:
		s.state = StatePlaying
		go s.process()
		return nil
	case StatePlaying:
		return errors.New("gumbleffmpeg: stream already playing")
	case StateStopped:
		return errors.New("gumbleffmpeg: stream has stopped")
	}

	// fresh stream
	if s.Source == nil {
		return errors.New("gumbleffmpeg: nil source")
	}

	args := s.Source.arguments()
	if s.Offset > 0 {
		args = append([]string{"-ss", strconv.FormatFloat(s.Offset.Seconds(), 'f', -1, 64)}, args...)
	}

	args = append(args, "-ac", strconv.Itoa(gumble.AudioChannels), "-ar", strconv.Itoa(gumble.AudioSampleRate), "-f", "s16le", "-")
	cmd := exec.Command(s.Command, args...)
	var err error
	s.pipe, err = cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := s.Source.start(cmd); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		s.Source.done()
		return err
	}
	s.wg.Add(1)
	s.cmd = cmd
	s.state = StatePlaying
	go s.process()
	return nil
}

// State returns the state of the stream.
func (s *Stream) State() State {
	s.l.Lock()
	defer s.l.Unlock()
	return s.state
}

// Pause pauses a playing stream.
func (s *Stream) Pause() error {
	s.l.Lock()
	if s.state != StatePlaying {
		s.l.Unlock()
		return errors.New("gumbleffmpeg: stream is not playing")
	}
	s.state = StatePaused
	s.l.Unlock()
	s.pause <- struct{}{}
	return nil
}

// Stop stops the stream.
func (s *Stream) Stop() error {
	s.l.Lock()
	switch s.state {
	case StateStopped, StateInitial:
		s.l.Unlock()
		return errors.New("gumbleffmpeg: stream is not playing nor paused")
	}
	s.cleanup()
	s.Wait()
	return nil
}

// Wait returns once the stream has stopped playing.
func (s *Stream) Wait() {
	s.wg.Wait()
}

// Elapsed returns the amount of audio that has been played by the stream.
func (s *Stream) Elapsed() time.Duration {
	return time.Duration(atomic.LoadInt64(&s.elapsed))
}

func (s *Stream) process() {
	// s.state has been set to StatePlaying

	interval := s.client.Config.AudioInterval
	frameSize := s.client.Config.AudioFrameSize()

	byteBuffer := make([]byte, frameSize*2)

	outgoing := s.client.AudioOutgoing()
	defer close(outgoing)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.pause:
			return
		case <-ticker.C:
			n, err := io.ReadFull(s.pipe, byteBuffer)
			if err != nil {
				if (err == io.EOF || err == io.ErrUnexpectedEOF) && n > 0 {
					for i := n; i < len(byteBuffer); i++ {
						byteBuffer[i] = 0
					}
				} else {
					s.l.Lock()
					s.cleanup()
					return
				}
			}
			int16Buffer := make([]int16, frameSize)
			for i := range int16Buffer {
				float := float32(int16(binary.LittleEndian.Uint16(byteBuffer[i*2 : (i+1)*2])))
				int16Buffer[i] = int16(s.Volume * float)
			}
			atomic.AddInt64(&s.elapsed, int64(interval))
			outgoing <- gumble.AudioBuffer(int16Buffer)
		}
	}
}

func (s *Stream) cleanup() {
	defer s.l.Unlock()
	// s.l has been acquired
	if s.state == StateStopped {
		return
	}
	s.cmd.Process.Kill()
	s.cmd.Wait()
	s.Source.done()
	for len(s.pause) > 0 {
		<-s.pause
	}
	s.state = StateStopped
	s.wg.Done()
}
