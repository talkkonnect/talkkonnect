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

// multicast.go re-transmits audio as RTP to a multicast group, so IP PA speakers
// and SIP desk phones hear what this node hears. Two things feed it:
//
//   - received Mumble audio, tapped with a second gumble AudioListener (the same
//     technique avrecord.go uses) and filtered by the <include>/<exclude> user
//     lists and the channel scope, and
//   - <multimedia> announcements, when the profile sets <multicast>true</multicast>.
//
// Everything lands in one mixer: several people talking at once are summed into a
// single RTP stream with one SSRC, because a receiver handed two streams on the
// same group plays neither well. The mixer runs on its own 20 ms clock and stops
// sending when nobody is talking, marking the next packet as a new talk spurt.
//
// Mumble audio is 48 kHz; RTP goes out at 8 kHz (see multicastRTPClockRate), so
// each source is decimated by 6 behind an anti-alias low-pass.
//
// Nothing on the receiving path may block: gumble hands each listener its packets
// over an unbuffered channel, serially, so a slow multicast leg would stall all
// incoming audio for the whole session. Feeding the mixer only appends to a
// bounded buffer under a short mutex, and drops with a rate-limited log when the
// buffer is full.

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

const (
	multicastDefaultPacketMS   = 20  // 20 ms per packet is what hardware decoders expect
	multicastDefaultHangoverMS = 200 // how long a silent source holds the stream open
	multicastDefaultVolume     = 100
	multicastDefaultL16PT      = 96
	multicastPrerollFrames     = 3   // buffer 60 ms before a source joins the mix
	multicastQueueFrames       = 100 // ~2 s of 20 ms frames per source
	multicastDecimation        = gumble.AudioSampleRate / multicastRTPClockRate
)

// multicastConfig is the snapshot of <multicast> the sender runs on. Taking a
// copy at start keeps a config reload from changing the socket under the mixer.
type multicastConfig struct {
	Enabled        bool
	Dests          []multicastDest
	Interface      string
	PacketMS       int
	L16PayloadType int
	AllChannels    bool
	HangoverMS     int
	Include        map[string]bool // lowercased user names, empty means everyone
	Exclude        map[string]bool // lowercased user names, wins over Include
}

// SamplesPerFrame is the RTP payload size in samples, 160 at 20 ms / 8 kHz.
func (c multicastConfig) SamplesPerFrame() int {
	return multicastRTPClockRate * c.PacketMS / 1000
}

func (c multicastConfig) frameInterval() time.Duration {
	return time.Duration(c.PacketMS) * time.Millisecond
}

func (c multicastConfig) hangoverTicks() int {
	ticks := c.HangoverMS / c.PacketMS
	if ticks < 1 {
		ticks = 1
	}
	return ticks
}

var (
	multicastCfgMu sync.RWMutex
	multicastCfg   multicastConfig
)

// multicastConfigureFromXML derives the runtime snapshot from Config. Called from
// readxmlconfig on first load and on reload, in the same place
// internetRadioConfigureFromXML is called.
func multicastConfigureFromXML() {
	src := Config.Global.Software.Multicast

	cfg := multicastConfig{
		Enabled:        src.Enabled,
		Interface:      strings.TrimSpace(src.Interface),
		PacketMS:       src.PacketMS,
		L16PayloadType: src.L16PayloadType,
		AllChannels:    src.AllChannels,
		HangoverMS:     src.HangoverMS,
		Include:        multicastNameSet(src.Include.User),
		Exclude:        multicastNameSet(src.Exclude.User),
	}
	if cfg.PacketMS <= 0 {
		cfg.PacketMS = multicastDefaultPacketMS
	}
	if cfg.L16PayloadType <= 0 {
		cfg.L16PayloadType = multicastDefaultL16PT
	}
	if cfg.HangoverMS <= 0 {
		cfg.HangoverMS = multicastDefaultHangoverMS
	}

	volume := src.Volume
	if volume <= 0 {
		volume = multicastDefaultVolume
	}
	if group := strings.TrimSpace(src.Group); group != "" {
		cfg.Dests = []multicastDest{{
			Address: group,
			Port:    src.Port,
			Codec:   strings.ToLower(strings.TrimSpace(src.Codec)),
			TTL:     src.TTL,
			Gain:    volume,
		}}
	}

	multicastCfgMu.Lock()
	multicastCfg = cfg
	multicastCfgMu.Unlock()
}

