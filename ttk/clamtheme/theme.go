// Package clamtheme registers the "clam" TTK theme.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/clamtheme"
package clamtheme

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // ensure default theme init runs first
)

// Palette is the set of colours a clam-style theme is drawn with, as
// 0xRRGGBB values. Light is clam's own (clamTheme.tcl).
type Palette struct {
	Frame, Dark, Darker, Darkest uint64 // window background and its shades
	Lighter, Light               uint64 // highlights
	Foreground, DisabledFg       uint64
	Field                        uint64 // entry, spinbox and tree background
	Select, SelectForeground     uint64 // selection and progress bars
	HeadingHover                 uint64
	AltIndicator                 uint64
	DisabledAltIndicator         uint64
}

// Light is the palette of the "clam" theme.
var Light = Palette{
	Frame:                0xdcdad5,
	Dark:                 0xcfcdc8,
	Darker:               0xbab5ab,
	Darkest:              0x9e9a91,
	Lighter:              0xeeebe7,
	Light:                0xffffff,
	Foreground:           0x000000,
	DisabledFg:           0x999999,
	Field:                0xffffff,
	Select:               0x4a6984,
	SelectForeground:     0xffffff,
	HeadingHover:         0xececec,
	AltIndicator:         0x5895bc,
	DisabledAltIndicator: 0xa0a0a0,
}

func init() {
	ttk.RegisterTheme(New("clam", Light))
}

// New builds a theme with clam's layouts and elements in the given
// palette. Register the result with ttk.RegisterTheme.
func New(name string, p Palette) *ttk.Theme {
	// The default theme is the parent; this package's blank import of
	// defaulttheme registers it first.
	defaultTheme := ttk.LookupTheme("default")

	theme := ttk.NewTheme(name, defaultTheme)

	// Override root style.
	root := theme.GetStyle(".")
	root.Defaults["-background"] = p.Frame
	root.Defaults["-foreground"] = p.Foreground
	root.Defaults["-borderwidth"] = 2
	root.Defaults["-bordercolor"] = p.Darkest
	root.Defaults["-darkcolor"] = p.Dark
	root.Defaults["-lightcolor"] = p.Lighter
	root.Defaults["-troughcolor"] = p.Darker

	root.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: p.Lighter},
	}
	root.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.DisabledFg},
	}

	// Clam border element.
	theme.RegisterElement("border", borderFactory(p))

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
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Darker},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: p.Lighter},
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
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Darker},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: p.Lighter},
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
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Darker},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: p.Lighter},
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
	ttreeview.Defaults["-fieldbackground"] = p.Field
	ttreeview.Defaults["-selectbackground"] = p.Select
	ttreeview.Defaults["-selectforeground"] = p.SelectForeground
	ttreeview.Defaults["-bordercolor"] = p.Darkest

	// Treeview.Heading style.
	tvHeading := theme.GetStyle("Treeview.Heading")
	tvHeading.Defaults["-background"] = p.Frame
	tvHeading.Defaults["-foreground"] = p.Foreground
	tvHeading.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Darker},
		{Spec: ttk.StateSpec{OnBits: ttk.StateHover}, Value: p.HeadingHover},
	}

	// TScrollbar styles — troughcolor matches Tcl clam's $colors(-darker).
	vScrollbar := theme.GetStyle("Vertical.TScrollbar")
	vScrollbar.Defaults["-troughcolor"] = p.Darker

	hScrollbar := theme.GetStyle("Horizontal.TScrollbar")
	hScrollbar.Defaults["-troughcolor"] = p.Darker

	// TSpinbox style.
	tspinbox := theme.GetStyle("TSpinbox")
	tspinbox.Defaults["-background"] = p.Frame
	tspinbox.Defaults["-foreground"] = p.Foreground
	tspinbox.Defaults["-fieldbackground"] = p.Field

	// TEntry style.
	tentry := theme.GetStyle("TEntry")
	tentry.Defaults["-background"] = p.Frame
	tentry.Defaults["-foreground"] = p.Foreground
	tentry.Defaults["-fieldbackground"] = p.Field
	tentry.Defaults["-selectbackground"] = p.Select
	tentry.Defaults["-selectforeground"] = p.SelectForeground
	tentry.Defaults["-insertwidth"] = 1
	tentry.Defaults["-padding"] = ttk.Padding{Left: 1, Top: 1, Right: 1, Bottom: 1}
	tentry.Defaults["-insertcolor"] = p.Foreground

	// TEntry layout: background → highlight → border → padding
	theme.RegisterLayout("TEntry",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	// TSizegrip style.
	tsizegrip := theme.GetStyle("TSizegrip")
	tsizegrip.Defaults["-background"] = p.Frame

	// Progressbar styles — troughcolor matches Tcl clam's $colors(-darker).
	hProgress := theme.GetStyle("Horizontal.TProgressbar")
	hProgress.Defaults["-troughcolor"] = p.Darker
	hProgress.Defaults["-barcolor"] = p.Select

	vProgress := theme.GetStyle("Vertical.TProgressbar")
	vProgress.Defaults["-troughcolor"] = p.Darker
	vProgress.Defaults["-barcolor"] = p.Select

	// TCheckbutton style — clam-style flat indicators.
	// Matches Tcl clam: white fill with light blue (#5895bc) alternate.
	tcheckbutton := theme.GetStyle("TCheckbutton")
	tcheckbutton.Defaults["-padding"] = screenunit.Pt(1.5)
	tcheckbutton.Defaults["-indicatormargin"] = "0.75p 0.75p 3p 0.75p"
	tcheckbutton.Defaults["-upperbordercolor"] = p.Darkest
	tcheckbutton.Defaults["-lowerbordercolor"] = p.Dark
	tcheckbutton.Defaults["-indicatorbackground"] = p.Light
	tcheckbutton.Defaults["-indicatorforeground"] = p.Foreground
	tcheckbutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: p.DisabledAltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: p.AltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
	}
	tcheckbutton.Maps["-indicatorforeground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.DisabledFg},
	}

	// TRadiobutton style — clam-style flat indicators.
	tradiobutton := theme.GetStyle("TRadiobutton")
	tradiobutton.Defaults["-padding"] = screenunit.Pt(1.5)
	tradiobutton.Defaults["-indicatormargin"] = "0.75p 0.75p 3p 0.75p"
	tradiobutton.Defaults["-upperbordercolor"] = p.Darkest
	tradiobutton.Defaults["-lowerbordercolor"] = p.Dark
	tradiobutton.Defaults["-indicatorbackground"] = p.Light
	tradiobutton.Defaults["-indicatorforeground"] = p.Foreground
	tradiobutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: p.Frame},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate | ttk.StateDisabled}, Value: p.DisabledAltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: p.AltIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.Frame},
	}
	tradiobutton.Maps["-indicatorforeground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: p.DisabledFg},
	}

	return theme
}

