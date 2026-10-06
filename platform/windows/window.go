//go:build windows

package windows

import (
	"syscall"

	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// --- WindowManager implementation ---

func (d *WindowsDisplay) CreateWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint,
	depth int, class uint, valueMask uint64, attrs *platform.WindowAttrs) platform.WindowID {

	parentHWND := toHWND(parent)
	isTopLevel := false
	isOverride := false

	// Determine window style.
	// On X11, windows whose parent is the root window are top-level windows
	// managed by the window manager. On Windows, we must use
	// WS_OVERLAPPEDWINDOW (or WS_POPUP) for these instead of WS_CHILD.
	style := uint32(w32.WS_CHILD | w32.WS_CLIPSIBLINGS | w32.WS_CLIPCHILDREN)
	exStyle := uint32(0)

	if attrs != nil && attrs.OverrideRedirect {
		// Override-redirect → popup (for menus, tooltips).
		style = w32.WS_POPUP | w32.WS_CLIPSIBLINGS | w32.WS_CLIPCHILDREN
		exStyle = w32.WS_EX_TOPMOST | w32.WS_EX_TOOLWINDOW
		parentHWND = 0 // no parent for popups
		isOverride = true
	} else if parentHWND == d.rootHWND {
		// Parent is the display root → this is a top-level window.
		style = w32.WS_OVERLAPPEDWINDOW | w32.WS_CLIPSIBLINGS | w32.WS_CLIPCHILDREN
		exStyle = w32.WS_EX_APPWINDOW
		parentHWND = 0 // top-level windows have no parent HWND
		isTopLevel = true

		// Use CW_USEDEFAULT for initial placement of top-level windows
		// unless specific coordinates were provided.
		if x == 0 && y == 0 {
			x = int(w32.CW_USEDEFAULT)
			y = int(w32.CW_USEDEFAULT)
		}
	}

	// InputOnly windows: create without visible rendering.
	if class == platform.InputOnly {
		exStyle |= w32.WS_EX_LAYERED | w32.WS_EX_NOACTIVATE
	}

	bgPixel := uint64(0x00FFFFFF) // default white
	if attrs != nil {
		bgPixel = attrs.BackgroundPixel
	}

	// For top-level windows, adjust the outer size so that the client area
	// matches the requested width/height. WS_OVERLAPPEDWINDOW adds title bar
	// and borders which reduce the usable client area.
	outerW, outerH := int32(width), int32(height)
	if isTopLevel {
		rect := w32.RECT{Left: 0, Top: 0, Right: int32(width), Bottom: int32(height)}
		w32.AdjustWindowRectEx(&rect, style, false, exStyle)
		outerW = rect.Right - rect.Left
		outerH = rect.Bottom - rect.Top
	}

	className := syscall.StringToUTF16Ptr(windowClassName)
	hwnd := w32.CreateWindowEx(exStyle, className, nil, style,
		int32(x), int32(y), outerW, outerH,
		parentHWND, 0, d.hInstance, nil)
	if hwnd == 0 {
		return 0
	}
	// Windows puts a new child at the bottom of its siblings, X on top;
	// raise it as Tk_MakeWindow does (tk/win/tkWinWindow.c).
	if !isTopLevel && !isOverride {
		w32.SetWindowPos(hwnd, w32.HWND_TOP, 0, 0, 0, 0,
			w32.SWP_NOMOVE|w32.SWP_NOSIZE|w32.SWP_NOACTIVATE)
	}

	var eventMask int64
	if attrs != nil {
		eventMask = attrs.EventMask
	}

	d.windowMu.Lock()
	d.windowData[hwnd] = &windowInfo{
		hwnd:       hwnd,
		parent:     toHWND(parent),
		eventMask:  eventMask,
		bgPixel:    bgPixel,
		isOverride: isOverride,
		isTopLevel: isTopLevel,
	}
	d.windowMu.Unlock()

	return fromHWND(hwnd)
}

func (d *WindowsDisplay) CreateSimpleWindow(parent platform.WindowID, x, y int, width, height, borderWidth uint,
	border, background uint64) platform.WindowID {
	attrs := &platform.WindowAttrs{
		BackgroundPixel: background,
		BorderPixel:     border,
	}
	return d.CreateWindow(parent, x, y, width, height, borderWidth, 0, platform.InputOutput, 0, attrs)
}

func (d *WindowsDisplay) DestroyWindow(w platform.WindowID) {
	hwnd := toHWND(w)
	w32.DestroyWindow(hwnd)
	d.forgetWindow(hwnd)
}

// forgetWindow drops the state kept for hwnd (WM_DESTROY also calls it
// for each destroyed descendant).
func (d *WindowsDisplay) forgetWindow(hwnd w32.HWND) {
	d.windowMu.Lock()
	delete(d.windowData, hwnd)
	d.windowMu.Unlock()
	if d.hoverHWND == hwnd {
		d.hoverHWND = 0
	}
	propsMu.Lock()
	delete(propsDB, fromHWND(hwnd))
	propsMu.Unlock()
}

