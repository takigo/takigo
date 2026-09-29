//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>

// set_dashes wraps XSetDashes.
static void set_dashes(Display *dpy, GC gc, int dash_offset, char *dash_list, int n) {
	XSetDashes(dpy, gc, dash_offset, dash_list, n);
}
*/
import "C"
import "unsafe"

// XPoint represents an X11 point with int16 coordinates.
type XPoint struct {
	X, Y int16
}

// The points are handed to Xlib in place, so XPoint must match C's XPoint.
var _ [unsafe.Sizeof(XPoint{}) - unsafe.Sizeof(C.XPoint{})]struct{}
var _ [unsafe.Sizeof(C.XPoint{}) - unsafe.Sizeof(XPoint{})]struct{}

// cPoints passes points to Xlib without copying.
func cPoints(points []XPoint) *C.XPoint {
	return (*C.XPoint)(unsafe.Pointer(&points[0]))
}

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

// DrawLines draws connected line segments on a drawable.
func (d *Display) DrawLines(drawable Drawable, gc GC, points []XPoint, mode int) {
	if len(points) < 2 {
		return
	}
	C.XDrawLines(d.ptr, C.Drawable(drawable), C.GC(gc),
		cPoints(points), C.int(len(points)), C.int(mode))
}

// FillPolygon fills a polygon defined by points on a drawable.
func (d *Display) FillPolygon(drawable Drawable, gc GC, points []XPoint, shape, mode int) {
	if len(points) < 3 {
		return
	}
	C.XFillPolygon(d.ptr, C.Drawable(drawable), C.GC(gc),
		cPoints(points), C.int(len(points)), C.int(shape), C.int(mode))
}

// SetDashes sets the dash pattern for a GC.
func (d *Display) SetDashes(gc GC, dashOffset int, dashList []byte) {
	if len(dashList) == 0 {
		return
	}
	C.set_dashes(d.ptr, C.GC(gc), C.int(dashOffset),
		(*C.char)(unsafe.Pointer(&dashList[0])), C.int(len(dashList)))
}
