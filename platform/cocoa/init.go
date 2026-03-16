//go:build darwin

package cocoa

import (
	clib "github.com/msorc/takigo/internal/cocoa"
	"github.com/msorc/takigo/platform"
)

// InitPredefinedAtoms sets the platform-level predefined atom variables
// to their macOS-emulated values. Call this once after creating the display server.
func InitPredefinedAtoms() {
	platform.XA_WM_NAME = platform.AtomID(clib.InternAtom("WM_NAME", false))
	platform.XA_STRING = platform.AtomID(clib.InternAtom("STRING", false))
	platform.XA_WM_NORMAL_HINTS = platform.AtomID(clib.InternAtom("WM_NORMAL_HINTS", false))
	platform.XA_PRIMARY = platform.AtomID(clib.InternAtom("PRIMARY", false))
	platform.XA_SECONDARY = platform.AtomID(clib.InternAtom("SECONDARY", false))
	platform.XA_ATOM = platform.AtomID(clib.InternAtom("ATOM", false))
	platform.XA_CARDINAL = platform.AtomID(clib.InternAtom("CARDINAL", false))
	platform.XA_WINDOW = platform.AtomID(clib.InternAtom("WINDOW", false))
}
