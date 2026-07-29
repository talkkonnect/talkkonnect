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
 */

package talkkonnect

// multicastrtp.go carries the on-the-wire half of multicast paging: G.711 codecs
// and a native RTP sender. It is a port of gochimesd's daemon/rtp.go so both
// programs put the identical framing on the network.
//
// Standard IP PA endpoints (CyberData, Algo, Barix, Advanced Network Devices)
// and SIP desk phones (Yealink) join a multicast group and decode RTP carrying
// G.711 (u-law/A-law) at 8 kHz mono, 20 ms per packet. That is the only format
// they accept: L16 rides a *dynamic* payload type and G.711-only receivers drop
// it silently, so the group transmits correctly yet the speaker stays mute.
//
// The G.711 conversion tables and the segment search below are transcriptions of
// the canonical ITU-T G.711 reference implementation (the public-domain Sun
// g711.c), so they interoperate bit-exactly with hardware codecs.

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

// RTP static payload types (RFC 3551).
const (
	ptPCMU = 0 // G.711 u-law
	ptPCMA = 8 // G.711 A-law
)

// multicastRTPClockRate is the RTP clock, and therefore the PCM sample rate, of
// every codec talkkonnect sends. 8 kHz is what hardware decoders expect and what
// gochimesd emits, L16 included, so one receiver configuration serves both.
const multicastRTPClockRate = 8000

// multicastDest is one RTP destination: a multicast group and port with its own
// codec, TTL and software gain. The sender takes a slice of these so more groups
// are a configuration change rather than a code change.
type multicastDest struct {
	Address string
	Port    int
	Codec   string // pcmu | pcma | l16
	TTL     int
	Gain    int // software volume percent, 100 = unity
}

// Addr returns the address:port destination string, for logs.
func (d multicastDest) Addr() string { return fmt.Sprintf("%s:%d", d.Address, d.Port) }

// --- G.711 codec (ITU-T reference, public-domain Sun g711.c) ----------------

var (
	segEndULaw = [8]int32{0x3F, 0x7F, 0xFF, 0x1FF, 0x3FF, 0x7FF, 0xFFF, 0x1FFF}
	segEndALaw = [8]int32{0x1F, 0x3F, 0x7F, 0xFF, 0x1FF, 0x3FF, 0x7FF, 0xFFF}
)

func mcSegmentSearch(val int32, table *[8]int32) int32 {
	for i := int32(0); i < 8; i++ {
		if val <= table[i] {
			return i
		}
	}
	return 8
}

// linearToULaw encodes a 16-bit PCM sample to 8-bit G.711 u-law.
func linearToULaw(pcm int16) byte {
	const bias = 0x84
	const clip = 8159
	val := int32(pcm) >> 2 // 16-bit -> 14-bit
	var mask int32
	if val < 0 {
		val = -val
		mask = 0x7F
	} else {
		mask = 0xFF
	}
	if val > clip {
		val = clip
	}
	val += bias >> 2
	seg := mcSegmentSearch(val, &segEndULaw)
	if seg >= 8 {
		return byte(0x7F ^ mask)
	}
	uval := (seg << 4) | ((val >> (seg + 1)) & 0x0F)
	return byte(uval ^ mask)
}

// ulawToLinear decodes an 8-bit u-law value back to a 16-bit PCM sample. Only
// the tests need it, but it is what makes the encoder verifiable.
func ulawToLinear(u byte) int16 {
	const bias = 0x84
	u = ^u
	t := ((int32(u) & 0x0F) << 3) + bias
	t <<= (int32(u) & 0x70) >> 4
	if u&0x80 != 0 {
		return int16(bias - t)
	}
	return int16(t - bias)
}

// linearToALaw encodes a 16-bit PCM sample to 8-bit G.711 A-law.
func linearToALaw(pcm int16) byte {
	val := int32(pcm) >> 3 // 16-bit -> 13-bit
	var mask int32
	if val >= 0 {
		mask = 0xD5
	} else {
		mask = 0x55
		val = -val - 1
	}
	seg := mcSegmentSearch(val, &segEndALaw)
	if seg >= 8 {
		return byte(0x7F ^ mask)
	}
	aval := seg << 4
	if seg < 2 {
		aval |= (val >> 1) & 0x0F
	} else {
		aval |= (val >> seg) & 0x0F
	}
	return byte(aval ^ mask)
}

