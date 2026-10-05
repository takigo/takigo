//go:build windows

package windows

import (
	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// --- GrabManager implementation ---

func (d *WindowsDisplay) GrabPointer(grabWindow platform.WindowID, ownerEvents bool, eventMask uint,
	pointerMode, keyboardMode int, confineTo platform.WindowID, cursor platform.CursorID, time platform.Timestamp) int {
	w32.SetCapture(toHWND(grabWindow))
	return platform.GrabSuccess
}

func (d *WindowsDisplay) UngrabPointer(time platform.Timestamp) {
	w32.ReleaseCapture()
}

func (d *WindowsDisplay) GrabKeyboard(grabWindow platform.WindowID, ownerEvents bool,
	pointerMode, keyboardMode int, time platform.Timestamp) int {
	w32.SetFocus(toHWND(grabWindow))
	return platform.GrabSuccess
}

func (d *WindowsDisplay) UngrabKeyboard(time platform.Timestamp) {
	// No direct keyboard ungrab on Windows; focus is managed by SetFocus.
}
