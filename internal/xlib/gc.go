//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <X11/Xlib.h>
#include <X11/Xutil.h>
*/
import "C"
import "unsafe"

// GCValues holds GC configuration values.
type GCValues struct {
	Foreground uint64
	Background uint64
	LineWidth  int
	Function   int
}

// CreateGC creates a new graphics context.
func (d *Display) CreateGC(drawable Drawable, valueMask uint64, values *GCValues) GC {
	var xgcv C.XGCValues
	if valueMask&GCForeground != 0 {
		xgcv.foreground = C.ulong(values.Foreground)
	}
	if valueMask&GCBackground != 0 {
		xgcv.background = C.ulong(values.Background)
	}
	if valueMask&GCLineWidth != 0 {
		xgcv.line_width = C.int(values.LineWidth)
	}
	if valueMask&GCFunction != 0 {
		xgcv.function = C.int(values.Function)
	}
	return GC(C.XCreateGC(d.ptr, C.Drawable(drawable), C.ulong(valueMask), &xgcv))
}

// FreeGC frees a graphics context.
func (d *Display) FreeGC(gc GC) {
	C.XFreeGC(d.ptr, C.GC(gc))
}

// SetForeground sets the foreground color of a GC.
func (d *Display) SetForeground(gc GC, pixel uint64) {
	C.XSetForeground(d.ptr, C.GC(gc), C.ulong(pixel))
}

// SetBackground sets the background color of a GC.
func (d *Display) SetBackground(gc GC, pixel uint64) {
	C.XSetBackground(d.ptr, C.GC(gc), C.ulong(pixel))
}

// SetLineAttributes sets line drawing attributes of a GC.
func (d *Display) SetLineAttributes(gc GC, lineWidth uint, lineStyle, capStyle, joinStyle int) {
	C.XSetLineAttributes(d.ptr, C.GC(gc), C.uint(lineWidth),
		C.int(lineStyle), C.int(capStyle), C.int(joinStyle))
}

// SetFillStyle sets the fill style of a GC (FillSolid, FillStippled, etc.).
func (d *Display) SetFillStyle(gc GC, fillStyle int) {
	C.XSetFillStyle(d.ptr, C.GC(gc), C.int(fillStyle))
}

// SetStipple sets the stipple pixmap (depth-1 bitmap) for a GC.
func (d *Display) SetStipple(gc GC, stipple Pixmap) {
	C.XSetStipple(d.ptr, C.GC(gc), C.Pixmap(stipple))
}

// SetTSOrigin sets the tile/stipple origin of a GC.
func (d *Display) SetTSOrigin(gc GC, x, y int) {
	C.XSetTSOrigin(d.ptr, C.GC(gc), C.int(x), C.int(y))
}

// CreateBitmapFromData creates a depth-1 pixmap from XBM-format bit data.
func (d *Display) CreateBitmapFromData(drawable Drawable, bits []byte, width, height uint) Pixmap {
	return Pixmap(C.XCreateBitmapFromData(d.ptr, C.Drawable(drawable),
		(*C.char)(unsafe.Pointer(&bits[0])), C.uint(width), C.uint(height)))
}
