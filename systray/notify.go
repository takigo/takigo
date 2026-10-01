package systray

import "errors"

// ErrNoNotifier is returned by Notify when the desktop has no notification
// service to deliver to.
var ErrNoNotifier = errors.New("systray: no notification service")

// Notify shows a desktop notification, like Tk's "tk sysnotify". On Linux
// and the BSDs it calls org.freedesktop.Notifications on the session bus;
// on macOS it goes through osascript. Windows is not implemented and
// returns ErrUnsupported.
func Notify(title, message string) error {
	return notify(title, message)
}
