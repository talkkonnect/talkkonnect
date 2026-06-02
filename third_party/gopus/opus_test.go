//go:build cgo

package gopus_test

import (
	"testing"

	"github.com/talkkonnect/gopus"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	const sampleRate = 48000
	const channels = 1
	const frameSize = 960

	enc, err := gopus.NewEncoder(sampleRate, channels, gopus.Voip)
	if err != nil {
		t.Fatalf("NewEncoder: %v", err)
	}
	enc.SetBitrate(gopus.BitrateMaximum)

	dec, err := gopus.NewDecoder(sampleRate, channels)
	if err != nil {
		t.Fatalf("NewDecoder: %v", err)
	}

	pcm := make([]int16, frameSize*channels)
	for i := range pcm {
		pcm[i] = int16(i % 32767)
	}

	encoded, err := enc.Encode(pcm, frameSize, 4000)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("Encode returned empty packet")
	}

	decoded, err := dec.Decode(encoded, frameSize, false)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(decoded) == 0 {
		t.Fatal("Decode returned empty PCM")
	}
}
