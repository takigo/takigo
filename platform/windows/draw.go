//go:build windows

package windows

import (
	"math"
	"unsafe"

	w32 "github.com/takigo/takigo/internal/win32"
	"github.com/takigo/takigo/platform"
)

// --- Drawer implementation ---

// render runs draw on the drawable's DC. With a stippled fill style it
// does what RenderObject in tk/win/tkWinDraw.c does: draw works on a copy
// of the box left, top .. right, bottom, and only the pixels where the
// stipple is set are copied back.
//
// The copy-back departs from Tk. Tk selects the stipple into the
// destination as a pattern brush, sets the brush origin to the GC's
// ts origin and does one BitBlt with the ternary raster op COPYFG
// (0x00CA0749, dest = (pat & src) | (!pat & dst)). Done that way, Wine
// 11.17 put the pattern at the brush origin plus the destination
// rectangle's position, so anything with a non-zero ts origin (canvas
// bitmap items, stipples in a scrolled canvas) came out wrapped. Here the
// brush paints a mask with a plain PatBlt and the mask is combined with
// SRCINVERT and SRCAND only, which costs a second scratch bitmap and two
// more blits.
//
// Not established: whether that is a Wine bug (Tk's stipple is a
// device-dependent monochrome bitmap, ours a 1-bit DIB section, and the
// COPYFG path was not tried with the former), and how either method
// behaves on real Windows, where only Tk's is proven. If COPYFG turns out
// to be right there, it can replace the mask.
func (d *WindowsDisplay) render(drawable platform.DrawableID, g *gcState, left, top, right, bottom int32, draw func(dc w32.HDC)) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	var pattern w32.HBRUSH
	if g.fillStyle == platform.FillStippled {
		if pix := d.getPixmap(g.stipple); pix != nil {
			pattern = pix.pattern
		}
	}
	if pattern == 0 {
		draw(hdc)
		return
	}

	dw, dh := d.drawableSize(drawable)
	left, top = max(left, 0), max(top, 0)
	right, bottom = min(right, dw), min(bottom, dh)
	w, h := right-left, bottom-top
	if w <= 0 || h <= 0 {
		return
	}

	memDC, freeMem := scratchDC(hdc, w, h)
	defer freeMem()
	maskDC, freeMask := scratchDC(hdc, w, h)
	defer freeMask()

	w32.BitBlt(memDC, 0, 0, w, h, hdc, left, top, w32.SRCCOPY)
	w32.SetViewportOrgEx(memDC, -left, -top, nil)
	draw(memDC)
	w32.SetViewportOrgEx(memDC, 0, 0, nil)

	// The mask is white where the stipple is set: a monochrome brush
	// paints its 0 bits in the text colour and its 1 bits in the
	// background colour.
	w32.SetTextColor(maskDC, w32.RGB(0, 0, 0))
	w32.SetBkColor(maskDC, w32.RGB(255, 255, 255))
	w32.SetBrushOrgEx(maskDC, int32(g.tsX)-left, int32(g.tsY)-top, nil)
	oldBrush := w32.SelectObject(maskDC, w32.HGDIOBJ(pattern))
	w32.PatBlt(maskDC, 0, 0, w, h, w32.PATCOPY)
	w32.SelectObject(maskDC, oldBrush)

	// mem = (mem ^ dst) & mask, then dst ^= mem: dst takes the drawn
	// pixels where the mask is white and keeps its own elsewhere.
	w32.BitBlt(memDC, 0, 0, w, h, hdc, left, top, w32.SRCINVERT)
	w32.BitBlt(memDC, 0, 0, w, h, maskDC, 0, 0, w32.SRCAND)
	w32.BitBlt(hdc, left, top, w, h, memDC, 0, 0, w32.SRCINVERT)
}

// scratchDC returns a memory DC holding a w x h bitmap compatible with
// hdc, and the function that frees both.
func scratchDC(hdc w32.HDC, w, h int32) (w32.HDC, func()) {
	dc := w32.CreateCompatibleDC(hdc)
	bmp := w32.CreateCompatibleBitmap(hdc, w, h)
	old := w32.SelectObject(dc, w32.HGDIOBJ(bmp))
	return dc, func() {
		w32.SelectObject(dc, old)
		w32.DeleteObject(w32.HGDIOBJ(bmp))
		w32.DeleteDC(dc)
	}
}

