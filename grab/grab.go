// Package grab provides local and global grab support for modal dialogs.
// It ports the essential subset of tk/generic/tkGrab.c.
package grab

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// GrabState indicates a window's relationship to the current grab.
type GrabState int

const (
	GrabNone     GrabState = iota // no grab active, or window is in a different app
	GrabInTree                    // window is inside the grab subtree
	GrabAncestor                  // window is an ancestor of the grab window
	GrabExcluded                  // window is outside the grab tree
)

// Manager manages grab state for the application.
type Manager struct {
	display    platform.DisplayServer
	dispatcher *event.Dispatcher

	// Current grab window (nil = no grab).
	grabWin *window.Window

	// Whether the grab is global (X server grab) or local (software only).
	grabGlobal bool
}

// NewManager creates a new grab manager.
func NewManager(display platform.DisplayServer, dispatcher *event.Dispatcher) *Manager {
	return &Manager{
		display:    display,
		dispatcher: dispatcher,
	}
}

// Current returns the current grab window, or nil if no grab is active.
func (m *Manager) Current() *window.Window {
	return m.grabWin
}

// IsGlobal returns true if the current grab is a global (X server) grab.
func (m *Manager) IsGlobal() bool {
	return m.grabGlobal
}

// Set activates a grab on the given window.
// If global is true, an X server pointer+keyboard grab is used.
// Otherwise, the grab is enforced in software by the event dispatcher.
func (m *Manager) Set(w *window.Window, global bool) bool {
	if w == nil || w.PlatformID == platform.WindowID(0) {
		return false
	}

	// Release any existing grab first.
	if m.grabWin != nil {
		m.Release()
	}

	m.grabWin = w
	m.grabGlobal = global

	if global {
		// Grab pointer.
		result := m.display.GrabPointer(w.PlatformID, true,
			uint(platform.ButtonPressMask|platform.ButtonReleaseMask|platform.PointerMotionMask|platform.ButtonMotionMask),
			platform.GrabModeAsync, platform.GrabModeAsync,
			platform.WindowID(platform.None), platform.CursorID(0), platform.CurrentTime)
		if result != platform.GrabSuccess {
			m.grabWin = nil
			m.grabGlobal = false
			return false
		}

		// Grab keyboard.
		result = m.display.GrabKeyboard(w.PlatformID, false,
			platform.GrabModeAsync, platform.GrabModeAsync, platform.CurrentTime)
		if result != platform.GrabSuccess {
			m.display.UngrabPointer(platform.CurrentTime)
			m.grabWin = nil
			m.grabGlobal = false
			return false
		}
	}

	return true
}

// Release releases the current grab.
func (m *Manager) Release() {
	if m.grabWin == nil {
		return
	}

	if m.grabGlobal {
		m.display.UngrabPointer(platform.CurrentTime)
		m.display.UngrabKeyboard(platform.CurrentTime)
	}

	m.grabWin = nil
	m.grabGlobal = false
}

// State returns the grab state for a given window relative to the current grab.
func (m *Manager) State(w *window.Window) GrabState {
	if m.grabWin == nil {
		return GrabNone
	}

	if w == m.grabWin {
		return GrabInTree
	}

	// Check if w is a descendant of the grab window.
	for p := w.Parent; p != nil; p = p.Parent {
		if p == m.grabWin {
			return GrabInTree
		}
	}

	// Check if w is an ancestor of the grab window.
	for p := m.grabWin.Parent; p != nil; p = p.Parent {
		if p == w {
			return GrabAncestor
		}
	}

	return GrabExcluded
}

// ShouldRedirect returns true if an event targeting w should be redirected
// to the grab window. For local grabs, pointer and key events outside
// the grab tree are redirected.
func (m *Manager) ShouldRedirect(w *window.Window) bool {
	if m.grabWin == nil {
		return false
	}
	if m.grabGlobal {
		// Global grabs are handled by the X server.
		return false
	}
	return m.State(w) == GrabExcluded
}

// RedirectTarget returns the grab window if events should be redirected,
// or nil if no redirection is needed.
func (m *Manager) RedirectTarget(w *window.Window) *window.Window {
	if m.ShouldRedirect(w) {
		return m.grabWin
	}
	return nil
}
