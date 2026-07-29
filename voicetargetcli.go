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

// The "vt" bottom CLI / SSH console command lists the voice targets from the XML
// config and appends new ones to it. Adding writes the config file and updates
// the in-memory Config, so "voicetargetset <id>" can use a target right away
// without a restart.

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// voiceTargetMaxID is Mumble's highest usable voice target slot.
const voiceTargetMaxID uint32 = 31

const voiceTargetCLIUsage = `Usage:
  vt list                                              show the voice targets in the XML config
  vt set <id>                                          activate a configured target (same as voicetargetset)
  vt clear                                             back to normal channel speech (same as vt set 0)
  vt next | vt prev                                    step through the configured targets
  vt whisper <name>                                    whisper to one online user, no config needed
  vt add <id> user <name> [<name> ...]                 whisper target: one or more users
  vt add <id> channel <name> [recursive=true] [links=true] [group=<name>]
                                                       shout target: a channel
Notes:
  <id> is 1 to 31. Target 0 means normal channel speech and cannot be configured.
  Channel and user names with spaces go in double quotes.
  Adding appends to the account talkkonnect is using, writes the XML config
  (previous contents kept as <config>.bak) and takes effect immediately.
  "vt set" and "vt clear" always work; the equivalent voicetargetset command needs
  a matching <command action="voicetargetset"> under <http> in the XML config.`

// voiceTargetConfigAccountIndex returns the index into Config.Accounts.Account of
// the account whose <voicetargets> cmdSendVoiceTargets actually reads: the
// AccountIndex-th account marked default. Returns -1 when there is no such
// account, so "vt add" refuses rather than writing somewhere unused.
func voiceTargetConfigAccountIndex() int {
	counter := 0
	for i, account := range Config.Accounts.Account {
		if !account.Default {
			continue
		}
		if counter == AccountIndex {
			return i
		}
		counter++
	}
	return -1
}

// voiceTargetCLITokens splits a command line on whitespace, keeping double quoted
// runs together so user and channel names may contain spaces.
func voiceTargetCLITokens(line string) []string {
	var out []string
	var cur strings.Builder
	inQuote := false
	started := false

	flush := func() {
		if started {
			out = append(out, cur.String())
			cur.Reset()
			started = false
		}
	}

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuote = !inQuote
			started = true
		case (c == ' ' || c == '\t') && !inQuote:
			flush()
		default:
			cur.WriteByte(c)
			started = true
		}
	}
	flush()
	return out
}

// voiceTargetCLIParseBool accepts the usual spellings, defaulting to def when the
// value is empty.
func voiceTargetCLIParseBool(s string, def bool) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return def, nil
	}
	v, err := strconv.ParseBool(strings.ToLower(s))
	if err != nil {
		return false, fmt.Errorf("%q is not true or false", s)
	}
	return v, nil
}

// voiceTargetCLISubcommands are the forms Tab completes, one spelling each so a
// completion never has to pick between an alias and its long name.
var voiceTargetCLISubcommands = []string{"add", "clear", "help", "list", "next", "prev", "set", "whisper"}

// bottomCLITabCompleteVT completes "vt" lines: the subcommand, then the ids that
// are actually configured for "vt set" and the users who are actually online for
// "vt whisper", so the caller does not have to run "vt list" first.
func bottomCLITabCompleteVT(line string) (newLine string, bell bool) {
	trailingSpace := len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t')
	lineTR := strings.TrimRight(line, " \t")
	fields := strings.Fields(lineTR)

	filter := func(cands []string, prefix string) []string {
		var out []string
		for _, c := range cands {
			if strings.HasPrefix(strings.ToLower(c), strings.ToLower(prefix)) {
				out = append(out, c)
			}
		}
		return out
	}

	var matches []string
	switch {
	case len(fields) == 1 && !trailingSpace:
		if !strings.HasPrefix("vt", strings.ToLower(fields[0])) {
			return line, true
		}
		return "vt ", false
	case len(fields) == 1 && trailingSpace:
		matches = voiceTargetCLISubcommands
	case len(fields) == 2 && !trailingSpace:
		matches = filter(voiceTargetCLISubcommands, fields[1])
	case len(fields) == 2 && trailingSpace, len(fields) == 3 && !trailingSpace:
		arg := ""
		if len(fields) == 3 {
			arg = fields[2]
		}
		switch strings.ToLower(fields[1]) {
		case "set", "s":
			matches = filter(voiceTargetConfiguredIDs(), arg)
		case "whisper", "wh":
			matches = filter(voiceTargetOnlineUserNames(), arg)
		default:
			return line, true
		}
	default:
		return line, true
	}

	if len(matches) == 0 {
		return line, true
	}

	// Rebuild from the fields that are already complete so the caret keeps the
	// spacing the user typed for them.
	prefixFields := fields
	if !trailingSpace {
		prefixFields = fields[:len(fields)-1]
	}
	pre := strings.Join(prefixFields, " ") + " "

	if len(matches) == 1 {
		return pre + matches[0] + " ", false
	}
	lcp := bottomCLILongestCommonPrefix(matches)
	typed := ""
	if !trailingSpace {
		typed = fields[len(fields)-1]
	}
	if len(lcp) <= len(typed) {
		return line, true
	}
	return pre + lcp, false
}

