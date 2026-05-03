/*
 * Remote SSH console when talkkonnect runs as a go-daemon child (no GNU screen).
 * github.com/talkkonnect/gosshd attaches with "screen -xS tk", which fails without screen.
 */

package talkkonnect

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
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

	// Do not write to the session channel until the client has received a successful
	// shell reply; PuTTY and Windows OpenSSH otherwise keep input buffered or disabled.
	shellReady := make(chan struct{}, 1)
	go func() {
		for req := range requests {
			switch req.Type {
			case "env":
				// PuTTY often sends env requests; rejecting them can confuse the client.
				if req.WantReply {
					_ = req.Reply(true, nil)
				}
			case "pty-req":
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
				// no local PTY to resize
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
	r := bufio.NewReader(connection)
	lineIn := &sshLineInputState{}

	fmt.Fprintf(out, "\r\nTalKKonnect remote console (daemon). Type ? for menu, q to disconnect.\r\n\r\n")
	for {
		fmt.Fprintf(out, "%s> ", bottomCLIPromptLabel())
		line, err := sshReadLine(r, out, lineIn)
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
		disconnect := b.bottomCLIDispatchRemoteLine(out, line)
		if disconnect {
			return
		}
		// Ensure the next prompt starts on a new row (API + quick-menu paths often lack a trailing newline).
		_, _ = fmt.Fprintf(out, "\r\n")
	}
}

// sshSyncedChannelWriter serializes all writes to the SSH channel and converts LF
// to CRLF so PuTTY / Windows Terminal column alignment stays correct.
type sshSyncedChannelWriter struct {
	w  io.Writer
	mu *sync.Mutex
}

func (s *sshSyncedChannelWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	var b strings.Builder
	for i := 0; i < len(p); i++ {
		if p[i] == '\n' && (i == 0 || p[i-1] != '\r') {
			b.WriteByte('\r')
		}
		b.WriteByte(p[i])
	}
	_, err := s.w.Write([]byte(b.String()))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// sshLineInputState drops a single LF that follows a CR line terminator (CRLF clients).
type sshLineInputState struct {
	dropNextLF bool
}

const sshMaxLineBytes = 4096

// sshReadLine reads until CR or LF with local editing. Does not use Peek after CR
// (that blocked until a second key and broke PuTTY / Windows Terminal).
func sshReadLine(r *bufio.Reader, out *sshSyncedChannelWriter, st *sshLineInputState) (string, error) {
	var line []byte
	writeEcho := func(p []byte) {
		if len(p) > 0 {
			_, _ = out.Write(p)
		}
	}

	if st.dropNextLF {
		st.dropNextLF = false
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				return "", io.EOF
			}
			return "", err
		}
		if b != '\n' {
			if uerr := r.UnreadByte(); uerr != nil {
				return "", uerr
			}
		}
	}

	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF && len(line) > 0 {
				return string(line), nil
			}
			return "", err
		}
		switch b {
		case '\n':
			return string(line), nil
		case '\r':
			st.dropNextLF = true
			return string(line), nil
		case 127, 8: // DEL, BS
			if len(line) > 0 {
				line = line[:len(line)-1]
				writeEcho([]byte{'\b', ' ', '\b'})
			}
		case 21: // ^U kill line
			for len(line) > 0 {
				line = line[:len(line)-1]
				writeEcho([]byte{'\b', ' ', '\b'})
			}
		case 3: // ^C discard current line
			line = line[:0]
			writeEcho([]byte("^C\r\n"))
		default:
			if b >= 32 || b == '\t' {
				if len(line) >= sshMaxLineBytes {
					return string(line), fmt.Errorf("line too long")
				}
				line = append(line, b)
				writeEcho([]byte{b})
			}
		}
	}
}
