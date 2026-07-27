package talkkonnect

import (
	"testing"
	"time"
)

func TestScanDwellAndHangDefaults(t *testing.T) {
	defer func() {
		Config.Global.Software.ChannelScan.DwellTimeMsecs = 0
		Config.Global.Software.ChannelScan.HangTimeMsecs = 0
	}()

	Config.Global.Software.ChannelScan.DwellTimeMsecs = 0
	Config.Global.Software.ChannelScan.HangTimeMsecs = 0
	if got, want := scanDwellDuration(), scanDefaultDwellMsecs*time.Millisecond; got != want {
		t.Errorf("unset dwell time = %v, want %v", got, want)
	}
	if got, want := scanHangDuration(), scanDefaultHangMsecs*time.Millisecond; got != want {
		t.Errorf("unset hang time = %v, want %v", got, want)
	}

	Config.Global.Software.ChannelScan.DwellTimeMsecs = 10
	Config.Global.Software.ChannelScan.HangTimeMsecs = 10
	if got, want := scanDwellDuration(), scanMinDwellMsecs*time.Millisecond; got != want {
		t.Errorf("too short dwell time = %v, want clamp to %v", got, want)
	}
	if got, want := scanHangDuration(), scanMinHangMsecs*time.Millisecond; got != want {
		t.Errorf("too short hang time = %v, want clamp to %v", got, want)
	}

	Config.Global.Software.ChannelScan.DwellTimeMsecs = 1500
	Config.Global.Software.ChannelScan.HangTimeMsecs = 2500
	if got, want := scanDwellDuration(), 1500*time.Millisecond; got != want {
		t.Errorf("configured dwell time = %v, want %v", got, want)
	}
	if got, want := scanHangDuration(), 2500*time.Millisecond; got != want {
		t.Errorf("configured hang time = %v, want %v", got, want)
	}
}

func TestScanSkipped(t *testing.T) {
	defer func() { Config.Global.Software.ChannelScan.SkipChannels = "" }()

	Config.Global.Software.ChannelScan.SkipChannels = " root , 42 "

	cases := []struct {
		channel scanChannelStruct
		want    bool
	}{
		{scanChannelStruct{chanID: 0, chanName: "Root"}, true},    // name match is case insensitive
		{scanChannelStruct{chanID: 42, chanName: "HAM-CB"}, true}, // channel id match
		{scanChannelStruct{chanID: 7, chanName: "HAM-CB"}, false},
	}

	for _, c := range cases {
		if got := scanSkipped(c.channel); got != c.want {
			t.Errorf("scanSkipped(%v) = %v, want %v", c.channel, got, c.want)
		}
	}

	Config.Global.Software.ChannelScan.SkipChannels = ""
	if scanSkipped(scanChannelStruct{chanID: 1, chanName: "Root"}) {
		t.Error("empty skip list should not skip any channel")
	}
}

func TestScanVoiceActivity(t *testing.T) {
	scanVoiceActivity.Range(func(key, _ interface{}) bool {
		scanVoiceActivity.Delete(key)
		return true
	})

	if scanVoiceActiveOn("HAM-CB", time.Second) {
		t.Error("channel without traffic reported as active")
	}

	noteScanVoiceActivity("HAM-CB")
	if !scanVoiceActiveOn("HAM-CB", time.Second) {
		t.Error("channel with fresh traffic reported as idle")
	}
	if scanVoiceActiveOn("NakonNayok", time.Second) {
		t.Error("traffic leaked to another channel")
	}
	if scanVoiceActiveOn("HAM-CB", time.Nanosecond) {
		t.Error("traffic older than the hang time reported as active")
	}

	noteScanVoiceActivity("")
	if _, found := scanVoiceActivity.Load(""); found {
		t.Error("empty channel name should not be recorded")
	}
}

func TestScanDwellOnQuietChannel(t *testing.T) {
	b, restore := scanDwellTestSetup(t, 600, 500)
	defer restore()

	stop := make(chan struct{})
	start := time.Now()
	if !b.scanDwellOn(scanChannelStruct{chanID: 1, chanName: "HAM-CB"}, stop) {
		t.Fatal("dwell on a quiet channel should carry on scanning")
	}
	if elapsed := time.Since(start); elapsed < 500*time.Millisecond {
		t.Errorf("left the channel after %v, want at least the dwell time", elapsed)
	}
	if ScanIsHolding() {
		t.Error("scanner should not be holding on a quiet channel")
	}
}

func TestScanHoldsOnBusyChannel(t *testing.T) {
	b, restore := scanDwellTestSetup(t, 300, 500)
	defer restore()

	stop := make(chan struct{})
	busyUntil := time.Now().Add(700 * time.Millisecond)
	go func() {
		for time.Now().Before(busyUntil) {
			noteScanVoiceActivity("HAM-CB")
			time.Sleep(50 * time.Millisecond)
		}
	}()

	start := time.Now()
	if !b.scanDwellOn(scanChannelStruct{chanID: 1, chanName: "HAM-CB"}, stop) {
		t.Fatal("dwell should carry on scanning after the traffic stopped")
	}
	// held for as long as there was traffic and then for the hang time on top
	if elapsed := time.Since(start); elapsed < 1100*time.Millisecond {
		t.Errorf("left the busy channel after %v, want traffic time plus hang time", elapsed)
	}
	if ScanIsHolding() {
		t.Error("holding flag should be cleared when the scan carries on")
	}
}

func TestScanDwellStops(t *testing.T) {
	b, restore := scanDwellTestSetup(t, 5000, 500)
	defer restore()

	stop := make(chan struct{})
	close(stop)

	if b.scanDwellOn(scanChannelStruct{chanID: 1, chanName: "HAM-CB"}, stop) {
		t.Error("dwell should report a stop request instead of carrying on")
	}
}

func scanDwellTestSetup(t *testing.T, dwellMsecs int, hangMsecs int) (*Talkkonnect, func()) {
	t.Helper()

	connected := IsConnected
	dwell := Config.Global.Software.ChannelScan.DwellTimeMsecs
	hang := Config.Global.Software.ChannelScan.HangTimeMsecs

	IsConnected = true
	Config.Global.Software.ChannelScan.DwellTimeMsecs = dwellMsecs
	Config.Global.Software.ChannelScan.HangTimeMsecs = hangMsecs
	scanVoiceActivity.Range(func(key, _ interface{}) bool {
		scanVoiceActivity.Delete(key)
		return true
	})

	return &Talkkonnect{}, func() {
		IsConnected = connected
		Config.Global.Software.ChannelScan.DwellTimeMsecs = dwell
		Config.Global.Software.ChannelScan.HangTimeMsecs = hang
		scanHolding.Store(false)
	}
}

func TestScanNotRunningWhenDisconnected(t *testing.T) {
	connected := IsConnected
	defer func() { IsConnected = connected }()

	IsConnected = false
	b := &Talkkonnect{}
	b.Scan()

	if ScanIsRunning() {
		t.Error("scan started while disconnected")
	}
}
