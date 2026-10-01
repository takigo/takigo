// Package appearance reports whether the desktop asks applications for a
// light or a dark look. Tk has no counterpart; ttk.UseSystemTheme uses it
// to pick a theme.
package appearance

import (
	"os"
	"strings"
)

// Mode is the look the desktop asks for.
type Mode int

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

// System returns the desktop's current preference. It asks the desktop
// each time, so call it again to notice a change. The environment variable
// TAKIGO_APPEARANCE (light or dark) overrides the desktop.
func System() Mode {
	switch strings.ToLower(os.Getenv("TAKIGO_APPEARANCE")) {
	case "dark":
		return Dark
	case "light":
		return Light
	}
	return system()
}