// drawableSize returns the size of a pixmap or of a window's client area.
func (d *WindowsDisplay) drawableSize(drawable platform.DrawableID) (w, h int32) {
	if pix := d.getPixmap(platform.PixmapID(drawable)); pix != nil {
		return int32(pix.width), int32(pix.height)
	}
	var rect w32.RECT
	w32.GetClientRect(toDrawableHWND(drawable), &rect)
	return rect.Right, rect.Bottom
}

// gdiPoints converts X points (relative to the previous one with
// CoordModePrevious) and returns their bounding box.
func gdiPoints(points []platform.Point, mode int) (pts []w32.POINT, left, top, right, bottom int32) {
	pts = make([]w32.POINT, len(points))
	var cx, cy int32
	for i, p := range points {
		if mode == platform.CoordModePrevious {
			cx += int32(p.X)
			cy += int32(p.Y)
		} else {
			cx, cy = int32(p.X), int32(p.Y)
		}
		pts[i] = w32.POINT{X: cx, Y: cy}
		if i == 0 {
			left, top, right, bottom = cx, cy, cx, cy
		}
		left, top = min(left, cx), min(top, cy)
		right, bottom = max(right, cx), max(bottom, cy)
	}
	return pts, left, top, right + 1, bottom + 1
}

func (d *WindowsDisplay) FillRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	g := d.getGC(gc)
	if g == nil {
		return
	}

	rect := w32.RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + int(width)),
		Bottom: int32(y + int(height)),
	}
	d.render(drawable, g, rect.Left, rect.Top, rect.Right, rect.Bottom, func(hdc w32.HDC) {
		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}
		w32.FillRect(hdc, &rect, g.getBrush())
	})
}

func (d *WindowsDisplay) DrawRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	g := d.getGC(gc)
	if g == nil {
		return
	}

	left, top := int32(x), int32(y)
	right, bottom := int32(x+int(width)+1), int32(y+int(height)+1)
	lw := max(int32(g.lineWidth), 1)
	d.render(drawable, g, left-lw, top-lw, right+lw, bottom+lw, func(hdc w32.HDC) {
		oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(g.getPen()))
		defer w32.SelectObject(hdc, oldPen)

		// Use null brush for outline only.
		oldBrush := w32.SelectObject(hdc, w32.GetStockObject(w32.NULL_BRUSH))
		defer w32.SelectObject(hdc, oldBrush)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		w32.GdiRectangle(hdc, left, top, right, bottom)
	})
}

func (d *WindowsDisplay) DrawLine(drawable platform.DrawableID, gc platform.GCID, x1, y1, x2, y2 int) {
	g := d.getGC(gc)
	if g == nil {
		return
	}

	lw := max(int32(g.lineWidth), 1)
	left, top := int32(min(x1, x2)), int32(min(y1, y2))
	right, bottom := int32(max(x1, x2))+1, int32(max(y1, y2))+1
	d.render(drawable, g, left-lw, top-lw, right+lw, bottom+lw, func(hdc w32.HDC) {
		oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(g.getPen()))
		defer w32.SelectObject(hdc, oldPen)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		w32.MoveToEx(hdc, int32(x1), int32(y1), nil)
		w32.LineTo(hdc, int32(x2), int32(y2))
	})
}

func (d *WindowsDisplay) DrawLines(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, mode int) {
	if len(points) < 2 {
		return
	}

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pts, left, top, right, bottom := gdiPoints(points, mode)
	lw := max(int32(g.lineWidth), 1)
	d.render(drawable, g, left-lw, top-lw, right+lw, bottom+lw, func(hdc w32.HDC) {
		oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(g.getPen()))
		defer w32.SelectObject(hdc, oldPen)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		w32.Polyline(hdc, &pts[0], int32(len(pts)))
	})
}

func (d *WindowsDisplay) FillPolygon(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, shape, mode int) {
	if len(points) < 3 {
		return
	}

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pts, left, top, right, bottom := gdiPoints(points, mode)
	d.render(drawable, g, left, top, right, bottom, func(hdc w32.HDC) {
		oldBrush := w32.SelectObject(hdc, w32.HGDIOBJ(g.getBrush()))
		defer w32.SelectObject(hdc, oldBrush)

		oldPen := w32.SelectObject(hdc, w32.GetStockObject(w32.NULL_PEN))
		defer w32.SelectObject(hdc, oldPen)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		w32.GdiPolygon(hdc, &pts[0], int32(len(pts)))
	})
}

