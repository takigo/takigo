//go:build linux || freebsd || openbsd || netbsd

package xlib

import (
	"math"
	"math/bits"
	"testing"

	"github.com/takigo/takigo/internal/displaylock"
)

// TestPutImageRGBARoundTrip puts opaque colours on a pixmap and reads them
// back: the pixel layout of an image follows the depth of the drawable, and
// a 32-bit image on a 16-bit drawable is rejected by the server.
func TestPutImageRGBARoundTrip(t *testing.T) {
	displaylock.Require(t)

	dpy, err := OpenDisplay("")
	if err != nil {
		t.Fatalf("OpenDisplay failed: %v", err)
	}
	defer dpy.Close()

	screen := dpy.DefaultScreen()
	visual := dpy.DefaultVisual(screen)
	depth := dpy.DefaultDepth(screen)
	rm, gm, bm := visual.Masks()
	if rm == 0 || gm == 0 || bm == 0 {
		t.Skip("colormapped visual: the pixels are colour cells, not masks")
	}

	const w, h = 4, 1
	colors := [w][3]uint8{
		{255, 0, 0},
		{0, 255, 0},
		{0, 0, 255},
		{255, 255, 255},
	}
	rgba := make([]byte, 0, w*h*4)
	for _, c := range colors {
		rgba = append(rgba, c[0], c[1], c[2], 255)
	}

	pm := dpy.CreatePixmap(Drawable(dpy.DefaultRootWindow()), w, h, uint(depth))
	defer dpy.FreePixmap(pm)
	gc := dpy.CreateGC(Drawable(pm), 0, nil)
	defer dpy.FreeGC(gc)

	dpy.PutImageRGBA(Drawable(pm), gc, visual, depth, rgba, w*4, w, h, 0, 0, 0, 0, w, h, 0)
	got := dpy.GetImageRGBA(Drawable(pm), visual, 0, 0, w, h)
	if got == nil {
		t.Fatal("GetImageRGBA returned nil")
	}
	// The channels of the visual may be narrower than 8 bits (5-6-5).
	tolerance := func(mask uint64) float64 {
		bits := bits.Len64(mask >> bits.TrailingZeros64(mask))
		return math.Ldexp(255, -bits) + 1
	}
	for x, c := range colors {
		i := x * 4
		for ch, want := range []uint8{c[0], c[1], c[2]} {
			mask := [3]uint64{rm, gm, bm}[ch]
			if diff := math.Abs(float64(got[i+ch]) - float64(want)); diff > tolerance(mask) {
				t.Errorf("pixel %d channel %d = %d, want %d (tolerance %g)",
					x, ch, got[i+ch], want, tolerance(mask))
			}
		}
	}
}

// TestCreatePixmapRefusesBadSize: a size outside the protocol's range must
// not reach the server, where it is a BadValue followed by a BadDrawable
// for every request on the pixmap.
func TestCreatePixmapRefusesBadSize(t *testing.T) {
	displaylock.Require(t)

	dpy, err := OpenDisplay("")
	if err != nil {
		t.Fatalf("OpenDisplay failed: %v", err)
	}
	defer dpy.Close()

	root := Drawable(dpy.DefaultRootWindow())
	depth := uint(dpy.DefaultDepth(dpy.DefaultScreen()))
	before := ErrorCount()
	for _, size := range [][2]uint{{0, 10}, {10, 0}, {maxPixmapSize + 1, 10}, {10, 1 << 16}, {1<<31 + 539, 242}} {
		if pm := dpy.CreatePixmap(root, size[0], size[1], depth); pm != 0 {
			t.Errorf("CreatePixmap(%d, %d) = %#x, want 0", size[0], size[1], pm)
		}
	}
	pm := dpy.CreatePixmap(root, 10, 10, depth)
	if pm == 0 {
		t.Error("CreatePixmap(10, 10) = 0")
	} else {
		dpy.FreePixmap(pm)
	}
	dpy.Sync(false)
	if n := ErrorCount() - before; n != 0 {
		t.Errorf("%d X errors, want 0", n)
	}
}
