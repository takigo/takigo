//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

var (
	user32 = syscall.NewLazyDLL("user32.dll")

	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDestroyWindow      = user32.NewProc("DestroyWindow")
	procShowWindow         = user32.NewProc("ShowWindow")
	procUpdateWindow       = user32.NewProc("UpdateWindow")
	procMoveWindow         = user32.NewProc("MoveWindow")
	procSetWindowPos       = user32.NewProc("SetWindowPos")
	procGetWindowRect      = user32.NewProc("GetWindowRect")
	procGetClientRect      = user32.NewProc("GetClientRect")
	procSetWindowTextW     = user32.NewProc("SetWindowTextW")
	procSetWindowLongPtrW  = user32.NewProc("SetWindowLongPtrW")
	procGetWindowLongPtrW  = user32.NewProc("GetWindowLongPtrW")
	procSetClassLongPtrW   = user32.NewProc("SetClassLongPtrW")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procGetDC              = user32.NewProc("GetDC")
	procReleaseDC          = user32.NewProc("ReleaseDC")
	procBeginPaint         = user32.NewProc("BeginPaint")
	procEndPaint           = user32.NewProc("EndPaint")
	procInvalidateRect     = user32.NewProc("InvalidateRect")
	procPeekMessageW       = user32.NewProc("PeekMessageW")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procPostMessageW       = user32.NewProc("PostMessageW")
	procSendMessageW       = user32.NewProc("SendMessageW")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procSetCapture         = user32.NewProc("SetCapture")
	procReleaseCapture     = user32.NewProc("ReleaseCapture")
	procGetCapture         = user32.NewProc("GetCapture")
	procSetFocusW          = user32.NewProc("SetFocus")
	procGetFocusW          = user32.NewProc("GetFocus")
	procSetActiveWindow    = user32.NewProc("SetActiveWindow")
	procLoadCursorW        = user32.NewProc("LoadCursorW")
	procSetCursor          = user32.NewProc("SetCursor")
	procOpenClipboard      = user32.NewProc("OpenClipboard")
	procCloseClipboard     = user32.NewProc("CloseClipboard")
	procEmptyClipboard     = user32.NewProc("EmptyClipboard")
	procGetClipboardData   = user32.NewProc("GetClipboardData")
	procSetClipboardData   = user32.NewProc("SetClipboardData")
	procGetSystemMetrics   = user32.NewProc("GetSystemMetrics")
	procGetSysColor        = user32.NewProc("GetSysColor")
	procMapWindowPoints    = user32.NewProc("MapWindowPoints")
	procScreenToClient     = user32.NewProc("ScreenToClient")
	procClientToScreen     = user32.NewProc("ClientToScreen")
	procSetTimer           = user32.NewProc("SetTimer")
	procKillTimer          = user32.NewProc("KillTimer")
	procGetKeyState        = user32.NewProc("GetKeyState")
	procGetCursorPos       = user32.NewProc("GetCursorPos")
	procFillRect           = user32.NewProc("FillRect")
	procIsWindowVisible    = user32.NewProc("IsWindowVisible")
	procIsIconic           = user32.NewProc("IsIconic")
	procTrackMouseEvent    = user32.NewProc("TrackMouseEvent")
	procRedrawWindow       = user32.NewProc("RedrawWindow")
	procAdjustWindowRectEx = user32.NewProc("AdjustWindowRectEx")
)

func RegisterClassEx(wc *WNDCLASSEXW) ATOM {
	r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(wc)))
	return ATOM(r)
}

func CreateWindowEx(exStyle uint32, className, windowName *uint16,
	style uint32, x, y, width, height int32,
	parent HWND, menu HMENU, instance HINSTANCE, param unsafe.Pointer) HWND {
	r, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x), uintptr(y), uintptr(width), uintptr(height),
		uintptr(parent), uintptr(menu), uintptr(instance),
		uintptr(param))
	return HWND(r)
}

func DestroyWindow(hwnd HWND) bool {
	r, _, _ := procDestroyWindow.Call(uintptr(hwnd))
	return r != 0
}

func ShowWindow(hwnd HWND, cmdShow int32) bool {
	r, _, _ := procShowWindow.Call(uintptr(hwnd), uintptr(cmdShow))
	return r != 0
}

