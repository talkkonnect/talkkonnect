package talkkonnect

import (
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// Every shipped config that has a <quansheng> section must parse into the struct.
func TestShippedConfigsQuanshengSection(t *testing.T) {
	files := []string{"talkkonnect-v4.xml", "sample-configs/talkkonnect-v4.xml", "sample-configs/talkkonnect-current-sample-config.xml",
		"sample-configs/talkkonnect-v4-rpi4.xml", "sample-configs/talkkonnect-v4-pc-x86.xml"}
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		var c ConfigStruct
		if err := xml.Unmarshal(b, &c); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		q := c.Global.Hardware.Quansheng
		if q.URL != "http://127.0.0.1:8099" || q.TimeoutMsecs != 3000 || q.MumbleToRadio.PTTTimeoutSecs != 30 ||
			q.MumbleToRadio.HangTimeMsecs != 300 || q.MumbleToRadio.MaxTXSecs != 180 || q.RadioToMumble.HangTimeMsecs != 500 {
			t.Errorf("%s: quansheng section not parsed: %+v", f, q)
		}
		if q.Enabled || q.MumbleToRadio.Enabled || q.RadioToMumble.Enabled {
			t.Errorf("%s: shipped quansheng section must be disabled", f)
		}
		if !c.Global.Software.PrintVariables.PrintQuansheng {
			t.Errorf("%s: printquansheng missing", f)
		}
	}
}

// fakeQdock records PTT requests like a qdock server.
type fakeQdock struct {
	mu   sync.Mutex
	ptts []string
	keys []string
}

func (f *fakeQdock) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer tok" {
		http.Error(w, `{"error":"bad token"}`, http.StatusUnauthorized)
		return
	}
	switch r.URL.Path {
	case "/api/v1/ptt":
		r.ParseForm()
		f.mu.Lock()
		f.ptts = append(f.ptts, r.PostForm.Get("state"))
		f.mu.Unlock()
		w.Write([]byte(`{"on":true}`))
	case "/api/v1/sequence":
		r.ParseForm()
		f.mu.Lock()
		f.keys = append(f.keys, r.PostForm.Get("keys"))
		f.mu.Unlock()
		w.Write([]byte(`{"ok":true}`))
	case "/api/v1/state":
		w.Write([]byte(`{"connected":true,"frequencies":["145.500"]}`))
	default:
		http.NotFound(w, r)
	}
}

func (f *fakeQdock) log() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.ptts...)
}

// The quanshengkeys HTTP API command takes the sequence from the keys query parameter.
func TestQuanshengHTTPAPIKeys(t *testing.T) {
	fq := &fakeQdock{}
	ts := httptest.NewServer(fq)
	defer ts.Close()

	b, err := os.ReadFile("talkkonnect-v4.xml")
	if err != nil {
		t.Fatal(err)
	}
	var shipped ConfigStruct
	if err := xml.Unmarshal(b, &shipped); err != nil {
		t.Fatal(err)
	}
	savedCmds, savedQ := Config.Global.Software.RemoteControl.HTTP.Command, Config.Global.Hardware.Quansheng
	defer func() {
		Config.Global.Software.RemoteControl.HTTP.Command, Config.Global.Hardware.Quansheng = savedCmds, savedQ
	}()
	Config.Global.Software.RemoteControl.HTTP.Command = shipped.Global.Software.RemoteControl.HTTP.Command
	q := &Config.Global.Hardware.Quansheng
	q.Enabled, q.URL, q.Token, q.TimeoutMsecs = true, ts.URL, "tok", 2000

	tk := &Talkkonnect{}
	rec := httptest.NewRecorder()
	tk.HandleRemoteAPICommand(rec, remoteAPIQuery{Command: "quanshengkeys"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing keys: %v %s", rec.Code, rec.Body)
	}
	rec = httptest.NewRecorder()
	tk.HandleRemoteAPICommand(rec, remoteAPIQuery{Command: "quanshengkeys", APIKeys: "menu up up menu"})
	if rec.Code != http.StatusOK || len(fq.keys) != 1 || fq.keys[0] != "menu up up menu" {
		t.Fatalf("keys: %v %s %v", rec.Code, rec.Body, fq.keys)
	}
}

func TestQuanshengMumbleToRadio(t *testing.T) {
	fq := &fakeQdock{}
	ts := httptest.NewServer(fq)
	defer ts.Close()

	saved := Config.Global.Hardware.Quansheng
	defer func() { Config.Global.Hardware.Quansheng = saved }()
	q := &Config.Global.Hardware.Quansheng
	q.Enabled, q.URL, q.Token, q.TimeoutMsecs = true, ts.URL, "tok", 2000
	q.MumbleToRadio.Enabled, q.MumbleToRadio.PTTTimeoutSecs, q.MumbleToRadio.HangTimeMsecs, q.MumbleToRadio.MaxTXSecs = true, 30, 150, 180

	quansheng.started.Store(true)
	quansheng.radioKeyed.Store(false)
	quansheng.radioRX.Store(false)
	defer quansheng.started.Store(false)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { quanshengMumbleToRadioWorker(ctx); close(done) }()

	waitFor := func(what string, cond func() bool) {
		t.Helper()
		for end := time.Now().Add(2 * time.Second); !cond(); time.Sleep(5 * time.Millisecond) {
			if time.Now().After(end) {
				t.Fatalf("timed out waiting for %s (ptt log %v)", what, fq.log())
			}
		}
	}

	QuanshengOnMumbleTalking()
	waitFor("key up", func() bool { return quansheng.radioKeyed.Load() })
	// a gap shorter than the hang time keeps the radio keyed
	QuanshengOnMumbleSilence()
	time.Sleep(50 * time.Millisecond)
	QuanshengOnMumbleTalking()
	time.Sleep(200 * time.Millisecond)
	if !quansheng.radioKeyed.Load() || len(fq.log()) != 1 {
		t.Fatalf("unkeyed during a short gap: %v", fq.log())
	}
	QuanshengOnMumbleSilence()
	waitFor("unkey after hang time", func() bool { return !quansheng.radioKeyed.Load() })
	if got := fq.log(); len(got) != 2 || got[0] != "on" || got[1] != "off" {
		t.Fatalf("ptt log %v", got)
	}

	// never key the radio while it is receiving (it would be relaying RF into mumble)
	quansheng.radioRX.Store(true)
	QuanshengOnMumbleTalking()
	time.Sleep(100 * time.Millisecond)
	if quansheng.radioKeyed.Load() || len(fq.log()) != 2 {
		t.Fatalf("keyed while radio receiving: %v", fq.log())
	}
	QuanshengOnMumbleSilence()
	quansheng.radioRX.Store(false)

	// shutdown unkeys a keyed radio
	QuanshengOnMumbleTalking()
	waitFor("key up again", func() bool { return quansheng.radioKeyed.Load() })
	cancel()
	<-done
	if got := fq.log(); got[len(got)-1] != "off" || quansheng.radioKeyed.Load() {
		t.Fatalf("not unkeyed on shutdown: %v", got)
	}
}
