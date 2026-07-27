/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * talkkonnect is the based on talkiepi and barnard by Daniel Chote and Tim Cooper
 *
 * The Initial Developer of the Original Code is
 * Suvir Kumar <suvir@talkkonnect.com>
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * channelscan.go implements the "Scan Channels" function. It walks the accessable
 * channels of the connected Mumble server one by one, dwelling on each for a
 * configurable time, and holds on any channel where voice traffic is heard until
 * the traffic stops (plus a hang time), then carries on scanning.
 *
 */

package talkkonnect

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	scanDefaultDwellMsecs = 3000
	scanDefaultHangMsecs  = 4000
	scanMinDwellMsecs     = 500
	scanMinHangMsecs      = 500
	scanPollInterval      = 100 * time.Millisecond
)

type scanChannelStruct struct {
	chanID   uint32
	chanName string
}

var (
	scanMu            sync.Mutex
	scanRunning       bool
	scanStopRequested bool
	scanReturnHome    bool
	scanStopChan      chan struct{}

	// scanActive/scanHolding are read from the audio path and the status API, keep them lock free.
	scanActive  atomic.Bool
	scanHolding atomic.Bool

	// scanVoiceActivity maps a channel name to the unix nano timestamp of the last voice packet
	// heard on it. Only written while a scan is running.
	scanVoiceActivity sync.Map
)

// ScanIsRunning reports whether the channel scanner is currently walking channels.
func ScanIsRunning() bool {
	return scanActive.Load()
}

// ScanIsHolding reports whether the scanner is parked on a channel because of traffic.
func ScanIsHolding() bool {
	return scanHolding.Load()
}

// noteScanVoiceActivity is called from the incoming audio path (only while scanning) to
// remember when voice was last heard on a given channel.
func noteScanVoiceActivity(channelName string) {
	if channelName == "" {
		return
	}
	scanVoiceActivity.Store(channelName, time.Now().UnixNano())
}

// scanVoiceActiveOn reports whether voice was heard on channelName within the last "within" period.
func scanVoiceActiveOn(channelName string, within time.Duration) bool {
	value, found := scanVoiceActivity.Load(channelName)
	if !found {
		return false
	}
	last, ok := value.(int64)
	if !ok {
		return false
	}
	return time.Since(time.Unix(0, last)) <= within
}

func scanDwellDuration() time.Duration {
	msecs := Config.Global.Software.ChannelScan.DwellTimeMsecs
	if msecs <= 0 {
		msecs = scanDefaultDwellMsecs
	}
	if msecs < scanMinDwellMsecs {
		msecs = scanMinDwellMsecs
	}
	return time.Duration(msecs) * time.Millisecond
}

func scanHangDuration() time.Duration {
	msecs := Config.Global.Software.ChannelScan.HangTimeMsecs
	if msecs <= 0 {
		msecs = scanDefaultHangMsecs
	}
	if msecs < scanMinHangMsecs {
		msecs = scanMinHangMsecs
	}
	return time.Duration(msecs) * time.Millisecond
}

// scanSkipped reports whether a channel is excluded from the scan by the
// comma separated skipchannels config item (channel name or channel ID).
func scanSkipped(channel scanChannelStruct) bool {
	for _, item := range strings.Split(Config.Global.Software.ChannelScan.SkipChannels, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.EqualFold(item, channel.chanName) {
			return true
		}
		if id, err := strconv.Atoi(item); err == nil && uint32(id) == channel.chanID {
			return true
		}
	}
	return false
}

// scanChannelList returns the accessable channels (in channel ID order) that take part in the scan.
func (b *Talkkonnect) scanChannelList() []scanChannelStruct {
	b.ListChannels(false)

	var list []scanChannelStruct
	for _, ch := range ChannelsList {
		chanName, found := AccessableChannelMap[ch.chanID]
		if !found {
			continue
		}
		candidate := scanChannelStruct{chanID: uint32(ch.chanID), chanName: chanName}
		if scanSkipped(candidate) {
			log.Printf("debug: Channel Scan Skipping Channel ID %v Name %v (Excluded by Config)\n", candidate.chanID, candidate.chanName)
			continue
		}
		list = append(list, candidate)
	}
	return list
}

