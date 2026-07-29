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

// Voice targets are spliced into talkkonnect.xml as text rather than by
// re-marshalling ConfigStruct. talkkonnect.xml is hand maintained and full of
// comments, and xml.Marshal would throw all of them away along with the tag
// order and the empty-element style, so everything outside the <voicetargets>
// element being edited is left byte for byte identical.

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// xmlIndentUnit is the indentation added per level for elements this file creates.
const xmlIndentUnit = "  "

// voiceTargetChannelEntry is one <channel> under a voice target.
type voiceTargetChannelEntry struct {
	Name      string
	Recursive bool
	Links     bool
	Group     string
}

// xmlSpan locates one element in a document. Start is the offset of '<',
// InnerStart and InnerEnd bracket the content, and End is just past the closing
// '>'. For a self-closing element InnerStart, InnerEnd and End all coincide.
type xmlSpan struct {
	Start       int
	InnerStart  int
	InnerEnd    int
	End         int
	SelfClosing bool
}

func xmlIsNameBoundary(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '>' || c == '/'
}

// xmlTagEnd returns the offset just past the '>' that closes the tag starting at
// i, ignoring '>' inside quoted attribute values.
func xmlTagEnd(doc string, i int) (end int, selfClosing bool, ok bool) {
	var quote byte
	for j := i; j < len(doc); j++ {
		c := doc[j]
		switch {
		case quote != 0:
			if c == quote {
				quote = 0
			}
		case c == '"' || c == '\'':
			quote = c
		case c == '>':
			return j + 1, j > i && doc[j-1] == '/', true
		}
	}
	return 0, false, false
}

// xmlSkipToTag returns the offset of the next element tag at or after i and
// before to, stepping over comments, CDATA, processing instructions and doctype
// declarations so that a commented-out example cannot be mistaken for markup.
func xmlSkipToTag(doc string, i, to int) (int, bool) {
	if to > len(doc) {
		to = len(doc)
	}
	for i < to {
		lt := strings.IndexByte(doc[i:to], '<')
		if lt < 0 {
			return 0, false
		}
		i += lt
		rest := doc[i:]
		switch {
		case strings.HasPrefix(rest, "<!--"):
			e := strings.Index(rest, "-->")
			if e < 0 {
				return 0, false
			}
			i += e + 3
		case strings.HasPrefix(rest, "<![CDATA["):
			e := strings.Index(rest, "]]>")
			if e < 0 {
				return 0, false
			}
			i += e + 3
		case strings.HasPrefix(rest, "<?"), strings.HasPrefix(rest, "<!"):
			e := strings.IndexByte(rest, '>')
			if e < 0 {
				return 0, false
			}
			i += e + 1
		default:
			return i, true
		}
	}
	return 0, false
}

// xmlTagNameAt reads the element name of the tag starting at t, for both open
// and closing tags.
func xmlTagNameAt(doc string, t int) string {
	j := t + 1
	if j < len(doc) && doc[j] == '/' {
		j++
	}
	k := j
	for k < len(doc) && !xmlIsNameBoundary(doc[k]) {
		k++
	}
	if k > len(doc) {
		return ""
	}
	return doc[j:k]
}

// xmlFindMatchingClose returns the offset of the </name> closing an element whose
// content starts at from, counting nested elements of the same name.
func xmlFindMatchingClose(doc, name string, from int) (int, bool) {
	depth := 0
	i := from
	for i < len(doc) {
		t, ok := xmlSkipToTag(doc, i, len(doc))
		if !ok {
			return 0, false
		}
		end, selfClosing, ok := xmlTagEnd(doc, t)
		if !ok {
			return 0, false
		}
		if xmlTagNameAt(doc, t) == name && !selfClosing {
			if doc[t+1] == '/' {
				if depth == 0 {
					return t, true
				}
				depth--
			} else {
				depth++
			}
		}
		i = end
	}
	return 0, false
}