// multicastApplyReload re-derives the snapshot after a config reload and starts,
// stops or restarts the sender to match. The include and exclude lists and the
// channel scope are read per packet, so those take effect without a restart; the
// socket and clock settings need one.
func multicastApplyReload() {
	previous := multicastConfigSnapshot()
	multicastConfigureFromXML()
	current := multicastConfigSnapshot()
	running := MulticastIsRunning()

	switch {
	case !current.Enabled && running:
		log.Println("info: Multicast Disabled By Config Reload")
		StopMulticastSender()
	case current.Enabled && !running:
		StartMulticastSender(appTalkkonnect)
	case current.Enabled && running && multicastWireChanged(previous, current):
		log.Println("info: Multicast Settings Changed By Config Reload, Restarting Sender")
		StopMulticastSender()
		StartMulticastSender(appTalkkonnect)
	}
}

// multicastWireChanged reports whether anything the open socket or the mixer clock
// depends on differs between two snapshots.
func multicastWireChanged(a, b multicastConfig) bool {
	if a.Interface != b.Interface || a.PacketMS != b.PacketMS ||
		a.L16PayloadType != b.L16PayloadType || a.HangoverMS != b.HangoverMS {
		return true
	}
	if len(a.Dests) != len(b.Dests) {
		return true
	}
	for i := range a.Dests {
		if a.Dests[i] != b.Dests[i] {
			return true
		}
	}
	return false
}

func multicastConfigSnapshot() multicastConfig {
	multicastCfgMu.RLock()
	defer multicastCfgMu.RUnlock()
	return multicastCfg
}

func multicastNameSet(names []string) map[string]bool {
	out := make(map[string]bool, len(names))
	for _, name := range names {
		if name = strings.ToLower(strings.TrimSpace(name)); name != "" {
			out[name] = true
		}
	}
	return out
}

