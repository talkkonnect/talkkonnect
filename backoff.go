/*
 * Exponential backoff helpers for Mumble and MQTT (1s initial, 2× per step, 8s cap).
 */

package talkkonnect

import (
	"context"
	"time"
)

// mumbleMQTTBackoffDelays are sleeps after failed attempts 1..4 before the next try (5 attempts total).
var mumbleMQTTBackoffDelays = []time.Duration{
	time.Second,
	2 * time.Second,
	4 * time.Second,
	8 * time.Second,
}

func sleepBackoff(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
