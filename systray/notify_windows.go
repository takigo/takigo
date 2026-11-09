//go:build windows

package systray

// notify shows a balloon on the newest tray icon, like WinSysNotifyCmd,
// which needs "tk systray" to have created one.
func notify(title, message string) error {
	currentMu.Lock()
	t := current
	currentMu.Unlock()
	if t == nil || !t.Exists() {
		return ErrNoTray
	}
	return t.display.Server.TrayNotify(t.win, title, message)
}