// checkMulticastConfigSanity validates <multicast> and repairs what it can in
// place, disabling the section when it cannot. It returns the number of warnings
// raised, which CheckConfigSanity adds to its own count.
func checkMulticastConfigSanity() int {
	warnings := 0
	mc := &Config.Global.Software.Multicast

	group := strings.TrimSpace(mc.Group)
	ip := net.ParseIP(group)
	switch {
	case group == "":
		log.Print("warn: Config Error [Section MULTICAST] No group address configured, Disabling Multicast")
		mc.Enabled = false
		warnings++
	case ip == nil || ip.To4() == nil:
		log.Printf("warn: Config Error [Section MULTICAST] Group %q is not an IPv4 address, Disabling Multicast", group)
		mc.Enabled = false
		warnings++
	case !ip.IsMulticast():
		// Sending to a unicast address would work, but it is never what someone
		// writing a <multicast> section meant.
		log.Printf("warn: Config Error [Section MULTICAST] Group %v is not in the multicast range 224.0.0.0/4, Disabling Multicast", group)
		mc.Enabled = false
		warnings++
	}

	if mc.Port <= 0 || mc.Port > 65535 {
		log.Printf("warn: Config Error [Section MULTICAST] Port %v out of range, setting to 5004", mc.Port)
		mc.Port = 5004
		warnings++
	} else if mc.Port%2 != 0 {
		// RTP convention: the even port carries media, the odd one above it RTCP.
		// Some receivers will not accept an odd media port at all.
		log.Printf("warn: Config Error [Section MULTICAST] Port %v is odd; RTP media ports should be even (the odd port above is reserved for RTCP)", mc.Port)
		warnings++
	}

	codec := strings.ToLower(strings.TrimSpace(mc.Codec))
	switch codec {
	case "":
		mc.Codec = "pcmu"
	case "pcmu", "pcma", "alaw", "g711a", "l16", "pcm", "raw":
		mc.Codec = codec
	default:
		log.Printf("warn: Config Error [Section MULTICAST] Codec %q not supported (pcmu, pcma, l16), setting to pcmu", mc.Codec)
		mc.Codec = "pcmu"
		warnings++
	}
	if mcNormalizeCodec(mc.Codec) == "l16" {
		// Ported from gochimesd, which learned this the hard way: L16 rides a
		// dynamic payload type and G.711-only receivers drop it without a word.
		log.Printf("warn: Config Error [Section MULTICAST] Codec l16 is uncompressed PCM on a dynamic payload type; most hardware IP phones and PA devices decode only G.711 and will play SILENCE. Use pcmu unless every listener on %v supports L16", mc.Group)
		warnings++
	}

	if mc.TTL <= 0 {
		mc.TTL = 1
	} else if mc.TTL > 255 {
		log.Printf("warn: Config Error [Section MULTICAST] TTL %v out of range, setting to 1", mc.TTL)
		mc.TTL = 1
		warnings++
	}

	switch mc.PacketMS {
	case 0:
		mc.PacketMS = multicastDefaultPacketMS
	case 10, 20, 30, 40, 60:
	default:
		log.Printf("warn: Config Error [Section MULTICAST] packetms %v not supported (10, 20, 30, 40, 60), setting to %v", mc.PacketMS, multicastDefaultPacketMS)
		mc.PacketMS = multicastDefaultPacketMS
		warnings++
	}

	if mc.L16PayloadType == 0 {
		mc.L16PayloadType = multicastDefaultL16PT
	} else if mc.L16PayloadType < 96 || mc.L16PayloadType > 127 {
		log.Printf("warn: Config Error [Section MULTICAST] l16payloadtype %v is not a dynamic payload type (96-127), setting to %v", mc.L16PayloadType, multicastDefaultL16PT)
		mc.L16PayloadType = multicastDefaultL16PT
		warnings++
	}

	if mc.Volume == 0 {
		mc.Volume = multicastDefaultVolume
	} else if mc.Volume < 1 || mc.Volume > 200 {
		log.Printf("warn: Config Error [Section MULTICAST] volume %v%% out of range (1-200), setting to %v%%", mc.Volume, multicastDefaultVolume)
		mc.Volume = multicastDefaultVolume
		warnings++
	}

	if mc.HangoverMS == 0 {
		mc.HangoverMS = multicastDefaultHangoverMS
	} else if mc.HangoverMS < 0 || mc.HangoverMS > 5000 {
		log.Printf("warn: Config Error [Section MULTICAST] hangoverms %v out of range (0-5000), setting to %v", mc.HangoverMS, multicastDefaultHangoverMS)
		mc.HangoverMS = multicastDefaultHangoverMS
		warnings++
	}

	if iface := strings.TrimSpace(mc.Interface); iface != "" {
		if _, err := net.InterfaceByName(iface); err != nil {
			log.Printf("warn: Config Error [Section MULTICAST] Interface %q not found on this host; multicast will follow the default route", iface)
			warnings++
		}
	}

	// A name in both lists is a contradiction, and exclude wins, so say so rather
	// than leaving someone wondering why their include entry is silent.
	for _, name := range mc.Include.User {
		for _, other := range mc.Exclude.User {
			if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(other)) {
				log.Printf("warn: Config Error [Section MULTICAST] User %q is in both include and exclude; exclude wins", strings.TrimSpace(name))
				warnings++
			}
		}
	}

	if !multicastFFmpegAvailable() {
		for _, profile := range Config.Global.Multimedia.ID {
			if profile.Enabled && profile.Params.Multicast {
				log.Printf("warn: Config Error [Section MULTICAST] ffmpeg not found on PATH; multimedia profile %q cannot be multicast", profile.Value)
				warnings++
			}
		}
	}

	return warnings
}

// multicastFFmpegAvailable reports whether announcements can be decoded at all.
func multicastFFmpegAvailable() bool {
	_, err := exec.LookPath(ffmpegBinaryPath())
	return err == nil
}

// --- source buffers ---------------------------------------------------------

// multicastSource is one contributor to the mix: a talking user or a playing
// announcement. pending holds 8 kHz samples waiting to be packetized.
type multicastSource struct {
	name      string
	pending   []int16
	res       *mcResampler // nil when the source is already at 8 kHz
	active    bool         // in the mix, i.e. past the preroll
	idleTicks int
	finished  bool // producer is gone, drop the source once it drains
	dropped   uint64
	lastDrop  time.Time
}

