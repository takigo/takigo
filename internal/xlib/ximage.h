/* XImage layout helpers shared by the cgo files of this package. */
#ifndef TAKIGO_XIMAGE_H
#define TAKIGO_XIMAGE_H

#include <X11/Xlib.h>

/* takigo_bits_per_pixel returns the width of a pixel of a drawable of the
 * given depth in an XImage: 8 bits for the depths up to 8, 16 for a 5-6-5
 * visual, 32 for the 24/32-bit ones. Xlib derives the same value from the
 * depth and ignores the bitmap_pad given to XCreateImage.
 */
static int takigo_bits_per_pixel(int depth)
{
	if (depth <= 8) return 8;
	if (depth <= 16) return 16;
	return 32;
}

/* takigo_scanline_bytes returns the number of bytes a scanline of w pixels
 * of bpp bits takes; the X server pads every scanline to a multiple of 4
 * bytes.
 */
static int takigo_scanline_bytes(int bpp, int w)
{
	return ((bpp * w + 31) >> 3) & ~3;
}

/* takigo_image_layout returns the width of a pixel of a drawable of the
 * given depth in an XImage and whether the image's pixels are in
 * MSBFirst order. Xlib derives the width from the depth as well, and the
 * image must have the width and the byte order Xlib expects, so ask it.
 */
static void takigo_image_layout(Display *dpy, Visual *visual, int depth, int w,
	int *bppOut, int *msbFirstOut)
{
	int bpp = takigo_bits_per_pixel(depth);
	int msbFirst = (ImageByteOrder(dpy) == MSBFirst);
	XImage *probe = XCreateImage(dpy, visual, (unsigned)depth, ZPixmap, 0,
		NULL, w, 1, 32, 0);
	if (!probe) {
		*bppOut = bpp;
		*msbFirstOut = msbFirst;
		return;
	}
	if (probe->bits_per_pixel) {
		bpp = probe->bits_per_pixel;
	}
	msbFirst = (probe->byte_order == MSBFirst);
	XDestroyImage(probe);
	*bppOut = bpp;
	*msbFirstOut = msbFirst;
}

/* takigo_store16 and takigo_store32 write a pixel value in the byte order
 * of the image (MSBFirst or LSBFirst).
 */
static void takigo_store16(unsigned char *p, unsigned long v, int msbFirst)
{
	if (msbFirst) {
		p[0] = (unsigned char)(v >> 8);
		p[1] = (unsigned char)v;
	} else {
		p[0] = (unsigned char)v;
		p[1] = (unsigned char)(v >> 8);
	}
}

static void takigo_store32(unsigned char *p, unsigned long v, int msbFirst)
{
	if (msbFirst) {
		p[0] = (unsigned char)(v >> 24);
		p[1] = (unsigned char)(v >> 16);
		p[2] = (unsigned char)(v >> 8);
		p[3] = (unsigned char)v;
	} else {
		p[0] = (unsigned char)v;
		p[1] = (unsigned char)(v >> 8);
		p[2] = (unsigned char)(v >> 16);
		p[3] = (unsigned char)(v >> 24);
	}
}

#endif
