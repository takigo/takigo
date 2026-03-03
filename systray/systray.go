//go:build linux

// Package systray provides X11 system tray (_NET_SYSTEM_TRAY) support.
// It creates a small window that is docked into the desktop's system
// tray / notification area.
package systray

import (
	"fmt"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

const (
	systemTrayRequestDock = 0
	trayIconSize          = 24
)

// TrayIcon represents an icon in the system tray.
type TrayIcon struct {
	win     platform.WindowID
	display platform.DisplayServer
	app     widget.AppContext

	tooltip      string
	clickHandler func()
}

// TrayOption configures a TrayIcon.
type TrayOption func(*TrayIcon)

func TrayTooltip(s string) TrayOption       { return func(t *TrayIcon) { t.tooltip = s } }
func TrayClickHandler(fn func()) TrayOption { return func(t *TrayIcon) { t.clickHandler = fn } }

// New creates a new system tray icon and docks it.
// Returns an error if no system tray manager is running.
func New(app widget.AppContext, winDisplay *window.Display, opts ...TrayOption) (*TrayIcon, error) {
	d := winDisplay.Server

	// Find the system tray manager.
	screen := winDisplay.Screen
	trayAtomName := fmt.Sprintf("_NET_SYSTEM_TRAY_S%d", screen)
	trayAtom := d.InternAtom(trayAtomName, false)
	manager := d.GetSelectionOwner(trayAtom)
	if manager == platform.WindowID(0) {
		return nil, fmt.Errorf("systray: no system tray manager found")
	}

	// Create the tray icon window (24x24).
	attrs := &platform.WindowAttrs{
		BackgroundPixel: 0,
		EventMask: int64(
			platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.ExposureMask |
				platform.StructureNotifyMask),
	}

	iconWin := d.CreateWindow(
		winDisplay.RootWindow,
		0, 0, trayIconSize, trayIconSize, 0,
		platform.CopyFromParent, platform.InputOutput,
		platform.CWBackPixel|platform.CWEventMask,
		attrs,
	)

	t := &TrayIcon{
		win:     iconWin,
		display: d,
		app:     app,
	}

	for _, opt := range opts {
		opt(t)
	}

	// Set tooltip via _NET_WM_NAME if provided.
	if t.tooltip != "" {
		t.SetTooltip(t.tooltip)
	}

	// Send SYSTEM_TRAY_REQUEST_DOCK message to the tray manager.
	opcodeAtom := d.InternAtom("_NET_SYSTEM_TRAY_OPCODE", false)
	d.SendClientMessage(
		iconWin, manager, opcodeAtom,
		int64(platform.CurrentTime),
		systemTrayRequestDock,
		int64(iconWin),
		0, 0,
	)
	d.Flush()

	// Bind click handler.
	if t.clickHandler != nil {
		app.Dispatcher().Bind(iconWin, event.ButtonPressMask, func(ev *event.Event) {
			if ev.Button == 1 && t.clickHandler != nil {
				t.clickHandler()
			}
		})
	}

	// Bind expose to draw a simple icon placeholder.
	app.Dispatcher().Bind(iconWin, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		t.draw()
	})

	return t, nil
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

// Destroy removes the tray icon.
func (t *TrayIcon) Destroy() {
	if t.win == platform.WindowID(0) {
		return
	}
	t.display.DestroyWindow(t.win)
	t.win = platform.WindowID(0)
}

// draw paints a simple placeholder icon (a filled circle).
func (t *TrayIcon) draw() {
	if t.win == platform.WindowID(0) {
		return
	}
	// Create a temporary GC for drawing.
	gc := t.display.CreateGC(platform.WindowDrawable(t.win), 0, &platform.GCValues{})
	defer t.display.FreeGC(gc)

	// Fill background with dark grey.
	t.display.SetForeground(gc, 0x404040)
	t.display.FillRectangle(platform.WindowDrawable(t.win), gc, 0, 0, trayIconSize, trayIconSize)

	// Draw a lighter circle in the center.
	t.display.SetForeground(gc, 0x80B0FF)
	t.display.FillArc(platform.WindowDrawable(t.win), gc, 4, 4, trayIconSize-8, trayIconSize-8, 0, 360*64)

	t.display.Flush()
}