// --- the sender -------------------------------------------------------------

type multicastSender struct {
	mu       sync.Mutex
	running  bool
	cfg      multicastConfig
	streams  []*rtpStream
	sources  map[string]*multicastSource
	detacher gumble.Detacher
	cancel   context.CancelFunc
	done     chan struct{}
	spurt    bool // true while packets are flowing
	packets  uint64
	started  time.Time
}

var globalMulticast multicastSender

// StartMulticastSender opens the multicast socket and attaches the RX audio tap.
// It is idempotent, and a no-op when <multicast enabled="false">.
func StartMulticastSender(b *Talkkonnect) {
	if b == nil || b.Config == nil {
		log.Println("error: cannot start multicast without Mumble config")
		return
	}

	cfg := multicastConfigSnapshot()
	if !cfg.Enabled {
		log.Println("info: Multicast Disabled in Config, Not Starting")
		return
	}
	if len(cfg.Dests) == 0 {
		log.Println("error: multicast is enabled but no group is configured; not starting")
		return
	}

	globalMulticast.mu.Lock()
	if globalMulticast.running {
		globalMulticast.mu.Unlock()
		log.Println("info: multicast sender is already running")
		return
	}

	// One bad destination must not take the others down with it.
	var streams []*rtpStream
	for _, dest := range cfg.Dests {
		stream, err := newRTPStream(dest, cfg.Interface, cfg.L16PayloadType)
		if err != nil {
			log.Printf("error: multicast destination %s not opened: %v", dest.Addr(), err)
			continue
		}
		streams = append(streams, stream)
	}
	if len(streams) == 0 {
		globalMulticast.mu.Unlock()
		log.Println("error: multicast has no usable destination; not starting")
		return
	}

	parent := b.MasterCtx
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancel(parent)

	globalMulticast.cfg = cfg
	globalMulticast.streams = streams
	globalMulticast.sources = make(map[string]*multicastSource)
	globalMulticast.cancel = cancel
	globalMulticast.done = make(chan struct{})
	globalMulticast.spurt = false
	globalMulticast.packets = 0
	globalMulticast.started = time.Now()
	globalMulticast.running = true
	done := globalMulticast.done
	globalMulticast.mu.Unlock()

	SafeGo(func() { globalMulticast.pace(ctx, done) })

	// Attached to b.Config rather than to the session, so the tap and the socket
	// survive a Mumble reconnect the way the traffic recorder does.
	globalMulticast.mu.Lock()
	globalMulticast.detacher = b.Config.AttachAudio(&globalMulticast)
	globalMulticast.mu.Unlock()

	dest := cfg.Dests[0]
	log.Printf("info: Multicast Started to %s codec %s ttl %v packet %vms interface %q volume %v%%",
		dest.Addr(), mcNormalizeCodec(dest.Codec), dest.TTL, cfg.PacketMS, cfg.Interface, dest.Gain)
	if len(cfg.Include) > 0 {
		log.Printf("info: Multicast Includes Only Users %v", multicastSortedNames(cfg.Include))
	}
	if len(cfg.Exclude) > 0 {
		log.Printf("info: Multicast Excludes Users %v", multicastSortedNames(cfg.Exclude))
	}
	if !cfg.AllChannels {
		log.Println("info: Multicast Carries The Joined Channel Only, Set <allchannels>true</allchannels> To Widen")
	}
}

// StopMulticastSender detaches the tap, stops the mixer and closes the sockets.
func StopMulticastSender() {
	globalMulticast.mu.Lock()
	if !globalMulticast.running {
		globalMulticast.mu.Unlock()
		return
	}
	detacher := globalMulticast.detacher
	cancel := globalMulticast.cancel
	done := globalMulticast.done
	globalMulticast.detacher = nil
	globalMulticast.cancel = nil
	globalMulticast.running = false
	globalMulticast.mu.Unlock()

	if detacher != nil {
		detacher.Detach()
	}
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}

	globalMulticast.mu.Lock()
	for _, stream := range globalMulticast.streams {
		stream.close()
	}
	globalMulticast.streams = nil
	globalMulticast.sources = nil
	globalMulticast.mu.Unlock()

	log.Println("info: Multicast Stopped")
}