// Scan toggles channel scanning on and off.
func (b *Talkkonnect) Scan() {
	if !(IsConnected) {
		sshRemoteReplyF("Not connected to Mumble; scan unavailable.\n")
		return
	}

	scanMu.Lock()
	running := scanRunning
	scanMu.Unlock()

	if running {
		b.ScanStop("requested by user")
		return
	}

	b.ScanStart()
}

// ScanStart begins scanning the accessable channels in a background goroutine.
func (b *Talkkonnect) ScanStart() {
	if !(IsConnected) {
		sshRemoteReplyF("Not connected to Mumble; scan unavailable.\n")
		return
	}

	if b.Client == nil || b.Client.Self == nil || b.Client.Self.Channel == nil {
		log.Println("warn: Channel Scan Not Possible, Client Not Ready")
		sshRemoteReplyF("Channel scan not possible, client not ready.\n")
		return
	}

	channels := b.scanChannelList()
	if len(channels) < 2 {
		log.Printf("warn: Channel Scan Needs At Least 2 Accessable Channels, Found %v\n", len(channels))
		sshRemoteReplyF("Channel scan needs at least 2 accessable channels (found %v).\n", len(channels))
		return
	}

	scanMu.Lock()
	if scanRunning {
		scanMu.Unlock()
		log.Println("info: Channel Scan Already Running")
		return
	}
	stop := make(chan struct{})
	scanStopChan = stop
	scanStopRequested = false
	scanReturnHome = Config.Global.Software.ChannelScan.ReturnToStartChannel
	scanRunning = true
	scanMu.Unlock()

	scanActive.Store(true)
	scanHolding.Store(false)
	scanVoiceActivity.Range(func(key, _ interface{}) bool {
		scanVoiceActivity.Delete(key)
		return true
	})

	homeChannel := b.Client.Self.Channel.Name
	log.Printf("info: Channel Scan Started on %v Channels (Dwell %v Hang %v Home Channel %v)\n", len(channels), scanDwellDuration(), scanHangDuration(), homeChannel)
	sshRemoteReplyF("Channel scan started on %v channels (dwell %v, hang %v).\n", len(channels), scanDwellDuration(), scanHangDuration())
	TTSEvent("startscanning")

	SafeGo(func() {
		b.scanLoop(channels, homeChannel, stop)
	})
}

// ScanStop asks a running scan to stop, the scanning goroutine does the tidying up.
func (b *Talkkonnect) ScanStop(reason string) {
	scanMu.Lock()
	if !scanRunning || scanStopRequested {
		scanMu.Unlock()
		return
	}
	scanStopRequested = true
	if scanStopChan != nil {
		close(scanStopChan)
	}
	scanMu.Unlock()

	log.Printf("info: Channel Scan Stopping (%v)\n", reason)
}

// ScanStopForChannelChange stops a running scan when the channel is changed by hand
// (menu, GPIO, keyboard or remote API) and leaves the client on the newly chosen channel
// instead of jumping back to the channel the scan started from.
func (b *Talkkonnect) ScanStopForChannelChange() {
	scanMu.Lock()
	if !scanRunning {
		scanMu.Unlock()
		return
	}
	scanReturnHome = false
	scanMu.Unlock()

	b.ScanStop("manual channel change")
}

