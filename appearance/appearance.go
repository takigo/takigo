// Package appearance reports whether the desktop asks applications for a
// light or a dark look. Tk has no counterpart; ttk.UseSystemTheme uses it
// to pick a theme.
package appearance

import (
	"os"
	"strings"
	"sync"
	"time"
)

// lookupTTL is how long a desktop lookup is reused. The lookup can be
// slow: the XDG portal call dials the session bus, and macOS forks
// "defaults", so a call in an event handler must not repeat it.
const lookupTTL = 30 * time.Second

// lookupCache holds the last desktop lookup and when it was made.
type lookupCache struct {
	mu    sync.Mutex
	mode  Mode
	ok    bool
	at    time.Time
	valid bool
}

// get returns the cached mode, or looks it up and caches the result.
func (c *lookupCache) get(lookup func() (Mode, bool)) (Mode, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.valid && time.Since(c.at) < lookupTTL {
		return c.mode, c.ok
	}
	mode, ok := lookup()
	c.mode, c.ok, c.at, c.valid = mode, ok, time.Now(), true
	return mode, ok
}

// reset drops the cached lookup.
func (c *lookupCache) reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.valid = false
}

// systemCache holds the platform's last desktop lookup.
var systemCache lookupCache

// cachedLookup returns the desktop's answer, asking the desktop at most
// once every lookupTTL.
func cachedLookup(lookup func() (Mode, bool)) (Mode, bool) {
	return systemCache.get(lookup)
}

// resetCache drops the cached lookup; for tests.
func resetCache() {
	systemCache.reset()
}

// Mode is the look the desktop asks for.
type Mode int

// The appearances.
const (
	Light Mode = iota
	Dark
)

func (m Mode) String() string {
	if m == Dark {
		return "dark"
	}
	return "light"
}

// System returns the desktop's current preference. The desktop is asked
// at most once every 30 seconds, because the lookup can block (it dials
// the session bus on X11, and forks "defaults" on macOS); call it again
// later to notice a change. The environment variable TAKIGO_APPEARANCE
// (light or dark) overrides the desktop and is not cached.
func System() Mode {
	switch strings.ToLower(os.Getenv("TAKIGO_APPEARANCE")) {
	case "dark":
		return Dark
	case "light":
		return Light
	}
	return system()
}