// MulticastIsRunning reports whether packets can currently flow.
func MulticastIsRunning() bool {
	globalMulticast.mu.Lock()
	defer globalMulticast.mu.Unlock()
	return globalMulticast.running
}

// pace is the mixer clock: one packet per tick per destination, for as long as
// any source has audio.
func (m *multicastSender) pace(ctx context.Context, done chan struct{}) {
	defer close(done)

	m.mu.Lock()
	cfg := m.cfg
	m.mu.Unlock()

	spf := cfg.SamplesPerFrame()
	ticker := time.NewTicker(cfg.frameInterval())
	defer ticker.Stop()

	acc := make([]int32, spf)
	frame := make([]int16, spf)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !m.mixTick(acc, frame, spf, cfg.hangoverTicks()) {
				continue
			}
			m.writeFrame(frame)
		}
	}
}

// mixTick sums one packet worth of audio from every active source into frame and
// reports whether anything was mixed. Sources past their hangover are dropped.
func (m *multicastSender) mixTick(acc []int32, frame []int16, spf, hangoverTicks int) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range acc {
		acc[i] = 0
	}
	mixed := false

	for key, src := range m.sources {
		if !src.active {
			// Wait for a little audio to pile up before joining the mix, so
			// network jitter on the Mumble side does not chop the RTP stream
			// into a series of one-packet spurts.
			if len(src.pending) >= multicastPrerollFrames*spf {
				src.active = true
				src.idleTicks = 0
			} else if src.finished && len(src.pending) == 0 {
				delete(m.sources, key)
				continue
			} else {
				continue
			}
		}

		n := len(src.pending)
		if n > spf {
			n = spf
		}
		if n == 0 {
			src.idleTicks++
			if src.idleTicks >= hangoverTicks {
				src.active = false
				if src.finished {
					delete(m.sources, key)
				}
			}
			continue
		}

		src.idleTicks = 0
		for i := 0; i < n; i++ {
			acc[i] += int32(src.pending[i])
		}
		// Compact rather than reslice, so the backing array cannot creep.
		src.pending = append(src.pending[:0], src.pending[n:]...)
		if src.finished && len(src.pending) == 0 {
			delete(m.sources, key)
		}
		mixed = true
	}

	if !mixed {
		// Nobody is talking: stop sending and mark the next packet as the start
		// of a new spurt so a hardware decoder resets its jitter buffer.
		if m.spurt {
			m.spurt = false
			for _, stream := range m.streams {
				stream.startSpurt()
			}
		}
		return false
	}

	for i := range frame {
		frame[i] = mcClampToInt16(acc[i])
	}
	m.spurt = true
	return true
}

// writeFrame fans one mixed frame out to every destination. Socket writes happen
// outside the mixer lock so a stalled NIC cannot block the audio tap.
func (m *multicastSender) writeFrame(frame []int16) {
	m.mu.Lock()
	streams := make([]*rtpStream, len(m.streams))
	copy(streams, m.streams)
	m.mu.Unlock()

	for _, stream := range streams {
		if err := stream.writeFrame(frame); err != nil {
			log.Printf("warn: multicast send to %v failed: %v", stream.dst, err)
		}
	}
	atomic.AddUint64(&m.packets, 1)
}

// --- feeding the mixer ------------------------------------------------------

// feed48k hands Mumble-rate audio to the mixer, resampling to 8 kHz on the way.
func (m *multicastSender) feed48k(key, name string, pcm []int16) {
	m.feed(key, name, pcm, true)
}

// feed8k hands audio that is already at the RTP clock rate to the mixer.
func (m *multicastSender) feed8k(key, name string, pcm []int16) {
	m.feed(key, name, pcm, false)
}

