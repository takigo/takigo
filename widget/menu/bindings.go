// Menu event bindings, porting tk/library/menu.tcl.
package menu

import (
	"unicode"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

func bindMenu(m *Menu, app widget.AppContext) {
	w := m.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		m.Display()
	})

	// Motion → activate entry under pointer.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		idx := m.entryAtY(ev.Y)
		m.activate(idx)
	})

	// Enter → activate.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		idx := m.entryAtY(ev.Y)
		m.activate(idx)
	})

	// Leave → deactivate.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		m.activate(-1)
	})

	// Button release → invoke.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		idx := m.entryAtY(ev.Y)
		if idx == -2 {
			// Tearoff grip clicked — detach the menu.
			m.Detach()
		} else if idx >= 0 {
			m.invoke(idx)
		} else {
			// Click outside entries — unpost.
			m.Unpost()
		}
	})

	// Button press outside menu → unpost.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		// If click is outside menu bounds, unpost.
		if ev.X < 0 || ev.X >= w.Width || ev.Y < 0 || ev.Y >= w.Height {
			m.Unpost()
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		ks := ev.KeySym
		switch {
		case ks == platform.XK_Escape:
			m.Unpost()

		case ks == platform.XK_Up:
			moveActiveEntry(m, -1)

		case ks == platform.XK_Down:
			moveActiveEntry(m, 1)

		case ks == platform.XK_Right:
			// Enter cascade submenu.
			if m.activeIndex >= 0 && m.activeIndex < len(m.entries) {
				e := &m.entries[m.activeIndex]
				if e.Type == Cascade && e.SubMenu != nil {
					m.postCascade(m.activeIndex)
				}
			}

		case ks == platform.XK_Left:
			// Close cascade (parent will handle this via unpost).
			m.Unpost()

		case ks == platform.XK_Return:
			if m.activeIndex >= 0 {
				m.invoke(m.activeIndex)
			}

		default:
			// Letter navigation: match underlined character (or first char as fallback).
			r := platform.KeySymToRune(ks)
			if r <= 0 {
				return
			}
			r = unicode.ToLower(r)
			// Search from entry after the active one, wrapping around.
			start := m.activeIndex + 1
			if start < 0 {
				start = 0
			}
			n := len(m.entries)
			for i := 0; i < n; i++ {
				idx := (start + i) % n
				e := &m.entries[idx]
				if e.Type == Separator || e.State == widget.StateDisabled {
					continue
				}
				runes := []rune(e.Label)
				if len(runes) == 0 {
					continue
				}
				var matchRune rune
				if e.Underline >= 0 && e.Underline < len(runes) {
					matchRune = unicode.ToLower(runes[e.Underline])
				} else {
					matchRune = unicode.ToLower(runes[0])
				}
				if matchRune == r {
					m.invoke(idx)
					return
				}
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
