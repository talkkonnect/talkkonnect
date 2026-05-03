package talkkonnect

import (
	"fmt"
	"io"
	"sync"
)

// sshRemoteReply mirrors human-readable command output to the SSH console while
// the daemon session is handling a line (see bottomCLIDispatchRemoteLine).
var sshRemoteReply struct {
	mu sync.Mutex
	w  io.Writer
}

func sshRemoteReplyAttach(w io.Writer) {
	sshRemoteReply.mu.Lock()
	sshRemoteReply.w = w
	sshRemoteReply.mu.Unlock()
}

func sshRemoteReplyDetach() {
	sshRemoteReply.mu.Lock()
	sshRemoteReply.w = nil
	sshRemoteReply.mu.Unlock()
}

func sshRemoteReplyF(format string, args ...interface{}) {
	sshRemoteReply.mu.Lock()
	w := sshRemoteReply.w
	sshRemoteReply.mu.Unlock()
	if w == nil {
		return
	}
	_, _ = fmt.Fprintf(w, format, args...)
}