func (m *multicastSender) feed(key, name string, pcm []int16, needsResample bool) {
	if len(pcm) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.running || m.sources == nil {
		return
	}

	src := m.sources[key]
	if src == nil {
		src = &multicastSource{name: name}
		if needsResample {
			src.res = &mcResampler{}
		}
		m.sources[key] = src
		log.Printf("debug: multicast source %q added", name)
	}
	src.finished = false

	if src.res != nil {
		src.pending = src.res.resample(pcm, src.pending)
	} else {
		src.pending = append(src.pending, pcm...)
	}

	// Bounded buffer: if the mixer cannot keep up, throw away the oldest audio
	// rather than growing without limit or blocking the producer.
	max := multicastQueueFrames * m.cfg.SamplesPerFrame()
	if over := len(src.pending) - max; over > 0 {
		src.pending = append(src.pending[:0], src.pending[over:]...)
		src.dropped += uint64(over)
		if time.Since(src.lastDrop) > 5*time.Second {
			log.Printf("warn: multicast buffer full for %q, dropped %v sample(s)", src.name, src.dropped)
			src.dropped = 0
			src.lastDrop = time.Now()
		}
	}
}

// finishSource marks a producer as gone. The mixer keeps the source until its
// buffer drains, then forgets it.
func (m *multicastSender) finishSource(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if src := m.sources[key]; src != nil {
		src.finished = true
	}
}

// --- the received-audio tap -------------------------------------------------

// OnAudioStream is the gumble AudioListener entry point, one call per talking
// user. It mirrors avrecord.go's recorder: a SafeGo goroutine draining e.C, with
// no blocking work in the loop.
func (m *multicastSender) OnAudioStream(e *gumble.AudioStreamEvent) {
	if e == nil || e.User == nil {
		return
	}
	key := "user:" + strconv.FormatUint(uint64(e.User.Session), 10)

	SafeGo(func() {
		defer m.finishSource(key)
		for packet := range e.C {
			if packet == nil || len(packet.AudioBuffer) == 0 {
				continue
			}

			user := packet.Sender
			if user == nil {
				user = e.User
			}
			channelName := ""
			if user.Channel != nil {
				channelName = user.Channel.Name
			}
			if !multicastAllowsUser(user.Name, channelName) {
				continue
			}

			// Every listener is handed the same packet pointer, so the buffer is
			// read-only here; feed copies what it keeps.
			m.feed48k(key, user.Name, packet.AudioBuffer)
		}
	})
}

// multicastAllowsUser applies the exclude list, then the include list, then the
// channel scope. Exclude wins; an empty include list means everyone.
func multicastAllowsUser(userName, channelName string) bool {
	cfg := multicastConfigSnapshot()
	name := strings.ToLower(strings.TrimSpace(userName))

	if cfg.Exclude[name] {
		return false
	}
	if len(cfg.Include) > 0 && !cfg.Include[name] {
		return false
	}

	// A user talkkonnect is ignoring locally should not reach the PA either.
	if Config.Global.Software.IgnoreUser.IgnoreUserEnabled && len(Config.Global.Software.IgnoreUser.IgnoreUserRegex) > 0 {
		if checkRegex(Config.Global.Software.IgnoreUser.IgnoreUserRegex, userName) {
			return false
		}
	}

	if cfg.AllChannels {
		return true
	}
	return multicastIsOwnChannel(channelName)
}

// multicastIsOwnChannel reports whether channelName is the channel talkkonnect is
// joined to. Audio heard through <listentochannels> or an incoming whisper comes
// from elsewhere and stays off the group unless <allchannels> is set.
func multicastIsOwnChannel(channelName string) bool {
	b := appTalkkonnect
	if b == nil || b.Client == nil || b.Client.Self == nil || b.Client.Self.Channel == nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(channelName), b.Client.Self.Channel.Name)
}

func multicastSortedNames(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	return out
}

// --- 48 kHz -> 8 kHz resampling ---------------------------------------------

// mcFIRTaps is the length of the anti-alias low-pass in front of the decimator.
// 25 taps at 48 kHz is a few hundred multiplies per output sample band, cheap
// next to Opus decoding, and it keeps everything above 4 kHz from folding back
// into the voice band the way a plain sample-averaging decimator would.
const mcFIRTaps = 25

// mcDecimCutoffHz sits below the 4 kHz Nyquist limit of the 8 kHz output while
// leaving the intelligibility range of speech intact.
const mcDecimCutoffHz = 3400.0

