// Package classictheme registers the "classic" TTK theme.
// Implements the classic Motif-like Tk look with custom elements and layouts.
// Import with blank identifier to auto-register: _ "github.com/msorc/takigo/ttk/classictheme"
package classictheme

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/defaulttheme" // ensure default theme init runs first
)

// Classic theme colors.
const (
	frameColor   uint64 = 0xd9d9d9
	activeBg     uint64 = 0xececec
	disabledFg   uint64 = 0xa3a3a3
	selectBg     uint64 = 0xc3c3c3
	troughBg     uint64 = 0xb3b3b3
	indicator    uint64 = 0xb03060
	altIndicator uint64 = 0xb05e5e
)

// --- highlightElement ---
// Draws a Motif-style focus highlight border.
// When focused: draws highlightcolor ring; otherwise draws background color (invisible).
type highlightElement struct{ ctx *ttk.DrawContext }

func newHighlightElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &highlightElement{ctx: ctx}
}

func (e *highlightElement) Size(state ttk.State) (int, int, ttk.Padding) {
	ht := ttk.LookupInt(e.ctx.Style, "-highlightthickness", state, 0)
	return 0, 0, ttk.UniformPadding(ht)
}

func (e *highlightElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	ht := ttk.LookupInt(e.ctx.Style, "-highlightthickness", state, 0)
	if ht <= 0 {
		return
	}
	var color uint64
	if state&ttk.StateFocus != 0 {
		color = ttk.LookupColor(e.ctx.Style, "-highlightcolor", state, 0x000000)
	} else {
		color = ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	}
	d.SetForeground(gc, color)
	for i := range ht {
		d.DrawRectangle(drawable, gc, box.X+i, box.Y+i,
			uint(box.Width-1-2*i), uint(box.Height-1-2*i))
	}
}

// --- buttonBorderElement ---
// Motif-style button border: same as BorderElement but registered as "Button.border".
// Supports a simple version of the default ring by checking -default style option.
type buttonBorderElement struct{ ctx *ttk.DrawContext }

func newButtonBorderElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &buttonBorderElement{ctx: ctx}
}

func (e *buttonBorderElement) Size(state ttk.State) (int, int, ttk.Padding) {
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	return 0, 0, ttk.UniformPadding(bw)
}

func (e *buttonBorderElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	relief := ttk.LookupRelief(e.ctx.Style, "-relief", state, option.ReliefFlat)
	if bw <= 0 || relief == option.ReliefFlat {
		return
	}
	border := draw.NewBorderFromPixel(bg)
	draw.Draw3DRectangle(d, drawable, gc, border, box.X, box.Y, box.Width, box.Height, bw, relief)
}

// --- squareIndicatorElement ---
// Checkbutton.indicator: 3D raised/sunken square (Motif-style).
// Relief is read from -indicatorrelief; style can state-map it to sunken when selected.
type squareIndicatorElement struct{ ctx *ttk.DrawContext }

func newSquareIndicatorElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &squareIndicatorElement{ctx: ctx}
}

func (e *squareIndicatorElement) Size(state ttk.State) (int, int, ttk.Padding) {
	sz := ttk.LookupInt(e.ctx.Style, "-indicatorsize", state, 12)
	// margin: left=0, top=2, right=4, bottom=2 (matches C indicatormargin "0 2 4 2")
	return sz + 4, sz + 4, ttk.Padding{}
}

func (e *squareIndicatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	sz := ttk.LookupInt(e.ctx.Style, "-indicatorsize", state, 12)
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	fill := ttk.LookupColor(e.ctx.Style, "-indicatorbackground", state, bg)
	relief := ttk.LookupRelief(e.ctx.Style, "-indicatorrelief", state, option.ReliefRaised)
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)

	x := box.X
	y := box.Y + (box.Height-sz)/2

	d.SetForeground(gc, fill)
	d.FillRectangle(drawable, gc, x, y, uint(sz), uint(sz))

	if state&ttk.StateDisabled == 0 {
		border := draw.NewBorderFromPixel(bg)
		draw.Draw3DRectangle(d, drawable, gc, border, x, y, sz, sz, bw, relief)
	}
}