func UpdateWindow(hwnd HWND) bool {
	r, _, _ := procUpdateWindow.Call(uintptr(hwnd))
	return r != 0
}

func MoveWindow(hwnd HWND, x, y, width, height int32, repaint bool) bool {
	var bRepaint uintptr
	if repaint {
		bRepaint = 1
	}
	r, _, _ := procMoveWindow.Call(uintptr(hwnd),
		uintptr(x), uintptr(y), uintptr(width), uintptr(height), bRepaint)
	return r != 0
}

func SetWindowPos(hwnd, hWndInsertAfter HWND, x, y, cx, cy int32, flags uint32) bool {
	r, _, _ := procSetWindowPos.Call(uintptr(hwnd), uintptr(hWndInsertAfter),
		uintptr(x), uintptr(y), uintptr(cx), uintptr(cy), uintptr(flags))
	return r != 0
}

func GetWindowRect(hwnd HWND, rect *RECT) bool {
	r, _, _ := procGetWindowRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return r != 0
}

func GetClientRect(hwnd HWND, rect *RECT) bool {
	r, _, _ := procGetClientRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)))
	return r != 0
}

func SetWindowText(hwnd HWND, text *uint16) bool {
	r, _, _ := procSetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(text)))
	return r != 0
}

func SetWindowLongPtr(hwnd HWND, index int32, newLong uintptr) uintptr {
	r, _, _ := procSetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index), newLong)
	return r
}

func GetWindowLongPtr(hwnd HWND, index int32) uintptr {
	r, _, _ := procGetWindowLongPtrW.Call(uintptr(hwnd), uintptr(index))
	return r
}

func SetClassLongPtr(hwnd HWND, index int32, newLong uintptr) uintptr {
	r, _, _ := procSetClassLongPtrW.Call(uintptr(hwnd), uintptr(index), newLong)
	return r
}