var mcDecimFIR = mcBuildLowPass(mcFIRTaps, mcDecimCutoffHz, float64(gumble.AudioSampleRate))

// mcBuildLowPass returns a Hamming-windowed sinc low-pass normalised to unity DC
// gain. Computed at init rather than pasted as a coefficient table so the cutoff
// stays readable next to the sample rates it depends on.
func mcBuildLowPass(taps int, cutoffHz, sampleRate float64) []float32 {
	out := make([]float32, taps)
	fc := cutoffHz / sampleRate // normalised cutoff, cycles per sample
	mid := float64(taps-1) / 2
	var sum float64
	raw := make([]float64, taps)
	for i := 0; i < taps; i++ {
		x := float64(i) - mid
		var sinc float64
		if x == 0 {
			sinc = 2 * fc
		} else {
			sinc = math.Sin(2*math.Pi*fc*x) / (math.Pi * x)
		}
		window := 0.54 - 0.46*math.Cos(2*math.Pi*float64(i)/float64(taps-1))
		raw[i] = sinc * window
		sum += raw[i]
	}
	for i := range raw {
		out[i] = float32(raw[i] / sum)
	}
	return out
}

// mcResampler decimates 48 kHz mono PCM to 8 kHz. State carries across calls, so
// no sample is lost at packet boundaries.
type mcResampler struct {
	hist  [mcFIRTaps]float32
	pos   int // next write position in the circular history
	phase int // input samples since the last output
}

// resample appends the decimated output to out and returns it. The filter is
// symmetric, so the history can be convolved oldest-first without reversing the
// coefficients.
func (r *mcResampler) resample(in []int16, out []int16) []int16 {
	for _, sample := range in {
		r.hist[r.pos] = float32(sample)
		r.pos++
		if r.pos == mcFIRTaps {
			r.pos = 0
		}
		r.phase++
		if r.phase < multicastDecimation {
			continue
		}
		r.phase = 0

		var accumulator float32
		idx := r.pos // oldest sample in the history
		for i := 0; i < mcFIRTaps; i++ {
			accumulator += mcDecimFIR[i] * r.hist[idx]
			idx++
			if idx == mcFIRTaps {
				idx = 0
			}
		}
		out = append(out, mcClampToInt16(int32(accumulator)))
	}
	return out
}

// --- announcement input -----------------------------------------------------

// multicastPlayFile decodes one media file straight to the RTP sample rate and
// feeds it to the mixer as a named source, blocking until the file is done or ctx
// is cancelled. volume is a percentage, offset and duration are seconds and are
// ignored when zero.
func multicastPlayFile(ctx context.Context, sourceKey, name, path string, volume int, offset, duration float32) error {
	if !MulticastIsRunning() {
		return fmt.Errorf("multicast is not running")
	}

	cfg := multicastConfigSnapshot()
	spf := cfg.SamplesPerFrame()

	frames, wait, err := decodePCMStreamViaFFmpeg(ctx, path, multicastRTPClockRate, spf, offset, duration)
	if err != nil {
		return err
	}
	defer globalMulticast.finishSource(sourceKey)

	for frame := range frames {
		if volume > 0 && volume != 100 {
			frame = mcApplyGain(frame, volume)
		}
		globalMulticast.feed8k(sourceKey, name, frame)

		// The mixer drains one frame per packet interval; keep roughly that pace
		// so a long file does not sit in the buffer being dropped at the far end.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if !multicastWaitForRoom(ctx, sourceKey, spf) {
			return ctx.Err()
		}
	}
	return wait()
}

// multicastWaitForRoom paces the producer against the mixer: it returns once the
// source has less than the preroll buffered, or false if ctx ended.
func multicastWaitForRoom(ctx context.Context, sourceKey string, spf int) bool {
	const highWater = multicastPrerollFrames * 2
	for {
		globalMulticast.mu.Lock()
		src := globalMulticast.sources[sourceKey]
		pending := 0
		if src != nil {
			pending = len(src.pending)
		}
		running := globalMulticast.running
		globalMulticast.mu.Unlock()

		if !running || pending < highWater*spf {
			return running
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(time.Duration(spf) * time.Second / multicastRTPClockRate):
		}
	}
}

