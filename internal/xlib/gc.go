package xlib

/*
#include <X11/Xlib.h>
*/
import "C"

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
