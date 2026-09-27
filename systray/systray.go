//go:build linux || freebsd || openbsd || netbsd

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

	// Bind click handlers.
	app.Dispatcher().Bind(iconWin, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 && t.clickHandler != nil {
			t.clickHandler()
		} else if ev.Button == 3 && t.rightClickHandler != nil {
			t.rightClickHandler(ev.RootX, ev.RootY)
		}
	})

	// Bind expose to draw a simple icon placeholder.
	app.Dispatcher().Bind(iconWin, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		t.draw()
	})

	return t, nil
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
