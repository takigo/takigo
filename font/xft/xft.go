//go:build linux || freebsd || openbsd || netbsd

// Package xft implements font.Font with Xft and fontconfig for the X11
// backend, as tk/unix/tkUnixRFont.c does.
package xft

/*
#cgo pkg-config: xft fontconfig
#include <X11/Xlib.h>
#include <X11/Xft/Xft.h>
#include <fontconfig/fontconfig.h>
#include <math.h>
#include <stdlib.h>
#include <string.h>

// alloc_xft_color makes the XftColor for r, g, b with XftColorAllocValue,
// as LookUpColor in tkUnixRFont.c does: it derives the pixel for the
// visual (allocating a cell on colormapped visuals), which Xft's core
// rendering path needs.
static XftColor alloc_xft_color(Display *dpy, Visual *visual, Colormap cmap,
	unsigned short r, unsigned short g, unsigned short b) {
	XRenderColor rc;
	rc.red = r; rc.green = g; rc.blue = b; rc.alpha = 0xFFFF;
	XftColor c;
	if (!XftColorAllocValue(dpy, visual, cmap, &rc, &c)) {
		c.pixel = 0;
		c.color = rc;
	}
	return c;
}

// no_unref_fonts makes XftFontClose destroy fonts at once instead of keeping
// them in Xft's per-display cache of unreferenced fonts. Cached fonts are
// otherwise destroyed by Xft's XCloseDisplay hook after libXrender has
// already dropped the display, so freeing their glyph sets sends requests
// with a stale RENDER opcode (garbage BadRequest/BadLength errors, and on a
// later display at the same address, corrupted replies or a hang).
static void no_unref_fonts(Display *dpy) {
	FcPattern *p = FcPatternCreate();
	if (!p) return;
	FcPatternAddInteger(p, XFT_MAX_UNREF_FONTS, 0);
	XftDefaultSet(dpy, p);
}

// open_xft_font_rotated creates a rotated variant of base_font.
// sin_a and cos_a are sin/cos of the rotation angle (clockwise on screen,
// matching Tk's canvas -angle convention).
static XftFont* open_xft_font_rotated(Display *dpy, int screen,
		XftFont *base_font, double sin_a, double cos_a) {
	FcPattern *pat = FcPatternDuplicate(base_font->pattern);
	if (!pat) return NULL;
	FcMatrix mat;
	mat.xx = mat.yy = cos_a;
	mat.xy = -sin_a;
	mat.yx = sin_a;
	FcPatternAddMatrix(pat, FC_MATRIX, &mat);
	FcConfigSubstitute(NULL, pat, FcMatchPattern);
	XftDefaultSubstitute(dpy, screen, pat);
	FcResult result;
	FcPattern *matched = FcFontMatch(NULL, pat, &result);
	FcPatternDestroy(pat);
	if (!matched) return NULL;
	XftFont *font = XftFontOpenPattern(dpy, matched);
	return font; // matched consumed by XftFontOpenPattern
}

// find_fallback_font finds the best font that contains ucs4, matching the
// size/weight/slant of base_pattern. Returns NULL if none found.
// The returned XftFont must be closed by the caller.
static XftFont* find_fallback_font(Display *dpy, int screen,
		FcPattern *base_pattern, FcChar32 ucs4) {
	FcCharSet *cs = FcCharSetCreate();
	if (!cs) return NULL;
	FcCharSetAddChar(cs, ucs4);

	FcPattern *pat = FcPatternCreate();
	if (!pat) { FcCharSetDestroy(cs); return NULL; }

	// Copy size/weight/slant from base pattern.
	double size; int weight, slant;
	if (FcPatternGetDouble(base_pattern, FC_SIZE, 0, &size) == FcResultMatch)
		FcPatternAddDouble(pat, FC_SIZE, size);
	if (FcPatternGetInteger(base_pattern, FC_WEIGHT, 0, &weight) == FcResultMatch)
		FcPatternAddInteger(pat, FC_WEIGHT, weight);
	if (FcPatternGetInteger(base_pattern, FC_SLANT, 0, &slant) == FcResultMatch)
		FcPatternAddInteger(pat, FC_SLANT, slant);

	FcPatternAddCharSet(pat, FC_CHARSET, cs);
	FcCharSetDestroy(cs);

	FcConfigSubstitute(NULL, pat, FcMatchPattern);
	XftDefaultSubstitute(dpy, screen, pat);

	FcResult result;
	FcPattern *matched = FcFontMatch(NULL, pat, &result);
	FcPatternDestroy(pat);
	if (!matched) return NULL;
	XftFont *font = XftFontOpenPattern(dpy, matched); // matched consumed
	return font;
}

// xft_char_exists wraps XftCharExists.
static FcBool xft_char_exists(Display *dpy, XftFont *font, FcChar32 ucs4) {
	return XftCharExists(dpy, font, ucs4);
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

// Helper for listing font families.
static FcObjectSet *FcObjectSetBuild_helper(void) {
	return FcObjectSetBuild(FC_FAMILY, NULL);
}
static FcPattern *fc_fontset_get_font(FcFontSet *fs, int i) {
	return fs->fonts[i];
}
static FcResult fc_pattern_get_family(FcPattern *p, FcChar8 **family) {
	return FcPatternGetString(p, FC_FAMILY, 0, family);
}
*/
import "C"
import (
	"fmt"
	"math"
	"unicode/utf8"
	"unsafe"

	"github.com/takigo/takigo/font"

	"github.com/takigo/takigo/internal/xlib"
	"github.com/takigo/takigo/platform"
)

