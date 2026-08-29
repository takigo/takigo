//go:build windows

package windows

import (
	"time"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// WinRawEvent holds the parsed data from a Windows message, stored in
// platform.RawEvent.Data.
type WinRawEvent struct {
	Window       platform.WindowID
	X, Y         int
	RootX, RootY int
	State        uint   // modifier state (X11-style masks)
	KeyCode      uint   // virtual key code
	KeySym       uint64 // translated keysym
	Str          string // character string from WM_CHAR
	Button       uint
	Width        int
	Height       int
	Time         uint64
	MessageType  platform.AtomID
	MessageData  [5]int64
	FocusMode    int
	FocusDetail  int
	// Expose fields
	ExposeX, ExposeY, ExposeW, ExposeH int
	ExposeCount                        int
}

// EventParser implements platform.EventParser for Windows events.
type EventParser struct{}

// Verify at compile time.
var _ platform.EventParser = (*EventParser)(nil)

func (p *EventParser) ParseKeyEvent(ev *platform.RawEvent) platform.KeyEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.KeyEvent{
		EventWindow: raw.Window,
		X:           raw.X,
		Y:           raw.Y,
		RootX:       raw.RootX,
		RootY:       raw.RootY,
		State:       raw.State,
		KeyCode:     raw.KeyCode,
		KeySym:      platform.KeySym(raw.KeySym),
		Str:         raw.Str,
		Time:        platform.Timestamp(raw.Time),
	}
}

func (p *EventParser) ParseKeyEventIM(ev *platform.RawEvent) platform.KeyEvent {
	// On Windows, IM handling is done via WM_CHAR which already provides
	// the composed string. No separate IM parsing needed.
	return p.ParseKeyEvent(ev)
}

func (p *EventParser) ParseButtonEvent(ev *platform.RawEvent) platform.ButtonEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.ButtonEvent{
		EventWindow: raw.Window,
		X:           raw.X,
		Y:           raw.Y,
		RootX:       raw.RootX,
		RootY:       raw.RootY,
		State:       raw.State,
		Button:      raw.Button,
		Time:        platform.Timestamp(raw.Time),
	}
}

func (p *EventParser) ParseMotionEvent(ev *platform.RawEvent) platform.MotionEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.MotionEvent{
		EventWindow: raw.Window,
		X:           raw.X,
		Y:           raw.Y,
		RootX:       raw.RootX,
		RootY:       raw.RootY,
		State:       raw.State,
		Time:        platform.Timestamp(raw.Time),
	}
}

func (p *EventParser) ParseExposeEvent(ev *platform.RawEvent) platform.ExposeEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.ExposeEvent{
		EventWindow: raw.Window,
		X:           raw.ExposeX,
		Y:           raw.ExposeY,
		Width:       raw.ExposeW,
		Height:      raw.ExposeH,
		Count:       raw.ExposeCount,
	}
}

func (p *EventParser) ParseConfigureEvent(ev *platform.RawEvent) platform.ConfigureEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.ConfigureEvent{
		EventWindow: raw.Window,
		X:           raw.X,
		Y:           raw.Y,
		Width:       raw.Width,
		Height:      raw.Height,
	}
}

func (p *EventParser) ParseClientMessageEvent(ev *platform.RawEvent) platform.ClientMessageEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.ClientMessageEvent{
		EventWindow: raw.Window,
		MessageType: raw.MessageType,
		Format:      32,
		Data:        raw.MessageData,
	}
}

func (p *EventParser) ParseCrossingEvent(ev *platform.RawEvent) platform.CrossingEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.CrossingEvent{
		EventWindow: raw.Window,
		X:           raw.X,
		Y:           raw.Y,
		State:       raw.State,
		Time:        platform.Timestamp(raw.Time),
	}
}

func (p *EventParser) ParseDestroyEvent(ev *platform.RawEvent) platform.DestroyEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.DestroyEvent{
		EventWindow: raw.Window,
	}
}

func (p *EventParser) ParseFocusEvent(ev *platform.RawEvent) platform.FocusEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.FocusEvent{
		EventWindow: raw.Window,
		Mode:        raw.FocusMode,
		Detail:      raw.FocusDetail,
	}
}

