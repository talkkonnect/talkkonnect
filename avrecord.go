/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * avrecord.go -> record incoming Mumble voice traffic as raw Opus packets in
 * a custom .mrec binary format. Each block is:
 *   [8] timestamp ms, [4] session id, [2] sequence, [2] username len,
 *   [2] opus payload len, [username], [opus payload].
 * Video capture uses external tools elsewhere.
 */

package talkkonnect

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

const (
	mrecBlockHeaderSize = 18
	defaultMrecBufSize  = 64 * 1024
	defaultChannelBuf   = 4096
	defaultFlushMS      = 1000
	defaultMaxFileSize  = 100 * 1024 * 1024
	defaultBaseName     = "talkkonnect"
	defaultIndexLog     = "recording_index.log"
)

type recordEntry struct {
	sessionID uint32
	sequence  uint16
	username  string
	payload   []byte
	received  time.Time
}

type mrecWriter struct {
	savePath      string
	baseName      string
	maxFileSize   int64
	indexLogPath  string
	flushInterval time.Duration

	mu           sync.Mutex
	file         *os.File
	buf          *bufio.Writer
	bytesWritten int64
	fileStart    time.Time
	currentPath  string
	header       [mrecBlockHeaderSize]byte
}

type opusRecorder struct {
	mu          sync.Mutex
	running     bool
	detacher    gumble.Detacher
	recordChan  chan recordEntry
	stopChan    chan struct{}
	writerDone  chan struct{}
	dropCount   uint64
	dropLogMu   sync.Mutex
	dropLastLog time.Time
}

var globalOpusRecorder opusRecorder

func resolvedAudioRecordSettings() (savePath, baseName, indexLog string, maxFileSize int64, channelBuf, flushMS int) {
	cfg := Config.Global.Hardware.AudioRecordFunction

	savePath = cfg.RecordSavePath
	baseName = cfg.RecordBaseName
	if baseName == "" {
		baseName = defaultBaseName
	}

	maxFileSize = cfg.MaxFileSize
	if maxFileSize <= 0 {
		maxFileSize = defaultMaxFileSize
	}

	indexLog = cfg.RecordIndexLog
	if indexLog == "" {
		indexLog = defaultIndexLog
	}

	channelBuf = cfg.ChannelBufferSize
	if channelBuf <= 0 {
		channelBuf = defaultChannelBuf
	}

	flushMS = cfg.WriteFlushInterval
	if flushMS <= 0 {
		flushMS = defaultFlushMS
	}

	return savePath, baseName, indexLog, maxFileSize, channelBuf, flushMS
}

// StartOpusTrafficRecording attaches an audio listener and writes raw Opus packets to .mrec files.
func StartOpusTrafficRecording(b *Talkkonnect) {
	if b == nil || b.Config == nil {
		log.Println("error: cannot start Opus traffic recording without Mumble config")
		return
	}

	globalOpusRecorder.mu.Lock()
	defer globalOpusRecorder.mu.Unlock()

	if globalOpusRecorder.running {
		log.Println("info: Opus traffic recording is already running")
		return
	}

	savePath, baseName, indexLog, maxFileSize, channelBuf, flushMS := resolvedAudioRecordSettings()
	if savePath == "" {
		log.Println("error: recordsavepath is empty; cannot start Opus traffic recording")
		return
	}

	createDirIfNotExist(savePath)

	indexLogPath := indexLog
	if !filepath.IsAbs(indexLogPath) {
		indexLogPath = filepath.Join(savePath, indexLog)
	}

	writer := &mrecWriter{
		savePath:      savePath,
		baseName:      baseName,
		maxFileSize:   maxFileSize,
		indexLogPath:  indexLogPath,
		flushInterval: time.Duration(flushMS) * time.Millisecond,
	}

	globalOpusRecorder.recordChan = make(chan recordEntry, channelBuf)
	globalOpusRecorder.stopChan = make(chan struct{})
	globalOpusRecorder.writerDone = make(chan struct{})

	go writer.run(globalOpusRecorder.recordChan, globalOpusRecorder.stopChan, globalOpusRecorder.writerDone)

	globalOpusRecorder.detacher = b.Config.AttachAudio(&globalOpusRecorder)
	globalOpusRecorder.running = true

	log.Printf("info: Opus traffic recording started; save path=%s base=%s max size=%d bytes\n", savePath, baseName, maxFileSize)
}

// StopOpusTrafficRecording detaches the listener and flushes the active .mrec file.
func StopOpusTrafficRecording() {
	globalOpusRecorder.mu.Lock()
	if !globalOpusRecorder.running {
		globalOpusRecorder.mu.Unlock()
		return
	}

	globalOpusRecorder.running = false
	if globalOpusRecorder.detacher != nil {
		globalOpusRecorder.detacher.Detach()
		globalOpusRecorder.detacher = nil
	}

	stopChan := globalOpusRecorder.stopChan
	writerDone := globalOpusRecorder.writerDone
	globalOpusRecorder.mu.Unlock()

	close(stopChan)
	<-writerDone

	log.Println("info: Opus traffic recording stopped")
}

func (r *opusRecorder) OnAudioStream(e *gumble.AudioStreamEvent) {
	SafeGo(func() {
		for packet := range e.C {
			if packet == nil || len(packet.OpusPayload) == 0 {
				continue
			}

			payload := make([]byte, len(packet.OpusPayload))
			copy(payload, packet.OpusPayload)

			var sessionID uint32
			var username string
			if packet.Sender != nil {
				sessionID = packet.Sender.Session
				username = cleanstring(packet.Sender.Name)
			} else if e.User != nil {
				sessionID = e.User.Session
				username = cleanstring(e.User.Name)
			}

			r.enqueue(recordEntry{
				sessionID: sessionID,
				sequence:  packet.Sequence,
				username:  username,
				payload:   payload,
				received:  time.Now(),
			})
		}
	})
}

