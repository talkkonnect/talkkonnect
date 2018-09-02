package volume

import (
	"errors"
	"math"

	"github.com/go-ole/go-ole"
	"github.com/moutend/go-wca"
)

// GetVolume returns the current volume (0 to 100).
func GetVolume() (int, error) {
	vol, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		var level float32
		err := aev.GetMasterVolumeLevelScalar(&level)
		vol := int(math.Floor(float64(level*100.0 + 0.5)))
		return vol, err
	})
	if vol == nil {
		return 0, err
	}
	return vol.(int), err
}

// SetVolume sets the sound volume to the specified value.
func SetVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("out of valid volume range")
	}
	_, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		err := aev.SetMasterVolumeLevelScalar(float32(volume)/100, nil)
		return nil, err
	})
	return err
}

// IncreaseVolume increases (or decreases) the audio volume by the specified value.
func IncreaseVolume(diff int) error {
	_, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		var level float32
		err := aev.GetMasterVolumeLevelScalar(&level)
		if err != nil {
			return nil, err
		}
		vol := math.Min(math.Max(math.Floor(float64(level*100.0+0.5))+float64(diff), 0.0), 100.0)
		err = aev.SetMasterVolumeLevelScalar(float32(vol)/100, nil)
		return nil, err
	})
	return err
}

// GetMuted returns the current muted status.
func GetMuted() (bool, error) {
	muted, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		var muted bool
		err := aev.GetMute(&muted)
		return muted, err
	})
	if muted == nil {
		return false, err
	}
	return muted.(bool), err
}

// Mute mutes the audio.
func Mute() error {
	_, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		err := aev.SetMute(true, nil)
		return nil, err
	})
	return err
}

// Unmute unmutes the audio.
func Unmute() error {
	_, err := invoke(func(aev *wca.IAudioEndpointVolume) (interface{}, error) {
		err := aev.SetMute(false, nil)
		return nil, err
	})
	return err
}

func invoke(f func(aev *wca.IAudioEndpointVolume) (interface{}, error)) (ret interface{}, err error) {
	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return
	}
	defer ole.CoUninitialize()

	var mmde *wca.IMMDeviceEnumerator
	if err = wca.CoCreateInstance(wca.CLSID_MMDeviceEnumerator, 0, wca.CLSCTX_ALL, wca.IID_IMMDeviceEnumerator, &mmde); err != nil {
		return
	}
	defer mmde.Release()

	var mmd *wca.IMMDevice
	if err = mmde.GetDefaultAudioEndpoint(wca.ERender, wca.EConsole, &mmd); err != nil {
		return
	}
	defer mmd.Release()

	var ps *wca.IPropertyStore
	if err = mmd.OpenPropertyStore(wca.STGM_READ, &ps); err != nil {
		return
	}
	defer ps.Release()

	var pv wca.PROPVARIANT
	if err = ps.GetValue(&wca.PKEY_Device_FriendlyName, &pv); err != nil {
		return
	}

	var aev *wca.IAudioEndpointVolume
	if err = mmd.Activate(wca.IID_IAudioEndpointVolume, wca.CLSCTX_ALL, nil, &aev); err != nil {
		return
	}
	defer aev.Release()

	ret, err = f(aev)
	return
}
