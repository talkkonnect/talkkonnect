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
 *   [8] file offset ms, [8] realtime unix ms, [4] session id, [2] sequence,
 *   [2] username len, [2] opus payload len, [username], [opus payload].
 * Filenames are: {basename}-{server}-{channel}-{timestamp}.mrec
 * Video capture uses external tools elsewhere.
 */

package talkkonnect

import (
	"bufio"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/talkkonnect/gumble/gumble"
)

const (
	mrecBlockHeaderSize = 26
	defaultMrecBufSize  = 64 * 1024
	defaultChannelBuf   = 4096
	defaultFlushMS      = 1000
	defaultMaxFileSize  = 100 * 1024 * 1024
	defaultBaseName     = "talkkonnect"
	defaultIndexLog     = "recording_index.log"
	defaultMySQLPort    = 3306

	transmissionSweepInterval = 500 * time.Millisecond
	transmissionIdleTimeout   = 600 * time.Millisecond
	minTransmissionDuration   = 200 * time.Millisecond
)

var recordDBSchema = []string{
	`CREATE TABLE IF NOT EXISTS mrec_files (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		filename VARCHAR(1024) NOT NULL,
		server_name VARCHAR(255) NOT NULL,
		channel_name VARCHAR(255) NOT NULL,
		file_start_time DATETIME(3) NOT NULL,
		UNIQUE KEY uk_mrec_files_filename (filename)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE TABLE IF NOT EXISTS transmissions (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		file_id BIGINT UNSIGNED NOT NULL,
		session_id INT UNSIGNED NOT NULL,
		username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
		start_time DATETIME(3) NOT NULL,
		end_time DATETIME(3) NOT NULL,
		duration_ms BIGINT NOT NULL,
		offset_start_ms BIGINT NOT NULL,
		offset_end_ms BIGINT NOT NULL,
		notes TEXT NULL,
		CONSTRAINT fk_file
			FOREIGN KEY (file_id)
			REFERENCES mrec_files (id)
			ON DELETE CASCADE
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	`CREATE INDEX IF NOT EXISTS idx_transmissions_time
		ON transmissions (start_time DESC, end_time DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_transmissions_duration
		ON transmissions (duration_ms)`,
	`CREATE INDEX IF NOT EXISTS idx_transmissions_username
		ON transmissions (username)`,
	`CREATE INDEX IF NOT EXISTS idx_transmissions_file_id
		ON transmissions (file_id)`,
	`ALTER TABLE transmissions
		MODIFY username VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL`,
}

type recordEntry struct {
	sessionID uint32
	sequence  uint16
	username  string
	payload   []byte
	received  time.Time
}

type ActiveTransmission struct {
	FileID        int64
	SessionID     uint32
	Username      string
	StartTime     time.Time
	EndTime       time.Time
	OffsetStartMs int64
	OffsetEndMs   int64
	LastPacketAt  time.Time
}

type RecordingManager struct {
	DB           *sql.DB
	ActiveBursts map[uint32]*ActiveTransmission
	mu           sync.Mutex
	stopCh       chan struct{}
	doneCh       chan struct{}
}

type mrecWriter struct {
	savePath      string
	baseName      string
	maxFileSize   int64
	indexLogPath  string
	flushInterval time.Duration
	tk            *Talkkonnect
	recManager    *RecordingManager

	mu            sync.Mutex
	file          *os.File
	buf           *bufio.Writer
	bytesWritten  int64
	fileStart     time.Time
	currentPath   string
	currentFileID int64
	header        [mrecBlockHeaderSize]byte
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

func mysqlDSNFromConfig() (string, bool) {
	cfg := Config.Global.Hardware.AudioRecordFunction.RecordDB
	if !cfg.Enabled {
		return "", false
	}

	host := strings.TrimSpace(cfg.Host)
	user := strings.TrimSpace(cfg.User)
	database := strings.TrimSpace(cfg.Database)
	if host == "" || user == "" || database == "" {
		return "", false
	}

	port := cfg.Port
	if port <= 0 {
		port = defaultMySQLPort
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		user, cfg.Password, host, port, database)
	return dsn, true
}

func NewRecordingManager(dsn string) (*RecordingManager, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql database: %w", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping mysql database: %w", err)
	}

	for _, stmt := range recordDBSchema {
		if _, err := db.Exec(stmt); err != nil {
			db.Close()
			return nil, fmt.Errorf("init mysql schema: %w", err)
		}
	}

	rm := &RecordingManager{
		DB:           db,
		ActiveBursts: make(map[uint32]*ActiveTransmission),
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
	go rm.sweepClosedTransmissions()
	return rm, nil
}

func (rm *RecordingManager) RegisterFile(filename, serverName, channelName string, fileStart time.Time) (int64, error) {
	res, err := rm.DB.Exec(
		`INSERT INTO mrec_files (filename, server_name, channel_name, file_start_time) VALUES (?, ?, ?, ?)`,
		filename, serverName, channelName, fileStart.UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (rm *RecordingManager) ProcessPacket(fileID int64, sessionID uint32, username string, offsetMs int64, packetTime time.Time) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	burst, exists := rm.ActiveBursts[sessionID]
	if !exists {
		rm.ActiveBursts[sessionID] = &ActiveTransmission{
			FileID:        fileID,
			SessionID:     sessionID,
			Username:      username,
			StartTime:     packetTime,
			EndTime:       packetTime,
			OffsetStartMs: offsetMs,
			OffsetEndMs:   offsetMs,
			LastPacketAt:  time.Now(),
		}
		return
	}

	burst.EndTime = packetTime
	burst.OffsetEndMs = offsetMs
	burst.LastPacketAt = time.Now()
	if username != "" {
		burst.Username = username
	}
}

func (rm *RecordingManager) sweepClosedTransmissions() {
	defer close(rm.doneCh)

	ticker := time.NewTicker(transmissionSweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.stopCh:
			return
		case <-ticker.C:
			rm.sweepOnce()
		}
	}
}

func (rm *RecordingManager) sweepOnce() {
	rm.mu.Lock()
	now := time.Now()
	var toWrite []ActiveTransmission

	for sessionID, burst := range rm.ActiveBursts {
		if now.Sub(burst.LastPacketAt) <= transmissionIdleTimeout {
			continue
		}

		if burst.EndTime.Sub(burst.StartTime) >= minTransmissionDuration {
			toWrite = append(toWrite, *burst)
		}
		delete(rm.ActiveBursts, sessionID)
	}
	rm.mu.Unlock()

	for i := range toWrite {
		b := &toWrite[i]
		rm.writeToDB(b, b.EndTime.Sub(b.StartTime).Milliseconds())
	}
}

func (rm *RecordingManager) FlushAll() {
	rm.mu.Lock()
	toWrite := make([]ActiveTransmission, 0, len(rm.ActiveBursts))
	for sessionID, burst := range rm.ActiveBursts {
		if burst.EndTime.Sub(burst.StartTime) >= minTransmissionDuration {
			toWrite = append(toWrite, *burst)
		}
		delete(rm.ActiveBursts, sessionID)
	}
	rm.mu.Unlock()

	for i := range toWrite {
		b := &toWrite[i]
		rm.writeToDB(b, b.EndTime.Sub(b.StartTime).Milliseconds())
	}
}

func (rm *RecordingManager) writeToDB(b *ActiveTransmission, durationMs int64) {
	_, err := rm.DB.Exec(
		`INSERT INTO transmissions
			(file_id, session_id, username, start_time, end_time, duration_ms, offset_start_ms, offset_end_ms)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		b.FileID,
		b.SessionID,
		b.Username,
		b.StartTime.UTC(),
		b.EndTime.UTC(),
		durationMs,
		b.OffsetStartMs,
		b.OffsetEndMs,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mrec: mysql transmission insert error: %v\n", err)
	}
}

