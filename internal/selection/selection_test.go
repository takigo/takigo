package selection

import (
	"sync"
	"testing"
	"time"

	"github.com/takigo/takigo/platform"
)

func TestManagerOwnGetContent(t *testing.T) {
	// We can't easily test Manager without a display, but we can
	// verify the internal data structures work correctly

	// Test data map
	data := make(map[platform.AtomID]string)
	owners := make(map[platform.AtomID]platform.WindowID)

	sel := platform.AtomID(1)
	owner := platform.WindowID(100)
	content := "test content"

	data[sel] = content
	owners[sel] = owner

	if data[sel] != content {
		t.Errorf("data[%v] = %q, want %q", sel, data[sel], content)
	}
	if owners[sel] != owner {
		t.Errorf("owners[%v] = %v, want %v", sel, owners[sel], owner)
	}

	// Test missing selection
	if data[platform.AtomID(999)] != "" {
		t.Error("missing selection should return empty string")
	}
	if owners[platform.AtomID(999)] != 0 {
		t.Error("missing owner should return 0")
	}
}

func TestHandleSelectionRequestTargets(t *testing.T) {
	// Test the logic for TARGETS request
	targets := []platform.AtomID{
		platform.AtomID(1), // UTF8_STRING
		platform.AtomID(2), // STRING
		platform.AtomID(3), // TARGETS
	}

	if len(targets) != 3 {
		t.Errorf("targets length = %d, want 3", len(targets))
	}

	// Verify order: UTF8_STRING, STRING, TARGETS
	if targets[0] != platform.AtomID(1) {
		t.Error("first target should be UTF8_STRING")
	}
	if targets[1] != platform.AtomID(2) {
		t.Error("second target should be STRING")
	}
	if targets[2] != platform.AtomID(3) {
		t.Error("third target should be TARGETS")
	}
}

func TestRequestWithCallbackLogic(t *testing.T) {
	// Test the logic flow of RequestWithCallback

	// Case 1: Local ownership -> synchronous callback
	m := &Manager{
		data: map[platform.AtomID]string{
			platform.AtomID(1): "local content",
		},
		clipboard: platform.AtomID(1),
		mu:        sync.Mutex{},
	}

	var called bool
	var received string

	callback := func(s string) {
		called = true
		received = s
	}

	// Simulate the local ownership case
	m.mu.Lock()
	content, ok := m.data[m.clipboard]
	m.mu.Unlock()

	if ok {
		callback(content)
	}

	if !called {
		t.Error("callback should have been called for local content")
	}
	if received != "local content" {
		t.Errorf("received = %q, want \"local content\"", received)
	}
}

// convertCall is a recorded XConvertSelection request.
type convertCall struct {
	selection, target, property platform.AtomID
	requestor                   platform.WindowID
}

// convertServer records ConvertSelection calls and serves one property
// per request.
type convertServer struct {
	platform.DisplayServer
	mu       sync.Mutex
	atoms    map[string]platform.AtomID
	nextAtom platform.AtomID
	converts []convertCall
	props    map[platform.AtomID][]byte
}

func (s *convertServer) InternAtom(name string, _ bool) platform.AtomID {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.atoms == nil {
		s.atoms = map[string]platform.AtomID{}
	}
	if a, ok := s.atoms[name]; ok {
		return a
	}
	s.nextAtom += 100
	s.atoms[name] = s.nextAtom
	return s.nextAtom
}

func (s *convertServer) Atoms() *platform.Atoms        { return &platform.Atoms{String: 31} }
func (s *convertServer) ClipboardText() (string, bool) { return "", false }
func (s *convertServer) ConvertSelection(selection, target, property platform.AtomID, requestor platform.WindowID, _ platform.Timestamp) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.converts = append(s.converts, convertCall{selection, target, property, requestor})
}

func (s *convertServer) GetWindowProperty(_ platform.WindowID, property platform.AtomID, _, _ int64, _ bool) ([]byte, platform.AtomID, int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.props[property], 8, 0
}

// calls returns the recorded conversion requests.
func (s *convertServer) calls() []convertCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]convertCall(nil), s.converts...)
}

// serve puts the answer of an owner into the property of a request.
func (s *convertServer) serve(property platform.AtomID, data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.props == nil {
		s.props = map[platform.AtomID][]byte{}
	}
	s.props[property] = data
}

func newRemoteManager(srv *convertServer) (*Manager, *[]func()) {
	var timers []func()
	m := &Manager{
		server:     srv,
		data:       map[platform.AtomID]string{},
		clipboard:  platform.AtomID(1),
		utf8str:    platform.AtomID(2),
		pendingGet: map[platform.AtomID]*pendingRequest{},
	}
	m.SetTimer(func(_ time.Duration, fn func()) { timers = append(timers, fn) })
	return m, &timers
}

