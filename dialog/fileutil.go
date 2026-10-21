package dialog

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// splitPatterns turns a file type's pattern list ("*.c *.h") into its
// glob patterns; Tk file types carry a list of extensions the same way.
func splitPatterns(pattern string) []string {
	f := strings.Fields(pattern)
	if len(f) == 0 {
		return []string{"*"}
	}
	return f
}

// matchesPatterns reports whether name matches any glob; an empty list or
// "*" accepts everything.
func matchesPatterns(name string, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, p := range patterns {
		if p == "*" {
			return true
		}
		if ok, _ := filepath.Match(p, name); ok {
			return true
		}
	}
	return false
}

// humanSize formats a byte count the way file managers do.
func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	v := float64(n) / float64(div)
	suffix := "KMGTPE"[exp : exp+1]
	if v < 10 {
		return fmt.Sprintf("%.1f %sB", v, suffix)
	}
	return fmt.Sprintf("%.0f %sB", v, suffix)
}

// expandUser expands a leading ~ or ~user, like Tcl's file join.
func expandUser(p string) string {
	if p == "" || p[0] != '~' {
		return p
	}
	rest := p[1:]
	name, tail, _ := strings.Cut(rest, "/")
	var home string
	if name == "" {
		home, _ = os.UserHomeDir()
	} else if u, err := user.Lookup(name); err == nil {
		home = u.HomeDir
	}
	if home == "" {
		return p
	}
	if tail == "" && !strings.Contains(rest, "/") {
		return home
	}
	return filepath.Join(home, tail)
}

// resolveFlag classifies what the user typed, as ResolveFile in
// tkfbox.tcl does.
type resolveFlag int

const (
	resolveOK      resolveFlag = iota // existing file (file != "") or directory (file == "")
	resolvePattern                    // existing directory plus a glob
	resolveNewFile                    // existing directory, file does not exist
	resolvePath                       // the directory does not exist
)

// resolveFile interprets text relative to dir. A relative or ~ path is
// joined to dir, $VAR is expanded from the environment and a missing
// extension gets defaultExt, except for directories.
func resolveFile(dir, text, defaultExt string) (flag resolveFlag, outDir, file string) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "$") {
		text = os.ExpandEnv(text)
	}
	text = expandUser(text)
	path := text
	if !filepath.IsAbs(path) {
		path = filepath.Join(dir, path)
	}
	path = filepath.Clean(path)
	if strings.HasSuffix(text, string(filepath.Separator)) && path != string(filepath.Separator) {
		// "name/" asks for a directory: never add an extension to it.
		defaultExt = ""
	}

	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return resolveOK, path, ""
	}
	if err != nil && defaultExt != "" && filepath.Ext(path) == "" {
		withExt := path + defaultExt
		if _, err2 := os.Stat(withExt); err2 == nil || !strings.ContainsAny(filepath.Base(path), "*?[") {
			path = withExt
			_, err = os.Stat(path)
		}
	}
	if err == nil {
		return resolveOK, filepath.Dir(path), filepath.Base(path)
	}
	parent, base := filepath.Dir(path), filepath.Base(path)
	if pi, perr := os.Stat(parent); perr == nil && pi.IsDir() {
		if strings.ContainsAny(base, "*?[") {
			return resolvePattern, parent, base
		}
		return resolveNewFile, parent, base
	}
	return resolvePath, parent, base
}

// completeName ports CompleteEnt: the unique candidate extending prefix,
// or the longest common prefix of the candidates that do. ok is false
// when nothing (new) can be completed.
func completeName(prefix string, candidates []string) (string, bool) {
	var targets []string
	for _, c := range candidates {
		if strings.HasPrefix(c, prefix) {
			targets = append(targets, c)
		}
	}
	switch {
	case len(targets) == 0:
		return prefix, false
	case len(targets) == 1:
		return targets[0], targets[0] != prefix
	}
	common := targets[0]
	for _, t := range targets[1:] {
		for !strings.HasPrefix(t, common) {
			common = common[:len(common)-1]
		}
	}
	return common, common != prefix
}

// parseNameList splits a multiple-selection entry: names separated by
// spaces, with "double quotes" around names that contain spaces. Inside
// the quotes a backslash escapes a backslash or a double quote, except
// that a backslash and quote at the end of a name (a typed Windows
// directory such as "C:\My Files\") close the quotes.
func parseNameList(text string) []string {
	var names []string
	var cur strings.Builder
	inQuote, has := false, false
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch {
		case inQuote && c == '\\' && i+1 < len(text) && (text[i+1] == '\\' || text[i+1] == '"' && !endsName(text, i+2)):
			cur.WriteByte(text[i+1])
			has = true
			i++
		case c == '"':
			inQuote = !inQuote
			has = true
		case c == ' ' && !inQuote:
			if has {
				names = append(names, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteByte(c)
			has = true
		}
	}
	if has {
		names = append(names, cur.String())
	}
	return names
}

// endsName reports whether a name ends at text[i].
func endsName(text string, i int) bool {
	return i >= len(text) || text[i] == ' '
}

// quoteNames is the inverse of parseNameList: names with a space, a
// double quote or a backslash are quoted, and a backslash or a double
// quote inside them is escaped with a backslash.
func quoteNames(names []string) string {
	parts := make([]string, len(names))
	for i, n := range names {
		if strings.ContainsAny(n, " \"\\") {
			var b strings.Builder
			b.WriteByte('"')
			for j := 0; j < len(n); j++ {
				if n[j] == '"' || n[j] == '\\' {
					b.WriteByte('\\')
				}
				b.WriteByte(n[j])
			}
			b.WriteByte('"')
			n = b.String()
		}
		parts[i] = n
	}
	return strings.Join(parts, " ")
}

// dirHistory is the file browser's back/forward list of visited
// directories; pos indexes the current one.
type dirHistory struct {
	dirs []string
	pos  int
}

// visit records dir as the current directory, dropping the forward
// entries; revisiting the current directory changes nothing.
func (h *dirHistory) visit(dir string) {
	if len(h.dirs) > 0 && h.dirs[h.pos] == dir {
		return
	}
	if len(h.dirs) > 0 {
		h.dirs = h.dirs[:h.pos+1]
	}
	h.dirs = append(h.dirs, dir)
	h.pos = len(h.dirs) - 1
}

func (h *dirHistory) canBack() bool    { return h.pos > 0 }
func (h *dirHistory) canForward() bool { return h.pos+1 < len(h.dirs) }

// back moves to the previous directory and returns it.
func (h *dirHistory) back() (string, bool) {
	if !h.canBack() {
		return "", false
	}
	h.pos--
	return h.dirs[h.pos], true
}

// forward moves to the next directory and returns it.
func (h *dirHistory) forward() (string, bool) {
	if !h.canForward() {
		return "", false
	}
	h.pos++
	return h.dirs[h.pos], true
}
