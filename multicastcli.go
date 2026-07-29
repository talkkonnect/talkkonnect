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
 */

package talkkonnect

// The "mc" bottom CLI / SSH console command inspects and switches RTP multicast
// output. It calls the sender directly rather than going through the HTTP command
// allow-list, so it works whether or not multicaston / multicastoff are enabled
// under <http> in the XML config.

import (
	"fmt"
	"io"
	"strings"
)

const multicastCLIUsage = `Usage:
  mc status                 show the multicast destination, filters and live state
  mc on                     start sending received audio to the multicast group
  mc off                    stop sending
  mc toggle                 flip between the two
Notes:
  The group, codec, ttl, interface and packet size come from the <multicast>
  section of talkkonnect.xml; change them with "cfg set global.software.multicast.…"
  followed by "cfg save" and a config reload, which restarts the sender.
  Audio goes out as 8 kHz mono RTP, 20 ms per packet by default, which is what
  hardware IP speakers and SIP desk phones decode.`

// multicastCLISubcommands are the forms Tab completes.
var multicastCLISubcommands = []string{"help", "off", "on", "status", "toggle"}

// bottomCLIHandleMulticastLine runs one "mc ..." line.
func (b *Talkkonnect) bottomCLIHandleMulticastLine(w io.Writer, line string) {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) == 0 || !strings.EqualFold(parts[0], "mc") {
		return
	}

	sub := "status"
	if len(parts) > 1 {
		sub = strings.ToLower(parts[1])
	}

	switch sub {
	case "status", "show", "s":
		b.multicastCLIStatus(w)
	case "on", "start":
		b.cmdMulticastOn()
		b.multicastCLIStatus(w)
	case "off", "stop":
		b.cmdMulticastOff()
		b.multicastCLIStatus(w)
	case "toggle", "t":
		b.cmdMulticastToggle()
		b.multicastCLIStatus(w)
	case "help", "h", "?":
		fmt.Fprintln(w, multicastCLIUsage)
	default:
		fmt.Fprintf(w, "Unknown mc subcommand %q.\n%s\n", parts[1], multicastCLIUsage)
	}
}

// multicastCLIStatus prints where audio is going and who is being carried.
func (b *Talkkonnect) multicastCLIStatus(w io.Writer) {
	status := multicastUISnapshot()
	cfg := multicastConfigSnapshot()

	state := "stopped"
	if status.Running {
		state = "running"
	}
	fmt.Fprintf(w, "Multicast %s (config enabled=%v)\n", state, status.Enabled)

	if status.Group == "" {
		fmt.Fprintln(w, "  no group configured, set global.software.multicast.group")
	} else {
		fmt.Fprintf(w, "  destination %s codec %s ttl %d packet %dms volume %d%%\n",
			status.Group, status.Codec, status.TTL, status.PacketMS, status.Volume)
	}
	if cfg.Interface == "" {
		fmt.Fprintln(w, "  interface   default route")
	} else {
		fmt.Fprintf(w, "  interface   %s\n", cfg.Interface)
	}

	scope := "the joined channel only"
	if cfg.AllChannels {
		scope = "every channel talkkonnect hears"
	}
	fmt.Fprintf(w, "  carrying    %s\n", scope)

	if len(cfg.Include) == 0 {
		fmt.Fprintln(w, "  include     every talker")
	} else {
		fmt.Fprintf(w, "  include     %s\n", strings.Join(multicastSortedNames(cfg.Include), ", "))
	}
	if len(cfg.Exclude) > 0 {
		fmt.Fprintf(w, "  exclude     %s\n", strings.Join(multicastSortedNames(cfg.Exclude), ", "))
	}

	if len(status.Sources) == 0 {
		fmt.Fprintf(w, "  live        idle, %d packet(s) sent\n", status.Packets)
	} else {
		fmt.Fprintf(w, "  live        mixing %s, %d packet(s) sent\n", strings.Join(status.Sources, ", "), status.Packets)
	}
}

// bottomCLITabCompleteMC completes "mc" lines: the verb, and nothing after it
// since no subcommand takes an argument.
func bottomCLITabCompleteMC(line string) (newLine string, bell bool) {
	trailingSpace := len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t')
	lineTR := strings.TrimRight(line, " \t")
	fields := strings.Fields(lineTR)

	var matches []string
	switch {
	case len(fields) == 1 && !trailingSpace:
		if !strings.HasPrefix("mc", strings.ToLower(fields[0])) {
			return line, true
		}
		return "mc ", false
	case len(fields) == 1 && trailingSpace:
		matches = multicastCLISubcommands
	case len(fields) == 2 && !trailingSpace:
		for _, candidate := range multicastCLISubcommands {
			if strings.HasPrefix(candidate, strings.ToLower(fields[1])) {
				matches = append(matches, candidate)
			}
		}
	default:
		return line, true
	}

	if len(matches) == 0 {
		return line, true
	}
	if len(matches) == 1 {
		return "mc " + matches[0] + " ", false
	}
	lcp := bottomCLILongestCommonPrefix(matches)
	typed := ""
	if !trailingSpace && len(fields) == 2 {
		typed = fields[1]
	}
	if len(lcp) <= len(typed) {
		return line, true
	}
	return "mc " + lcp, false
}
