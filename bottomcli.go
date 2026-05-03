/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Bottom-terminal interactive CLI: terminal layout and input loop are adapted from
 * github.com/talkkonnect/virtualkeyz2 (virtualkeyz2.go technician menu): DECSTBM
 * scrolling region, reserved bottom status row, synchronized log writes, raw /dev/tty
 * line editing with history and Tab completion.
 */

package talkkonnect

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/term"
)

const bottomCLIHistoryMax = 100

func bottomTerminalCLIShouldWrap() bool {
	if os.Getenv("TALKKONNECT_NO_BOTTOM_CLI") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

var (
	bottomCLIMu            sync.Mutex
	bottomCLIEnabled       bool
	bottomCLITerminalRows  int
	bottomCLIInputDraft    []byte
	bottomCLIHistoryMu     sync.Mutex
	bottomCLIHistory       []string
	bottomCLIPromptDefault = "talkkonnect"
)

type bottomCLIEchoWriter struct{}

func (bottomCLIEchoWriter) Write(p []byte) (int, error) {
	n := len(p)
	if n == 0 {
		return 0, nil
	}
	bottomCLISyncPrint(func(w io.Writer) {
		_, _ = w.Write(p)
	})
	return n, nil
}

func bottomCLIQueryRows() int {
	if _, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil && h >= 2 {
		return h
	}
	if s := os.Getenv("LINES"); s != "" {
		var n int
		_, _ = fmt.Sscanf(s, "%d", &n)
		if n >= 2 {
			return n
		}
	}
	return 24
}

func bottomCLIPromptLabel() string {
	if Config.Accounts.Account != nil && AccountIndex < len(Config.Accounts.Account) && AccountIndex >= 0 {
		if n := strings.TrimSpace(Config.Accounts.Account[AccountIndex].Name); n != "" {
			return n
		}
	}
	return bottomCLIPromptDefault
}

func bottomCLIMoveToScrollRegionBottomUnlocked(w io.Writer) {
	if !bottomCLIEnabled || bottomCLITerminalRows < 2 {
		return
	}
	_, _ = fmt.Fprintf(w, "\033[%d;1H", bottomCLITerminalRows-1)
}

func bottomCLIPaintPromptRowUnlocked(w io.Writer) {
	if !bottomCLIEnabled || bottomCLITerminalRows < 2 {
		return
	}
	_, _ = fmt.Fprintf(w, "\033[%d;1H\033[K", bottomCLITerminalRows)
	_, _ = fmt.Fprint(w, bottomCLIPromptLabel())
	_, _ = fmt.Fprint(w, "> ")
}

func bottomCLIPaintPromptAndDraftUnlocked(w io.Writer) {
	bottomCLIPaintPromptRowUnlocked(w)
	if len(bottomCLIInputDraft) > 0 {
		_, _ = w.Write(bottomCLIInputDraft)
	}
}

func bottomCLIEnableLayout() {
	rows := bottomCLIQueryRows()
	if rows < 2 {
		return
	}
	bottomCLIMu.Lock()
	bottomCLITerminalRows = rows
	bottomCLIEnabled = true
	_, _ = fmt.Fprintf(os.Stdout, "\033[1;%dr\033[1;1H", rows-1)
	bottomCLIPaintPromptAndDraftUnlocked(os.Stdout)
	bottomCLIMu.Unlock()
}

func bottomCLIDisableLayout() {
	bottomCLIMu.Lock()
	bottomCLIEnabled = false
	_, _ = fmt.Fprint(os.Stdout, "\033[r\n")
	bottomCLIMu.Unlock()
}

func bottomCLITerminalHardReset() {
	const seq = "\033[0m\033[?25h\033[r\033c"
	_, _ = fmt.Fprint(os.Stdout, seq)
	if t, err := os.OpenFile("/dev/tty", os.O_WRONLY, 0); err == nil {
		_, _ = fmt.Fprint(t, seq)
		_ = t.Close()
	}
}

func bottomCLIClearScreenAndRelayout() {
	rows := bottomCLIQueryRows()
	if rows < 2 {
		bottomCLIMu.Lock()
		rows = bottomCLITerminalRows
		bottomCLIMu.Unlock()
	}
	if rows < 2 {
		rows = 24
	}
	bottomCLIMu.Lock()
	defer bottomCLIMu.Unlock()
	bottomCLITerminalRows = rows
	bottomCLIEnabled = true
	_, _ = fmt.Fprint(os.Stdout, "\033[2J\033[1;1H")
	_, _ = fmt.Fprintf(os.Stdout, "\033[1;%dr\033[1;1H", rows-1)
	bottomCLIPaintPromptAndDraftUnlocked(os.Stdout)
}

func bottomCLISyncPrint(f func(w io.Writer)) {
	bottomCLIMu.Lock()
	defer bottomCLIMu.Unlock()
	bottomCLIMoveToScrollRegionBottomUnlocked(os.Stdout)
	f(os.Stdout)
	bottomCLIPaintPromptAndDraftUnlocked(os.Stdout)
}

type bottomCLILogWriter struct {
	down io.Writer
	buf  []byte
}

func newBottomCLILogWriter(downstream io.Writer) io.Writer {
	return &bottomCLILogWriter{down: downstream}
}

func (c *bottomCLILogWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	c.buf = append(c.buf, p...)
	for {
		idx := bytes.IndexByte(c.buf, '\n')
		if idx < 0 {
			return n, nil
		}
		line := c.buf[:idx+1]
		c.buf = append([]byte(nil), c.buf[idx+1:]...)
		bottomCLIMu.Lock()
		bottomCLIMoveToScrollRegionBottomUnlocked(c.down)
		_, _ = c.down.Write(line)
		bottomCLIPaintPromptAndDraftUnlocked(c.down)
		bottomCLIMu.Unlock()
	}
}

func bottomCLIAppendHistory(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	bottomCLIHistoryMu.Lock()
	defer bottomCLIHistoryMu.Unlock()
	bottomCLIHistory = append(bottomCLIHistory, line)
	if len(bottomCLIHistory) > bottomCLIHistoryMax {
		bottomCLIHistory = bottomCLIHistory[len(bottomCLIHistory)-bottomCLIHistoryMax:]
	}
}

func bottomCLIHistorySnapshot() []string {
	bottomCLIHistoryMu.Lock()
	defer bottomCLIHistoryMu.Unlock()
	return append([]string(nil), bottomCLIHistory...)
}

func bottomCLICompletionCandidates() []string {
	seen := make(map[string]struct{})
	var out []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		low := strings.ToLower(s)
		if _, ok := seen[low]; ok {
			return
		}
		seen[low] = struct{}{}
		out = append(out, s)
	}
	for _, c := range Config.Global.Software.RemoteControl.HTTP.Command {
		if c.Enabled && c.Action != "" {
			add(c.Action)
		}
	}
	for _, s := range []string{
		"help", "?", "menu", "cfg", "clearhist", "c", "clear", "cls", "q", "quit", "exit", "...", "…",
		"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
		"a", "b", "d", "e", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "z",
	} {
		add(s)
	}
	return out
}

