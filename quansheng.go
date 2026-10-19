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
 * talkkonnect is the based on talkiepi and barnard by Daniel Chote and Tim Cooper
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
 * quansheng.go -> talkkonnect interface to a Quansheng UV-K5 radio (dock firmware)
 * controlled through the qdock REST daemon (github.com/talkkonnect/quansheng)
 *
 * mumble -> radio : key the radio while mumble voice is being received
 * radio -> mumble : transmit into mumble while the radio's squelch is open
 */

package talkkonnect

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// quanshengRadioState is the part of qdock's /api/v1/state and lcd events we use.
type quanshengRadioState struct {
	Connected bool `json:"connected"`
	Version   *struct {
		Firmware string `json:"firmware"`
	} `json:"version"`
	TX          bool     `json:"tx"`
	RX          bool     `json:"rx"`
	VFOState    string   `json:"vfo_state"`
	Frequencies []string `json:"frequencies"`
	PTT         struct {
		On bool `json:"on"`
	} `json:"ptt"`
	Status struct {
		BatteryVolts   float64 `json:"battery_volts"`
		BatteryPercent int     `json:"battery_percent"`
	} `json:"status"`
}

type quanshengEvent struct {
	Type      string               `json:"type"`
	Connected *bool                `json:"connected"`
	LCD       *quanshengRadioState `json:"lcd"`
}

var (
	quanshengHTTP         = &http.Client{}
	quanshengMumbleNotify = make(chan struct{}, 1) // wakes the mumble -> radio worker
)

var quansheng struct {
	started       atomic.Bool
	mumbleTalking atomic.Bool // mumble voice is being received (set by the RX dispatcher)
	radioKeyed    atomic.Bool // we have the radio keyed for mumble voice
	radioRX       atomic.Bool // radio squelch open (from qdock events)
	txByRadio     atomic.Bool // talkkonnect is transmitting because the radio is receiving

	mu       sync.Mutex
	vfoState string
	freqs    string
}

func quanshengTimeout() time.Duration {
	return time.Duration(Config.Global.Hardware.Quansheng.TimeoutMsecs) * time.Millisecond
}

// quanshengCall performs one qdock API request. form is sent as a POST body when not nil.
func quanshengCall(ctx context.Context, timeout time.Duration, method, path string, form url.Values) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(Config.Global.Hardware.Quansheng.URL, "/")+path, body)
	if err != nil {
		return nil, err
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if Config.Global.Hardware.Quansheng.Token != "" {
		req.Header.Set("Authorization", "Bearer "+Config.Global.Hardware.Quansheng.Token)
	}
	resp, err := quanshengHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode/100 != 2 {
		var e struct{ Error string }
		if json.Unmarshal(b, &e) != nil || e.Error == "" {
			e.Error = strings.TrimSpace(string(b))
		}
		return nil, fmt.Errorf("qdock replied %v: %v", resp.StatusCode, e.Error)
	}
	return b, nil
}

func quanshengPTT(ctx context.Context, on bool) error {
	form := url.Values{"state": {"off"}}
	timeout := quanshengTimeout()
	if on {
		form = url.Values{"state": {"on"}, "timeout_s": {fmt.Sprint(Config.Global.Hardware.Quansheng.MumbleToRadio.PTTTimeoutSecs)}}
		// qdock delays the first key-up after it connects (the radio refuses TX for ~6s after a hello)
		timeout = max(timeout, 10*time.Second)
	}
	_, err := quanshengCall(ctx, timeout, "POST", "/api/v1/ptt", form)
	return err
}

func quanshengKeys(ctx context.Context, keys string) error {
	_, err := quanshengCall(ctx, max(quanshengTimeout(), 30*time.Second), "POST", "/api/v1/sequence", url.Values{"keys": {keys}, "gap_ms": {"150"}})
	return err
}

func quanshengGetState(ctx context.Context) (quanshengRadioState, error) {
	var st quanshengRadioState
	b, err := quanshengCall(ctx, quanshengTimeout(), "GET", "/api/v1/state", nil)
	if err == nil {
		err = json.Unmarshal(b, &st)
	}
	if err != nil {
		return st, err
	}
	// the VFO frequencies are only in the display mirror
	var lcd struct {
		Frequencies []string `json:"frequencies"`
	}
	if b, err := quanshengCall(ctx, quanshengTimeout(), "GET", "/api/v1/lcd", nil); err == nil && json.Unmarshal(b, &lcd) == nil {
		st.Frequencies = lcd.Frequencies
	}
	return st, nil
}