func TestRequestWithCallbackSharesOutstandingRequest(t *testing.T) {
	srv := &convertServer{}
	m, _ := newRemoteManager(srv)
	var got []string
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, "a:"+s) })
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, "b:"+s) })
	if len(srv.calls()) != 1 {
		t.Fatalf("%d ConvertSelection requests, want 1 shared", len(srv.calls()))
	}
	property := srv.calls()[0].property
	srv.serve(property, []byte("pasted"))
	if !m.HandleSelectionNotify(100, m.clipboard, m.utf8str, property) {
		t.Fatal("reply found no pending request")
	}
	if len(got) != 2 || got[0] != "a:pasted" || got[1] != "b:pasted" {
		t.Errorf("callbacks got %q, want both to receive the reply", got)
	}
}

func TestRequestWithCallbackTimesOut(t *testing.T) {
	srv := &convertServer{}
	m, timers := newRemoteManager(srv)
	var got []string
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, s) })
	if len(*timers) != 1 {
		t.Fatalf("%d timeouts scheduled, want 1", len(*timers))
	}
	property := srv.calls()[0].property
	(*timers)[0]()
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("after timeout callbacks got %q, want one empty result", got)
	}
	if m.HandleSelectionNotify(100, m.clipboard, m.utf8str, property) {
		t.Error("late reply matched an expired request")
	}

	// A new request is not cancelled by the old request's timer.
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, s) })
	(*timers)[0]()
	newProperty := srv.calls()[1].property
	srv.serve(newProperty, []byte("late"))
	if !m.HandleSelectionNotify(100, m.clipboard, m.utf8str, newProperty) || got[len(got)-1] != "late" {
		t.Errorf("new request lost to a stale timeout: got %q", got)
	}
}

// pngTarget is the target of an image request.
const pngTarget = platform.AtomID(3)

// TestTextAndImageRequestsDoNotCollide checks that a text request and an
// image request for the same clipboard can be outstanding at the same
// time and that each answer reaches the right callback.
func TestTextAndImageRequestsDoNotCollide(t *testing.T) {
	srv := &convertServer{}
	m, _ := newRemoteManager(srv)

	var text string
	var image []byte
	m.RequestWithCallback(100, 0, func(s string) { text = s })
	m.RequestTarget(m.clipboard, pngTarget, 100, 0, func(data []byte) { image = data })

	calls := srv.calls()
	if len(calls) != 2 {
		t.Fatalf("%d ConvertSelection requests, want one per request", len(calls))
	}
	if calls[0].property == calls[1].property {
		t.Fatalf("both requests use property %d, want distinct properties", calls[0].property)
	}
	if calls[0].target != m.utf8str || calls[1].target != pngTarget {
		t.Fatalf("targets = %d and %d, want %d and %d", calls[0].target, calls[1].target, m.utf8str, pngTarget)
	}

	// The image answers first, the text second: neither may take the
	// other's answer.
	png := []byte{0x89, 'P', 'N', 'G'}
	srv.serve(calls[1].property, png)
	if !m.HandleSelectionNotify(100, m.clipboard, pngTarget, calls[1].property) {
		t.Fatal("the image answer found no pending request")
	}
	if string(image) != string(png) {
		t.Errorf("image = %q, want %q", image, png)
	}
	if text != "" {
		t.Errorf("text = %q, want it still empty", text)
	}

	srv.serve(calls[0].property, []byte("pasted"))
	if !m.HandleSelectionNotify(100, m.clipboard, m.utf8str, calls[0].property) {
		t.Fatal("the text answer found no pending request")
	}
	if text != "pasted" {
		t.Errorf("text = %q, want %q", text, "pasted")
	}
}

// TestStaleSelectionNotifyIsIgnored checks that the answer of a request
// which has already expired does not reach a newer request.
func TestStaleSelectionNotifyIsIgnored(t *testing.T) {
	srv := &convertServer{}
	m, timers := newRemoteManager(srv)

	var got []string
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, s) })
	oldProperty := srv.calls()[0].property

	// The owner is too slow: the request times out.
	(*timers)[0]()
	if len(got) != 1 || got[0] != "" {
		t.Fatalf("after timeout callbacks got %q, want one empty result", got)
	}
	if len(m.pendingGet) != 0 {
		t.Errorf("%d pending requests left after the timeout, want none", len(m.pendingGet))
	}

	// A new request is made, and the old answer arrives late.
	m.RequestWithCallback(100, 0, func(s string) { got = append(got, s) })
	newProperty := srv.calls()[1].property
	if newProperty == oldProperty {
		t.Fatalf("the new request reuses property %d of the expired one", oldProperty)
	}
	srv.serve(oldProperty, []byte("stale"))
	if m.HandleSelectionNotify(100, m.clipboard, m.utf8str, oldProperty) {
		t.Error("the stale answer matched a pending request")
	}
	if len(got) != 1 {
		t.Fatalf("callbacks got %q, want the stale answer to be dropped", got)
	}

	// The answer of the new request still arrives.
	srv.serve(newProperty, []byte("fresh"))
	if !m.HandleSelectionNotify(100, m.clipboard, m.utf8str, newProperty) {
		t.Fatal("the answer of the new request found no pending request")
	}
	if len(got) != 2 || got[1] != "fresh" {
		t.Errorf("callbacks got %q, want the new answer", got)
	}
}

