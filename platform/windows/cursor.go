//go:build windows

package windows

import (
	"github.com/takigo/takigo/cursor"
	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// shapeToWinCursor maps abstract cursor shapes to Windows IDC_* constants.
var shapeToWinCursor = map[cursor.Shape]uint16{
	cursor.Arrow:             w32.IDC_ARROW,
	cursor.Crosshair:         w32.IDC_CROSS,
	cursor.Fleur:             w32.IDC_SIZEALL,
	cursor.Hand1:             w32.IDC_HAND,
	cursor.Hand2:             w32.IDC_HAND,
	cursor.LeftPtr:           w32.IDC_ARROW,
	cursor.Plus:              w32.IDC_CROSS,
	cursor.QuestionArrow:     w32.IDC_HELP,
	cursor.SBHDoubleArrow:    w32.IDC_SIZEWE,
	cursor.SBVDoubleArrow:    w32.IDC_SIZENS,
	cursor.SizingAngle:       w32.IDC_SIZENWSE,
	cursor.TopLeftArrow:      w32.IDC_ARROW,
	cursor.Watch:             w32.IDC_WAIT,
	cursor.XTerm:             w32.IDC_IBEAM,
	cursor.BottomRightCorner: w32.IDC_SIZENWSE,
}

// --- CursorManager implementation ---

func (d *WindowsDisplay) CreateFontCursor(shape uint) platform.CursorID {
	// Map X11 cursor font index to Windows cursor.
	// For now, treat the shape as a generic IDC value.
	hcursor := w32.LoadCursor(0, w32.MAKEINTRESOURCE(w32.IDC_ARROW))
	return platform.CursorID(uintptr(hcursor))
}

// Cursors are per window: the window class is shared by every takigo
// window, so setting its cursor changed the whole application. The window
// procedure applies the cursor on WM_SETCURSOR, as tkWinPointer.c does.
func (d *WindowsDisplay) DefineCursor(w platform.WindowID, cursorID platform.CursorID) {
	d.setWindowCursor(toHWND(w), w32.HCURSOR(uintptr(cursorID)))
}

// setWindowCursor records hwnd's cursor and shows it at once if the
// pointer is over hwnd.
func (d *WindowsDisplay) setWindowCursor(hwnd w32.HWND, c w32.HCURSOR) {
	d.windowMu.Lock()
	if info := d.windowData[hwnd]; info != nil {
		info.cursor = c
	}
	d.windowMu.Unlock()
	if hwnd == d.hoverHWND {
		w32.SetCursorFunc(d.cursorFor(hwnd))
	}
}

// cursorFor returns the cursor for hwnd: its own, else the nearest
// ancestor's, else the arrow.
func (d *WindowsDisplay) cursorFor(hwnd w32.HWND) w32.HCURSOR {
	d.windowMu.RLock()
	defer d.windowMu.RUnlock()
	for info := d.windowData[hwnd]; info != nil; info = d.windowData[info.parent] {
		if info.cursor != 0 {
			return info.cursor
		}
		if info.isTopLevel || info.parent == 0 {
			break
		}
	}
	return w32.LoadCursor(0, w32.MAKEINTRESOURCE(w32.IDC_ARROW))
}

func (d *WindowsDisplay) SetCursorShape(w platform.WindowID, shape cursor.Shape) {
	d.cursorMu.Lock()
	hcursor, ok := d.cursorCache[shape]
	d.cursorMu.Unlock()

	if !ok {
		// Look up the Windows cursor for this abstract shape.
		idcID := uint16(w32.IDC_ARROW)
		if id, found := shapeToWinCursor[shape]; found {
			idcID = id
		}
		hcursor = w32.LoadCursor(0, w32.MAKEINTRESOURCE(idcID))

		d.cursorMu.Lock()
		d.cursorCache[shape] = hcursor
		d.cursorMu.Unlock()
	}

	d.setWindowCursor(toHWND(w), hcursor)
}

func (d *WindowsDisplay) UndefineCursor(w platform.WindowID) {
	d.setWindowCursor(toHWND(w), 0)
}

func (d *WindowsDisplay) FreeCursor(cursorID platform.CursorID) {
	// Windows system cursors don't need to be freed.
}
