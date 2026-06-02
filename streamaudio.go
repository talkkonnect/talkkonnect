/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Preloaded PCM for low-latency roger beep injection on the transmit audio stream.
 */

package talkkonnect

import (
	"bytes"
	"encoding/binary"
	"log"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

var (
	rogerBeepMu     sync.RWMutex
	rogerBeepPCM    []int16
	rogerBeepVolume float32 = 1
)

// preloadEventSounds decodes enabled hot-path event sounds into memory (48 kHz mono PCM).
func preloadEventSounds() {
	eventSound := findEventSound("rogerbeep")
	if !eventSound.Enabled || eventSound.FileName == "" {
		rogerBeepMu.Lock()
		rogerBeepPCM = nil
		rogerBeepMu.Unlock()
		return
	}

	vol := float32(1)
	if v, err := strconv.ParseFloat(eventSound.Volume, 32); err == nil {
		vol = float32(v) / 100
	}

	pcm, err := loadPCMSoundViaFFmpeg(eventSound.FileName)
	if err != nil {
		log.Printf("warn: roger beep preload failed for %s: %v (will use ffmpeg fallback on TX end)", eventSound.FileName, err)
		rogerBeepMu.Lock()
		rogerBeepPCM = nil
		rogerBeepVolume = vol
		rogerBeepMu.Unlock()
		return
	}

	rogerBeepMu.Lock()
	rogerBeepPCM = pcm
	rogerBeepVolume = vol
	rogerBeepMu.Unlock()
	log.Printf("info: Preloaded roger beep %s (%d samples, %.0f ms)", eventSound.FileName, len(pcm), float64(len(pcm))/float64(gumble.AudioSampleRate)*1000)
}

func loadPCMSoundViaFFmpeg(path string) ([]int16, error) {
	ffmpeg := "ffmpeg"
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		ffmpeg = p
	} else if _, err := exec.LookPath("/usr/bin/ffmpeg"); err == nil {
		ffmpeg = "/usr/bin/ffmpeg"
	}

	args := []string{
		"-nostdin", "-loglevel", "error",
		"-i", path,
		"-ac", "1",
		"-ar", strconv.Itoa(gumble.AudioSampleRate),
		"-f", "s16le",
		"-",
	}
	cmd := exec.Command(ffmpeg, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	raw := stdout.Bytes()
	if len(raw)%2 != 0 {
		raw = raw[:len(raw)-1]
	}
	samples := make([]int16, len(raw)/2)
	for i := range samples {
		samples[i] = int16(binary.LittleEndian.Uint16(raw[i*2 : i*2+2]))
	}
	return samples, nil
}

// rogerBeepCached reports whether inline roger beep PCM is ready.
func rogerBeepCached() ([]int16, float32, bool) {
	rogerBeepMu.RLock()
	defer rogerBeepMu.RUnlock()
	if len(rogerBeepPCM) == 0 {
		return nil, 0, false
	}
	pcm := make([]int16, len(rogerBeepPCM))
	copy(pcm, rogerBeepPCM)
	return pcm, rogerBeepVolume, true
}

// rogerBeepNeedsFallback is true when roger is enabled but preload did not succeed.
func rogerBeepNeedsFallback() bool {
	eventSound := findEventSound("rogerbeep")
	if !eventSound.Enabled {
		return false
	}
	rogerBeepMu.RLock()
	ok := len(rogerBeepPCM) > 0
	rogerBeepMu.RUnlock()
	return !ok
}

// pumpPCMFrames sends PCM to an open AudioOutgoing channel at the client's frame interval.
func pumpPCMFrames(client *gumble.Client, outgoing chan<- gumble.AudioBuffer, pcm []int16, volume float32) {
	if client == nil || outgoing == nil || len(pcm) == 0 {
		return
	}

	interval := client.Config.AudioInterval
	frameSize := client.Config.AudioFrameSize()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	frame := make([]int16, frameSize)
	offset := 0
	first := true

	for offset < len(pcm) {
		if !first {
			<-ticker.C
		}
		first = false

		n := frameSize
		if offset+n > len(pcm) {
			n = len(pcm) - offset
		}
		copy(frame, pcm[offset:offset+n])
		for i := n; i < frameSize; i++ {
			frame[i] = 0
		}
		for i := range frame {
			frame[i] = int16(volume * float32(frame[i]))
		}
		outgoing <- gumble.AudioBuffer(append([]int16(nil), frame...))
		offset += n
		if n < frameSize {
			break
		}
	}
}

// playRogerBeepTail injects the cached roger beep on the same outgoing stream before it closes.
func playRogerBeepTail(client *gumble.Client, outgoing chan<- gumble.AudioBuffer) bool {
	pcm, vol, ok := rogerBeepCached()
	if !ok {
		return false
	}
	GPIOOutPin("transmit", "on")
	log.Println("debug: Rogerbeep Playing (inline PCM)")
	pumpPCMFrames(client, outgoing, pcm, vol)
	GPIOOutPin("transmit", "off")
	return true
}
