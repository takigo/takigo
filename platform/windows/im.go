//go:build windows

package windows

import "github.com/msorc/takigo/platform"

// --- InputMethodManager implementation ---
// Windows handles IME natively through the WM_CHAR/WM_UNICHAR/WM_IME_* messages.
// The current implementation provides stubs; full IMM32 support can be added later.

func (d *WindowsDisplay) InitIM(root platform.WindowID) {
	// Windows IME is initialized automatically by the OS.
}

func (d *WindowsDisplay) HasIM() bool {
	// Return false since we handle IME through WM_CHAR in the WndProc,
	// not through a separate IM interface like XIM.
	return false
}

func (d *WindowsDisplay) SetICFocus(w platform.WindowID) {
	// No-op: Windows IME focus follows the keyboard focus automatically.
}

func (d *WindowsDisplay) UnsetICFocus() {
	// No-op.
}
