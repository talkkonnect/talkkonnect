/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * radio-streamer.go -- streaming internet radio for talkkonnect (ffmpeg → ALSA)
 */

package talkkonnect

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type irStation struct {
	name    string
	url     string
	vol     int
	backend string
}

type streamingFFmpeg struct {
	mu              sync.Mutex
	cmd             *exec.Cmd
	ytdlpCmd        *exec.Cmd
	pipeW           io.WriteCloser
	in              io.WriteCloser
	isPlaying       bool
	lastStopPlanned bool
}

func (f *streamingFFmpeg) playing() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.isPlaying
}

func (f *streamingFFmpeg) takeWasPlannedStop() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	v := f.lastStopPlanned
	f.lastStopPlanned = false
	return v
}

func (f *streamingFFmpeg) close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.isPlaying && f.cmd == nil && f.ytdlpCmd == nil {
		return
	}
	f.lastStopPlanned = true
	f.isPlaying = false

	if f.ytdlpCmd != nil && f.ytdlpCmd.Process != nil {
		_ = f.ytdlpCmd.Process.Kill()
	}
	if f.pipeW != nil {
		_ = f.pipeW.Close()
		f.pipeW = nil
	}
	if f.cmd != nil && f.cmd.Process != nil {
		if f.in != nil {
			_, _ = f.in.Write([]byte("q"))
			_ = f.in.Close()
			f.in = nil
		} else {
			_ = f.cmd.Process.Kill()
		}
	} else if f.in != nil {
		_ = f.in.Close()
		f.in = nil
	}
	f.cmd = nil
	f.ytdlpCmd = nil
}

// forceKill terminates the ffmpeg child immediately (graceful quit via stdin is not reliable on exit).
func (f *streamingFFmpeg) forceKill() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastStopPlanned = true
	if f.ytdlpCmd != nil && f.ytdlpCmd.Process != nil {
		_ = f.ytdlpCmd.Process.Kill()
	}
	if f.pipeW != nil {
		_ = f.pipeW.Close()
		f.pipeW = nil
	}
	if f.cmd != nil && f.cmd.Process != nil {
		_ = f.cmd.Process.Kill()
	}
	if f.in != nil {
		_ = f.in.Close()
		f.in = nil
	}
	f.cmd = nil
	f.ytdlpCmd = nil
	f.isPlaying = false
}

func (f *streamingFFmpeg) start(streamURL string, gain float64, alsaDevice, ffmpegPath string, useYtDlp bool, ytDlpPath, ytFormat string, onExit func(wasPlanned bool)) error {
	if useYtDlp {
		return f.startYoutubePipe(streamURL, gain, alsaDevice, ffmpegPath, ytDlpPath, ytFormat, onExit)
	}
	return f.startDirectHTTP(streamURL, gain, alsaDevice, ffmpegPath, onExit)
}

func (f *streamingFFmpeg) startDirectHTTP(streamURL string, gain float64, alsaDevice, ffmpegPath string, onExit func(wasPlanned bool)) error {
	f.mu.Lock()
	if f.isPlaying {
		f.mu.Unlock()
		f.close()
		f.mu.Lock()
	}
	f.lastStopPlanned = false
	volStr := fmt.Sprintf("%.6f", gain)
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", streamURL,
		"-filter:a", "volume=" + volStr,
		"-f", "alsa", alsaDevice,
	}
	cmd := exec.Command(ffmpegPath, args...)
	in, err := cmd.StdinPipe()
	if err != nil {
		f.mu.Unlock()
		return err
	}
	if err := cmd.Start(); err != nil {
		f.mu.Unlock()
		_ = in.Close()
		return err
	}
	f.cmd = cmd
	f.in = in
	f.ytdlpCmd = nil
	f.pipeW = nil
	f.isPlaying = true
	proc := cmd
	f.mu.Unlock()

	go func() {
		err := proc.Wait()
		wasPlanned := f.takeWasPlannedStop()
		f.mu.Lock()
		f.isPlaying = false
		f.cmd = nil
		f.in = nil
		f.mu.Unlock()
		if err != nil && !wasPlanned {
			log.Printf("warn: internet radio ffmpeg exited: %v\n", err)
		}
		if onExit != nil {
			onExit(wasPlanned)
		}
	}()
	return nil
}

