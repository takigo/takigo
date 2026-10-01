//go:build !windows && !darwin

package systray

import (
	"errors"
	"testing"
)

// The test must not pop a notification on the developer's desktop, so it
// only covers the path with no bus to talk to.
func TestNotifyWithoutBus(t *testing.T) {
	t.Setenv("DBUS_SESSION_BUS_ADDRESS", "unix:path=/nonexistent/takigo-test-bus")
	if err := Notify("title", "message"); !errors.Is(err, ErrNoNotifier) {
		t.Errorf("Notify without a bus = %v, want ErrNoNotifier", err)
	}
}
