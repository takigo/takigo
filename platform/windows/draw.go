//go:build windows

package windows

import (
	"math"
	"unsafe"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// --- Drawer implementation ---

func (d *WindowsDisplay) FillRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	brush := g.getBrush()

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	rect := w32.RECT{
		Left:   int32(x),
		Top:    int32(y),
		Right:  int32(x + int(width)),
		Bottom: int32(y + int(height)),
	}
	w32.FillRect(hdc, &rect, brush)
}

func (d *WindowsDisplay) DrawRectangle(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pen := g.getPen()
	oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(pen))
	defer w32.SelectObject(hdc, oldPen)

	// Use null brush for outline only.
	nullBrush := w32.GetStockObject(w32.NULL_BRUSH)
	oldBrush := w32.SelectObject(hdc, nullBrush)
	defer w32.SelectObject(hdc, oldBrush)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	w32.GdiRectangle(hdc, int32(x), int32(y), int32(x+int(width)+1), int32(y+int(height)+1))
}

func (d *WindowsDisplay) DrawLine(drawable platform.DrawableID, gc platform.GCID, x1, y1, x2, y2 int) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pen := g.getPen()
	oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(pen))
	defer w32.SelectObject(hdc, oldPen)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	w32.MoveToEx(hdc, int32(x1), int32(y1), nil)
	w32.LineTo(hdc, int32(x2), int32(y2))
}

func (d *WindowsDisplay) DrawLines(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, mode int) {
	if len(points) < 2 {
		return
	}

	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pen := g.getPen()
	oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(pen))
	defer w32.SelectObject(hdc, oldPen)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	// Convert points. If CoordModePrevious, accumulate offsets.
	pts := make([]w32.POINT, len(points))
	if mode == platform.CoordModePrevious {
		var cx, cy int16
		for i, p := range points {
			cx += p.X
			cy += p.Y
			pts[i] = w32.POINT{X: int32(cx), Y: int32(cy)}
		}
	} else {
		for i, p := range points {
			pts[i] = w32.POINT{X: int32(p.X), Y: int32(p.Y)}
		}
	}

	w32.Polyline(hdc, &pts[0], int32(len(pts)))
}

func (d *WindowsDisplay) FillPolygon(drawable platform.DrawableID, gc platform.GCID, points []platform.Point, shape, mode int) {
	if len(points) < 3 {
		return
	}

	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	brush := g.getBrush()
	oldBrush := w32.SelectObject(hdc, w32.HGDIOBJ(brush))
	defer w32.SelectObject(hdc, oldBrush)

	pen := w32.GetStockObject(w32.NULL_PEN)
	oldPen := w32.SelectObject(hdc, pen)
	defer w32.SelectObject(hdc, oldPen)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	pts := make([]w32.POINT, len(points))
	if mode == platform.CoordModePrevious {
		var cx, cy int16
		for i, p := range points {
			cx += p.X
			cy += p.Y
			pts[i] = w32.POINT{X: int32(cx), Y: int32(cy)}
		}
	} else {
		for i, p := range points {
			pts[i] = w32.POINT{X: int32(p.X), Y: int32(p.Y)}
		}
	}

	w32.GdiPolygon(hdc, &pts[0], int32(len(pts)))
}

func (d *WindowsDisplay) FillArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	brush := g.getBrush()
	oldBrush := w32.SelectObject(hdc, w32.HGDIOBJ(brush))
	defer w32.SelectObject(hdc, oldBrush)

	pen := w32.GetStockObject(w32.NULL_PEN)
	oldPen := w32.SelectObject(hdc, pen)
	defer w32.SelectObject(hdc, oldPen)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	left := int32(x)
	top := int32(y)
	right := int32(x + int(width))
	bottom := int32(y + int(height))

	// X11 angles are in 64ths of a degree, counterclockwise from 3 o'clock.
	startX, startY, endX, endY := arcPoints(x, y, int(width), int(height), angle1, angle2)

	if angle2 >= 360*64 || angle2 <= -360*64 {
		// Full ellipse.
		w32.GdiEllipse(hdc, left, top, right, bottom)
	} else {
		w32.Pie(hdc, left, top, right, bottom, startX, startY, endX, endY)
	}
}

func (d *WindowsDisplay) DrawArc(drawable platform.DrawableID, gc platform.GCID, x, y int, width, height uint, angle1, angle2 int) {
	hdc, cleanup := d.getDrawableDC(drawable)
	defer cleanup()

	g := d.getGC(gc)
	if g == nil {
		return
	}

	pen := g.getPen()
	oldPen := w32.SelectObject(hdc, w32.HGDIOBJ(pen))
	defer w32.SelectObject(hdc, oldPen)

	if g.function != platform.GXcopy {
		w32.SetROP2(hdc, gcROP2(g.function))
	}

	left := int32(x)
	top := int32(y)
	right := int32(x + int(width))
	bottom := int32(y + int(height))

	startX, startY, endX, endY := arcPoints(x, y, int(width), int(height), angle1, angle2)

	if angle2 >= 360*64 || angle2 <= -360*64 {
		w32.GdiEllipse(hdc, left, top, right, bottom)
	} else {
		w32.GdiArc(hdc, left, top, right, bottom, startX, startY, endX, endY)
	}
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
