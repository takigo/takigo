//go:build windows

package win32

import (
	"syscall"
	"unsafe"
)

var (
	gdi32 = syscall.NewLazyDLL("gdi32.dll")

	procCreateCompatibleDC      = gdi32.NewProc("CreateCompatibleDC")
	procCreateCompatibleBitmap  = gdi32.NewProc("CreateCompatibleBitmap")
	procDeleteDC                = gdi32.NewProc("DeleteDC")
	procDeleteObject            = gdi32.NewProc("DeleteObject")
	procSelectObject            = gdi32.NewProc("SelectObject")
	procGetStockObject          = gdi32.NewProc("GetStockObject")
	procCreateSolidBrush        = gdi32.NewProc("CreateSolidBrush")
	procCreatePen               = gdi32.NewProc("CreatePen")
	procExtCreatePen            = gdi32.NewProc("ExtCreatePen")
	procRectangle               = gdi32.NewProc("Rectangle")
	procEllipse                 = gdi32.NewProc("Ellipse")
	procMoveToEx                = gdi32.NewProc("MoveToEx")
	procLineTo                  = gdi32.NewProc("LineTo")
	procPolyline                = gdi32.NewProc("Polyline")
	procPolygon                 = gdi32.NewProc("Polygon")
	procArc                     = gdi32.NewProc("Arc")
	procChord                   = gdi32.NewProc("Chord")
	procPie                     = gdi32.NewProc("Pie")
	procPatBlt                  = gdi32.NewProc("PatBlt")
	procBitBlt                  = gdi32.NewProc("BitBlt")
	procStretchBlt              = gdi32.NewProc("StretchBlt")
	procSetROP2                 = gdi32.NewProc("SetROP2")
	procSetBkMode               = gdi32.NewProc("SetBkMode")
	procSetBkColor              = gdi32.NewProc("SetBkColor")
	procSetTextColor            = gdi32.NewProc("SetTextColor")
	procCreateFontIndirectW     = gdi32.NewProc("CreateFontIndirectW")
	procTextOutW                = gdi32.NewProc("TextOutW")
	procGetTextExtentPoint32W   = gdi32.NewProc("GetTextExtentPoint32W")
	procGetTextMetricsW         = gdi32.NewProc("GetTextMetricsW")
	procGetDeviceCaps           = gdi32.NewProc("GetDeviceCaps")
	procSetPixel                = gdi32.NewProc("SetPixel")
	procGetPixel                = gdi32.NewProc("GetPixel")
	procCreateDIBSection        = gdi32.NewProc("CreateDIBSection")
	procSetDIBitsToDevice       = gdi32.NewProc("SetDIBitsToDevice")
	procSetArcDirection         = gdi32.NewProc("SetArcDirection")
	procGetCharWidth32W         = gdi32.NewProc("GetCharWidth32W")
	procEnumFontFamiliesExW     = gdi32.NewProc("EnumFontFamiliesExW")
	procSaveDC                  = gdi32.NewProc("SaveDC")
	procRestoreDC               = gdi32.NewProc("RestoreDC")
)

func CreateCompatibleDC(hdc HDC) HDC {
	r, _, _ := procCreateCompatibleDC.Call(uintptr(hdc))
	return HDC(r)
}

func CreateCompatibleBitmap(hdc HDC, width, height int32) HBITMAP {
	r, _, _ := procCreateCompatibleBitmap.Call(uintptr(hdc), uintptr(width), uintptr(height))
	return HBITMAP(r)
}

func DeleteDC(hdc HDC) bool {
	r, _, _ := procDeleteDC.Call(uintptr(hdc))
	return r != 0
}

func DeleteObject(obj HGDIOBJ) bool {
	r, _, _ := procDeleteObject.Call(uintptr(obj))
	return r != 0
}

func SelectObject(hdc HDC, obj HGDIOBJ) HGDIOBJ {
	r, _, _ := procSelectObject.Call(uintptr(hdc), uintptr(obj))
	return HGDIOBJ(r)
}

func GetStockObject(index int32) HGDIOBJ {
	r, _, _ := procGetStockObject.Call(uintptr(index))
	return HGDIOBJ(r)
}