func (p *EventParser) ParsePropertyEvent(ev *platform.RawEvent) platform.PropertyEvent {
	raw := ev.Data.(*WinRawEvent)
	return platform.PropertyEvent{
		EventWindow: raw.Window,
		Atom:        raw.MessageType,
	}
}

func (p *EventParser) ParseSelectionRequestEvent(ev *platform.RawEvent) platform.SelectionRequestParsed {
	return platform.SelectionRequestParsed{} // clipboard handled differently on Windows
}

func (p *EventParser) ParseSelectionClearEvent(ev *platform.RawEvent) platform.SelectionClearParsed {
	return platform.SelectionClearParsed{}
}

func (p *EventParser) ParseSelectionNotifyEvent(ev *platform.RawEvent) platform.SelectionNotifyParsed {
	return platform.SelectionNotifyParsed{}
}

// --- EventSource implementation ---

func (d *WindowsDisplay) NextEvent() *platform.RawEvent {
	return <-d.eventCh
}

func (d *WindowsDisplay) PeekEvent() *platform.RawEvent {
	select {
	case ev := <-d.eventCh:
		return ev
	default:
		return nil
	}
}

func (d *WindowsDisplay) FilterEvent(ev *platform.RawEvent) bool {
	return false // No IM filtering on Windows currently.
}

// --- WndProc: converts Windows messages to platform events ---

