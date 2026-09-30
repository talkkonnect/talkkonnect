package gumble

import (
	"sync"
	"testing"
)

// Permission() read client.permissions with no lock while the reader goroutine
// replaces and mutates that map under client.volatile (see handlers.go). This
// test is only meaningful under -race.
func TestChannelPermissionUnderConcurrentWrite(t *testing.T) {
	c := &Client{
		Channels:    make(Channels),
		permissions: make(map[uint32]*Permission),
	}
	ch := &Channel{ID: 1, Name: "test", client: c}
	c.Channels[1] = ch
	p := Permission(PermissionEnter)
	c.permissions[1] = &p

	var wg sync.WaitGroup
	wg.Add(2)

	// Stands in for handlers.go's permission-query handling.
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			c.volatile.Lock()
			c.permissions = make(map[uint32]*Permission)
			np := Permission(PermissionEnter)
			c.permissions[1] = &np
			c.volatile.Unlock()
		}
	}()

	// Stands in for any caller outside the event-handler goroutine.
	go func() {
		defer wg.Done()
		for i := 0; i < 2000; i++ {
			_ = ch.Permission()
		}
	}()

	wg.Wait()
}
