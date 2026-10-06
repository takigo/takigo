// Package focus manages keyboard focus and Tab/Shift-Tab traversal.
// It ports tk/generic/tkFocus.c and tk/library/focus.tcl.
package focus

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// FocusableChecker determines if a widget can receive focus.
type FocusableChecker func(w *window.Window) bool

// Manager tracks focus state across toplevels and handles traversal.
type Manager struct {
	dispatcher *event.Dispatcher
	display    platform.DisplayServer
	winDisplay *window.Display // for LookupWindow from event IDs

	// Per-toplevel focus: which widget last had focus in each toplevel.
	toplevelFocus map[*window.Window]*window.Window

	// toplevelReady tracks toplevels that have received a real X FocusIn event,
	// meaning the WM has made them viewable. SetInputFocus is only safe once ready.
	toplevelReady map[*window.Window]bool

	// Currently focused widget on the display.
	focusWin *window.Window

	// Callback to check if a window is focusable.
	IsFocusable FocusableChecker

	// Callback invoked when focus changes (for widgets to redraw).
	OnFocusChange func(lost, gained *window.Window)

	// As in Tk (tkFocus.c) the X input focus stays on toplevels: xFocus
	// is the toplevel SetFocus last gave it to. Keys reaching a toplevel
	// or its children are redirected to focusWin (FilterEvent).
	xFocus platform.WindowID

	// appLost is set once the application's last focused toplevel lost the
	// X focus to another client; SetFocus then only records the widget,
	// which gets the focus back when a toplevel is focused again (focus
	// without -force).
	appLost bool

	// implicit is set while the application has the focus only because
	// the X focus is PointerRoot and the pointer is over one of its
	// toplevels (no window manager); it ends when the pointer leaves.
	implicit bool
}

// NewManager creates a new focus manager.
func NewManager(dispatcher *event.Dispatcher, display platform.DisplayServer, winDisplay *window.Display) *Manager {
	m := &Manager{
		dispatcher:    dispatcher,
		display:       display,
		winDisplay:    winDisplay,
		toplevelFocus: make(map[*window.Window]*window.Window),
		toplevelReady: make(map[*window.Window]bool),
		IsFocusable: func(w *window.Window) bool {
			return w.Flags&window.FlagFocusable != 0 && w.PlatformID != platform.WindowID(0)
		},
	}
	return m
}

// FocusWindow returns the widget that currently has focus, or nil.
func (m *Manager) FocusWindow() *window.Window {
	return m.focusWin
}

// SetFocus moves focus to the given window (Tk's focus command). The
// window becomes its toplevel's focus; if the application has the focus
// it gets FocusIn (and the old focus FocusOut) at once and the X focus
// moves to its toplevel, otherwise it gets the focus when the toplevel is
// next focused.
func (m *Manager) SetFocus(w *window.Window) {
	if w == nil || w.PlatformID == platform.WindowID(0) {
		return
	}
	tl := findToplevel(w)
	if tl != nil {
		m.toplevelFocus[tl] = w
	}
	if m.appLost {
		return
	}
	m.moveTo(w)
	m.claimX(tl)
}

// moveTo makes w the focus window, dispatching FocusOut to the old one and
// FocusIn to w (GenerateFocusEvents).
func (m *Manager) moveTo(w *window.Window) {
	old := m.focusWin
	if old == w {
		return
	}
	m.focusWin = w
	if m.OnFocusChange != nil {
		m.OnFocusChange(old, w)
	}
	if old != nil && old.PlatformID != 0 {
		m.dispatcher.Dispatch(&event.Event{Type: event.FocusOutType, Window: old.PlatformID})
	}
	if w != nil {
		m.dispatcher.Dispatch(&event.Event{Type: event.FocusInType, Window: w.PlatformID})
	}
}

// claimX gives the X focus to toplevel tl, once X has shown it viewable
// (a FocusIn), unless the focus is implicit (focus follows the pointer).
func (m *Manager) claimX(tl *window.Window) {
	if tl == nil || tl.PlatformID == 0 || !m.toplevelReady[tl] || m.implicit || m.xFocus == tl.PlatformID {
		return
	}
	m.xFocus = tl.PlatformID
	m.display.SetInputFocus(tl.PlatformID, platform.RevertToParent, platform.CurrentTime)
}

