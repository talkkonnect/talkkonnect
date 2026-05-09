package talkkonnect

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
)

var mumbleReconnectMu sync.Mutex

// startSignalShutdownBridge listens for SIGINT/SIGTERM and cancels MasterCtx once.
func (b *Talkkonnect) startSignalShutdownBridge() {
	if b.MasterCtx == nil {
		return
	}
	SafeGo(func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigs)
		select {
		case sig := <-sigs:
			log.Printf("info: Received signal %v, initiating graceful shutdown\n", sig)
			if b.masterCancel != nil {
				b.masterCancel()
			}
		case <-b.MasterCtx.Done():
		}
	})
}

// connectMumbleWithBackoff dials the Mumble server with 1s,2s,4s,8s pauses between failed attempts (5 tries). On exhaustion returns an error.
func (b *Talkkonnect) connectMumbleWithBackoff(ctx context.Context) error {
	attempts := 1 + len(mumbleMQTTBackoffDelays)
	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			d := mumbleMQTTBackoffDelays[i-1]
			if err := sleepBackoff(ctx, d); err != nil {
				return err
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		lastErr = b.tryDialMumbleOnce()
		if lastErr == nil {
			return nil
		}
		log.Printf("warn: Mumble dial attempt %d/%d failed: %v\n", i+1, attempts, lastErr)
	}
	return fmt.Errorf("could not connect after exponential backoff (last error: %w)", lastErr)
}

func (b *Talkkonnect) tryDialMumbleOnce() error {
	_, err := gumble.Ping(b.Address, time.Second*1, time.Second*5)
	if err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	_, err = gumble.DialWithDialer(new(net.Dialer), b.Address, b.Config, &b.TLSConfig)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	log.Printf("info: Connected to Server Successfully\n")
	b.startConnectionContext()
	b.OpenStream()
	pstream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(""), 0)
	IsConnected = true
	return nil
}

func (b *Talkkonnect) destroyMumbleStreamIfPresent() {
	if pstream != nil {
		_ = pstream.Stop()
		pstream = nil
	}
	if b.Stream == nil {
		return
	}
	b.Destroy()
	b.Stream = nil
}

func (b *Talkkonnect) scheduleMumbleReconnect(reason string) {
	log.Printf("warn: Mumble session ended (%s); reconnecting with exponential backoff\n", reason)
	SafeGo(func() {
		mumbleReconnectMu.Lock()
		defer mumbleReconnectMu.Unlock()

		ctx := b.MasterCtx
		if ctx == nil {
			ctx = context.Background()
		}
		b.destroyMumbleStreamIfPresent()
		IsConnected = false

		if err := b.connectMumbleWithBackoff(ctx); err != nil {
			log.Printf("alert: Mumble reconnect failed: %v\n", err)
			FatalCleanUp("Unable to reconnect to Mumble after backoff: " + err.Error())
			return
		}
		if Register[AccountIndex] && b.Client != nil && !b.Client.Self.IsRegistered() {
			b.Client.Self.Register()
			log.Println("alert: Client Is Now Registered")
		}
		log.Println("info: Mumble session restored after disconnect")
	})
}
