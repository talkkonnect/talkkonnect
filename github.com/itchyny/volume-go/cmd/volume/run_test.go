package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/itchyny/volume-go"
)

func TestMain(m *testing.M) {
	origVolume, err := volume.GetVolume()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get volume failed: %+v\n", err)
		os.Exit(1)
	}
	origMuted, err := volume.GetMuted()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get muted failed: %+v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	_ = volume.SetVolume(origVolume)
	if origMuted {
		_ = volume.Mute()
	} else {
		_ = volume.Unmute()
	}
	os.Exit(code)
}

func TestRunVersion(t *testing.T) {
	out := new(bytes.Buffer)
	if err := run([]string{"version"}, out); err != nil {
		t.Errorf("version failed: %+v", err)
	}
	if !strings.Contains(out.String(), "volume version ") {
		t.Errorf("version failed: (got: %+v, expected to contain: %+v)", out.String(), "volume version ")
	}
}

func TestRunHelp(t *testing.T) {
	out := new(bytes.Buffer)
	if err := run([]string{"help"}, out); err != nil {
		t.Errorf("version failed: %+v", err)
	}
	for _, expected := range []string{"USAGE:", "COMMANDS:", "VERSION:", "AUTHOR:"} {
		if !strings.Contains(out.String(), expected) {
			t.Errorf("version failed: (got: %+v, expected to contain: %+v)", out.String(), expected)
		}
	}
}

func TestRunStatus(t *testing.T) {
	_ = volume.SetVolume(17)
	_ = volume.Unmute()
	out := new(bytes.Buffer)
	if err := run([]string{"status"}, out); err != nil {
		t.Errorf("status failed: %+v", err)
	}
	got := out.String()
	expected := []string{"volume: 17\n", "muted: false\n"}
	if !strings.Contains(got, expected[0]) {
		v, _ := strconv.Atoi(strings.Split(got, "\n")[0][8:])
		if vol := 17; math.Abs(float64(v-vol)) > 4 {
			t.Errorf("get status failed: (got: %+v, expected: %+v)", got, strings.Join(expected, ""))
		} else {
			t.Logf("get status difference (possibly amixer on Linux): (got: %+v, expected: %+v)", got, strings.Join(expected, ""))
		}
	}
	if !strings.Contains(got, expected[1]) {
		t.Errorf("get status failed: %q should contain %q", got, strings.Join(expected, ""))
	}
}

func TestRunGetSet(t *testing.T) {
	out := new(bytes.Buffer)
	if err := run([]string{"set", "13"}, out); err != nil {
		t.Errorf("set volume failed: %+v", err)
	}
	if err := run([]string{"get"}, out); err != nil {
		t.Errorf("get volume failed: %+v", err)
	}
	expected := "13\n"
	if out.String() != expected {
		t.Errorf("set volume failed: (got: %+v, expected: %+v)", out.String(), expected)
	}
}

func TestRunUp(t *testing.T) {
	_ = volume.SetVolume(17)
	{
		if err := run([]string{"up"}, nil); err != nil {
			t.Errorf("up volume failed: %+v", err)
		}
		vol, _ := volume.GetVolume()
		expected := 17 + 6
		if vol != expected {
			if math.Abs(float64(vol-expected)) > 4 {
				t.Errorf("up volume failed: (got: %+v, expected: %+v)", vol, expected)
			} else {
				t.Logf("up volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", vol, expected)
			}
		}
	}
	{
		if err := run([]string{"up", "3"}, nil); err != nil {
			t.Errorf("up volume failed: %+v", err)
		}
		vol, _ := volume.GetVolume()
		expected := 17 + 6 + 3
		if vol != expected {
			if math.Abs(float64(vol-expected)) > 4 {
				t.Errorf("up volume failed: (got: %+v, expected: %+v)", vol, expected)
			} else {
				t.Logf("up volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", vol, expected)
			}
		}
	}
}

func TestRunDown(t *testing.T) {
	_ = volume.SetVolume(17)
	{
		if err := run([]string{"down"}, nil); err != nil {
			t.Errorf("down volume failed: %+v", err)
		}
		vol, _ := volume.GetVolume()
		expected := 17 - 6
		if vol != expected {
			if math.Abs(float64(vol-expected)) > 4 {
				t.Errorf("down volume failed: (got: %+v, expected: %+v)", vol, expected)
			} else {
				t.Logf("down volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", vol, expected)
			}
		}
	}
	{
		if err := run([]string{"down", "3"}, nil); err != nil {
			t.Errorf("down volume failed: %+v", err)
		}
		vol, _ := volume.GetVolume()
		expected := 17 - 6 - 3
		if vol != expected {
			if math.Abs(float64(vol-expected)) > 4 {
				t.Errorf("down volume failed: (got: %+v, expected: %+v)", vol, expected)
			} else {
				t.Logf("down volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", vol, expected)
			}
		}
	}
}

func TestRunMute(t *testing.T) {
	_ = volume.Unmute()
	if err := run([]string{"mute"}, nil); err != nil {
		t.Errorf("mute failed: %+v", err)
	}
	muted, _ := volume.GetMuted()
	expected := true
	if muted != expected {
		t.Errorf("mute failed: (got: %+v, expected: %+v)", muted, expected)
	}
}

func TestRunUnmute(t *testing.T) {
	_ = volume.Mute()
	if err := run([]string{"unmute"}, nil); err != nil {
		t.Errorf("unmute failed: %+v", err)
	}
	muted, _ := volume.GetMuted()
	expected := false
	if muted != expected {
		t.Errorf("usermute failed: (got: %+v, expected: %+v)", muted, expected)
	}
}
