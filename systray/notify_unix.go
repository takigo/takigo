//go:build !windows && !darwin

package systray

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/msorc/takigo/internal/dbus"
)

func notify(title, message string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNoNotifier, err)
	}
	defer conn.Close()
	// Notify(app_name, replaces_id, app_icon, summary, body, actions,
	// hints, expire_timeout); -1 leaves the timeout to the server.
	_, err = conn.Call("org.freedesktop.Notifications", "/org/freedesktop/Notifications",
		"org.freedesktop.Notifications", "Notify", "susssasa{sv}i",
		filepath.Base(os.Args[0]), uint32(0), "", title, message,
		[]string{}, map[string]dbus.Variant{}, int32(-1))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNoNotifier, err)
	}
	return nil
}
