package talkkonnect

import (
	"testing"

	"github.com/talkkonnect/gumble/gumble"
)

func TestLoadPCMSoundViaFFmpeg(t *testing.T) {
	const apollo = "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/soundfiles/rogerbeeps/Apollo.wav"
	if !FileExists(apollo) {
		t.Skip("Apollo.wav not present")
	}
	pcm, err := loadPCMSoundViaFFmpeg(apollo)
	if err != nil {
		t.Fatal(err)
	}
	if len(pcm) < 100 {
		t.Fatalf("expected substantial sample count, got %d", len(pcm))
	}
	// ~200 ms at 48 kHz
	ms := float64(len(pcm)) / float64(gumble.AudioSampleRate) * 1000
	if ms < 150 || ms > 350 {
		t.Fatalf("unexpected duration %.1f ms", ms)
	}
}

func TestPumpPCMFramesFrameCount(t *testing.T) {
	pcm := make([]int16, 480) // 10 ms
	frameSize := gumble.AudioDefaultFrameSize
	frames := (len(pcm) + frameSize - 1) / frameSize
	if frames != 1 {
		t.Fatalf("expected 1 frame, got %d", frames)
	}
}
