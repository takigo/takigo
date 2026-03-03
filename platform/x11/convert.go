//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"unsafe"

	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/platform"
)

// toXGC converts a platform.GCID to xlib.GC.
// xlib.GC is a C pointer type, so we must go through unsafe.Pointer.
func toXGC(gc platform.GCID) xlib.GC {
	return xlib.GC(unsafe.Pointer(gc))
}

// fromXGC converts an xlib.GC to platform.GCID.
func fromXGC(gc xlib.GC) platform.GCID {
	return platform.GCID(uintptr(unsafe.Pointer(gc)))
}