func (b *Talkkonnect) scanLoop(channels []scanChannelStruct, homeChannel string, stop <-chan struct{}) {
	defer func() {
		scanActive.Store(false)
		scanHolding.Store(false)

		scanMu.Lock()
		returnHome := scanReturnHome
		scanRunning = false
		scanStopRequested = false
		scanReturnHome = false
		scanStopChan = nil
		scanMu.Unlock()

		if returnHome && IsConnected && homeChannel != "" {
			log.Printf("info: Channel Scan Returning to Home Channel %v\n", homeChannel)
			b.ChangeChannel(homeChannel)
		}

		currentChannel := homeChannel
		if !returnHome && b.Client != nil && b.Client.Self != nil && b.Client.Self.Channel != nil {
			// when returning home the server has not confirmed the move yet, so only trust
			// the client channel when we are staying where the scan left us
			currentChannel = b.Client.Self.Channel.Name
		}

		log.Printf("info: Channel Scan Stopped on Channel %v\n", currentChannel)
		sshRemoteReplyF("Channel scan stopped on channel %v.\n", currentChannel)
		b.scanDisplay("Scan Stopped", currentChannel)
		TTSEvent("stopscanning")
	}()

	index := 0
	if b.Client != nil && b.Client.Self != nil && b.Client.Self.Channel != nil {
		for i, ch := range channels {
			if ch.chanID == b.Client.Self.Channel.ID {
				index = i
				break
			}
		}
	}

	for {
		select {
		case <-stop:
			return
		default:
		}

		if !(IsConnected) {
			log.Println("alert: Channel Scan Aborted, Disconnected From Server")
			return
		}

		index = (index + 1) % len(channels)

		if !b.scanMoveTo(channels[index]) {
			// channel disappeared from the server, take a short breath before trying the next one
			if !scanSleep(stop, scanPollInterval) {
				return
			}
			continue
		}

		if !b.scanDwellOn(channels[index], stop) {
			return
		}
	}
}

// scanMoveTo moves into the channel to be scanned, it returns false when the channel is gone.
func (b *Talkkonnect) scanMoveTo(scanChannel scanChannelStruct) bool {
	if b.Client == nil || b.Client.Self == nil {
		return false
	}

	channel := b.Client.Channels[scanChannel.chanID]
	if channel == nil {
		log.Printf("warn: Channel Scan Cannot Find Channel ID %v Name %v Anymore\n", scanChannel.chanID, scanChannel.chanName)
		return false
	}

	if b.Client.Self.Channel == nil || b.Client.Self.Channel.ID != scanChannel.chanID {
		b.Client.Self.Move(channel)
	}

	b.BackLightTimer()
	log.Printf("info: Channel Scan Now on Channel ID %v Name %v\n", scanChannel.chanID, scanChannel.chanName)
	b.scanDisplay("Scanning", scanChannel.chanName)
	return true
}

// scanDwellOn stays on a channel for the dwell time and holds on it for as long as there is
// traffic (or we are transmitting) plus the hang time. It returns false when the scan should stop.
func (b *Talkkonnect) scanDwellOn(scanChannel scanChannelStruct, stop <-chan struct{}) bool {
	dwell := scanDwellDuration()
	hang := scanHangDuration()
	deadline := time.Now().Add(dwell)

	ticker := time.NewTicker(scanPollInterval)
	defer ticker.Stop()

	holding := false

	for {
		select {
		case <-stop:
			return false
		case <-ticker.C:
			if !(IsConnected) {
				log.Println("alert: Channel Scan Aborted, Disconnected From Server")
				return false
			}

			busy := b.IsTransmitting || scanVoiceActiveOn(scanChannel.chanName, hang)

			if busy {
				if !holding {
					holding = true
					scanHolding.Store(true)
					b.BackLightTimer()
					log.Printf("info: Channel Scan Holding on Channel ID %v Name %v (Activity Detected)\n", scanChannel.chanID, scanChannel.chanName)
					sshRemoteReplyF("Channel scan holding on channel %v (activity detected).\n", scanChannel.chanName)
					b.scanDisplay("Scan Hold", scanChannel.chanName)
				}
				continue
			}

			if holding {
				// traffic finished and the hang time expired, carry on with the next channel
				holding = false
				scanHolding.Store(false)
				log.Printf("info: Channel Scan Resuming After Hang Time on Channel %v\n", scanChannel.chanName)
				return true
			}

			if time.Now().After(deadline) {
				return true
			}
		}
	}
}

// scanSleep waits for the given duration, it returns false when the scan was asked to stop.
func scanSleep(stop <-chan struct{}, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-stop:
		return false
	case <-timer.C:
		return true
	}
}

func (b *Talkkonnect) scanDisplay(status string, channelName string) {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return
	}
	if LCDEnabled {
		LcdText[1] = status + " " + channelName
		LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
	}
	if OLEDEnabled {
		oledDisplay(false, 1, OLEDStartColumn, status+" "+channelName)
	}
}
