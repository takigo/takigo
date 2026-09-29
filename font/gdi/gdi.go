//go:build windows

// Package gdi implements font.Font with GDI for the Windows backend, as
// tk/win/tkWinFont.c does.
package gdi

import (
	"github.com/msorc/takigo/font"
	"sort"
	"sync"
	"syscall"
	"unicode/utf16"
	"unsafe"

	w32 "github.com/msorc/takigo/internal/win32"
	"github.com/msorc/takigo/platform"
)

// DCResolver returns an HDC for a given DrawableID and a cleanup function.
// For pixmap drawables the cleanup is a no-op; for windows the DC must be released.
type DCResolver func(drawable platform.DrawableID) (w32.HDC, func())

// GDIFont implements font.Font and platform.DrawableFont using Windows GDI.
type GDIFont struct {
	hfont     w32.HFONT
	screenDC  w32.HDC
	attrs     font.Attributes
	metrics   font.Metrics
	resolveDC DCResolver

	// measureDC is a memory DC with hfont selected, made on the first
	// MeasureString and kept (Tk keeps its font's DC state likewise)
	// instead of a CreateCompatibleDC/DeleteDC pair per measurement.
	measureMu sync.Mutex
	measureDC w32.HDC
	oldFont   w32.HGDIOBJ
}

// OpenGDI creates a GDI font from font attributes.
// The resolver, if non-nil, is used by DrawString to obtain the correct HDC
// for any DrawableID (window or pixmap).
func OpenGDI(screenDC w32.HDC, attrs font.Attributes, resolver DCResolver) (*GDIFont, error) {
	var lf w32.LOGFONTW

	// Convert size. Positive Size = points, negative = pixels.
	if attrs.Size < 0 {
		lf.LfHeight = int32(attrs.Size) // negative pixel size → negative height
	} else if attrs.Size > 0 {
		// Convert points to pixels: height = -MulDiv(pointSize, dpi, 72)
		dpi := w32.GetDeviceCaps(screenDC, w32.LOGPIXELSY)
		if dpi == 0 {
			dpi = 96
		}
		lf.LfHeight = -int32(float64(attrs.Size) * float64(dpi) / 72.0)
	} else {
		// Default size.
		dpi := w32.GetDeviceCaps(screenDC, w32.LOGPIXELSY)
		if dpi == 0 {
			dpi = 96
		}
		lf.LfHeight = -int32(12.0 * float64(dpi) / 72.0)
	}

	// Weight.
	lf.LfWeight = w32.FW_NORMAL
	if attrs.Weight == font.WeightBold {
		lf.LfWeight = w32.FW_BOLD
	}

	// Slant.
	if attrs.Slant == font.SlantItalic || attrs.Slant == font.SlantOblique {
		lf.LfItalic = 1
	}

	// Underline/strikeout.
	if attrs.Underline {
		lf.LfUnderline = 1
	}
	if attrs.Overstrike {
		lf.LfStrikeOut = 1
	}

	lf.LfCharSet = w32.DEFAULT_CHARSET
	lf.LfOutPrecision = w32.OUT_TT_PRECIS
	lf.LfClipPrecision = w32.CLIP_DEFAULT_PRECIS
	lf.LfQuality = w32.CLEARTYPE_QUALITY
	lf.LfPitchAndFamily = w32.DEFAULT_PITCH | w32.FF_DONTCARE

	// Copy family name.
	family := attrs.Family
	if family == "" || family == "sans-serif" {
		family = "Segoe UI"
	}
	if family == "monospace" || family == "Courier" {
		family = "Consolas"
	}
	if family == "serif" || family == "Times" {
		family = "Times New Roman"
	}
	utf16Family := syscall.StringToUTF16(family)
	for i := 0; i < len(utf16Family) && i < 31; i++ {
		lf.LfFaceName[i] = utf16Family[i]
	}

	hfont := w32.CreateFontIndirect(&lf)
	if hfont == 0 {
		// Fallback to default GUI font.
		hfont = w32.HFONT(w32.GetStockObject(w32.DEFAULT_GUI_FONT))
	}

	f := &GDIFont{
		hfont:     hfont,
		screenDC:  screenDC,
		attrs:     attrs,
		resolveDC: resolver,
	}

	// Query metrics using a temporary DC.
	memDC := w32.CreateCompatibleDC(screenDC)
	oldFont := w32.SelectObject(memDC, w32.HGDIOBJ(hfont))
	var tm w32.TEXTMETRICW
	w32.GetTextMetrics(memDC, &tm)
	w32.SelectObject(memDC, oldFont)
	w32.DeleteDC(memDC)

	f.metrics = font.Metrics{
		Ascent:   int(tm.TmAscent),
		Descent:  int(tm.TmDescent),
		MaxWidth: int(tm.TmMaxCharWidth),
		Fixed:    (tm.TmPitchAndFamily & w32.TMPF_FIXED_PITCH) == 0, // counter-intuitive: TMPF_FIXED_PITCH means NOT fixed
	}

	// Update attrs with actual family name if we substituted.
	f.attrs.Family = family

	return f, nil
}

