package talkkonnect

import (
	"encoding/xml"
	"os"
	"testing"
)

// shippedConfigs are the sample configs a user starts from. A command handler is
// only reachable over the HTTP API when one of these declares it, so they are the
// other half of the API surface and drift between them and the handler map is what
// this file guards. talkkonnect.xml is the working config and is gitignored, so it
// is checked when present and skipped when it is not — talkkonnect-v4.xml is
// tracked and always covers a fresh clone.
var shippedConfigs = []string{"talkkonnect.xml", "talkkonnect-v4.xml"}

// eachShippedConfig runs check over every sample config that exists, and fails if
// none did — an empty sweep would otherwise pass every test in this file.
func eachShippedConfig(t *testing.T, check func(path string, cfg ConfigStruct)) {
	t.Helper()
	checked := 0
	for _, path := range shippedConfigs {
		data, err := os.ReadFile(path)
		if os.IsNotExist(err) {
			t.Logf("skipping %s: not present", path)
			continue
		}
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		var cfg ConfigStruct
		if err := xml.Unmarshal(data, &cfg); err != nil {
			t.Fatalf("unmarshal %s: %v", path, err)
		}
		checked++
		check(path, cfg)
	}
	if checked == 0 {
		t.Fatalf("none of %v were readable", shippedConfigs)
	}
}

// TestShippedHTTPCommandsHaveHandlers catches a <command action="…"> that no
// built-in handler answers — a typo or a renamed handler. HandleRemoteAPICommand
// rejects those with a 400 at call time, which reads like a broken API rather than
// a broken config, so fail here instead.
func TestShippedHTTPCommandsHaveHandlers(t *testing.T) {
	handlers := (&Talkkonnect{}).remoteAPICommandHandlers()

	eachShippedConfig(t, func(path string, cfg ConfigStruct) {
		commands := cfg.Global.Software.RemoteControl.HTTP.Command
		if len(commands) == 0 {
			t.Fatalf("%s declares no <http> commands", path)
		}
		for _, cmd := range commands {
			if _, ok := handlers[cmd.Action]; !ok {
				t.Errorf("%s: <command action=%q> has no built-in handler", path, cmd.Action)
			}
		}
	})
}

// TestShippedHTTPCommandsCoverWebMonitor pins the commands tk-webmonitor's clickable
// UI needs. Each one addresses something by name or flips a mode, so none of them
// can be reached by the step-through buttons if the declaration goes missing — the
// symptom is a 404 on one control while the rest of the dashboard keeps working.
func TestShippedHTTPCommandsCoverWebMonitor(t *testing.T) {
	wanted := []string{
		"joinchannel",
		"whisperuser",
		"whisperclear",
		"setrxvolume",
		"multicasttoggle",
	}

	eachShippedConfig(t, func(path string, cfg ConfigStruct) {
		declared := make(map[string]bool)
		for _, cmd := range cfg.Global.Software.RemoteControl.HTTP.Command {
			declared[cmd.Action] = true
		}
		for _, action := range wanted {
			if !declared[action] {
				t.Errorf("%s: missing <command action=%q>", path, action)
			}
		}
	})
}

// TestMulticastCommandsHaveNoParams checks the multicast handlers stay parameterless.
// HandleRemoteAPICommand picks its call path from funcparamname: an empty one calls
// the handler with no arguments, and b.Call fails with "the number of params is not
// adapted" — a 500 — if the handler ever grows one.
func TestMulticastCommandsHaveNoParams(t *testing.T) {
	eachShippedConfig(t, func(path string, cfg ConfigStruct) {
		for _, cmd := range cfg.Global.Software.RemoteControl.HTTP.Command {
			switch cmd.Action {
			case "multicaston", "multicastoff", "multicasttoggle":
				if cmd.Funcparamname != "" {
					t.Errorf("%s: %s declares funcparamname=%q, want empty", path, cmd.Action, cmd.Funcparamname)
				}
			}
		}
	})
}
