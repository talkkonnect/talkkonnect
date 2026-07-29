package talkkonnect

import (
	"encoding/binary"
	"math"
	"net"
	"testing"
	"time"
)

// TestG711RoundTrip checks that the u-law and A-law encoders decode back to a
// value close to the original (G.711 is lossy but must stay well correlated).
func TestG711RoundTrip(t *testing.T) {
	for _, v := range []int16{0, 100, -100, 1000, -1000, 8000, -8000, 32000, -32000, 32767, -32768} {
		u := ulawToLinear(linearToULaw(v))
		if tol := int(math.Abs(float64(v))/8) + 256; mcAbsDiff(u, v) > tol {
			t.Errorf("u-law round trip %d -> %d (tol %d)", v, u, tol)
		}
		a := alawToLinear(linearToALaw(v))
		if tol := int(math.Abs(float64(v))/8) + 256; mcAbsDiff(a, v) > tol {
			t.Errorf("A-law round trip %d -> %d (tol %d)", v, a, tol)
		}
	}
	// Canonical reference value: u-law of digital silence is 0xFF.
	if got := linearToULaw(0); got != 0xFF {
		t.Errorf("linearToULaw(0) = 0x%02X, want 0xFF", got)
	}
}

// TestRTPWireFormat sends real RTP packets over a loopback socket and verifies
// the header fields (V2, payload type, marker on the first packet of a spurt,
// incrementing sequence and 8 kHz timestamp) and that the payload decodes back to
// PCM. This is exactly the framing hardware PA endpoints consume, so it is the
// contract that must not drift.
func TestRTPWireFormat(t *testing.T) {
	rx, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer rx.Close()
	port := rx.LocalAddr().(*net.UDPAddr).Port

	dest := multicastDest{Address: "127.0.0.1", Port: port, Codec: "pcmu", TTL: 1}
	s, err := newRTPStream(dest, "", 96)
	if err != nil {
		t.Fatalf("newRTPStream: %v", err)
	}
	defer s.close()

	const spf = 160 // 20 ms @ 8 kHz
	frame := make([]int16, spf)
	for i := range frame {
		frame[i] = int16(8000 * math.Sin(2*math.Pi*440*float64(i)/8000))
	}

	const nPackets = 3
	for i := 0; i < nPackets; i++ {
		if err := s.writeFrame(frame); err != nil {
			t.Fatalf("writeFrame %d: %v", i, err)
		}
	}

	var prevSeq uint16
	var prevTS uint32
	buf := make([]byte, 2048)
	for i := 0; i < nPackets; i++ {
		_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := rx.ReadFromUDP(buf)
		if err != nil {
			t.Fatalf("read packet %d: %v", i, err)
		}
		if want := 12 + spf; n != want {
			t.Fatalf("packet %d length = %d, want %d", i, n, want)
		}
		if buf[0] != 0x80 {
			t.Errorf("packet %d byte0 = 0x%02X, want 0x80 (V2)", i, buf[0])
		}
		pt := buf[1] & 0x7F
		marker := buf[1]&0x80 != 0
		if pt != ptPCMU {
			t.Errorf("packet %d payload type = %d, want %d", i, pt, ptPCMU)
		}
		if i == 0 && !marker {
			t.Errorf("packet 0 missing the RTP marker bit")
		}
		if i > 0 && marker {
			t.Errorf("packet %d unexpectedly has the marker bit", i)
		}
		seq := binary.BigEndian.Uint16(buf[2:4])
		ts := binary.BigEndian.Uint32(buf[4:8])
		if i > 0 {
			if seq != prevSeq+1 {
				t.Errorf("packet %d seq = %d, want %d", i, seq, prevSeq+1)
			}
			if ts != prevTS+spf {
				t.Errorf("packet %d ts = %d, want %d", i, ts, prevTS+spf)
			}
		}
		prevSeq, prevTS = seq, ts

		got := mcDecodePayload(pt, buf[12:n])
		if len(got) != spf {
			t.Fatalf("decoded %d samples, want %d", len(got), spf)
		}
		if corr := mcCorrelation(frame, got); corr < 0.99 {
			t.Errorf("packet %d payload correlation = %.3f, want > 0.99", i, corr)
		}
	}

	// After a silent gap the sender must mark the next packet as a new spurt.
	s.startSpurt()
	if err := s.writeFrame(frame); err != nil {
		t.Fatalf("writeFrame after spurt reset: %v", err)
	}
	_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := rx.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read spurt packet: %v", err)
	}
	if n < 12 || buf[1]&0x80 == 0 {
		t.Error("first packet of a new spurt is missing the marker bit")
	}
}

// TestRTPCodecPayloads pins the payload type and packet size of each codec, the
// numbers a receiver is configured with.
func TestRTPCodecPayloads(t *testing.T) {
	cases := []struct {
		codec     string
		wantPT    byte
		wantBytes int // payload bytes for a 160 sample frame
	}{
		{"pcmu", ptPCMU, 160},
		{"", ptPCMU, 160}, // unset falls back to the interoperable codec
		{"pcma", ptPCMA, 160},
		{"l16", 96, 320},
	}
	frame := make([]int16, 160)
	for i := range frame {
		frame[i] = int16(i * 100)
	}
	for _, tc := range cases {
		pt, ok := mcPayloadTypeFor(tc.codec, 96)
		if !ok || pt != tc.wantPT {
			t.Errorf("payload type for %q = %d (ok %v), want %d", tc.codec, pt, ok, tc.wantPT)
		}
		if got := len(mcEncodePayload(mcNormalizeCodec(tc.codec), frame)); got != tc.wantBytes {
			t.Errorf("%q payload = %d bytes, want %d", tc.codec, got, tc.wantBytes)
		}
	}

	// L16 must go out big-endian, which is the opposite of talkkonnect's
	// internal little-endian PCM.
	out := mcEncodePayload("l16", []int16{0x0102})
	if out[0] != 0x01 || out[1] != 0x02 {
		t.Errorf("l16 payload = % X, want 01 02 (network byte order)", out)
	}
}

// TestApplyGainSaturates makes sure a hot mix distorts rather than wrapping into
// inverted phase.
func TestApplyGainSaturates(t *testing.T) {
	got := mcApplyGain([]int16{20000, -20000, 0}, 300)
	want := []int16{32767, -32768, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("gain 300%% sample %d = %d, want %d", i, got[i], want[i])
		}
	}
}

func mcAbsDiff(a, b int16) int {
	d := int(a) - int(b)
	if d < 0 {
		return -d
	}
	return d
}

// mcCorrelation returns the normalized cross-correlation of two equal length signals.
func mcCorrelation(a, b []int16) float64 {
	var sa, sb, sab float64
	for i := range a {
		x, y := float64(a[i]), float64(b[i])
		sa += x * x
		sb += y * y
		sab += x * y
	}
	if sa == 0 || sb == 0 {
		return 0
	}
	return sab / math.Sqrt(sa*sb)
}
