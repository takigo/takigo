//go:build darwin

package appearance

import (
	"os/exec"
	"strings"
)

// system reads the AppleInterfaceStyle default, which is "Dark" in dark
// mode and absent otherwise.
func system() Mode {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err == nil && strings.EqualFold(strings.TrimSpace(string(out)), "dark") {
		return Dark
	}
	return Light
}