func bottomCLIFilterPrefixLower(cands []string, lowPrefix string) []string {
	var m []string
	for _, s := range cands {
		if strings.HasPrefix(strings.ToLower(s), lowPrefix) {
			m = append(m, s)
		}
	}
	return m
}

func bottomCLILongestCommonPrefix(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	ref := strings.ToLower(strs[0])
	for i := 0; i < len(ref); i++ {
		ch := ref[i]
		for j := 1; j < len(strs); j++ {
			s := strings.ToLower(strs[j])
			if i >= len(s) || s[i] != ch {
				return strs[0][:i]
			}
		}
	}
	return strs[0]
}

func bottomCLITabCompleteLine(line string) (newLine string, bell bool) {
	line = strings.TrimRight(line, "\n\r")
	if strings.HasPrefix(strings.ToLower(strings.TrimLeft(line, " \t")), "cfg") {
		return bottomCLITabCompleteCfg(line)
	}
	fields := strings.Fields(line)
	if len(fields) > 1 {
		return line, true
	}
	partial := ""
	if len(fields) == 1 {
		partial = fields[0]
	}
	matches := bottomCLIFilterPrefixLower(bottomCLICompletionCandidates(), strings.ToLower(partial))
	if len(matches) == 0 {
		return line, true
	}
	if len(matches) == 1 {
		return matches[0] + " ", false
	}
	lcp := bottomCLILongestCommonPrefix(matches)
	lowPart := strings.ToLower(partial)
	if !strings.HasPrefix(strings.ToLower(lcp), lowPart) || len(strings.ToLower(lcp)) == len(lowPart) {
		return line, true
	}
	return lcp, false
}