// alawToLinear decodes an 8-bit A-law value back to a 16-bit PCM sample.
func alawToLinear(a byte) int16 {
	a ^= 0x55
	t := (int32(a) & 0x0F) << 4
	seg := (int32(a) & 0x70) >> 4
	switch seg {
	case 0:
		t += 8
	case 1:
		t += 0x108
	default:
		t += 0x108
		t <<= uint(seg - 1)
	}
	if a&0x80 != 0 {
		return int16(t)
	}
	return int16(-t)
}

// --- RTP multicast sender ---------------------------------------------------

// rtpStream is a single RTP/UDP multicast sender bound to one destination.
type rtpStream struct {
	pc    *ipv4.PacketConn
	dst   *net.UDPAddr
	codec string
	pt    byte
	gain  int
	ssrc  uint32
	seq   uint16
	ts    uint32
	first bool // set the RTP marker bit on the first packet of a talk spurt
	hdr   [12]byte
	pkt   []byte // reused packet buffer, header + payload
}

// newRTPStream opens a multicast UDP socket for dest, applies TTL and egress
// interface, and seeds a random SSRC / sequence / timestamp per RFC 3550.
//
// The socket is deliberately unconnected (ListenUDP on an ephemeral port rather
// than DialUDP) with the group supplied per write, which is what lets one socket
// serve one destination while TTL and interface are set through ipv4.PacketConn.
func newRTPStream(dest multicastDest, iface string, l16PayloadType int) (*rtpStream, error) {
	ip := net.ParseIP(dest.Address)
	if ip == nil || ip.To4() == nil {
		return nil, fmt.Errorf("invalid multicast address %q", dest.Address)
	}
	uc, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, fmt.Errorf("open udp socket: %w", err)
	}
	pc := ipv4.NewPacketConn(uc)

	ttl := dest.TTL
	if ttl <= 0 {
		ttl = 1
	}
	if err := pc.SetMulticastTTL(ttl); err != nil {
		_ = pc.Close()
		return nil, fmt.Errorf("set multicast ttl: %w", err)
	}

	// A mistyped interface name is otherwise the hardest failure here to
	// diagnose: the kernel quietly sends out of the default route instead, so
	// packets exist on the wire but never reach the PA VLAN. Log and carry on.
	if iface != "" {
		if ifi, err := net.InterfaceByName(iface); err != nil {
			log.Printf("warn: multicast interface %q not found, using the default route: %v", iface, err)
		} else if err := pc.SetMulticastInterface(ifi); err != nil {
			log.Printf("warn: cannot send multicast out of interface %q, using the default route: %v", iface, err)
		}
	}

	pt, ok := mcPayloadTypeFor(dest.Codec, l16PayloadType)
	if !ok {
		_ = pc.Close()
		return nil, fmt.Errorf("unsupported codec %q", dest.Codec)
	}

	gain := dest.Gain
	if gain <= 0 {
		gain = 100
	}

	s := &rtpStream{
		pc:    pc,
		dst:   &net.UDPAddr{IP: ip, Port: dest.Port},
		codec: mcNormalizeCodec(dest.Codec),
		pt:    pt,
		gain:  gain,
		first: true,
	}
	s.ssrc = mcRandUint32()
	s.seq = uint16(mcRandUint32())
	s.ts = mcRandUint32()
	return s, nil
}

// writeFrame encodes one PCM frame and sends it as a single RTP packet, applying
// this destination's software gain on a copy so the shared frame that fans out to
// the other destinations is never mutated.
func (s *rtpStream) writeFrame(pcm []int16) error {
	if s.gain != 100 {
		pcm = mcApplyGain(pcm, s.gain)
	}
	payload := mcEncodePayload(s.codec, pcm)

	// 12-byte RTP header (RFC 3550): V=2, P=0, X=0, CC=0.
	s.hdr[0] = 0x80
	s.hdr[1] = s.pt
	if s.first {
		s.hdr[1] |= 0x80 // marker bit, first packet of a spurt
		s.first = false
	}
	binary.BigEndian.PutUint16(s.hdr[2:4], s.seq)
	binary.BigEndian.PutUint32(s.hdr[4:8], s.ts)
	binary.BigEndian.PutUint32(s.hdr[8:12], s.ssrc)

	// One reused buffer rather than a fresh allocation per packet: this runs 50
	// times a second per destination for as long as anyone is talking.
	need := len(s.hdr) + len(payload)
	if cap(s.pkt) < need {
		s.pkt = make([]byte, need)
	}
	s.pkt = s.pkt[:need]
	copy(s.pkt, s.hdr[:])
	copy(s.pkt[len(s.hdr):], payload)

	if _, err := s.pc.WriteTo(s.pkt, nil, s.dst); err != nil {
		return err
	}
	s.seq++
	s.ts += uint32(len(pcm)) // one timestamp tick per sample at 8 kHz
	return nil
}