// XftFont wraps an Xft font handle and provides text measurement and drawing.
type XftFont struct {
	display  *xlib.Display
	font     *C.XftFont
	attrs    font.Attributes
	metrics  font.Metrics
	screen   C.int
	visual   *C.Visual
	colormap C.Colormap

	// colors caches XftColors by 16-bit RGB (see xftColor).
	colors map[uint64]C.XftColor

	// Cache of rotated font variants, keyed by angle×10 (integer tenths of degrees).
	rotatedVariants map[int64]*C.XftFont

	// Per-rune font selection cache: maps rune → XftFont to use for that rune.
	// Value is f.font (primary) or a fallback font handle.
	fontByRune map[rune]*C.XftFont
	// Set of opened fallback fonts (for cleanup in Close).
	fallbackFonts map[*C.XftFont]bool
	// fallbackOrder lists fallbackFonts in the order they were opened; a
	// rune the primary font lacks is looked up in them before fontconfig.
	fallbackOrder []*C.XftFont

	// Glyph advance cache for MeasureString: asciiAdvance[r] is the width
	// of ASCII rune r plus one (0 = not measured yet); advance holds the
	// rest. Xft sums advances without kerning, so a string's width is the
	// sum of its runes' widths.
	asciiAdvance [128]int32
	advance      map[rune]int
}

