//go:build !unix

// Package displaylock serializes display tests across test binaries; see
// displaylock_unix.go. Windows gives each window its own surface, so there
// is nothing to serialize.
package displaylock

import "testing"

// Acquire does nothing on this platform.
func Acquire(testing.TB) {}
