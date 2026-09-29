//go:build linux || freebsd || openbsd || netbsd

package x11

import (
	"math/bits"
	"sync"

	"github.com/msorc/takigo/internal/xlib"
)

// pixelFormat converts the logical pixels the rest of takigo uses
// (0xRRGGBB, as color.Cache computes them and as the Windows and Cocoa
// backends read them) into the default visual's pixels: through its
// channel masks for TrueColor and DirectColor at any depth (5-6-5,
// 10-10-10, ...), and through allocated colour cells, as Tk_GetColor does,
// for colormapped visuals. A 24/32-bit 0xRRGGBB visual needs no change.
type pixelFormat struct {
	identity bool
	masked   bool
	shift    [3]int
	max      [3]uint64

	dpy  *xlib.Display
	cmap xlib.Colormap
	mu   sync.Mutex
	// cells caches allocated colour cells by logical pixel.
	cells map[uint64]uint64
}

func newPixelFormat(dpy *xlib.Display, screen int) *pixelFormat {
	v := dpy.DefaultVisual(screen)
	pf := &pixelFormat{dpy: dpy, cmap: dpy.DefaultColormap(screen)}
	switch v.Class() {
	case xlib.TrueColor, xlib.DirectColor:
		r, g, b := v.Masks()
		if r == 0xff0000 && g == 0xff00 && b == 0xff {
			pf.identity = true
			return pf
		}
		pf.masked = true
		for i, m := range [3]uint64{r, g, b} {
			pf.shift[i] = bits.TrailingZeros64(m)
			pf.max[i] = m >> pf.shift[i]
		}
	default:
		pf.cells = map[uint64]uint64{}
	}
	return pf
}

// pixel returns the visual's pixel for logical pixel p (0xRRGGBB).
func (pf *pixelFormat) pixel(p uint64) uint64 {
	if pf.identity {
		return p
	}
	c := [3]uint64{p >> 16 & 0xff, p >> 8 & 0xff, p & 0xff}
	if pf.masked {
		var out uint64
		for i, v := range c {
			out |= (v*pf.max[i] + 127) / 255 << pf.shift[i]
		}
		return out
	}
	pf.mu.Lock()
	defer pf.mu.Unlock()
	if px, ok := pf.cells[p]; ok {
		return px
	}
	px, ok := pf.dpy.AllocColor(pf.cmap, uint16(c[0]*257), uint16(c[1]*257), uint16(c[2]*257))
	if !ok {
		// The colormap is full: the nearer of black and white.
		px = pf.dpy.BlackPixel(pf.dpy.DefaultScreen())
		if c[0]*30+c[1]*59+c[2]*11 >= 128*100 {
			px = pf.dpy.WhitePixel(pf.dpy.DefaultScreen())
		}
	}
	pf.cells[p] = px
	return px
}
