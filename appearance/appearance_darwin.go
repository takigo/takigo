//go:build darwin

package appearance

import (
	"os/exec"
	"strings"
)

// system reads the AppleInterfaceStyle default, which is "Dark" in dark
// mode and absent otherwise. The read forks "defaults", so it is cached.
func system() Mode {
	m, _ := cachedLookup(defaultsMode)
	return m
}

func defaultsMode() (Mode, bool) {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err == nil && strings.EqualFold(strings.TrimSpace(string(out)), "dark") {
		return Dark, true
	}
	return Light, true
}
