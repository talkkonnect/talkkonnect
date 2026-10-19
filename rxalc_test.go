package talkkonnect

import (
	"encoding/binary"
	"math"
	"testing"
)

const testFrame = 480 // 10 ms at 48 kHz

func toneFrame(amplitude float64, frameIndex int) []int16 {
	out := make([]int16, testFrame)
	for i := range out {
		t := float64(frameIndex*testFrame+i) / 48000
		out[i] = int16(amplitude * math.Sin(2*math.Pi*440*t))
	}
	return out
}

func rmsDBFS(raw []byte) float64 {
	var sum float64
	n := len(raw) / 2
	for i := 0; i < n; i++ {
		f := float64(int16(binary.LittleEndian.Uint16(raw[i*2:])))
		sum += f * f
	}
	return 20 * math.Log10(math.Sqrt(sum/float64(n))/rxALCFullScale)
}

// runTone feeds 3 s of tone through a fresh ALC and returns the output level of the last frame.
func runTone(t *testing.T, user string, amplitude float64, p rxALCParams) float64 {
	t.Helper()
	rxALCGains.Delete(user)
	a := newRxALC(user)
	raw := make([]byte, testFrame*2)
	for f := 0; f < 300; f++ {
		a.process(toneFrame(amplitude, f), raw, p)
	}
	return rmsDBFS(raw)
}

func TestRXALCNormalizesLoudAndSoft(t *testing.T) {
	p := RXALCConfig{Enabled: true}.params()
	loud := runTone(t, "loud", 20000, p) // ~ -7 dBFS RMS
	soft := runTone(t, "soft", 2600, p)  // ~ -25 dBFS RMS
	for name, got := range map[string]float64{"loud": loud, "soft": soft} {
		if math.Abs(got-rxALCDefaultTargetDBFS) > 1 {
			t.Errorf("%s talker output %.1f dBFS, want %.1f ±1", name, got, rxALCDefaultTargetDBFS)
		}
	}
}

func TestRXALCRespectsMaxGain(t *testing.T) {
	p := RXALCConfig{Enabled: true}.params()
	in := 20 * math.Log10(300/math.Sqrt2/rxALCFullScale) // ~ -43 dBFS, needs +25 dB
	got := runTone(t, "veryquiet", 300, p)
	if math.Abs(got-(in+rxALCDefaultMaxGainDB)) > 1 {
		t.Errorf("very quiet talker output %.1f dBFS, want %.1f (capped at +%v dB)", got, in+rxALCDefaultMaxGainDB, rxALCDefaultMaxGainDB)
	}
}

func TestRXALCHoldsGainBelowGate(t *testing.T) {
	p := RXALCConfig{Enabled: true}.params()
	rxALCGains.Delete("gate")
	a := newRxALC("gate")
	raw := make([]byte, testFrame*2)
	for f := 0; f < 300; f++ {
		a.process(toneFrame(50, f), raw, p) // ~ -59 dBFS, below the -50 dBFS gate
	}
	if a.gain != 1.0 {
		t.Errorf("gain moved to %v on sub-gate noise, want 1.0", a.gain)
	}
}

func TestRXALCDoesNotMutateInput(t *testing.T) {
	p := RXALCConfig{Enabled: true}.params()
	in := toneFrame(2600, 0)
	orig := append([]int16(nil), in...)
	newRxALC("immut").process(in, make([]byte, len(in)*2), p)
	for i := range in {
		if in[i] != orig[i] {
			t.Fatal("process modified its input buffer")
		}
	}
}

func TestRXALCRemembersGainPerUser(t *testing.T) {
	p := RXALCConfig{Enabled: true}.params()
	runTone(t, "repeat", 2600, p)
	if g := newRxALC("repeat").gain; g <= 1.5 {
		t.Errorf("new stream for known soft talker starts at gain %v, want learned boost", g)
	}
}
