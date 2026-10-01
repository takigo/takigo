// Package selection implements selection handling (copy/paste).
// It ports tk/generic/tkSelect.c and tk/generic/tkClipboard.c.
package selection

import (
	"encoding/binary"
	"slices"
	"sync"
	"time"
	"unicode/utf8"

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

	// INCR transfers (ICCCM 2.7.2): outgoing ones we serve as owner, keyed
	// by requestor window and property; incoming ones by our requestor.
	outgoing map[incrKey]*outgoingIncr
	incoming map[platform.WindowID]*incomingIncr

	// formats holds what OwnFormats offers besides text, by selection and
	// target.
	formats map[platform.AtomID]map[platform.AtomID][]byte

	// Atoms.
	clipboard platform.AtomID
	utf8str   platform.AtomID
	targets   platform.AtomID
	incr      platform.AtomID
}

// selBytesAtOnce is Tk's TK_SEL_BYTES_AT_ONCE: longer selections are sent
// with INCR, in chunks of this size.
const selBytesAtOnce = 4000

type incrKey struct {
	window   platform.WindowID
	property platform.AtomID
}

// outgoingIncr is a selection we send in chunks; each chunk is written
// when the requestor deletes the previous one.
type outgoingIncr struct {
	typ  platform.AtomID
	data []byte
	sent bool // the terminating zero-length chunk has been written
}

// incomingIncr is a selection we receive in chunks.
type incomingIncr struct {
	property  platform.AtomID
	typ       platform.AtomID
	buf       []byte
	callbacks []func(string)
	gen       uint64
}

// requestTimeout is how long a clipboard owner may stay silent before the
// request fails, like the 5 idle seconds of tkUnixSelect.c SelTimeoutProc.
const requestTimeout = 5 * time.Second

// pendingRequest is an outstanding clipboard request and everyone waiting
// for its answer.
type pendingRequest struct {
	gen               uint64
	selection, target platform.AtomID
	callbacks         []func(string)
}

// NewManager creates a new selection manager.
func NewManager(server platform.DisplayServer, dispatcher *event.Dispatcher) *Manager {
	m := &Manager{
		server:     server,
		dispatcher: dispatcher,
		data:       make(map[platform.AtomID]string),
		formats:    make(map[platform.AtomID]map[platform.AtomID][]byte),
		owners:     make(map[platform.AtomID]platform.WindowID),
		pendingGet: make(map[platform.WindowID]*pendingRequest),
		clipboard:  server.InternAtom("CLIPBOARD", false),
		utf8str:    server.InternAtom("UTF8_STRING", false),
		targets:    server.InternAtom("TARGETS", false),
		incr:       server.InternAtom("INCR", false),
		outgoing:   make(map[incrKey]*outgoingIncr),
		incoming:   make(map[platform.WindowID]*incomingIncr),
	}

	return m
}

// Own claims ownership of a selection and stores content.
func (m *Manager) Own(selection platform.AtomID, owner platform.WindowID, content string, time platform.Timestamp) {
	m.mu.Lock()
	m.data[selection] = content
	m.owners[selection] = owner
	delete(m.formats, selection)
	m.mu.Unlock()

	m.server.SetSelectionOwner(selection, owner, time)
	if selection == m.clipboard {
		m.server.SetClipboardText(content)
	}
}