// startYoutubePipe streams YouTube / YouTube Music pages via yt-dlp stdout → ffmpeg (pipe:0) → ALSA.
func (f *streamingFFmpeg) startYoutubePipe(pageURL string, gain float64, alsaDevice, ffmpegPath, ytDlpPath, format string, onExit func(wasPlanned bool)) error {
	f.mu.Lock()
	if f.isPlaying {
		f.mu.Unlock()
		f.close()
		f.mu.Lock()
	}
	f.lastStopPlanned = false
	volStr := fmt.Sprintf("%.6f", gain)

	pr, pw := io.Pipe()
	ytArgs := []string{
		"-f", format,
		"-o", "-",
		"--no-playlist",
		"--no-warnings",
		pageURL,
	}
	ytCmd := exec.Command(ytDlpPath, ytArgs...)
	ytCmd.Stdout = pw
	ytCmd.Stderr = os.Stderr

	ffArgs := []string{
		"-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-filter:a", "volume=" + volStr,
		"-f", "alsa", alsaDevice,
	}
	ffCmd := exec.Command(ffmpegPath, ffArgs...)
	ffCmd.Stdin = pr

	if err := ffCmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		f.mu.Unlock()
		return err
	}
	if err := ytCmd.Start(); err != nil {
		if ffCmd.Process != nil {
			_ = ffCmd.Process.Kill()
		}
		_ = pw.Close()
		_ = pr.Close()
		f.mu.Unlock()
		return err
	}

	f.cmd = ffCmd
	f.ytdlpCmd = ytCmd
	f.pipeW = pw
	f.in = nil
	f.isPlaying = true
	ffProc := ffCmd
	ytProc := ytCmd
	pipeWriter := pw
	f.mu.Unlock()

	go func() {
		_ = ytProc.Wait()
		_ = pipeWriter.Close()
		err := ffProc.Wait()
		wasPlanned := f.takeWasPlannedStop()
		f.mu.Lock()
		f.isPlaying = false
		f.cmd = nil
		f.ytdlpCmd = nil
		f.pipeW = nil
		f.mu.Unlock()
		if err != nil && !wasPlanned {
			log.Printf("warn: internet radio ffmpeg (youtube pipe) exited: %v\n", err)
		}
		if onExit != nil {
			onExit(wasPlanned)
		}
	}()
	return nil
}

type internetRadioMgr struct {
	mu             sync.Mutex
	stations       []irStation
	currentIdx     int
	intentOn       bool
	interrupted    bool
	ducking        bool
	manualPaused   bool
	gainTrim       float64
	autoResume     *time.Timer
	ff             streamingFFmpeg
	retryTimer     *time.Timer
	inhibitRetries bool
}

var ir internetRadioMgr

func irDefaultStations() []irStation {
	return []irStation{
		{"comedy", "https://listen.181fm.com/181-comedy_128k.mp3", 0, ""},
		{"WBEZ 91.5", "http://stream.wbez.org/wbez128.mp3", 0, ""},
		{"WGN", "http://provisioning.streamtheworld.com/pls/WGNPLUSAM.pls", 0, ""},
	}
}

func irLoadStationList() []irStation {
	var out []irStation
	for _, s := range Config.Global.StreamingRadio.Stations.Station {
		u := strings.TrimSpace(s.URL)
		if u == "" {
			continue
		}
		out = append(out, irStation{strings.TrimSpace(s.Name), u, s.Volume, strings.TrimSpace(s.Backend)})
	}
	if len(out) == 0 {
		return irDefaultStations()
	}
	return out
}

func internetRadioConfigureFromXML() {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.stations = irLoadStationList()
	if ir.gainTrim <= 0 {
		ir.gainTrim = 1.0
	}
	if ir.currentIdx >= len(ir.stations) {
		ir.currentIdx = 0
	}
	if !Config.Global.StreamingRadio.Enabled {
		ir.cancelTimerLocked()
		ir.cancelRetryLocked()
		ir.inhibitRetries = true
		ir.ff.close()
		ir.intentOn = false
		ir.interrupted = false
		ir.ducking = false
		ir.manualPaused = false
		ir.currentIdx = -1
	}
}

func internetRadioStationCount() int {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if !Config.Global.StreamingRadio.Enabled {
		return 0
	}
	return len(ir.stations)
}

func irAlsaDevice() string {
	s := strings.TrimSpace(Config.Global.StreamingRadio.AlsaDevice)
	if s == "" {
		return "default"
	}
	return s
}

func irFFmpegPath() string {
	s := strings.TrimSpace(Config.Global.StreamingRadio.FFmpegPath)
	if s == "" {
		return "/usr/bin/ffmpeg"
	}
	return s
}

func irYtDlpPath() string {
	s := strings.TrimSpace(Config.Global.StreamingRadio.YtDlpPath)
	if s == "" {
		return "/usr/bin/yt-dlp"
	}
	return s
}