// xmlFindElementSpan finds the first <name> element that starts inside
// doc[from:to].
func xmlFindElementSpan(doc, name string, from, to int) (xmlSpan, bool) {
	i := from
	for i < to {
		t, ok := xmlSkipToTag(doc, i, to)
		if !ok {
			return xmlSpan{}, false
		}
		end, selfClosing, ok := xmlTagEnd(doc, t)
		if !ok {
			return xmlSpan{}, false
		}
		if doc[t+1] != '/' && xmlTagNameAt(doc, t) == name {
			if selfClosing {
				return xmlSpan{Start: t, InnerStart: end, InnerEnd: end, End: end, SelfClosing: true}, true
			}
			closeStart, ok := xmlFindMatchingClose(doc, name, end)
			if !ok {
				return xmlSpan{}, false
			}
			return xmlSpan{Start: t, InnerStart: end, InnerEnd: closeStart, End: closeStart + len(name) + 3}, true
		}
		i = end
	}
	return xmlSpan{}, false
}

// xmlNthElementSpan returns the span of the n-th (zero based) <name> element in
// the document. Sibling elements are counted in document order, which is the
// same order encoding/xml fills the matching config slice in.
func xmlNthElementSpan(doc, name string, n int) (xmlSpan, bool) {
	from := 0
	for c := 0; ; c++ {
		span, ok := xmlFindElementSpan(doc, name, from, len(doc))
		if !ok {
			return xmlSpan{}, false
		}
		if c == n {
			return span, true
		}
		from = span.End
	}
}

// xmlAttrValue reads one attribute out of the open tag of span.
func xmlAttrValue(doc string, span xmlSpan, attr string) (string, bool) {
	end, _, ok := xmlTagEnd(doc, span.Start)
	if !ok {
		return "", false
	}
	tag := doc[span.Start:end]
	needle := attr + "="
	for idx := 0; ; {
		k := strings.Index(tag[idx:], needle)
		if k < 0 {
			return "", false
		}
		k += idx
		idx = k + len(needle)
		if k == 0 || !xmlIsNameBoundary(tag[k-1]) {
			continue
		}
		if idx >= len(tag) {
			return "", false
		}
		quote := tag[idx]
		if quote != '"' && quote != '\'' {
			continue
		}
		e := strings.IndexByte(tag[idx+1:], quote)
		if e < 0 {
			return "", false
		}
		return tag[idx+1 : idx+1+e], true
	}
}

// xmlIndentAt returns the whitespace preceding pos on its own line, or "" when
// pos is not the first non-blank thing on that line.
func xmlIndentAt(doc string, pos int) string {
	if pos > len(doc) {
		return ""
	}
	lineStart := strings.LastIndexByte(doc[:pos], '\n') + 1
	indent := doc[lineStart:pos]
	if strings.TrimSpace(indent) != "" {
		return ""
	}
	return indent
}

// xmlChildIndent picks the indentation for a new child of parent. New children
// are appended, so it follows the last existing child — indentation is not
// consistent throughout the shipped configs and matching the neighbour the
// insertion actually sits next to reads better than matching the first one.
// With no children to copy it indents one unit in from the parent.
func xmlChildIndent(doc string, parent xmlSpan) string {
	indent := ""
	for from := parent.InnerStart; from < parent.InnerEnd; {
		t, ok := xmlSkipToTag(doc, from, parent.InnerEnd)
		if !ok {
			break
		}
		end, selfClosing, ok := xmlTagEnd(doc, t)
		if !ok || doc[t+1] == '/' {
			break
		}
		if own := xmlIndentAt(doc, t); own != "" {
			indent = own
		}
		if selfClosing {
			from = end
			continue
		}
		name := xmlTagNameAt(doc, t)
		closeStart, ok := xmlFindMatchingClose(doc, name, end)
		if !ok {
			break
		}
		from = closeStart + len(name) + 3
	}
	if indent != "" {
		return indent
	}
	return xmlIndentAt(doc, parent.Start) + xmlIndentUnit
}

