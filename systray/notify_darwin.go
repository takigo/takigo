//go:build darwin

package systray

import (
	"fmt"
	"os/exec"
	"strconv"
)

// notify runs AppleScript's "display notification", which needs no bundle
// or entitlement, unlike UNUserNotificationCenter.
func notify(title, message string) error {
	script := "display notification " + strconv.Quote(message) + " with title " + strconv.Quote(title)
	if err := exec.Command("osascript", "-e", script).Run(); err != nil {
		return fmt.Errorf("%w: %w", ErrNoNotifier, err)
	}
	return nil
}
