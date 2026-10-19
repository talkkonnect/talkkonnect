package talkkonnect

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestQuanshengE2E drives a real radio through a running qdock. It only runs when
// QUANSHENG_E2E_URL is set (e.g. http://127.0.0.1:8099; QUANSHENG_E2E_TOKEN if needed).
// It presses UP then DOWN and keys the transmitter for about a second: use a dummy load.
func TestQuanshengE2E(t *testing.T) {
	base := os.Getenv("QUANSHENG_E2E_URL")
	if base == "" {
		t.Skip("QUANSHENG_E2E_URL not set")
	}
	saved := Config.Global.Hardware.Quansheng
	defer func() { Config.Global.Hardware.Quansheng = saved }()
	q := &Config.Global.Hardware.Quansheng
	q.Enabled, q.URL, q.Token, q.TimeoutMsecs = true, base, os.Getenv("QUANSHENG_E2E_TOKEN"), 3000
	q.MumbleToRadio.Enabled, q.MumbleToRadio.PTTTimeoutSecs, q.MumbleToRadio.HangTimeMsecs, q.MumbleToRadio.MaxTXSecs = true, 10, 200, 30
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	st, err := quanshengGetState(ctx)
	if err != nil || !st.Connected || len(st.Frequencies) == 0 {
		t.Fatalf("state %+v %v", st, err)
	}
	t.Logf("radio firmware %v VFO %v battery %.2fV", st.Version.Firmware, st.Frequencies, st.Status.BatteryVolts)
	(&Talkkonnect{}).cmdQuanshengStatus()

	// event stream: lcd events with the display
	rx := make(chan bool, 16)
	events := make(chan struct{}, 1)
	go func() {
		evctx, evcancel := context.WithTimeout(ctx, 3*time.Second)
		defer evcancel()
		quanshengFollowEvents(evctx, &http.Client{}, rx)
		events <- struct{}{}
	}()

	freq := func() string {
		s, err := quanshengGetState(ctx)
		if err != nil || len(s.Frequencies) == 0 {
			return ""
		}
		return s.Frequencies[0]
	}
	start := freq()
	(&Talkkonnect{}).cmdQuanshengKeys("up")
	time.Sleep(600 * time.Millisecond)
	up := freq()
	(&Talkkonnect{}).cmdQuanshengKeys("down")
	time.Sleep(600 * time.Millisecond)
	back := freq()
	t.Logf("VFO A %v -> up %v -> down %v", start, up, back)
	if up == start || back != start {
		t.Fatalf("keys did not move the VFO and back: %v %v %v", start, up, back)
	}
	<-events
	quansheng.mu.Lock()
	seen := quansheng.freqs
	quansheng.mu.Unlock()
	if !strings.Contains(seen, start) && !strings.Contains(seen, up) {
		t.Fatalf("event stream did not deliver the display, last freqs %q", seen)
	}

	// mumble -> radio: key the real radio for ~1s via the worker
	quansheng.started.Store(true)
	defer quansheng.started.Store(false)
	quansheng.radioRX.Store(false)
	done := make(chan struct{})
	wctx, wcancel := context.WithCancel(ctx)
	go func() { quanshengMumbleToRadioWorker(wctx); close(done) }()
	QuanshengOnMumbleTalking()
	deadline := time.Now().Add(10 * time.Second) // the first key-up can wait out the radio's post-hello TX block
	for !quansheng.radioKeyed.Load() && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(700 * time.Millisecond)
	s, _ := quanshengGetState(ctx)
	t.Logf("while mumble talking: qdock ptt=%v radio tx=%v vfo_state=%q", s.PTT.On, s.TX, s.VFOState)
	QuanshengOnMumbleSilence()
	// hang time, then the radio takes a few hundred ms to redraw without the TX marker
	var after quanshengRadioState
	for end := time.Now().Add(3 * time.Second); time.Now().Before(end); time.Sleep(100 * time.Millisecond) {
		if after, _ = quanshengGetState(ctx); !after.TX && !after.PTT.On {
			break
		}
	}
	wcancel()
	<-done
	if !s.PTT.On || !s.TX {
		t.Fatalf("radio did not transmit for mumble voice: %+v", s)
	}
	if after.PTT.On || after.TX || quansheng.radioKeyed.Load() {
		t.Fatalf("radio not unkeyed after mumble voice ended: %+v", after)
	}
	t.Logf("after mumble silence + hang: qdock ptt=%v radio tx=%v", after.PTT.On, after.TX)
}