func bottomCLIReadCSI(tty *os.File) ([]byte, error) {
	b := make([]byte, 1)
	if _, err := tty.Read(b); err != nil {
		return nil, err
	}
	if b[0] != '[' && b[0] != 'O' {
		return []byte{b[0]}, nil
	}
	out := []byte{b[0]}
	for {
		if _, err := tty.Read(b); err != nil {
			return out, err
		}
		out = append(out, b[0])
		if b[0] >= 0x40 && b[0] <= 0x7e {
			break
		}
	}
	return out, nil
}

func bottomCLIRedrawInputLine(line []byte) {
	bottomCLIMu.Lock()
	defer bottomCLIMu.Unlock()
	bottomCLIInputDraft = append([]byte(nil), line...)
	bottomCLIPaintPromptRowUnlocked(os.Stdout)
	if len(line) > 0 {
		_, _ = os.Stdout.Write(line)
	}
}

func bottomCLIReadLine(tty *os.File) (string, error) {
	fd := int(tty.Fd())
	old, err := term.MakeRaw(fd)
	if err != nil {
		r := bufio.NewReader(tty)
		s, e := r.ReadString('\n')
		if e != nil {
			return "", e
		}
		s = strings.TrimSuffix(s, "\r")
		s = strings.TrimSuffix(s, "\n")
		return s, nil
	}
	defer func() {
		_ = term.Restore(fd, old)
		bottomCLIMu.Lock()
		bottomCLIInputDraft = nil
		bottomCLIMu.Unlock()
	}()

	var line []byte
	histIdx := -1
	redraw := func() { bottomCLIRedrawInputLine(line) }
	redraw()

	upSeq := []byte("\x1b[A")
	downSeq := []byte("\x1b[B")
	upSS3 := []byte("\x1bOA")
	downSS3 := []byte("\x1bOB")

	buf := make([]byte, 1)
	for {
		n, err := tty.Read(buf)
		if err != nil {
			return "", err
		}
		if n == 0 {
			continue
		}
		b := buf[0]
		switch {
		case b == '\r' || b == '\n':
			bottomCLIMu.Lock()
			_, _ = fmt.Fprint(os.Stdout, "\n")
			bottomCLIMu.Unlock()
			return string(line), nil
		case b == 127 || b == 8:
			if len(line) > 0 {
				line = line[:len(line)-1]
				histIdx = -1
				redraw()
			}
		case b == '\t':
			histIdx = -1
			nl, bell := bottomCLITabCompleteLine(string(line))
			line = []byte(nl)
			if bell {
				_, _ = tty.Write([]byte{'\a'})
			}
			redraw()
		case b == 27:
			csi, err := bottomCLIReadCSI(tty)
			if err != nil {
				return "", err
			}
			seq := append([]byte{27}, csi...)
			hist := bottomCLIHistorySnapshot()
			nh := len(hist)
			switch {
			case bytes.Equal(seq, upSeq) || bytes.Equal(seq, upSS3):
				if nh == 0 {
					redraw()
					continue
				}
				if histIdx < 0 {
					histIdx = nh - 1
				} else if histIdx > 0 {
					histIdx--
				}
				line = append([]byte(nil), hist[histIdx]...)
				redraw()
			case bytes.Equal(seq, downSeq) || bytes.Equal(seq, downSS3):
				if histIdx < 0 {
					continue
				}
				if histIdx < nh-1 {
					histIdx++
					line = append([]byte(nil), hist[histIdx]...)
				} else {
					histIdx = -1
					line = nil
				}
				redraw()
			}
		case b >= 32 && b < 127:
			histIdx = -1
			line = append(line, b)
			redraw()
		case b == 3:
			line = nil
			histIdx = -1
			redraw()
		}
	}
}

