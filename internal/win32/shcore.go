//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

var (
	shcore = syscall.NewLazyDLL("shcore.dll")

	procSetProcessDpiAwareness = shcore.NewProc("SetProcessDpiAwareness")
	procGetDpiForMonitor       = shcore.NewProc("GetDpiForMonitor")

	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
)

// DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 is ((DPI_AWARENESS_CONTEXT)-4).
const DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = ^uintptr(3)

// SetProcessDpiAwarenessContext sets the process DPI awareness context
// (Windows 10 1703 and later); it returns an error where it is missing.
func SetProcessDpiAwarenessContext(ctx uintptr) error {
	if err := procSetProcessDpiAwarenessContext.Find(); err != nil {
		return err
	}
	if r, _, err := procSetProcessDpiAwarenessContext.Call(ctx); r == 0 {
		return err
	}
	return nil
}

// DPI awareness levels.
const (
	PROCESS_DPI_UNAWARE           = 0
	PROCESS_SYSTEM_DPI_AWARE      = 1
	PROCESS_PER_MONITOR_DPI_AWARE = 2
)

// Monitor DPI types.
const (
	MDT_EFFECTIVE_DPI = 0
	MDT_ANGULAR_DPI   = 1
	MDT_RAW_DPI       = 2
)

// SetProcessDpiAwareness sets the DPI awareness for the process.
// Returns an error HRESULT on failure.
func SetProcessDpiAwareness(awareness int32) error {
	if err := procSetProcessDpiAwareness.Find(); err != nil {
		return err
	}
	r, _, _ := procSetProcessDpiAwareness.Call(uintptr(awareness))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}

// GetDpiForMonitor retrieves the DPI for a monitor.
func GetDpiForMonitor(hMonitor uintptr, dpiType int32, dpiX, dpiY *uint32) error {
	r, _, _ := procGetDpiForMonitor.Call(hMonitor, uintptr(dpiType),
		uintptr(unsafe.Pointer(dpiX)), uintptr(unsafe.Pointer(dpiY)))
	if r != 0 {
		return syscall.Errno(r)
	}
	return nil
}