// --- diamondIndicatorElement ---
// Radiobutton.indicator: 3D raised/sunken diamond (Motif-style).
type diamondIndicatorElement struct{ ctx *ttk.DrawContext }

func newDiamondIndicatorElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &diamondIndicatorElement{ctx: ctx}
}

func (e *diamondIndicatorElement) Size(state ttk.State) (int, int, ttk.Padding) {
	sz := ttk.LookupInt(e.ctx.Style, "-indicatorsize", state, 12)
	// sz+3 to match C's DiamondIndicatorElementSize which adds 3 for the polygon
	return sz + 3 + 4, sz + 3 + 4, ttk.Padding{}
}

func (e *diamondIndicatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	sz := ttk.LookupInt(e.ctx.Style, "-indicatorsize", state, 12)
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	fill := ttk.LookupColor(e.ctx.Style, "-indicatorbackground", state, bg)
	relief := ttk.LookupRelief(e.ctx.Style, "-indicatorrelief", state, option.ReliefRaised)
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)

	total := sz + 3
	x := box.X
	y := box.Y + (box.Height-total)/2
	radius := sz / 2

	// Diamond vertices: left, bottom, right, top (as in C code).
	pts := []draw.Point{
		{X: x, Y: y + radius},
		{X: x + radius, Y: y + sz - 1},
		{X: x + sz - 1, Y: y + radius},
		{X: x + radius, Y: y},
	}

	d.SetForeground(gc, fill)
	draw.FillPolygon(d, drawable, gc, pts)

	if state&ttk.StateDisabled == 0 {
		border := draw.NewBorderFromPixel(bg)
		drawDiamond3D(d, drawable, gc, border, pts, bw, relief)
	}
}

// drawDiamond3D draws 3D shaded edges for a diamond (4-point polygon).
// Points order: left, bottom, right, top.
func drawDiamond3D(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, pts []draw.Point, bw int, relief option.Relief) {
	var lightPx, darkPx uint64
	switch relief {
	case option.ReliefRaised, option.ReliefRidge:
		lightPx = border.LightPixel
		darkPx = border.DarkPixel
	case option.ReliefSunken, option.ReliefGroove:
		lightPx = border.DarkPixel
		darkPx = border.LightPixel
	default:
		lightPx = border.BgPixel
		darkPx = border.BgPixel
	}

	left, bottom, right, top := pts[0], pts[1], pts[2], pts[3]
	for i := range bw {
		// Upper half: top→left and top→right (light for raised).
		d.SetForeground(gc, lightPx)
		d.DrawLine(drawable, gc, top.X, top.Y+i, left.X+i, left.Y)
		d.DrawLine(drawable, gc, top.X, top.Y+i, right.X-i, right.Y)
		// Lower half: bottom→left and bottom→right (dark for raised).
		d.SetForeground(gc, darkPx)
		d.DrawLine(drawable, gc, bottom.X, bottom.Y-i, left.X+i, left.Y)
		d.DrawLine(drawable, gc, bottom.X, bottom.Y-i, right.X-i, right.Y)
	}
}

// --- menuIndicatorElement ---
// Menubutton.indicator: small raised rectangular grip.
type menuIndicatorElement struct{ ctx *ttk.DrawContext }

func newMenuIndicatorElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &menuIndicatorElement{ctx: ctx}
}

func (e *menuIndicatorElement) Size(state ttk.State) (int, int, ttk.Padding) {
	// indicatorwidth≈15, indicatorheight≈6, plus left margin 5.
	return 15 + 5, 6, ttk.Padding{}
}