func (rm *RecordingManager) Close() {
	rm.FlushAll()
	close(rm.stopCh)
	<-rm.doneCh
	_ = rm.DB.Close()
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

	var recManager *RecordingManager
	if dsn, ok := mysqlDSNFromConfig(); ok {
		var err error
		recManager, err = NewRecordingManager(dsn)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mrec: mysql index disabled: %v\n", err)
		}
	}

	writer := &mrecWriter{
		savePath:      savePath,
		baseName:      baseName,
		maxFileSize:   maxFileSize,
		indexLogPath:  indexLogPath,
		flushInterval: time.Duration(flushMS) * time.Millisecond,
		tk:            b,
		recManager:    recManager,
	}

	globalOpusRecorder.recordChan = make(chan recordEntry, channelBuf)
	globalOpusRecorder.stopChan = make(chan struct{})
	globalOpusRecorder.writerDone = make(chan struct{})

	go writer.run(globalOpusRecorder.recordChan, globalOpusRecorder.stopChan, globalOpusRecorder.writerDone)

	globalOpusRecorder.detacher = b.Config.AttachAudio(&globalOpusRecorder)
	globalOpusRecorder.running = true

	log.Printf("info: Opus traffic recording started; save path=%s base=%s max size=%d bytes\n", savePath, baseName, maxFileSize)
	if recManager != nil {
		cfg := Config.Global.Hardware.AudioRecordFunction.RecordDB
		port := cfg.Port
		if port <= 0 {
			port = defaultMySQLPort
		}
		log.Printf("info: Opus transmission index database %s@%s:%d/%s\n",
			cfg.User, cfg.Host, port, cfg.Database)
	}
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
				username = recordUsername(packet.Sender.Name)
			} else if e.User != nil {
				sessionID = e.User.Session
				username = recordUsername(e.User.Name)
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
	defer func() {
		if w.recManager != nil {
			w.recManager.Close()
		}
	}()

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
	binary.BigEndian.PutUint64(w.header[8:], uint64(entry.received.UnixMilli()))
	binary.BigEndian.PutUint32(w.header[16:], entry.sessionID)
	binary.BigEndian.PutUint16(w.header[20:], entry.sequence)
	binary.BigEndian.PutUint16(w.header[22:], uint16(len(usernameBytes)))
	binary.BigEndian.PutUint16(w.header[24:], uint16(len(entry.payload)))

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

	if w.recManager != nil && w.currentFileID > 0 {
		w.recManager.ProcessPacket(w.currentFileID, entry.sessionID, entry.username, ts, entry.received)
	}

	return nil
}