func bottomCLIPartsToRemoteQuery(parts []string) remoteAPIQuery {
	q := remoteAPIQuery{}
	if len(parts) == 0 {
		return q
	}
	q.Command = strings.ToLower(strings.TrimSpace(parts[0]))
	if q.Command == "voicetargetset" && len(parts) >= 2 {
		q.ID, _ = strconv.Atoi(parts[1])
	}
	if q.Command == "ttsannouncement" && len(parts) >= 2 {
		q.APITTSMessage = strings.Join(parts[1:], " ")
	}
	return q
}

// bottomCLIMenuBanner mirrors talkkonnectMenu() in talkkonnect.go (LCD key help).
const bottomCLIMenuBanner = `
------------------------------------------------------------------------------------
 talkkonnect Mumble SBC Client <suvir@talkkonnect.com>
------------------------------------------------------------------------------------	
  ?  Display this menu       2  Channel UP (+)           3  Channel Down (-)
  4  Mute/Unmute speaker     5  Digital volume up (+)    6  Digital volume down (-)
  7  Start transmitting      8  Stop transmitting        9  List online users
  0  Show uptime
------------------------------------------------------------------------------------
  a  List API commands (log) b  Playback/stop stream      d  Dump XML config
  e  Send email              g  GPS position              h  XML config checker (sanity)
  i  Traffic record          j  Mic record                k  Traffic & mic record
  l  Clear screen (LCD/OLED) m  Radio channel (+)         n  Radio channel (-)
  o  Ping servers            p  Panic simulation          q  Repeat TX loop test
  r  Scan channels           s  Thanks/acknowledge        t  Show uptime
  u  Display version         v  Online radio on/off       w  Dump XML config
  x  Previous server         z  Next server                    
------------------------------------------------------------------------------------
 CLI Commands:
  menu / ? / help           Show this banner
  cfg keys|list|set|save|restart   Inspect or change config (Tab completes cfg set paths)
  c / clear / cls           Clear terminal + restore bottom prompt
  q / quit / exit           Close bottom CLI (talkkonnect keeps running)
  ... or …                  Quit talkkonnect (SIGTERM)
------------------------------------------------------------------------------------
Visit us at www.talkkonnect.com and github.com/talkkonnect
Thanks to Global Coders Co., Ltd. for their sponsorship 	
------------------------------------------------------------------------------------
`

