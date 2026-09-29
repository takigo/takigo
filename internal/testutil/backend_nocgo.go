//go:build !cgo && !windows

package testutil

// haveBackend reports whether this build has a display backend: without
// cgo there is neither X11 nor Cocoa.
const haveBackend = false
