// Package focus manages keyboard focus and Tab/Shift-Tab traversal.
// It ports tk/generic/tkFocus.c and tk/library/focus.tcl.
package focus

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/window"
)

// FocusableChecker determines if a widget can receive focus.
type FocusableChecker func(w *window.Window) bool

// Manager tracks focus state across toplevels and handles traversal.
type Manager struct {
	dispatcher *event.Dispatcher
	display    *xlib.Display

	// Per-toplevel focus: which widget last had focus in each toplevel.
	toplevelFocus map[*window.Window]*window.Window

	// Currently focused widget on the display.
	focusWin *window.Window

	// Callback to check if a window is focusable.
	IsFocusable FocusableChecker

	// Callback invoked when focus changes (for widgets to redraw).
	OnFocusChange func(lost, gained *window.Window)
}

// NewManager creates a new focus manager.
func NewManager(dispatcher *event.Dispatcher, display *xlib.Display) *Manager {
	m := &Manager{
		dispatcher:    dispatcher,
		display:       display,
		toplevelFocus: make(map[*window.Window]*window.Window),
		IsFocusable: func(w *window.Window) bool {
			// Default: all mapped non-toplevel windows are focusable.
			return w.IsMapped() || w.XWindow != xlib.Window(0)
		},
	}
	return m
}

// FocusWindow returns the widget that currently has focus, or nil.
func (m *Manager) FocusWindow() *window.Window {
	return m.focusWin
}

// SetFocus moves focus to the given window.
func (m *Manager) SetFocus(w *window.Window) {
	if w == nil || w.XWindow == xlib.Window(0) {
		return
	}

	old := m.focusWin
	if old == w {
		return
	}

	// Update per-toplevel tracking.
	tl := findToplevel(w)
	if tl != nil {
		m.toplevelFocus[tl] = w
	}

	m.focusWin = w

	// Notify widgets of focus change.
	if m.OnFocusChange != nil {
		m.OnFocusChange(old, w)
	}

	// Generate FocusOut/FocusIn events.
	if old != nil {
		m.dispatcher.Dispatch(&event.Event{
			Type:   event.FocusOutType,
			Window: old.XWindow,
		})
	}
	m.dispatcher.Dispatch(&event.Event{
		Type:   event.FocusInType,
		Window: w.XWindow,
	})

	// Tell X to direct keyboard input to this window's toplevel.
	if tl != nil && tl.XWindow != xlib.Window(0) {
		m.display.SetInputFocus(tl.XWindow, xlib.RevertToParent, xlib.CurrentTime)
	}
}

// HandleFocusIn processes an X FocusIn event on a toplevel.
// It restores focus to the last focused widget within that toplevel.
func (m *Manager) HandleFocusIn(w *window.Window) {
	tl := findToplevel(w)
	if tl == nil {
		tl = w
	}

	// Restore the remembered focus widget for this toplevel.
	if remembered, ok := m.toplevelFocus[tl]; ok && remembered != nil {
		m.SetFocus(remembered)
	} else {
		// Default to the toplevel itself.
		m.SetFocus(tl)
	}
}

// HandleFocusOut processes an X FocusOut event.
func (m *Manager) HandleFocusOut(w *window.Window) {
	if m.focusWin != nil {
		old := m.focusWin
		m.focusWin = nil
		if m.OnFocusChange != nil {
			m.OnFocusChange(old, nil)
		}
		m.dispatcher.Dispatch(&event.Event{
			Type:   event.FocusOutType,
			Window: old.XWindow,
		})
	}
}

// FocusNext moves focus to the next focusable widget (Tab key).
func (m *Manager) FocusNext() {
	current := m.focusWin
	if current == nil {
		return
	}
	tl := findToplevel(current)
	if tl == nil {
		return
	}

	next := m.nextFocusable(tl, current, true)
	if next != nil {
		m.SetFocus(next)
	}
}

// FocusPrev moves focus to the previous focusable widget (Shift-Tab).
func (m *Manager) FocusPrev() {
	current := m.focusWin
	if current == nil {
		return
	}
	tl := findToplevel(current)
	if tl == nil {
		return
	}

	next := m.nextFocusable(tl, current, false)
	if next != nil {
		m.SetFocus(next)
	}
}

// nextFocusable finds the next (or previous) focusable widget using
// depth-first pre-order traversal of the widget tree.
func (m *Manager) nextFocusable(toplevel, current *window.Window, forward bool) *window.Window {
	all := flattenTree(toplevel)
	if len(all) == 0 {
		return nil
	}

	// Find current index.
	idx := -1
	for i, w := range all {
		if w == current {
			idx = i
			break
		}
	}

	n := len(all)
	for i := 1; i < n; i++ {
		var candidate *window.Window
		if forward {
			candidate = all[(idx+i)%n]
		} else {
			candidate = all[(idx-i+n)%n]
		}
		if m.IsFocusable(candidate) {
			return candidate
		}
	}

	return nil
}

// flattenTree returns all windows in the tree rooted at w
// in depth-first pre-order.
func flattenTree(w *window.Window) []*window.Window {
	var result []*window.Window
	var walk func(w *window.Window)
	walk = func(w *window.Window) {
		result = append(result, w)
		for _, child := range w.Children {
			// Skip embedded toplevels.
			if child.IsTopLevel() {
				continue
			}
			walk(child)
		}
	}
	walk(w)
	return result
}

// HandleDestroyWindow cleans up focus state when a window is destroyed.
func (m *Manager) HandleDestroyWindow(w *window.Window) {
	if m.focusWin == w {
		tl := findToplevel(w)
		if tl != nil && tl != w {
			m.SetFocus(tl)
		} else {
			m.focusWin = nil
		}
	}

	// Clean up toplevel focus entries.
	for tl, fw := range m.toplevelFocus {
		if fw == w {
			m.toplevelFocus[tl] = tl // reset to toplevel
		}
		if tl == w {
			delete(m.toplevelFocus, tl)
		}
	}
}

// BindTraversal binds Tab and Shift-Tab on a window for focus traversal.
func (m *Manager) BindTraversal(w *window.Window) {
	m.dispatcher.Bind(w.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Tab {
			if ev.State&xlib.ShiftMask != 0 {
				m.FocusPrev()
			} else {
				m.FocusNext()
			}
		}
	})
}

// findToplevel walks up the parent chain to find the toplevel window.
func findToplevel(w *window.Window) *window.Window {
	for w != nil {
		if w.IsTopLevel() {
			return w
		}
		w = w.Parent
	}
	return nil
}
