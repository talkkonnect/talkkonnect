// +build !windows

package volume

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execCmd(cmdArgs []string) ([]byte, error) {
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Env = append(os.Environ(), cmdEnv()...)
	out, err := cmd.Output()
	if err != nil {
		err = fmt.Errorf(`failed to execute "%v" (%+v)`, strings.Join(cmdArgs, " "), err)
	}
	return out, err
}

// GetVolume returns the current volume (0 to 100).
func GetVolume() (int, error) {
	out, err := execCmd(getVolumeCmd())
	if err != nil {
		return 0, err
	}
	return parseVolume(string(out))
}

// SetVolume sets the sound volume to the specified value.
func SetVolume(volume int) error {
	if volume < 0 || 100 < volume {
		return errors.New("out of valid volume range")
	}
	_, err := execCmd(setVolumeCmd(volume))
	return err
}

// IncreaseVolume increases (or decreases) the audio volume by the specified value.
func IncreaseVolume(diff int) error {
	_, err := execCmd(increaseVolumeCmd(diff))
	return err
}

// GetMuted returns the current muted status.
func GetMuted() (bool, error) {
	out, err := execCmd(getMutedCmd())
	if err != nil {
		return false, err
	}
	return parseMuted(string(out))
}

// Mute mutes the audio.
func Mute() error {
	_, err := execCmd(muteCmd())
	return err
}

// Unmute unmutes the audio.
func Unmute() error {
	_, err := execCmd(unmuteCmd())
	return err
}
