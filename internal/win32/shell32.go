//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

var (
	shell32               = syscall.NewLazyDLL("shell32.dll")
	procShell_NotifyIconW = shell32.NewProc("Shell_NotifyIconW")
)

// Shell_NotifyIconW messages and NOTIFYICONDATAW flags.
const (
	NIM_ADD    = 0x0
	NIM_MODIFY = 0x1
	NIM_DELETE = 0x2

	NIF_MESSAGE = 0x01
	NIF_ICON    = 0x02
	NIF_TIP     = 0x04
	NIF_INFO    = 0x10

	NIIF_USER = 0x4

	// WM_APP_TRAY is the callback message of the tray icons.
	WM_APP_TRAY = 0x8002
)

// GUID is a Windows GUID.
type GUID struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

// NOTIFYICONDATAW is the Shell_NotifyIconW record.
type NOTIFYICONDATAW struct {
	CbSize            uint32
	HWnd              HWND
	UID               uint32
	UFlags            uint32
	UCallbackMessage  uint32
	HIcon             HICON
	SzTip             [128]uint16
	DwState           uint32
	DwStateMask       uint32
	SzInfo            [256]uint16
	UTimeoutOrVersion uint32
	SzInfoTitle       [64]uint16
	DwInfoFlags       uint32
	GuidItem          GUID
	HBalloonIcon      HICON
}

// Shell_NotifyIcon adds, modifies or deletes a notification-area icon.
func Shell_NotifyIcon(message uint32, data *NOTIFYICONDATAW) bool {
	data.CbSize = uint32(unsafe.Sizeof(*data))
	r, _, _ := procShell_NotifyIconW.Call(uintptr(message), uintptr(unsafe.Pointer(data)))
	return r != 0
}

// CopyUTF16 copies s into dst as a NUL-terminated UTF-16 string,
// truncated to fit.
func CopyUTF16(dst []uint16, s string) {
	u, _ := syscall.UTF16FromString(s)
	n := copy(dst[:len(dst)-1], u)
	dst[n] = 0
}
