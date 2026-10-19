/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package talkkonnect

import (
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
)

// RX ALC (automatic level control): normalizes each remote talker's voice to a
// common loudness so loud and soft senders play back at about the same level.

const (
	rxALCDefaultTargetDBFS = -18.0
	rxALCDefaultMaxGainDB  = 12.0
	rxALCDefaultMinGainDB  = -12.0
	rxALCDefaultGateDBFS   = -50.0
	rxALCDefaultAttackMs   = 20.0
	rxALCDefaultReleaseMs  = 400.0
	rxALCFullScale         = 32768.0
	rxALCOverGap           = 400 * time.Millisecond // packet gap that ends a talker's over
)

type RXALCConfig struct {
	Enabled    bool    `xml:"enabled,attr"`
	TargetDBFS float64 `xml:"targetdbfs,attr"`
	MaxGainDB  float64 `xml:"maxgaindb,attr"`
	MinGainDB  float64 `xml:"mingaindb,attr"`
	GateDBFS   float64 `xml:"gatedbfs,attr"`
	AttackMs   float64 `xml:"attackms,attr"`
	ReleaseMs  float64 `xml:"releasems,attr"`
}

// rxALCParams is RXALCConfig resolved to linear values, with defaults filled in.
type rxALCParams struct {
	target, gate     float64 // linear RMS, int16 scale
	minGain, maxGain float64 // linear
	attackMs         float64
	releaseMs        float64
}

func dbToLinear(db float64) float64 { return math.Pow(10, db/20) }

func (c RXALCConfig) params() rxALCParams {
	orDefault := func(v, def float64) float64 {
		if v == 0 {
			return def
		}
		return v
	}
	p := rxALCParams{
		target:    rxALCFullScale * dbToLinear(orDefault(c.TargetDBFS, rxALCDefaultTargetDBFS)),
		gate:      rxALCFullScale * dbToLinear(orDefault(c.GateDBFS, rxALCDefaultGateDBFS)),
		maxGain:   dbToLinear(orDefault(c.MaxGainDB, rxALCDefaultMaxGainDB)),
		minGain:   dbToLinear(orDefault(c.MinGainDB, rxALCDefaultMinGainDB)),
		attackMs:  orDefault(c.AttackMs, rxALCDefaultAttackMs),
		releaseMs: orDefault(c.ReleaseMs, rxALCDefaultReleaseMs),
	}
	if p.minGain > p.maxGain {
		p.minGain, p.maxGain = p.maxGain, p.minGain
	}
	return p
}

// smoothingCoeff converts a time constant into a per-frame one-pole coefficient.
func smoothingCoeff(timeMs float64, frame time.Duration) float64 {
	frameMs := float64(frame) / float64(time.Millisecond)
	if timeMs <= 0 || frameMs <= 0 {
		return 1
	}
	return 1 - math.Exp(-frameMs/timeMs)
}

// rxALCGains remembers each talker's learned gain (by user name) so their next
// transmission starts at the right level instead of re-converging from unity.
var rxALCGains sync.Map

type rxALC struct {
	user string
	gain float64

	// per-over accumulators (speech frames only), reset by endOver
	overStart    time.Time
	speechFrames int
	inPower      float64
	outPower     float64
}

func newRxALC(user string) *rxALC {
	if g, ok := rxALCGains.Load(user); ok {
		return &rxALC{user: user, gain: g.(float64)}
	}
	return &rxALC{user: user, gain: 1.0}
}

// process writes gain-adjusted little-endian PCM for in into raw (len(raw) >= 2*len(in)).
// It never mutates in, since other gumble audio listeners may share the packet buffer.
func (a *rxALC) process(in []int16, raw []byte, p rxALCParams) {
	if len(in) == 0 {
		return
	}
	var sum float64
	for _, s := range in {
		f := float64(s)
		sum += f * f
	}
	rms := math.Sqrt(sum / float64(len(in)))

	start := a.gain
	// Below the gate (silence, background hiss) hold the gain so noise is not pumped up.
	if rms > p.gate {
		want := math.Max(p.minGain, math.Min(p.maxGain, p.target/rms))
		frame := time.Duration(len(in)) * time.Second / 48000
		coeff := smoothingCoeff(p.releaseMs, frame) // slow when boosting
		if want < a.gain {
			coeff = smoothingCoeff(p.attackMs, frame) // fast when cutting
		}
		a.gain += (want - a.gain) * coeff
		rxALCGains.Store(a.user, a.gain)
	}

	n := float64(len(in))
	var outSum float64
	for i, s := range in {
		g := start + (a.gain-start)*float64(i+1)/n // ramp across the frame to avoid zipper noise
		v := math.Max(-32768, math.Min(32767, math.Round(float64(s)*g)))
		outSum += v * v
		binary.LittleEndian.PutUint16(raw[i*2:], uint16(int16(v)))
	}

	if rms > p.gate {
		if a.speechFrames == 0 {
			a.overStart = time.Now()
		}
		a.speechFrames++
		a.inPower += sum / n
		a.outPower += outSum / n
		a.publish(true)
	}
}

