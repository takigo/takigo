// Menu event bindings, porting tk/library/menu.tcl.
package menu

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
)

func bindMenu(m *Menu, app widget.AppContext) {
	w := m.Win

	// Expose.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		m.Display()
	})

	// Motion → activate entry under pointer.
	app.Dispatcher().Bind(w.XWindow, event.MotionMask, func(ev *event.Event) {
		idx := m.entryAtY(ev.Y)
		m.activate(idx)
	})

	// Enter → activate.
	app.Dispatcher().Bind(w.XWindow, event.EnterMask, func(ev *event.Event) {
		idx := m.entryAtY(ev.Y)
		m.activate(idx)
	})

	// Leave → deactivate.
	app.Dispatcher().Bind(w.XWindow, event.LeaveMask, func(ev *event.Event) {
		m.activate(-1)
	})

	// Button release → invoke.
	app.Dispatcher().Bind(w.XWindow, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		idx := m.entryAtY(ev.Y)
		if idx >= 0 {
			m.invoke(idx)
		} else {
			// Click outside entries — unpost.
			m.Unpost()
		}
	})

	// Button press outside menu → unpost.
	app.Dispatcher().Bind(w.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		// If click is outside menu bounds, unpost.
		if ev.X < 0 || ev.X >= w.Width || ev.Y < 0 || ev.Y >= w.Height {
			m.Unpost()
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(w.XWindow, event.KeyPressMask, func(ev *event.Event) {
		switch ev.KeySym {
		case xlib.XK_Escape:
			m.Unpost()

		case xlib.XK_Up:
			moveActiveEntry(m, -1)

		case xlib.XK_Down:
			moveActiveEntry(m, 1)

		case xlib.XK_Right:
			// Enter cascade submenu.
			if m.activeIndex >= 0 && m.activeIndex < len(m.entries) {
				e := &m.entries[m.activeIndex]
				if e.Type == Cascade && e.SubMenu != nil {
					m.postCascade(m.activeIndex)
				}
			}

		case xlib.XK_Left:
			// Close cascade (parent will handle this via unpost).
			m.Unpost()

		case xlib.XK_Return:
			if m.activeIndex >= 0 {
				m.invoke(m.activeIndex)
			}
		}
	})

	_ = app
}

func moveActiveEntry(m *Menu, direction int) {
	if len(m.entries) == 0 {
		return
	}

	start := m.activeIndex
	if start < 0 {
		if direction > 0 {
			start = -1
		} else {
			start = len(m.entries)
		}
	}

	idx := start
	for {
		idx += direction
		if idx < 0 {
			idx = len(m.entries) - 1
		}
		if idx >= len(m.entries) {
			idx = 0
		}
		// Skip separators and disabled entries.
		e := &m.entries[idx]
		if e.Type != Separator && e.State != widget.StateDisabled {
			m.activate(idx)
			return
		}
		if idx == start {
			return // wrapped around, nothing to activate
		}
	}
}
