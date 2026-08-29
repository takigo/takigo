// Package busy provides a busy window overlay that blocks user interaction
// with a target window while a long operation is in progress.
package busy

import (
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// BusyWin is an InputOnly overlay window that intercepts all input
// to the target window beneath it.
type BusyWin struct {
	overlay platform.WindowID
	display platform.DisplayServer
	target  *window.Window
}

// Hold creates an InputOnly overlay covering the target window,
// preventing all mouse and keyboard interaction.
func Hold(app widget.AppContext, target *window.Window) *BusyWin {
	if target == nil || target.PlatformID == platform.WindowID(0) {
		return nil
	}

	d := target.Display.Server
	w := uint(target.Width)
	h := uint(target.Height)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	attrs := &platform.WindowAttrs{
		EventMask: int64(
			platform.KeyPressMask |
				platform.KeyReleaseMask |
				platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.PointerMotionMask),
	}

	overlay := d.CreateWindow(
		target.PlatformID,
		0, 0, w, h, 0,
		0, // depth=0 for InputOnly
		platform.InputOnly,
		platform.CWEventMask,
		attrs,
	)

	d.RaiseWindow(overlay)
	d.MapWindow(overlay)
	d.Flush()

	// Set the busy cursor (watch cursor).
	d.SetCursorShape(overlay, uint(cursor.Watch))

	return &BusyWin{
		overlay: overlay,
		display: d,
		target:  target,
	}
}

// Release destroys the overlay window, restoring interaction.
func (b *BusyWin) Release() {
	if b == nil || b.overlay == platform.WindowID(0) {
		return
	}
	b.display.DestroyWindow(b.overlay)
	b.overlay = platform.WindowID(0)
	b.display.Flush()
}
