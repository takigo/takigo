// Radiobutton event bindings, porting library/button.tcl behavior.

package radiobutton

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

func bindRadiobutton(r *Radiobutton, app widget.AppContext) {
	w := r.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		r.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			r.Display()
		}
	})

	// Enter → active.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		if r.State == widget.StateDisabled {
			return
		}
		r.State = widget.StateActive
		r.Display()
	})

	// Leave → normal.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if r.State == widget.StateDisabled {
			return
		}
		r.State = widget.StateNormal
		r.pressed = false
		r.Display()
	})

	// Button1 press.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if r.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			r.pressed = true
		}
	})

	// Button1 release → select.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if r.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 && r.pressed {
			r.pressed = false
			if ev.X >= 0 && ev.X < w.Width && ev.Y >= 0 && ev.Y < w.Height {
				r.Select()
			}
		}
	})

	// FocusIn/FocusOut — track focus state and redraw highlight.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		r.HasFocus = ev.Type == event.FocusInType
		r.Display()
	})

	// Space key → select.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			r.Select()
		}
	})
}