func (d *WindowsDisplay) MapWindow(w platform.WindowID) {
	hwnd := toHWND(w)
	info := d.getWindowInfo(hwnd)
	if info != nil && info.isOverride {
		w32.ShowWindow(hwnd, w32.SW_SHOWNOACTIVATE)
	} else if info != nil && info.isTopLevel {
		w32.ShowWindow(hwnd, w32.SW_SHOWNORMAL)
	} else {
		w32.ShowWindow(hwnd, w32.SW_SHOW)
	}
	w32.UpdateWindow(hwnd)
}

func (d *WindowsDisplay) MapRaised(w platform.WindowID) {
	hwnd := toHWND(w)
	w32.ShowWindow(hwnd, w32.SW_SHOW)
	w32.SetWindowPos(hwnd, w32.HWND_TOP, 0, 0, 0, 0,
		w32.SWP_NOMOVE|w32.SWP_NOSIZE|w32.SWP_SHOWWINDOW)
}

func (d *WindowsDisplay) UnmapWindow(w platform.WindowID) {
	w32.ShowWindow(toHWND(w), w32.SW_HIDE)
}

func (d *WindowsDisplay) RaiseWindow(w platform.WindowID) {
	w32.SetWindowPos(toHWND(w), w32.HWND_TOP, 0, 0, 0, 0,
		w32.SWP_NOMOVE|w32.SWP_NOSIZE|w32.SWP_NOACTIVATE)
}

func (d *WindowsDisplay) LowerWindow(w platform.WindowID) {
	w32.SetWindowPos(toHWND(w), w32.HWND_BOTTOM, 0, 0, 0, 0,
		w32.SWP_NOMOVE|w32.SWP_NOSIZE|w32.SWP_NOACTIVATE)
}

func (d *WindowsDisplay) MoveWindow(w platform.WindowID, x, y int) {
	w32.SetWindowPos(toHWND(w), 0, int32(x), int32(y), 0, 0,
		w32.SWP_NOSIZE|w32.SWP_NOZORDER|w32.SWP_NOACTIVATE)
}

func (d *WindowsDisplay) ResizeWindow(w platform.WindowID, width, height uint) {
	hwnd := toHWND(w)
	outerW, outerH := int32(width), int32(height)

	// For top-level windows, convert client area size to outer window size.
	info := d.getWindowInfo(hwnd)
	if info != nil && info.isTopLevel {
		style := uint32(w32.GetWindowLongPtr(hwnd, w32.GWL_STYLE))
		exStyle := uint32(w32.GetWindowLongPtr(hwnd, w32.GWL_EXSTYLE))
		rect := w32.RECT{Left: 0, Top: 0, Right: int32(width), Bottom: int32(height)}
		w32.AdjustWindowRectEx(&rect, style, false, exStyle)
		outerW = rect.Right - rect.Left
		outerH = rect.Bottom - rect.Top
	}

	w32.SetWindowPos(hwnd, 0, 0, 0, outerW, outerH,
		w32.SWP_NOMOVE|w32.SWP_NOZORDER|w32.SWP_NOACTIVATE)
}

func (d *WindowsDisplay) MoveResizeWindow(w platform.WindowID, x, y int, width, height uint) {
	hwnd := toHWND(w)
	outerW, outerH := int32(width), int32(height)

	// For top-level windows, convert client area size to outer window size.
	info := d.getWindowInfo(hwnd)
	if info != nil && info.isTopLevel {
		style := uint32(w32.GetWindowLongPtr(hwnd, w32.GWL_STYLE))
		exStyle := uint32(w32.GetWindowLongPtr(hwnd, w32.GWL_EXSTYLE))
		rect := w32.RECT{Left: 0, Top: 0, Right: int32(width), Bottom: int32(height)}
		w32.AdjustWindowRectEx(&rect, style, false, exStyle)
		outerW = rect.Right - rect.Left
		outerH = rect.Bottom - rect.Top
	}

	w32.MoveWindow(hwnd, int32(x), int32(y), outerW, outerH, true)
}

func (d *WindowsDisplay) SelectInput(w platform.WindowID, eventMask int64) {
	hwnd := toHWND(w)
	d.windowMu.Lock()
	if info, ok := d.windowData[hwnd]; ok {
		info.eventMask = eventMask
	}
	d.windowMu.Unlock()
}

func (d *WindowsDisplay) StoreName(w platform.WindowID, name string) {
	w32.SetWindowText(toHWND(w), syscall.StringToUTF16Ptr(name))
}

// ChildAt is not implemented: nothing on Windows walks the window tree.
func (d *WindowsDisplay) ChildAt(platform.WindowID, int, int) platform.WindowID { return 0 }

func (d *WindowsDisplay) TranslateCoordinates(src, dst platform.WindowID, srcX, srcY int) (int, int) {
	pt := w32.POINT{X: int32(srcX), Y: int32(srcY)}
	w32.MapWindowPoints(toHWND(src), toHWND(dst), &pt, 1)
	return int(pt.X), int(pt.Y)
}