// voiceTargetConfiguredIDs lists the ids selectable right now, 0 included because
// clearing is a selection too.
func voiceTargetConfiguredIDs() []string {
	out := []string{"0"}
	accountIndex := voiceTargetConfigAccountIndex()
	if accountIndex < 0 {
		return out
	}
	for _, target := range Config.Accounts.Account[accountIndex].Voicetargets.ID {
		out = append(out, strconv.FormatUint(uint64(target.Value), 10))
	}
	return out
}

// voiceTargetOnlineUserNames lists the other users on the server, since whispering
// to yourself is not useful. Names with spaces are quoted so the completed line
// tokenizes back to one name.
func voiceTargetOnlineUserNames() []string {
	b := appTalkkonnect
	if !IsConnected || b == nil || b.Client == nil {
		return nil
	}
	var out []string
	for _, user := range b.Client.Users {
		if user == nil || user.Name == "" || (b.Client.Self != nil && user.Name == b.Client.Self.Name) {
			continue
		}
		if strings.ContainsAny(user.Name, " \t") {
			out = append(out, `"`+user.Name+`"`)
			continue
		}
		out = append(out, user.Name)
	}
	return out
}

// bottomCLIHandleVoiceTargetLine runs one "vt ..." line.
func (b *Talkkonnect) bottomCLIHandleVoiceTargetLine(w io.Writer, line string) {
	parts := voiceTargetCLITokens(strings.TrimSpace(line))
	if len(parts) == 0 || !strings.EqualFold(parts[0], "vt") {
		return
	}

	sub := "list"
	if len(parts) > 1 {
		sub = strings.ToLower(parts[1])
	}

	switch sub {
	case "list", "show", "l":
		b.voiceTargetCLIList(w)
	case "set", "s":
		b.voiceTargetCLISet(w, parts[2:])
	case "clear", "off", "none":
		b.voiceTargetCLISet(w, []string{"0"})
	case "next", "prev", "previous":
		b.voiceTargetCLIStep(w, sub)
	case "whisper", "wh":
		b.voiceTargetCLIWhisper(w, parts[2:])
	case "add", "a":
		b.voiceTargetCLIAdd(w, parts[2:])
	case "help", "h", "?":
		fmt.Fprintln(w, voiceTargetCLIUsage)
	default:
		fmt.Fprintf(w, "Unknown vt subcommand %q.\n%s\n", parts[1], voiceTargetCLIUsage)
	}
}

// voiceTargetCLISet activates a target by id, or clears with id 0. It goes
// straight to cmdSendVoiceTargets rather than through HandleRemoteAPICommand, so
// it works regardless of which commands the XML enables under <http>.
func (b *Talkkonnect) voiceTargetCLISet(w io.Writer, args []string) {
	if len(args) == 0 {
		fmt.Fprintf(w, "vt set: give a target id, 0 to %d\n", voiceTargetMaxID)
		return
	}

	id, err := strconv.ParseUint(strings.TrimSpace(args[0]), 10, 32)
	if err != nil || uint32(id) > voiceTargetMaxID {
		fmt.Fprintf(w, "vt set: %q is not a voice target id, expected 0 to %d\n", args[0], voiceTargetMaxID)
		return
	}
	targetID := uint32(id)

	if !IsConnected || b.Client == nil {
		fmt.Fprintln(w, "Not connected to Mumble; cannot change the voice target.")
		return
	}

	// Warn instead of leaving the caller guessing why nothing changed: an id with
	// no <voicetargets> entry for the account in use is a no-op inside
	// cmdSendVoiceTargets, which only logs at debug level.
	if targetID > 0 && !voiceTargetConfigured(targetID) {
		fmt.Fprintf(w, "vt set: no voice target %d is configured for the account in use, see \"vt list\"\n", targetID)
		return
	}

	b.cmdSendVoiceTargets(targetID)

	if active := uiVoiceTargetSnapshot(); active.ID == 0 {
		fmt.Fprintln(w, "Voice target cleared, talking normally to the joined channel.")
	} else {
		fmt.Fprintf(w, "Voice target %d active (%s %s).\n", active.ID, active.Kind, strings.Join(active.Names, ", "))
	}
}

