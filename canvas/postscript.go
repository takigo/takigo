package canvas

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
)

// PSContext carries the state of a single canvas PostScript generation.
//
// Mirrors TkPostscriptInfo from tk/generic/tkCanvPs.c:43-85. Items call
// methods on this context to emit PostScript fragments; the context owns the
// output buffer, font table, and optional color/font mappings.
type PSContext struct {
	buf       strings.Builder
	itemStart int // offset of the current item's first byte in buf

	// Page geometry (PostScript points; 72 pt/in).
	PageX, PageY float64
	PageWidth    float64
	PageHeight   float64
	Scale        float64
	Anchor       option.Anchor
	Rotate       bool
	Prolog       bool

	// Source rectangle in canvas pixel coordinates.
	X, Y, X2, Y2  int
	Width, Height int

	// ColorLevel: 0=monochrome, 1=grayscale, 2=color (set from -colormode).
	ColorLevel int

	// Prepass=true: items only register font names, output is discarded.
	Prepass bool

	// Font names collected during pre-pass.
	fonts map[string]bool

	// Optional -colormap: color name → PS-string to emit instead of setrgbcolor.
	Colormap map[string]string

	// Optional -fontmap: font name → {PS-name, size-points}.
	Fontmap map[string][2]string

	// Title used in %%Title header.
	Title string
}

// NewPSContext creates a PSContext with sensible defaults mirroring Tk's
// TkCanvPostscriptObjCmd initial state.
func NewPSContext() *PSContext {
	return &PSContext{
		PageX:      72 * 4.25, // default Tk page position
		PageY:      72 * 5.5,
		PageWidth:  0, // 0 = unset; computed from width if zero
		PageHeight: 0,
		Scale:      1.0,
		Anchor:     option.AnchorCenter,
		Rotate:     false,
		Prolog:     true,
		ColorLevel: 2, // default "color"
		fonts:      make(map[string]bool),
	}
}

// String returns the accumulated PostScript output.
func (ps *PSContext) String() string { return ps.buf.String() }

// Reset empties the buffer (used between pre-pass and real pass).
func (ps *PSContext) Reset() { ps.buf.Reset() }

// ResetItemBuf marks the start of an item's output. After item.Postscript()
// runs, TakeItemBuf returns just the bytes written since this call so the
// caller can wrap them in gsave/grestore.
func (ps *PSContext) ResetItemBuf() { ps.itemStart = ps.buf.Len() }

// TakeItemBuf returns the bytes written since the last ResetItemBuf and
// resets itemStart.
func (ps *PSContext) TakeItemBuf() string {
	end := ps.buf.Len()
	if ps.itemStart >= end {
		ps.itemStart = end
		return ""
	}
	out := ps.buf.String()[ps.itemStart:end]
	ps.itemStart = end
	return out
}

// Fonts returns the set of font names registered during the pre-pass.
func (ps *PSContext) Fonts() []string {
	out := make([]string, 0, len(ps.fonts))
	for n := range ps.fonts {
		out = append(out, n)
	}
	return out
}

// write appends raw text to the buffer. Used by per-item emitters that
// build their own PS directly.
func (ps *PSContext) write(s string) { ps.buf.WriteString(s) }

// psPrintf formats a PostScript fragment with %g-style numbers.
func (ps *PSContext) psPrintf(format string, args ...float64) {
	converted := make([]any, len(args))
	for i, a := range args {
		converted[i] = a
	}
	ps.buf.WriteString(fmt.Sprintf(format, converted...))
}

// writef appends a sprintf-style format string.
func (ps *PSContext) writef(format string, args ...any) {
	ps.buf.WriteString(fmt.Sprintf(format, args...))
}

// registerFont notes that a font is in use; called from the Font emitter
// during pre-pass so the header can list it in %%DocumentNeededResources.
func (ps *PSContext) registerFont(name string) {
	ps.fonts[name] = true
}

// ---- shared PostScript emitters ----

// PsY converts a canvas-y coordinate to a PostScript y-coordinate (y axis
// flips). Mirrors tk/generic/tkCanvPs.c:1026-1029 (Tk_PostscriptY).
func (ps *PSContext) PsY(y int) float64 {
	return float64(ps.Y2 - y)
}

// Path emits "moveto" + N×"lineto" for the given canvas-coordinate points.
// Mirrors tk/generic/tkCanvPs.c:1062-1075 (Tk_PostscriptPath).
func (ps *PSContext) Path(pts []float64) {
	if len(pts) < 2 {
		return
	}
	ps.writef("%.15g %.15g moveto\n", pts[0], ps.PsY(int(pts[1])))
	for i := 2; i+1 < len(pts); i += 2 {
		ps.writef("%.15g %.15g lineto\n", pts[i], ps.PsY(int(pts[i+1])))
	}
}