// ClamBorderElement draws a custom 2px border in the clam style.
type ClamBorderElement struct {
	ctx *ttk.DrawContext
	p   Palette
}

func borderFactory(p Palette) ttk.ElementFactory {
	return func(ctx *ttk.DrawContext) ttk.Element {
		return &ClamBorderElement{ctx: ctx, p: p}
	}
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
		d.SetForeground(gc, e.p.Darkest)
		drawRectOutline(d, drawable, gc, x, y, w, h)

		// Inner border: different for each edge.
		var topLeft, bottomRight uint64
		switch relief {
		case option.ReliefRaised:
			topLeft = e.p.Light
			bottomRight = e.p.Dark
		case option.ReliefSunken:
			topLeft = e.p.Dark
			bottomRight = e.p.Light
		default:
			// For groove/ridge, use standard draw.
			bg := ttk.LookupColor(e.ctx.Style, "-background", state, e.p.Frame)
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
		d.SetForeground(gc, e.p.Darkest)
		drawRectOutline(d, drawable, gc, x, y, w, h)
	}
}

func drawRectOutline(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, x, y, w, h int) {
	d.DrawLine(drawable, gc, x, y, x+w-1, y)         // top
	d.DrawLine(drawable, gc, x, y+h-1, x+w-1, y+h-1) // bottom
	d.DrawLine(drawable, gc, x, y, x, y+h-1)         // left
	d.DrawLine(drawable, gc, x+w-1, y, x+w-1, y+h-1) // right
}