func (r *opusRecorder) enqueue(entry recordEntry) {
	r.mu.Lock()
	ch := r.recordChan
	r.mu.Unlock()
	if ch == nil {
		return
	}

	select {
	case ch <- entry:
	default:
		atomic.AddUint64(&r.dropCount, 1)
		r.logDropRateLimited()
	}
}

func (r *opusRecorder) logDropRateLimited() {
	r.dropLogMu.Lock()
	defer r.dropLogMu.Unlock()
	if time.Since(r.dropLastLog) < 5*time.Second {
		return
	}
	dropped := atomic.SwapUint64(&r.dropCount, 0)
	r.dropLastLog = time.Now()
	fmt.Fprintf(os.Stderr, "mrec: record channel full, dropped %d packet(s)\n", dropped)
}

func (w *mrecWriter) run(ch <-chan recordEntry, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)

	flushTicker := time.NewTicker(w.flushInterval)
	defer flushTicker.Stop()

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				w.closeFile()
				return
			}
			if err := w.writeEntry(entry); err != nil {
				fmt.Fprintf(os.Stderr, "mrec: write error: %v\n", err)
				w.closeFile()
				return
			}
		case <-flushTicker.C:
			w.flush()
		case <-stop:
			for {
				select {
				case entry, ok := <-ch:
					if !ok {
						w.closeFile()
						return
					}
					if err := w.writeEntry(entry); err != nil {
						fmt.Fprintf(os.Stderr, "mrec: write error during shutdown: %v\n", err)
					}
				default:
					w.closeFile()
					return
				}
			}
		}
	}
}

func (w *mrecWriter) appendIndexLog(absPath string, started time.Time) error {
	f, err := os.OpenFile(w.indexLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s %s\n", absPath, started.Format(time.RFC3339))
	return err
}

func (w *mrecWriter) writeEntry(entry recordEntry) error {
	if len(entry.payload) > 0xFFFF {
		fmt.Fprintf(os.Stderr, "mrec: dropping oversized opus payload (%d bytes)\n", len(entry.payload))
		return nil
	}

	usernameBytes := []byte(entry.username)
	if len(usernameBytes) > 0xFFFF {
		fmt.Fprintf(os.Stderr, "mrec: dropping oversized username (%d bytes)\n", len(usernameBytes))
		return nil
	}

	blockSize := int64(mrecBlockHeaderSize + len(usernameBytes) + len(entry.payload))

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openNewFileLocked(); err != nil {
			return err
		}
	}

	if w.maxFileSize > 0 && w.bytesWritten > 0 && w.bytesWritten+blockSize > w.maxFileSize {
		if err := w.rotateLocked(); err != nil {
			return err
		}
	}

	ts := entry.received.Sub(w.fileStart).Milliseconds()
	if ts < 0 {
		ts = 0
	}

	binary.BigEndian.PutUint64(w.header[0:], uint64(ts))
	binary.BigEndian.PutUint32(w.header[8:], entry.sessionID)
	binary.BigEndian.PutUint16(w.header[12:], entry.sequence)
	binary.BigEndian.PutUint16(w.header[14:], uint16(len(usernameBytes)))
	binary.BigEndian.PutUint16(w.header[16:], uint16(len(entry.payload)))

	if _, err := w.buf.Write(w.header[:]); err != nil {
		return err
	}
	if len(usernameBytes) > 0 {
		if _, err := w.buf.Write(usernameBytes); err != nil {
			return err
		}
	}
	if _, err := w.buf.Write(entry.payload); err != nil {
		return err
	}

	w.bytesWritten += blockSize
	return nil
}

func (w *mrecWriter) rotateLocked() error {
	if err := w.flushLocked(); err != nil {
		return err
	}
	if w.file != nil {
		if err := w.file.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "mrec: sync before rotate: %v\n", err)
		}
	}
	return w.openNewFileLocked()
}

func (w *mrecWriter) flush() {
	w.mu.Lock()
	defer w.mu.Unlock()
	_ = w.flushLocked()
}

func (w *mrecWriter) flushLocked() error {
	if w.buf == nil {
		return nil
	}
	return w.buf.Flush()
}

func (w *mrecWriter) closeFile() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closeFileLocked()
}

func (w *mrecWriter) closeFileLocked() {
	if w.buf != nil {
		_ = w.buf.Flush()
		w.buf = nil
	}
	if w.file != nil {
		_ = w.file.Sync()
		_ = w.file.Close()
		w.file = nil
	}
	w.bytesWritten = 0
	w.currentPath = ""
}

func (w *mrecWriter) openNewFileLocked() error {
	now := time.Now()
	filename := fmt.Sprintf("%s_%s.mrec", w.baseName, now.Format("20060102_150405"))
	path := filepath.Join(w.savePath, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	if w.file != nil {
		_ = w.file.Close()
	}

	w.file = f
	w.buf = bufio.NewWriterSize(f, defaultMrecBufSize)
	w.bytesWritten = 0
	w.fileStart = now
	w.currentPath = absPath

	if err := w.appendIndexLog(absPath, now); err != nil {
		fmt.Fprintf(os.Stderr, "mrec: index log error: %v\n", err)
	}

	log.Printf("info: opened mrec file %s\n", absPath)
	return nil
}
