/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * Bottom CLI cfg commands: set/save/restart with reflect-driven paths (similar idea to
 * github.com/talkkonnect/virtualkeyz2 technician cfg commands).
 */

package talkkonnect

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/talkkonnect/colog"
)

var (
	cfgLeafCacheMu sync.Mutex
	cfgLeafPaths   []string
	cfgLeafPathSet map[string]struct{}
)

func cfgXMLSegment(sf reflect.StructField) string {
	if sf.Name == "XMLName" {
		return ""
	}
	tag := sf.Tag.Get("xml")
	if tag == "" || tag == "-" {
		return strings.ToLower(sf.Name)
	}
	first := strings.Split(tag, ",")[0]
	if first == "" {
		return strings.ToLower(sf.Name)
	}
	return strings.ToLower(first)
}

func cfgLeafPathsRefresh() {
	cfgLeafCacheMu.Lock()
	defer cfgLeafCacheMu.Unlock()
	if len(cfgLeafPaths) > 0 {
		return
	}
	var out []string
	cfgWalkLeaves(reflect.ValueOf(&Config).Elem(), "", &out)
	sort.Strings(out)
	cfgLeafPaths = out
	set := make(map[string]struct{}, len(out))
	for _, p := range out {
		set[p] = struct{}{}
	}
	cfgLeafPathSet = set
}

func cfgWalkLeaves(v reflect.Value, prefix string, out *[]string) {
	if !v.IsValid() {
		return
	}
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			sf := t.Field(i)
			if sf.PkgPath != "" {
				continue
			}
			seg := cfgXMLSegment(sf)
			if seg == "" {
				continue
			}
			next := seg
			if prefix != "" {
				next = prefix + "." + seg
			}
			cfgWalkLeaves(v.Field(i), next, out)
		}
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return
		}
		for i := 0; i < v.Len(); i++ {
			seg := strconv.Itoa(i)
			next := seg
			if prefix != "" {
				next = prefix + "." + seg
			}
			cfgWalkLeaves(v.Index(i), next, out)
		}
	default:
		if prefix != "" {
			*out = append(*out, prefix)
		}
	}
}

func cfgIsLeafPath(p string) bool {
	cfgLeafPathsRefresh()
	_, ok := cfgLeafPathSet[strings.ToLower(strings.TrimSpace(p))]
	return ok
}

func cfgDistinctFirstPathSegments() []string {
	cfgLeafPathsRefresh()
	seen := make(map[string]struct{})
	var out []string
	for _, leaf := range cfgLeafPaths {
		i := strings.IndexByte(leaf, '.')
		seg := leaf
		if i >= 0 {
			seg = leaf[:i]
		}
		if seg == "" {
			continue
		}
		if _, ok := seen[seg]; !ok {
			seen[seg] = struct{}{}
			out = append(out, seg)
		}
	}
	sort.Strings(out)
	return out
}

func cfgDistinctFollowingPathPrefixes(pathSoFar string) []string {
	cfgLeafPathsRefresh()
	pathSoFar = strings.Trim(strings.ToLower(strings.TrimSpace(pathSoFar)), ".")
	if pathSoFar == "" {
		return cfgDistinctFirstPathSegments()
	}
	needle := pathSoFar + "."
	seen := make(map[string]struct{})
	var out []string
	for _, leaf := range cfgLeafPaths {
		low := strings.ToLower(leaf)
		if !strings.HasPrefix(low, needle) {
			continue
		}
		rest := leaf[len(needle):]
		if rest == "" {
			continue
		}
		dot := strings.IndexByte(rest, '.')
		seg := rest
		if dot >= 0 {
			seg = rest[:dot]
		}
		if seg == "" {
			continue
		}
		full := pathSoFar + "." + seg
		if _, ok := seen[full]; !ok {
			seen[full] = struct{}{}
			out = append(out, full)
		}
	}
	sort.Strings(out)
	return out
}

func cfgFilterLeafPathPrefix(lowPrefix string) []string {
	cfgLeafPathsRefresh()
	if lowPrefix == "" {
		return append([]string(nil), cfgLeafPaths...)
	}
	var m []string
	for _, p := range cfgLeafPaths {
		if strings.HasPrefix(strings.ToLower(p), lowPrefix) {
			m = append(m, p)
		}
	}
	return m
}

func cfgStructFieldBySegment(st reflect.Value, seg string) (reflect.Value, error) {
	if st.Kind() != reflect.Struct {
		return reflect.Value{}, fmt.Errorf("not a struct")
	}
	low := strings.ToLower(seg)
	t := st.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.PkgPath != "" {
			continue
		}
		name := cfgXMLSegment(sf)
		if name == "" {
			continue
		}
		if name == low {
			return st.Field(i), nil
		}
	}
	return reflect.Value{}, fmt.Errorf("unknown field %q", seg)
}

