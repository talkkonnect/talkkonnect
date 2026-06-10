/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package talkkonnect

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/volume-go"
)

// UIChannelUser is one user in the current Mumble channel.
type UIChannelUser struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Self   bool   `json:"self"`
}

// UIChannelNode is one row in the server channel tree for framebuffer clients.
type UIChannelNode struct {
	Name       string `json:"name"`
	Depth      int    `json:"depth"`
	UserCount  int    `json:"userCount"`
	Active     bool   `json:"active"`
	Accessible bool   `json:"accessible"`
}

// UILastMessage is the most recent Mumble text message for framebuffer clients.
type UILastMessage struct {
	Sender string `json:"sender,omitempty"`
	Text   string `json:"text,omitempty"`
}

var (
	lastUIMessageSender string
	lastUIMessageText   string
)

// RecordUILastMessage stores the latest received Mumble text message for /uistatus.
func RecordUILastMessage(sender, text string) {
	lastUIMessageSender = strings.TrimSpace(sender)
	lastUIMessageText = strings.TrimSpace(text)
}

// ClearUILastMessage removes the stored Mumble text message from /uistatus.
func ClearUILastMessage() {
	lastUIMessageSender = ""
	lastUIMessageText = ""
}

func lastUIMessageSnapshot() UILastMessage {
	return UILastMessage{
		Sender: lastUIMessageSender,
		Text:   lastUIMessageText,
	}
}

// UIStatus is JSON telemetry for external framebuffer / dashboard clients.
type UIStatus struct {
	Connected      bool                `json:"connected"`
	Transmitting   bool                `json:"transmitting"`
	ServerName     string              `json:"serverName"`
	Server         string              `json:"server"`
	Channel        string              `json:"channel"`
	UsersOnline    int                 `json:"usersOnline"`
	ChannelUsers   []UIChannelUser     `json:"channelUsers"`
	ChannelTree    []UIChannelNode     `json:"channelTree"`
	Receiving      bool                `json:"receiving"`
	LastSpeaker    string              `json:"lastSpeaker"`
	LastMessage    UILastMessage       `json:"lastMessage,omitempty"`
	RXVolume       int                 `json:"rxVolume"`
	Muted          bool                `json:"muted"`
	InternetRadio  InternetRadioStatus `json:"internetRadio"`
	IPAddress      string              `json:"ipAddress"`
	Bitrate        string              `json:"bitrate"`
	UptimeSec      int64               `json:"uptimeSec"`
	Activity       string              `json:"activity"`
	MumbleUsername string              `json:"mumbleUsername"`
	Version        string              `json:"version"`
}

func primaryLocalIPv4() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip != nil {
				return ip.String()
			}
		}
	}
	return ""
}

func mumbleUserStatus(name, lastSpeaker string, muted, selfMuted, suppressed bool) string {
	if lastSpeaker != "" && strings.EqualFold(name, lastSpeaker) {
		return "Speaking"
	}
	if muted || selfMuted || suppressed {
		return "Muted"
	}
	return "idle"
}

