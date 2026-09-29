//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>

static int visual_class(Visual *v) {
#if defined(__cplusplus) || defined(c_plusplus)
	return v->c_class;
#else
	return v->class;
#endif
}

static int alloc_color(Display *dpy, Colormap cmap, unsigned short r, unsigned short g,
	unsigned short b, unsigned long *pixel) {
	XColor c;
	c.red = r; c.green = g; c.blue = b;
	c.flags = DoRed | DoGreen | DoBlue;
	if (!XAllocColor(dpy, cmap, &c)) return 0;
	*pixel = c.pixel;
	return 1;
}

// put_pixels puts a w x h block of ready visual pixels with XPutPixel, for
// visuals whose pixel layout is not 32-bit 0xRRGGBB.
static void put_pixels(Display *dpy, Drawable d, GC gc, Visual *visual, int depth,
	unsigned long *pixels, int w, int h, int dstX, int dstY) {
	XImage *img = XCreateImage(dpy, visual, depth, ZPixmap, 0, NULL, w, h, 32, 0);
	if (!img) return;
	img->data = (char *)malloc((size_t)img->bytes_per_line * h);
	if (!img->data) { XDestroyImage(img); return; }
	for (int y = 0; y < h; y++)
		for (int x = 0; x < w; x++)
			XPutPixel(img, x, y, pixels[y * w + x]);
	XPutImage(dpy, d, gc, img, 0, 0, dstX, dstY, w, h);
	XDestroyImage(img); // frees data
}
*/
import "C"

import "unsafe"

// Visual classes (X.h).
const (
	StaticGray  = 0
	GrayScale   = 1
	StaticColor = 2
	PseudoColor = 3
	TrueColor   = 4
	DirectColor = 5
)

// Class returns the visual's class (TrueColor, PseudoColor, ...).
func (v *Visual) Class() int { return int(C.visual_class(v.ptr)) }

// Masks returns the visual's red, green and blue pixel masks.
func (v *Visual) Masks() (r, g, b uint64) {
	return uint64(v.ptr.red_mask), uint64(v.ptr.green_mask), uint64(v.ptr.blue_mask)
}

// AllocColor allocates the closest read-only colour cell for 16-bit r, g, b
// in cmap (XAllocColor).
func (d *Display) AllocColor(cmap Colormap, r, g, b uint16) (uint64, bool) {
	var p C.ulong
	if C.alloc_color(d.ptr, C.Colormap(cmap), C.ushort(r), C.ushort(g), C.ushort(b), &p) == 0 {
		return 0, false
	}
	return uint64(p), true
}

// PutPixels puts w x h visual pixels (row-major) at (dstX, dstY).
func (d *Display) PutPixels(drawable Drawable, gc GC, visual *Visual, depth int, pixels []uint64, w, h, dstX, dstY int) {
	if w <= 0 || h <= 0 || len(pixels) < w*h {
		return
	}
	cp := make([]C.ulong, w*h)
	for i := range cp {
		cp[i] = C.ulong(pixels[i])
	}
	C.put_pixels(d.ptr, C.Drawable(drawable), C.GC(gc), visual.ptr, C.int(depth),
		(*C.ulong)(unsafe.Pointer(&cp[0])), C.int(w), C.int(h), C.int(dstX), C.int(dstY))
}
