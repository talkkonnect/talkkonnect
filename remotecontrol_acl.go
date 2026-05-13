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
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	remoteControlACLEnabled bool
	remoteControlACLNets    []*net.IPNet
	remoteControlACLMu      sync.RWMutex
)

// reloadRemoteControlNetworkACLFromConfig parses <networkacl> under <remotecontrol> and
// refreshes the in-memory CIDR list used by the HTTP API and SSH remote console.
func reloadRemoteControlNetworkACLFromConfig() {
	cfg := Config.Global.Software.RemoteControl.NetworkACL
	var nets []*net.IPNet
	for _, n := range cfg.Network {
		cidr := strings.TrimSpace(n.CIDR)
		if cidr == "" {
			continue
		}
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			log.Printf("warn: Remote control network ACL: invalid CIDR %q: %v (skipped)", cidr, err)
			continue
		}
		nets = append(nets, ipnet)
	}
	active := cfg.Enabled && len(nets) > 0
	if cfg.Enabled && len(nets) == 0 {
		log.Println("warn: Remote control network ACL enabled but no valid CIDR entries; ACL inactive (all addresses allowed)")
	}

	remoteControlACLMu.Lock()
	remoteControlACLEnabled = active
	remoteControlACLNets = nets
	remoteControlACLMu.Unlock()

	if remoteControlACLEnabled {
		log.Printf("info: Remote control network ACL active with %d CIDR rule(s) (HTTP API and SSH console)\n", len(nets))
	}
}

func remoteControlIPAllowed(ip net.IP) bool {
	if ip == nil {
		return false
	}
	remoteControlACLMu.RLock()
	defer remoteControlACLMu.RUnlock()
	if !remoteControlACLEnabled {
		return true
	}
	for _, n := range remoteControlACLNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// remoteControlHTTPClientIPAllowed returns true if the request client may use remote control HTTP endpoints.
func remoteControlHTTPClientIPAllowed(r *http.Request) bool {
	if r == nil {
		return false
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return remoteControlIPAllowed(net.ParseIP(host))
}

// remoteControlTCPPeerAllowed returns true if the TCP peer may use the SSH remote console.
func remoteControlTCPPeerAllowed(addr net.Addr) bool {
	switch a := addr.(type) {
	case *net.TCPAddr:
		return remoteControlIPAllowed(a.IP)
	default:
		return false
	}
}