func (e *menuIndicatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	border := draw.NewBorderFromPixel(bg)
	// Left margin = 5.
	x := box.X + 5
	w := box.Width - 5
	h := 6
	y := box.Y + (box.Height-h)/2
	if w < 4 {
		return
	}
	d.SetForeground(gc, border.BgPixel)
	d.FillRectangle(drawable, gc, x, y, uint(w), uint(h))
	draw.Draw3DRectangle(d, drawable, gc, border, x, y, w, h, 2, option.ReliefRaised)
}

// --- arrowElement ---
// 3D shaded triangle arrows for scrollbar/spinbox (up/down/left/right).
type arrowDir int

const (
	arrowUp arrowDir = iota
	arrowDown
	arrowLeft
	arrowRight
)

type arrowElement struct {
	ctx *ttk.DrawContext
	dir arrowDir
}

func newArrowElementFactory(dir arrowDir) ttk.ElementFactory {
	return func(ctx *ttk.DrawContext) ttk.Element {
		return &arrowElement{ctx: ctx, dir: dir}
	}
}

func (e *arrowElement) Size(state ttk.State) (int, int, ttk.Padding) {
	sz := ttk.LookupInt(e.ctx.Style, "-arrowsize", state, 15)
	return sz, sz, ttk.Padding{}
}

func (e *arrowElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	bw := ttk.LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	relief := ttk.LookupRelief(e.ctx.Style, "-relief", state, option.ReliefRaised)
	border := draw.NewBorderFromPixel(bg)

	// 3px padding around the triangle interior.
	pad := 3
	sz := min(box.Width, box.Height) - 2*pad
	if sz < 2 {
		return
	}

	ax := box.X + (box.Width-sz)/2
	ay := box.Y + (box.Height-sz)/2

	var pts []draw.Point
	switch e.dir {
	case arrowUp:
		pts = []draw.Point{
			{X: ax + sz/2, Y: ay},
			{X: ax + sz - 1, Y: ay + sz - 1},
			{X: ax, Y: ay + sz - 1},
		}
	case arrowDown:
		pts = []draw.Point{
			{X: ax, Y: ay},
			{X: ax + sz - 1, Y: ay},
			{X: ax + sz/2, Y: ay + sz - 1},
		}
	case arrowLeft:
		pts = []draw.Point{
			{X: ax + sz - 1, Y: ay},
			{X: ax, Y: ay + sz/2},
			{X: ax + sz - 1, Y: ay + sz - 1},
		}
	case arrowRight:
		pts = []draw.Point{
			{X: ax, Y: ay},
			{X: ax + sz - 1, Y: ay + sz/2},
			{X: ax, Y: ay + sz - 1},
		}
	}

	d.SetForeground(gc, border.BgPixel)
	draw.FillPolygon(d, drawable, gc, pts)
	drawArrow3D(d, drawable, gc, border, pts, bw, relief, e.dir)
}