func cfgEnsureSliceLen(sv reflect.Value, want int) error {
	if sv.Kind() != reflect.Slice || !sv.CanSet() {
		return fmt.Errorf("cannot grow slice")
	}
	elem := sv.Type().Elem()
	for sv.Len() <= want {
		z := reflect.New(elem).Elem()
		sv.Set(reflect.Append(sv, z))
	}
	return nil
}

func cfgNavigateSettable(root reflect.Value, path string) (reflect.Value, error) {
	segs := strings.Split(strings.Trim(strings.TrimSpace(path), "."), ".")
	if len(segs) == 1 && segs[0] == "" {
		return reflect.Value{}, fmt.Errorf("empty path")
	}
	cur := root
	for _, seg := range segs {
		for cur.Kind() == reflect.Pointer {
			if cur.IsNil() {
				return reflect.Value{}, fmt.Errorf("nil pointer in path")
			}
			cur = cur.Elem()
		}
		switch cur.Kind() {
		case reflect.Struct:
			nx, err := cfgStructFieldBySegment(cur, seg)
			if err != nil {
				return reflect.Value{}, err
			}
			cur = nx
		case reflect.Slice:
			idx, err := strconv.Atoi(seg)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("invalid slice index %q", seg)
			}
			if !cur.CanSet() {
				return reflect.Value{}, fmt.Errorf("slice not settable")
			}
			if err := cfgEnsureSliceLen(cur, idx); err != nil {
				return reflect.Value{}, err
			}
			cur = cur.Index(idx)
		default:
			return reflect.Value{}, fmt.Errorf("cannot traverse %s via %q", cur.Kind(), seg)
		}
	}
	for cur.Kind() == reflect.Pointer {
		if cur.IsNil() {
			return reflect.Value{}, fmt.Errorf("nil leaf pointer")
		}
		cur = cur.Elem()
	}
	return cur, nil
}

func cfgAssignLeaf(dst reflect.Value, s string) error {
	if !dst.CanSet() {
		return fmt.Errorf("value not settable")
	}
	s = strings.TrimSpace(s)
	switch dst.Kind() {
	case reflect.String:
		dst.SetString(s)
		return nil
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		dst.SetBool(b)
		return nil
	case reflect.Int64:
		if dst.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(s)
			if err != nil {
				return err
			}
			dst.SetInt(int64(d))
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Int32:
		if dst.Type() == reflect.TypeOf(rune(0)) {
			n, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return err
			}
			dst.SetInt(n)
			return nil
		}
		n, err := strconv.ParseInt(s, 10, 32)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16:
		bits := 32
		switch dst.Kind() {
		case reflect.Int8:
			bits = 8
		case reflect.Int16:
			bits = 16
		}
		n, err := strconv.ParseInt(s, 10, bits)
		if err != nil {
			return err
		}
		dst.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bits := 64
		switch dst.Kind() {
		case reflect.Uint8:
			bits = 8
		case reflect.Uint16:
			bits = 16
		case reflect.Uint32:
			bits = 32
		}
		n, err := strconv.ParseUint(s, 10, bits)
		if err != nil {
			return err
		}
		if dst.OverflowUint(n) {
			return fmt.Errorf("overflow")
		}
		dst.SetUint(n)
		return nil
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		dst.SetFloat(f)
		return nil
	case reflect.Struct:
		return fmt.Errorf("cannot set struct %s", dst.Type())
	default:
		return fmt.Errorf("unsupported type %s", dst.Kind())
	}
}

func cfgSetPathValue(path, value string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("empty path")
	}
	if !cfgIsLeafPath(path) {
		return fmt.Errorf("unknown or non-leaf path %q (use Tab after cfg set)", path)
	}
	dst, err := cfgNavigateSettable(reflect.ValueOf(&Config).Elem(), path)
	if err != nil {
		return err
	}
	if err := cfgAssignLeaf(dst, value); err != nil {
		return err
	}
	low := strings.ToLower(path)
	if strings.HasSuffix(low, ".loglevel") || low == "loglevel" {
		ApplyCologMinLevelFromConfig()
	}
	return nil
}

func cfgExtractAfterCfgSetCI(line string) string {
	li := strings.Index(strings.ToLower(line), "cfg set ")
	if li < 0 {
		return ""
	}
	return line[li+len("cfg set "):]
}

func cfgSetSplitPathValue(tail string) (path, value string) {
	tail = strings.TrimSpace(tail)
	if tail == "" {
		return "", ""
	}
	i := strings.IndexByte(tail, ' ')
	if i < 0 {
		return tail, ""
	}
	return strings.TrimSpace(tail[:i]), strings.TrimSpace(tail[i+1:])
}

