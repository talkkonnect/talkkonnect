package talkkonnect

import (
	"context"
	"encoding/xml"
	"math"
	"net"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

// multicastTestDoc exercises the shapes the XML has to support: the enabled
// attribute on the section, scalar elements, the two user lists, and the
// per-profile multicast flag under <multimedia>.
const multicastTestDoc = `<?xml version="1.0" encoding="UTF-8"?>
<document>
  <global>
    <software>
      <multicast enabled="true">
        <group>239.0.1.10</group>
        <port>5004</port>
        <codec>pcmu</codec>
        <ttl>2</ttl>
        <interface>ens18</interface>
        <packetms>20</packetms>
        <l16payloadtype>96</l16payloadtype>
        <volume>80</volume>
        <allchannels>true</allchannels>
        <hangoverms>300</hangoverms>
        <include>
          <user>zoran-laptop</user>
          <user>Suvir</user>
        </include>
        <exclude>
          <user>noisybox</user>
        </exclude>
      </multicast>
    </software>
    <multimedia>
      <id value="page_all" enabled="true">
        <params>
          <localplay>false</localplay>
          <playintostream>false</playintostream>
          <multicast>true</multicast>
        </params>
      </id>
    </multimedia>
  </global>
</document>`

func TestMulticastConfigParse(t *testing.T) {
	var cfg ConfigStruct
	if err := xml.Unmarshal([]byte(multicastTestDoc), &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	mc := cfg.Global.Software.Multicast
	if !mc.Enabled {
		t.Error("enabled attribute did not parse")
	}
	if mc.Group != "239.0.1.10" || mc.Port != 5004 || mc.Codec != "pcmu" {
		t.Errorf("destination = %v:%v codec %v", mc.Group, mc.Port, mc.Codec)
	}
	if mc.TTL != 2 || mc.PacketMS != 20 || mc.L16PayloadType != 96 || mc.Volume != 80 || mc.HangoverMS != 300 {
		t.Errorf("scalars = ttl %v packetms %v l16pt %v volume %v hangover %v",
			mc.TTL, mc.PacketMS, mc.L16PayloadType, mc.Volume, mc.HangoverMS)
	}
	if !mc.AllChannels {
		t.Error("allchannels did not parse")
	}
	if len(mc.Include.User) != 2 || mc.Include.User[0] != "zoran-laptop" {
		t.Errorf("include = %v", mc.Include.User)
	}
	if len(mc.Exclude.User) != 1 || mc.Exclude.User[0] != "noisybox" {
		t.Errorf("exclude = %v", mc.Exclude.User)
	}
	if len(cfg.Global.Multimedia.ID) != 1 || !cfg.Global.Multimedia.ID[0].Params.Multicast {
		t.Error("per profile <multicast> did not parse")
	}
}

// TestMulticastConfigDefaults checks the snapshot fills in what the XML leaves out.
func TestMulticastConfigDefaults(t *testing.T) {
	restore := multicastTestConfig(t, "239.0.1.10", 5004)
	defer restore()

	Config.Global.Software.Multicast.PacketMS = 0
	Config.Global.Software.Multicast.L16PayloadType = 0
	Config.Global.Software.Multicast.HangoverMS = 0
	Config.Global.Software.Multicast.Volume = 0
	multicastConfigureFromXML()

	cfg := multicastConfigSnapshot()
	if cfg.PacketMS != multicastDefaultPacketMS || cfg.SamplesPerFrame() != 160 {
		t.Errorf("packetms = %v (%v samples), want %v (160)", cfg.PacketMS, cfg.SamplesPerFrame(), multicastDefaultPacketMS)
	}
	if cfg.L16PayloadType != multicastDefaultL16PT || cfg.HangoverMS != multicastDefaultHangoverMS {
		t.Errorf("l16pt = %v hangover = %v", cfg.L16PayloadType, cfg.HangoverMS)
	}
	if len(cfg.Dests) != 1 || cfg.Dests[0].Gain != multicastDefaultVolume {
		t.Errorf("dests = %+v, want one at unity gain", cfg.Dests)
	}
	if got := cfg.hangoverTicks(); got != multicastDefaultHangoverMS/multicastDefaultPacketMS {
		t.Errorf("hangoverTicks = %v", got)
	}
}

// TestMulticastSanityChecksRepairConfig covers the values a hand written config
// most often gets wrong.
func TestMulticastSanityChecksRepairConfig(t *testing.T) {
	restore := multicastTestConfig(t, "10.0.0.5", 5005) // unicast address, odd port
	defer restore()
	Config.Global.Software.Multicast.Codec = "opus"
	Config.Global.Software.Multicast.PacketMS = 17
	Config.Global.Software.Multicast.L16PayloadType = 8
	Config.Global.Software.Multicast.TTL = 999

	if warnings := checkMulticastConfigSanity(); warnings == 0 {
		t.Fatal("expected warnings for a unicast group, bad codec, odd port and bad payload type")
	}

	mc := Config.Global.Software.Multicast
	if mc.Enabled {
		t.Error("a non-multicast group should disable the section")
	}
	if mc.Codec != "pcmu" {
		t.Errorf("codec = %q, want pcmu", mc.Codec)
	}
	if mc.PacketMS != multicastDefaultPacketMS {
		t.Errorf("packetms = %v, want %v", mc.PacketMS, multicastDefaultPacketMS)
	}
	if mc.L16PayloadType != multicastDefaultL16PT {
		t.Errorf("l16payloadtype = %v, want %v", mc.L16PayloadType, multicastDefaultL16PT)
	}
	if mc.TTL != 1 {
		t.Errorf("ttl = %v, want 1", mc.TTL)
	}
}

func TestMulticastUserFilter(t *testing.T) {
	restore := multicastTestConfig(t, "239.0.1.10", 5004)
	defer restore()

	savedApp := appTalkkonnect
	t.Cleanup(func() { appTalkkonnect = savedApp })
	appTalkkonnect = &Talkkonnect{Client: &gumble.Client{
		Self: &gumble.User{Name: "self", Channel: &gumble.Channel{Name: "Operations"}},
	}}

	cases := []struct {
		name        string
		include     []string
		exclude     []string
		allChannels bool
		user        string
		channel     string
		want        bool
	}{
		{name: "empty include carries everyone", user: "anyone", channel: "Operations", want: true},
		{name: "include limits", include: []string{"alice"}, user: "bob", channel: "Operations", want: false},
		{name: "include admits", include: []string{"alice"}, user: "alice", channel: "Operations", want: true},
		{name: "include is case insensitive", include: []string{"Alice"}, user: "ALICE", channel: "Operations", want: true},
		{name: "exclude wins over include", include: []string{"alice"}, exclude: []string{"alice"}, user: "alice", channel: "Operations", want: false},
		{name: "other channel is dropped", user: "alice", channel: "Monitoring", want: false},
		{name: "allchannels widens", allChannels: true, user: "alice", channel: "Monitoring", want: true},
	}

	for _, tc := range cases {
		Config.Global.Software.Multicast.Include.User = tc.include
		Config.Global.Software.Multicast.Exclude.User = tc.exclude
		Config.Global.Software.Multicast.AllChannels = tc.allChannels
		multicastConfigureFromXML()

		if got := multicastAllowsUser(tc.user, tc.channel); got != tc.want {
			t.Errorf("%s: allows(%q in %q) = %v, want %v", tc.name, tc.user, tc.channel, got, tc.want)
		}
	}
}

// TestMulticastResampler checks the 48 kHz to 8 kHz decimator: the sample count,
// that state carries across calls, and that the anti-alias filter actually filters
// (a plain averaging decimator would fold 6 kHz back into the voice band).
func TestMulticastResampler(t *testing.T) {
	const inputSamples = 4800 // 100 ms at 48 kHz

	toneAt := func(hz float64) []int16 {
		out := make([]int16, inputSamples)
		for i := range out {
			out[i] = int16(10000 * math.Sin(2*math.Pi*hz*float64(i)/float64(gumble.AudioSampleRate)))
		}
		return out
	}

	// Fed in two halves, the output must be the same length as one whole pass:
	// no samples lost at the boundary.
	var res mcResampler
	tone := toneAt(1000)
	var out []int16
	out = res.resample(tone[:len(tone)/2], out)
	out = res.resample(tone[len(tone)/2:], out)
	if want := inputSamples / multicastDecimation; len(out) != want {
		t.Errorf("resampled %d samples, want %d", len(out), want)
	}

	peak := func(samples []int16) int {
		max := 0
		// Skip the filter's group delay at the start.
		for _, s := range samples[mcFIRTaps:] {
			v := int(s)
			if v < 0 {
				v = -v
			}
			if v > max {
				max = v
			}
		}
		return max
	}

	var passband, stopband mcResampler
	pass := peak(passband.resample(toneAt(1000), nil))
	stop := peak(stopband.resample(toneAt(6000), nil))

	if pass < 8000 {
		t.Errorf("1 kHz peak = %d, want the passband left near unity", pass)
	}
	if stop > pass/4 {
		t.Errorf("6 kHz peak = %d against 1 kHz %d: above 4 kHz must be attenuated, not aliased", stop, pass)
	}
}

// TestMulticastMixerEndToEnd runs the real sender against a loopback socket: no
// packets while idle, two sources summed once they are talking, and a marker bit
// starting each spurt.
func TestMulticastMixerEndToEnd(t *testing.T) {
	rx, port := multicastTestSocket(t)
	restore := multicastTestConfig(t, "127.0.0.1", port)
	defer restore()

	b := &Talkkonnect{Config: &gumble.Config{}}
	StartMulticastSender(b)
	defer StopMulticastSender()
	if !MulticastIsRunning() {
		t.Fatal("sender did not start")
	}

	// Idle: silence suppression means nothing on the wire.
	_ = rx.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
	buf := make([]byte, 2048)
	if n, _, err := rx.ReadFromUDP(buf); err == nil {
		t.Fatalf("read %d bytes while idle, want no packets", n)
	}

	// Two sources, each a constant level, so the mix is easy to verify.
	const level = 4000
	frame := make([]int16, 480) // 60 ms at 8 kHz, past the preroll
	for i := range frame {
		frame[i] = level
	}
	globalMulticast.feed8k("user:1", "alice", frame)
	globalMulticast.feed8k("user:2", "bob", frame)

	_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := rx.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read mixed packet: %v", err)
	}
	if want := 12 + 160; n != want {
		t.Fatalf("packet length = %d, want %d", n, want)
	}
	if buf[1]&0x80 == 0 {
		t.Error("first packet of the spurt is missing the marker bit")
	}

	pcm := mcDecodePayload(buf[1]&0x7F, buf[12:n])
	got := int(pcm[len(pcm)/2])
	if got < 2*level-500 || got > 2*level+500 {
		t.Errorf("mixed sample = %d, want about %d (both sources summed)", got, 2*level)
	}

	if status := multicastUISnapshot(); len(status.Sources) != 2 {
		t.Errorf("uistatus sources = %v, want alice and bob", status.Sources)
	}
}

// TestMulticastMixerClipsRatherThanWraps makes sure a hot mix saturates: wrapping
// would invert the waveform and sound far worse than clipping.
func TestMulticastMixerClipsRatherThanWraps(t *testing.T) {
	rx, port := multicastTestSocket(t)
	restore := multicastTestConfig(t, "127.0.0.1", port)
	defer restore()

	b := &Talkkonnect{Config: &gumble.Config{}}
	StartMulticastSender(b)
	defer StopMulticastSender()

	frame := make([]int16, 480)
	for i := range frame {
		frame[i] = 30000
	}
	globalMulticast.feed8k("user:1", "alice", frame)
	globalMulticast.feed8k("user:2", "bob", frame)

	buf := make([]byte, 2048)
	_ = rx.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := rx.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	pcm := mcDecodePayload(buf[1]&0x7F, buf[12:n])
	if got := pcm[len(pcm)/2]; got < 30000 {
		t.Errorf("clipped sample = %d, want it pinned near full scale, not wrapped", got)
	}
}

// TestMulticastRXTap drives the received-audio path the way gumble does: hand the
// listener a stream event and push 48 kHz packets at it, then check that resampled
// audio reaches the group. A user on the exclude list must produce nothing.
func TestMulticastRXTap(t *testing.T) {
	rx, port := multicastTestSocket(t)
	restore := multicastTestConfig(t, "127.0.0.1", port)
	defer restore()

	b := &Talkkonnect{Config: &gumble.Config{}}
	StartMulticastSender(b)
	defer StopMulticastSender()

	// 10 ms of 48 kHz tone, the frame size gumble delivers.
	frame := make(gumble.AudioBuffer, gumble.AudioDefaultFrameSize)
	for i := range frame {
		frame[i] = int16(9000 * math.Sin(2*math.Pi*440*float64(i)/float64(gumble.AudioSampleRate)))
	}

	feed := func(userName string, session uint32, packets int) {
		user := &gumble.User{Name: userName, Session: session, Channel: &gumble.Channel{Name: "Operations"}}
		stream := make(chan *gumble.AudioPacket, packets)
		globalMulticast.OnAudioStream(&gumble.AudioStreamEvent{User: user, C: stream})
		for i := 0; i < packets; i++ {
			stream <- &gumble.AudioPacket{Sender: user, AudioBuffer: frame}
		}
		close(stream)
	}

	feed("alice", 7, 30) // 300 ms, comfortably past the preroll

	buf := make([]byte, 2048)
	_ = rx.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, _, err := rx.ReadFromUDP(buf)
	if err != nil {
		t.Fatalf("no packet from the RX tap: %v", err)
	}
	if want := 12 + 160; n != want {
		t.Fatalf("packet length = %d, want %d", n, want)
	}
	pcm := mcDecodePayload(buf[1]&0x7F, buf[12:n])
	peak := 0
	for _, s := range pcm {
		v := int(s)
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	if peak < 3000 {
		t.Errorf("resampled peak = %d, want the tone to survive the 48k to 8k conversion", peak)
	}

	// An excluded talker must not reach the group. Drain first, then let the
	// hangover expire so the mixer is idle before the excluded user talks.
	Config.Global.Software.Multicast.Exclude.User = []string{"bob"}
	multicastConfigureFromXML()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		_ = rx.SetReadDeadline(time.Now().Add(400 * time.Millisecond))
		if _, _, err := rx.ReadFromUDP(buf); err != nil {
			break // the group has gone quiet
		}
	}

	feed("bob", 8, 30)
	_ = rx.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if n, _, err := rx.ReadFromUDP(buf); err == nil {
		t.Errorf("read %d bytes for an excluded talker, want silence", n)
	}
}

// TestMulticastPlayFileEndToEnd is the announcement path: ffmpeg decodes a
// generated tone and the mixer puts it on the group.
func TestMulticastPlayFileEndToEnd(t *testing.T) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not on PATH")
	}

	wav := filepath.Join(t.TempDir(), "tone.wav")
	gen := exec.Command("ffmpeg", "-nostdin", "-loglevel", "error", "-f", "lavfi",
		"-i", "sine=frequency=440:duration=1", "-ac", "1", "-ar", "8000", wav)
	if out, err := gen.CombinedOutput(); err != nil {
		t.Skipf("cannot generate test tone: %v %s", err, out)
	}

	rx, port := multicastTestSocket(t)
	restore := multicastTestConfig(t, "127.0.0.1", port)
	defer restore()

	b := &Talkkonnect{Config: &gumble.Config{}}
	StartMulticastSender(b)
	defer StopMulticastSender()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan error, 1)
	SafeGo(func() {
		done <- multicastPlayFile(ctx, "media:test", "test", wav, 100, 0, 0)
	})

	buf := make([]byte, 2048)
	for i := 0; i < 5; i++ {
		_ = rx.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _, err := rx.ReadFromUDP(buf)
		if err != nil {
			t.Fatalf("read announcement packet %d: %v", i, err)
		}
		if want := 12 + 160; n != want {
			t.Fatalf("announcement packet %d length = %d, want %d", i, n, want)
		}
		pcm := mcDecodePayload(buf[1]&0x7F, buf[12:n])
		energy := 0
		for _, s := range pcm {
			energy += int(s) * int(s) / len(pcm)
		}
		if i > 0 && energy == 0 {
			t.Errorf("announcement packet %d is silent", i)
		}
	}

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("multicastPlayFile: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Error("multicastPlayFile did not finish")
	}
}