// xmlInsertChildren splices lines in as the last children of parent, each on its
// own line indented with childIndent, leaving the closing tag where it was.
func xmlInsertChildren(doc string, parent xmlSpan, childIndent string, lines []string) string {
	at := parent.InnerEnd
	head := doc[:at]
	closeIndent := xmlIndentAt(doc, at)

	if closeIndent != "" || strings.HasSuffix(head, "\n") {
		// The closing tag starts its own line: lift its indentation out of the
		// way and put it back after the new children.
		head = head[:len(head)-len(closeIndent)]
	} else {
		// An element written on a single line, e.g. <users></users>.
		closeIndent = xmlIndentAt(doc, parent.Start)
		head += "\n"
	}

	var b strings.Builder
	b.WriteString(head)
	for _, line := range lines {
		b.WriteString(childIndent)
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString(closeIndent)
	b.WriteString(doc[at:])
	return b.String()
}

// xmlEnsureOpenContainer makes sure parent has a <name> child that can take
// children of its own, expanding a self-closing <name/> and creating the element
// when it is absent. Offsets change, so callers must re-find their spans.
func xmlEnsureOpenContainer(doc string, parent xmlSpan, name string) string {
	span, found := xmlFindElementSpan(doc, name, parent.InnerStart, parent.InnerEnd)

	if found && !span.SelfClosing {
		return doc
	}

	if found && span.SelfClosing {
		indent := xmlIndentAt(doc, span.Start)
		return doc[:span.Start] + "<" + name + ">\n" + indent + "</" + name + ">" + doc[span.End:]
	}

	return xmlInsertChildren(doc, parent, xmlChildIndent(doc, parent), []string{
		"<" + name + ">",
		"</" + name + ">",
	})
}

func xmlEscapeString(s string) string {
	var b strings.Builder
	if err := xml.EscapeText(&b, []byte(s)); err != nil {
		return s
	}
	return b.String()
}

// voiceTargetXMLUserLines renders <user> elements.
func voiceTargetXMLUserLines(users []string) []string {
	out := make([]string, 0, len(users))
	for _, u := range users {
		out = append(out, "<user>"+xmlEscapeString(u)+"</user>")
	}
	return out
}

// voiceTargetXMLChannelLines renders <channel> blocks, nested one unit deep.
func voiceTargetXMLChannelLines(channels []voiceTargetChannelEntry) []string {
	out := make([]string, 0, len(channels)*6)
	for _, c := range channels {
		out = append(out,
			"<channel>",
			xmlIndentUnit+"<name>"+xmlEscapeString(c.Name)+"</name>",
			xmlIndentUnit+"<recursive>"+strconv.FormatBool(c.Recursive)+"</recursive>",
			xmlIndentUnit+"<links>"+strconv.FormatBool(c.Links)+"</links>",
			xmlIndentUnit+"<group>"+xmlEscapeString(c.Group)+"</group>",
			"</channel>",
		)
	}
	return out
}

// voiceTargetXMLIDBlock renders a whole <id value="N"> element, including only
// the containers it actually needs.
func voiceTargetXMLIDBlock(targetID uint32, users []string, channels []voiceTargetChannelEntry) []string {
	out := []string{fmt.Sprintf("<id value=\"%d\">", targetID)}
	if len(users) > 0 {
		out = append(out, xmlIndentUnit+"<users>")
		for _, line := range voiceTargetXMLUserLines(users) {
			out = append(out, xmlIndentUnit+xmlIndentUnit+line)
		}
		out = append(out, xmlIndentUnit+"</users>")
	}
	if len(channels) > 0 {
		out = append(out, xmlIndentUnit+"<channels>")
		for _, line := range voiceTargetXMLChannelLines(channels) {
			out = append(out, xmlIndentUnit+xmlIndentUnit+line)
		}
		out = append(out, xmlIndentUnit+"</channels>")
	}
	out = append(out, "</id>")
	return out
}

// voiceTargetXMLFindID locates <id value="targetID"> inside a <voicetargets> span.
func voiceTargetXMLFindID(doc string, voicetargets xmlSpan, targetID uint32) (xmlSpan, bool) {
	from := voicetargets.InnerStart
	for from < voicetargets.InnerEnd {
		span, ok := xmlFindElementSpan(doc, "id", from, voicetargets.InnerEnd)
		if !ok {
			return xmlSpan{}, false
		}
		if raw, ok := xmlAttrValue(doc, span, "value"); ok {
			if v, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 32); err == nil && uint32(v) == targetID {
				return span, true
			}
		}
		from = span.End
	}
	return xmlSpan{}, false
}

