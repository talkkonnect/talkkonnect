package openal

import "strings"

// ParseDeviceList splits a null-separated OpenAL device list.
func ParseDeviceList(specifier string) []string {
	if specifier == "" {
		return nil
	}
	parts := strings.Split(specifier, "\x00")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

// PlaybackDevices returns available OpenAL playback device names.
func PlaybackDevices() []string {
	param := uint32(DeviceSpecifier)
	if IsALCExtensionPresent(ExtensionEnumerateAll) {
		param = AllDevicesSpecifier
	}
	return ParseDeviceList(GetDeviceString(nil, param))
}

// CaptureDevices returns available OpenAL capture device names.
func CaptureDevices() []string {
	return ParseDeviceList(GetDeviceString(nil, CaptureDeviceSpecifier))
}