// drawArrow3D draws 3D shaded edges for a filled triangle arrow.
func drawArrow3D(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, pts []draw.Point, bw int, relief option.Relief, dir arrowDir) {
	var lightPx, darkPx uint64
	if relief == option.ReliefRaised {
		lightPx = border.LightPixel
		darkPx = border.DarkPixel
	} else {
		lightPx = border.DarkPixel
		darkPx = border.LightPixel
	}
	if bw < 1 {
		bw = 1
	}
	for i := 0; i < bw; i++ {
		switch dir {
		case arrowUp:
			// tip→right is right-going (dark), tip→left is left-going (light),
			// base is horizontal bottom (dark).
			p0, p1, p2 := pts[0], pts[1], pts[2] // tip, bottom-right, bottom-left
			d.SetForeground(gc, lightPx)
			d.DrawLine(drawable, gc, p0.X, p0.Y+i, p2.X+i, p2.Y-i)
			d.SetForeground(gc, darkPx)
			d.DrawLine(drawable, gc, p0.X, p0.Y+i, p1.X-i, p1.Y-i)
			d.DrawLine(drawable, gc, p2.X+i, p2.Y-i, p1.X-i, p1.Y-i)
		case arrowDown:
			p0, p1, p2 := pts[0], pts[1], pts[2] // top-left, top-right, tip
			d.SetForeground(gc, lightPx)
			d.DrawLine(drawable, gc, p0.X+i, p0.Y+i, p1.X-i, p1.Y+i)
			d.DrawLine(drawable, gc, p0.X+i, p0.Y+i, p2.X, p2.Y-i)
			d.SetForeground(gc, darkPx)
			d.DrawLine(drawable, gc, p1.X-i, p1.Y+i, p2.X, p2.Y-i)
		case arrowLeft:
			p0, p1, p2 := pts[0], pts[1], pts[2] // top-right, tip, bottom-right
			d.SetForeground(gc, lightPx)
			d.DrawLine(drawable, gc, p0.X-i, p0.Y+i, p1.X+i, p1.Y)
			d.SetForeground(gc, darkPx)
			d.DrawLine(drawable, gc, p2.X-i, p2.Y-i, p1.X+i, p1.Y)
			d.DrawLine(drawable, gc, p0.X-i, p0.Y+i, p2.X-i, p2.Y-i)
		case arrowRight:
			p0, p1, p2 := pts[0], pts[1], pts[2] // top-left, tip, bottom-left
			d.SetForeground(gc, lightPx)
			d.DrawLine(drawable, gc, p0.X+i, p0.Y+i, p0.X+i, p2.Y-i)
			d.SetForeground(gc, darkPx)
			d.DrawLine(drawable, gc, p0.X+i, p0.Y+i, p1.X-i, p1.Y)
			d.DrawLine(drawable, gc, p0.X+i, p2.Y-i, p1.X-i, p1.Y)
		}
	}
}

// --- sliderElement ---
// Scale slider: raised 3D box with a center groove line (Motif-style).
type sliderElement struct{ ctx *ttk.DrawContext }

func newSliderElementFactory(ctx *ttk.DrawContext) ttk.Element {
	return &sliderElement{ctx: ctx}
}

func (e *sliderElement) Size(state ttk.State) (int, int, ttk.Padding) {
	length := ttk.LookupInt(e.ctx.Style, "-sliderlength", state, 30)
	thickness := ttk.LookupInt(e.ctx.Style, "-sliderthickness", state, 15)
	orient := ttk.LookupInt(e.ctx.Style, "-sliderorient", state, 0) // 0=horizontal
	if orient == 1 {
		return thickness, length, ttk.Padding{}
	}
	return length, thickness, ttk.Padding{}
}

func (e *sliderElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	bw := ttk.LookupInt(e.ctx.Style, "-sliderborderwidth", state, 2)
	relief := ttk.LookupRelief(e.ctx.Style, "-sliderrelief", state, option.ReliefRaised)
	orient := ttk.LookupInt(e.ctx.Style, "-sliderorient", state, 0)

	border := draw.NewBorderFromPixel(bg)
	d.SetForeground(gc, border.BgPixel)
	d.FillRectangle(drawable, gc, box.X, box.Y, uint(box.Width), uint(box.Height))
	draw.Draw3DRectangle(d, drawable, gc, border, box.X, box.Y, box.Width, box.Height, bw, relief)

	// Draw center groove line (light then dark, creating a groove effect).
	if relief != option.ReliefFlat {
		if orient == 0 { // horizontal slider: vertical center line
			if box.Width > 4 {
				cx := box.X + box.Width/2
				d.SetForeground(gc, border.DarkPixel)
				d.DrawLine(drawable, gc, cx-1, box.Y+bw, cx-1, box.Y+box.Height-bw)
				d.SetForeground(gc, border.LightPixel)
				d.DrawLine(drawable, gc, cx, box.Y+bw, cx, box.Y+box.Height-bw)
			}
		} else { // vertical slider: horizontal center line
			if box.Height > 4 {
				cy := box.Y + box.Height/2
				d.SetForeground(gc, border.DarkPixel)
				d.DrawLine(drawable, gc, box.X+bw, cy-1, box.X+box.Width-bw, cy-1)
				d.SetForeground(gc, border.LightPixel)
				d.DrawLine(drawable, gc, box.X+bw, cy, box.X+box.Width-bw, cy)
			}
		}
	}
}