// Color emits "R G B setrgbcolor AdjustColor\n" — or a gray/halftone
// equivalent when ColorLevel < 2 — and looks up the color in Colormap if
// one is supplied. Mirrors tk/generic/tkCanvPs.c:690-737 (Tk_PostscriptColor).
func (ps *PSContext) Color(c *color.ColorRef) {
	if ps.Prepass {
		return
	}
	if c == nil {
		return
	}
	if name, ok := lookupColormap(ps.Colormap, c); ok {
		ps.writef("%s\n", name)
		return
	}
	r, g, b := rgbFractions(c)
	switch ps.ColorLevel {
	case 0: // monochrome
		gray := (r + g + b) / 3
		ps.writef("%.3f setgray\n", gray)
	case 1: // grayscale
		// Tk uses Rec. 601 luma: 0.299*R + 0.587*G + 0.114*B.
		gray := 0.299*r + 0.587*g + 0.114*b
		ps.writef("%.3f setgray\n", gray)
	default: // color
		ps.writef("%.3f %.3f %.3f setrgbcolor AdjustColor\n", r, g, b)
	}
}

func lookupColormap(m map[string]string, c *color.ColorRef) (string, bool) {
	if len(m) == 0 || c == nil {
		return "", false
	}
	// Tk's colormap is keyed by color name; we accept a hex "#RRGGBB" key as
	// a convenience since takigo ColorRefs don't carry the original name.
	key := fmt.Sprintf("#%02x%02x%02x", c.Red>>8, c.Green>>8, c.Blue>>8)
	if v, ok := m[key]; ok {
		return v, true
	}
	return "", false
}

func rgbFractions(c *color.ColorRef) (r, g, b float64) {
	return float64(c.Red>>8) / 255.0,
		float64(c.Green>>8) / 255.0,
		float64(c.Blue>>8) / 255.0
}

// Font emits a PostScript "findfont/scalefont/setfont" sequence. During
// pre-pass it only registers the font name in the font table. Mirrors
// tk/generic/tkCanvPs.c:763-825 (Tk_PostscriptFont).
func (ps *PSContext) Font(f font.Font) {
	if f == nil {
		return
	}
	if ps.Fontmap != nil {
		// Honour -fontmap override if present.
		if v, ok := ps.Fontmap[f.Attrs().Family]; ok {
			psName, size := v[0], psParseFloat(v[1])
			if psName != "" && size > 0 {
				encode := " ISOEncode"
				if strings.HasPrefix(strings.ToLower(psName), "symbol") {
					encode = ""
				}
				ps.registerFont(psName)
				if ps.Prepass {
					return
				}
				ps.writef("/%s findfont %d scalefont%s setfont\n",
					psName, int(size+0.5), encode)
				return
			}
		}
	}
	name, _ := psFontName(f)
	ps.registerFont(name)
	if ps.Prepass {
		return
	}
	ps.write(psFontEmit(f))
}

// Bitmap emits a 1-bit raster as PostScript <hex> image data for use with
// `imagemask`. Mirrors tk/generic/tkCanvImg.c:754-759 + tk/generic/tkImgBmap.c.
func (ps *PSContext) Bitmap(bits []byte, w, h int) {
	if ps.Prepass {
		return
	}
	if w == 0 || h == 0 {
		return
	}
	// Tk packs bits MSB-first per byte in the hex data; takigo's XBMData is
	// already MSB-first (LSB-first on wire but stored MSB-first per scanline).
	// We just hex-encode the bytes verbatim.
	hex := make([]byte, 0, len(bits)*2)
	for _, b := range bits {
		hex = append(hex, hexDigit[(b>>4)&0xF], hexDigit[b&0xF])
	}
	ps.writef("%d %d true [", w, h)
	ps.writef("%d %d 0 0 0 0", w, h)
	ps.writef("] {<%s>} imagemask\n", string(hex))
}

var hexDigit = []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b', 'c', 'd', 'e', 'f'}

// Stipple fills the current clipping path with the given 1-bit stipple
// pattern. Mirrors tk/generic/tkCanvPs.c + tk/generic/tkImgBmap.c.
func (ps *PSContext) Stipple(bits []byte, w, h int) {
	if ps.Prepass {
		return
	}
	if w == 0 || h == 0 {
		return
	}
	hex := make([]byte, 0, len(bits)*2)
	for _, b := range bits {
		hex = append(hex, hexDigit[(b>>4)&0xF], hexDigit[b&0xF])
	}
	ps.writef("gsave StrokeClip %d %d true [", w, h)
	ps.writef("%d %d 0 0 0 0", w, h)
	ps.writef("] {<%s>} imagemask grestore newpath\n", string(hex))
}