func (d *WindowsDisplay) wndProc(hwnd w32.HWND, msg uint32, wParam w32.WPARAM, lParam w32.LPARAM) w32.LRESULT {
	wid := fromHWND(hwnd)
	now := uint64(time.Now().UnixMilli())

	switch msg {
	case w32.WM_PAINT:
		var ps w32.PAINTSTRUCT
		w32.BeginPaint(hwnd, &ps)
		w32.EndPaint(hwnd, &ps)
		// Post expose event for the paint rect.
		raw := &WinRawEvent{
			Window:  wid,
			ExposeX: int(ps.RcPaint.Left),
			ExposeY: int(ps.RcPaint.Top),
			ExposeW: int(ps.RcPaint.Right - ps.RcPaint.Left),
			ExposeH: int(ps.RcPaint.Bottom - ps.RcPaint.Top),
			Time:    now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.ExposeEvent_,
			EventWindow: wid,
		})
		return 0

	case w32.WM_SIZE:
		w := int(w32.LOWORD(uintptr(lParam)))
		h := int(w32.HIWORD(uintptr(lParam)))
		// Ignore zero-dimension events (minimized windows, intermediate states).
		// A zero width or height would cause layout to collapse.
		if w <= 0 || h <= 0 {
			return 0
		}
		raw := &WinRawEvent{
			Window: wid,
			Width:  w,
			Height: h,
			Time:   now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.ConfigureNotifyEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_MOVE:
		x := int(w32.GET_X_LPARAM(lParam))
		y := int(w32.GET_Y_LPARAM(lParam))
		// Include current client size so ConfigureNotify has valid dimensions.
		var rect w32.RECT
		w32.GetClientRect(hwnd, &rect)
		cw := int(rect.Right - rect.Left)
		ch := int(rect.Bottom - rect.Top)
		if cw <= 0 || ch <= 0 {
			return 0
		}
		raw := &WinRawEvent{
			Window: wid,
			X:      x,
			Y:      y,
			Width:  cw,
			Height: ch,
			Time:   now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.ConfigureNotifyEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_KEYDOWN, w32.WM_SYSKEYDOWN:
		vk := uint(wParam)
		state := d.getModifierState()
		keysym := vkToKeySym(vk)
		raw := &WinRawEvent{
			Window:  wid,
			KeyCode: vk,
			KeySym:  keysym,
			State:   state,
			Time:    now,
		}
		// Check if we have a pending char from WM_CHAR.
		if d.hasChar {
			raw.Str = string(d.pendingChar)
			d.hasChar = false
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.KeyPressEvent,
			EventWindow: wid,
		})
		if msg == w32.WM_SYSKEYDOWN {
			return w32.DefWindowProc(hwnd, msg, wParam, lParam)
		}
		return 0

	case w32.WM_KEYUP, w32.WM_SYSKEYUP:
		vk := uint(wParam)
		state := d.getModifierState()
		keysym := vkToKeySym(vk)
		raw := &WinRawEvent{
			Window:  wid,
			KeyCode: vk,
			KeySym:  keysym,
			State:   state,
			Time:    now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.KeyReleaseEvent,
			EventWindow: wid,
		})
		if msg == w32.WM_SYSKEYUP {
			return w32.DefWindowProc(hwnd, msg, wParam, lParam)
		}
		return 0

	case w32.WM_CHAR:
		// Store the character for the next WM_KEYDOWN.
		ch := rune(wParam)
		d.pendingChar = ch
		d.hasChar = true
		// Also post a key press for the character if no pending KEYDOWN consumed it.
		// This handles standalone WM_CHAR from IME.
		keysym := uint64(ch)
		if ch < 128 {
			keysym = uint64(ch) // ASCII maps directly to keysym
		}
		raw := &WinRawEvent{
			Window: wid,
			KeySym: keysym,
			Str:    string(ch),
			State:  d.getModifierState(),
			Time:   now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.KeyPressEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_MOUSEMOVE:
		x := int(w32.GET_X_LPARAM(lParam))
		y := int(w32.GET_Y_LPARAM(lParam))
		rootX, rootY := d.clientToScreen(hwnd, x, y)
		state := d.mouseModifierState(wParam)
		raw := &WinRawEvent{
			Window: wid,
			X:      x,
			Y:      y,
			RootX:  rootX,
			RootY:  rootY,
			State:  state,
			Time:   now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.MotionNotifyEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_LBUTTONDOWN:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 1, platform.ButtonPressEvent)
		return 0
	case w32.WM_LBUTTONUP:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 1, platform.ButtonReleaseEvent)
		return 0
	case w32.WM_MBUTTONDOWN:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 2, platform.ButtonPressEvent)
		return 0
	case w32.WM_MBUTTONUP:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 2, platform.ButtonReleaseEvent)
		return 0
	case w32.WM_RBUTTONDOWN:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 3, platform.ButtonPressEvent)
		return 0
	case w32.WM_RBUTTONUP:
		d.handleMouseButton(hwnd, wid, lParam, wParam, now, 3, platform.ButtonReleaseEvent)
		return 0

	case w32.WM_MOUSEWHEEL:
		delta := w32.GET_WHEEL_DELTA_WPARAM(wParam)
		var button uint
		if delta > 0 {
			button = 4 // scroll up
		} else {
			button = 5 // scroll down
		}
		// WM_MOUSEWHEEL coords are screen-relative.
		sx := int(w32.GET_X_LPARAM(lParam))
		sy := int(w32.GET_Y_LPARAM(lParam))
		pt := w32.POINT{X: int32(sx), Y: int32(sy)}
		w32.ScreenToClient(hwnd, &pt)

		state := d.mouseModifierState(wParam)
		raw := &WinRawEvent{
			Window: wid,
			X:      int(pt.X),
			Y:      int(pt.Y),
			RootX:  sx,
			RootY:  sy,
			Button: button,
			State:  state,
			Time:   now,
		}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.ButtonPressEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_SETFOCUS:
		d.focusWindow = wid
		raw := &WinRawEvent{Window: wid, Time: now}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.FocusInEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_KILLFOCUS:
		raw := &WinRawEvent{Window: wid, Time: now}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.FocusOutEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_DESTROY:
		raw := &WinRawEvent{Window: wid, Time: now}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.DestroyNotifyEvent,
			EventWindow: wid,
		})
		return 0

	case w32.WM_CLOSE:
		// Convert to WM_DELETE_WINDOW client message, matching the X11 convention:
		// message_type = WM_PROTOCOLS, data[0] = WM_DELETE_WINDOW.
		deleteAtom := d.internAtom("WM_DELETE_WINDOW", false)
		protocolsAtom := d.internAtom("WM_PROTOCOLS", false)
		raw := &WinRawEvent{
			Window:      wid,
			MessageType: protocolsAtom,
			Time:        now,
		}
		raw.MessageData[0] = int64(deleteAtom)
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.ClientMessageEvent_,
			EventWindow: wid,
		})
		return 0

	case w32.WM_ERASEBKGND:
		// Paint background with the window's background pixel to prevent flicker.
		info := d.getWindowInfo(hwnd)
		if info != nil {
			hdc := w32.HDC(uintptr(wParam))
			var rect w32.RECT
			w32.GetClientRect(hwnd, &rect)
			brush := w32.CreateSolidBrush(pixelToCOLORREF(info.bgPixel))
			w32.FillRect(hdc, &rect, brush)
			w32.DeleteObject(w32.HGDIOBJ(brush))
		}
		return 1

	case w32.WM_MOUSELEAVE:
		raw := &WinRawEvent{Window: wid, Time: now}
		d.postEvent(&platform.RawEvent{
			Data:        raw,
			EventType:   platform.LeaveNotifyEvent,
			EventWindow: wid,
		})
		return 0
	}

	return w32.DefWindowProc(hwnd, msg, wParam, lParam)
}

