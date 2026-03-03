//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/platform"
)

// InitPredefinedAtoms sets the platform-level predefined atom variables
// to their X11 values. Call this once after creating the display server.
func InitPredefinedAtoms() {
	platform.XA_WM_NAME = platform.AtomID(xlib.XA_WM_NAME)
	platform.XA_STRING = platform.AtomID(xlib.XA_STRING)
	platform.XA_WM_NORMAL_HINTS = platform.AtomID(xlib.XA_WM_NORMAL_HINTS)
	platform.XA_PRIMARY = platform.AtomID(xlib.XA_PRIMARY)
	platform.XA_SECONDARY = platform.AtomID(xlib.XA_SECONDARY)
	platform.XA_ATOM = platform.AtomID(xlib.XA_ATOM)
	platform.XA_CARDINAL = platform.AtomID(xlib.XA_CARDINAL)
	platform.XA_WINDOW = platform.AtomID(xlib.XA_WINDOW)
}
