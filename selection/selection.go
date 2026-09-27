// Package selection implements selection handling (copy/paste).
// It ports tk/generic/tkSelect.c and tk/generic/tkClipboard.c.
package selection

import (
	"sync"
	"time"

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
	// pendingGet holds, per requestor, the clipboard request awaiting its
	// SelectionNotify; at most one is outstanding, as in Tk.
	pendingGet map[platform.WindowID]*pendingRequest
	nextGen    uint64
	after      func(time.Duration, func())

	// Atoms.
	clipboard platform.AtomID
	utf8str   platform.AtomID
	targets   platform.AtomID
}

// requestTimeout is how long a clipboard owner may stay silent before the
// request fails, like the 5 idle seconds of tkUnixSelect.c SelTimeoutProc.
const requestTimeout = 5 * time.Second

// pendingRequest is an outstanding clipboard request and everyone waiting
// for its answer.
type pendingRequest struct {
	gen       uint64
	callbacks []func(string)
}

// NewManager creates a new selection manager.
func NewManager(server platform.DisplayServer, dispatcher *event.Dispatcher) *Manager {
	m := &Manager{
		server:     server,
		dispatcher: dispatcher,
		data:       make(map[platform.AtomID]string),
		owners:     make(map[platform.AtomID]platform.WindowID),
		pendingGet: make(map[platform.WindowID]*pendingRequest),
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
	m.Own(m.server.Atoms().Primary, owner, content, time)
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
			m.server.Atoms().String,
			m.targets,
		})
	} else if target == m.utf8str || target == m.server.Atoms().String {
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

// SetTimer installs the scheduler used to time out clipboard requests
// (the event loop's After). Without one, requests wait indefinitely.
func (m *Manager) SetTimer(after func(time.Duration, func())) {
	m.mu.Lock()
	m.after = after
	m.mu.Unlock()
}

// RequestWithCallback retrieves CLIPBOARD content. If we own it locally the
// callback is invoked synchronously. Otherwise an async XConvertSelection
// request is sent; the caller must handle SelectionNotify and call
// HandleSelectionNotify to deliver the result. A request made while
// another from the same window is outstanding shares its answer. If the
// owner does not answer within requestTimeout, callbacks receive "".
func (m *Manager) RequestWithCallback(requestor platform.WindowID, ts platform.Timestamp, callback func(string)) {
	m.mu.Lock()
	if content, ok := m.data[m.clipboard]; ok {
		m.mu.Unlock()
		callback(content)
		return
	}
	if p := m.pendingGet[requestor]; p != nil {
		p.callbacks = append(p.callbacks, callback)
		m.mu.Unlock()
		return
	}
	m.nextGen++
	gen := m.nextGen
	m.pendingGet[requestor] = &pendingRequest{gen: gen, callbacks: []func(string){callback}}
	after := m.after
	m.mu.Unlock()

	m.Request(m.clipboard, requestor, ts)
	if after != nil {
		after(requestTimeout, func() { m.expire(requestor, gen) })
	}
}

// expire fails request gen of requestor if it is still unanswered; a
// reply arriving afterwards finds nothing pending and is ignored.
func (m *Manager) expire(requestor platform.WindowID, gen uint64) {
	m.mu.Lock()
	p := m.pendingGet[requestor]
	if p == nil || p.gen != gen {
		m.mu.Unlock()
		return
	}
	delete(m.pendingGet, requestor)
	m.mu.Unlock()
	for _, cb := range p.callbacks {
		cb("")
	}
}

// HandleSelectionNotify is called when a SelectionNotify event arrives for a
// window that previously called RequestWithCallback. It reads the property,
// fires the pending callback, and returns true if a callback was pending.
func (m *Manager) HandleSelectionNotify(requestor platform.WindowID, property platform.AtomID) bool {
	m.mu.Lock()
	p := m.pendingGet[requestor]
	delete(m.pendingGet, requestor)
	m.mu.Unlock()
	if p == nil {
		return false
	}
	var text string
	if property != 0 {
		text = m.ReadProperty(requestor, property)
	}
	for _, cb := range p.callbacks {
		cb(text)
	}
	return true
}

// UTF8StringAtom returns the UTF8_STRING atom.
func (m *Manager) UTF8StringAtom() platform.AtomID {
	return m.utf8str
}