// voiceTargetConfigured reports whether the account cmdSendVoiceTargets reads has
// an entry for targetID.
func voiceTargetConfigured(targetID uint32) bool {
	accountIndex := voiceTargetConfigAccountIndex()
	if accountIndex < 0 {
		return false
	}
	for _, target := range Config.Accounts.Account[accountIndex].Voicetargets.ID {
		if target.Value == targetID {
			return true
		}
	}
	return false
}

// voiceTargetCLIStep walks the configured targets the way the rotary encoder does.
func (b *Talkkonnect) voiceTargetCLIStep(w io.Writer, sub string) {
	if !IsConnected || b.Client == nil {
		fmt.Fprintln(w, "Not connected to Mumble; cannot change the voice target.")
		return
	}
	// VTMove indexes Config.Accounts.Account[AccountIndex] directly, so check the
	// bounds it assumes before calling it.
	if AccountIndex < 0 || AccountIndex >= len(Config.Accounts.Account) ||
		len(Config.Accounts.Account[AccountIndex].Voicetargets.ID) == 0 {
		fmt.Fprintln(w, "No voice targets are configured for the account in use, see \"vt list\".")
		return
	}

	direction := "up"
	if sub != "next" {
		direction = "down"
	}
	b.VTMove(direction)

	if active := uiVoiceTargetSnapshot(); active.ID == 0 {
		fmt.Fprintln(w, "No voice target active, talking normally to the joined channel.")
	} else {
		fmt.Fprintf(w, "Voice target %d active (%s %s).\n", active.ID, active.Kind, strings.Join(active.Names, ", "))
	}
}

// voiceTargetCLIWhisper whispers to a user named on the command line, which is the
// one form that needs no <voicetargets> entry at all.
func (b *Talkkonnect) voiceTargetCLIWhisper(w io.Writer, args []string) {
	name := strings.TrimSpace(strings.Join(args, " "))
	if name == "" {
		fmt.Fprintln(w, "vt whisper: give the name of an online user")
		return
	}
	if !IsConnected || b.Client == nil {
		fmt.Fprintln(w, "Not connected to Mumble; cannot set a whisper target.")
		return
	}
	if b.Client.Users.Find(name) == nil {
		fmt.Fprintf(w, "vt whisper: %q is not on the server right now, whisper target unchanged\n", name)
		return
	}

	b.cmdWhisperUser(name)
	// cmdWhisperUser reports through sshRemoteReplyF, which the local bottom CLI
	// does not attach to, so echo the outcome on the caller's writer as well.
	fmt.Fprintf(w, "Whispering to %s on voice target %d.\n", name, remoteAPIWhisperTargetID)
}

// voiceTargetCLIList prints every configured target, flagging the ones whose user
// or channel is not on the server right now — a target naming a channel that does
// not exist silently falls back to plain channel speech when it is selected.
func (b *Talkkonnect) voiceTargetCLIList(w io.Writer) {
	if len(Config.Accounts.Account) == 0 {
		fmt.Fprintln(w, "No accounts are configured.")
		return
	}

	activeIdx := voiceTargetConfigAccountIndex()
	connected := IsConnected && b.Client != nil

	for i, account := range Config.Accounts.Account {
		marker := ""
		if i == activeIdx {
			marker = "   <- in use"
		}
		fmt.Fprintf(w, "account %d %q default=%v%s\n", i, account.Name, account.Default, marker)

		if len(account.Voicetargets.ID) == 0 {
			fmt.Fprintln(w, "  no voice targets configured")
			continue
		}

		for _, target := range account.Voicetargets.ID {
			fmt.Fprintf(w, "  id %d\n", target.Value)
			for _, user := range target.Users.User {
				note := ""
				if connected && b.Client.Users.Find(user) == nil {
					note = "   (not on the server right now)"
				}
				fmt.Fprintf(w, "    user    %s%s\n", user, note)
			}
			for _, channel := range target.Channels.Channel {
				note := ""
				if connected && b.Client.Channels.Find(channel.Name) == nil {
					note = "   (no such channel on the server)"
				}
				fmt.Fprintf(w, "    channel %s recursive=%v links=%v group=%q%s\n",
					channel.Name, channel.Recursive, channel.Links, channel.Group, note)
			}
			if len(target.Users.User) == 0 && len(target.Channels.Channel) == 0 {
				fmt.Fprintln(w, "    empty, selecting this target does nothing")
			}
		}
	}

	if active := uiVoiceTargetSnapshot(); active.ID == 0 {
		fmt.Fprintln(w, "Active: none, talking normally to the joined channel.")
	} else {
		fmt.Fprintf(w, "Active: id %d (%s %s)\n", active.ID, active.Kind, strings.Join(active.Names, ", "))
	}
}

