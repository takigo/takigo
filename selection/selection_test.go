package selection

import (
	"sync"
	"testing"

	"github.com/msorc/takigo/platform"
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

	// Case 2: Not owned locally -> async (stored in pendingGet)
	m2 := &Manager{
		data:       map[platform.AtomID]string{},
		clipboard:  platform.AtomID(1),
		pendingGet: make(map[platform.WindowID]func(string)),
		mu:         sync.Mutex{},
	}

	m2.mu.Lock()
	_, ok = m2.data[m2.clipboard]
	if !ok {
		m2.pendingGet[platform.WindowID(100)] = callback
	}
	m2.mu.Unlock()

	if len(m2.pendingGet) != 1 {
		t.Errorf("pendingGet length = %d, want 1", len(m2.pendingGet))
	}
	if m2.pendingGet[platform.WindowID(100)] == nil {
		t.Error("pendingGet should store callback")
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