//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <stdlib.h>
#include <string.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>

// put_rgba_image converts Go RGBA data to X11 BGRA format, pre-composites
// alpha against bgPixel, and puts the image onto a drawable.
static void put_rgba_image(Display *dpy, Drawable d, GC gc, Visual *visual,
	int depth, unsigned char *rgba, int stride,
	int imgW, int imgH,
	int srcX, int srcY,
	int dstX, int dstY,
	int w, int h,
	unsigned long bgPixel)
{
	// Extract bg RGB components from pixel value.
	unsigned char bgR = (bgPixel >> 16) & 0xFF;
	unsigned char bgG = (bgPixel >> 8) & 0xFF;
	unsigned char bgB = bgPixel & 0xFF;

	// Allocate 32-bit BGRA buffer for XImage.
	int rowBytes = w * 4;
	unsigned char *buf = (unsigned char *)malloc(rowBytes * h);
	if (!buf) return;

	for (int y = 0; y < h; y++) {
		unsigned char *src = rgba + (srcY + y) * stride + srcX * 4;
		unsigned char *dst = buf + y * rowBytes;
		for (int x = 0; x < w; x++) {
			unsigned char r = src[0];
			unsigned char g = src[1];
			unsigned char b = src[2];
			unsigned char a = src[3];
			if (a == 255) {
				dst[0] = b;
				dst[1] = g;
				dst[2] = r;
				dst[3] = 0;
			} else if (a == 0) {
				dst[0] = bgB;
				dst[1] = bgG;
				dst[2] = bgR;
				dst[3] = 0;
			} else {
				// Go's image.RGBA is premultiplied: out = src + bg*(1-a).
				int ob = b + bgB * (255 - a) / 255;
				int og = g + bgG * (255 - a) / 255;
				int or_ = r + bgR * (255 - a) / 255;
				dst[0] = ob > 255 ? 255 : ob;
				dst[1] = og > 255 ? 255 : og;
				dst[2] = or_ > 255 ? 255 : or_;
				dst[3] = 0;
			}
			src += 4;
			dst += 4;
		}
	}

	XImage *img = XCreateImage(dpy, visual, depth, ZPixmap, 0,
		(char *)buf, w, h, 32, rowBytes);
	if (img) {
		XPutImage(dpy, d, gc, img, 0, 0, dstX, dstY, w, h);
		img->data = NULL; // prevent XDestroyImage from freeing our buffer
		XDestroyImage(img);
	}
	free(buf);
}
*/
import "C"
import "unsafe"

// CreatePixmap creates a pixmap of the given dimensions and depth.
func (d *Display) CreatePixmap(drawable Drawable, width, height, depth uint) Pixmap {
	return Pixmap(C.XCreatePixmap(d.ptr, C.Drawable(drawable),
		C.uint(width), C.uint(height), C.uint(depth)))
}

// FreePixmap frees a previously created pixmap.
func (d *Display) FreePixmap(pixmap Pixmap) {
	C.XFreePixmap(d.ptr, C.Pixmap(pixmap))
}

// CopyArea copies a rectangular area from one drawable to another.
func (d *Display) CopyArea(src, dst Drawable, gc GC, srcX, srcY int, width, height uint, dstX, dstY int) {
	C.XCopyArea(d.ptr, C.Drawable(src), C.Drawable(dst), C.GC(gc),
		C.int(srcX), C.int(srcY), C.uint(width), C.uint(height),
		C.int(dstX), C.int(dstY))
}

// PutImageRGBA converts Go RGBA pixel data to X11 format, pre-composites
// alpha against bgPixel, and puts the image onto a drawable.
func (d *Display) PutImageRGBA(drawable Drawable, gc GC, visual *Visual, depth int,
	rgbaData []byte, stride int, imgW, imgH int,
	srcX, srcY, dstX, dstY, w, h int, bgPixel uint64) {

	if len(rgbaData) == 0 || w <= 0 || h <= 0 {
		return
	}
	C.put_rgba_image(d.ptr, C.Drawable(drawable), C.GC(gc),
		visual.ptr, C.int(depth),
		(*C.uchar)(unsafe.Pointer(&rgbaData[0])), C.int(stride),
		C.int(imgW), C.int(imgH),
		C.int(srcX), C.int(srcY),
		C.int(dstX), C.int(dstY),
		C.int(w), C.int(h),
		C.ulong(bgPixel))
}

// PixmapDrawable converts a Pixmap to a Drawable for use in drawing functions.
func PixmapDrawable(p Pixmap) Drawable {
	return Drawable(p)
}
