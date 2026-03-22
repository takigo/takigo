//go:build windows

package win32

import (
	"sync"
	"syscall"
)

// WndProcCallback is the function signature for window procedure handlers.
type WndProcCallback func(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT

// globalWndProc is the registered global window procedure.
var (
	globalWndProcMu sync.RWMutex
	globalWndProc   WndProcCallback
	wndProcPtr      uintptr
)

// SetGlobalWndProc sets the global window procedure handler.
// Must be called before creating any windows.
func SetGlobalWndProc(fn WndProcCallback) {
	globalWndProcMu.Lock()
	globalWndProc = fn
	globalWndProcMu.Unlock()
}

// GetWndProcPtr returns the syscall.NewCallback pointer for the window procedure.
// It lazily initializes the callback on first call.
func GetWndProcPtr() uintptr {
	if wndProcPtr == 0 {
		wndProcPtr = syscall.NewCallback(wndProcBridge)
	}
	return wndProcPtr
}

// wndProcBridge is the actual Windows callback that bridges to the Go handler.
func wndProcBridge(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	globalWndProcMu.RLock()
	fn := globalWndProc
	globalWndProcMu.RUnlock()

	if fn != nil {
		return fn(hwnd, msg, wParam, lParam)
	}
	return DefWindowProc(hwnd, msg, wParam, lParam)
}