// Outline emits the PostScript fragment that sets line width, dash pattern,
// color, and finally strokes (or fills with stipple + StrokeClip).
// Mirrors tk/generic/tkCanvUtil.c:1399-1504 (Tk_CanvasPsOutline).
//
// If stipple != nil, the current path is filled via StrokeClip + TkStippleFill
// instead of stroked.
func (ps *PSContext) Outline(width int, dash []int, offset int, c *color.ColorRef, stipple []byte) {
	if ps.Prepass {
		return
	}
	ps.writef("%.15g setlinewidth\n", float64(width))

	ps.write("[")
	if len(dash) > 0 {
		// Tk doubles the pattern if it has odd length (so the dash list is
		// always even in setdash).
		d := dash
		if len(d)%2 == 1 {
			d = append(d, dash...)
		}
		ps.write(strconv.Itoa(d[0]))
		for _, v := range d[1:] {
			ps.writef(" %d", v)
		}
		ps.writef("] %d setdash\n", offset)
	} else {
		ps.write("] 0 setdash\n")
	}

	ps.Color(c)

	if len(stipple) > 0 {
		ps.Stipple(stipple /*w*/, 0 /*h*/, 0)
		// Stroked-from-clipped path: replace currentpath with the clip and
		// emit nothing further; the calling item already drew the fill.
		ps.write("StrokeClip\n")
	} else {
		ps.write("stroke\n")
	}
}

// EmitHeader writes the EPS header: %%BoundingBox, font list, preamble, and
// the page-setup (translate/rotate/scale/clip-path). Mirrors the body of
// TkCanvPostscriptObjCmd from tk/generic/tkCanvPs.c:425-546.
func (ps *PSContext) EmitHeader() {
	if ps.Prolog {
		ps.write("%!PS-Adobe-3.0 EPSF-3.0\n")
		ps.write("%%Creator: takigo Canvas Widget\n")
		ps.writef("%%%%Title: %s\n", ps.Title)
		ps.write("%%CreationDate: takigo\n")
		// BoundingBox computed below once we know clip.
		ps.writef("%%%%BoundingBox: 0 0 %d %d\n", ps.Width, ps.Height)
		ps.write("%%Pages: 1\n")
		ps.write("%%DocumentData: Clean7Bit\n")
		if ps.Rotate {
			ps.write("%%Orientation: Landscape\n")
		} else {
			ps.write("%%Orientation: Portrait\n")
		}
		for _, name := range sortedFonts(ps.fonts) {
			ps.writef("%%%%DocumentNeededResources: font %s\n", name)
		}
		ps.write("%%EndComments\n\n")

		// Prolog (Tk's preamble from mkpsenc.tcl).
		ps.write(tkPsPreamble)
		ps.write("\n")

		// Document setup.
		ps.writef("%%%%BeginSetup\n/CL %d def\n", ps.ColorLevel)
		for _, name := range sortedFonts(ps.fonts) {
			ps.writef("%%%%IncludeResource: font %s\n", name)
		}
		ps.write("%%EndSetup\n\n")

		// Page setup.
		ps.write("%%Page: 1 1\nsave\n")
		ps.writef("%.1f %.1f translate\n", ps.PageX, ps.PageY)
		if ps.Rotate {
			ps.write("90 rotate\n")
		}
		ps.writef("%.4g %.4g scale\n", ps.Scale, ps.Scale)
		// Translate so the canvas origin matches the page anchor.
		ps.writef("%d %d translate\n", -ps.X, 0)

		// Clipping rectangle.
		ps.writef("%d %.15g moveto %d %.15g lineto %d %.15g lineto %d %.15g lineto closepath clip newpath\n",
			ps.X, ps.PsY(ps.Y),
			ps.X2, ps.PsY(ps.Y),
			ps.X2, ps.PsY(ps.Y2),
			ps.X, ps.PsY(ps.Y2),
		)
	}
}

func sortedFonts(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for n := range m {
		out = append(out, n)
	}
	// stable alphabetical
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}

// EmitTrailer writes the page-trailer + %%EOF.
func (ps *PSContext) EmitTrailer() {
	if ps.Prolog {
		ps.write("restore showpage\n\n")
		ps.write("%%Trailer\n")
		ps.write("end\n")
		ps.write("%%EOF\n")
	}
}

func psParseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
