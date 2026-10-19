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
	"log"
	"math"
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
	for i, s := range in {
		g := start + (a.gain-start)*float64(i+1)/n // ramp across the frame to avoid zipper noise
		v := math.Max(-32768, math.Min(32767, math.Round(float64(s)*g)))
		binary.LittleEndian.PutUint16(raw[i*2:], uint16(int16(v)))
	}
}

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