// QuanshengStart connects to qdock and starts the gateway workers. They stop when ctx is cancelled.
func (b *Talkkonnect) QuanshengStart(ctx context.Context) {
	cfg := Config.Global.Hardware.Quansheng
	quansheng.started.Store(true)

	st, err := quanshengGetState(ctx)
	switch {
	case err != nil:
		log.Printf("warn: Quansheng qdock at %v not reachable (%v), will keep retrying\n", cfg.URL, err)
	case !st.Connected:
		log.Printf("warn: Quansheng qdock at %v is running but the radio is not connected\n", cfg.URL)
	default:
		fw := ""
		if st.Version != nil {
			fw = st.Version.Firmware
		}
		log.Printf("info: Quansheng radio connected via qdock %v firmware %v VFO %v battery %.2fV\n", cfg.URL, fw, strings.Join(st.Frequencies, " / "), st.Status.BatteryVolts)
	}

	if keys := strings.TrimSpace(cfg.StartupKeys); keys != "" && err == nil && st.Connected {
		if err := quanshengKeys(ctx, keys); err != nil {
			log.Printf("error: Quansheng startup keys %q failed: %v\n", keys, err)
		} else {
			log.Printf("info: Quansheng startup keys %q sent\n", keys)
		}
	}

	SafeGo(func() { b.quanshengEventLoop(ctx) })
	if cfg.MumbleToRadio.Enabled {
		SafeGo(func() { quanshengMumbleToRadioWorker(ctx) })
	}
}

// QuanshengShutdown unkeys the radio if talkkonnect keyed it.
func QuanshengShutdown() {
	if !Config.Global.Hardware.Quansheng.Enabled || !quansheng.started.Load() {
		return
	}
	if quansheng.radioKeyed.Load() {
		if err := quanshengPTT(context.Background(), false); err != nil {
			log.Printf("error: Quansheng unkey on shutdown failed: %v\n", err)
		}
		quansheng.radioKeyed.Store(false)
	}
}

// QuanshengOnMumbleTalking and QuanshengOnMumbleSilence are called from the RX dispatcher on the
// start and end of received mumble voice. They never block.
func QuanshengOnMumbleTalking() { quanshengSetMumbleTalking(true) }
func QuanshengOnMumbleSilence() { quanshengSetMumbleTalking(false) }

func quanshengSetMumbleTalking(talking bool) {
	if !Config.Global.Hardware.Quansheng.Enabled || !Config.Global.Hardware.Quansheng.MumbleToRadio.Enabled || !quansheng.started.Load() {
		return
	}
	if quansheng.mumbleTalking.Swap(talking) != talking {
		select {
		case quanshengMumbleNotify <- struct{}{}:
		default:
		}
	}
}

// quanshengMumbleToRadioWorker keys the radio while mumble voice is received, refreshes the
// qdock PTT deadline while it lasts and unkeys after the hang time.
func quanshengMumbleToRadioWorker(ctx context.Context) {
	cfg := Config.Global.Hardware.Quansheng.MumbleToRadio
	hang := time.Duration(cfg.HangTimeMsecs) * time.Millisecond
	maxTX := time.Duration(cfg.MaxTXSecs) * time.Second
	refresh := time.NewTicker(time.Duration(cfg.PTTTimeoutSecs) * time.Second / 3)
	defer refresh.Stop()
	var (
		hangC     <-chan time.Time
		hangTimer *time.Timer
		keyedAt   time.Time
		capped    bool
	)
	unkey := func(why string) {
		if err := quanshengPTT(context.Background(), false); err != nil {
			log.Printf("error: Quansheng unkey failed: %v\n", err)
		}
		quansheng.radioKeyed.Store(false)
		log.Printf("info: Quansheng radio unkeyed (%v)\n", why)
	}
	for {
		select {
		case <-ctx.Done():
			if quansheng.radioKeyed.Load() {
				unkey("talkkonnect stopping")
			}
			return
		case <-quanshengMumbleNotify:
		case <-hangC:
			hangC = nil
			if !quansheng.mumbleTalking.Load() && quansheng.radioKeyed.Load() {
				unkey("mumble voice ended")
			}
			continue
		case <-refresh.C:
			if !quansheng.radioKeyed.Load() {
				continue
			}
			if time.Since(keyedAt) >= maxTX {
				capped = true
				unkey(fmt.Sprintf("reached maxtxsecs %v", cfg.MaxTXSecs))
				continue
			}
			if err := quanshengPTT(ctx, true); err != nil {
				log.Printf("error: Quansheng PTT refresh failed: %v\n", err)
			}
			continue
		}

		talking := quansheng.mumbleTalking.Load()
		switch {
		case talking && hangC != nil:
			hangTimer.Stop()
			hangC = nil
		case talking && !quansheng.radioKeyed.Load():
			if capped {
				continue // wait for this mumble transmission to end before keying again
			}
			if quansheng.radioRX.Load() || quansheng.txByRadio.Load() {
				log.Println("warn: Quansheng radio is receiving, not keying it for mumble voice")
				continue
			}
			if err := quanshengPTT(ctx, true); err != nil {
				log.Printf("error: Quansheng key radio failed: %v\n", err)
				continue
			}
			quansheng.radioKeyed.Store(true)
			keyedAt = time.Now()
			log.Println("info: Quansheng radio keyed for mumble voice")
		case !talking:
			capped = false
			if quansheng.radioKeyed.Load() && hangC == nil {
				hangTimer = time.NewTimer(hang)
				hangC = hangTimer.C
			}
		}
	}
}

