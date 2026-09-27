package canvas

import (
	"io"
	"strings"

	"github.com/msorc/takigo/option"
)

// PostscriptOption configures a Canvas.Postscript call.
//
// Mirrors the `$canvas postscript ...` subcommand options in
// tk/generic/tkCanvPs.c:92-126.
type PostscriptOption func(*psConfig)

type psConfig struct {
	File      string            // -file <path>
	Channel   io.Writer         // -channel equivalent (Go idiom: pass an io.Writer)
	ColorMap  map[string]string // -colormap (color name → PS string)
	FontMap   map[string][2]string
	ColorMode string // "monochrome" | "gray" | "color"
	X, Y      int
	Width     int
	Height    int
	PageX     float64
	PageY     float64
	PageW     float64
	PageH     float64
	Anchor    option.Anchor
	Rotate    bool
	Prolog    bool
	Title     string
}

// PSFile sets the output path (Tk: -file). When set, the PostScript is
// written to that file and the return value of Canvas.Postscript is "".
func PSFile(path string) PostscriptOption {
	return func(c *psConfig) { c.File = path }
}

// PSWriter sets the output writer. Takes precedence over -file.
func PSWriter(w io.Writer) PostscriptOption {
	return func(c *psConfig) { c.Channel = w }
}

// PSColorMode sets -colormode: "monochrome", "gray", or "color".
func PSColorMode(mode string) PostscriptOption {
	return func(c *psConfig) { c.ColorMode = strings.ToLower(mode) }
}

// PSColormapVar sets the color map (-colormap). Keys are "#RRGGBB" hex
// strings; values are PS fragments to emit instead of setrgbcolor.
func PSColormapVar(m map[string]string) PostscriptOption {
	return func(c *psConfig) { c.ColorMap = m }
}

// PSFontMapVar sets the font map (-fontmap). Keys are font family names;
// values are {PS-font-name, size-as-string} pairs.
func PSFontMapVar(m map[string][2]string) PostscriptOption {
	return func(c *psConfig) { c.FontMap = m }
}

// PSRegion sets the canvas pixel rectangle to print. Defaults to the
// visible canvas.
func PSRegion(x, y, w, h int) PostscriptOption {
	return func(c *psConfig) { c.X, c.Y, c.Width, c.Height = x, y, w, h }
}

// PSPageX sets -pagex (PostScript points; default 72*4.25).
func PSPageX(x float64) PostscriptOption { return func(c *psConfig) { c.PageX = x } }

// PSPageY sets -pagey (PostScript points; default 72*5.5).
func PSPageY(y float64) PostscriptOption { return func(c *psConfig) { c.PageY = y } }

// PSPageWidth sets -pagewidth.
func PSPageWidth(w float64) PostscriptOption { return func(c *psConfig) { c.PageW = w } }

// PSPageHeight sets -pageheight.
func PSPageHeight(h float64) PostscriptOption { return func(c *psConfig) { c.PageH = h } }

// PSPageAnchor sets -pageanchor.
func PSPageAnchor(a option.Anchor) PostscriptOption { return func(c *psConfig) { c.Anchor = a } }

// PSRotate enables landscape output (-rotate).
func PSRotate(b bool) PostscriptOption { return func(c *psConfig) { c.Rotate = b } }

// PSProlog toggles the standard Tk preamble (-prolog, default true).
func PSProlog(b bool) PostscriptOption { return func(c *psConfig) { c.Prolog = b } }

// PSTitle sets the %%Title header field.
func PSTitle(t string) PostscriptOption { return func(c *psConfig) { c.Title = t } }

// noPrologFlag is set when -prolog is omitted (defaults to true).
func (c *psConfig) prologSet() bool { return c.Prolog }

// colorLevel converts the -colormode string to Tk's numeric level.
func (c *psConfig) colorLevel() int {
	switch c.ColorMode {
	case "monochrome", "mono":
		return 0
	case "gray", "grey":
		return 1
	default:
		return 2 // "color" or unset
	}
}