func irYtDlpFormat() string {
	s := strings.TrimSpace(Config.Global.StreamingRadio.YtDlpFormat)
	if s == "" {
		return "bestaudio/best"
	}
	return s
}

func irURLLooksLikeYoutube(url string) bool {
	u := strings.ToLower(url)
	return strings.Contains(u, "youtube.com") || strings.Contains(u, "youtu.be")
}

// irStationUsesYtDlp selects yt-dlp→ffmpeg pipe playback for YouTube Music / YouTube URLs when enabled.
// Per-station Backend: empty or "auto" uses global YoutubeMusicPlayback + URL detection; "youtube" forces yt-dlp; "http"/"direct"/"ffmpeg" forces plain ffmpeg -i.
func irStationUsesYtDlp(st irStation) bool {
	b := strings.ToLower(strings.TrimSpace(st.backend))
	switch b {
	case "youtube", "yt-dlp", "ytdlp", "youtubemusic":
		return true
	case "http", "direct", "ffmpeg", "url":
		return false
	default:
		if !Config.Global.StreamingRadio.YoutubeMusicPlayback {
			return false
		}
		return irURLLooksLikeYoutube(st.url)
	}
}

func interruptionModeNormalized() string {
	return strings.ToLower(strings.TrimSpace(Config.Global.StreamingRadio.InterruptionMode))
}

func (ir *internetRadioMgr) cancelTimerLocked() {
	if ir.autoResume != nil {
		ir.autoResume.Stop()
		ir.autoResume = nil
	}
}

func (ir *internetRadioMgr) cancelRetryLocked() {
	if ir.retryTimer != nil {
		ir.retryTimer.Stop()
		ir.retryTimer = nil
	}
}

func (ir *internetRadioMgr) resetAutoResumeTimerLocked() {
	ir.cancelTimerLocked()
	d := time.Duration(Config.Global.StreamingRadio.AutoResumeDelay) * time.Second
	if d <= 0 {
		d = 15 * time.Second
	}
	ir.autoResume = time.AfterFunc(d, func() {
		ir.mu.Lock()
		defer ir.mu.Unlock()
		ir.tryAutoResumeLocked()
	})
}

func (ir *internetRadioMgr) tryAutoResumeLocked() {
	if !Config.Global.StreamingRadio.Enabled || !ir.intentOn || ir.manualPaused {
		return
	}
	if ir.interrupted {
		ir.resumePlaybackLocked(false)
	}
}

func (ir *internetRadioMgr) computeGainLocked(duck bool) float64 {
	master := float64(Config.Global.StreamingRadio.MasterVolume) / 100.0
	if master <= 0 {
		master = 0.5
	}
	stVol := 1.0
	if ir.currentIdx >= 0 && ir.currentIdx < len(ir.stations) {
		v := ir.stations[ir.currentIdx].vol
		if v > 0 {
			stVol = float64(v) / 100.0
		}
	}
	if ir.gainTrim <= 0 {
		ir.gainTrim = 1.0
	}
	g := master * stVol * ir.gainTrim
	if duck {
		dp := float64(Config.Global.StreamingRadio.DuckVolumePercent) / 100.0
		if dp <= 0 || dp > 1 {
			dp = 0.1
		}
		g *= dp
	}
	if g > 2.0 {
		g = 2.0
	}
	if g < 0.00005 {
		g = 0.00005
	}
	return g
}

func (ir *internetRadioMgr) applyInterruptLocked() {
	mode := interruptionModeNormalized()
	switch mode {
	case "duck":
		ir.ducking = true
		ir.interrupted = true
		if ir.currentIdx < 0 || ir.currentIdx >= len(ir.stations) {
			return
		}
		g := ir.computeGainLocked(true)
		ir.inhibitRetries = true
		ir.ff.close()
		ir.inhibitRetries = false
		st := ir.stations[ir.currentIdx]
		err := ir.ff.start(st.url, g, irAlsaDevice(), irFFmpegPath(), irStationUsesYtDlp(st), irYtDlpPath(), irYtDlpFormat(), ir.makeExitHandler())
		if err != nil {
			log.Println("error: internet radio duck playback:", err)
		}
		internetRadioLCDStatus("[Radio: Duck]")
	default:
		ir.ducking = false
		ir.interrupted = true
		ir.inhibitRetries = true
		ir.ff.close()
		ir.inhibitRetries = false
		internetRadioLCDStatus("[Radio: Paused]")
	}
}

