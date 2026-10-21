//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>

#include "xerror.h"
#include "ximage.h"

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
	// The pixel width follows the depth: 8 bits for the depths up to 8,
	// 16 for a 5-6-5 visual, 32 for 24/32-bit ones. A 32-bit image
	// cannot be put on a 16-bit drawable, the server rejects it.
	int bpp, msbFirst;
	takigo_image_layout(dpy, visual, depth, w, &bpp, &msbFirst);
	// The server pads every scanline to a multiple of 4 bytes.
	int bpl = takigo_scanline_bytes(bpp, w);
	// The buffer is ours: XCreateImage allocates one when data is NULL,
	// and replacing img->data with another one leaks that first.
	char *buf = (char *)malloc((size_t)bpl * (size_t)h);
	if (!buf) return;
	XImage *img = XCreateImage(dpy, visual, depth, ZPixmap, 0, buf, w, h, bpp, bpl);
	if (!img) { free(buf); return; }
	if (img->bits_per_pixel != bpp || img->bytes_per_line != bpl) {
		// Xlib chose another layout; its bytes_per_line does not match
		// the buffer we allocated, so do not put anything.
		img->data = NULL;
		XDestroyImage(img);
		free(buf);
		return;
	}
	for (int y = 0; y < h; y++)
		for (int x = 0; x < w; x++)
			XPutPixel(img, x, y, pixels[y * w + x]);
	takigo_trap_errors(dpy);
	XPutImage(dpy, d, gc, img, 0, 0, dstX, dstY, w, h);
	takigo_untrap_errors(dpy);
	img->data = NULL; // our buffer, freed below
	XDestroyImage(img);
	free(buf);
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
	trapMu.Lock()
	defer trapMu.Unlock()
	C.put_pixels(d.ptr, C.Drawable(drawable), C.GC(gc), visual.ptr, C.int(depth),
		(*C.ulong)(unsafe.Pointer(&cp[0])), C.int(w), C.int(h), C.int(dstX), C.int(dstY))
}
