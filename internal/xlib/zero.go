//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <X11/Xlib.h>
*/
import "C"

// IsZeroGC returns true if the GC is the zero value.
func IsZeroGC(gc GC) bool {
	return C.GC(gc) == nil
}

// ZeroGC returns the zero value for GC.
func ZeroGC() GC {
	return GC(nil)
}
