// +build !windows,!darwin

package volume

import (
	"errors"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var useAmixer bool

func init() {
	if _, err := exec.LookPath("pactl"); err != nil {
		useAmixer = true
	}
}

func cmdEnv() []string {
	return []string{"LANG=C", "LC_ALL=C"}
}

func getVolumeCmd() []string {
	if useAmixer {
		// Modified From Master to Speaker by Suvir
		//return []string{"amixer", "get", "Master"}
		return []string{"amixer", "get", "Speaker"}
	}
	return []string{"pactl", "list", "sinks"}
}

var volumePattern = regexp.MustCompile(`\d+%`)

func parseVolume(out string) (int, error) {
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		s := strings.TrimLeft(line, " \t")
		if useAmixer && strings.Contains(s, "Playback") && strings.Contains(s, "%") ||
			!useAmixer && strings.HasPrefix(s, "Volume:") {
			volumeStr := volumePattern.FindString(s)
			return strconv.Atoi(volumeStr[:len(volumeStr)-1])
		}
	}
	return 0, errors.New("no volume found")
}

func setVolumeCmd(volume int) []string {
	if useAmixer {
		return []string{"amixer", "set", "Master", strconv.Itoa(volume) + "%"}
	}
	return []string{"pactl", "set-sink-volume", "0", strconv.Itoa(volume) + "%"}
}

func increaseVolumeCmd(diff int) []string {
	var sign string
	if diff >= 0 {
		sign = "+"
	} else if useAmixer {
		diff = -diff
		sign = "-"
	}
	if useAmixer {
		return []string{"amixer", "set", "Master", strconv.Itoa(diff) + "%" + sign}
	}
	return []string{"pactl", "--", "set-sink-volume", "0", sign + strconv.Itoa(diff) + "%"}
}

func getMutedCmd() []string {
	if useAmixer {
		// Change Master to Speaker By Suvir Kumar
		//return []string{"amixer", "get", "Master"}
		return []string{"amixer", "get", "Speaker"}
	}
	return []string{"pactl", "list", "sinks"}
}

func parseMuted(out string) (bool, error) {
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		s := strings.TrimLeft(line, " \t")
		if useAmixer && strings.Contains(s, "Playback") && strings.Contains(s, "%") ||
			!useAmixer && strings.HasPrefix(s, "Mute: ") {
			if strings.Contains(s, "[off]") || strings.Contains(s, "yes") {
				return true, nil
			} else if strings.Contains(s, "[on]") || strings.Contains(s, "no") {
				return false, nil
			}
		}
	}
	return false, errors.New("no muted information found")
}

func muteCmd() []string {
	if useAmixer {
		// Changed Master to Speaker
		//return []string{"amixer", "", "", "set", "Speaker", "mute"}
		return []string{"amixer", "set", "Speaker", "mute"}
	}
	return []string{"pactl", "set-sink-mute", "0", "1"}
}

func unmuteCmd() []string {
	if useAmixer {
		//Changed Master to Speaker
		return []string{"amixer", "set", "Speaker", "unmute"}
	}
	return []string{"pactl", "set-sink-mute", "0", "0"}
}
