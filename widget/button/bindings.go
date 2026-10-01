// Button event bindings, porting library/button.tcl behavior.

package button

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// bindButton sets up the standard button event bindings:
// - Enter → active state, highlight
// - Leave → normal state
// - Button1 press → sunken relief
// - Button1 release → invoke if still inside
func bindButton(b *Button, app widget.AppContext) {
	w := b.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		b.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			b.Display()
		}
	})

	// Enter → active.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		if b.State == widget.StateDisabled {
			return
		}
		b.State = widget.StateActive
		b.Display()
	})

	// Leave → normal.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if b.State == widget.StateDisabled {
			return
		}
		b.State = widget.StateNormal
		b.pressed = false
		b.Display()
	})

	// Button1 press → sunken.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if b.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			b.pressed = true
			b.Display()
		}
	})

	// Button1 release → invoke.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if b.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 && b.pressed {
			b.pressed = false
			b.Display()
			// Only invoke if the pointer is still inside the button.
			if ev.X >= 0 && ev.X < w.Width && ev.Y >= 0 && ev.Y < w.Height {
				b.Invoke()
			}
		}
	})

	// Space key → invoke (matches Tk's "bind Button <space>" binding).
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			b.Invoke()
		}
	})
}