func CreateSolidBrush(color COLORREF) HBRUSH {
	r, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return HBRUSH(r)
}

func CreatePen(style, width int32, color COLORREF) HPEN {
	r, _, _ := procCreatePen.Call(uintptr(style), uintptr(width), uintptr(color))
	return HPEN(r)
}

func ExtCreatePen(style uint32, width uint32, lb *LOGBRUSH, numEntries uint32, style2 *uint32) HPEN {
	r, _, _ := procExtCreatePen.Call(uintptr(style), uintptr(width),
		uintptr(unsafe.Pointer(lb)), uintptr(numEntries), uintptr(unsafe.Pointer(style2)))
	return HPEN(r)
}

func GdiRectangle(hdc HDC, left, top, right, bottom int32) bool {
	r, _, _ := procRectangle.Call(uintptr(hdc),
		uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return r != 0
}

func GdiEllipse(hdc HDC, left, top, right, bottom int32) bool {
	r, _, _ := procEllipse.Call(uintptr(hdc),
		uintptr(left), uintptr(top), uintptr(right), uintptr(bottom))
	return r != 0
}

func MoveToEx(hdc HDC, x, y int32, prevPoint *POINT) bool {
	r, _, _ := procMoveToEx.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(unsafe.Pointer(prevPoint)))
	return r != 0
}

func LineTo(hdc HDC, x, y int32) bool {
	r, _, _ := procLineTo.Call(uintptr(hdc), uintptr(x), uintptr(y))
	return r != 0
}

func Polyline(hdc HDC, points *POINT, count int32) bool {
	r, _, _ := procPolyline.Call(uintptr(hdc), uintptr(unsafe.Pointer(points)), uintptr(count))
	return r != 0
}

func GdiPolygon(hdc HDC, points *POINT, count int32) bool {
	r, _, _ := procPolygon.Call(uintptr(hdc), uintptr(unsafe.Pointer(points)), uintptr(count))
	return r != 0
}

func GdiArc(hdc HDC, x1, y1, x2, y2, x3, y3, x4, y4 int32) bool {
	r, _, _ := procArc.Call(uintptr(hdc),
		uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2),
		uintptr(x3), uintptr(y3), uintptr(x4), uintptr(y4))
	return r != 0
}

func Chord(hdc HDC, x1, y1, x2, y2, x3, y3, x4, y4 int32) bool {
	r, _, _ := procChord.Call(uintptr(hdc),
		uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2),
		uintptr(x3), uintptr(y3), uintptr(x4), uintptr(y4))
	return r != 0
}

func Pie(hdc HDC, x1, y1, x2, y2, x3, y3, x4, y4 int32) bool {
	r, _, _ := procPie.Call(uintptr(hdc),
		uintptr(x1), uintptr(y1), uintptr(x2), uintptr(y2),
		uintptr(x3), uintptr(y3), uintptr(x4), uintptr(y4))
	return r != 0
}

func PatBlt(hdc HDC, x, y, w, h int32, rop uint32) bool {
	r, _, _ := procPatBlt.Call(uintptr(hdc),
		uintptr(x), uintptr(y), uintptr(w), uintptr(h), uintptr(rop))
	return r != 0
}

func BitBlt(hdcDest HDC, x, y, cx, cy int32, hdcSrc HDC, x1, y1 int32, rop uint32) bool {
	r, _, _ := procBitBlt.Call(uintptr(hdcDest),
		uintptr(x), uintptr(y), uintptr(cx), uintptr(cy),
		uintptr(hdcSrc), uintptr(x1), uintptr(y1), uintptr(rop))
	return r != 0
}

func SetROP2(hdc HDC, mode int32) int32 {
	r, _, _ := procSetROP2.Call(uintptr(hdc), uintptr(mode))
	return int32(r)
}

func SetBkMode(hdc HDC, mode int32) int32 {
	r, _, _ := procSetBkMode.Call(uintptr(hdc), uintptr(mode))
	return int32(r)
}

func SetBkColor(hdc HDC, color COLORREF) COLORREF {
	r, _, _ := procSetBkColor.Call(uintptr(hdc), uintptr(color))
	return COLORREF(r)
}

