package geometry

import "github.com/msorc/takigo/window"

// WhenIdle runs arrange once the event loop is idle, however many times
// it is requested before then, as Tk's geometry managers do with
// Tcl_DoWhenIdle and a REQUESTED_RELAYOUT flag. pending is the caller's
// per-container flag. Without an idle scheduler arrange runs at once.
func WhenIdle(container *window.Window, pending *bool, arrange func()) {
	d := container.Display
	if d == nil || d.DoWhenIdle == nil {
		arrange()
		return
	}
	if *pending {
		return
	}
	*pending = true
	d.DoWhenIdle(func() {
		*pending = false
		arrange()
	})
}