// OpenXft opens a font via Xft/fontconfig.
func OpenXft(display *xlib.Display, screen int, visual *xlib.Visual, colormap xlib.Colormap, attrs font.Attributes) (*XftFont, error) {
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
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
	if attrs.Weight == font.WeightBold {
		weight = C.FC_WEIGHT_BOLD
	}

	slant := C.int(C.FC_SLANT_ROMAN)
	switch attrs.Slant {
	case font.SlantItalic:
		slant = C.FC_SLANT_ITALIC
	case font.SlantOblique:
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
	f.metrics = font.Metrics{
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
func (f *XftFont) Attrs() font.Attributes {
	return f.attrs
}

// Metrics returns the font metrics.
func (f *XftFont) Metrics() font.Metrics {
	return f.metrics
}

// MeasureString returns the pixel width of a string, using font fallback.
func (f *XftFont) MeasureString(s string) int {
	if !utf8.ValidString(s) {
		return f.measureUncached(s)
	}
	total := 0
	for _, r := range s {
		total += f.runeAdvance(r)
	}
	return total
}

// runeAdvance returns the advance width of r in the font that draws it.
func (f *XftFont) runeAdvance(r rune) int {
	if r < 128 {
		if w := f.asciiAdvance[r]; w != 0 {
			return int(w) - 1
		}
		w := f.measureUncached(string(r))
		f.asciiAdvance[r] = int32(w) + 1
		return w
	}
	if w, ok := f.advance[r]; ok {
		return w
	}
	if f.advance == nil {
		f.advance = make(map[rune]int)
	}
	w := f.measureUncached(string(r))
	f.advance[r] = w
	return w
}

// measureUncached measures s with XftTextExtentsUtf8, run by run.
func (f *XftFont) measureUncached(s string) int {
	if len(s) == 0 {
		return 0
	}
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
	dpy := (*C.Display)(f.display.Ptr())
	total := 0
	for _, run := range f.runsFor(s) {
		cs := C.CString(run.text)
		var extents C.XGlyphInfo
		C.XftTextExtentsUtf8(dpy, run.font, (*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(run.text)), &extents)
		C.free(unsafe.Pointer(cs))
		total += int(extents.xOff)
	}
	return total
}

// fontRun is a contiguous run of UTF-8 text that uses the same XftFont.
type fontRun struct {
	font *C.XftFont
	text string
}

// fontForRune returns the best XftFont for drawing rune r (with caching).
// The caller holds xlib.XftMu.
func (f *XftFont) fontForRune(r rune) *C.XftFont {
	if f.fontByRune == nil {
		f.fontByRune = make(map[rune]*C.XftFont)
	}
	if xf, ok := f.fontByRune[r]; ok {
		return xf
	}
	dpy := (*C.Display)(f.display.Ptr())
	var chosen *C.XftFont
	if C.xft_char_exists(dpy, f.font, C.FcChar32(r)) != 0 {
		chosen = f.font
	} else {
		// A fallback opened for an earlier rune often covers this one
		// (a CJK paragraph), which is far cheaper than FcFontMatch.
		for _, fb := range f.fallbackOrder {
			if C.xft_char_exists(dpy, fb, C.FcChar32(r)) != 0 {
				chosen = fb
				break
			}
		}
	}
	if chosen == nil {
		fb := C.find_fallback_font(dpy, f.screen, f.font.pattern, C.FcChar32(r))
		switch {
		case fb == nil:
			chosen = f.font // no fallback; will draw a box
		case f.fallbackFonts[fb]:
			// Xft handed back a font already open here with one more
			// reference; keep a single one.
			C.XftFontClose(dpy, fb)
			chosen = fb
		default:
			if f.fallbackFonts == nil {
				f.fallbackFonts = make(map[*C.XftFont]bool)
			}
			f.fallbackFonts[fb] = true
			f.fallbackOrder = append(f.fallbackOrder, fb)
			chosen = fb
		}
	}
	f.fontByRune[r] = chosen
	return chosen
}

// runsFor segments s into contiguous runs, each using the same XftFont.
func (f *XftFont) runsFor(s string) []fontRun {
	if len(s) == 0 {
		return nil
	}
	var runs []fontRun
	pos := 0
	curStart := 0
	var curFont *C.XftFont
	first := true
	for pos < len(s) {
		ch, size := utf8.DecodeRuneInString(s[pos:])
		chFont := f.fontForRune(ch)
		if first || chFont != curFont {
			if !first {
				runs = append(runs, fontRun{curFont, s[curStart:pos]})
			}
			curFont = chFont
			curStart = pos
			first = false
		}
		pos += size
	}
	if curStart < len(s) {
		runs = append(runs, fontRun{curFont, s[curStart:]})
	}
	return runs
}

// DrawString draws a string on a drawable at the given baseline position.
// Implements platform.DrawableFont.
func (f *XftFont) DrawString(drawable platform.DrawableID, x, y int, s string, pixel uint64, r, g, b uint16) {
	f.drawStringXlib(xlib.Drawable(drawable), x, y, s, pixel, r, g, b)
}

// DrawStringAngle draws a string rotated by angleDeg degrees (clockwise on screen,
// matching Tk's canvas -angle convention). The (x, y) is the baseline start in
// screen coordinates, already adjusted for anchor and rotation.
func (f *XftFont) DrawStringAngle(drawable platform.DrawableID, x, y int, angleDeg float64, s string, pixel uint64, r, g, b uint16) {
	if len(s) == 0 {
		return
	}
	if angleDeg == 0 {
		f.DrawString(drawable, x, y, s, pixel, r, g, b)
		return
	}
	xlib.XftMu.Lock()
	rotFont := f.getOrCreateRotated(angleDeg)
	if rotFont == nil {
		xlib.XftMu.Unlock()
		f.DrawString(drawable, x, y, s, pixel, r, g, b)
		return
	}
	defer xlib.XftMu.Unlock()
	dpy := (*C.Display)(f.display.Ptr())
	draw := f.xftDraw(dpy, xlib.Drawable(drawable))
	if draw == nil {
		return
	}
	cs := C.CString(s)
	defer C.free(unsafe.Pointer(cs))
	color := f.xftColor(dpy, r, g, b)
	C.XftDrawStringUtf8(draw, &color, rotFont, C.int(x), C.int(y),
		(*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(s)))
}

// xftColor returns the XftColor for 16-bit r, g, b, allocated once per
// font. The caller holds xlib.XftMu.
func (f *XftFont) xftColor(dpy *C.Display, r, g, b uint16) C.XftColor {
	key := uint64(r)<<32 | uint64(g)<<16 | uint64(b)
	if c, ok := f.colors[key]; ok {
		return c
	}
	c := C.alloc_xft_color(dpy, f.visual, f.colormap, C.ushort(r), C.ushort(g), C.ushort(b))
	if f.colors == nil {
		f.colors = map[uint64]C.XftColor{}
	}
	f.colors[key] = c
	return c
}

// xftDraws keeps one XftDraw per display and retargets it with
// XftDrawChange, as Tk keeps fontPtr->ftDraw, instead of creating and
// destroying one (with its RENDER Picture) per string. Its drawable must
// be forgotten before that drawable goes away, or Xft later frees a
// Picture the server already dropped with its window (RenderBadPicture):
// the X11 backend calls ForgetDrawable from FreePixmap and ForgetAll from
// DestroyWindow and Close. Guarded by xlib.XftMu.
var xftDraws = map[*C.Display]*cachedDraw{}

type cachedDraw struct {
	draw     *C.XftDraw
	drawable xlib.Drawable
	visual   *C.Visual
	colormap C.Colormap
}

// xftDraw returns the display's XftDraw aimed at drawable. The caller
// holds xlib.XftMu.
func (f *XftFont) xftDraw(dpy *C.Display, drawable xlib.Drawable) *C.XftDraw {
	c := xftDraws[dpy]
	if c != nil && (c.visual != f.visual || c.colormap != f.colormap) {
		C.XftDrawDestroy(c.draw)
		delete(xftDraws, dpy)
		c = nil
	}
	if c == nil {
		draw := C.XftDrawCreate(dpy, C.Drawable(drawable), f.visual, f.colormap)
		if draw == nil {
			return nil
		}
		xftDraws[dpy] = &cachedDraw{draw, drawable, f.visual, f.colormap}
		return draw
	}
	if c.drawable != drawable {
		C.XftDrawChange(c.draw, C.Drawable(drawable))
		c.drawable = drawable
	}
	return c.draw
}

// ForgetDrawable drops the cached XftDraw of display d if it targets
// drawable; call it before freeing drawable.
func ForgetDrawable(d *xlib.Display, drawable xlib.Drawable) {
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
	dpy := (*C.Display)(d.Ptr())
	if c := xftDraws[dpy]; c != nil && c.drawable == drawable {
		C.XftDrawDestroy(c.draw)
		delete(xftDraws, dpy)
	}
}

// ForgetAll drops the cached XftDraw of display d; call it before
// destroying a window (which also destroys its descendants) or closing d.
func ForgetAll(d *xlib.Display) {
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
	dpy := (*C.Display)(d.Ptr())
	if c := xftDraws[dpy]; c != nil {
		C.XftDrawDestroy(c.draw)
		delete(xftDraws, dpy)
	}
}

// getOrCreateRotated returns (creating if needed) a rotated XFT font for the
// given angle. The caller holds xlib.XftMu.
func (f *XftFont) getOrCreateRotated(angleDeg float64) *C.XftFont {
	key := int64(angleDeg * 10)
	if rf, ok := f.rotatedVariants[key]; ok {
		return rf
	}
	if f.rotatedVariants == nil {
		f.rotatedVariants = make(map[int64]*C.XftFont)
	}
	rad := angleDeg * math.Pi / 180.0
	sinA := math.Sin(rad)
	cosA := math.Cos(rad)
	dpy := (*C.Display)(f.display.Ptr())
	rf := C.open_xft_font_rotated(dpy, f.screen, f.font, C.double(sinA), C.double(cosA))
	f.rotatedVariants[key] = rf // nil means "not available" — don't retry
	return rf
}

// drawStringXlib is the internal Xft implementation with font fallback.
func (f *XftFont) drawStringXlib(drawable xlib.Drawable, x, y int, s string, pixel uint64, r, g, b uint16) {
	if len(s) == 0 {
		return
	}
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()

	dpy := (*C.Display)(f.display.Ptr())
	draw := f.xftDraw(dpy, drawable)
	if draw == nil {
		return
	}

	color := f.xftColor(dpy, r, g, b)

	for _, run := range f.runsFor(s) {
		cs := C.CString(run.text)
		C.XftDrawStringUtf8(draw, &color, run.font, C.int(x), C.int(y),
			(*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(run.text)))
		var extents C.XGlyphInfo
		C.XftTextExtentsUtf8(dpy, run.font, (*C.FcChar8)(unsafe.Pointer(cs)), C.int(len(run.text)), &extents)
		C.free(unsafe.Pointer(cs))
		x += int(extents.xOff)
	}
}

// ListFamilies returns a sorted list of available font family names
// from fontconfig.
func ListFamilies() []string {
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
	pattern := C.FcPatternCreate()
	objectSet := C.FcObjectSetBuild_helper()
	fontSet := C.FcFontList(nil, pattern, objectSet)
	C.FcPatternDestroy(pattern)
	C.FcObjectSetDestroy(objectSet)

	if fontSet == nil {
		return nil
	}
	defer C.FcFontSetDestroy(fontSet)

	seen := make(map[string]bool)
	var families []string

	for i := C.int(0); i < fontSet.nfont; i++ {
		p := C.fc_fontset_get_font(fontSet, i)
		var family *C.FcChar8
		if C.fc_pattern_get_family(p, &family) == C.FcResultMatch {
			name := C.GoString((*C.char)(unsafe.Pointer(family)))
			if !seen[name] {
				seen[name] = true
				families = append(families, name)
			}
		}
	}

	// Sort families.
	sortStrings(families)
	return families
}

func sortStrings(s []string) {
	// Simple insertion sort — font family lists are typically <500 items.
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}

// Close releases font resources.
func (f *XftFont) Close() {
	xlib.XftMu.Lock()
	defer xlib.XftMu.Unlock()
	dpy := (*C.Display)(f.display.Ptr())
	C.no_unref_fonts(dpy)
	if f.font != nil {
		C.XftFontClose(dpy, f.font)
		f.font = nil
	}
	for fb := range f.fallbackFonts {
		C.XftFontClose(dpy, fb)
	}
	f.fallbackFonts = nil
	f.fallbackOrder = nil
	for _, c := range f.colors {
		C.XftColorFree(dpy, f.visual, f.colormap, &c)
	}
	f.colors = nil
	f.fontByRune = nil
	for _, rf := range f.rotatedVariants {
		if rf != nil {
			C.XftFontClose(dpy, rf)
		}
	}
	f.rotatedVariants = nil
}
