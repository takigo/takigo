//go:build windows

package windows

import (
	"time"
	"unicode/utf16"
	"unicode/utf8"

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
		// TranslateMessage queues this key's WM_CHAR behind the WM_KEYDOWN;
		// take it now, as tkWinX.c GetTranslatedKey does, so one KeyPress
		// carries both the keysym and the text.
		// WM_SYSCHAR stays queued: DefWindowProc needs it for Alt+Space.
		if msg == w32.WM_KEYDOWN {
			if str := d.takeTranslatedChars(); str != "" {
				raw.Str = str
				if keysym < 0x100 {
					r, _ := utf8.DecodeRuneInString(str)
					if ks := platform.RuneToKeySym(r); ks != 0 {
						raw.KeySym = uint64(ks)
					}
				}
			}
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
		// A WM_CHAR not claimed by its WM_KEYDOWN (IME commits, pasted
		// input): deliver it as a KeyPress, joining UTF-16 surrogate pairs.
		unit := uint16(wParam)
		if utf16.IsSurrogate(rune(unit)) && unit < 0xdc00 {
			d.highSurrogate = unit
			return 0
		}
		r := rune(unit)
		if d.highSurrogate != 0 {
			r = utf16.DecodeRune(rune(d.highSurrogate), r)
			d.highSurrogate = 0
		}
		if r == utf8.RuneError {
			return 0
		}
		raw := &WinRawEvent{
			Window: wid,
			KeySym: uint64(platform.RuneToKeySym(r)),
			Str:    string(r),
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

// vkToXK maps Windows virtual-key codes to X11-style keysyms.
// Ported from tk/win/tkWinKey.c's keymap table; comments preserve
// the original VK_* and XK_* names for traceability.
var vkToXK = map[uint]uint64{
	w32.VK_CANCEL:    0xff69, // XK_Cancel
	w32.VK_BACK:      0xff08, // XK_BackSpace
	w32.VK_TAB:       0xff09, // XK_Tab
	w32.VK_CLEAR:     0xff0b, // XK_Clear
	w32.VK_RETURN:    0xff0d, // XK_Return
	w32.VK_SHIFT:     0xffe1, // XK_Shift_L
	w32.VK_CONTROL:   0xffe3, // XK_Control_L
	w32.VK_MENU:      0xffe9, // XK_Alt_L
	w32.VK_PAUSE:     0xff13, // XK_Pause
	w32.VK_CAPITAL:   0xffe5, // XK_Caps_Lock
	w32.VK_ESCAPE:    0xff1b, // XK_Escape
	w32.VK_SPACE:     0x0020, // XK_space
	w32.VK_PRIOR:     0xff55, // XK_Prior (Page Up)
	w32.VK_NEXT:      0xff56, // XK_Next (Page Down)
	w32.VK_END:       0xff57, // XK_End
	w32.VK_HOME:      0xff50, // XK_Home
	w32.VK_LEFT:      0xff51, // XK_Left
	w32.VK_UP:        0xff52, // XK_Up
	w32.VK_RIGHT:     0xff53, // XK_Right
	w32.VK_DOWN:      0xff54, // XK_Down
	w32.VK_SELECT:    0xff60, // XK_Select
	w32.VK_PRINT:     0xff61, // XK_Print
	w32.VK_EXECUTE:   0xff62, // XK_Execute
	w32.VK_INSERT:    0xff63, // XK_Insert
	w32.VK_DELETE:    0xffff, // XK_Delete
	w32.VK_HELP:      0xff6a, // XK_Help
	w32.VK_LWIN:      0xffeb, // XK_Super_L
	w32.VK_RWIN:      0xffec, // XK_Super_R
	w32.VK_APPS:      0xff67, // XK_Menu
	w32.VK_NUMLOCK:   0xff7f, // XK_Num_Lock
	w32.VK_SCROLL:    0xff14, // XK_Scroll_Lock
	w32.VK_MULTIPLY:  0xffaa, // XK_KP_Multiply
	w32.VK_ADD:       0xffab, // XK_KP_Add
	w32.VK_SEPARATOR: 0xffac, // XK_KP_Separator
	w32.VK_SUBTRACT:  0xffad, // XK_KP_Subtract
	w32.VK_DECIMAL:   0xffae, // XK_KP_Decimal
	w32.VK_DIVIDE:    0xffaf, // XK_KP_Divide
	w32.VK_F1:        0xffbe, // XK_F1
	w32.VK_F2:        0xffbf, // XK_F2
	w32.VK_F3:        0xffc0, // XK_F3
	w32.VK_F4:        0xffc1, // XK_F4
	w32.VK_F5:        0xffc2, // XK_F5
	w32.VK_F6:        0xffc3, // XK_F6
	w32.VK_F7:        0xffc4, // XK_F7
	w32.VK_F8:        0xffc5, // XK_F8
	w32.VK_F9:        0xffc6, // XK_F9
	w32.VK_F10:       0xffc7, // XK_F10
	w32.VK_F11:       0xffc8, // XK_F11
	w32.VK_F12:       0xffc9, // XK_F12
	w32.VK_F13:       0xffca, // XK_F13
	w32.VK_F14:       0xffcb, // XK_F14
	w32.VK_F15:       0xffcc, // XK_F15
	w32.VK_F16:       0xffcd, // XK_F16
	w32.VK_F17:       0xffce, // XK_F17
	w32.VK_F18:       0xffcf, // XK_F18
	w32.VK_F19:       0xffd0, // XK_F19
	w32.VK_F20:       0xffd1, // XK_F20
	w32.VK_F21:       0xffd2, // XK_F21
	w32.VK_F22:       0xffd3, // XK_F22
	w32.VK_F23:       0xffd4, // XK_F23
	w32.VK_F24:       0xffd5, // XK_F24
}

// vkToKeySym translates a Windows virtual key code to an X11-style keysym.
// This is ported from tk/win/tkWinKey.c's keymap table.
func vkToKeySym(vk uint) uint64 {
	if sym, ok := vkToXK[vk]; ok {
		return sym
	}
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
	return 0 // NoSymbol
}

// takeTranslatedChars removes the WM_CHAR messages queued directly
// behind the current WM_KEYDOWN and returns their text, joining
// surrogate pairs.
func (d *WindowsDisplay) takeTranslatedChars() string {
	var units []uint16
	var msg w32.MSG
	for w32.PeekMessage(&msg, 0, 0, 0, w32.PM_NOREMOVE) {
		if msg.Message != w32.WM_CHAR {
			break
		}
		w32.PeekMessage(&msg, 0, msg.Message, msg.Message, w32.PM_REMOVE)
		units = append(units, uint16(msg.WParam))
	}
	return string(utf16.Decode(units))
}
