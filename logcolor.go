/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Full-line ANSI coloring for colog output by severity level.
 */

package talkkonnect

import (
	"bytes"
	"os"

	"github.com/talkkonnect/colog"
	"golang.org/x/term"
)

const (
	ansiReset   = "\033[0m"
	ansiTrace   = "\033[0;90m"  // bright black / gray
	ansiDebug   = "\033[0;36m"  // cyan
	ansiInfo    = "\033[0;32m"  // green
	ansiWarning = "\033[0;33m"  // yellow
	ansiError   = "\033[1;31m"  // bold red
	ansiAlert   = "\033[1;37;41m" // bold white on red
)

// fullLineColorFormatter wraps each log line in one ANSI foreground (and alert: background)
// so the entire line is colored by level. Respects NO_COLOR; enables when colog reports a TTY
// or when stdout is a terminal (covers io.MultiWriter and bottom CLI wrappers).
type fullLineColorFormatter struct {
	plain     *colog.StdFormatter
	outIsTerm bool
}

func newFullLineColorFormatter() *fullLineColorFormatter {
	return &fullLineColorFormatter{
		plain: &colog.StdFormatter{NoColors: true},
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
	// Inner formatter must stay plain; we apply color to the whole line ourselves.
	f.plain.ColorSupported(false)
	f.plain.NoColors = true
}

func levelLinePrefix(level colog.Level) string {
	switch level {
	case colog.LTrace:
		return ansiTrace
	case colog.LDebug:
		return ansiDebug
	case colog.LInfo:
		return ansiInfo
	case colog.LWarning:
		return ansiWarning
	case colog.LError:
		return ansiError
	case colog.LAlert:
		return ansiAlert
	default:
		return ansiTrace
	}
}

func (f *fullLineColorFormatter) Format(e *colog.Entry) ([]byte, error) {
	f.plain.NoColors = true
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
