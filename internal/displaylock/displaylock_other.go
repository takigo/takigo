//go:build !unix

// Package displaylock serializes display tests across test binaries; see
// displaylock_unix.go. Windows gives each window its own surface, so there
// is nothing to serialize.
package displaylock

import (
	"os"
	"testing"
)

// Acquire does nothing on this platform.
func Acquire(testing.TB) {}

// UseVirtualDisplay does nothing on this platform.
func UseVirtualDisplay() {}

// Require skips t when no display is available; there is no lock to hold.
func Require(t testing.TB) {
	t.Helper()
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display available (set DISPLAY or use xvfb-run)")
	}
}
