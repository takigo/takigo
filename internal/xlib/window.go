//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

// WindowAttributes holds attributes for window creation.
type WindowAttributes struct {
	BackgroundPixel  uint64
	BorderPixel      uint64
	BitGravity       int
	EventMask        int64
	Colormap         Colormap
	OverrideRedirect bool
}

// CreateWindow creates a new X11 window.
func (d *Display) CreateWindow(parent Window, x, y int, width, height, borderWidth uint,
	depth int, class uint, visual *Visual, valueMask uint64, attrs *WindowAttributes) Window {

	var xattrs C.XSetWindowAttributes
	if valueMask&CWBackPixel != 0 {
		xattrs.background_pixel = C.ulong(attrs.BackgroundPixel)
	}
	if valueMask&CWBorderPixel != 0 {
		xattrs.border_pixel = C.ulong(attrs.BorderPixel)
	}
	if valueMask&CWBitGravity != 0 {
		xattrs.bit_gravity = C.int(attrs.BitGravity)
	}
	if valueMask&CWEventMask != 0 {
		xattrs.event_mask = C.long(attrs.EventMask)
	}
	if valueMask&CWColormap != 0 {
		xattrs.colormap = C.Colormap(attrs.Colormap)
	}
	if valueMask&CWOverrideRedirect != 0 {
		if attrs.OverrideRedirect {
			xattrs.override_redirect = 1
		}
	}

	var vis *C.Visual
	if visual != nil {
		vis = visual.ptr
	}

	w := C.XCreateWindow(d.ptr, C.Window(parent),
		C.int(x), C.int(y), C.uint(width), C.uint(height), C.uint(borderWidth),
		C.int(depth), C.uint(class), vis, C.ulong(valueMask), &xattrs)
	return Window(w)
}

// CreateSimpleWindow creates a simple X11 window.
func (d *Display) CreateSimpleWindow(parent Window, x, y int, width, height, borderWidth uint, border, background uint64) Window {
	w := C.XCreateSimpleWindow(d.ptr, C.Window(parent),
		C.int(x), C.int(y), C.uint(width), C.uint(height), C.uint(borderWidth),
		C.ulong(border), C.ulong(background))
	return Window(w)
}

// SetWindowBackground sets the background pixel of a window.
func (d *Display) SetWindowBackground(w Window, pixel uint64) {
	C.XSetWindowBackground(d.ptr, C.Window(w), C.ulong(pixel))
}

// DestroyWindow destroys an X11 window.
func (d *Display) DestroyWindow(w Window) {
	C.XDestroyWindow(d.ptr, C.Window(w))
}

// MapWindow maps a window to the screen.
func (d *Display) MapWindow(w Window) {
	C.XMapWindow(d.ptr, C.Window(w))
}

// MapRaised maps a window and raises it to the top.
func (d *Display) MapRaised(w Window) {
	C.XMapRaised(d.ptr, C.Window(w))
}

// UnmapWindow unmaps a window from the screen.
func (d *Display) UnmapWindow(w Window) {
	C.XUnmapWindow(d.ptr, C.Window(w))
}

// RaiseWindow raises a window to the top of the stacking order.
func (d *Display) RaiseWindow(w Window) {
	C.XRaiseWindow(d.ptr, C.Window(w))
}

// LowerWindow lowers a window to the bottom of the stacking order.
func (d *Display) LowerWindow(w Window) {
	C.XLowerWindow(d.ptr, C.Window(w))
}

// MoveWindow moves a window.
func (d *Display) MoveWindow(w Window, x, y int) {
	C.XMoveWindow(d.ptr, C.Window(w), C.int(x), C.int(y))
}

// ResizeWindow resizes a window.
func (d *Display) ResizeWindow(w Window, width, height uint) {
	C.XResizeWindow(d.ptr, C.Window(w), C.uint(width), C.uint(height))
}

// MoveResizeWindow moves and resizes a window.
func (d *Display) MoveResizeWindow(w Window, x, y int, width, height uint) {
	C.XMoveResizeWindow(d.ptr, C.Window(w), C.int(x), C.int(y), C.uint(width), C.uint(height))
}

// CreateFontCursor creates a cursor from the standard X11 cursor font.
func (d *Display) CreateFontCursor(shape uint) Cursor {
	return Cursor(C.XCreateFontCursor(d.ptr, C.uint(shape)))
}

// DefineCursor sets the cursor for a window.
func (d *Display) DefineCursor(w Window, cursor Cursor) {
	C.XDefineCursor(d.ptr, C.Window(w), C.Cursor(cursor))
}

// DefineCursorFromFont creates a cursor from the font and sets it on a window.
func (d *Display) DefineCursorFromFont(w Window, shape uint) {
	cursor := d.CreateFontCursor(shape)
	d.DefineCursor(w, cursor)
}

// UndefineCursor reverts a window to its parent's cursor.
func (d *Display) UndefineCursor(w Window) {
	C.XUndefineCursor(d.ptr, C.Window(w))
}

// FreeCursor frees a cursor.
func (d *Display) FreeCursor(cursor Cursor) {
	C.XFreeCursor(d.ptr, C.Cursor(cursor))
}

// SelectInput selects the event mask for a window.
func (d *Display) SelectInput(w Window, eventMask int64) {
	C.XSelectInput(d.ptr, C.Window(w), C.long(eventMask))
}

// StoreName sets the window name (title).
func (d *Display) StoreName(w Window, name string) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.XStoreName(d.ptr, C.Window(w), cname)
}

// SetWMProtocols sets the WM_PROTOCOLS property on a window.
func (d *Display) SetWMProtocols(w Window, protocols []Atom) int {
	if len(protocols) == 0 {
		return 0
	}
	return int(C.XSetWMProtocols(d.ptr, C.Window(w), (*C.Atom)(&protocols[0]), C.int(len(protocols))))
}

// TranslateCoordinates translates coordinates from src to dst window.
func (d *Display) TranslateCoordinates(src, dst Window, srcX, srcY int) (dstX, dstY int) {
	var dx, dy C.int
	var child C.Window
	C.XTranslateCoordinates(d.ptr, C.Window(src), C.Window(dst), C.int(srcX), C.int(srcY), &dx, &dy, &child)
	return int(dx), int(dy)
}
