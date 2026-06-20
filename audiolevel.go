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
	"sync/atomic"
)

const pcmLevelFullScale = 20000 // int16 peak treated as 100% on the VU meter

var (
	rxAudioPeak atomic.Uint32
	txAudioPeak atomic.Uint32
)

func pcmLevelPercent(samples []int16) int {
	if len(samples) == 0 {
		return 0
	}
	var peak int32
	for _, sample := range samples {
		v := int32(sample)
		if v < 0 {
			v = -v
		}
		if v > peak {
			peak = v
		}
	}
	lvl := int(float64(peak) * 100.0 / pcmLevelFullScale)
	if lvl > 100 {
		lvl = 100
	}
	return lvl
}

func observeAudioPeak(meter *atomic.Uint32, samples []int16) {
	lvl := uint32(pcmLevelPercent(samples))
	for {
		old := meter.Load()
		if lvl <= old {
			return
		}
		if meter.CompareAndSwap(old, lvl) {
			return
		}
	}
}

// RecordRXAudioLevel tracks peak RX PCM level for /uistatus VU clients.
func RecordRXAudioLevel(samples []int16) {
	observeAudioPeak(&rxAudioPeak, samples)
}

// RecordTXAudioLevel tracks peak TX PCM level for /uistatus VU clients.
func RecordTXAudioLevel(samples []int16) {
	observeAudioPeak(&txAudioPeak, samples)
}

// AudioLevelSnapshot returns peak RX/TX levels since the last snapshot and resets meters.
func AudioLevelSnapshot() (rxLevel, txLevel int) {
	return int(rxAudioPeak.Swap(0)), int(txAudioPeak.Swap(0))
}