// bottomCLIExecuteQuickMenu runs single-key shortcuts matching talkkonnectMenu() in talkkonnect.go.
// Returns true if the line was handled and should not be passed to HandleRemoteAPICommand.
// If auxOut is non-nil (e.g. SSH daemon console), short user messages go there instead of the bottom-terminal sync writer.
func (b *Talkkonnect) bottomCLIExecuteQuickMenu(key string, auxOut io.Writer) bool {
	switch key {
	case "2":
		b.cmdChannelUp()
		log.Println("info: menu: channel up")
	case "3":
		b.cmdChannelDown()
		log.Println("info: menu: channel down")
	case "4":
		b.cmdMuteUnmute("toggle")
		log.Println("info: menu: mute/unmute")
	case "5":
		b.cmdVolumeRXUp()
		log.Println("info: menu: digital volume up")
	case "6":
		b.cmdVolumeRXDown()
		log.Println("info: menu: digital volume down")
	case "7":
		b.cmdStartTransmitting()
		log.Println("info: menu: start transmitting")
	case "8":
		b.cmdStopTransmitting()
		log.Println("info: menu: stop transmitting")
	case "9":
		b.cmdListOnlineUsers()
		log.Println("info: menu: list online users")
	case "0":
		b.cmdShowUptime()
		log.Println("info: menu: show uptime")
	case "a":
		listAPI()
		if auxOut != nil {
			fmt.Fprintln(auxOut, "API command list written to log (see talkkonnect log).")
		} else {
			bottomCLISyncPrint(func(w io.Writer) {
				fmt.Fprintln(w, "API command list written to log (see talkkonnect log / screen above).")
			})
		}
		log.Println("info: menu: list API commands")
	case "b":
		b.cmdPlayback()
		log.Println("info: menu: playback/stop stream")
	case "d":
		b.cmdDumpXMLConfig()
		log.Println("info: menu: dump XML config")
	case "e":
		b.cmdSendEmail()
		log.Println("info: menu: send email")
	case "g":
		b.cmdGPSPosition()
		log.Println("info: menu: GPS position")
	case "h":
		cmdSanityCheck()
		log.Println("info: menu: XML config checker (sanity)")
	case "i":
		b.cmdAudioTrafficRecord()
		log.Println("info: menu: traffic record")
	case "j":
		b.cmdAudioMicRecord()
		log.Println("info: menu: mic record")
	case "k":
		b.cmdAudioMicTrafficRecord()
		log.Println("info: menu: traffic & mic record")
	case "l":
		b.cmdClearScreen()
		log.Println("info: menu: clear screen")
	case "m":
		b.cmdRadioChannelMove("Up")
		log.Println("info: menu: radio channel (+)")
	case "n":
		b.cmdRadioChannelMove("Down")
		log.Println("info: menu: radio channel (-)")
	case "o":
		b.cmdPingServers()
		log.Println("info: menu: ping servers")
	case "p":
		b.cmdPanicSimulation()
		log.Println("info: menu: panic simulation")
	case "q":
		b.cmdRepeatTxLoop()
		log.Println("info: menu: repeat TX loop test")
	case "r":
		b.cmdScanChannels()
		log.Println("info: menu: scan channels")
	case "s":
		cmdThanks()
		log.Println("info: menu: thanks/acknowledge")
	case "t":
		b.cmdShowUptime()
		log.Println("info: menu: show uptime")
	case "u":
		b.cmdDisplayVersion()
		log.Println("info: menu: display version")
	case "v":
		if len(player.stations) == 0 {
			log.Println("warn: menu: online radio has no stations")
		} else if player.current_station >= 0 {
			player.Stop()
			log.Println("info: menu: online radio stop")
		} else {
			player.Play(0)
			log.Println("info: menu: online radio start (first station)")
		}
	case "w":
		b.cmdDumpXMLConfig()
		log.Println("info: menu: dump XML config (w)")
	case "x":
		b.cmdConnPreviousServer()
		log.Println("info: menu: previous server")
	case "z":
		b.cmdConnNextServer()
		log.Println("info: menu: next server")
	default:
		return false
	}
	return true
}

// bottomCLIDispatchRemoteLine runs one user line for the SSH daemon console (no bottom-terminal layout).
func (b *Talkkonnect) bottomCLIDispatchRemoteLine(w io.Writer, line string) (disconnectSession bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return false
	}
	if w != nil {
		sshRemoteReplyAttach(w)
		defer sshRemoteReplyDetach()
	}
	bottomCLIAppendHistory(line)
	key := strings.ToLower(line)
	switch key {
	case "c", "cls", "clear":
		_, _ = fmt.Fprint(w, "\r\n\x1b[2J\x1b[H")
		log.Println("info: Remote SSH console: screen cleared.")
	case "q", "quit", "exit":
		_, _ = fmt.Fprintln(w, "Disconnected.")
		log.Println("info: Remote SSH console session closed by user (q).")
		return true
	case "...", "…":
		_, _ = fmt.Fprintln(w, "Shutdown requested; sending SIGTERM.")
		log.Println("info: Shutdown requested from remote SSH console (...); sending SIGTERM.")
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		return true
	case "?", "help", "menu":
		_, _ = fmt.Fprint(w, bottomCLIMenuBanner)
	case "clearhist":
		bottomCLIHistoryMu.Lock()
		bottomCLIHistory = nil
		bottomCLIHistoryMu.Unlock()
		_, _ = fmt.Fprintln(w, "Command history cleared.")
		log.Println("info: Remote SSH console: command history cleared (clearhist).")
	default:
		if b.bottomCLIExecuteQuickMenu(key, w) {
			return false
		}
		parts := strings.Fields(line)
		if len(parts) >= 1 && strings.EqualFold(parts[0], "cfg") {
			bottomCLIHandleCfgLine(w, line)
			return false
		}
		q := bottomCLIPartsToRemoteQuery(parts)
		b.HandleRemoteAPICommand(w, q)
	}
	return false
}

