// Package selection implements selection handling (copy/paste).
// It ports tk/generic/tkSelect.c and tk/generic/tkClipboard.c.
package selection

import (
	"encoding/binary"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
)

// Manager handles selection ownership and transfers.
type Manager struct {
	server     platform.DisplayServer
	dispatcher *event.Dispatcher

	mu   sync.Mutex
	data map[platform.AtomID]string // selection atom → content
	// Owner window for each selection.
	owners map[platform.AtomID]platform.WindowID
	// pendingGet holds the clipboard requests awaiting their
	// SelectionNotify, keyed by the property each of them uses. Every
	// request gets a property of its own, so that a text and an image
	// request can be outstanding at the same time without their answers
	// colliding.
	pendingGet map[platform.AtomID]*pendingRequest
	nextGen    uint64
	nextProp   uint64
	freeProps  []platform.AtomID // properties of answered requests, ready for reuse
	// propPrefix names the properties of this manager.
	propPrefix string
	after      func(time.Duration, func())

	// INCR transfers (ICCCM 2.7.2): outgoing ones we serve as owner, keyed
	// by requestor window and property; incoming ones by their property.
	outgoing map[incrKey]*outgoingIncr
	incoming map[platform.AtomID]*incomingIncr

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

// managerSeq numbers the managers, so that the properties of two
// managers on one display never share a name (and so an answer meant for
// one of them is not taken for the other's).
var managerSeq atomic.Uint64

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
	requestor platform.WindowID
	typ       platform.AtomID
	buf       []byte
	callbacks []func(string)
	gen       uint64
}

// requestTimeout is how long a clipboard owner may stay silent before the
// request fails, like the 5 idle seconds of tkUnixSelect.c SelTimeoutProc.
const requestTimeout = 5 * time.Second

