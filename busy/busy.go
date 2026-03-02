// Package busy provides a busy window overlay that blocks user interaction
// with a target window while a long operation is in progress.
package busy

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// BusyWin is an InputOnly overlay window that intercepts all input
// to the target window beneath it.
type BusyWin struct {
	overlay xlib.Window
	display *xlib.Display
	target  *window.Window
}

// Hold creates an InputOnly overlay covering the target window,
// preventing all mouse and keyboard interaction.
func Hold(app widget.AppContext, target *window.Window) *BusyWin {
	if target == nil || target.XWindow == xlib.Window(0) {
		return nil
	}

	d := target.Display.XDisplay
	w := uint(target.Width)
	h := uint(target.Height)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	attrs := &xlib.WindowAttributes{
		EventMask: int64(
			xlib.KeyPressMask |
				xlib.KeyReleaseMask |
				xlib.ButtonPressMask |
				xlib.ButtonReleaseMask |
				xlib.PointerMotionMask),
	}

	overlay := d.CreateWindow(
		target.XWindow,
		0, 0, w, h, 0,
		0,             // depth=0 for InputOnly
		xlib.InputOnly,
		nil, // visual=nil for InputOnly
		xlib.CWEventMask,
		attrs,
	)

	d.RaiseWindow(overlay)
	d.MapWindow(overlay)
	d.Flush()

	// Set the busy cursor (watch cursor).
	// X11 cursor font index 150 = "watch"
	d.DefineCursorFromFont(overlay, 150)

	return &BusyWin{
		overlay: overlay,
		display: d,
		target:  target,
	}
}

// Release destroys the overlay window, restoring interaction.
func (b *BusyWin) Release() {
	if b == nil || b.overlay == xlib.Window(0) {
		return
	}
	b.display.DestroyWindow(b.overlay)
	b.overlay = xlib.Window(0)
	b.display.Flush()
}