func (d *WindowsDisplay) FillArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	g := d.getGC(gc)
	if g == nil {
		return
	}

	left := int32(x)
	top := int32(y)
	right := int32(x + int(width))
	bottom := int32(y + int(height))

	// X11 angles are in 64ths of a degree, counterclockwise from 3 o'clock.
	startX, startY, endX, endY := arcPoints(x, y, int(width), int(height), angle1, angle2)

	d.render(drawable, g, left, top, right+1, bottom+1, func(hdc w32.HDC) {
		oldBrush := w32.SelectObject(hdc, w32.HGDIOBJ(g.getBrush()))
		defer w32.SelectObject(hdc, oldBrush)

		oldPen := w32.SelectObject(hdc, w32.GetStockObject(w32.NULL_PEN))
		defer w32.SelectObject(hdc, oldPen)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		if angle2 >= 360*64 || angle2 <= -360*64 {
			// Full ellipse.
			w32.GdiEllipse(hdc, left, top, right, bottom)
		} else {
			w32.Pie(hdc, left, top, right, bottom, startX, startY, endX, endY)
		}
	})
}

func (d *WindowsDisplay) DrawArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	g := d.getGC(gc)
	if g == nil {
		return
	}

	left := int32(x)
	top := int32(y)
	right := int32(x + int(width))
	bottom := int32(y + int(height))

	startX, startY, endX, endY := arcPoints(x, y, int(width), int(height), angle1, angle2)

	lw := max(int32(g.lineWidth), 1)
	d.render(drawable, g, left-lw, top-lw, right+lw+1, bottom+lw+1, func(hdc w32.HDC) {
		oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(g.getPen()))
		defer w32.SelectObject(hdc, oldPen)

		// Ellipse fills with the DC's brush; only the outline is wanted.
		oldBrush := w32.SelectObject(hdc, w32.GetStockObject(w32.NULL_BRUSH))
		defer w32.SelectObject(hdc, oldBrush)

		if g.function != platform.GXcopy {
			w32.SetROP2(hdc, gcROP2(g.function))
		}

		if angle2 >= 360*64 || angle2 <= -360*64 {
			w32.GdiEllipse(hdc, left, top, right, bottom)
		} else {
			w32.GdiArc(hdc, left, top, right, bottom, startX, startY, endX, endY)
		}
	})
}

func (d *WindowsDisplay) ClearWindow(w platform.WindowID) {
	hwnd := toHWND(w)
	var rect w32.RECT
	w32.GetClientRect(hwnd, &rect)

	info := d.getWindowInfo(hwnd)
	bgPixel := uint64(0x00FFFFFF)
	if info != nil {
		bgPixel = info.bgPixel
	}

	hdc := w32.GetDC(hwnd)
	defer w32.ReleaseDC(hwnd, hdc)

	brush := w32.CreateSolidBrush(pixelToCOLORREF(bgPixel))
	defer w32.DeleteObject(w32.HGDIOBJ(brush))
	w32.FillRect(hdc, &rect, brush)
}

func (d *WindowsDisplay) SetWindowBackground(w platform.WindowID, pixel uint64) {
	hwnd := toHWND(w)
	d.windowMu.Lock()
	if info, ok := d.windowData[hwnd]; ok {
		info.bgPixel = pixel
	}
	d.windowMu.Unlock()
}

func (d *WindowsDisplay) ClearArea(w platform.WindowID, x, y int, width, height uint, exposures bool) {
	hwnd := toHWND(w)
	info := d.getWindowInfo(hwnd)
	bgPixel := uint64(0x00FFFFFF)
	if info != nil {
		bgPixel = info.bgPixel
	}

	hdc := w32.GetDC(hwnd)
	defer w32.ReleaseDC(hwnd, hdc)

	brush := w32.CreateSolidBrush(pixelToCOLORREF(bgPixel))
	defer w32.DeleteObject(w32.HGDIOBJ(brush))

	rect := w32.RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + int(width)),
		Bottom: int32(y + int(height)),
	}
	w32.FillRect(hdc, &rect, brush)

	if exposures {
		w32.InvalidateRect(hwnd, &rect, false)
	}
}