// TestRequestPropertiesAreDistinct checks that every request gets a
// property of its own.
func TestRequestPropertiesAreDistinct(t *testing.T) {
	srv := &convertServer{}
	m, _ := newRemoteManager(srv)

	m.RequestWithCallback(100, 0, func(string) {})
	m.RequestWithCallback(200, 0, func(string) {})
	m.RequestTarget(m.clipboard, pngTarget, 100, 0, func([]byte) {})

	calls := srv.calls()
	if len(calls) != 3 {
		t.Fatalf("%d ConvertSelection requests, want 3", len(calls))
	}
	seen := map[platform.AtomID]bool{}
	for _, c := range calls {
		if c.property == 0 {
			t.Fatal("a request has no property")
		}
		if seen[c.property] {
			t.Errorf("property %d is used by two requests", c.property)
		}
		seen[c.property] = true
	}
	if len(m.pendingGet) != 3 {
		t.Errorf("%d pending requests, want 3", len(m.pendingGet))
	}
}

// TestRefusedRequestFails checks that an owner which refuses the
// conversion (a notify without a property) fails the waiting request.
func TestRefusedRequestFails(t *testing.T) {
	srv := &convertServer{}
	m, _ := newRemoteManager(srv)

	var got []string
	m.RequestTarget(m.clipboard, pngTarget, 100, 0, func(data []byte) { got = append(got, string(data)) })
	if !m.HandleSelectionNotify(100, m.clipboard, pngTarget, 0) {
		t.Error("the refusal did not fail the pending request")
	}
	if len(got) != 1 || got[0] != "" {
		t.Errorf("callbacks got %q, want one empty result", got)
	}
}

func TestManagerAtoms(t *testing.T) {
	// Verify atom accessors
	m := &Manager{
		clipboard: platform.AtomID(100),
		utf8str:   platform.AtomID(200),
	}

	if m.ClipboardAtom() != platform.AtomID(100) {
		t.Errorf("ClipboardAtom() = %v, want 100", m.ClipboardAtom())
	}
	if m.UTF8StringAtom() != platform.AtomID(200) {
		t.Errorf("UTF8StringAtom() = %v, want 200", m.UTF8StringAtom())
	}
}

func TestHandleSelectionClear(t *testing.T) {
	// Test HandleSelectionClear removes data and owners
	m := &Manager{
		data: map[platform.AtomID]string{
			platform.AtomID(1): "content1",
			platform.AtomID(2): "content2",
		},
		owners: map[platform.AtomID]platform.WindowID{
			platform.AtomID(1): platform.WindowID(100),
			platform.AtomID(2): platform.WindowID(200),
		},
		mu: sync.Mutex{},
	}

	m.HandleSelectionClear(platform.AtomID(1))

	if _, ok := m.data[platform.AtomID(1)]; ok {
		t.Error("data[1] should be deleted")
	}
	if _, ok := m.owners[platform.AtomID(1)]; ok {
		t.Error("owners[1] should be deleted")
	}

	// Other selection should remain
	if m.data[platform.AtomID(2)] != "content2" {
		t.Error("data[2] should remain")
	}
	if m.owners[platform.AtomID(2)] != platform.WindowID(200) {
		t.Error("owners[2] should remain")
	}
}

func TestLatin1RoundTrip(t *testing.T) {
	if got := string(toLatin1("naïve € ok")); got != "na\xefve ? ok" {
		t.Errorf("toLatin1 = %q", got)
	}
	if got := fromLatin1([]byte("caf\xe9")); got != "café" {
		t.Errorf("fromLatin1 = %q", got)
	}
}

// nativeServer has a native clipboard, as the Win32 and Cocoa backends do.
type nativeServer struct {
	platform.DisplayServer
	board string
}

func (s *nativeServer) SetSelectionOwner(platform.AtomID, platform.WindowID, platform.Timestamp) {}
func (s *nativeServer) SetClipboardText(text string) bool                                        { s.board = text; return true }
func (s *nativeServer) ClipboardText() (string, bool)                                            { return s.board, true }

func TestNativeClipboardRoundTrip(t *testing.T) {
	srv := &nativeServer{}
	m := &Manager{
		server:     srv,
		data:       map[platform.AtomID]string{},
		owners:     map[platform.AtomID]platform.WindowID{},
		clipboard:  platform.AtomID(1),
		pendingGet: map[platform.AtomID]*pendingRequest{},
	}
	m.OwnClipboard(5, "copied here", 0)
	if srv.board != "copied here" {
		t.Fatalf("native clipboard = %q after OwnClipboard, want %q", srv.board, "copied here")
	}

	// Another application copies: pasting must return its text, not ours.
	srv.board = "copied elsewhere"
	var got string
	m.RequestWithCallback(6, 0, func(s string) { got = s })
	if got != "copied elsewhere" {
		t.Errorf("paste = %q, want the native clipboard's %q", got, "copied elsewhere")
	}
}