// --- sashElement ---
// Panedwindow sash: sunken line + raised handle square.
type sashElement struct {
	ctx    *ttk.DrawContext
	orient ttk.Orientation
}

func newSashElementFactory(orient ttk.Orientation) ttk.ElementFactory {
	return func(ctx *ttk.DrawContext) ttk.Element {
		return &sashElement{ctx: ctx, orient: orient}
	}
}

func (e *sashElement) Size(state ttk.State) (int, int, ttk.Padding) {
	thickness := ttk.LookupInt(e.ctx.Style, "-sashthickness", state, 6)
	handleSize := ttk.LookupInt(e.ctx.Style, "-handlesize", state, 8)
	sashPad := ttk.LookupInt(e.ctx.Style, "-sashpad", state, 2)
	if thickness < handleSize+2*sashPad {
		thickness = handleSize + 2*sashPad
	}
	if e.orient == ttk.Horizontal {
		return 0, thickness, ttk.Padding{}
	}
	return thickness, 0, ttk.Padding{}
}

func (e *sashElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box ttk.Box, state ttk.State) {
	bg := ttk.LookupColor(e.ctx.Style, "-background", state, frameColor)
	relief := ttk.LookupRelief(e.ctx.Style, "-sashrelief", state, option.ReliefSunken)
	handleSize := ttk.LookupInt(e.ctx.Style, "-handlesize", state, 8)
	handlePad := ttk.LookupInt(e.ctx.Style, "-handlepad", state, 8)
	border := draw.NewBorderFromPixel(bg)

	var lightPx, darkPx uint64
	switch relief {
	case option.ReliefRaised, option.ReliefRidge:
		lightPx = border.LightPixel
		darkPx = border.DarkPixel
	case option.ReliefSunken, option.ReliefGroove:
		lightPx = border.DarkPixel
		darkPx = border.LightPixel
	default:
		lightPx = border.LightPixel
		darkPx = border.DarkPixel
	}

	// Draw sash line.
	if e.orient == ttk.Horizontal {
		y := box.Y + box.Height/2 - 1
		d.SetForeground(gc, lightPx)
		d.DrawLine(drawable, gc, box.X, y, box.X+box.Width, y)
		d.SetForeground(gc, darkPx)
		d.DrawLine(drawable, gc, box.X, y+1, box.X+box.Width, y+1)
	} else {
		x := box.X + box.Width/2 - 1
		d.SetForeground(gc, lightPx)
		d.DrawLine(drawable, gc, x, box.Y, x, box.Y+box.Height)
		d.SetForeground(gc, darkPx)
		d.DrawLine(drawable, gc, x+1, box.Y, x+1, box.Y+box.Height)
	}

	// Draw handle (small raised square).
	if handleSize > 0 {
		var hx, hy int
		if e.orient == ttk.Horizontal {
			hx = box.X + handlePad
			hy = box.Y + (box.Height-handleSize)/2
		} else {
			hx = box.X + (box.Width-handleSize)/2
			hy = box.Y + handlePad
		}
		d.SetForeground(gc, border.BgPixel)
		d.FillRectangle(drawable, gc, hx, hy, uint(handleSize), uint(handleSize))
		draw.Draw3DRectangle(d, drawable, gc, border, hx, hy, handleSize, handleSize, 1, option.ReliefRaised)
	}
}