// OwnFormats claims a selection and offers data in the given targets
// (atoms such as "image/png" or "text/uri-list") instead of text. It is an
// X11 mechanism: the native clipboards of Windows and macOS only get text.
func (m *Manager) OwnFormats(selection platform.AtomID, owner platform.WindowID, formats map[platform.AtomID][]byte, time platform.Timestamp) {
	m.mu.Lock()
	m.data[selection] = ""
	m.owners[selection] = owner
	m.formats[selection] = formats
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
	formats := m.formats[selection]
	m.mu.Unlock()

	if !ok {
		// We don't own this selection — refuse.
		m.server.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	if property == 0 {
		property = target
	}

	if formats != nil {
		// Offered with OwnFormats: exactly those targets.
		data, have := formats[target]
		switch {
		case target == m.targets:
			atoms := []platform.AtomID{m.targets}
			for t := range formats {
				atoms = append(atoms, t)
			}
			slices.Sort(atoms)
			m.server.ChangePropertyAtoms(requestor, property, atoms)
		case !have:
			m.server.SendSelectionNotify(requestor, selection, target, 0, time)
			return
		case len(data) > selBytesAtOnce:
			m.startIncr(requestor, property, target, data)
		default:
			m.server.ChangeProperty(requestor, property, target, 8, platform.PropModeReplace, data, len(data))
		}
		m.server.SendSelectionNotify(requestor, selection, target, property, time)
		return
	}

	if target == m.targets {
		// Respond with supported targets.
		m.server.ChangePropertyAtoms(requestor, property, []platform.AtomID{
			m.utf8str,
			m.server.Atoms().String,
			m.targets,
		})
	} else if target == m.utf8str || target == m.server.Atoms().String {
		data := []byte(content)
		if target == m.server.Atoms().String {
			data = toLatin1(content)
		}
		if len(data) > selBytesAtOnce {
			m.startIncr(requestor, property, target, data)
		} else {
			m.server.ChangeProperty(requestor, property, target, 8, platform.PropModeReplace, data, len(data))
		}
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
	delete(m.formats, selection)
	m.mu.Unlock()
}

// Request requests the content of a selection from its current owner.
// The result comes back as a SelectionNotify event with the data in a property.
func (m *Manager) Request(selection platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	m.convert(selection, m.utf8str, requestor, time)
}

func (m *Manager) convert(selection, target platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	property := m.server.InternAtom("TAKIGO_SEL", false)
	m.server.ConvertSelection(selection, target, property, requestor, time)
}

// RequestTarget asks the owner of selection for its data in the given
// target and calls callback with the bytes, or with nil if the owner
// refuses, does not answer within requestTimeout, or another request from
// the same window for something else is still outstanding. The caller
// must route SelectionNotify to HandleSelectionNotify, as for
// RequestWithCallback.
func (m *Manager) RequestTarget(selection, target platform.AtomID, requestor platform.WindowID, ts platform.Timestamp, callback func([]byte)) {
	deliver := func(s string) {
		if s == "" {
			callback(nil)
			return
		}
		callback([]byte(s))
	}
	m.mu.Lock()
	if formats, own := m.formats[selection]; own {
		data := formats[target]
		m.mu.Unlock()
		callback(data)
		return
	}
	if p := m.pendingGet[requestor]; p != nil {
		if p.selection != selection || p.target != target {
			m.mu.Unlock()
			callback(nil)
			return
		}
		p.callbacks = append(p.callbacks, deliver)
		m.mu.Unlock()
		return
	}
	m.nextGen++
	gen := m.nextGen
	m.pendingGet[requestor] = &pendingRequest{gen: gen, selection: selection, target: target, callbacks: []func(string){deliver}}
	after := m.after
	m.mu.Unlock()

	m.convert(selection, target, requestor, ts)
	if after != nil {
		after(requestTimeout, func() { m.expire(requestor, gen) })
	}
}

// maxPropWords bounds a single property read, in 32-bit units (4 MiB).
const maxPropWords = 1024 * 1024

// ReadProperty reads the result of a selection request from a window property.
func (m *Manager) ReadProperty(w platform.WindowID, property platform.AtomID) string {
	data, typ, _ := m.server.GetWindowProperty(w, property, 0, maxPropWords, true)
	return m.decode(data, typ)
}

// decode converts selection bytes of type typ to a Go string: STRING is
// ISO 8859-1 (ICCCM), anything else is taken as UTF-8.
func (m *Manager) decode(data []byte, typ platform.AtomID) string {
	if typ == m.server.Atoms().String {
		return fromLatin1(data)
	}
	return string(data)
}

// startIncr begins an INCR transfer of data to requestor: the property
// gets type INCR and the total size, and each chunk follows once the
// requestor deletes the previous one (tkUnixSelect.c ConvertSelection).
func (m *Manager) startIncr(requestor platform.WindowID, property, typ platform.AtomID, data []byte) {
	m.mu.Lock()
	m.outgoing[incrKey{requestor, property}] = &outgoingIncr{typ: typ, data: data}
	m.mu.Unlock()
	m.server.SelectInput(requestor, platform.PropertyChangeMask)
	size := make([]byte, 4)
	binary.NativeEndian.PutUint32(size, uint32(len(data)))
	m.server.ChangeProperty(requestor, property, m.incr, 32, platform.PropModeReplace, size, 1)
}

// startIncoming waits for the chunks of an INCR transfer to requestor.
func (m *Manager) startIncoming(requestor platform.WindowID, property platform.AtomID, p *pendingRequest) {
	m.mu.Lock()
	m.nextGen++
	in := &incomingIncr{property: property, callbacks: p.callbacks, gen: m.nextGen}
	m.incoming[requestor] = in
	m.mu.Unlock()
	m.armIncomingTimeout(requestor, in.gen)
}

// armIncomingTimeout fails an INCR transfer that stalls for requestTimeout;
// each chunk re-arms it, as Tk resets retrPtr->idleTime.
func (m *Manager) armIncomingTimeout(requestor platform.WindowID, gen uint64) {
	m.mu.Lock()
	after := m.after
	m.mu.Unlock()
	if after == nil {
		return
	}
	after(requestTimeout, func() {
		m.mu.Lock()
		in := m.incoming[requestor]
		if in == nil || in.gen != gen {
			m.mu.Unlock()
			return
		}
		delete(m.incoming, requestor)
		m.mu.Unlock()
		for _, cb := range in.callbacks {
			cb("")
		}
	})
}

// HandlePropertyNotify drives INCR transfers: as owner, a requestor deleting
// our property asks for the next chunk; as requestor, a new value on our
// property is the next chunk, and an empty one ends the transfer. It
// returns true if the event belonged to a transfer.
func (m *Manager) HandlePropertyNotify(w platform.WindowID, property platform.AtomID, deleted bool) bool {
	if deleted {
		return m.sendNextChunk(w, property)
	}
	m.mu.Lock()
	in := m.incoming[w]
	if in == nil || in.property != property {
		m.mu.Unlock()
		return false
	}
	m.mu.Unlock()

	data, typ, _ := m.server.GetWindowProperty(w, property, 0, maxPropWords, true)
	if typ == 0 {
		return true
	}
	m.mu.Lock()
	if len(data) > 0 {
		in.buf = append(in.buf, data...)
		in.typ = typ
		m.nextGen++
		in.gen = m.nextGen
		gen := in.gen
		m.mu.Unlock()
		m.armIncomingTimeout(w, gen)
		return true
	}
	delete(m.incoming, w)
	m.mu.Unlock()
	text := m.decode(in.buf, in.typ)
	for _, cb := range in.callbacks {
		cb(text)
	}
	return true
}

// sendNextChunk writes the next chunk of an outgoing INCR transfer, or
// the terminating empty chunk, after the requestor deleted the property.
func (m *Manager) sendNextChunk(requestor platform.WindowID, property platform.AtomID) bool {
	key := incrKey{requestor, property}
	m.mu.Lock()
	out := m.outgoing[key]
	if out == nil {
		m.mu.Unlock()
		return false
	}
	if out.sent {
		delete(m.outgoing, key)
		m.mu.Unlock()
		m.server.SelectInput(requestor, 0)
		return true
	}
	chunk := out.data[:min(selBytesAtOnce, len(out.data))]
	out.data = out.data[len(chunk):]
	out.sent = len(chunk) == 0
	m.mu.Unlock()
	m.server.ChangeProperty(requestor, property, out.typ, 8, platform.PropModeReplace, chunk, len(chunk))
	return true
}

// toLatin1 encodes s as ISO 8859-1 for the STRING target; characters
// outside Latin-1 become '?'.
func toLatin1(s string) []byte {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xff {
			r = '?'
		}
		b = append(b, byte(r))
	}
	return b
}

// fromLatin1 decodes ISO 8859-1 bytes.
func fromLatin1(b []byte) string {
	buf := make([]byte, 0, len(b))
	for _, c := range b {
		buf = utf8.AppendRune(buf, rune(c))
	}
	return string(buf)
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
	// A native clipboard holds what was last copied by any application,
	// including our own Own calls.
	if text, ok := m.server.ClipboardText(); ok {
		callback(text)
		return
	}
	m.mu.Lock()
	if content, ok := m.data[m.clipboard]; ok {
		m.mu.Unlock()
		callback(content)
		return
	}
	if p := m.pendingGet[requestor]; p != nil {
		if p.selection != m.clipboard || p.target != m.utf8str {
			// A request for something else is outstanding on this window.
			m.mu.Unlock()
			callback("")
			return
		}
		p.callbacks = append(p.callbacks, callback)
		m.mu.Unlock()
		return
	}
	m.nextGen++
	gen := m.nextGen
	m.pendingGet[requestor] = &pendingRequest{gen: gen, selection: m.clipboard, target: m.utf8str, callbacks: []func(string){callback}}
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
		data, typ, _ := m.server.GetWindowProperty(requestor, property, 0, maxPropWords, true)
		if typ != 0 && typ == m.incr {
			// The owner sends the text in chunks; reading (and so deleting)
			// the INCR property asked for the first one.
			m.startIncoming(requestor, property, p)
			return true
		}
		text = m.decode(data, typ)
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
