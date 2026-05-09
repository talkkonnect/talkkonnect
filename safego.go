/*
 * Safe goroutine wrapper: recover panics in non-critical workers without killing the process.
 */

package talkkonnect

import (
	"log"
	"runtime/debug"
)

// SafeGo runs fn in a new goroutine with panic recovery. Use for UI helpers, MQTT publish
// callbacks, GPIO timers, and other non-critical tasks so the Mumble/audio path keeps running.
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("alert: recovered panic in goroutine: %v\n%s", r, debug.Stack())
			}
		}()
		fn()
	}()
}
