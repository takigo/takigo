//go:build windows

package windows

import (
	"syscall"
	"time"

	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// The tray icons port tk/win/tkWinSysTray.c: each icon has a hidden
// window that receives its callback message, which becomes a ButtonPress
// event for that window.

// trayIcon is the Shell_NotifyIconW state of one icon.
type trayIcon struct {
	hicon w32.HICON
}

// trayData fills the fields every Shell_NotifyIconW call needs.
func trayData(hwnd w32.HWND) w32.NOTIFYICONDATAW {
	return w32.NOTIFYICONDATAW{HWnd: hwnd, UID: 1}
}

// AddTrayIcon creates the icon's hidden window and adds the icon.
func (d *WindowsDisplay) AddTrayIcon(icon *platform.IconImage, tooltip string) platform.WindowID {
	hwnd := w32.CreateWindowEx(0, syscall.StringToUTF16Ptr(windowClassName),
		syscall.StringToUTF16Ptr(""), 0, 0, 0, 1, 1, 0, 0, d.hInstance, nil)
	if hwnd == 0 {
		return 0
	}
	t := &trayIcon{}
	nid := trayData(hwnd)
	nid.UFlags = w32.NIF_MESSAGE | w32.NIF_TIP
	nid.UCallbackMessage = w32.WM_APP_TRAY
	if icon != nil {
		t.hicon = createIcon(icon)
		nid.HIcon = t.hicon
		nid.UFlags |= w32.NIF_ICON
	}
	w32.CopyUTF16(nid.SzTip[:], tooltip)
	if !w32.Shell_NotifyIcon(w32.NIM_ADD, &nid) {
		if t.hicon != 0 {
			w32.DestroyIcon(t.hicon)
		}
		w32.DestroyWindow(hwnd)
		return 0
	}
	d.windowMu.Lock()
	d.windowData[hwnd] = &windowInfo{hwnd: hwnd, tray: t}
	d.windowMu.Unlock()
	return fromHWND(hwnd)
}

func (d *WindowsDisplay) trayOf(w platform.WindowID) (w32.HWND, *trayIcon) {
	hwnd := toHWND(w)
	d.windowMu.RLock()
	defer d.windowMu.RUnlock()
	if info := d.windowData[hwnd]; info != nil && info.tray != nil {
		return hwnd, info.tray
	}
	return 0, nil
}

// ModifyTrayIcon replaces the icon and the tooltip (NIM_MODIFY).
func (d *WindowsDisplay) ModifyTrayIcon(w platform.WindowID, icon *platform.IconImage, tooltip string) {
	hwnd, t := d.trayOf(w)
	if t == nil {
		return
	}
	nid := trayData(hwnd)
	nid.UFlags = w32.NIF_TIP
	w32.CopyUTF16(nid.SzTip[:], tooltip)
	old := w32.HICON(0)
	if icon != nil {
		old, t.hicon = t.hicon, createIcon(icon)
		nid.HIcon = t.hicon
		nid.UFlags |= w32.NIF_ICON
	}
	w32.Shell_NotifyIcon(w32.NIM_MODIFY, &nid)
	if old != 0 {
		w32.DestroyIcon(old)
	}
}

// RemoveTrayIcon deletes the icon and its window.
func (d *WindowsDisplay) RemoveTrayIcon(w platform.WindowID) {
	hwnd, t := d.trayOf(w)
	if t == nil {
		return
	}
	nid := trayData(hwnd)
	w32.Shell_NotifyIcon(w32.NIM_DELETE, &nid)
	if t.hicon != 0 {
		w32.DestroyIcon(t.hicon)
	}
	d.windowMu.Lock()
	delete(d.windowData, hwnd)
	d.windowMu.Unlock()
	w32.DestroyWindow(hwnd)
}

// TrayNotify shows a balloon on the icon, as WinSysNotifyCmd does: NIF_INFO
// with NIIF_USER, so the balloon shows the icon itself.
func (d *WindowsDisplay) TrayNotify(w platform.WindowID, title, message string) error {
	hwnd, t := d.trayOf(w)
	if t == nil {
		return platform.ErrNoTray
	}
	nid := trayData(hwnd)
	nid.UFlags = w32.NIF_INFO
	nid.DwInfoFlags = w32.NIIF_USER
	w32.CopyUTF16(nid.SzInfoTitle[:], title)
	w32.CopyUTF16(nid.SzInfo[:], message)
	if !w32.Shell_NotifyIcon(w32.NIM_MODIFY, &nid) {
		return syscall.Errno(w32.GetLastError())
	}
	return nil
}

// trayCallback turns the icon's callback message into a ButtonPress event
// at the pointer. Like systray.tcl's _win_callback, a double click counts
// as a click.
func (d *WindowsDisplay) trayCallback(hwnd w32.HWND, lParam w32.LPARAM) {
	var button uint
	switch uint32(lParam) & 0xffff {
	case w32.WM_LBUTTONDOWN, w32.WM_LBUTTONDBLCLK:
		button = 1
	case w32.WM_RBUTTONDOWN, w32.WM_RBUTTONDBLCLK:
		button = 3
	default:
		return
	}
	var pt w32.POINT
	w32.GetCursorPos(&pt)
	// The window must be in front for a menu posted now to close when
	// the user clicks elsewhere (KB135788).
	w32.SetForegroundWindow(hwnd)
	wid := fromHWND(hwnd)
	d.postEvent(&platform.RawEvent{
		Data: &WinRawEvent{Window: wid, X: int(pt.X), Y: int(pt.Y), RootX: int(pt.X), RootY: int(pt.Y),
			Button: button, Time: uint64(time.Now().UnixMilli())},
		EventType:   platform.ButtonPressEvent,
		EventWindow: wid,
	})
}

var _ platform.TrayManager = (*WindowsDisplay)(nil)