func (b *Talkkonnect) channelUsersSnapshot() []UIChannelUser {
	if b == nil || b.Client == nil || b.Client.Self == nil || b.Client.Self.Channel == nil {
		return nil
	}

	selfName := b.Client.Self.Name
	activeSpeaker := ""
	if ReceivingVoice {
		activeSpeaker = LastSpeaker
	}
	var out []UIChannelUser
	for _, usr := range b.Client.Self.Channel.Users {
		if usr == nil || strings.TrimSpace(usr.Name) == "" {
			continue
		}
		out = append(out, UIChannelUser{
			Name:   usr.Name,
			Status: mumbleUserStatus(usr.Name, activeSpeaker, usr.Muted, usr.SelfMuted, usr.Suppressed),
			Self:   strings.EqualFold(usr.Name, selfName),
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Status == "Speaking" && out[j].Status != "Speaking" {
			return true
		}
		if out[j].Status == "Speaking" && out[i].Status != "Speaking" {
			return false
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})

	return out
}

func channelAccessibleForUI(ch *gumble.Channel) bool {
	if ch == nil {
		return false
	}
	if perm := ch.Permission(); perm != nil {
		return perm.Has(gumble.PermissionEnter)
	}
	return true
}

func sortedChannelChildren(ch *gumble.Channel) []*gumble.Channel {
	if ch == nil || len(ch.Children) == 0 {
		return nil
	}
	out := make([]*gumble.Channel, 0, len(ch.Children))
	for _, c := range ch.Children {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Position != out[j].Position {
			return out[i].Position < out[j].Position
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func (b *Talkkonnect) appendChannelTree(out *[]UIChannelNode, ch *gumble.Channel, depth int, activeID uint32) {
	if ch == nil {
		return
	}
	*out = append(*out, UIChannelNode{
		Name:       ch.Name,
		Depth:      depth,
		UserCount:  len(ch.Users),
		Active:     ch.ID == activeID,
		Accessible: channelAccessibleForUI(ch),
	})
	for _, child := range sortedChannelChildren(ch) {
		b.appendChannelTree(out, child, depth+1, activeID)
	}
}

func (b *Talkkonnect) channelTreeSnapshot() []UIChannelNode {
	if b == nil || b.Client == nil || b.Client.Self == nil {
		return nil
	}
	active := b.Client.Self.Channel
	var activeID uint32
	if active != nil {
		activeID = active.ID
	}
	if active != nil && RootChannel == nil {
		return []UIChannelNode{{
			Name:       active.Name,
			Depth:      0,
			UserCount:  len(active.Users),
			Active:     true,
			Accessible: channelAccessibleForUI(active),
		}}
	}
	if RootChannel == nil {
		return nil
	}
	var out []UIChannelNode
	b.appendChannelTree(&out, RootChannel, 0, activeID)
	return out
}

func (b *Talkkonnect) buildUIStatus() UIStatus {
	st := UIStatus{
		Connected:     IsConnected,
		Transmitting:  b != nil && b.IsTransmitting,
		ServerName:    strings.TrimSpace(b.Name),
		Server:        b.Address,
		LastSpeaker:   LastSpeaker,
		LastMessage:   lastUIMessageSnapshot(),
		Receiving:     ReceivingVoice,
		InternetRadio: InternetRadioStatusSnapshot(),
		IPAddress:     primaryLocalIPv4(),
		UptimeSec:     int64(time.Since(StartTime).Seconds()),
		Version:       talkkonnectVersion,
	}

	if vol, err := volume.GetVolume(Config.Global.Software.Settings.OutputVolControlDevice); err == nil {
		st.RXVolume = vol
	}

	if b != nil && b.Config != nil && strings.TrimSpace(b.Config.Username) != "" {
		st.MumbleUsername = b.Config.Username
	} else if AccountIndex >= 0 && AccountIndex < len(Username) {
		st.MumbleUsername = strings.TrimSpace(Username[AccountIndex])
	}

	if b != nil && b.Client != nil && b.Client.Self != nil && b.Client.Self.Channel != nil {
		st.Channel = b.Client.Self.Channel.Name
		st.ChannelUsers = b.channelUsersSnapshot()
		st.UsersOnline = len(st.ChannelUsers)
		st.ChannelTree = b.channelTreeSnapshot()
	}
	if muted, err := volume.GetMuted(Config.Global.Software.Settings.OutputDevice); err == nil {
		st.Muted = muted
	}

	switch {
	case st.Transmitting:
		st.Activity = "tx"
	case ReceivingVoice:
		st.Activity = "rx"
	case st.InternetRadio.Playing:
		st.Activity = "radio"
	case st.Connected:
		st.Activity = "idle"
	default:
		st.Activity = "offline"
	}

	if st.InternetRadio.Playing {
		st.Bitrate = "stream"
	}

	return st
}

func (b *Talkkonnect) httpUIStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !remoteControlHTTPClientIPAllowed(r) {
		http.Error(w, "403 forbidden: client address not allowed by remote control network ACL", http.StatusForbidden)
		return
	}

	st := b.buildUIStatus()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(st)
}

// UIStatusURL returns the local HTTP URL for framebuffer clients.
func UIStatusURL(listenPort string) string {
	port := strings.TrimSpace(listenPort)
	if port == "" {
		port = "8080"
	}
	return "http://127.0.0.1:" + port + "/uistatus"
}