// endOver logs a one-line summary of the talker's transmission and marks them idle.
// Called when packets stop arriving for a short while and when the stream closes.
func (a *rxALC) endOver() {
	if a.speechFrames == 0 {
		return
	}
	st := a.publish(false)
	log.Printf("info: RX ALC user=%q in=%.1f dBFS gain=%+.1f dB out=%.1f dBFS (%.1fs speech)",
		a.user, st.InDBFS, st.GainDB, st.OutDBFS, st.SpeechSecs)
	a.speechFrames, a.inPower, a.outPower = 0, 0, 0
}

// RXALCUserStatus is one talker's ALC state for /rxalc and /uistatus.
type RXALCUserStatus struct {
	User       string    `json:"user"`
	Talking    bool      `json:"talking"`
	InDBFS     float64   `json:"inDbfs"`  // average speech level received (current or last over)
	GainDB     float64   `json:"gainDb"`  // gain currently applied
	OutDBFS    float64   `json:"outDbfs"` // average level after ALC
	SpeechSecs float64   `json:"speechSecs"`
	Overs      int       `json:"overs"`
	LastHeard  time.Time `json:"lastHeard"`
}

var (
	rxALCStatusMu sync.Mutex
	rxALCStatus   = map[string]*RXALCUserStatus{}
)