func (f *GDIFont) Attrs() font.Attributes { return f.attrs }
func (f *GDIFont) Metrics() font.Metrics  { return f.metrics }

func (f *GDIFont) MeasureString(s string) int {
	if s == "" {
		return 0
	}
	f.measureMu.Lock()
	defer f.measureMu.Unlock()
	if f.measureDC == 0 {
		if f.hfont == 0 {
			return 0
		}
		f.measureDC = w32.CreateCompatibleDC(f.screenDC)
		f.oldFont = w32.SelectObject(f.measureDC, w32.HGDIOBJ(f.hfont))
	}
	u := utf16.Encode([]rune(s))
	var size w32.SIZE
	w32.GetTextExtentPoint32(f.measureDC, &u[0], int32(len(u)), &size)
	return int(size.CX)
}

func (f *GDIFont) Close() {
	f.measureMu.Lock()
	if f.measureDC != 0 {
		w32.SelectObject(f.measureDC, f.oldFont)
		w32.DeleteDC(f.measureDC)
		f.measureDC = 0
	}
	f.measureMu.Unlock()
	if f.hfont != 0 {
		w32.DeleteObject(w32.HGDIOBJ(f.hfont))
		f.hfont = 0
	}
}

// DrawString implements platform.DrawableFont.
func (f *GDIFont) DrawString(drawable platform.DrawableID, x, y int, s string, pixel uint64, r, g, b uint16) {
	if s == "" {
		return
	}

	// Resolve the drawable to an HDC. The resolver knows how to handle both
	// window HWNDs and pixmap IDs (looking up the pixmap's memory DC).
	var hdc w32.HDC
	var cleanup func()
	if f.resolveDC != nil {
		hdc, cleanup = f.resolveDC(drawable)
	} else {
		// Fallback: treat as HWND.
		hwnd := w32.HWND(uintptr(drawable))
		hdc = w32.GetDC(hwnd)
		cleanup = func() { w32.ReleaseDC(hwnd, hdc) }
	}
	defer cleanup()

	if hdc == 0 {
		return
	}

	oldFont := w32.SelectObject(hdc, w32.HGDIOBJ(f.hfont))
	defer w32.SelectObject(hdc, oldFont)

	w32.SetBkMode(hdc, w32.TRANSPARENT)
	w32.SetTextColor(hdc, w32.RGB(byte(r>>8), byte(g>>8), byte(b>>8)))

	utf16Str := syscall.StringToUTF16(s)
	count := len(utf16Str)
	if count > 0 && utf16Str[count-1] == 0 {
		count--
	}
	w32.TextOut(hdc, int32(x), int32(y-f.metrics.Ascent), &utf16Str[0], int32(count))
}

// Verify DrawableFont at compile time.
var _ platform.DrawableFont = (*GDIFont)(nil)

// Go never frees a syscall.NewCallback and caps how many exist, so the
// enumeration callback is created once and reports to enumSink, which
// enumMu guards for the duration of one synchronous EnumFontFamiliesEx.
var (
	enumMu   sync.Mutex
	enumSink func(name string)
	enumCB   = sync.OnceValue(func() uintptr {
		return syscall.NewCallback(func(lf *w32.LOGFONTW, tm *w32.TEXTMETRICW, fontType uint32, lParam w32.LPARAM) uintptr {
			enumSink(utf16ToString(lf.LfFaceName[:]))
			return 1 // continue enumeration
		})
	})
)

// ListFamilies returns all available font families on Windows.
func ListFamilies() []string {
	screenDC := w32.GetDC(0)
	defer w32.ReleaseDC(0, screenDC)

	var families []string
	seen := make(map[string]bool)

	var lf w32.LOGFONTW
	lf.LfCharSet = w32.DEFAULT_CHARSET

	enumMu.Lock()
	enumSink = func(name string) {
		if !seen[name] {
			seen[name] = true
			families = append(families, name)
		}
	}
	w32.EnumFontFamiliesEx(screenDC, &lf, enumCB(), 0, 0)
	enumSink = nil
	enumMu.Unlock()

	sort.Strings(families)
	return families
}

// utf16ToString converts a fixed-size UTF-16 array to a Go string.
func utf16ToString(s []uint16) string {
	for i, v := range s {
		if v == 0 {
			return string(syscall.UTF16ToString(s[:i]))
		}
	}
	return string(syscall.UTF16ToString(s))
}

// Ensure unused import doesn't cause error.
var _ = unsafe.Sizeof(0)