func (w *mrecWriter) rotateLocked() error {
	if w.recManager != nil {
		w.recManager.FlushAll()
	}

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
	if w.recManager != nil {
		w.recManager.FlushAll()
	}

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
	w.currentFileID = 0
}

func (w *mrecWriter) openNewFileLocked() error {
	now := time.Now()
	server, channel := w.currentServerChannel()
	filename := fmt.Sprintf(
		"%s-%s-%s-%s.mrec",
		mrecFilenamePart(w.baseName),
		mrecFilenamePart(server),
		mrecFilenamePart(channel),
		now.Format("20060102-150405"),
	)
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

	if w.recManager != nil {
		fileID, err := w.recManager.RegisterFile(absPath, server, channel, now)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mrec: mysql file registry error: %v\n", err)
			w.currentFileID = 0
		} else {
			w.currentFileID = fileID
		}
	}

	if err := w.appendIndexLog(absPath, now); err != nil {
		fmt.Fprintf(os.Stderr, "mrec: index log error: %v\n", err)
	}

	log.Printf("info: opened mrec file %s\n", absPath)
	return nil
}

func (w *mrecWriter) currentServerChannel() (server, channel string) {
	if w.tk == nil {
		return "", ""
	}

	server = cleanstring(strings.TrimSpace(w.tk.Name))
	if w.tk.Client != nil && w.tk.Client.Self != nil && w.tk.Client.Self.Channel != nil {
		channel = cleanstring(strings.TrimSpace(w.tk.Client.Self.Channel.Name))
	} else {
		channel = cleanstring(strings.TrimSpace(w.tk.ChannelName))
	}
	return server, channel
}

func mrecFilenamePart(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "unknown"
	}

	var b strings.Builder
	b.Grow(len(s))
	lastHyphen := false
	for _, r := range s {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		case '-':
			b.WriteRune('-')
			lastHyphen = true
		default:
			b.WriteRune(r)
			lastHyphen = false
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "unknown"
	}
	return out
}
