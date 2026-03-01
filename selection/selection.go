// Package selection implements X11 selection handling (copy/paste).
// It ports tk/generic/tkSelect.c and tk/generic/tkClipboard.c.
package selection

import (
	"sync"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
)

// Manager handles X11 selection ownership and transfers.
type Manager struct {
	display    *xlib.Display
	dispatcher *event.Dispatcher

	mu   sync.Mutex
	data map[xlib.Atom]string // selection atom → content
	// Owner window for each selection.
	owners map[xlib.Atom]xlib.Window

	// Atoms.
	clipboard xlib.Atom
	utf8str   xlib.Atom
	targets   xlib.Atom
}

// NewManager creates a new selection manager.
func NewManager(display *xlib.Display, dispatcher *event.Dispatcher) *Manager {
	m := &Manager{
		display:    display,
		dispatcher: dispatcher,
		data:       make(map[xlib.Atom]string),
		owners:     make(map[xlib.Atom]xlib.Window),
		clipboard:  display.InternAtom("CLIPBOARD", false),
		utf8str:    display.InternAtom("UTF8_STRING", false),
		targets:    display.InternAtom("TARGETS", false),
	}

	// Listen for selection-related events globally.
	dispatcher.BindGlobal(event.AllEventsMask, func(ev *event.Event) {
		switch ev.Type {
		case event.PropertyType:
			// Used for incremental transfers — not implemented yet.
		}
	})

	return m
}

// Own claims ownership of a selection and stores content.
func (m *Manager) Own(selection xlib.Atom, owner xlib.Window, content string, time xlib.Time) {
	m.mu.Lock()
	m.data[selection] = content
	m.owners[selection] = owner
	m.mu.Unlock()

	m.display.SetSelectionOwner(selection, owner, time)
}

// OwnPrimary claims PRIMARY selection.
func (m *Manager) OwnPrimary(owner xlib.Window, content string, time xlib.Time) {
	m.Own(xlib.XA_PRIMARY, owner, content, time)
}

// OwnClipboard claims CLIPBOARD selection.
func (m *Manager) OwnClipboard(owner xlib.Window, content string, time xlib.Time) {
	m.Own(m.clipboard, owner, content, time)
}

// GetContent returns the stored content for a selection.
func (m *Manager) GetContent(selection xlib.Atom) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[selection]
}

// HandleSelectionRequest processes a SelectionRequest event from another client.
// It sends back the selection data via a property and SelectionNotify.
func (m *Manager) HandleSelectionRequest(requestor xlib.Window, selection, target, property xlib.Atom, time xlib.Time) {
	m.mu.Lock()
	content, ok := m.data[selection]
	m.mu.Unlock()

	if !ok {
		// We don't own this selection — refuse.
		m.display.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	if property == 0 {
		property = target
	}

	if target == m.targets {
		// Respond with supported targets.
		m.display.ChangePropertyAtoms(requestor, property, []xlib.Atom{
			m.utf8str,
			xlib.XA_STRING,
			m.targets,
		})
	} else if target == m.utf8str || target == xlib.XA_STRING {
		m.display.ChangePropertyString(requestor, property, target, content)
	} else {
		// Unsupported target — refuse.
		m.display.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	m.display.SendSelectionNotify(requestor, selection, target, property, time)
}

// HandleSelectionClear is called when we lose selection ownership.
func (m *Manager) HandleSelectionClear(selection xlib.Atom) {
	m.mu.Lock()
	delete(m.data, selection)
	delete(m.owners, selection)
	m.mu.Unlock()
}

// Request requests the content of a selection from its current owner.
// The result comes back as a SelectionNotify event with the data in a property.
func (m *Manager) Request(selection xlib.Atom, requestor xlib.Window, time xlib.Time) {
	property := m.display.InternAtom("TAKIGO_SEL", false)
	m.display.ConvertSelection(selection, m.utf8str, property, requestor, time)
}

// ReadProperty reads the result of a selection request from a window property.
func (m *Manager) ReadProperty(w xlib.Window, property xlib.Atom) string {
	data, _, _ := m.display.GetWindowProperty(w, property, 0, 1024*1024, true)
	return string(data)
}

// ClipboardAtom returns the CLIPBOARD atom.
func (m *Manager) ClipboardAtom() xlib.Atom {
	return m.clipboard
}

// UTF8StringAtom returns the UTF8_STRING atom.
func (m *Manager) UTF8StringAtom() xlib.Atom {
	return m.utf8str
}