func (d *WindowsDisplay) handleMouseButton(hwnd w32.HWND, wid platform.WindowID,
	lParam w32.LPARAM, wParam w32.WPARAM, now uint64, button uint, eventType int) {
	x := int(w32.GET_X_LPARAM(lParam))
	y := int(w32.GET_Y_LPARAM(lParam))
	rootX, rootY := d.clientToScreen(hwnd, x, y)
	state := d.mouseModifierState(wParam)

	raw := &WinRawEvent{
		Window: wid,
		X:      x,
		Y:      y,
		RootX:  rootX,
		RootY:  rootY,
		Button: button,
		State:  state,
		Time:   now,
	}
	d.postEvent(&platform.RawEvent{
		Data:        raw,
		EventType:   eventType,
		EventWindow: wid,
	})
}

func (d *WindowsDisplay) clientToScreen(hwnd w32.HWND, x, y int) (int, int) {
	pt := w32.POINT{X: int32(x), Y: int32(y)}
	w32.ClientToScreen(hwnd, &pt)
	return int(pt.X), int(pt.Y)
}

// getModifierState returns the current keyboard modifier state as X11-style masks.
func (d *WindowsDisplay) getModifierState() uint {
	var state uint
	if w32.GetKeyState(w32.VK_SHIFT) < 0 {
		state |= platform.ShiftMask
	}
	if w32.GetKeyState(w32.VK_CONTROL) < 0 {
		state |= platform.ControlMask
	}
	if w32.GetKeyState(w32.VK_MENU) < 0 {
		state |= platform.Mod1Mask // Alt
	}
	if w32.GetKeyState(w32.VK_CAPITAL)&1 != 0 {
		state |= platform.LockMask
	}
	if w32.GetKeyState(w32.VK_LWIN) < 0 || w32.GetKeyState(w32.VK_RWIN) < 0 {
		state |= platform.Mod4Mask // Super/Win
	}
	return state
}

// mouseModifierState returns X11-style state from WM_*BUTTON wParam.
func (d *WindowsDisplay) mouseModifierState(wParam w32.WPARAM) uint {
	state := d.getModifierState()
	wp := uint(wParam)
	if wp&w32.MK_LBUTTON != 0 {
		state |= platform.Button1Mask
	}
	if wp&w32.MK_MBUTTON != 0 {
		state |= platform.Button2Mask
	}
	if wp&w32.MK_RBUTTON != 0 {
		state |= platform.Button3Mask
	}
	return state
}

