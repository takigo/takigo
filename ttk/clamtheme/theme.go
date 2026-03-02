// Package clamtheme registers the "clam" TTK theme.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/clamtheme"
package clamtheme

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // ensure default theme init runs first
)

// Clam theme colors.
const (
	frameColor uint64 = 0xdcdad5
	darkColor  uint64 = 0xcfcdc8
	darkerColor uint64 = 0xbab5ab
	darkestColor uint64 = 0x9e9a91
	lightColor uint64 = 0xffffff
)

func init() {
	// Ensure default theme is loaded first (it's our parent).
	// The blank import of defaulttheme in the demo does this.
	defaultTheme := ttk.CurrentTheme()

	theme := ttk.NewTheme("clam", defaultTheme)

	// Override root background.
	root := theme.GetStyle(".")
	root.Defaults["-background"] = frameColor
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-borderwidth"] = 2

	// Clam border element.
	theme.RegisterElement("border", newClamBorderFactory)

	// TButton style overrides.
	tbutton := theme.GetStyle("TButton")
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}
	tbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
	}

	ttk.RegisterTheme(theme)
}

// ClamBorderElement draws a custom 2px border in the clam style.
type ClamBorderElement struct {
	ctx *ttk.DrawContext
}

func newClamBorderFactory(ctx *ttk.DrawContext) ttk.Element {
	return &ClamBorderElement{ctx: ctx}
}

func (e *ClamBorderElement) Size(state ttk.State) (int, int, ttk.Padding) {
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	return 0, 0, ttk.UniformPadding(bw)
}

func (e *ClamBorderElement) Draw(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, box ttk.Box, state ttk.State) {
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	relief := ttk.LookupRelief(e.ctx.Style, "-relief", state, option.ReliefFlat)

	if bw <= 0 || relief == option.ReliefFlat {
		return
	}

	x, y, w, h := box.X, box.Y, box.Width, box.Height

	if bw >= 2 {
		// Outer border: darkest color.
		d.SetForeground(gc, darkestColor)
		drawRectOutline(d, drawable, gc, x, y, w, h)

		// Inner border: different for each edge.
		var topLeft, bottomRight uint64
		switch relief {
		case option.ReliefRaised:
			topLeft = lightColor
			bottomRight = darkColor
		case option.ReliefSunken:
			topLeft = darkColor
			bottomRight = lightColor
		default:
			// For groove/ridge, use standard draw.
			bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
			border := draw.NewBorderFromPixel(bg)
			draw.Draw3DRectangle(d, drawable, gc, border, x, y, w, h, bw, relief)
			return
		}

		d.SetForeground(gc, topLeft)
		// Top inner.
		d.DrawLine(drawable, gc, x+1, y+1, x+w-2, y+1)
		// Left inner.
		d.DrawLine(drawable, gc, x+1, y+1, x+1, y+h-2)

		d.SetForeground(gc, bottomRight)
		// Bottom inner.
		d.DrawLine(drawable, gc, x+1, y+h-2, x+w-2, y+h-2)
		// Right inner.
		d.DrawLine(drawable, gc, x+w-2, y+1, x+w-2, y+h-2)
	} else {
		// Single pixel border.
		d.SetForeground(gc, darkestColor)
		drawRectOutline(d, drawable, gc, x, y, w, h)
	}
}

func drawRectOutline(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC, x, y, w, h int) {
	d.DrawLine(drawable, gc, x, y, x+w-1, y)         // top
	d.DrawLine(drawable, gc, x, y+h-1, x+w-1, y+h-1) // bottom
	d.DrawLine(drawable, gc, x, y, x, y+h-1)         // left
	d.DrawLine(drawable, gc, x+w-1, y, x+w-1, y+h-1) // right
}
