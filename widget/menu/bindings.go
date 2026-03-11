// Menu event bindings, porting tk/library/menu.tcl.
package menu

import (
	"unicode"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// Focus mode / detail constants (mirrors platform package values for readability).
const (
	focusModeNormal   = platform.FocusModeNormal
	focusDetailInferior = platform.FocusDetailInferior
	focusDetailPointer  = platform.FocusDetailPointer
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

	// Motion → activate entry under pointer; also marks that pointer has moved
	// since the menu was posted (so the first ButtonRelease is not spuriously ignored).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		m.motionSincePost = true
		idx := m.entryAt(ev.X, ev.Y)
		m.activate(idx)
	})

	// Enter → activate.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		idx := m.entryAt(ev.X, ev.Y)
		m.activate(idx)
	})

	// Leave → deactivate.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		m.activate(-1)
	})

	// FocusOut → unpost when focus leaves to another application.
	// Ignored when: triggered by us posting a cascade submenu (suppressFocusOut),
	// or when it's a pointer/inferior detail (synthetic / child-window focus).
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type != event.FocusOutType {
			return
		}
		if m.suppressFocusOut {
			m.suppressFocusOut = false
			return
		}
		// Ignore grab-induced and synthetic focus changes.
		if ev.FocusMode != focusModeNormal {
			return
		}
		if ev.FocusDetail == focusDetailInferior || ev.FocusDetail == focusDetailPointer {
			return
		}
		m.unpostChain()
	})

	// Button release → invoke.
	// Ignore the very first release after posting (the button-up from the
	// click that opened the menu) unless the pointer has moved since then.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		if !m.motionSincePost {
			// First release with no motion: stay open (click-to-open mode).
			return
		}
		idx := m.entryAt(ev.X, ev.Y)
		if idx == -2 {
			// Tearoff grip clicked — detach the menu.
			m.Detach()
		} else if idx >= 0 {
			m.invoke(idx)
		} else {
			// Click outside entries — unpost entire chain.
			m.unpostChain()
		}
	})

	// Button press outside menu → unpost.
	// With owner_events=true, clicks outside all client windows are still routed
	// here (grab window). Use RootX/RootY to confirm the click is truly outside.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.RootX < m.screenX || ev.RootX >= m.screenX+w.Width ||
			ev.RootY < m.screenY || ev.RootY >= m.screenY+w.Height {
			m.unpostChain()
		}
	})

	// BindGlobal: close menu when user clicks any other widget inside our app.
	// With owner_events=true, clicks on our own widgets are delivered normally,
	// so the per-window handler above never fires for them.  BindGlobal fills
	// this gap.  We skip the very first ButtonPress (the click that opened the
	// menu) via skipGlobalButtonPress, and ignore clicks on the menu itself or
	// on an active cascade submenu.
	app.Dispatcher().BindGlobal(event.ButtonPressMask, func(ev *event.Event) {
		if !m.posted {
			return
		}
		if m.skipGlobalButtonPress {
			m.skipGlobalButtonPress = false
			return
		}
		// If a cascade is posted, let it handle its own closure.
		if m.postedCascade != nil && m.postedCascade.IsPosted() {
			return
		}
		// Don't unpost if the click is within our own menu window.
		if ev.RootX >= m.screenX && ev.RootX < m.screenX+w.Width &&
			ev.RootY >= m.screenY && ev.RootY < m.screenY+w.Height {
			return
		}
		m.unpostChain()
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