// pendingRequest is an outstanding clipboard request and everyone waiting
// for its answer. The property is the one the answer arrives in; it
// belongs to this request alone.
type pendingRequest struct {
	gen               uint64
	selection, target platform.AtomID
	property          platform.AtomID
	requestor         platform.WindowID
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
		pendingGet: make(map[platform.AtomID]*pendingRequest),
		clipboard:  server.InternAtom("CLIPBOARD", false),
		utf8str:    server.InternAtom("UTF8_STRING", false),
		targets:    server.InternAtom("TARGETS", false),
		incr:       server.InternAtom("INCR", false),
		outgoing:   make(map[incrKey]*outgoingIncr),
		incoming:   make(map[platform.AtomID]*incomingIncr),
		propPrefix: fmt.Sprintf("TAKIGO_SEL_%d_", managerSeq.Add(1)),
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
	// Copy what the reply needs out of the map: it may change while the
	// reply is being built.
	var (
		haveFormats bool
		data        []byte
		haveData    bool
		atoms       []platform.AtomID
	)
	if formats, has := m.formats[selection]; has {
		haveFormats = true
		data, haveData = formats[target]
		atoms = make([]platform.AtomID, 0, len(formats))
		for t := range formats {
			atoms = append(atoms, t)
		}
		slices.Sort(atoms)
	}
	m.mu.Unlock()

	if !ok {
		// We don't own this selection — refuse.
		m.server.SendSelectionNotify(requestor, selection, target, 0, time)
		return
	}

	if property == 0 {
		property = target
	}

	if haveFormats {
		// Offered with OwnFormats: exactly those targets.
		switch {
		case target == m.targets:
			m.server.ChangePropertyAtoms(requestor, property, append([]platform.AtomID{m.targets}, atoms...))
		case !haveData:
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

// Request asks the owner of selection for its text and returns the
// property its answer arrives in; the caller reads it with
// ReadProperty. Use RequestWithCallback or RequestTarget to be told the
// content.
func (m *Manager) Request(selection platform.AtomID, requestor platform.WindowID, time platform.Timestamp) platform.AtomID {
	return m.request(selection, m.utf8str, requestor, time, nil)
}

// request asks the owner of selection for its data in target and waits
// for the answer, which arrives as a SelectionNotify carrying the
// property of the request. The callbacks receive the content. A request
// for the same selection and target from the same window that is still
// outstanding shares its answer.
func (m *Manager) request(selection, target platform.AtomID, requestor platform.WindowID, ts platform.Timestamp, callbacks []func(string)) platform.AtomID {
	m.mu.Lock()
	if p := m.findPending(requestor, selection, target); p != nil {
		p.callbacks = append(p.callbacks, callbacks...)
		property := p.property
		m.mu.Unlock()
		return property
	}
	m.mu.Unlock()

	property := m.newProperty()
	m.mu.Lock()
	if p := m.findPending(requestor, selection, target); p != nil {
		// Another request for the same thing started meanwhile.
		p.callbacks = append(p.callbacks, callbacks...)
		property := p.property
		m.mu.Unlock()
		return property
	}
	m.nextGen++
	gen := m.nextGen
	m.pendingGet[property] = &pendingRequest{
		gen:       gen,
		selection: selection,
		target:    target,
		property:  property,
		requestor: requestor,
		callbacks: callbacks,
	}
	after := m.after
	m.mu.Unlock()

	m.convert(selection, target, property, requestor, ts)
	if after != nil {
		after(requestTimeout, func() { m.expire(property, gen) })
	}
	return property
}

// newProperty returns the property of a new request. Every outstanding
// request gets a property of its own, so that the answers to two of them
// can never be confused with one another.
//
// X keeps an atom until the server goes away, so a property whose answer
// was read (and deleted) goes back to a pool for the next request. One
// that timed out does not: a late answer to the old request must not be
// taken for the answer to a newer one, so it is never handed out again.
func (m *Manager) newProperty() platform.AtomID {
	m.mu.Lock()
	if n := len(m.freeProps); n > 0 {
		property := m.freeProps[n-1]
		m.freeProps = m.freeProps[:n-1]
		m.mu.Unlock()
		return property
	}
	m.nextProp++
	n := m.nextProp
	prefix := m.propPrefix
	m.mu.Unlock()
	if prefix == "" {
		prefix = "TAKIGO_SEL_"
	}
	return m.server.InternAtom(fmt.Sprintf("%s%d", prefix, n), false)
}

// releaseProperty returns the property of an answered request to the pool.
func (m *Manager) releaseProperty(property platform.AtomID) {
	m.mu.Lock()
	m.freeProps = append(m.freeProps, property)
	m.mu.Unlock()
}

// findPending returns the outstanding request of requestor for the given
// selection and target, if there is one: it shares its answer.
func (m *Manager) findPending(requestor platform.WindowID, selection, target platform.AtomID) *pendingRequest {
	for _, p := range m.pendingGet {
		if p.requestor == requestor && p.selection == selection && p.target == target {
			return p
		}
	}
	return nil
}

func (m *Manager) convert(selection, target, property platform.AtomID, requestor platform.WindowID, time platform.Timestamp) {
	m.server.ConvertSelection(selection, target, property, requestor, time)
}

// RequestTarget asks the owner of selection for its data in the given
// target and calls callback with the bytes, or with nil if the owner
// refuses or does not answer within requestTimeout. A request for
// another target of the same selection may be outstanding at the same
// time; each gets its own answer. The caller must route
// SelectionNotify to HandleSelectionNotify, as for
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
	m.mu.Unlock()

	m.request(selection, target, requestor, ts, []func(string){deliver})
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
	in := &incomingIncr{property: property, requestor: requestor, callbacks: p.callbacks, gen: m.nextGen}
	m.incoming[property] = in
	m.mu.Unlock()
	m.armIncomingTimeout(property, in.gen)
}

// armIncomingTimeout fails an INCR transfer that stalls for requestTimeout;
// each chunk re-arms it, as Tk resets retrPtr->idleTime.
func (m *Manager) armIncomingTimeout(property platform.AtomID, gen uint64) {
	m.mu.Lock()
	after := m.after
	m.mu.Unlock()
	if after == nil {
		return
	}
	after(requestTimeout, func() {
		m.mu.Lock()
		in := m.incoming[property]
		if in == nil || in.gen != gen {
			m.mu.Unlock()
			return
		}
		delete(m.incoming, property)
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
	in := m.incoming[property]
	if in == nil || in.requestor != w {
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
		m.armIncomingTimeout(property, gen)
		return true
	}
	delete(m.incoming, property)
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
// another from the same window for the same thing is outstanding shares
// its answer; a request for a different target (an image, say) runs
// beside it, each with its own property. If the owner does not answer
// within requestTimeout, callbacks receive "".
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
	m.mu.Unlock()

	m.request(m.clipboard, m.utf8str, requestor, ts, []func(string){callback})
}

// expire fails the request of the given property if it is still
// unanswered; a reply arriving afterwards names a property nobody
// waits for any more and is ignored.
func (m *Manager) expire(property platform.AtomID, gen uint64) {
	m.mu.Lock()
	p := m.pendingGet[property]
	if p == nil || p.gen != gen {
		m.mu.Unlock()
		return
	}
	delete(m.pendingGet, property)
	m.mu.Unlock()
	for _, cb := range p.callbacks {
		cb("")
	}
}

// HandleSelectionNotify is called when a SelectionNotify event arrives for
// a request this manager made. It reads the property, fires the pending
// callback, and returns true if a callback was pending. A reply whose
// property is not the one of a pending request belongs to a request that
// has already expired (or to another client) and is ignored.
func (m *Manager) HandleSelectionNotify(requestor platform.WindowID, selection, target, property platform.AtomID) bool {
	m.mu.Lock()
	p := m.pendingGet[property]
	if p == nil {
		if property != 0 {
			m.mu.Unlock()
			return false
		}
		// The owner refused the conversion: it sends the notify
		// without a property. It does not say which of our requests
		// it refuses, so fail those of the requestor window for the
		// same selection and target.
		var failed []func(string)
		for prop, req := range m.pendingGet {
			if req.requestor == requestor && req.selection == selection && req.target == target {
				failed = append(failed, req.callbacks...)
				delete(m.pendingGet, prop)
			}
		}
		m.mu.Unlock()
		for _, cb := range failed {
			cb("")
		}
		return len(failed) > 0
	}
	delete(m.pendingGet, property)
	m.mu.Unlock()

	var text string
	if property != 0 {
		data, typ, _ := m.server.GetWindowProperty(p.requestor, property, 0, maxPropWords, true)
		if typ != 0 && typ == m.incr {
			// The owner sends the text in chunks; reading (and so deleting)
			// the INCR property asked for the first one.
			m.startIncoming(p.requestor, property, p)
			return true
		}
		text = m.decode(data, typ)
		m.releaseProperty(property)
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
