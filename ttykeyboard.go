/*
 * ttykeyboard.go -> evdev listener for built-in / console keyboard devices.
 */

package talkkonnect

import "log"

func (b *Talkkonnect) TTYKeyboard() {
	if len(TTYKeyMap) == 0 {
		return
	}
	if len(Config.Global.Hardware.TTYKeyboard.TTYKeyboardPaths) == 0 {
		log.Println("warn: TTY keyboard bindings are configured but no ttykeyboarddevpath is set; TTY keyboard disabled")
		return
	}
	for _, devicePath := range Config.Global.Hardware.TTYKeyboard.TTYKeyboardPaths {
		go b.evdevKeyboardListener(devicePath, "TTY Keyboard", TTYKeyMap, Config.Global.Hardware.TTYKeyboard.NumlockScanID)
	}
}
