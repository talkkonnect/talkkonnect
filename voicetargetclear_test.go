package talkkonnect

import (
	"bytes"
	"strings"
	"testing"

	"github.com/talkkonnect/gumble/gumble"
)

// TestSendVoiceTargetsZeroClears covers the case that regressed: target 0 is the
// "talk to the joined channel" state rather than a slot under <voicetargets>, so
// cmdSendVoiceTargets(0) has to clear on its own instead of scanning the config
// for an id it will never find.
func TestSendVoiceTargetsZeroClears(t *testing.T) {
	savedConfig := Config
	savedConnected := IsConnected
	savedUI := uiVoiceTargetSnapshot()
	t.Cleanup(func() {
		Config = savedConfig
		IsConnected = savedConnected
		restoreUIVoiceTarget(savedUI)
	})

	// No account declares a target 0, which is the normal case, so scanning the
	// config can never satisfy the request.
	Config = ConfigStruct{}
	IsConnected = false

	RecordUIVoiceTarget(3, "channel", "Operations")

	b := &Talkkonnect{}
	b.cmdSendVoiceTargets(0)

	if active := uiVoiceTargetSnapshot(); active.ID != 0 || active.Kind != "" || len(active.Names) != 0 {
		t.Fatalf("voicetargetset 0 left target %+v active, want cleared", active)
	}
}

// TestVoiceTargetCLITabComplete checks the "vt" completions, including that the
// ids offered for "vt set" come from the config rather than a fixed list.
func TestVoiceTargetCLITabComplete(t *testing.T) {
	savedConfig := Config
	t.Cleanup(func() { Config = savedConfig })
	Config = vtParse(t, vtTestDoc) // account 0 is default and has targets 1 and 3

	if got := voiceTargetConfiguredIDs(); strings.Join(got, ",") != "0,1,3" {
		t.Errorf("configured ids = %v, want [0 1 3]", got)
	}

	cases := []struct {
		line     string
		wantLine string
		wantBell bool
	}{
		{"vt", "vt ", false},
		{"vt s", "vt set ", false},
		{"vt cl", "vt clear ", false},
		{"vt wh", "vt whisper ", false},
		{"vt set 3", "vt set 3 ", false},
		{"vt set 9", "vt set 9", true},   // not configured, nothing to complete
		{"vt zz", "vt zz", true},         // no such subcommand
		{"vt list x", "vt list x", true}, // list takes no argument
	}
	for _, tc := range cases {
		gotLine, gotBell := bottomCLITabCompleteVT(tc.line)
		if gotLine != tc.wantLine || gotBell != tc.wantBell {
			t.Errorf("complete(%q) = (%q, %v), want (%q, %v)", tc.line, gotLine, gotBell, tc.wantLine, tc.wantBell)
		}
	}
}

// TestVoiceTargetCLISetRejectsUnconfiguredID keeps "vt set" from looking like it
// worked when the id has no entry for the account in use, which is all
// cmdSendVoiceTargets would do on its own.
func TestVoiceTargetCLISetRejectsUnconfiguredID(t *testing.T) {
	savedConfig := Config
	savedConnected := IsConnected
	savedUI := uiVoiceTargetSnapshot()
	t.Cleanup(func() {
		Config = savedConfig
		IsConnected = savedConnected
		restoreUIVoiceTarget(savedUI)
	})

	Config = vtParse(t, vtTestDoc)
	IsConnected = true
	RecordUIVoiceTarget(1, "user", "zoran-laptop")

	// A non-nil Client is enough: the id check runs before the client is touched.
	b := &Talkkonnect{Client: &gumble.Client{}}

	var out bytes.Buffer
	b.voiceTargetCLISet(&out, []string{"9"})

	if !strings.Contains(out.String(), "no voice target 9") {
		t.Errorf("output = %q, want a complaint about target 9 not being configured", out.String())
	}
	if active := uiVoiceTargetSnapshot(); active.ID != 1 {
		t.Errorf("active target = %+v, want the previous target 1 untouched", active)
	}
}