func (ir *internetRadioMgr) makeExitHandler() func(bool) {
	return func(wasPlanned bool) {
		ir.mu.Lock()
		defer ir.mu.Unlock()
		if wasPlanned || ir.inhibitRetries {
			return
		}
		if !ir.intentOn || ir.manualPaused || !Config.Global.StreamingRadio.Enabled {
			return
		}
		// Deliberate stop/pause interrupt: ffmpeg already torn down; do not auto-retry.
		if ir.interrupted && interruptionModeNormalized() != "duck" {
			return
		}
		ir.scheduleReconnectLocked()
	}
}

func (ir *internetRadioMgr) scheduleReconnectLocked() {
	ir.cancelRetryLocked()
	wait := time.Duration(Config.Global.StreamingRadio.StreamRetrySecs) * time.Second
	if wait <= 0 {
		wait = 5 * time.Second
	}
	ir.retryTimer = time.AfterFunc(wait, func() {
		ir.mu.Lock()
		defer ir.mu.Unlock()
		if !ir.intentOn || ir.manualPaused || !Config.Global.StreamingRadio.Enabled {
			return
		}
		if ir.interrupted {
			return
		}
		log.Println("info: internet radio reconnecting after stream error…")
		ir.resumePlaybackLocked(false)
	})
}

func (ir *internetRadioMgr) resumePlaybackLocked(fromUser bool) {
	if ir.currentIdx < 0 || ir.currentIdx >= len(ir.stations) {
		if len(ir.stations) > 0 {
			ir.currentIdx = 0
		} else {
			return
		}
	}
	ir.interrupted = false
	ir.ducking = false
	g := ir.computeGainLocked(false)
	ir.inhibitRetries = true
	ir.ff.close()
	ir.inhibitRetries = false
	st := ir.stations[ir.currentIdx]
	err := ir.ff.start(st.url, g, irAlsaDevice(), irFFmpegPath(), irStationUsesYtDlp(st), irYtDlpPath(), irYtDlpFormat(), ir.makeExitHandler())
	if err != nil {
		log.Println("error: internet radio playback:", err)
		if !fromUser {
			ir.scheduleReconnectLocked()
		}
		return
	}
	internetRadioLCDNowPlaying(ir.stations[ir.currentIdx].name)
}

