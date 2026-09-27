//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"unsafe"

	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/platform"
)

// toXGC converts a platform.GCID to xlib.GC. GCIDs hold C pointers,
// which the Go GC does not track, so reinterpreting the bits is safe
// (and, unlike a uintptr-to-unsafe.Pointer conversion, vet-clean).
func toXGC(gc platform.GCID) xlib.GC {
	return *(*xlib.GC)(unsafe.Pointer(&gc))
}

// fromXGC converts an xlib.GC to platform.GCID.
func fromXGC(gc xlib.GC) platform.GCID {
	return platform.GCID(uintptr(unsafe.Pointer(gc)))
}
