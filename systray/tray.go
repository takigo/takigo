// Package systray provides X11 system tray (_NET_SYSTEM_TRAY) support.
// It creates a small window that is docked into the desktop's system
// tray / notification area. On other platforms New returns
// ErrUnsupported.
package systray

import (
	"errors"

	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// ErrUnsupported is returned by New on platforms without a tray backend.
var ErrUnsupported = errors.New("systray: not supported on this platform")

// TrayIcon represents an icon in the system tray.
type TrayIcon struct {
	win     platform.WindowID
	display platform.DisplayServer
	app     widget.AppContext

	tooltip           string
	clickHandler      func()
	rightClickHandler func(x, y int)
}

// TrayOption configures a TrayIcon.
type TrayOption func(*TrayIcon)

func TrayTooltip(s string) TrayOption       { return func(t *TrayIcon) { t.tooltip = s } }
func TrayClickHandler(fn func()) TrayOption { return func(t *TrayIcon) { t.clickHandler = fn } }
func TrayRightClickHandler(fn func(x, y int)) TrayOption {
	return func(t *TrayIcon) { t.rightClickHandler = fn }
}

// SetTooltip sets the tooltip text for the tray icon.
func (t *TrayIcon) SetTooltip(s string) {
	t.tooltip = s
	if t.win == platform.WindowID(0) {
		return
	}
	utf8Atom := t.display.InternAtom("UTF8_STRING", false)
	netWmName := t.display.InternAtom("_NET_WM_NAME", false)
	data := []byte(s)
	t.display.ChangeProperty(t.win, netWmName, utf8Atom, 8, platform.PropModeReplace, data, len(data))
}

// Destroy removes the tray icon and unbinds its event handlers.
func (t *TrayIcon) Destroy() {
	if t.win == platform.WindowID(0) {
		return
	}
	t.app.Dispatcher().Unbind(t.win)
	t.display.DestroyWindow(t.win)
	t.win = platform.WindowID(0)
}