// multicastTestSocket returns a bound loopback UDP socket and its port.
func multicastTestSocket(t *testing.T) (*net.UDPConn, int) {
	t.Helper()
	rx, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { rx.Close() })
	return rx, rx.LocalAddr().(*net.UDPAddr).Port
}

// multicastTestConfig points the sender at group:port and returns a func that puts
// the previous config and snapshot back.
func multicastTestConfig(t *testing.T, group string, port int) func() {
	t.Helper()

	savedConfig := Config
	savedSnapshot := multicastConfigSnapshot()

	Config.Global.Software.Multicast.Enabled = true
	Config.Global.Software.Multicast.Group = group
	Config.Global.Software.Multicast.Port = port
	Config.Global.Software.Multicast.Codec = "pcmu"
	Config.Global.Software.Multicast.TTL = 1
	Config.Global.Software.Multicast.Interface = ""
	Config.Global.Software.Multicast.PacketMS = 20
	Config.Global.Software.Multicast.L16PayloadType = 96
	Config.Global.Software.Multicast.Volume = 100
	Config.Global.Software.Multicast.AllChannels = true
	Config.Global.Software.Multicast.HangoverMS = 200
	Config.Global.Software.Multicast.Include.User = nil
	Config.Global.Software.Multicast.Exclude.User = nil
	multicastConfigureFromXML()

	// Names the tests use in log lines, so a failure says which case it was.
	t.Logf("multicast test destination %s:%s", group, strconv.Itoa(port))

	return func() {
		Config = savedConfig
		multicastCfgMu.Lock()
		multicastCfg = savedSnapshot
		multicastCfgMu.Unlock()
	}
}
