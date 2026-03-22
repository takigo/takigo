//go:build windows

package windows

import (
	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/platform"
)

// shapeToWinCursor maps abstract cursor shapes to Windows IDC_* constants.
var shapeToWinCursor = map[cursor.Shape]uint16{
	cursor.Arrow:          w32.IDC_ARROW,
	cursor.Crosshair:      w32.IDC_CROSS,
	cursor.Fleur:          w32.IDC_SIZEALL,
	cursor.Hand1:          w32.IDC_HAND,
	cursor.Hand2:          w32.IDC_HAND,
	cursor.LeftPtr:        w32.IDC_ARROW,
	cursor.Plus:           w32.IDC_CROSS,
	cursor.QuestionArrow:  w32.IDC_HELP,
	cursor.SBHDoubleArrow: w32.IDC_SIZEWE,
	cursor.SBVDoubleArrow: w32.IDC_SIZENS,
	cursor.SizingAngle:    w32.IDC_SIZENWSE,
	cursor.TopLeftArrow:   w32.IDC_ARROW,
	cursor.Watch:          w32.IDC_WAIT,
	cursor.XTerm:          w32.IDC_IBEAM,
	cursor.BottomRightCorner: w32.IDC_SIZENWSE,
}

// --- CursorManager implementation ---

func (d *WindowsDisplay) CreateFontCursor(shape uint) platform.CursorID {
	// Map X11 cursor font index to Windows cursor.
	// For now, treat the shape as a generic IDC value.
	hcursor := w32.LoadCursor(0, w32.MAKEINTRESOURCE(w32.IDC_ARROW))
	return platform.CursorID(uintptr(hcursor))
}

func (d *WindowsDisplay) DefineCursor(w platform.WindowID, cursorID platform.CursorID) {
	hcursor := w32.HCURSOR(uintptr(cursorID))
	w32.SetClassLongPtr(toHWND(w), w32.GCLP_HCURSOR, uintptr(hcursor))
}

func (d *WindowsDisplay) SetCursorShape(w platform.WindowID, shape uint) {
	d.cursorMu.Lock()
	hcursor, ok := d.cursorCache[shape]
	d.cursorMu.Unlock()

	if !ok {
		// Look up the Windows cursor for this abstract shape.
		idcID := uint16(w32.IDC_ARROW)
		if id, found := shapeToWinCursor[cursor.Shape(shape)]; found {
			idcID = id
		}
		hcursor = w32.LoadCursor(0, w32.MAKEINTRESOURCE(idcID))

		d.cursorMu.Lock()
		d.cursorCache[shape] = hcursor
		d.cursorMu.Unlock()
	}

	w32.SetClassLongPtr(toHWND(w), w32.GCLP_HCURSOR, uintptr(hcursor))
}

func (d *WindowsDisplay) UndefineCursor(w platform.WindowID) {
	// Reset to default arrow cursor.
	hcursor := w32.LoadCursor(0, w32.MAKEINTRESOURCE(w32.IDC_ARROW))
	w32.SetClassLongPtr(toHWND(w), w32.GCLP_HCURSOR, uintptr(hcursor))
}

func (d *WindowsDisplay) FreeCursor(cursorID platform.CursorID) {
	// Windows system cursors don't need to be freed.
}
