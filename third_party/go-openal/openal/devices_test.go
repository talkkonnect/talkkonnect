package openal_test

import (
	"testing"

	"github.com/talkkonnect/go-openal/openal"
)

func TestPlaybackDevices(t *testing.T) {
	devices := openal.PlaybackDevices()
	if len(devices) == 0 {
		t.Fatal("expected at least one OpenAL playback device")
	}
	t.Logf("playback devices: %v", devices)
}

func TestCaptureDevices(t *testing.T) {
	devices := openal.CaptureDevices()
	if len(devices) == 0 {
		t.Fatal("expected at least one OpenAL capture device")
	}
	t.Logf("capture devices: %v", devices)
}

func TestOpenDefaultDevices(t *testing.T) {
	playback, err := openal.OpenDeviceChecked("")
	if err != nil {
		t.Fatalf("default playback device: %v", err)
	}
	defer playback.CloseDevice()

	capture, err := openal.CaptureOpenDeviceChecked("", 48000, openal.FormatMono16, 960)
	if err != nil {
		t.Fatalf("default capture device: %v", err)
	}
	defer capture.CaptureCloseDevice()
}

func TestOpenInvalidDevice(t *testing.T) {
	_, err := openal.OpenDeviceChecked("talkkonnect-invalid-device-name")
	if err == nil {
		t.Fatal("expected error opening invalid playback device")
	}
}