// startSpurt marks the next packet as the start of a talk spurt, which is how a
// hardware decoder knows to reset its jitter buffer after a silent gap.
func (s *rtpStream) startSpurt() { s.first = true }

func (s *rtpStream) close() {
	if s.pc != nil {
		_ = s.pc.Close()
	}
}

// mcApplyGain returns a copy of pcm scaled by gainPercent/100 with saturation.
func mcApplyGain(pcm []int16, gainPercent int) []int16 {
	if gainPercent < 0 {
		gainPercent = 0
	}
	out := make([]int16, len(pcm))
	for i, s := range pcm {
		out[i] = mcClampToInt16(int32(s) * int32(gainPercent) / 100)
	}
	return out
}

// mcClampToInt16 saturates instead of wrapping, so a loud mix distorts rather
// than inverting phase.
func mcClampToInt16(v int32) int16 {
	if v > 32767 {
		return 32767
	}
	if v < -32768 {
		return -32768
	}
	return int16(v)
}

// mcEncodePayload converts a PCM frame to the destination's codec bytes.
func mcEncodePayload(codec string, pcm []int16) []byte {
	switch codec {
	case "pcma":
		out := make([]byte, len(pcm))
		for i, s := range pcm {
			out[i] = linearToALaw(s)
		}
		return out
	case "l16":
		// RFC 3551: L16 is network byte order, unlike the little-endian PCM
		// talkkonnect handles internally.
		out := make([]byte, len(pcm)*2)
		for i, s := range pcm {
			binary.BigEndian.PutUint16(out[i*2:], uint16(s))
		}
		return out
	default: // pcmu
		out := make([]byte, len(pcm))
		for i, s := range pcm {
			out[i] = linearToULaw(s)
		}
		return out
	}
}

// mcDecodePayload converts received RTP payload bytes back to PCM. talkkonnect
// only sends, so this exists to make the wire format testable.
func mcDecodePayload(pt byte, payload []byte) []int16 {
	switch pt {
	case ptPCMA:
		out := make([]int16, len(payload))
		for i, b := range payload {
			out[i] = alawToLinear(b)
		}
		return out
	case ptPCMU:
		out := make([]int16, len(payload))
		for i, b := range payload {
			out[i] = ulawToLinear(b)
		}
		return out
	default: // any dynamic payload type is L16 big-endian
		n := len(payload) / 2
		out := make([]int16, n)
		for i := 0; i < n; i++ {
			out[i] = int16(binary.BigEndian.Uint16(payload[i*2:]))
		}
		return out
	}
}

func mcPayloadTypeFor(codec string, l16PayloadType int) (byte, bool) {
	switch mcNormalizeCodec(codec) {
	case "pcmu":
		return ptPCMU, true
	case "pcma":
		return ptPCMA, true
	case "l16":
		if l16PayloadType <= 0 {
			l16PayloadType = 96
		}
		return byte(l16PayloadType), true
	default:
		return 0, false
	}
}

// mcNormalizeCodec accepts the spellings people write in the XML and falls back
// to the interoperable one.
func mcNormalizeCodec(codec string) string {
	switch codec {
	case "pcma", "alaw", "g711a":
		return "pcma"
	case "l16", "pcm", "raw":
		return "l16"
	default:
		return "pcmu"
	}
}

// mcRandUint32 returns a cryptographically random 32-bit value for the SSRC and
// the initial sequence number and timestamp, as RFC 3550 requires.
func mcRandUint32() uint32 {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0x1a2b3c4d // deterministic fallback; a collision is harmless here
	}
	return binary.BigEndian.Uint32(b[:])
}
