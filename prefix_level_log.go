/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Internal replacement for github.com/talkkonnect/colog: prefix-based severity,
 * minimum level filtering, and integration with the standard library log package.
 *
 * Level header parsing and layout follow the same rules as colog (MPL 2.0).
 */

package talkkonnect

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"golang.org/x/term"
)

// Severity levels (same semantics as former colog.Level).
type prefixLogLevel uint8

const (
	prefixLevelUnknown prefixLogLevel = iota
	prefixLevelTrace
	prefixLevelDebug
	prefixLevelInfo
	prefixLevelWarning
	prefixLevelError
	prefixLevelAlert
)

func (l prefixLogLevel) String() string {
	switch l {
	case prefixLevelTrace:
		return "trace"
	case prefixLevelDebug:
		return "debug"
	case prefixLevelInfo:
		return "info"
	case prefixLevelWarning:
		return "warning"
	case prefixLevelError:
		return "error"
	case prefixLevelAlert:
		return "alert"
	default:
		return "unknown"
	}
}

// prefixLogEntry is one line after the standard logger hands it to us.
type prefixLogEntry struct {
	Level   prefixLogLevel
	Time    time.Time
	Message []byte
	File    string
	Line    int
}

type prefixHeader struct {
	header string
	level  prefixLogLevel
}

var defaultPrefixHeaders = []prefixHeader{
	{"panic: ", prefixLevelAlert},
	{"alert: ", prefixLevelAlert},
	{"alr: ", prefixLevelAlert},
	{"a: ", prefixLevelAlert},
	{"error: ", prefixLevelError},
	{"err: ", prefixLevelError},
	{"e: ", prefixLevelError},
	{"warning: ", prefixLevelWarning},
	{"warn: ", prefixLevelWarning},
	{"wrn: ", prefixLevelWarning},
	{"w: ", prefixLevelWarning},
	{"info: ", prefixLevelInfo},
	{"inf: ", prefixLevelInfo},
	{"i: ", prefixLevelInfo},
	{"debug: ", prefixLevelDebug},
	{"dbg: ", prefixLevelDebug},
	{"d: ", prefixLevelDebug},
	{"trace: ", prefixLevelTrace},
	{"trc: ", prefixLevelTrace},
	{"t: ", prefixLevelTrace},
}

func init() {
	// Longest headers first so e.g. "info: " wins over "i: ".
	sort.Slice(defaultPrefixHeaders, func(i, j int) bool {
		return len(defaultPrefixHeaders[i].header) > len(defaultPrefixHeaders[j].header)
	})
}

type prefixLevelLogger struct {
	mu           sync.Mutex
	out          io.Writer
	minLevel     prefixLogLevel
	defaultLevel prefixLogLevel
	headers      []prefixHeader
	formatter    *fullLineColorFormatter
	customFmt    bool
}

var stdPrefixLogger = &prefixLevelLogger{
	minLevel:     prefixLevelTrace,
	defaultLevel: prefixLevelInfo,
	headers:      defaultPrefixHeaders,
	formatter:    newFullLineColorFormatter(),
	out:          os.Stderr,
}

// prefixLogRegister wires the standard library logger through stdPrefixLogger (same idea as colog.Register).
func prefixLogRegister() {
	stdPrefixLogger.mu.Lock()
	if !stdPrefixLogger.customFmt && stdPrefixLogger.formatter != nil {
		stdPrefixLogger.formatter.SetFlags(log.Flags())
	}
	stdPrefixLogger.mu.Unlock()

	log.SetPrefix("")
	log.SetFlags(0)
	log.SetOutput(stdPrefixLogger)
	prefixLogNotifyColorSupport()
}

func prefixLogSetFormatter(f *fullLineColorFormatter) {
	stdPrefixLogger.mu.Lock()
	stdPrefixLogger.customFmt = true
	stdPrefixLogger.formatter = f
	stdPrefixLogger.mu.Unlock()
	prefixLogNotifyColorSupport()
}

func prefixLogSetOutput(w io.Writer) {
	stdPrefixLogger.mu.Lock()
	stdPrefixLogger.out = w
	stdPrefixLogger.mu.Unlock()
	prefixLogNotifyColorSupport()
}

func prefixLogSetFlags(flags int) {
	stdPrefixLogger.mu.Lock()
	if stdPrefixLogger.formatter != nil {
		stdPrefixLogger.formatter.SetFlags(flags)
	}
	stdPrefixLogger.mu.Unlock()
}

func prefixLogSetMinLevel(l prefixLogLevel) {
	stdPrefixLogger.mu.Lock()
	stdPrefixLogger.minLevel = l
	stdPrefixLogger.mu.Unlock()
}

func prefixLogNotifyColorSupport() {
	stdPrefixLogger.mu.Lock()
	out := stdPrefixLogger.out
	f := stdPrefixLogger.formatter
	stdPrefixLogger.mu.Unlock()
	if f == nil {
		return
	}
	f.ColorSupported(prefixLogOutputSupportsColor(out))
}

