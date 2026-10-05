// Package geometry defines shared infrastructure for geometry management.
// It ports tk/generic/tkGeometry.c.
package geometry

import (
	"github.com/takigo/takigo/window"
)

// Manager is an alias for the interface defined in the window package.
// Geometry managers (pack, grid, place) implement window.GeomManager.
type Manager = window.GeomManager

// ManageGeometry associates a window with a geometry manager.
// If the window already has a different manager, the old manager's
// LostContentProc is called.
func ManageGeometry(w *window.Window, mgr Manager) {
	if w.GeomManager != nil && w.GeomManager != mgr {
		w.GeomManager.LostContentProc(w)
	}
	w.GeomManager = mgr
}

// GeometryRequest is called by a widget to inform its geometry manager
// of its preferred size. The manager's RequestProc is then called.
func GeometryRequest(w *window.Window, reqWidth, reqHeight int) {
	if reqWidth < 1 {
		reqWidth = 1
	}
	if reqHeight < 1 {
		reqHeight = 1
	}
	// Enforce minimum requested size (Tk_SetMinimumRequestSize).
	if w.MinReqWidth > 0 && reqWidth < w.MinReqWidth {
		reqWidth = w.MinReqWidth
	}
	if w.MinReqHeight > 0 && reqHeight < w.MinReqHeight {
		reqHeight = w.MinReqHeight
	}

	if w.ReqWidth == reqWidth && w.ReqHeight == reqHeight {
		return
	}

	w.ReqWidth = reqWidth
	w.ReqHeight = reqHeight

	if w.GeomManager != nil {
		w.GeomManager.RequestProc(w)
	}
}

// SetInternalBorder sets the internal border widths where child windows
// cannot be placed.
func SetInternalBorder(w *window.Window, left, right, top, bottom int) {
	w.InternalBorderLeft = left
	w.InternalBorderRight = right
	w.InternalBorderTop = top
	w.InternalBorderBottom = bottom
}

// SetInternalBorderUniform sets all internal borders to the same width.
func SetInternalBorderUniform(w *window.Window, width int) {
	SetInternalBorder(w, width, width, width, width)
}

// UsableWidth returns the width available for placing children.
func UsableWidth(w *window.Window) int {
	return w.Width - w.InternalBorderLeft - w.InternalBorderRight
}

// UsableHeight returns the height available for placing children.
func UsableHeight(w *window.Window) int {
	return w.Height - w.InternalBorderTop - w.InternalBorderBottom
}
