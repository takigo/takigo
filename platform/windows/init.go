//go:build windows

package windows

import "github.com/msorc/takigo/platform"

// Predefined atom IDs for Windows. Since Windows doesn't have X11 atoms,
// we use our own interned atom system. These are initialized with well-known
// names matching the X11 predefined atoms.
func InitPredefinedAtoms() {
	// Use the global display's atom system if available, or assign fixed IDs.
	// These match the X11 predefined atom values for compatibility.
	platform.XA_WM_NAME = platform.AtomID(1)
	platform.XA_STRING = platform.AtomID(2)
	platform.XA_WM_NORMAL_HINTS = platform.AtomID(3)
	platform.XA_PRIMARY = platform.AtomID(4)
	platform.XA_SECONDARY = platform.AtomID(5)
	platform.XA_ATOM = platform.AtomID(6)
	platform.XA_CARDINAL = platform.AtomID(7)
	platform.XA_WINDOW = platform.AtomID(8)
}
