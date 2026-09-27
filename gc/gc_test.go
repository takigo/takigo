package gc

import (
	"testing"

	"github.com/msorc/takigo/platform"
)

func TestValuesEquality(t *testing.T) {
	v1 := Values{Foreground: 0xFF0000, Background: 0x00FF00, LineWidth: 2, Function: 3}
	v2 := Values{Foreground: 0xFF0000, Background: 0x00FF00, LineWidth: 2, Function: 3}
	v3 := Values{Foreground: 0x00FF00, Background: 0xFF0000, LineWidth: 2, Function: 3}

	if v1 != v2 {
		t.Errorf("Equal Values should be equal")
	}
	if v1 == v3 {
		t.Errorf("Different Values should not be equal")
	}
}

func TestKeyEquality(t *testing.T) {
	k1 := key{Values: Values{Foreground: 1, Background: 2, LineWidth: 3, Function: 4}, depth: 24}
	k2 := key{Values: Values{Foreground: 1, Background: 2, LineWidth: 3, Function: 4}, depth: 24}
	k3 := key{Values: Values{Foreground: 1, Background: 2, LineWidth: 3, Function: 4}, depth: 32}

	if k1 != k2 {
		t.Errorf("Equal keys should be equal")
	}
	if k1 == k3 {
		t.Errorf("Keys with different depth should not be equal")
	}
}

func TestNewPool(t *testing.T) {
	// Test that NewPool creates a valid pool structure
	// We can't fully test without a display, but we can verify the constructor
	var server platform.DisplayServer = nil
	p := NewPool(server, 0, 24)

	if p == nil {
		t.Fatal("NewPool returned nil")
	}
	if p.server != server {
		t.Error("Pool.server not set")
	}
	if p.screen != 0 {
		t.Errorf("Pool.screen = %d, want 0", p.screen)
	}
	if p.depth != 24 {
		t.Errorf("Pool.depth = %d, want 24", p.depth)
	}
	if p.byValue == nil {
		t.Error("Pool.byValue is nil")
	}
	if p.byGC == nil {
		t.Error("Pool.byGC is nil")
	}
}

func TestPoolGetFreeNilServer(t *testing.T) {
	p := NewPool(nil, 0, 24)

	// Get with nil server should panic or error when creating GC
	defer func() {
		if r := recover(); r == nil {
			t.Error("Get with nil server should panic or error")
		}
	}()

	_ = p.Get(0, 0, &Values{})
}

func TestPoolFreeNotFound(t *testing.T) {
	p := NewPool(nil, 0, 24)

	// Free with unknown GC should not panic
	p.Free(platform.GCID(123))
}

func TestPoolCleanup(t *testing.T) {
	p := NewPool(nil, 0, 24)

	// Cleanup should not panic on empty pool
	p.Cleanup()
}
