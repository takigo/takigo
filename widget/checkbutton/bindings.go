// Checkbutton event bindings, porting library/button.tcl behavior.

package checkbutton

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

func bindCheckbutton(c *Checkbutton, app widget.AppContext) {
	w := c.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			c.Display()
		}
	})

	// Enter → active.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		c.State = widget.StateActive
		c.Display()
	})

	// Leave → normal.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		c.State = widget.StateNormal
		c.SetPressed(false)
		c.Display()
	})

	// Button1 press.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			c.SetPressed(true)
		}
	})

	// Button1 release → toggle.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if c.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 && c.Pressed() {
			c.SetPressed(false)
			if ev.X >= 0 && ev.X < w.Width && ev.Y >= 0 && ev.Y < w.Height {
				c.Toggle()
			}
		}
	})

	// FocusIn/FocusOut — track focus state and redraw highlight.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		c.SetFocused(ev.Type == event.FocusInType)
		c.Display()
	})

	// Space key → toggle.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			c.Toggle()
		}
	})
}
