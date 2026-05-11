/*
 * Remote SSH console when talkkonnect runs as a go-daemon child (no GNU screen).
 * github.com/talkkonnect/gosshd attaches with "screen -xS tk", which fails without screen.
 */

package talkkonnect

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/crypto/ssh"
)

func (b *Talkkonnect) runRemoteSSHConsoleDaemon() {
	cfg := Config.Global.Software.RemoteSSHConsole
	sshDaemonListenEmbeddedConsole(b, cfg.Username, cfg.Password, cfg.IDRSAFile, cfg.Listen)
}

func sshDaemonListenEmbeddedConsole(b *Talkkonnect, username, password, idrsafile, listenOn string) {
	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if c.User() == username && string(pass) == password {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},
	}

	privateBytes, err := os.ReadFile(idrsafile)
	if err != nil {
		log.Printf("Remote SSH console: failed to load private key (%v): %v", idrsafile, err)
		return
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Printf("Remote SSH console: failed to parse private key: %v", err)
		return
	}
	serverConfig.AddHostKey(private)

	listener, err := net.Listen("tcp", listenOn)
	if err != nil {
		log.Printf("Remote SSH console: failed to listen on %v: %v", listenOn, err)
		return
	}
	defer listener.Close()
	log.Printf("talkkonnect remote ssh console (daemon) listening on %v\n", listenOn)

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Remote SSH console: accept error: %v", err)
			continue
		}
		go handleDaemonSSHConn(b, tcpConn, serverConfig)
	}
}

func handleDaemonSSHConn(b *Talkkonnect, tcpConn net.Conn, serverConfig *ssh.ServerConfig) {
	defer tcpConn.Close()
	sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, serverConfig)
	if err != nil {
		log.Printf("Remote SSH console: handshake failed: %v", err)
		return
	}
	defer sshConn.Close()
	log.Printf("New SSH connection to talkkonnect daemon console from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
	go ssh.DiscardRequests(reqs)
	for newChannel := range chans {
		go handleDaemonSSHChannel(b, newChannel)
	}
}

func sshParsePtyReq(payload []byte) (cols, rows uint32, ok bool) {
	if len(payload) < 4 {
		return 0, 0, false
	}
	tlen := int(binary.BigEndian.Uint32(payload[0:4]))
	if tlen < 0 || len(payload) < 4+tlen+8 {
		return 0, 0, false
	}
	p := payload[4+tlen:]
	cols = binary.BigEndian.Uint32(p[0:4])
	rows = binary.BigEndian.Uint32(p[4:8])
	return cols, rows, true
}

func sshParseWindowChange(payload []byte) (cols, rows uint32, ok bool) {
	if len(payload) < 8 {
		return 0, 0, false
	}
	cols = binary.BigEndian.Uint32(payload[0:4])
	rows = binary.BigEndian.Uint32(payload[4:8])
	return cols, rows, true
}

// sshDaemonConsoleState keeps the same DECSTBM fixed-bottom layout as the local bottom CLI.
type sshDaemonConsoleState struct {
	out     *sshSyncedChannelWriter
	rows    atomic.Int32
	draft   []byte
	draftMu sync.Mutex
	logSk   *bottomCLISSHLogSink
}

func (s *sshDaemonConsoleState) rowsSnap() int {
	r := int(s.rows.Load())
	if r < 2 {
		return 24
	}
	return r
}

func (s *sshDaemonConsoleState) draftSnap() []byte {
	s.draftMu.Lock()
	defer s.draftMu.Unlock()
	return append([]byte(nil), s.draft...)
}

func (s *sshDaemonConsoleState) redrawLine(line []byte) {
	s.draftMu.Lock()
	s.draft = append([]byte(nil), line...)
	s.draftMu.Unlock()
	rs := s.rowsSnap()
	if rs < 2 {
		return
	}
	s.out.WithWireLock(func(w io.Writer) {
		bottomCLIPromptSSHWire(w, rs, bottomCLIPromptLabel(), line)
	})
}

func (s *sshDaemonConsoleState) applyScrollLayout() {
	rs := s.rowsSnap()
	if rs < 2 {
		return
	}
	s.out.WithWireLock(func(w io.Writer) {
		bottomCLIApplyScrollLayoutW(w, rs)
	})
}

func (s *sshDaemonConsoleState) clearAndRelayout() {
	rs := s.rowsSnap()
	s.out.WithWireLock(func(w io.Writer) {
		_, _ = w.Write([]byte("\033[2J\033[1;1H"))
		bottomCLIApplyScrollLayoutW(w, rs)
	})
	s.redrawLine(nil)
}

func (s *sshDaemonConsoleState) onResize(cols, rows int) {
	_ = cols
	if rows < 2 || rows > 4000 {
		return
	}
	s.rows.Store(int32(rows))
	s.applyScrollLayout()
	line := s.draftSnap()
	s.redrawLine(line)
}

func (s *sshDaemonConsoleState) registerLog() {
	s.logSk = bottomCLISSHRegisterLogSink(s.out, s.rowsSnap, s.draftSnap)
}

