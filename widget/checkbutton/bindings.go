// Checkbutton event bindings, porting library/button.tcl behavior.
package checkbutton

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/widget"
)

func bindCheckbutton(c *Checkbutton, app widget.AppContext) {
	w := c.Win

	// Expose.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			c.Display()
		}
	})

	// Enter → active.
	app.Dispatcher().Bind(w.XWindow, event.EnterMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		c.State = widget.StateActive
		c.Display()
	})

	// Leave → normal.
	app.Dispatcher().Bind(w.XWindow, event.LeaveMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		c.State = widget.StateNormal
		c.pressed = false
		c.Display()
	})

	// Button1 press.
	app.Dispatcher().Bind(w.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			c.pressed = true
		}
	})

	// Button1 release → toggle.
	app.Dispatcher().Bind(w.XWindow, event.ButtonReleaseMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 && c.pressed {
			c.pressed = false
			if ev.X >= 0 && ev.X < w.Width && ev.Y >= 0 && ev.Y < w.Height {
				c.Toggle()
			}
		}
	})
}