func internetRadioNotifyVoiceOrTX() {
	if !Config.Global.StreamingRadio.Enabled {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if len(ir.stations) == 0 {
		return
	}
	ir.manualPaused = false
	if !ir.intentOn {
		return
	}
	ir.resetAutoResumeTimerLocked()
	if ir.interrupted {
		return
	}
	ir.applyInterruptLocked()
}

func internetRadioQuickToggle(b *Talkkonnect) {
	if !Config.Global.StreamingRadio.Enabled {
		log.Println("warn: Internet streaming radio is disabled in config (<Radio><Enabled>)")
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if len(ir.stations) == 0 {
		ir.stations = irLoadStationList()
	}
	if len(ir.stations) == 0 {
		log.Println("warn: No internet radio stations configured")
		return
	}

	onAir := ir.intentOn && (ir.ff.playing() || ir.interrupted)
	if onAir {
		ir.cancelTimerLocked()
		ir.cancelRetryLocked()
		ir.intentOn = false
		ir.manualPaused = false
		ir.interrupted = false
		ir.ducking = false
		ir.inhibitRetries = true
		ir.ff.close()
		ir.inhibitRetries = false
		ir.currentIdx = -1
		internetRadioLCDStatus("[Radio: Off]")
		return
	}

	ir.intentOn = true
	ir.manualPaused = false
	if ir.currentIdx < 0 {
		ir.currentIdx = 0
	}
	ir.cancelTimerLocked()
	ir.cancelRetryLocked()
	ir.resumePlaybackLocked(true)
	internetRadioAnnounceIfNeeded(b, ir.stations[ir.currentIdx].name)
}

func internetRadioNextStation(b *Talkkonnect) {
	if !Config.Global.StreamingRadio.Enabled {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if len(ir.stations) == 0 {
		return
	}
	ir.currentIdx = (ir.currentIdx + 1) % len(ir.stations)
	name := ir.stations[ir.currentIdx].name
	if ir.intentOn && !ir.interrupted {
		ir.resumePlaybackLocked(true)
	} else if ir.intentOn && ir.interrupted {
		internetRadioLCDNowPlaying(name)
	} else {
		internetRadioLCDNowPlaying(name)
	}
	internetRadioAnnounceIfNeeded(b, name)
}

func internetRadioPrevStation(b *Talkkonnect) {
	if !Config.Global.StreamingRadio.Enabled {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if len(ir.stations) == 0 {
		return
	}
	ir.currentIdx--
	if ir.currentIdx < 0 {
		ir.currentIdx = len(ir.stations) - 1
	}
	name := ir.stations[ir.currentIdx].name
	if ir.intentOn && !ir.interrupted {
		ir.resumePlaybackLocked(true)
	} else {
		internetRadioLCDNowPlaying(name)
	}
	internetRadioAnnounceIfNeeded(b, name)
}

func internetRadioVolUp() {
	if !Config.Global.StreamingRadio.Enabled {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.gainTrim <= 0 {
		ir.gainTrim = 1.0
	}
	ir.gainTrim *= 1.12
	if ir.gainTrim > 4.0 {
		ir.gainTrim = 4.0
	}
	if ir.intentOn && !ir.interrupted && ir.ff.playing() {
		ir.resumePlaybackLocked(false)
	}
}

func internetRadioVolDown() {
	if !Config.Global.StreamingRadio.Enabled {
		return
	}
	ir.mu.Lock()
	defer ir.mu.Unlock()
	if ir.gainTrim <= 0 {
		ir.gainTrim = 1.0
	}
	ir.gainTrim /= 1.12
	if ir.gainTrim < 0.2 {
		ir.gainTrim = 0.2
	}
	if ir.intentOn && !ir.interrupted && ir.ff.playing() {
		ir.resumePlaybackLocked(false)
	}
}

func internetRadioAnnounceIfNeeded(b *Talkkonnect, stationName string) {
	if b == nil || !Config.Global.StreamingRadio.AnnounceStationTTS || !Config.Global.Software.TTS.Enabled {
		return
	}
	name := strings.TrimSpace(stationName)
	if name == "" {
		return
	}
	msg := "Radio station: " + name
	go b.Speak(msg, "local", Config.Global.Software.TTS.Volumelevel, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
}

func internetRadioLCDStatus(status string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}
	s := strings.TrimSpace(status)
	if OLEDEnabled && s != "" {
		oledDisplay(false, 7, OLEDStartColumn, s)
	}
	if LCDEnabled && s != "" {
		LcdText[3] = s
		LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
	}
}

func internetRadioLCDNowPlaying(name string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}
	line := name
	if len(line) > 20 {
		line = line[:17] + "..."
	}
	if OLEDEnabled {
		oledDisplay(false, 6, OLEDStartColumn, "[FM] "+line)
	}
	if LCDEnabled {
		LcdText[3] = "[FM] " + line
		LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
	}
}

// InternetRadioStatus is JSON-friendly internet radio state for external UI clients.
type InternetRadioStatus struct {
	Enabled      bool   `json:"enabled"`
	Playing      bool   `json:"playing"`
	Status       string `json:"status"`
	StationName  string `json:"stationName"`
	StationIndex int    `json:"stationIndex"`
	StationCount int    `json:"stationCount"`
	Volume       int    `json:"volume"`
}

func InternetRadioStatusSnapshot() InternetRadioStatus {
	ir.mu.Lock()
	defer ir.mu.Unlock()

	st := InternetRadioStatus{
		Enabled: Config.Global.StreamingRadio.Enabled,
	}
	if !st.Enabled {
		st.Status = "off"
		return st
	}

	st.StationCount = len(ir.stations)
	st.StationIndex = ir.currentIdx

	if ir.currentIdx >= 0 && ir.currentIdx < len(ir.stations) {
		st.StationName = ir.stations[ir.currentIdx].name
	}

	if ir.gainTrim <= 0 {
		ir.gainTrim = 1.0
	}
	st.Volume = int((ir.gainTrim / 4.0) * 100)
	if st.Volume > 100 {
		st.Volume = 100
	}
	if st.Volume < 0 {
		st.Volume = 0
	}

	switch {
	case !ir.intentOn:
		st.Status = "off"
	case ir.ducking:
		st.Status = "ducking"
		st.Playing = ir.ff.playing()
	case ir.interrupted || ir.manualPaused:
		st.Status = "paused"
	case ir.ff.playing():
		st.Status = "playing"
		st.Playing = true
	default:
		st.Status = "idle"
	}

	return st
}

// internetRadioShutdownKill stops timers and kills the internet-radio ffmpeg process. Safe to call multiple times.
func internetRadioShutdownKill() {
	ir.mu.Lock()
	defer ir.mu.Unlock()
	ir.cancelTimerLocked()
	ir.cancelRetryLocked()
	ir.inhibitRetries = true
	ir.ff.forceKill()
	ir.intentOn = false
	ir.interrupted = false
	ir.ducking = false
}
