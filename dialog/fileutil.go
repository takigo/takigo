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
			info, err = os.Stat(path)
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
// spaces, with "double quotes" around names that contain spaces.
func parseNameList(text string) []string {
	var names []string
	var cur strings.Builder
	inQuote, has := false, false
	for _, r := range text {
		switch {
		case r == '"':
			inQuote = !inQuote
			has = true
		case r == ' ' && !inQuote:
			if has {
				names = append(names, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	if has {
		names = append(names, cur.String())
	}
	return names
}

// quoteNames is the inverse of parseNameList.
func quoteNames(names []string) string {
	parts := make([]string, len(names))
	for i, n := range names {
		if strings.ContainsAny(n, " \"") {
			n = `"` + strings.ReplaceAll(n, `"`, ``) + `"`
		}
		parts[i] = n
	}
	return strings.Join(parts, " ")
}