// voiceTargetXMLAdd splices users and channels into voice target targetID of the
// accountIndex-th <account>, creating the <voicetargets>, <id>, <users> and
// <channels> elements as needed. The returned document is validated by the
// caller before it replaces the config file.
func voiceTargetXMLAdd(doc string, accountIndex int, targetID uint32, users []string, channels []voiceTargetChannelEntry) (string, error) {
	if len(users) == 0 && len(channels) == 0 {
		return "", fmt.Errorf("nothing to add")
	}

	account, ok := xmlNthElementSpan(doc, "account", accountIndex)
	if !ok {
		return "", fmt.Errorf("cannot find <account> number %d in the config file", accountIndex)
	}

	doc = xmlEnsureOpenContainer(doc, account, "voicetargets")

	// Every splice shifts the offsets after it, so spans are looked up again
	// after each step rather than adjusted by hand.
	account, ok = xmlNthElementSpan(doc, "account", accountIndex)
	if !ok {
		return "", fmt.Errorf("cannot find <account> number %d after adding <voicetargets>", accountIndex)
	}
	voicetargets, ok := xmlFindElementSpan(doc, "voicetargets", account.InnerStart, account.InnerEnd)
	if !ok {
		return "", fmt.Errorf("cannot find <voicetargets> in account number %d", accountIndex)
	}

	if _, found := voiceTargetXMLFindID(doc, voicetargets, targetID); !found {
		return xmlInsertChildren(doc, voicetargets, xmlChildIndent(doc, voicetargets),
			voiceTargetXMLIDBlock(targetID, users, channels)), nil
	}

	for _, step := range []struct {
		container string
		lines     []string
	}{
		{container: "users", lines: voiceTargetXMLUserLines(users)},
		{container: "channels", lines: voiceTargetXMLChannelLines(channels)},
	} {
		if len(step.lines) == 0 {
			continue
		}

		id, err := voiceTargetXMLReFindID(doc, accountIndex, targetID)
		if err != nil {
			return "", err
		}
		doc = xmlEnsureOpenContainer(doc, id, step.container)

		id, err = voiceTargetXMLReFindID(doc, accountIndex, targetID)
		if err != nil {
			return "", err
		}
		container, ok := xmlFindElementSpan(doc, step.container, id.InnerStart, id.InnerEnd)
		if !ok {
			return "", fmt.Errorf("cannot find <%s> in voice target %d", step.container, targetID)
		}
		doc = xmlInsertChildren(doc, container, xmlChildIndent(doc, container), step.lines)
	}

	return doc, nil
}

// voiceTargetXMLReFindID re-resolves the <id> span after the document changed.
func voiceTargetXMLReFindID(doc string, accountIndex int, targetID uint32) (xmlSpan, error) {
	account, ok := xmlNthElementSpan(doc, "account", accountIndex)
	if !ok {
		return xmlSpan{}, fmt.Errorf("cannot find <account> number %d in the config file", accountIndex)
	}
	voicetargets, ok := xmlFindElementSpan(doc, "voicetargets", account.InnerStart, account.InnerEnd)
	if !ok {
		return xmlSpan{}, fmt.Errorf("cannot find <voicetargets> in account number %d", accountIndex)
	}
	id, ok := voiceTargetXMLFindID(doc, voicetargets, targetID)
	if !ok {
		return xmlSpan{}, fmt.Errorf("cannot find voice target %d in account number %d", targetID, accountIndex)
	}
	return id, nil
}