// FilterEvent applies Tk's focus model to a real event before it is
// dispatched (TkFocusFilterEvent, TkFocusKeyEvent) and reports whether to
// dispatch it:
//   - FocusIn/FocusOut on child windows are dropped; on toplevels they
//     move the application focus and are replaced by FocusIn/FocusOut on
//     the focus windows.
//   - Leaving a toplevel ends an implicit (pointer) focus.
//   - Key events go to the focus window, wherever X delivered them.
func (m *Manager) FilterEvent(ev *event.Event) bool {
	var w *window.Window
	if m.winDisplay != nil {
		w = m.winDisplay.LookupWindow(ev.Window)
	}
	if w == nil {
		return true
	}
	switch ev.Type {
	case event.FocusInType, event.FocusOutType:
		if !w.IsTopLevel() || ev.FocusMode == platform.FocusModeGrab || ev.FocusMode == platform.FocusModeUngrab {
			return false
		}
		switch ev.FocusDetail {
		case platform.FocusDetailVirtual, platform.FocusDetailNonlinearVirtual,
			platform.FocusDetailInferior, platform.FocusDetailPointerRoot:
			return false
		}
		if ev.Type == event.FocusInType {
			m.toplevelFocusIn(w, ev.FocusDetail == platform.FocusDetailPointer)
		} else if ev.FocusDetail != platform.FocusDetailPointer {
			// A FocusOut NotifyPointer comes with our own XSetInputFocus
			// while the focus was implicit; the FocusIn that follows says
			// where it went.
			m.toplevelFocusOut(w)
		}
		return false
	case event.LeaveType:
		if m.implicit && w.IsTopLevel() && ev.FocusDetail != platform.FocusDetailInferior {
			m.toplevelFocusOut(w)
		}
	case event.KeyPressType, event.KeyReleaseType:
		target := m.focusWin
		if target == nil {
			target = m.toplevelFocus[findToplevel(w)]
		}
		if target != nil && target != w && target.PlatformID != 0 && !target.IsDestroyed() {
			if ox, oy, tl := offsetInToplevel(w); tl != nil {
				if tx, ty, ttl := offsetInToplevel(target); ttl == tl {
					ev.X, ev.Y = ev.X+ox-tx, ev.Y+oy-ty
				}
			}
			ev.Window = target.PlatformID
		}
	}
	return true
}

// offsetInToplevel returns the position of w's interior within its
// toplevel's, and the toplevel.
func offsetInToplevel(w *window.Window) (x, y int, tl *window.Window) {
	for ; w != nil && !w.IsTopLevel(); w = w.Parent {
		x += w.X + w.BorderWidth
		y += w.Y + w.BorderWidth
	}
	return x, y, w
}

// toplevelFocusIn handles the X focus arriving at toplevel tl: its
// remembered focus window (or tl) becomes the focus.
func (m *Manager) toplevelFocusIn(tl *window.Window, implicit bool) {
	m.toplevelReady[tl] = true
	m.appLost = false
	m.implicit = implicit
	if !implicit {
		m.xFocus = tl.PlatformID
	}
	w := m.toplevelFocus[tl]
	if w == nil || w.IsDestroyed() {
		w = tl
		m.toplevelFocus[tl] = tl
	}
	m.moveTo(w)
}

// toplevelFocusOut handles toplevel tl losing the X focus: if the focus
// window is in tl it gets FocusOut and the application has no focus.
func (m *Manager) toplevelFocusOut(tl *window.Window) {
	m.implicit = false
	if m.xFocus == tl.PlatformID {
		m.xFocus = 0
	}
	if m.focusWin != nil && findToplevel(m.focusWin) == tl {
		m.appLost = true
		m.moveTo(nil)
	}
}

// FocusNext moves focus to the next focusable widget (Tab key).
// eventWin is the window that received the key event; used to find the
// toplevel when no widget currently has focus.
func (m *Manager) FocusNext(eventWin *window.Window) {
	current := m.focusWin
	if current == nil {
		current = eventWin
	}
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
// eventWin is the window that received the key event; used to find the
// toplevel when no widget currently has focus.
func (m *Manager) FocusPrev(eventWin *window.Window) {
	current := m.focusWin
	if current == nil {
		current = eventWin
	}
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
		if tl != nil && tl != w && !tl.IsDestroyed() {
			m.SetFocus(tl)
		} else {
			m.focusWin = nil
		}
	}

	// Clean up toplevel focus and ready entries.
	for tl, fw := range m.toplevelFocus {
		if fw == w {
			m.toplevelFocus[tl] = tl // reset to toplevel
		}
		if tl == w {
			delete(m.toplevelFocus, tl)
			delete(m.toplevelReady, tl)
		}
	}
}

// BindTraversal binds Tab and Shift-Tab globally for focus traversal.
// Uses a global binding so it works regardless of which widget has focus.
func (m *Manager) BindTraversal(w *window.Window) {
	m.dispatcher.BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.Handled {
			return
		}
		// Look up the window that received the event so we can find
		// the toplevel even when no widget has focus yet.
		var eventWin *window.Window
		if m.winDisplay != nil {
			eventWin = m.winDisplay.LookupWindow(ev.Window)
		}
		switch ev.KeySym {
		case platform.XK_Tab:
			m.FocusNext(eventWin)
		case platform.XK_ISO_Left_Tab:
			m.FocusPrev(eventWin)
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
