/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Full-line ANSI coloring for prefix-level log output by severity.
 */

package talkkonnect

import (
	"bytes"
	"os"

	"golang.org/x/term"
)

const (
	ansiReset   = "\033[0m"
	ansiTrace   = "\033[0;90m"    // bright black / gray
	ansiDebug   = "\033[0;36m"    // cyan
	ansiInfo    = "\033[0;32m"    // green
	ansiWarning = "\033[0;33m"    // yellow
	ansiError   = "\033[1;31m"    // bold red
	ansiAlert   = "\033[1;37;41m" // bold white on red
)

// fullLineColorFormatter wraps each log line in one ANSI foreground (and alert: background)
// so the entire line is colored by level. Respects NO_COLOR; enables when the logger reports a TTY
// or when stdout is a terminal (covers io.MultiWriter and bottom CLI wrappers).
type fullLineColorFormatter struct {
	plain     *plainLogFormatter
	outIsTerm bool
}

func newFullLineColorFormatter() *fullLineColorFormatter {
	return &fullLineColorFormatter{
		plain: &plainLogFormatter{},
	}
}

func (f *fullLineColorFormatter) Flags() int {
	return f.plain.Flags()
}

func (f *fullLineColorFormatter) SetFlags(flags int) {
	f.plain.SetFlags(flags)
}

func (f *fullLineColorFormatter) ColorSupported(yes bool) {
	f.outIsTerm = yes
}

func levelLinePrefix(level prefixLogLevel) string {
	switch level {
	case prefixLevelTrace:
		return ansiTrace
	case prefixLevelDebug:
		return ansiDebug
	case prefixLevelInfo:
		return ansiInfo
	case prefixLevelWarning:
		return ansiWarning
	case prefixLevelError:
		return ansiError
	case prefixLevelAlert:
		return ansiAlert
	default:
		return ansiTrace
	}
}

func (f *fullLineColorFormatter) Format(e *prefixLogEntry) ([]byte, error) {
	out, err := f.plain.Format(e)
	if err != nil {
		return out, err
	}
	if os.Getenv("NO_COLOR") != "" {
		return out, nil
	}
	useColor := f.outIsTerm || term.IsTerminal(int(os.Stdout.Fd()))
	if !useColor {
		return out, nil
	}
	out = bytes.TrimSuffix(out, []byte("\n"))
	prefix := levelLinePrefix(e.Level)
	var buf bytes.Buffer
	buf.Grow(len(prefix) + len(out) + len(ansiReset) + 1)
	buf.WriteString(prefix)
	buf.Write(out)
	buf.WriteString(ansiReset)
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// sshPlainBracketPrefixes must stay aligned with plainLevelLabels in prefix_level_log.go
// (plainLogFormatter output). Used only for SSH log mirroring so files stay unescaped.
var sshPlainBracketPrefixes = []struct {
	prefix []byte
	level  prefixLogLevel
}{
	{[]byte("[ alert ] "), prefixLevelAlert},
	{[]byte("[ error ] "), prefixLevelError},
	{[]byte("[  warn ] "), prefixLevelWarning},
	{[]byte("[  info ] "), prefixLevelInfo},
	{[]byte("[ debug ] "), prefixLevelDebug},
	{[]byte("[ trace ] "), prefixLevelTrace},
}

// colorizePlainPrefixLogLineForSSH wraps one plain formatter line in full-line ANSI by level.
// If the line is already ANSI (e.g. local bottom CLI mirroring) or has no bracket level tag,
// returns a copy of line unchanged. Respects NO_COLOR.
func colorizePlainPrefixLogLineForSSH(line []byte) []byte {
	if len(line) == 0 {
		return line
	}
	out := append([]byte(nil), line...)
	if os.Getenv("NO_COLOR") != "" {
		return out
	}
	if line[0] == '\x1b' {
		return out
	}
	body := bytes.TrimSuffix(line, []byte("\n"))
	var lvl prefixLogLevel
	ok := false
	for _, ent := range sshPlainBracketPrefixes {
		if bytes.HasPrefix(body, ent.prefix) {
			lvl = ent.level
			ok = true
			break
		}
	}
	if !ok {
		return out
	}
	prefix := levelLinePrefix(lvl)
	var buf bytes.Buffer
	buf.Grow(len(prefix) + len(body) + len(ansiReset) + 1)
	buf.WriteString(prefix)
	buf.Write(body)
	buf.WriteString(ansiReset)
	buf.WriteByte('\n')
	return buf.Bytes()
}