func prefixLogOutputSupportsColor(out io.Writer) bool {
	if out == nil {
		return false
	}
	if cs, ok := out.(interface{ ColorSupported() bool }); ok {
		return cs.ColorSupported()
	}
	if runtime.GOOS == "windows" {
		return false
	}
	type fdGetter interface{ Fd() uintptr }
	og, ok := out.(fdGetter)
	if !ok {
		return false
	}
	return term.IsTerminal(int(og.Fd()))
}

func applyPrefixHeaders(headers []prefixHeader, defaultLevel prefixLogLevel, msg []byte) (prefixLogLevel, []byte) {
	for _, h := range headers {
		p := []byte(h.header)
		if bytes.HasPrefix(msg, p) {
			return h.level, bytes.TrimPrefix(msg, p)
		}
	}
	return defaultLevel, msg
}

func (cl *prefixLevelLogger) Write(p []byte) (n int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("prefix logger recovered panic: %v", r)
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
		}
	}()

	cl.mu.Lock()
	outWriter := cl.out
	formatter := cl.formatter
	minLevel := cl.minLevel
	defaultLevel := cl.defaultLevel
	headers := cl.headers
	cl.mu.Unlock()

	if outWriter == nil || formatter == nil {
		e := errors.New("prefix logger: missing output or formatter")
		_, _ = fmt.Fprintln(os.Stderr, e.Error())
		return 0, e
	}

	msg := bytes.TrimRight(p, "\n")
	level, body := applyPrefixHeaders(headers, defaultLevel, msg)

	e := &prefixLogEntry{
		Level:   level,
		Time:    time.Now(),
		Message: body,
	}

	if formatter.Flags()&(log.Lshortfile|log.Llongfile) != 0 {
		e.File, e.Line = prefixLogGetFileLine(5)
	}

	if level < minLevel {
		return len(p), nil
	}

	fp, err := formatter.Format(e)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "prefix logger: format: %v\n", err)
		return 0, err
	}
	_, err = outWriter.Write(fp)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func prefixLogGetFileLine(calldepth int) (string, int) {
	_, file, line, ok := runtime.Caller(calldepth)
	if !ok {
		return "???", 0
	}
	return file, line
}

// plainLogFormatter renders date/time, optional file:line, and message (no ANSI in the body).
type plainLogFormatter struct {
	mu   sync.Mutex
	flag int
}

func (sf *plainLogFormatter) Flags() int {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	return sf.flag
}

func (sf *plainLogFormatter) SetFlags(flags int) {
	sf.mu.Lock()
	defer sf.mu.Unlock()
	sf.flag = flags
}

var plainLevelLabels = map[prefixLogLevel][]byte{
	prefixLevelTrace:   []byte("[ trace ] "),
	prefixLevelDebug:   []byte("[ debug ] "),
	prefixLevelInfo:    []byte("[  info ] "),
	prefixLevelWarning: []byte("[  warn ] "),
	prefixLevelError:   []byte("[ error ] "),
	prefixLevelAlert:   []byte("[ alert ] "),
}

func (sf *plainLogFormatter) Format(e *prefixLogEntry) ([]byte, error) {
	sf.mu.Lock()
	flags := sf.flag
	sf.mu.Unlock()

	lbl := plainLevelLabels[e.Level]
	if lbl == nil {
		lbl = plainLevelLabels[prefixLevelTrace]
	}

	var header []byte
	prefixLogStdHeader(&header, flags, e.Time, "", e.File, e.Line)

	var message []byte
	message = append(message, lbl...)
	message = append(message, header...)
	message = append(message, e.Message...)
	return append(message, '\n'), nil
}

// prefixLogStdHeader mirrors colog.StdFormatter.stdHeader (log-style date/time and file line).
func prefixLogStdHeader(buf *[]byte, flags int, t time.Time, prefix, file string, line int) {
	*buf = append(*buf, prefix...)
	if flags&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		if flags&log.Ldate != 0 {
			year, month, day := t.Date()
			prefixLogItoa(buf, year, 4)
			*buf = append(*buf, '/')
			prefixLogItoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			prefixLogItoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if flags&(log.Ltime|log.Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			prefixLogItoa(buf, hour, 2)
			*buf = append(*buf, ':')
			prefixLogItoa(buf, min, 2)
			*buf = append(*buf, ':')
			prefixLogItoa(buf, sec, 2)
			if flags&log.Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				prefixLogItoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}
	if flags&(log.Lshortfile|log.Llongfile) != 0 {
		if flags&log.Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, fmt.Sprintf("%s:%d: ", file, line)...)
	}
}

func prefixLogItoa(buf *[]byte, i int, wid int) {
	u := uint(i)
	if u == 0 && wid <= 1 {
		*buf = append(*buf, '0')
		return
	}
	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}
	*buf = append(*buf, b[bp:]...)
}
