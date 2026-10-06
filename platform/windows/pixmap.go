//go:build windows

package windows

import (
	mathbits "math/bits"
	"unsafe"

	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// pixmapInfo holds state for an offscreen bitmap.
type pixmapInfo struct {
	hdc     w32.HDC
	hbitmap w32.HBITMAP
	oldBmp  w32.HGDIOBJ // previous bitmap, must be restored before deletion
	width   int
	height  int
	depth   int
	pattern w32.HBRUSH // depth-1 bitmaps: the bitmap as a brush, for stipples
}

// --- PixmapManager implementation ---

func (d *WindowsDisplay) CreatePixmap(drawable platform.DrawableID, width, height, depth uint) platform.PixmapID {
	// Get a compatible DC for the drawable.
	var srcDC w32.HDC
	if pix := d.getPixmap(platform.PixmapID(drawable)); pix != nil {
		srcDC = pix.hdc
	} else {
		hwnd := toDrawableHWND(drawable)
		srcDC = w32.GetDC(hwnd)
		defer w32.ReleaseDC(hwnd, srcDC)
	}

	memDC := w32.CreateCompatibleDC(srcDC)
	hbitmap := w32.CreateCompatibleBitmap(srcDC, int32(width), int32(height))
	oldBmp := w32.SelectObject(memDC, w32.HGDIOBJ(hbitmap))

	// Clear the pixmap to white.
	rect := w32.RECT{Left: 0, Top: 0, Right: int32(width), Bottom: int32(height)}
	brush := w32.CreateSolidBrush(w32.RGB(255, 255, 255))
	w32.FillRect(memDC, &rect, brush)
	w32.DeleteObject(w32.HGDIOBJ(brush))

	d.pixmapMu.Lock()
	id := platform.PixmapID(d.pixmapNext)
	d.pixmapNext++
	d.pixmaps[id] = &pixmapInfo{
		hdc:     memDC,
		hbitmap: hbitmap,
		oldBmp:  oldBmp,
		width:   int(width),
		height:  int(height),
		depth:   int(depth),
	}
	d.pixmapMu.Unlock()

	return id
}

func (d *WindowsDisplay) FreePixmap(pixmap platform.PixmapID) {
	d.pixmapMu.Lock()
	pix, ok := d.pixmaps[pixmap]
	if ok {
		delete(d.pixmaps, pixmap)
	}
	d.pixmapMu.Unlock()

	if ok && pix != nil {
		if pix.pattern != 0 {
			w32.DeleteObject(w32.HGDIOBJ(pix.pattern))
		}
		w32.SelectObject(pix.hdc, pix.oldBmp)
		w32.DeleteObject(w32.HGDIOBJ(pix.hbitmap))
		w32.DeleteDC(pix.hdc)
	}
}

func (d *WindowsDisplay) CreateBitmapFromData(drawable platform.DrawableID, bits []byte, width, height uint) platform.PixmapID {
	// Create a 1-bit depth pixmap from XBM-format data.
	var srcDC w32.HDC
	hwnd := toDrawableHWND(drawable)
	srcDC = w32.GetDC(hwnd)
	defer w32.ReleaseDC(hwnd, srcDC)

	memDC := w32.CreateCompatibleDC(srcDC)

	// Create a monochrome bitmap whose set bits are white: drawn through
	// a brush, white lets the source through (see render).
	bmi := struct {
		header w32.BITMAPINFOHEADER
		colors [2]uint32
	}{
		header: w32.BITMAPINFOHEADER{
			BiSize:     uint32(unsafe.Sizeof(w32.BITMAPINFOHEADER{})),
			BiWidth:    int32(width),
			BiHeight:   -int32(height), // top-down
			BiPlanes:   1,
			BiBitCount: 1,
		},
		colors: [2]uint32{0x00000000, 0x00FFFFFF},
	}
	var bitsPtr unsafe.Pointer
	hbitmap := w32.CreateDIBSection(memDC, (*w32.BITMAPINFO)(unsafe.Pointer(&bmi)), w32.DIB_RGB_COLORS, &bitsPtr, 0, 0)

	if hbitmap == 0 {
		// Fallback: create a compatible monochrome bitmap.
		hbitmap = w32.CreateCompatibleBitmap(srcDC, int32(width), int32(height))
	} else if bitsPtr != nil && len(bits) > 0 {
		// XBM rows are byte-aligned with the leftmost pixel in the low
		// bit; DIB rows are 32-bit aligned with it in the high bit.
		xbmStride := (int(width) + 7) / 8
		dibStride := ((int(width) + 31) / 32) * 4
		dst := (*[1 << 30]byte)(bitsPtr)
		for row := range int(height) {
			srcOff := row * xbmStride
			dstOff := row * dibStride
			for col := 0; col < xbmStride && srcOff+col < len(bits); col++ {
				dst[dstOff+col] = mathbits.Reverse8(bits[srcOff+col])
			}
		}
	}

	// The brush copies the bitmap; make it while no DC holds the bitmap.
	pattern := w32.CreatePatternBrush(hbitmap)

	oldBmp := w32.SelectObject(memDC, w32.HGDIOBJ(hbitmap))

	d.pixmapMu.Lock()
	id := platform.PixmapID(d.pixmapNext)
	d.pixmapNext++
	d.pixmaps[id] = &pixmapInfo{
		hdc:     memDC,
		hbitmap: hbitmap,
		oldBmp:  oldBmp,
		width:   int(width),
		height:  int(height),
		depth:   1,
		pattern: pattern,
	}
	d.pixmapMu.Unlock()

	return id
}

// getPixmap returns the pixmap info for the given ID, or nil.
func (d *WindowsDisplay) getPixmap(id platform.PixmapID) *pixmapInfo {
	d.pixmapMu.Lock()
	defer d.pixmapMu.Unlock()
	return d.pixmaps[id]
}

// getDrawableDC returns an HDC for the given drawable and a cleanup function.
// For window drawables, the DC must be released via the cleanup.
// For pixmap drawables, the cleanup is a no-op.
func (d *WindowsDisplay) getDrawableDC(drawable platform.DrawableID) (w32.HDC, func()) {
	// Check if it's a pixmap first.
	if pix := d.getPixmap(platform.PixmapID(drawable)); pix != nil {
		return pix.hdc, func() {}
	}
	// It's a window.
	hwnd := toDrawableHWND(drawable)
	hdc := w32.GetDC(hwnd)
	return hdc, func() { w32.ReleaseDC(hwnd, hdc) }
}
