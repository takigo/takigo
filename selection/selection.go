// Package selection implements selection handling (copy/paste).
// It ports tk/generic/tkSelect.c and tk/generic/tkClipboard.c.
package selection

import (
	"sync"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
)

// Manager handles selection ownership and transfers.
type Manager struct {
	server     platform.DisplayServer
	dispatcher *event.Dispatcher

	mu   sync.Mutex
	data map[platform.AtomID]string // selection atom → content
	// Owner window for each selection.
	owners map[platform.AtomID]platform.WindowID
	// pendingGet stores callbacks waiting for async SelectionNotify responses.
	pendingGet map[platform.WindowID]func(string)

	// Atoms.
	clipboard platform.AtomID
	utf8str   platform.AtomID
	targets   platform.AtomID
}

// NewManager creates a new selection manager.
func NewManager(server platform.DisplayServer, dispatcher *event.Dispatcher) *Manager {
	m := &Manager{
		server:     server,
		dispatcher: dispatcher,
		data:       make(map[platform.AtomID]string),
		owners:     make(map[platform.AtomID]platform.WindowID),
		pendingGet: make(map[platform.WindowID]func(string)),
		clipboard:  server.InternAtom("CLIPBOARD", false),
		utf8str:    server.InternAtom("UTF8_STRING", false),
		targets:    server.InternAtom("TARGETS", false),
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
func (m *Manager) Own(selection platform.AtomID, owner platform.WindowID, content string, time platform.Timestamp) {
	m.mu.Lock()
	m.data[selection] = content
	m.owners[selection] = owner
	m.mu.Unlock()

	m.server.SetSelectionOwner(selection, owner, time)
}

// OwnPrimary claims PRIMARY selection.
func (m *Manager) OwnPrimary(owner platform.WindowID, content string, time platform.Timestamp) {
	m.Own(platform.XA_PRIMARY, owner, content, time)
}

// OwnClipboard claims CLIPBOARD selection.
func (m *Manager) OwnClipboard(owner platform.WindowID, content string, time platform.Timestamp) {
	m.Own(m.clipboard, owner, content, time)
}

// GetContent returns the stored content for a selection.
func (m *Manager) GetContent(selection platform.AtomID) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.data[selection]
}

// HandleSelectionRequest processes a SelectionRequest event from another client.
// It sends back the selection data via a property and SelectionNotify.
func (m *Manager) HandleSelectionRequest(requestor platform.WindowID, selection, target, property platform.AtomID, time platform.Timestamp) {
	m.mu.Lock()
	content, ok := m.data[selection]
	m.mu.Unlock()

	if !ok {
		// We don't own this selection — refuse.
		m.server.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	if property == 0 {
		property = target
	}

	if target == m.targets {
		// Respond with supported targets.
		m.server.ChangePropertyAtoms(requestor, property, []platform.AtomID{
			m.utf8str,
			platform.XA_STRING,
			m.targets,
		})
	} else if target == m.utf8str || target == platform.XA_STRING {
		m.server.ChangePropertyString(requestor, property, target, content)
	} else {
		// Unsupported target — refuse.
		m.server.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	m.server.SendSelectionNotify(requestor, selection, target, property, time)
}

// HandleSelectionClear is called when we lose selection ownership.
func (m *Manager) HandleSelectionClear(selection platform.AtomID) {
	m.mu.Lock()
	delete(m.data, selection)
	delete(m.owners, selection)
	m.mu.Unlock()
}

// Request requests the content of a selection from its current owner.
// The result comes back as a SelectionNotify event with the data in a property.
func (m *Manager) Request(selection platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	property := m.server.InternAtom("TAKIGO_SEL", false)
	m.server.ConvertSelection(selection, m.utf8str, property, requestor, time)
}

// ReadProperty reads the result of a selection request from a window property.
func (m *Manager) ReadProperty(w platform.WindowID, property platform.AtomID) string {
	data, _, _ := m.server.GetWindowProperty(w, property, 0, 1024*1024, true)
	return string(data)
}

// ClipboardAtom returns the CLIPBOARD atom.
func (m *Manager) ClipboardAtom() platform.AtomID {
	return m.clipboard
}

// RequestWithCallback retrieves CLIPBOARD content. If we own it locally the
// callback is invoked synchronously. Otherwise an async XConvertSelection
// request is sent; the caller must handle SelectionNotify and call
// HandleSelectionNotify to deliver the result.
func (m *Manager) RequestWithCallback(requestor platform.WindowID, time platform.Timestamp, callback func(string)) {
	m.mu.Lock()
	content, ok := m.data[m.clipboard]
	if ok {
		m.mu.Unlock()
		callback(content)
		return
	}
	m.pendingGet[requestor] = callback
	m.mu.Unlock()
	m.Request(m.clipboard, requestor, time)
}

// HandleSelectionNotify is called when a SelectionNotify event arrives for a
// window that previously called RequestWithCallback. It reads the property,
// fires the pending callback, and returns true if a callback was pending.
func (m *Manager) HandleSelectionNotify(requestor platform.WindowID, property platform.AtomID) bool {
	m.mu.Lock()
	cb := m.pendingGet[requestor]
	delete(m.pendingGet, requestor)
	m.mu.Unlock()
	if cb == nil {
		return false
	}
	var text string
	if property != 0 {
		text = m.ReadProperty(requestor, property)
	}
	cb(text)
	return true
}

// UTF8StringAtom returns the UTF8_STRING atom.
func (m *Manager) UTF8StringAtom() platform.AtomID {
	return m.utf8str
}
