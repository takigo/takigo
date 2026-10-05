//go:build !windows && !darwin

package appearance

import (
	"os"
	"strings"

	"github.com/takigo/takigo/internal/dbus"
)

// system asks the XDG settings portal for org.freedesktop.appearance
// color-scheme (1 is "prefer dark"), which GNOME, KDE and the wlroots
// portals all answer, and falls back to a GTK_THEME ending in "dark".
func system() Mode {
	if m, ok := portalColorScheme(); ok {
		return m
	}
	if theme := strings.ToLower(os.Getenv("GTK_THEME")); strings.HasSuffix(theme, "dark") {
		return Dark
	}
	return Light
}

func portalColorScheme() (Mode, bool) {
	conn, err := dbus.SessionBus()
	if err != nil {
		return Light, false
	}
	defer conn.Close()
	reply, err := conn.Call("org.freedesktop.portal.Desktop", "/org/freedesktop/portal/desktop",
		"org.freedesktop.portal.Settings", "Read", "ss", "org.freedesktop.appearance", "color-scheme")
	if err != nil || len(reply) == 0 {
		return Light, false
	}
	return colorSchemeMode(reply[0])
}

// colorSchemeMode unwraps the portal's reply, a variant holding a variant
// holding a uint32 (Read) or the uint32 in one variant (ReadOne).
func colorSchemeMode(v any) (Mode, bool) {
	for {
		inner, ok := v.(dbus.Variant)
		if !ok {
			break
		}
		v = inner.Value
	}
	n, ok := v.(uint32)
	if !ok {
		return Light, false
	}
	if n == 1 {
		return Dark, true
	}
	return Light, true
}