// vkToKeySym translates a Windows virtual key code to an X11-style keysym.
// This is ported from tk/win/tkWinKey.c's keymap table.
func vkToKeySym(vk uint) uint64 {
	switch vk {
	case w32.VK_CANCEL:
		return 0xff69 // XK_Cancel
	case w32.VK_BACK:
		return 0xff08 // XK_BackSpace
	case w32.VK_TAB:
		return 0xff09 // XK_Tab
	case w32.VK_CLEAR:
		return 0xff0b // XK_Clear
	case w32.VK_RETURN:
		return 0xff0d // XK_Return
	case w32.VK_SHIFT:
		return 0xffe1 // XK_Shift_L
	case w32.VK_CONTROL:
		return 0xffe3 // XK_Control_L
	case w32.VK_MENU:
		return 0xffe9 // XK_Alt_L
	case w32.VK_PAUSE:
		return 0xff13 // XK_Pause
	case w32.VK_CAPITAL:
		return 0xffe5 // XK_Caps_Lock
	case w32.VK_ESCAPE:
		return 0xff1b // XK_Escape
	case w32.VK_SPACE:
		return 0x0020 // XK_space
	case w32.VK_PRIOR:
		return 0xff55 // XK_Prior (Page Up)
	case w32.VK_NEXT:
		return 0xff56 // XK_Next (Page Down)
	case w32.VK_END:
		return 0xff57 // XK_End
	case w32.VK_HOME:
		return 0xff50 // XK_Home
	case w32.VK_LEFT:
		return 0xff51 // XK_Left
	case w32.VK_UP:
		return 0xff52 // XK_Up
	case w32.VK_RIGHT:
		return 0xff53 // XK_Right
	case w32.VK_DOWN:
		return 0xff54 // XK_Down
	case w32.VK_SELECT:
		return 0xff60 // XK_Select
	case w32.VK_PRINT:
		return 0xff61 // XK_Print
	case w32.VK_EXECUTE:
		return 0xff62 // XK_Execute
	case w32.VK_INSERT:
		return 0xff63 // XK_Insert
	case w32.VK_DELETE:
		return 0xffff // XK_Delete
	case w32.VK_HELP:
		return 0xff6a // XK_Help
	case w32.VK_LWIN:
		return 0xffeb // XK_Super_L
	case w32.VK_RWIN:
		return 0xffec // XK_Super_R
	case w32.VK_APPS:
		return 0xff67 // XK_Menu
	case w32.VK_F1:
		return 0xffbe // XK_F1
	case w32.VK_F2:
		return 0xffbf
	case w32.VK_F3:
		return 0xffc0
	case w32.VK_F4:
		return 0xffc1
	case w32.VK_F5:
		return 0xffc2
	case w32.VK_F6:
		return 0xffc3
	case w32.VK_F7:
		return 0xffc4
	case w32.VK_F8:
		return 0xffc5
	case w32.VK_F9:
		return 0xffc6
	case w32.VK_F10:
		return 0xffc7
	case w32.VK_F11:
		return 0xffc8
	case w32.VK_F12:
		return 0xffc9
	case w32.VK_F13:
		return 0xffca
	case w32.VK_F14:
		return 0xffcb
	case w32.VK_F15:
		return 0xffcc
	case w32.VK_F16:
		return 0xffcd
	case w32.VK_F17:
		return 0xffce
	case w32.VK_F18:
		return 0xffcf
	case w32.VK_F19:
		return 0xffd0
	case w32.VK_F20:
		return 0xffd1
	case w32.VK_F21:
		return 0xffd2
	case w32.VK_F22:
		return 0xffd3
	case w32.VK_F23:
		return 0xffd4
	case w32.VK_F24:
		return 0xffd5
	case w32.VK_NUMLOCK:
		return 0xff7f // XK_Num_Lock
	case w32.VK_SCROLL:
		return 0xff14 // XK_Scroll_Lock
	default:
		// For 0-9, A-Z: the VK code equals the ASCII code.
		if vk >= 0x30 && vk <= 0x39 {
			return uint64(vk) // '0'-'9'
		}
		if vk >= 0x41 && vk <= 0x5A {
			return uint64(vk + 0x20) // lowercase 'a'-'z'
		}
		// Numpad keys.
		if vk >= w32.VK_NUMPAD0 && vk <= w32.VK_NUMPAD9 {
			return uint64(0xffb0 + (vk - w32.VK_NUMPAD0)) // XK_KP_0 .. XK_KP_9
		}
		switch vk {
		case w32.VK_MULTIPLY:
			return 0xffaa // XK_KP_Multiply
		case w32.VK_ADD:
			return 0xffab // XK_KP_Add
		case w32.VK_SEPARATOR:
			return 0xffac // XK_KP_Separator
		case w32.VK_SUBTRACT:
			return 0xffad // XK_KP_Subtract
		case w32.VK_DECIMAL:
			return 0xffae // XK_KP_Decimal
		case w32.VK_DIVIDE:
			return 0xffaf // XK_KP_Divide
		}
		return 0 // NoSymbol
	}
}