func (d *WindowsDisplay) CopyArea(src, dst platform.DrawableID, gc platform.GCID, srcX, srcY int, width, height uint, dstX, dstY int) {
	srcDC, srcCleanup := d.getDrawableDC(src)
	defer srcCleanup()
	dstDC, dstCleanup := d.getDrawableDC(dst)
	defer dstCleanup()

	g := d.getGC(gc)
	rop := uint32(w32.SRCCOPY)
	if g != nil {
		rop = gcBitBltROP(g.function)
	}

	w32.BitBlt(dstDC, int32(dstX), int32(dstY), int32(width), int32(height),
		srcDC, int32(srcX), int32(srcY), rop)
}

func (d *WindowsDisplay) PutImageRGBA(drawable platform.DrawableID, gc platform.GCID, depth int,
	rgbaData []byte, stride int, imgW, imgH int,
	srcX, srcY, dstX, dstY, w, h int, bgPixel uint64) {

	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	// Create a BITMAPINFO for a top-down 32-bit BGR(A) DIB.
	bmi := w32.BITMAPINFO{
		BmiHeader: w32.BITMAPINFOHEADER{
			BiSize:        uint32(unsafe.Sizeof(w32.BITMAPINFOHEADER{})),
			BiWidth:       int32(w),
			BiHeight:      -int32(h), // top-down
			BiPlanes:      1,
			BiBitCount:    32,
			BiCompression: w32.BI_RGB,
		},
	}

	var bits unsafe.Pointer
	memDC := w32.CreateCompatibleDC(hdc)
	defer w32.DeleteDC(memDC)

	dibBmp := w32.CreateDIBSection(memDC, &bmi, w32.DIB_RGB_COLORS, &bits, 0, 0)
	if dibBmp == 0 || bits == nil {
		return
	}
	defer w32.DeleteObject(w32.HGDIOBJ(dibBmp))

	oldBmp := w32.SelectObject(memDC, w32.HGDIOBJ(dibBmp))
	defer w32.SelectObject(memDC, oldBmp)

	// Convert RGBA → BGRA (Windows DIB format).
	dst := (*[1 << 30]byte)(bits)
	for row := range h {
		srcRow := (srcY + row) * stride
		dstRow := row * w * 4
		for col := range w {
			si := srcRow + (srcX+col)*4
			di := dstRow + col*4
			if si+3 < len(rgbaData) {
				r, g, b, a := rgbaData[si], rgbaData[si+1], rgbaData[si+2], rgbaData[si+3]
				// Composite over bgPixel like the X11 put_rgba_image: the
				// data is premultiplied, so out = src + bg*(1-a). BitBlt
				// ignores the alpha byte, so a partly transparent pixel
				// must be blended here.
				// A fully transparent pixel may still carry a colour;
				// put_rgba_image ignores it, so drop it here too.
				if a == 0 {
					r, g, b = 0, 0, 0
				}
				inv := uint32(255 - a)
				blend := func(c, bg byte) byte { return byte(min(uint32(c)+uint32(bg)*inv/255, 255)) }
				dst[di+0] = blend(b, byte(bgPixel))
				dst[di+1] = blend(g, byte(bgPixel>>8))
				dst[di+2] = blend(r, byte(bgPixel>>16))
				dst[di+3] = 0xFF
			}
		}
	}

	w32.BitBlt(hdc, int32(dstX), int32(dstY), int32(w), int32(h), memDC, 0, 0, w32.SRCCOPY)
}

// arcPoints converts X11 arc angles to Windows Arc start/end points.
// X11: angle in 64ths of a degree, counterclockwise from 3 o'clock.
// Windows: start/end points on the ellipse bounding box.
func arcPoints(x, y, w, h, angle1, angle2 int) (startX, startY, endX, endY int32) {
	cx := float64(x) + float64(w)/2.0
	cy := float64(y) + float64(h)/2.0
	rx := float64(w) / 2.0
	ry := float64(h) / 2.0

	// Start angle.
	a1 := float64(angle1) / 64.0 * math.Pi / 180.0
	startX = int32(cx + rx*math.Cos(a1))
	startY = int32(cy - ry*math.Sin(a1))

	// End angle (start + extent).
	a2 := float64(angle1+angle2) / 64.0 * math.Pi / 180.0
	endX = int32(cx + rx*math.Cos(a2))
	endY = int32(cy - ry*math.Sin(a2))

	return
}

func (d *WindowsDisplay) GetImageRGBA(platform.DrawableID, int, int, int, int) []byte {
	return nil
}