// voiceTargetCLIAdd parses the add form, writes the XML config and mirrors the
// change into the running Config.
func (b *Talkkonnect) voiceTargetCLIAdd(w io.Writer, args []string) {
	if len(args) < 2 {
		fmt.Fprintln(w, voiceTargetCLIUsage)
		return
	}

	id, err := strconv.ParseUint(strings.TrimSpace(args[0]), 10, 32)
	if err != nil {
		fmt.Fprintf(w, "vt add: %q is not a voice target id\n", args[0])
		return
	}
	targetID := uint32(id)
	if targetID == 0 || targetID > voiceTargetMaxID {
		fmt.Fprintf(w, "vt add: id must be 1 to %d (0 is normal channel speech)\n", voiceTargetMaxID)
		return
	}
	if targetID == remoteAPIWhisperTargetID {
		fmt.Fprintf(w, "vt add: warning, id %d is the slot ad-hoc whisper requests reuse, so a run time whisper will overwrite it\n", targetID)
	}

	accountIndex := voiceTargetConfigAccountIndex()
	if accountIndex < 0 {
		fmt.Fprintf(w, "vt add: no account number %d is marked default=\"true\", so a voice target added here would never be used\n", AccountIndex)
		return
	}

	var users []string
	var channels []voiceTargetChannelEntry

	switch strings.ToLower(args[1]) {
	case "user", "users", "u":
		for _, name := range args[2:] {
			if name = strings.TrimSpace(name); name != "" {
				users = append(users, name)
			}
		}
		if len(users) == 0 {
			fmt.Fprintln(w, "vt add: give at least one user name")
			return
		}
	case "channel", "chan", "c":
		if len(args) < 3 || strings.TrimSpace(args[2]) == "" {
			fmt.Fprintln(w, "vt add: give a channel name")
			return
		}
		entry := voiceTargetChannelEntry{Name: strings.TrimSpace(args[2])}
		for _, opt := range args[3:] {
			key, value, found := strings.Cut(opt, "=")
			if !found {
				fmt.Fprintf(w, "vt add: %q should be recursive=, links= or group=\n", opt)
				return
			}
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "recursive":
				if entry.Recursive, err = voiceTargetCLIParseBool(value, true); err != nil {
					fmt.Fprintf(w, "vt add: recursive %v\n", err)
					return
				}
			case "links":
				if entry.Links, err = voiceTargetCLIParseBool(value, true); err != nil {
					fmt.Fprintf(w, "vt add: links %v\n", err)
					return
				}
			case "group":
				entry.Group = strings.TrimSpace(value)
			default:
				fmt.Fprintf(w, "vt add: unknown option %q, expected recursive=, links= or group=\n", key)
				return
			}
		}
		channels = append(channels, entry)
	default:
		fmt.Fprintf(w, "vt add: expected \"user\" or \"channel\", got %q\n", args[1])
		return
	}

	if err := voiceTargetWriteToConfigFile(accountIndex, targetID, users, channels); err != nil {
		fmt.Fprintf(w, "vt add: %v\n", err)
		log.Printf("error: vt add: %v", err)
		return
	}

	voiceTargetApplyToRunningConfig(accountIndex, targetID, users, channels)

	for _, user := range users {
		fmt.Fprintf(w, "Added user %q to voice target %d of account %q.\n", user, targetID, Config.Accounts.Account[accountIndex].Name)
		if IsConnected && b.Client != nil && b.Client.Users.Find(user) == nil {
			fmt.Fprintf(w, "Note: %q is not on the server right now, so the target cannot be sent until they connect.\n", user)
		}
	}
	for _, channel := range channels {
		fmt.Fprintf(w, "Added channel %q (recursive=%v links=%v group=%q) to voice target %d of account %q.\n",
			channel.Name, channel.Recursive, channel.Links, channel.Group, targetID, Config.Accounts.Account[accountIndex].Name)
		if IsConnected && b.Client != nil && b.Client.Channels.Find(channel.Name) == nil {
			fmt.Fprintf(w, "Note: there is no channel %q on the server, so selecting this target will fall back to plain channel speech.\n", channel.Name)
		}
	}

	fmt.Fprintf(w, "Saved to %s (previous contents in %s.bak). Select it with \"voicetargetset %d\".\n", ConfigXMLFile, ConfigXMLFile, targetID)
	log.Printf("info: vt add wrote voice target %d to %q", targetID, ConfigXMLFile)
}

