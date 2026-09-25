// Package clamtheme registers the "clam" TTK theme.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/clamtheme"
package clamtheme

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // ensure default theme init runs first
)

// Clam theme colors (from clamTheme.tcl).
const (
	frameColor           uint64 = 0xdcdad5
	darkColor            uint64 = 0xcfcdc8
	darkerColor          uint64 = 0xbab5ab
	darkestColor         uint64 = 0x9e9a91
	lighterColor         uint64 = 0xeeebe7
	lightColor           uint64 = 0xffffff
	disabledFg           uint64 = 0x999999
	altIndicator         uint64 = 0x5895bc
	disabledAltIndicator uint64 = 0xa0a0a0
)

func init() {
	// Ensure default theme is loaded first (it's our parent).
	// The blank import of defaulttheme in the demo does this.
	defaultTheme := ttk.CurrentTheme()

	theme := ttk.NewTheme("clam", defaultTheme)

	// Override root style.
	root := theme.GetStyle(".")
	root.Defaults["-background"] = frameColor
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-borderwidth"] = 2
	root.Defaults["-bordercolor"] = darkestColor
	root.Defaults["-darkcolor"] = darkColor
	root.Defaults["-lightcolor"] = lighterColor
	root.Defaults["-troughcolor"] = darkerColor

	root.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: lighterColor},
	}
	root.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
	}

	// Clam border element.
	theme.RegisterElement("border", newClamBorderFactory)

	// TLabel style — set padding explicitly (each theme is self-contained).
	tlabel := theme.GetStyle("TLabel")
	tlabel.Defaults["-padding"] = ttk.Padding{Left: 4, Top: 2, Right: 4, Bottom: 2}
	tlabel.Defaults["-relief"] = option.ReliefFlat
	tlabel.Defaults["-borderwidth"] = 0

	// TFrame style.
	tframe := theme.GetStyle("TFrame")
	tframe.Defaults["-relief"] = option.ReliefFlat
	tframe.Defaults["-borderwidth"] = 0

	// TButton style overrides (matches clamTheme.tcl).
	// Clam buttons keep raised relief even when disabled — no relief map.
	tbutton := theme.GetStyle("TButton")
	tbutton.Defaults["-anchor"] = option.AnchorCenter
	tbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Defaults["-borderwidth"] = 2
	tbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: lighterColor},
	}

	// Toolbutton style for clam (matches clamTheme.tcl).
	toolbutton := theme.GetStyle("Toolbutton")
	toolbutton.Defaults["-anchor"] = option.AnchorCenter
	toolbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	toolbutton.Defaults["-relief"] = option.ReliefFlat
	toolbutton.Defaults["-borderwidth"] = 2
	toolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: option.ReliefRaised},
	}
	toolbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: lighterColor},
	}

	// TMenubutton.Toolbutton for clam.
	tmbToolbutton := theme.GetStyle("TMenubutton.Toolbutton")
	tmbToolbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tmbToolbutton.Defaults["-relief"] = option.ReliefFlat
	tmbToolbutton.Defaults["-borderwidth"] = 2
	tmbToolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: option.ReliefRaised},
	}
	tmbToolbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: lighterColor},
	}

	// TSeparator styles.
	tsepH := theme.GetStyle("TSeparator.Horizontal")
	tsepH.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepH.Defaults["-relief"] = option.ReliefFlat
	tsepV := theme.GetStyle("TSeparator.Vertical")
	tsepV.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepV.Defaults["-relief"] = option.ReliefFlat

	// TTreeview style.
	ttreeview := theme.GetStyle("TTreeview")
	ttreeview.Defaults["-fieldbackground"] = uint64(0xffffff)
	ttreeview.Defaults["-selectbackground"] = uint64(0x4a6984)
	ttreeview.Defaults["-selectforeground"] = uint64(0xffffff)
	ttreeview.Defaults["-bordercolor"] = darkestColor

	// Treeview.Heading style.
	tvHeading := theme.GetStyle("Treeview.Heading")
	tvHeading.Defaults["-background"] = frameColor
	tvHeading.Defaults["-foreground"] = uint64(0x000000)
	tvHeading.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: darkerColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: uint64(0xececec)},
	}

	// TScrollbar styles — troughcolor matches Tcl clam's $colors(-darker).
	vScrollbar := theme.GetStyle("Vertical.TScrollbar")
	vScrollbar.Defaults["-troughcolor"] = darkerColor

	hScrollbar := theme.GetStyle("Horizontal.TScrollbar")
	hScrollbar.Defaults["-troughcolor"] = darkerColor

	// TSpinbox style.
	tspinbox := theme.GetStyle("TSpinbox")
	tspinbox.Defaults["-background"] = frameColor
	tspinbox.Defaults["-foreground"] = uint64(0x000000)
	tspinbox.Defaults["-fieldbackground"] = uint64(0xffffff)

	// TEntry style.
	tentry := theme.GetStyle("TEntry")
	tentry.Defaults["-background"] = frameColor
	tentry.Defaults["-foreground"] = uint64(0x000000)
	tentry.Defaults["-fieldbackground"] = uint64(0xffffff)
	tentry.Defaults["-selectbackground"] = uint64(0x4a6984)
	tentry.Defaults["-selectforeground"] = uint64(0xffffff)
	tentry.Defaults["-insertwidth"] = 1
	tentry.Defaults["-padding"] = ttk.Padding{Left: 1, Top: 1, Right: 1, Bottom: 1}
	tentry.Defaults["-insertcolor"] = uint64(0x000000)

	// TEntry layout: background → highlight → border → padding
	theme.RegisterLayout("TEntry",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	// TSizegrip style.
	tsizegrip := theme.GetStyle("TSizegrip")
	tsizegrip.Defaults["-background"] = frameColor

	// Progressbar styles — troughcolor matches Tcl clam's $colors(-darker).
	hProgress := theme.GetStyle("Horizontal.TProgressbar")
	hProgress.Defaults["-troughcolor"] = darkerColor
	hProgress.Defaults["-barcolor"] = uint64(0x4a6984)

	vProgress := theme.GetStyle("Vertical.TProgressbar")
	vProgress.Defaults["-troughcolor"] = darkerColor
	vProgress.Defaults["-barcolor"] = uint64(0x4a6984)

	// TCheckbutton style — clam-style flat indicators.
	// Matches Tcl clam: white fill with light blue (#5895bc) alternate.
	tcheckbutton := theme.GetStyle("TCheckbutton")
	tcheckbutton.Defaults["-padding"] = "1.5p"
	tcheckbutton.Defaults["-indicatormargin"] = "0.75p 0.75p 3p 0.75p"
	tcheckbutton.Defaults["-upperbordercolor"] = darkestColor
	tcheckbutton.Defaults["-lowerbordercolor"] = darkColor
	tcheckbutton.Defaults["-indicatorbackground"] = lightColor
	tcheckbutton.Defaults["-indicatorforeground"] = uint64(0x000000)
	tcheckbutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: disabledAltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
	}
	tcheckbutton.Maps["-indicatorforeground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
	}

	// TRadiobutton style — clam-style flat indicators.
	tradiobutton := theme.GetStyle("TRadiobutton")
	tradiobutton.Defaults["-padding"] = "1.5p"
	tradiobutton.Defaults["-indicatormargin"] = "0.75p 0.75p 3p 0.75p"
	tradiobutton.Defaults["-upperbordercolor"] = darkestColor
	tradiobutton.Defaults["-lowerbordercolor"] = darkColor
	tradiobutton.Defaults["-indicatorbackground"] = lightColor
	tradiobutton.Defaults["-indicatorforeground"] = uint64(0x000000)
	tradiobutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: disabledAltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
	}
	tradiobutton.Maps["-indicatorforeground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
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

func (e *ClamBorderElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
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

func drawRectOutline(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, x, y, w, h int) {
	d.DrawLine(drawable, gc, x, y, x+w-1, y)         // top
	d.DrawLine(drawable, gc, x, y+h-1, x+w-1, y+h-1) // bottom
	d.DrawLine(drawable, gc, x, y, x, y+h-1)         // left
	d.DrawLine(drawable, gc, x+w-1, y, x+w-1, y+h-1) // right
}