func bottomCLITabCompleteCfg(trimmedLine string) (newLine string, bell bool) {
	trailingSpace := len(trimmedLine) > 0 && (trimmedLine[len(trimmedLine)-1] == ' ' || trimmedLine[len(trimmedLine)-1] == '\t')
	lineTR := strings.TrimRight(trimmedLine, " \t")
	low := strings.ToLower(lineTR)

	cfgSubs := []string{"help", "keys", "list", "restart", "save", "set"}

	var matches []string
	addSpace := false

	switch {
	case strings.HasPrefix(low, "cfg set"):
		tail := cfgExtractAfterCfgSetCI(lineTR)
		pathPart, valPart := cfgSetSplitPathValue(tail)
		if valPart != "" {
			return trimmedLine, true
		}
		if trailingSpace {
			switch {
			case pathPart == "":
				matches = cfgDistinctFirstPathSegments()
			case cfgIsLeafPath(pathPart):
				return trimmedLine, true
			default:
				matches = cfgDistinctFollowingPathPrefixes(pathPart)
			}
		} else {
			if pathPart == "" {
				matches = cfgDistinctFirstPathSegments()
			} else {
				matches = cfgFilterLeafPathPrefix(strings.ToLower(pathPart))
			}
		}
	case strings.HasPrefix(low, "cfg"):
		fields := strings.Fields(lineTR)
		switch {
		case len(fields) == 1 && trailingSpace:
			matches = cfgSubs
		case len(fields) == 1 && !trailingSpace:
			p := strings.TrimSpace(strings.TrimPrefix(low, "cfg"))
			if p == "" {
				matches = cfgSubs
			} else {
				for _, s := range cfgSubs {
					if strings.HasPrefix(s, p) {
						matches = append(matches, s)
					}
				}
			}
		case len(fields) == 2 && trailingSpace && strings.EqualFold(fields[1], "set"):
			matches = cfgDistinctFirstPathSegments()
		case len(fields) == 2 && !trailingSpace:
			p := strings.ToLower(fields[1])
			for _, s := range cfgSubs {
				if strings.HasPrefix(s, p) {
					matches = append(matches, s)
				}
			}
		default:
			return trimmedLine, true
		}
	default:
		return trimmedLine, true
	}

	if len(matches) == 0 {
		return trimmedLine, true
	}

	var pick string
	if len(matches) == 1 {
		pick = matches[0]
		if strings.HasPrefix(low, "cfg set") {
			addSpace = cfgIsLeafPath(pick)
		} else {
			addSpace = true
		}
	} else {
		pick = bottomCLILongestCommonPrefix(matches)
		if strings.HasPrefix(low, "cfg set") && !trailingSpace {
			tail := cfgExtractAfterCfgSetCI(lineTR)
			pathPart, _ := cfgSetSplitPathValue(tail)
			lp := strings.ToLower(pathPart)
			if !strings.HasPrefix(strings.ToLower(pick), lp) || len(strings.ToLower(pick)) == len(lp) {
				return trimmedLine, true
			}
		}
		addSpace = false
	}

	if strings.HasPrefix(low, "cfg set") {
		idx := strings.Index(strings.ToLower(lineTR), "cfg set ")
		if idx < 0 {
			return trimmedLine, true
		}
		pre := lineTR[:idx+len("cfg set ")]
		suffix := pick
		if addSpace {
			suffix += " "
		}
		return pre + suffix, false
	}

	fields := strings.Fields(lineTR)
	if len(fields) == 1 && trailingSpace {
		return fields[0] + " " + pick + " ", false
	}
	if len(fields) == 1 && !trailingSpace {
		return fields[0] + " " + pick + " ", false
	}
	if len(fields) == 2 && trailingSpace && strings.EqualFold(fields[1], "set") {
		return fields[0] + " " + fields[1] + " " + pick + " ", false
	}
	if len(fields) == 2 && !trailingSpace {
		return fields[0] + " " + pick + " ", false
	}
	return trimmedLine, true
}

func saveTalkkonnectXMLConfig() error {
	p := strings.TrimSpace(ConfigXMLFile)
	if p == "" {
		return fmt.Errorf("config file path is empty")
	}
	data, err := xml.MarshalIndent(Config, "", "  ")
	if err != nil {
		return err
	}
	out := append([]byte(xml.Header), data...)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, out, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, p); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func restartTalkkonnectProcess() error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	self, err = filepath.Abs(self)
	if err != nil {
		return err
	}
	args := append([]string{self}, os.Args[1:]...)
	env := os.Environ()
	return syscall.Exec(self, args, env)
}