// voiceTargetWriteToConfigFile splices the entries into the config file. The new
// document has to parse and contain what was just added before it replaces the
// original, so a splice that lands in the wrong place cannot destroy a config.
func voiceTargetWriteToConfigFile(accountIndex int, targetID uint32, users []string, channels []voiceTargetChannelEntry) error {
	path := strings.TrimSpace(ConfigXMLFile)
	if path == "" {
		return fmt.Errorf("config file path is empty")
	}

	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	updated, err := voiceTargetXMLAdd(string(original), accountIndex, targetID, users, channels)
	if err != nil {
		return err
	}

	if err := voiceTargetVerifyDocument(updated, accountIndex, targetID, users, channels); err != nil {
		return err
	}

	if err := os.WriteFile(path+".bak", original, 0644); err != nil {
		return fmt.Errorf("cannot write backup %s.bak: %v", path, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(updated), 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// voiceTargetVerifyDocument re-parses an edited config and confirms the entries
// really are where they were meant to go.
func voiceTargetVerifyDocument(doc string, accountIndex int, targetID uint32, users []string, channels []voiceTargetChannelEntry) error {
	var parsed ConfigStruct
	if err := xml.Unmarshal([]byte(doc), &parsed); err != nil {
		return fmt.Errorf("the edited config does not parse, nothing was written: %v", err)
	}
	if accountIndex >= len(parsed.Accounts.Account) {
		return fmt.Errorf("the edited config has no account number %d, nothing was written", accountIndex)
	}

	for _, target := range parsed.Accounts.Account[accountIndex].Voicetargets.ID {
		if target.Value != targetID {
			continue
		}
		for _, want := range users {
			found := false
			for _, got := range target.Users.User {
				if got == want {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("user %q is missing from voice target %d after the edit, nothing was written", want, targetID)
			}
		}
		for _, want := range channels {
			found := false
			for _, got := range target.Channels.Channel {
				if got.Name == want.Name && got.Recursive == want.Recursive && got.Links == want.Links {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("channel %q is missing from voice target %d after the edit, nothing was written", want.Name, targetID)
			}
		}
		return nil
	}

	return fmt.Errorf("voice target %d is missing from account number %d after the edit, nothing was written", targetID, accountIndex)
}

// voiceTargetApplyToRunningConfig mirrors the file change into the live Config so
// the new target works without a restart. Reflection is used because the account
// and target types are anonymous structs declared inline in ConfigStruct.
func voiceTargetApplyToRunningConfig(accountIndex int, targetID uint32, users []string, channels []voiceTargetChannelEntry) {
	if accountIndex < 0 || accountIndex >= len(Config.Accounts.Account) {
		return
	}
	account := &Config.Accounts.Account[accountIndex]

	position := -1
	for i := range account.Voicetargets.ID {
		if account.Voicetargets.ID[i].Value == targetID {
			position = i
			break
		}
	}
	if position < 0 {
		targets := reflect.ValueOf(&account.Voicetargets.ID).Elem()
		targets.Set(reflect.Append(targets, reflect.Zero(targets.Type().Elem())))
		position = len(account.Voicetargets.ID) - 1
		account.Voicetargets.ID[position].Value = targetID
	}
	target := &account.Voicetargets.ID[position]

	target.Users.User = append(target.Users.User, users...)

	for _, entry := range channels {
		list := reflect.ValueOf(&target.Channels.Channel).Elem()
		list.Set(reflect.Append(list, reflect.Zero(list.Type().Elem())))
		added := &target.Channels.Channel[len(target.Channels.Channel)-1]
		added.Name = entry.Name
		added.Recursive = entry.Recursive
		added.Links = entry.Links
		added.Group = entry.Group
	}
}