func init() {
	defaultTheme := ttk.CurrentTheme()
	theme := ttk.NewTheme("classic", defaultTheme)

	// Register classic elements.
	theme.RegisterElement("highlight", newHighlightElementFactory)
	theme.RegisterElement("Button.border", newButtonBorderElementFactory)
	theme.RegisterElement("Checkbutton.indicator", newSquareIndicatorElementFactory)
	theme.RegisterElement("Radiobutton.indicator", newDiamondIndicatorElementFactory)
	theme.RegisterElement("Menubutton.indicator", newMenuIndicatorElementFactory)
	theme.RegisterElement("uparrow", newArrowElementFactory(arrowUp))
	theme.RegisterElement("downarrow", newArrowElementFactory(arrowDown))
	theme.RegisterElement("leftarrow", newArrowElementFactory(arrowLeft))
	theme.RegisterElement("rightarrow", newArrowElementFactory(arrowRight))
	theme.RegisterElement("arrow", newArrowElementFactory(arrowUp))
	theme.RegisterElement("slider", newSliderElementFactory)
	theme.RegisterElement("hsash", newSashElementFactory(ttk.Horizontal))
	theme.RegisterElement("vsash", newSashElementFactory(ttk.Vertical))

	// Root style ".".
	root := theme.GetStyle(".")
	root.Defaults["-background"] = frameColor
	root.Defaults["-foreground"] = uint64(0x000000)
	root.Defaults["-borderwidth"] = 1
	root.Defaults["-troughcolor"] = troughBg
	root.Defaults["-selectbackground"] = selectBg
	root.Defaults["-selectforeground"] = uint64(0x000000)
	root.Defaults["-highlightthickness"] = 1
	root.Defaults["-highlightcolor"] = uint64(0x000000)

	root.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: activeBg},
	}
	root.Maps["-foreground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: disabledFg},
	}

	// TFrame style.
	tframe := theme.GetStyle("TFrame")
	tframe.Defaults["-relief"] = option.ReliefFlat
	tframe.Defaults["-borderwidth"] = 0

	// TLabel style.
	tlabel := theme.GetStyle("TLabel")
	tlabel.Defaults["-padding"] = ttk.Padding{Left: 4, Top: 2, Right: 4, Bottom: 2}
	tlabel.Defaults["-relief"] = option.ReliefFlat
	tlabel.Defaults["-borderwidth"] = 0

	// TButton style — classic Motif raised button.
	tbutton := theme.GetStyle("TButton")
	tbutton.Defaults["-anchor"] = option.AnchorCenter
	tbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tbutton.Defaults["-relief"] = option.ReliefRaised
	tbutton.Defaults["-borderwidth"] = 2
	tbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed, OffBits: ttk.StateDisabled}, Value: option.ReliefSunken},
	}
	tbutton.Maps["-background"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: activeBg},
	}

	// Toolbutton style.
	toolbutton := theme.GetStyle("Toolbutton")
	toolbutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	toolbutton.Defaults["-relief"] = option.ReliefFlat
	toolbutton.Defaults["-borderwidth"] = 2
	toolbutton.Defaults["-focussolid"] = 1
	theme.GetStyle("TNotebook.Tab").Defaults["-focussolid"] = 1
	toolbutton.Maps["-relief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: option.ReliefFlat},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StateActive}, Value: option.ReliefRaised},
	}

	// TMenubutton style.
	tmenubutton := theme.GetStyle("TMenubutton")
	tmenubutton.Defaults["-padding"] = ttk.Padding{Left: 8, Top: 4, Right: 8, Bottom: 4}
	tmenubutton.Defaults["-relief"] = option.ReliefRaised
	tmenubutton.Defaults["-borderwidth"] = 2

	// TCheckbutton — Motif-style indicator: raised when off, sunken when on.
	// Matches Tcl classic: red filled squares (Motif-style dark red #b03060
	// for selected, muted red #b05e5e for alternate).
	tCheckbutton := theme.GetStyle("TCheckbutton")
	tCheckbutton.Defaults["-indicatorrelief"] = option.ReliefRaised
	tCheckbutton.Defaults["-indicatorbackground"] = frameColor
	tCheckbutton.Defaults["-indicatorsize"] = 24
	tCheckbutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: indicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
	}
	tCheckbutton.Maps["-indicatorrelief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: option.ReliefRaised},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
	}

	// TRadiobutton — Motif-style diamond indicator: raised when unselected, sunken when selected.
	tRadiobutton := theme.GetStyle("TRadiobutton")
	tRadiobutton.Defaults["-indicatorrelief"] = option.ReliefRaised
	tRadiobutton.Defaults["-indicatorbackground"] = frameColor
	tRadiobutton.Defaults["-indicatorsize"] = 24
	tRadiobutton.Maps["-indicatorbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: indicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: altIndicator},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: frameColor},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
	}
	tRadiobutton.Maps["-indicatorrelief"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateAlternate}, Value: option.ReliefRaised},
		{Spec: ttk.StateSpec{OnBits: ttk.StateSelected}, Value: option.ReliefSunken},
		{Spec: ttk.StateSpec{OnBits: ttk.StatePressed}, Value: option.ReliefSunken},
	}

	// TSeparator styles.
	tsepH := theme.GetStyle("TSeparator.Horizontal")
	tsepH.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepH.Defaults["-relief"] = option.ReliefFlat
	tsepV := theme.GetStyle("TSeparator.Vertical")
	tsepV.Parent = theme.GetStyle("TSeparator") // takigo-only name, not a Tk style
	tsepV.Defaults["-relief"] = option.ReliefFlat

	// TScrollbar styles.
	vscroll := theme.GetStyle("Vertical.TScrollbar")
	vscroll.Defaults["-troughcolor"] = troughBg
	hscroll := theme.GetStyle("Horizontal.TScrollbar")
	hscroll.Defaults["-troughcolor"] = troughBg

	// TTreeview style.
	ttreeview := theme.GetStyle("TTreeview")
	ttreeview.Defaults["-fieldbackground"] = uint64(0xffffff)
	ttreeview.Defaults["-selectbackground"] = uint64(0x4a6984)
	ttreeview.Defaults["-selectforeground"] = uint64(0xffffff)

	// Treeview.Heading style.
	tvHeading := theme.GetStyle("Treeview.Heading")
	tvHeading.Defaults["-relief"] = option.ReliefRaised

	// Progressbar styles.
	hProgress := theme.GetStyle("Horizontal.TProgressbar")
	hProgress.Defaults["-barcolor"] = uint64(0x4a6984)
	vProgress := theme.GetStyle("Vertical.TProgressbar")
	vProgress.Defaults["-barcolor"] = uint64(0x4a6984)

	// Scale styles.
	hScale := theme.GetStyle("Horizontal.TScale")
	hScale.Defaults["-sliderlength"] = 30
	hScale.Defaults["-sliderthickness"] = 15
	hScale.Defaults["-sliderrelief"] = option.ReliefRaised
	hScale.Defaults["-sliderborderwidth"] = 2
	hScale.Defaults["-sliderorient"] = 0 // horizontal
	vScale := theme.GetStyle("Vertical.TScale")
	vScale.Defaults["-sliderlength"] = 30
	vScale.Defaults["-sliderthickness"] = 15
	vScale.Defaults["-sliderrelief"] = option.ReliefRaised
	vScale.Defaults["-sliderborderwidth"] = 2
	vScale.Defaults["-sliderorient"] = 1 // vertical

	// Sash styles.
	hSash := theme.GetStyle("Horizontal.Sash")
	hSash.Defaults["-sashrelief"] = option.ReliefSunken
	hSash.Defaults["-sashthickness"] = 6
	hSash.Defaults["-handlesize"] = 8
	hSash.Defaults["-handlepad"] = 8
	vSash := theme.GetStyle("Vertical.Sash")
	vSash.Defaults["-sashrelief"] = option.ReliefSunken
	vSash.Defaults["-sashthickness"] = 6
	vSash.Defaults["-handlesize"] = 8
	vSash.Defaults["-handlepad"] = 8

	// --- Layout overrides ---

	// TButton: highlight → Button.border → padding → label
	theme.RegisterLayout("TButton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("Button.border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF,
						ttk.L("label", ttk.Expand|ttk.FillF))))))

	// Toolbutton: same structure as TButton.
	theme.RegisterLayout("Toolbutton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("Button.border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF,
						ttk.L("label", ttk.Expand|ttk.FillF))))))

	// TMenubutton: highlight → Button.border → [indicator(right) + padding → label]
	theme.RegisterLayout("TMenubutton",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("Button.border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("Menubutton.indicator", ttk.PackRight),
					ttk.L("padding", ttk.FillXF,
						ttk.L("label", 0))))))

	// TEntry style.
	tentry := theme.GetStyle("TEntry")
	tentry.Defaults["-background"] = frameColor
	tentry.Defaults["-foreground"] = uint64(0x000000)
	tentry.Defaults["-fieldbackground"] = uint64(0xffffff)
	tentry.Defaults["-selectbackground"] = selectBg
	tentry.Defaults["-selectforeground"] = uint64(0xffffff)
	tentry.Defaults["-insertwidth"] = 1
	tentry.Defaults["-padding"] = ttk.Padding{Left: 1, Top: 1, Right: 1, Bottom: 1}
	tentry.Defaults["-insertcolor"] = uint64(0x000000)
	tentry.Maps["-fieldbackground"] = ttk.StateMap[any]{
		{Spec: ttk.StateSpec{OnBits: ttk.StateFocus}, Value: uint64(0xffffff)},
		{Spec: ttk.StateSpec{OnBits: ttk.StateDisabled}, Value: frameColor},
		{Spec: ttk.StateSpec{}, Value: uint64(0xffffff)},
	}

	// TEntry: highlight → field → padding → textarea (field=layout engine, textarea=NullElement)
	theme.RegisterLayout("TEntry",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	// TCombobox: highlight → border → [downarrow(right) + padding → textarea]
	theme.RegisterLayout("TCombobox",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF,
					ttk.L("downarrow", ttk.PackRight|ttk.FillYF),
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	// TSpinbox: highlight → border → [arrows(right) + padding → textarea]
	theme.RegisterLayout("TSpinbox",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF,
					ttk.L("uparrow", ttk.PackRight|ttk.PackTop),
					ttk.L("downarrow", ttk.PackRight|ttk.PackBottom),
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	// Horizontal.TScale: highlight → trough → slider
	theme.RegisterLayout("Horizontal.TScale",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("trough", ttk.Expand|ttk.FillF,
					ttk.L("slider", ttk.PackLeft)))))

	// Vertical.TScale: highlight → trough → slider
	theme.RegisterLayout("Vertical.TScale",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("trough", ttk.Expand|ttk.FillF,
					ttk.L("slider", ttk.PackTop)))))

	// Horizontal.Sash and Vertical.Sash.
	theme.RegisterLayout("Horizontal.Sash",
		ttk.L("hsash", ttk.FillXF))
	theme.RegisterLayout("Vertical.Sash",
		ttk.L("vsash", ttk.FillYF))

	// Treeview: highlight → border → padding → treearea
	theme.RegisterLayout("Treeview",
		ttk.L("background", ttk.Expand|ttk.FillF,
			ttk.L("highlight", ttk.Expand|ttk.FillF,
				ttk.L("border", ttk.Expand|ttk.FillF|ttk.Border,
					ttk.L("padding", ttk.Expand|ttk.FillF)))))

	ttk.RegisterTheme(theme)
}