func SetTextColor(hdc HDC, color COLORREF) COLORREF {
	r, _, _ := procSetTextColor.Call(uintptr(hdc), uintptr(color))
	return COLORREF(r)
}

func CreateFontIndirect(lf *LOGFONTW) HFONT {
	r, _, _ := procCreateFontIndirectW.Call(uintptr(unsafe.Pointer(lf)))
	return HFONT(r)
}

func TextOut(hdc HDC, x, y int32, str *uint16, count int32) bool {
	r, _, _ := procTextOutW.Call(uintptr(hdc), uintptr(x), uintptr(y),
		uintptr(unsafe.Pointer(str)), uintptr(count))
	return r != 0
}

func GetTextExtentPoint32(hdc HDC, str *uint16, count int32, size *SIZE) bool {
	r, _, _ := procGetTextExtentPoint32W.Call(uintptr(hdc),
		uintptr(unsafe.Pointer(str)), uintptr(count), uintptr(unsafe.Pointer(size)))
	return r != 0
}

func GetTextMetrics(hdc HDC, tm *TEXTMETRICW) bool {
	r, _, _ := procGetTextMetricsW.Call(uintptr(hdc), uintptr(unsafe.Pointer(tm)))
	return r != 0
}

func GetDeviceCaps(hdc HDC, index int32) int32 {
	r, _, _ := procGetDeviceCaps.Call(uintptr(hdc), uintptr(index))
	return int32(r)
}

func SetPixel(hdc HDC, x, y int32, color COLORREF) COLORREF {
	r, _, _ := procSetPixel.Call(uintptr(hdc), uintptr(x), uintptr(y), uintptr(color))
	return COLORREF(r)
}

func GetPixel(hdc HDC, x, y int32) COLORREF {
	r, _, _ := procGetPixel.Call(uintptr(hdc), uintptr(x), uintptr(y))
	return COLORREF(r)
}

func CreateDIBSection(hdc HDC, bmi *BITMAPINFO, usage uint32, bits *unsafe.Pointer,
	hSection HANDLE, offset uint32) HBITMAP {
	r, _, _ := procCreateDIBSection.Call(uintptr(hdc),
		uintptr(unsafe.Pointer(bmi)), uintptr(usage),
		uintptr(unsafe.Pointer(bits)), uintptr(hSection), uintptr(offset))
	return HBITMAP(r)
}

func SetDIBitsToDevice(hdc HDC, xDest, yDest, w, h int32, xSrc, ySrc int32,
	startScan, numScans uint32, bits unsafe.Pointer, bmi *BITMAPINFO, usage uint32) int32 {
	r, _, _ := procSetDIBitsToDevice.Call(uintptr(hdc),
		uintptr(xDest), uintptr(yDest), uintptr(w), uintptr(h),
		uintptr(xSrc), uintptr(ySrc),
		uintptr(startScan), uintptr(numScans),
		uintptr(bits), uintptr(unsafe.Pointer(bmi)), uintptr(usage))
	return int32(r)
}

func SetArcDirection(hdc HDC, dir int32) int32 {
	r, _, _ := procSetArcDirection.Call(uintptr(hdc), uintptr(dir))
	return int32(r)
}

func GetCharWidth32(hdc HDC, first, last uint32, widths *int32) bool {
	r, _, _ := procGetCharWidth32W.Call(uintptr(hdc), uintptr(first), uintptr(last),
		uintptr(unsafe.Pointer(widths)))
	return r != 0
}

func EnumFontFamiliesEx(hdc HDC, lf *LOGFONTW, proc uintptr, lParam LPARAM, flags DWORD) int32 {
	r, _, _ := procEnumFontFamiliesExW.Call(uintptr(hdc),
		uintptr(unsafe.Pointer(lf)), proc, uintptr(lParam), uintptr(flags))
	return int32(r)
}

func SaveDC(hdc HDC) int32 {
	r, _, _ := procSaveDC.Call(uintptr(hdc))
	return int32(r)
}

func RestoreDC(hdc HDC, savedDC int32) bool {
	r, _, _ := procRestoreDC.Call(uintptr(hdc), uintptr(savedDC))
	return r != 0
}