// ApplyCologMinLevelFromConfig applies Config.Global.Software.Settings.Loglevel to colog.
func ApplyCologMinLevelFromConfig() {
	switch strings.ToLower(strings.TrimSpace(Config.Global.Software.Settings.Loglevel)) {
	case "trace":
		colog.SetMinLevel(colog.LTrace)
		log.Println("info: Loglevel Set to Trace")
	case "debug":
		colog.SetMinLevel(colog.LDebug)
		log.Println("info: Loglevel Set to Debug")
	case "info":
		colog.SetMinLevel(colog.LInfo)
		log.Println("info: Loglevel Set to Info")
	case "warning":
		colog.SetMinLevel(colog.LWarning)
		log.Println("info: Loglevel Set to Warning")
	case "error":
		colog.SetMinLevel(colog.LError)
		log.Println("info: Loglevel Set to Error")
	case "alert":
		colog.SetMinLevel(colog.LAlert)
		log.Println("info: Loglevel Set to Alert")
	default:
		colog.SetMinLevel(colog.LInfo)
		log.Println("info: Default Loglevel unset in XML config automatically loglevel to Info")
	}
}

func bottomCLIHandleCfgLine(w io.Writer, line string) {
	parts := strings.Fields(strings.TrimSpace(line))
	if len(parts) < 1 || !strings.EqualFold(parts[0], "cfg") {
		return
	}
	if len(parts) < 2 {
		fmt.Fprintln(w, "Usage: cfg keys|list|set <path> <value>|save|restart")
		return
	}
	sub := strings.ToLower(parts[1])
	switch sub {
	case "keys", "help", "h", "?":
		cfgLeafPathsRefresh()
		fmt.Fprintf(w, "%d configurable leaf paths (dot-separated XML tag names, lowercase). Examples:\n", len(cfgLeafPaths))
		fmt.Fprintf(w, "  cfg set global.software.settings.loglevel debug\n")
		fmt.Fprintf(w, "  cfg set accounts.account.0.username mycall\n")
		fmt.Fprintln(w, "Note: some runtime data (e.g. account slices filled at startup) updates fully after cfg restart.")
		max := 40
		for i, p := range cfgLeafPaths {
			if i >= max {
				fmt.Fprintf(w, "... and %d more (Tab-complete after: cfg set )\n", len(cfgLeafPaths)-max)
				break
			}
			fmt.Fprintln(w, p)
		}
	case "list", "show", "l":
		cfgLeafPathsRefresh()
		max := 3000
		for i, p := range cfgLeafPaths {
			if i >= max {
				fmt.Fprintf(w, "... truncated (%d paths total)\n", len(cfgLeafPaths))
				break
			}
			dst, err := cfgNavigateSettable(reflect.ValueOf(&Config).Elem(), p)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "%s = %v\n", p, formatCfgLeafValue(dst))
		}
	case "set":
		tail := cfgExtractAfterCfgSetCI(line)
		path, val := cfgSetSplitPathValue(tail)
		if path == "" {
			fmt.Fprintln(w, "Usage: cfg set <dotted.path> <value>")
			fmt.Fprintln(w, "Example: cfg set global.software.settings.loglevel info")
			return
		}
		if err := cfgSetPathValue(path, val); err != nil {
			fmt.Fprintf(w, "cfg set: %v\n", err)
			log.Printf("warn: cfg set: %v", err)
			return
		}
		fmt.Fprintf(w, "Set %q OK (in memory). Use \"cfg save\" to write %s\n", path, ConfigXMLFile)
		log.Printf("info: cfg set %q", path)
	case "save", "write":
		if err := saveTalkkonnectXMLConfig(); err != nil {
			fmt.Fprintf(w, "cfg save: %v\n", err)
			log.Printf("warn: cfg save: %v", err)
			return
		}
		fmt.Fprintf(w, "Configuration saved to %q\n", ConfigXMLFile)
		log.Printf("info: cfg save wrote %q", ConfigXMLFile)
	case "restart", "reboot":
		fmt.Fprintln(w, "Restarting: replacing process via exec (same binary and arguments).")
		log.Println("info: cfg restart — exec same binary")
		bottomCLIDisableLayout()
		bottomCLITerminalHardReset()
		if err := restartTalkkonnectProcess(); err != nil {
			fmt.Fprintf(w, "cfg restart: %v\n", err)
			log.Printf("error: cfg restart: %v", err)
		}
	default:
		fmt.Fprintf(w, "Unknown cfg subcommand %q. Try: cfg keys\n", parts[1])
	}
}

func formatCfgLeafValue(v reflect.Value) string {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Bool:
		return strconv.FormatBool(v.Bool())
	case reflect.Int64:
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			return v.Interface().(time.Duration).String()
		}
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	default:
		return fmt.Sprint(v.Interface())
	}
}