func powerToDBFS(power float64) float64 {
	if power <= 0 {
		return -96
	}
	return math.Max(-96, 10*math.Log10(power/(rxALCFullScale*rxALCFullScale)))
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func (a *rxALC) publish(talking bool) RXALCUserStatus {
	frames := float64(a.speechFrames)
	rxALCStatusMu.Lock()
	defer rxALCStatusMu.Unlock()
	st, ok := rxALCStatus[a.user]
	if !ok {
		st = &RXALCUserStatus{User: a.user}
		rxALCStatus[a.user] = st
	}
	if st.Talking && !talking {
		st.Overs++
	}
	st.Talking = talking
	st.InDBFS = round1(powerToDBFS(a.inPower / frames))
	st.OutDBFS = round1(powerToDBFS(a.outPower / frames))
	st.GainDB = round1(20 * math.Log10(a.gain))
	st.SpeechSecs = round1(frames * 0.01)
	st.LastHeard = time.Now()
	return *st
}

// RXALCStatusSnapshot returns all talkers heard since startup, most recent first.
func RXALCStatusSnapshot() []RXALCUserStatus {
	rxALCStatusMu.Lock()
	out := make([]RXALCUserStatus, 0, len(rxALCStatus))
	for _, st := range rxALCStatus {
		out = append(out, *st)
	}
	rxALCStatusMu.Unlock()
	sort.Slice(out, func(i, j int) bool { return out[i].LastHeard.After(out[j].LastHeard) })
	return out
}

// resetRXALC forgets all learned gains and stats.
func resetRXALC() {
	rxALCGains.Clear()
	rxALCStatusMu.Lock()
	rxALCStatus = map[string]*RXALCUserStatus{}
	rxALCStatusMu.Unlock()
	log.Println("info: RX ALC learned gains and stats reset")
}

// httpRXALC serves /rxalc (live HTML view), /rxalc.json (JSON) and /rxalc?reset=1.
func (b *Talkkonnect) httpRXALC(w http.ResponseWriter, r *http.Request) {
	if !remoteControlHTTPClientIPAllowed(r) {
		http.Error(w, "403 forbidden: client address not allowed by remote control network ACL", http.StatusForbidden)
		return
	}
	if r.URL.Query().Get("reset") == "1" {
		resetRXALC()
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.URL.Path == "/rxalc.json" {
		w.Header().Set("Content-Type", "application/json")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(struct {
			Enabled bool              `json:"enabled"`
			Target  float64           `json:"targetDbfs"`
			Users   []RXALCUserStatus `json:"users"`
		}{
			Enabled: Config.Global.Software.Settings.RXALC.Enabled,
			Target:  round1(20 * math.Log10(Config.Global.Software.Settings.RXALC.params().target/rxALCFullScale)),
			Users:   RXALCStatusSnapshot(),
		})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(rxALCPage))
}

const rxALCPage = `<!doctype html>
<html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>RX ALC</title>
<style>
body{font-family:system-ui,sans-serif;margin:16px;background:#111;color:#ddd}
h1{font-size:18px;margin:0 0 4px}.sub{color:#888;font-size:13px;margin-bottom:12px}
table{border-collapse:collapse;width:100%;max-width:900px}
th,td{padding:6px 8px;border-bottom:1px solid #333;text-align:left;font-size:14px;white-space:nowrap}
th{color:#888;font-weight:500}.num{text-align:right;font-variant-numeric:tabular-nums}
.bar{position:relative;height:10px;width:160px;background:#222;border-radius:3px}
.bar i{position:absolute;top:0;bottom:0;left:0;border-radius:3px}
.bar b{position:absolute;top:-3px;bottom:-3px;width:2px;background:#fff}
.in i{background:#6a8cff}.out i{background:#3fbf7f}
.talk{color:#3fbf7f;font-weight:600}.boost{color:#3fbf7f}.cut{color:#e8a33d}
a{color:#6a8cff}
</style></head><body>
<h1>RX ALC per talker</h1>
<div class="sub" id="sub">loading…</div>
<table><thead><tr><th>User</th><th></th><th class="num">In dBFS</th><th>Input</th>
<th class="num">Gain dB</th><th class="num">Out dBFS</th><th>Output</th><th class="num">Overs</th><th>Last heard</th></tr></thead>
<tbody id="rows"></tbody></table>
<p class="sub"><a href="/rxalc.json">JSON</a> · <a href="#" id="reset">reset learned gains</a></p>
<script>
const pct=d=>Math.max(0,Math.min(100,(d+60)/60*100)); // -60..0 dBFS
function bar(cls,d,t){return '<div class="bar '+cls+'"><i style="width:'+pct(d)+'%"></i><b style="left:'+pct(t)+'%"></b></div>'}
function ago(s){const x=(Date.now()-new Date(s))/1000;return x<2?'now':x<120?Math.round(x)+'s ago':Math.round(x/60)+'m ago'}
function esc(s){return s.replace(/[&<>"]/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c]))}
async function tick(){
 try{
  const r=await fetch('/rxalc.json',{cache:'no-store'});const j=await r.json();
  document.getElementById('sub').textContent=(j.enabled?'enabled':'DISABLED (set <rxalc enabled="true"> in talkkonnect.xml)')+' · target '+j.targetDbfs+' dBFS · white tick = target';
  document.getElementById('rows').innerHTML=j.users.map(u=>'<tr><td>'+esc(u.user)+'</td><td class="talk">'+(u.talking?'● TALKING':'')+'</td>'
   +'<td class="num">'+u.inDbfs.toFixed(1)+'</td><td>'+bar('in',u.inDbfs,j.targetDbfs)+'</td>'
   +'<td class="num '+(u.gainDb>0.5?'boost':u.gainDb<-0.5?'cut':'')+'">'+(u.gainDb>0?'+':'')+u.gainDb.toFixed(1)+'</td>'
   +'<td class="num">'+u.outDbfs.toFixed(1)+'</td><td>'+bar('out',u.outDbfs,j.targetDbfs)+'</td>'
   +'<td class="num">'+u.overs+'</td><td>'+ago(u.lastHeard)+'</td></tr>').join('')||'<tr><td colspan="9" class="sub">No talkers heard yet</td></tr>';
 }catch(e){document.getElementById('sub').textContent='talkkonnect not reachable'}
}
document.getElementById('reset').onclick=e=>{e.preventDefault();fetch('/rxalc.json?reset=1').then(tick)};
tick();setInterval(tick,500);
</script></body></html>`

func logRXALCConfig() {
	c := Config.Global.Software.Settings.RXALC
	if !c.Enabled {
		log.Println("info: RX ALC                           ", false)
		return
	}
	p := c.params()
	log.Printf("info: RX ALC                            true (target %.1f dBFS, gain %.1f..%.1f dB, gate %.1f dBFS, attack %.0f ms, release %.0f ms)",
		20*math.Log10(p.target/rxALCFullScale), 20*math.Log10(p.minGain), 20*math.Log10(p.maxGain),
		20*math.Log10(p.gate/rxALCFullScale), p.attackMs, p.releaseMs)
}