func DefWindowProc(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	r, _, _ := procDefWindowProcW.Call(uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
	return LRESULT(r)
}

func GetDC(hwnd HWND) HDC {
	r, _, _ := procGetDC.Call(uintptr(hwnd))
	return HDC(r)
}

func ReleaseDC(hwnd HWND, hdc HDC) int32 {
	r, _, _ := procReleaseDC.Call(uintptr(hwnd), uintptr(hdc))
	return int32(r)
}

func BeginPaint(hwnd HWND, ps *PAINTSTRUCT) HDC {
	r, _, _ := procBeginPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return HDC(r)
}

func EndPaint(hwnd HWND, ps *PAINTSTRUCT) bool {
	r, _, _ := procEndPaint.Call(uintptr(hwnd), uintptr(unsafe.Pointer(ps)))
	return r != 0
}

func InvalidateRect(hwnd HWND, rect *RECT, erase bool) bool {
	var bErase uintptr
	if erase {
		bErase = 1
	}
	r, _, _ := procInvalidateRect.Call(uintptr(hwnd), uintptr(unsafe.Pointer(rect)), bErase)
	return r != 0
}

func PeekMessage(msg *MSG, hwnd HWND, msgFilterMin, msgFilterMax, removeMsg uint32) bool {
	r, _, _ := procPeekMessageW.Call(
		uintptr(unsafe.Pointer(msg)),
		uintptr(hwnd),
		uintptr(msgFilterMin), uintptr(msgFilterMax), uintptr(removeMsg))
	return r != 0
}

func TranslateMessage(msg *MSG) bool {
	r, _, _ := procTranslateMessage.Call(uintptr(unsafe.Pointer(msg)))
	return r != 0
}

func DispatchMessage(msg *MSG) LRESULT {
	r, _, _ := procDispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
	return LRESULT(r)
}

func PostMessage(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) bool {
	r, _, _ := procPostMessageW.Call(uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
	return r != 0
}

func SendMessage(hwnd HWND, msg uint32, wParam WPARAM, lParam LPARAM) LRESULT {
	r, _, _ := procSendMessageW.Call(uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam))
	return LRESULT(r)
}

func PostQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func SetCapture(hwnd HWND) HWND {
	r, _, _ := procSetCapture.Call(uintptr(hwnd))
	return HWND(r)
}

func ReleaseCapture() bool {
	r, _, _ := procReleaseCapture.Call()
	return r != 0
}

func GetCapture() HWND {
	r, _, _ := procGetCapture.Call()
	return HWND(r)
}

func SetFocus(hwnd HWND) HWND {
	r, _, _ := procSetFocusW.Call(uintptr(hwnd))
	return HWND(r)
}

func GetFocus() HWND {
	r, _, _ := procGetFocusW.Call()
	return HWND(r)
}

func SetActiveWindow(hwnd HWND) HWND {
	r, _, _ := procSetActiveWindow.Call(uintptr(hwnd))
	return HWND(r)
}

func LoadCursor(hInstance HINSTANCE, cursorName uintptr) HCURSOR {
	r, _, _ := procLoadCursorW.Call(uintptr(hInstance), cursorName)
	return HCURSOR(r)
}

func SetCursorFunc(cursor HCURSOR) HCURSOR {
	r, _, _ := procSetCursor.Call(uintptr(cursor))
	return HCURSOR(r)
}

func OpenClipboard(hwnd HWND) bool {
	r, _, _ := procOpenClipboard.Call(uintptr(hwnd))
	return r != 0
}

func CloseClipboard() bool {
	r, _, _ := procCloseClipboard.Call()
	return r != 0
}

func EmptyClipboard() bool {
	r, _, _ := procEmptyClipboard.Call()
	return r != 0
}

func GetClipboardData(format uint32) HANDLE {
	r, _, _ := procGetClipboardData.Call(uintptr(format))
	return HANDLE(r)
}

func SetClipboardData(format uint32, hMem HANDLE) HANDLE {
	r, _, _ := procSetClipboardData.Call(uintptr(format), uintptr(hMem))
	return HANDLE(r)
}

func GetSystemMetrics(index int32) int32 {
	r, _, _ := procGetSystemMetrics.Call(uintptr(index))
	return int32(r)
}

func GetSysColor(index int32) COLORREF {
	r, _, _ := procGetSysColor.Call(uintptr(index))
	return COLORREF(r)
}

func MapWindowPoints(hwndFrom, hwndTo HWND, points *POINT, count uint32) int32 {
	r, _, _ := procMapWindowPoints.Call(
		uintptr(hwndFrom), uintptr(hwndTo),
		uintptr(unsafe.Pointer(points)), uintptr(count))
	return int32(r)
}

func ScreenToClient(hwnd HWND, point *POINT) bool {
	r, _, _ := procScreenToClient.Call(uintptr(hwnd), uintptr(unsafe.Pointer(point)))
	return r != 0
}

func ClientToScreen(hwnd HWND, point *POINT) bool {
	r, _, _ := procClientToScreen.Call(uintptr(hwnd), uintptr(unsafe.Pointer(point)))
	return r != 0
}

func GetKeyState(vKey int32) int16 {
	r, _, _ := procGetKeyState.Call(uintptr(vKey))
	return int16(r)
}

func GetCursorPos(point *POINT) bool {
	r, _, _ := procGetCursorPos.Call(uintptr(unsafe.Pointer(point)))
	return r != 0
}

func FillRect(hdc HDC, rect *RECT, brush HBRUSH) int32 {
	r, _, _ := procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(rect)), uintptr(brush))
	return int32(r)
}

func IsWindowVisible(hwnd HWND) bool {
	r, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	return r != 0
}

func IsIconic(hwnd HWND) bool {
	r, _, _ := procIsIconic.Call(uintptr(hwnd))
	return r != 0
}

// AdjustWindowRectEx calculates the required size of the window rectangle,
// based on the desired client-rectangle size.
func AdjustWindowRectEx(rect *RECT, style uint32, menu bool, exStyle uint32) bool {
	var bMenu uintptr
	if menu {
		bMenu = 1
	}
	r, _, _ := procAdjustWindowRectEx.Call(uintptr(unsafe.Pointer(rect)),
		uintptr(style), bMenu, uintptr(exStyle))
	return r != 0
}

// UTF16PtrFromString converts a Go string to a null-terminated UTF-16 pointer.
func UTF16PtrFromString(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}