func (b *Talkkonnect) runBottomTerminalCLI() {
	// Brief yield so the first colog lines flush; layout runs before later Init/ClientStart logs.
	time.Sleep(50 * time.Millisecond)

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		log.Printf("info: Bottom CLI skipped (no /dev/tty: %v). Set TALKKONNECT_NO_BOTTOM_CLI=1 to silence tries.", err)
		return
	}
	defer tty.Close()

	bottomCLIEnableLayout()

	bottomCLISyncPrint(func(w io.Writer) {
		//fmt.Fprintln(w, "Type h for main menu (quick keys 0-9, a, l, p, …) or type an HTTP API command.")
	})

	for {
		bottomCLIMu.Lock()
		bottomCLIPaintPromptAndDraftUnlocked(os.Stdout)
		bottomCLIMu.Unlock()

		line, err := bottomCLIReadLine(tty)
		if err != nil {
			bottomCLIDisableLayout()
			if err == io.EOF {
				return
			}
			log.Printf("info: Bottom CLI read ended: %v", err)
			return
		}
		line = strings.TrimSpace(line)
		bottomCLIAppendHistory(line)
		if line == "" {
			continue
		}

		key := strings.ToLower(line)
		if key != "..." && line != "…" && key != "c" && key != "cls" && key != "clear" {
			bottomCLIMu.Lock()
			if bottomCLIEnabled && bottomCLITerminalRows >= 2 {
				_, _ = fmt.Fprintf(os.Stdout, "\033[%d;1H\n", bottomCLITerminalRows-1)
			}
			bottomCLIMu.Unlock()
		}

		switch key {
		case "c", "cls", "clear":
			bottomCLIClearScreenAndRelayout()
			log.Println("info: Bottom CLI: screen cleared.")
		case "q", "quit", "exit":
			bottomCLISyncPrint(func(w io.Writer) { fmt.Fprintln(w, "Bottom CLI closed (talkkonnect continues).") })
			log.Println("info: Bottom CLI closed by user (q).")
			bottomCLIDisableLayout()
			return
		case "...", "…":
			bottomCLIDisableLayout()
			bottomCLITerminalHardReset()
			log.Println("info: Shutdown requested from bottom CLI (...); sending SIGTERM.")
			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
			return
		case "?", "help", "menu":
			bottomCLISyncPrint(func(w io.Writer) { fmt.Fprint(w, bottomCLIMenuBanner) })
		case "clearhist":
			bottomCLIHistoryMu.Lock()
			bottomCLIHistory = nil
			bottomCLIHistoryMu.Unlock()
			bottomCLISyncPrint(func(w io.Writer) { fmt.Fprintln(w, "Command history cleared.") })
			log.Println("info: Bottom CLI: command history cleared (clearhist).")
		default:
			if b.bottomCLIExecuteQuickMenu(key, nil) {
				break
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 && strings.EqualFold(parts[0], "cfg") {
				bottomCLISyncPrint(func(out io.Writer) {
					bottomCLIHandleCfgLine(out, line)
				})
				break
			}
			q := bottomCLIPartsToRemoteQuery(parts)
			b.HandleRemoteAPICommand(bottomCLIEchoWriter{}, q)
		}
	}
}
