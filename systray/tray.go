// Package systray puts an icon in the desktop's system tray and shows
// desktop notifications, like Tk's "tk systray" and "tk sysnotify"
// (tk/library/systray.tcl). On X11 the icon is a window docked with the
// System Tray protocol (tk/unix/tkUnixSysTray.c), on Windows a
// Shell_NotifyIconW icon (tk/win/tkWinSysTray.c); macOS has no tray icon
// yet and New returns ErrUnsupported.
package systray

import (
	"errors"
	goimage "image"
	"image/draw"
	"runtime"
	"sync"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// ErrUnsupported is returned by New on platforms without a tray backend.
var ErrUnsupported = errors.New("systray: not supported on this platform")

// ErrNoTray is returned by Notify on Windows before a tray icon exists:
// the notification is a balloon on the icon.
var ErrNoTray = errors.New("systray: must create a system tray icon first")

// TrayIcon represents an icon in the system tray.
type TrayIcon struct {
	app     widget.AppContext
	display *window.Display
	win     platform.WindowID // the docked window (X11) or the icon's window (native)
	native  bool              // the backend's TrayManager holds the icon

	image             widget.WidgetImage
	tooltip           string
	clickHandler      func()
	rightClickHandler func(x, y int)

	balloon       *balloon // X11 tooltip window
	width, height int      // the docked window's size
}

// TrayOption configures a TrayIcon.
type TrayOption func(*TrayIcon)

// TrayImage sets the icon's image (-image). Windows needs an image with
// pixels, such as an *image.Photo.
func TrayImage(img widget.WidgetImage) TrayOption { return func(t *TrayIcon) { t.image = img } }

// TrayTooltip sets the tooltip text shown over the icon (-text).
func TrayTooltip(s string) TrayOption { return func(t *TrayIcon) { t.tooltip = s } }

// TrayClickHandler sets the command for a left click (-button1).
func TrayClickHandler(fn func()) TrayOption { return func(t *TrayIcon) { t.clickHandler = fn } }

// TrayRightClickHandler sets the command for a right click (-button3); it
// gets the pointer's root coordinates, where a popup menu goes.
func TrayRightClickHandler(fn func(x, y int)) TrayOption {
	return func(t *TrayIcon) { t.rightClickHandler = fn }
}

// current is the newest live tray icon of the process: Tk allows one per
// interpreter, and Windows notifications show on it.
var (
	currentMu sync.Mutex
	current   *TrayIcon
)

// New creates a tray icon. It fails with ErrUnsupported on macOS and when
// neither the platform nor a running X11 tray manager can hold it.
func New(app widget.AppContext, winDisplay *window.Display, opts ...TrayOption) (*TrayIcon, error) {
	if runtime.GOOS == "darwin" {
		return nil, ErrUnsupported
	}
	t := &TrayIcon{app: app, display: winDisplay}
	for _, opt := range opts {
		opt(t)
	}
	d := winDisplay.Server
	if w := d.AddTrayIcon(iconImage(t.image), t.tooltip); w != 0 {
		t.win, t.native = w, true
	} else if err := t.dock(); err != nil {
		return nil, err
	}
	app.Dispatcher().Bind(t.win, event.ButtonPressMask, t.press)
	currentMu.Lock()
	current = t
	currentMu.Unlock()
	return t, nil
}

func (t *TrayIcon) press(ev *event.Event) {
	switch {
	case ev.Button == 1 && t.clickHandler != nil:
		t.clickHandler()
	case ev.Button == 3 && t.rightClickHandler != nil:
		t.rightClickHandler(ev.RootX, ev.RootY)
	}
}

// Configure changes the icon's options ("tk systray configure").
func (t *TrayIcon) Configure(opts ...TrayOption) {
	oldImage := t.image
	for _, opt := range opts {
		opt(t)
	}
	if !t.Exists() {
		return
	}
	if t.native {
		var icon *platform.IconImage
		if t.image != oldImage {
			icon = iconImage(t.image)
		}
		t.display.Server.ModifyTrayIcon(t.win, icon, t.tooltip)
		return
	}
	t.draw()
}

// SetTooltip sets the tooltip text for the tray icon.
func (t *TrayIcon) SetTooltip(s string) { t.Configure(TrayTooltip(s)) }

// Exists reports whether the icon has not been destroyed.
func (t *TrayIcon) Exists() bool { return t.win != 0 }

// Destroy removes the tray icon and unbinds its event handlers.
func (t *TrayIcon) Destroy() {
	if !t.Exists() {
		return
	}
	t.app.Dispatcher().Unbind(t.win)
	if t.native {
		t.display.Server.RemoveTrayIcon(t.win)
	} else {
		t.balloon.close()
		t.display.Server.DestroyWindow(t.win)
	}
	t.win = 0
	currentMu.Lock()
	if current == t {
		current = nil
	}
	currentMu.Unlock()
}

// Notify shows a notification from the icon: a balloon on Windows, the
// desktop's notification elsewhere (see the package-level Notify).
func (t *TrayIcon) Notify(title, message string) error {
	if t.native {
		return t.display.Server.TrayNotify(t.win, title, message)
	}
	return notify(title, message)
}

// iconImage converts an image with pixels to the backend's icon format.
func iconImage(img widget.WidgetImage) *platform.IconImage {
	src, ok := img.(interface{ RGBA() *goimage.RGBA })
	if !ok || src.RGBA() == nil {
		return nil
	}
	b := src.RGBA().Bounds()
	n := goimage.NewNRGBA(goimage.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(n, n.Rect, src.RGBA(), b.Min, draw.Src)
	return &platform.IconImage{Width: b.Dx(), Height: b.Dy(), Pix: n.Pix}
}