// quanshengEventLoop follows qdock's server-sent events, keeping the radio state and driving the
// radio -> mumble direction. It reconnects until ctx is cancelled.
func (b *Talkkonnect) quanshengEventLoop(ctx context.Context) {
	rx := make(chan bool, 16)
	if Config.Global.Hardware.Quansheng.RadioToMumble.Enabled {
		SafeGo(func() { b.quanshengRadioToMumbleWorker(ctx, rx) })
	}
	streamClient := &http.Client{} // no timeout: the stream stays open
	failing := false
	for ctx.Err() == nil {
		err := quanshengFollowEvents(ctx, streamClient, rx)
		if ctx.Err() != nil {
			return
		}
		if !failing {
			log.Printf("warn: Quansheng qdock event stream lost (%v), reconnecting\n", err)
			failing = true
		}
		quanshengSetRadioRX(rx, false)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
		if _, err := quanshengGetState(ctx); err == nil && failing {
			log.Println("info: Quansheng qdock reachable again")
			failing = false
		}
	}
}

func quanshengFollowEvents(ctx context.Context, client *http.Client, rx chan<- bool) error {
	req, err := http.NewRequestWithContext(ctx, "GET", strings.TrimRight(Config.Global.Hardware.Quansheng.URL, "/")+"/api/v1/events", nil)
	if err != nil {
		return err
	}
	if Config.Global.Hardware.Quansheng.Token != "" {
		req.Header.Set("Authorization", "Bearer "+Config.Global.Hardware.Quansheng.Token)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("qdock replied %v", resp.Status)
	}
	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		data, ok := strings.CutPrefix(sc.Text(), "data: ")
		if !ok {
			continue
		}
		var ev quanshengEvent
		if json.Unmarshal([]byte(data), &ev) != nil {
			continue
		}
		switch {
		case ev.Connected != nil:
			if !*ev.Connected {
				log.Println("warn: Quansheng radio disconnected from qdock")
				quanshengSetRadioRX(rx, false)
			} else {
				log.Println("info: Quansheng radio connected to qdock")
			}
		case ev.LCD != nil:
			quanshengNoteDisplay(ev.LCD)
			quanshengSetRadioRX(rx, ev.LCD.RX)
		}
	}
	if err := sc.Err(); err != nil {
		return err
	}
	return errors.New("stream closed")
}

// quanshengNoteDisplay logs frequency changes and radio refusals such as "TX DISABLE".
func quanshengNoteDisplay(s *quanshengRadioState) {
	freqs := strings.Join(s.Frequencies, " / ")
	quansheng.mu.Lock()
	vfoChanged := s.VFOState != quansheng.vfoState
	freqChanged := freqs != "" && freqs != quansheng.freqs
	quansheng.vfoState = s.VFOState
	if freqs != "" {
		quansheng.freqs = freqs
	}
	quansheng.mu.Unlock()
	if vfoChanged && s.VFOState != "" {
		log.Printf("warn: Quansheng radio reports %v\n", s.VFOState)
	}
	if freqChanged && Config.Global.Software.PrintVariables.PrintQuansheng {
		log.Printf("info: Quansheng VFO %v\n", freqs)
	}
}

func quanshengSetRadioRX(rx chan<- bool, on bool) {
	if quansheng.radioRX.Swap(on) == on {
		return
	}
	select {
	case rx <- on:
	default:
	}
}

