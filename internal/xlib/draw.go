package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
*/
import "C"
import "unsafe"

// FillRectangle fills a rectangle on a drawable.
func (d *Display) FillRectangle(drawable Drawable, gc GC, x, y int, width, height uint) {
	C.XFillRectangle(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x), C.int(y), C.uint(width), C.uint(height))
}

// DrawRectangle draws a rectangle outline on a drawable.
func (d *Display) DrawRectangle(drawable Drawable, gc GC, x, y int, width, height uint) {
	C.XDrawRectangle(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x), C.int(y), C.uint(width), C.uint(height))
}

// DrawLine draws a line on a drawable.
func (d *Display) DrawLine(drawable Drawable, gc GC, x1, y1, x2, y2 int) {
	C.XDrawLine(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x1), C.int(y1), C.int(x2), C.int(y2))
}

// DrawString draws a string on a drawable using the GC's current font.
func (d *Display) DrawString(drawable Drawable, gc GC, x, y int, str string) {
	cstr := C.CString(str)
	defer C.free(unsafe.Pointer(cstr))
	C.XDrawString(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x), C.int(y), cstr, C.int(len(str)))
}

// ClearWindow clears the entire window using the background pixel.
func (d *Display) ClearWindow(w Window) {
	C.XClearWindow(d.ptr, C.Window(w))
}

// ClearArea clears a rectangular area in a window.
func (d *Display) ClearArea(w Window, x, y int, width, height uint, exposures bool) {
	var exp C.int
	if exposures {
		exp = 1
	}
	C.XClearArea(d.ptr, C.Window(w), C.int(x), C.int(y), C.uint(width), C.uint(height), exp)
}

// FillArc fills an arc on a drawable.
func (d *Display) FillArc(drawable Drawable, gc GC, x, y int, width, height uint, angle1, angle2 int) {
	C.XFillArc(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x), C.int(y), C.uint(width), C.uint(height),
		C.int(angle1), C.int(angle2))
}

// DrawArc draws an arc outline on a drawable.
func (d *Display) DrawArc(drawable Drawable, gc GC, x, y int, width, height uint, angle1, angle2 int) {
	C.XDrawArc(d.ptr, C.Drawable(drawable), C.GC(gc),
		C.int(x), C.int(y), C.uint(width), C.uint(height),
		C.int(angle1), C.int(angle2))
}
