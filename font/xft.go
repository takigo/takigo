package font

/*
#cgo pkg-config: xft fontconfig
#include <X11/Xlib.h>
#include <X11/Xft/Xft.h>
#include <fontconfig/fontconfig.h>
#include <stdlib.h>
#include <string.h>

// Helper to create an XftColor from pixel + RGB.
static XftColor make_xft_color(unsigned long pixel, unsigned short r, unsigned short g, unsigned short b) {
	XftColor c;
	c.pixel = pixel;
	c.color.red = r;
	c.color.green = g;
	c.color.blue = b;
	c.color.alpha = 0xFFFF;
	return c;
}

// Helpers wrapping FC_* string macros so cgo can use them.
static FcBool fc_pattern_add_family(FcPattern *p, const FcChar8 *family) {
	return FcPatternAddString(p, FC_FAMILY, family);
}
static FcBool fc_pattern_add_size(FcPattern *p, double size) {
	return FcPatternAddDouble(p, FC_SIZE, size);
}
static FcBool fc_pattern_add_pixel_size(FcPattern *p, double size) {
	return FcPatternAddDouble(p, FC_PIXEL_SIZE, size);
}
static FcBool fc_pattern_add_weight(FcPattern *p, int weight) {
	return FcPatternAddInteger(p, FC_WEIGHT, weight);
}
static FcBool fc_pattern_add_slant(FcPattern *p, int slant) {
	return FcPatternAddInteger(p, FC_SLANT, slant);
}
static FcResult fc_pattern_get_spacing(FcPattern *p, int *spacing) {
	return FcPatternGetInteger(p, FC_SPACING, 0, spacing);
}
*/
import "C"
import (
	"fmt"
	"unsafe"

	"github.com/msorc/takigo/internal/xlib"
)

// XftFont wraps an Xft font handle and provides text measurement and drawing.
type XftFont struct {
	display  *xlib.Display
	font     *C.XftFont
	draw     *C.XftDraw
	attrs    Attributes
	metrics  Metrics
	screen   C.int
	visual   *C.Visual
	colormap C.Colormap
}

// OpenXft opens a font via Xft/fontconfig.
func OpenXft(display *xlib.Display, screen int, visual *xlib.Visual, colormap xlib.Colormap, attrs Attributes) (*XftFont, error) {
	dpy := (*C.Display)(display.Ptr())
	cscreen := C.int(screen)

	// Build fontconfig pattern from attributes.
	family := attrs.Family
	if family == "" {
		family = "sans-serif"
	}
	cfamily := C.CString(family)
	defer C.free(unsafe.Pointer(cfamily))

	size := attrs.Size
	if size == 0 {
		size = 12
	}

	var pixelSize bool
	if size < 0 {
		size = -size
		pixelSize = true
	}

	weight := C.int(C.FC_WEIGHT_MEDIUM)
	if attrs.Weight == WeightBold {
		weight = C.FC_WEIGHT_BOLD
	}

	slant := C.int(C.FC_SLANT_ROMAN)
	switch attrs.Slant {
	case SlantItalic:
		slant = C.FC_SLANT_ITALIC
	case SlantOblique:
		slant = C.FC_SLANT_OBLIQUE
	}

	pattern := C.FcPatternCreate()
	C.fc_pattern_add_family(pattern, (*C.FcChar8)(unsafe.Pointer(cfamily)))
	if pixelSize {
		C.fc_pattern_add_pixel_size(pattern, C.double(size))
	} else {
		C.fc_pattern_add_size(pattern, C.double(size))
	}
	C.fc_pattern_add_weight(pattern, weight)
	C.fc_pattern_add_slant(pattern, slant)

	C.FcConfigSubstitute(nil, pattern, C.FcMatchPattern)
	C.XftDefaultSubstitute(dpy, cscreen, pattern)

	var result C.FcResult
	matched := C.FcFontMatch(nil, pattern, &result)
	C.FcPatternDestroy(pattern)

	if matched == nil {
		return nil, fmt.Errorf("font: no match for %q", family)
	}

	xftFont := C.XftFontOpenPattern(dpy, matched)
	if xftFont == nil {
		C.FcPatternDestroy(matched)
		return nil, fmt.Errorf("font: failed to open %q", family)
	}

	f := &XftFont{
		display:  display,
		font:     xftFont,
		attrs:    attrs,
		screen:   cscreen,
		visual:   (*C.Visual)(unsafe.Pointer(visual.Ptr())),
		colormap: C.Colormap(colormap),
	}

	// Extract metrics.
	f.metrics = Metrics{
		Ascent:   int(xftFont.ascent),
		Descent:  int(xftFont.descent),
		MaxWidth: int(xftFont.max_advance_width),
	}

	// Check if monospace via spacing property.
	var spacing C.int
	if C.fc_pattern_get_spacing(xftFont.pattern, &spacing) == C.FcResultMatch {
		f.metrics.Fixed = spacing != C.FC_PROPORTIONAL
	}

	return f, nil
}

// Attrs returns the font's attributes.
func (f *XftFont) Attrs() Attributes {
	return f.attrs
}

// Metrics returns the font metrics.
func (f *XftFont) Metrics() Metrics {
	return f.metrics
}

// MeasureString returns the pixel width of a string.
func (f *XftFont) MeasureString(s string) int {
	if len(s) == 0 {
		return 0
	}
	dpy := (*C.Display)(f.display.Ptr())
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	var extents C.XGlyphInfo
	C.XftTextExtentsUtf8(dpy, f.font, (*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(s)), &extents)
	return int(extents.xOff)
}

// DrawString draws a string on a drawable at the given position.
// The position is the baseline origin.
func (f *XftFont) DrawString(drawable xlib.Drawable, x, y int, s string, pixel uint64, r, g, b uint16) {
	if len(s) == 0 {
		return
	}

	dpy := (*C.Display)(f.display.Ptr())

	// Create or reuse XftDraw.
	if f.draw == nil {
		f.draw = C.XftDrawCreate(dpy, C.Drawable(drawable), f.visual, f.colormap)
	} else {
		C.XftDrawChange(f.draw, C.Drawable(drawable))
	}

	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))

	color := C.make_xft_color(C.ulong(pixel), C.ushort(r), C.ushort(g), C.ushort(b))
	C.XftDrawStringUtf8(f.draw, &color, f.font, C.int(x), C.int(y),
		(*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(s)))
}

// Close releases font resources.
func (f *XftFont) Close() {
	dpy := (*C.Display)(f.display.Ptr())
	if f.draw != nil {
		C.XftDrawDestroy(f.draw)
		f.draw = nil
	}
	if f.font != nil {
		C.XftFontClose(dpy, f.font)
		f.font = nil
	}
}