// ffmpegBinaryPath finds ffmpeg the way the rest of talkkonnect does.
func ffmpegBinaryPath() string {
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	if _, err := exec.LookPath("/usr/bin/ffmpeg"); err == nil {
		return "/usr/bin/ffmpeg"
	}
	return "ffmpeg"
}

// decodePCMStreamViaFFmpeg decodes path to signed 16-bit little-endian mono PCM
// at rate and emits it as spf-sized frames. Unlike loadPCMSoundViaFFmpeg it
// streams, so an announcement starts playing immediately and a long file never
// sits in memory. The returned func waits for ffmpeg and reports its error.
func decodePCMStreamViaFFmpeg(ctx context.Context, path string, rate, spf int, offset, duration float32) (<-chan []int16, func() error, error) {
	args := []string{"-nostdin", "-hide_banner", "-loglevel", "error"}
	if offset > 0 {
		args = append(args, "-ss", strconv.FormatFloat(float64(offset), 'f', 3, 32))
	}
	args = append(args, "-i", path)
	if duration > 0 {
		args = append(args, "-t", strconv.FormatFloat(float64(duration), 'f', 3, 32))
	}
	args = append(args, "-ac", "1", "-ar", strconv.Itoa(rate), "-f", "s16le", "-")

	cmd := exec.CommandContext(ctx, ffmpegBinaryPath(), args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	frames := make(chan []int16, multicastQueueFrames)
	waitErr := make(chan error, 1)

	SafeGo(func() {
		defer close(frames)
		buf := make([]byte, spf*2)
		for {
			n, rerr := io.ReadFull(stdout, buf)
			if n > 0 {
				samples := n / 2
				frame := make([]int16, samples)
				for i := 0; i < samples; i++ {
					frame[i] = int16(binary.LittleEndian.Uint16(buf[i*2:]))
				}
				select {
				case frames <- frame:
				case <-ctx.Done():
					waitErr <- cmd.Wait()
					return
				}
			}
			if rerr != nil {
				werr := cmd.Wait()
				// EOF and a short final read are the normal end of a file.
				if rerr != io.EOF && rerr != io.ErrUnexpectedEOF && ctx.Err() == nil {
					werr = rerr
				}
				waitErr <- werr
				return
			}
		}
	})

	return frames, func() error {
		select {
		case err := <-waitErr:
			if err != nil && ctx.Err() == nil {
				return err
			}
			return nil
		case <-time.After(5 * time.Second):
			return nil
		}
	}, nil
}

// --- status -----------------------------------------------------------------

// UIMulticast is the multicast state /uistatus reports and the mc CLI prints.
type UIMulticast struct {
	Enabled  bool     `json:"enabled"`
	Running  bool     `json:"running"`
	Group    string   `json:"group,omitempty"`
	Codec    string   `json:"codec,omitempty"`
	TTL      int      `json:"ttl,omitempty"`
	PacketMS int      `json:"packetMs,omitempty"`
	Volume   int      `json:"volume,omitempty"`
	Sources  []string `json:"sources,omitempty"`
	Packets  uint64   `json:"packets"`
}

// multicastUISnapshot reads live sender state. The sender is the single source of
// truth here, so there is no separate Record… mirror to keep in step.
func multicastUISnapshot() UIMulticast {
	cfg := multicastConfigSnapshot()
	out := UIMulticast{
		Enabled:  cfg.Enabled,
		PacketMS: cfg.PacketMS,
	}
	if len(cfg.Dests) > 0 {
		dest := cfg.Dests[0]
		out.Group = dest.Addr()
		out.Codec = mcNormalizeCodec(dest.Codec)
		out.TTL = dest.TTL
		out.Volume = dest.Gain
	}

	globalMulticast.mu.Lock()
	out.Running = globalMulticast.running
	for _, src := range globalMulticast.sources {
		if src.active {
			out.Sources = append(out.Sources, src.name)
		}
	}
	globalMulticast.mu.Unlock()
	out.Packets = atomic.LoadUint64(&globalMulticast.packets)
	return out
}