func (s *sshDaemonConsoleState) unregisterLog() {
	bottomCLISSHUnregisterLogSink(s.logSk)
	s.logSk = nil
}

func (s *sshDaemonConsoleState) auxWriter() io.Writer {
	if s == nil {
		return io.Discard
	}
	return &sshBottomAuxWriter{s: s}
}

// sshBottomAuxWriter routes command output into the scrolling region above the prompt.
type sshBottomAuxWriter struct{ s *sshDaemonConsoleState }

func (a *sshBottomAuxWriter) Write(p []byte) (int, error) {
	if a == nil || a.s == nil || len(p) == 0 {
		return len(p), nil
	}
	n := len(p)
	rs := a.s.rowsSnap()
	for _, chunk := range bytes.Split(p, []byte{'\n'}) {
		if len(chunk) == 0 {
			continue
		}
		line := append(append([]byte(nil), chunk...), '\n')
		a.s.out.WithWireLock(func(w io.Writer) {
			if rs >= 2 {
				bottomCLIMoveToScrollRegionBottomW(w, rs)
				_, _ = w.Write(sshBytesToCRLF(line))
				d := a.s.draftSnap()
				bottomCLIPromptSSHWire(w, rs, bottomCLIPromptLabel(), d)
			} else {
				_, _ = w.Write(sshBytesToCRLF(line))
			}
		})
	}
	return n, nil
}

var sshDaemonConsoleActive atomic.Pointer[sshDaemonConsoleState]

func handleDaemonSSHChannel(b *Talkkonnect, newChannel ssh.NewChannel) {
	if t := newChannel.ChannelType(); t != "session" {
		_ = newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	connection, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Remote SSH console: could not accept channel: %v", err)
		return
	}
	defer connection.Close()

	var termRows atomic.Int32
	termRows.Store(24)

	// Do not write to the session channel until the client has received a successful
	// shell reply; PuTTY and Windows OpenSSH otherwise keep input buffered or disabled.
	shellReady := make(chan struct{}, 1)
	go func() {
		for req := range requests {
			switch req.Type {
			case "env":
				if req.WantReply {
					_ = req.Reply(true, nil)
				}
			case "pty-req":
				if cols, rows, ok := sshParsePtyReq(req.Payload); ok && rows >= 2 && rows <= 4000 {
					termRows.Store(int32(rows))
					if c := sshDaemonConsoleActive.Load(); c != nil {
						c.onResize(int(cols), int(rows))
					}
				}
				if req.WantReply {
					_ = req.Reply(true, nil)
				}
			case "shell":
				if req.WantReply {
					_ = req.Reply(true, nil)
				}
				select {
				case shellReady <- struct{}{}:
				default:
				}
			case "window-change":
				if cols, rows, ok := sshParseWindowChange(req.Payload); ok && rows >= 2 && rows <= 4000 {
					termRows.Store(int32(rows))
					if c := sshDaemonConsoleActive.Load(); c != nil {
						c.onResize(int(cols), int(rows))
					}
				}
				if req.WantReply {
					_ = req.Reply(false, nil)
				}
			default:
				if req.WantReply {
					_ = req.Reply(false, nil)
				}
			}
		}
	}()

	select {
	case <-shellReady:
	case <-time.After(60 * time.Second):
		log.Printf("Remote SSH console: timeout waiting for shell request (client may use exec/subsystem only)")
		return
	}

	var outMu sync.Mutex
	out := &sshSyncedChannelWriter{w: connection, mu: &outMu}
	st := &sshDaemonConsoleState{out: out}
	st.rows.Store(termRows.Load())
	st.applyScrollLayout()
	_, _ = st.auxWriter().Write([]byte("TalKKonnect remote console (daemon). Type ? for menu, q to disconnect.\n\n"))
	st.redrawLine(nil)
	st.registerLog()
	sshDaemonConsoleActive.Store(st)
	defer func() {
		sshDaemonConsoleActive.Store(nil)
		st.unregisterLog()
	}()

	aux := st.auxWriter()
	for {
		line, err := bottomCLISSHReadLine(connection, st)
		if err != nil {
			if err != io.EOF {
				log.Printf("Remote SSH console: read error: %v", err)
			}
			return
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if b.bottomCLIDispatchRemoteLine(aux, line, st) {
			return
		}
	}
}

// sshSyncedChannelWriter serializes all writes to the SSH channel and converts LF
// to CRLF so PuTTY / Windows Terminal column alignment stays correct.
type sshSyncedChannelWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func sshBytesToCRLF(p []byte) []byte {
	var b strings.Builder
	b.Grow(len(p) + len(p)/16)
	for i := 0; i < len(p); i++ {
		if p[i] == '\n' && (i == 0 || p[i-1] != '\r') {
			b.WriteByte('\r')
		}
		b.WriteByte(p[i])
	}
	return []byte(b.String())
}

func (s *sshSyncedChannelWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.w.Write(sshBytesToCRLF(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// WithWireLock runs fn with the channel mutex held. Writes from fn should be raw
// wire bytes (use sshBytesToCRLF for text containing newlines).
func (s *sshSyncedChannelWriter) WithWireLock(fn func(w io.Writer)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn(s.w)
}