// quanshengRadioToMumbleWorker transmits into mumble while the radio receives a signal.
func (b *Talkkonnect) quanshengRadioToMumbleWorker(ctx context.Context, rx <-chan bool) {
	hang := time.Duration(Config.Global.Hardware.Quansheng.RadioToMumble.HangTimeMsecs) * time.Millisecond
	var (
		hangC     <-chan time.Time
		hangTimer *time.Timer
	)
	stop := func(why string) {
		if quansheng.txByRadio.Swap(false) && b.IsTransmitting {
			b.TransmitStop(false)
			log.Printf("info: Quansheng radio -> mumble transmit stopped (%v)\n", why)
		}
	}
	for {
		select {
		case <-ctx.Done():
			stop("talkkonnect stopping")
			return
		case <-hangC:
			hangC = nil
			if !quansheng.radioRX.Load() {
				stop("radio squelch closed")
			}
		case on := <-rx:
			switch {
			case on && hangC != nil:
				hangTimer.Stop()
				hangC = nil
			case on && !quansheng.txByRadio.Load():
				if quansheng.radioKeyed.Load() || b.IsTransmitting {
					continue // the radio is relaying mumble, or someone is already transmitting
				}
				b.TransmitStart()
				if b.IsTransmitting {
					quansheng.txByRadio.Store(true)
					log.Println("info: Quansheng radio -> mumble transmit started (radio squelch open)")
				}
			case !on && quansheng.txByRadio.Load() && hangC == nil:
				hangTimer = time.NewTimer(hang)
				hangC = hangTimer.C
			}
		}
	}
}

// cmdQuanshengStatus reports the radio state (HTTP API / MQTT command quanshengstatus).
func (b *Talkkonnect) cmdQuanshengStatus() {
	if !Config.Global.Hardware.Quansheng.Enabled {
		log.Println("warn: Quansheng radio control is disabled in talkkonnect.xml")
		sshRemoteReplyF("quansheng: disabled in talkkonnect.xml\n")
		return
	}
	st, err := quanshengGetState(context.Background())
	if err != nil {
		log.Printf("error: Quansheng status failed: %v\n", err)
		sshRemoteReplyF("quansheng: error: %v\n", err)
		return
	}
	fw := ""
	if st.Version != nil {
		fw = st.Version.Firmware
	}
	msg := fmt.Sprintf("quansheng: connected=%v firmware=%v vfo=%v tx=%v rx=%v ptt=%v battery=%.2fV(%v%%)",
		st.Connected, fw, strings.Join(st.Frequencies, "/"), st.TX, st.RX, st.PTT.On, st.Status.BatteryVolts, st.Status.BatteryPercent)
	if st.VFOState != "" {
		msg += " state=" + st.VFOState
	}
	log.Println("info: " + msg)
	sshRemoteReplyF("%v\n", msg)
}

// cmdQuanshengKeys presses a key sequence on the radio, e.g. "145500" or "menu up up menu".
func (b *Talkkonnect) cmdQuanshengKeys(keys string) {
	if !Config.Global.Hardware.Quansheng.Enabled {
		sshRemoteReplyF("quansheng: disabled in talkkonnect.xml\n")
		return
	}
	if err := quanshengKeys(context.Background(), keys); err != nil {
		log.Printf("error: Quansheng keys %q failed: %v\n", keys, err)
		sshRemoteReplyF("quansheng: keys %q failed: %v\n", keys, err)
		return
	}
	log.Printf("info: Quansheng keys %q sent\n", keys)
	sshRemoteReplyF("quansheng: keys %q sent\n", keys)
}

// cmdQuanshengPTTOn keys the radio for at most mumbletoradio ptttimeoutsecs.
func (b *Talkkonnect) cmdQuanshengPTTOn() {
	if !Config.Global.Hardware.Quansheng.Enabled {
		sshRemoteReplyF("quansheng: disabled in talkkonnect.xml\n")
		return
	}
	if err := quanshengPTT(context.Background(), true); err != nil {
		log.Printf("error: Quansheng PTT on failed: %v\n", err)
		sshRemoteReplyF("quansheng: ptt on failed: %v\n", err)
		return
	}
	log.Printf("info: Quansheng PTT on (auto off after %vs)\n", Config.Global.Hardware.Quansheng.MumbleToRadio.PTTTimeoutSecs)
	sshRemoteReplyF("quansheng: ptt on (auto off after %vs)\n", Config.Global.Hardware.Quansheng.MumbleToRadio.PTTTimeoutSecs)
}

// cmdQuanshengPTTOff unkeys the radio.
func (b *Talkkonnect) cmdQuanshengPTTOff() {
	if !Config.Global.Hardware.Quansheng.Enabled {
		sshRemoteReplyF("quansheng: disabled in talkkonnect.xml\n")
		return
	}
	if err := quanshengPTT(context.Background(), false); err != nil {
		log.Printf("error: Quansheng PTT off failed: %v\n", err)
		sshRemoteReplyF("quansheng: ptt off failed: %v\n", err)
		return
	}
	quansheng.radioKeyed.Store(false)
	log.Println("info: Quansheng PTT off")
	sshRemoteReplyF("quansheng: ptt off\n")
}
