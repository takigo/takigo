package systray

import "errors"

// ErrNoNotifier is returned by Notify when the desktop has no notification
// service to deliver to.
var ErrNoNotifier = errors.New("systray: no notification service")

// Notify shows a desktop notification, like Tk's "tk sysnotify". On Linux
// and the BSDs it calls org.freedesktop.Notifications on the session bus;
// on macOS it goes through osascript; on Windows it is a balloon on the
// tray icon, so it returns ErrNoTray until New has created one.
func Notify(title, message string) error {
	return notify(title, message)
}
