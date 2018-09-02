package volume

import (
	"fmt"
	"math"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	origVolume, err := GetVolume()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get volume failed: %+v\n", err)
		os.Exit(1)
	}
	origMuted, err := GetMuted()
	if err != nil {
		fmt.Fprintf(os.Stderr, "get muted failed: %+v\n", err)
		os.Exit(1)
	}
	code := m.Run()
	_ = SetVolume(origVolume)
	if origMuted {
		_ = Mute()
	} else {
		_ = Unmute()
	}
	os.Exit(code)
}

func TestGetVolume(t *testing.T) {
	_, err := GetVolume()
	if err != nil {
		t.Errorf("get volume failed: %+v", err)
	}
}

func TestSetVolume(t *testing.T) {
	for _, vol := range []int{0, 37, 54, 20, 10} {
		err := SetVolume(vol)
		if err != nil {
			t.Errorf("set volume failed: %+v", err)
		}
		v, err := GetVolume()
		if err != nil {
			t.Errorf("get volume failed: %+v", err)
		}
		if vol != v {
			if math.Abs(float64(v-vol)) > 4 {
				t.Errorf("set volume failed: (got: %+v, expected: %+v)", v, vol)
			} else {
				t.Logf("set volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", v, vol)
			}
		}
	}
}

func TestIncreaseVolume(t *testing.T) {
	vol := 17
	diff := 3
	err := SetVolume(vol)
	if err != nil {
		t.Errorf("set volume failed: %+v", err)
	}
	err = IncreaseVolume(diff)
	if err != nil {
		t.Errorf("increase volume failed: %+v", err)
	}
	v, err := GetVolume()
	if err != nil {
		t.Errorf("get volume failed: %+v", err)
	}
	if v != vol+diff {
		if vol := vol + diff; math.Abs(float64(v-vol)) > 4 {
			t.Errorf("increase volume failed: (got: %+v, expected: %+v)", v, vol)
		} else {
			t.Logf("increase volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", v, vol)
		}
	}
	err = IncreaseVolume(-diff)
	if err != nil {
		t.Errorf("increase volume failed: %+v", err)
	}
	v, err = GetVolume()
	if err != nil {
		t.Errorf("get volume failed: %+v", err)
	}
	if v != vol {
		if math.Abs(float64(v-vol)) > 4 {
			t.Errorf("increase volume failed: (got: %+v, expected: %+v)", v, vol)
		} else {
			t.Logf("increase volume difference (possibly amixer on Linux): (got: %+v, expected: %+v)", v, vol)
		}
	}
	err = IncreaseVolume(-100)
	if err != nil {
		t.Errorf("increase volume failed: %+v", err)
	}
	v, err = GetVolume()
	if err != nil {
		t.Errorf("get volume failed: %+v", err)
	}
	if v != 0 {
		t.Errorf("increase volume failed: (got: %+v, expected: %+v)", v, 0)
	}
}

func TestMute(t *testing.T) {
	err := Mute()
	if err != nil {
		t.Errorf("mute failed: %+v", err)
	}
	muted, err := GetMuted()
	if err != nil {
		t.Errorf("get muted failed: %+v", err)
	}
	if !muted {
		t.Errorf("mute failed: %t", muted)
	}
}

func TestUnmute(t *testing.T) {
	err := Unmute()
	if err != nil {
		t.Errorf("unmute failed: %+v", err)
	}
	muted, err := GetMuted()
	if err != nil {
		t.Errorf("get muted failed: %+v", err)
	}
	if muted {
		t.Errorf("unmute failed: %t", muted)
	}
}
