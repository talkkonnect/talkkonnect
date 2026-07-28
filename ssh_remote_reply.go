package talkkonnect

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// maxReplyCaptureBytes caps how much reply text one replyCapture keeps, so a
// chatty command (a long channel list, a big server list) cannot grow a request
// buffer without bound.
const maxReplyCaptureBytes = 64 << 10

// sshRemoteReply mirrors human-readable command output to every attached console
// writer while a command is running. Two listeners can be attached at once - an
// SSH daemon session (bottomCLIDispatchRemoteLine) and an HTTP API request
// (HandleRemoteAPICommand) - so the registry holds one writer per attachment
// instead of a single global writer.
var sshRemoteReply struct {
	mu      sync.Mutex
	nextID  int64
	writers map[int64]io.Writer
}

// sshRemoteReplyAttach registers w to receive command reply text and returns the
// token sshRemoteReplyDetach needs to unregister it again.
func sshRemoteReplyAttach(w io.Writer) int64 {
	sshRemoteReply.mu.Lock()
	defer sshRemoteReply.mu.Unlock()
	if sshRemoteReply.writers == nil {
		sshRemoteReply.writers = make(map[int64]io.Writer)
	}
	sshRemoteReply.nextID++
	id := sshRemoteReply.nextID
	sshRemoteReply.writers[id] = w
	return id
}

func sshRemoteReplyDetach(id int64) {
	sshRemoteReply.mu.Lock()
	delete(sshRemoteReply.writers, id)
	sshRemoteReply.mu.Unlock()
}

// sshRemoteReplyF fans one reply line out to every attached writer. The write
// happens under the lock so a writer is never used after its
// sshRemoteReplyDetach has returned: commands may emit reply lines from
// background goroutines that outlive the request which attached the writer.
func sshRemoteReplyF(format string, args ...interface{}) {
	sshRemoteReply.mu.Lock()
	defer sshRemoteReply.mu.Unlock()
	for _, w := range sshRemoteReply.writers {
		_, _ = fmt.Fprintf(w, format, args...)
	}
}

// replyCapture collects reply text for a caller that wants it as a string rather
// than as a console stream: the HTTP API hands the text back in its response body
// so dashboards such as tk-webmonitor can show what the command actually said.
// It is safe for concurrent use because commands may emit from goroutines.
type replyCapture struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (c *replyCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.buf.Len() >= maxReplyCaptureBytes {
		return len(p), nil
	}
	return c.buf.Write(p)
}

func (c *replyCapture) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}
