package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// --- BackgroundElement ---

// BackgroundElement fills its box with the -background color from the style.
type BackgroundElement struct {
	ctx *DrawContext
}

func NewBackgroundElementFactory(ctx *DrawContext) Element {
	return &BackgroundElement{ctx: ctx}
}

func (e *BackgroundElement) Size(State) (int, int, Padding) {
	return 0, 0, Padding{}
}

func (e *BackgroundElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(drawable, gc, box.X, box.Y, uint(box.Width), uint(box.Height))
}

// --- BorderElement ---

// BorderElement draws a 3D relief border.
type BorderElement struct {
	ctx *DrawContext
}

func NewBorderElementFactory(ctx *DrawContext) Element {
	return &BorderElement{ctx: ctx}
}

func (e *BorderElement) Size(state State) (int, int, Padding) {
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	return 0, 0, UniformPadding(bw)
}

func (e *BorderElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	relief := LookupRelief(e.ctx.Style, "-relief", state, option.ReliefFlat)

	if bw <= 0 || relief == option.ReliefFlat {
		return
	}
	border := draw.NewBorderFromPixel(bg)
	draw.Draw3DRectangle(d, drawable, gc, border,
		box.X, box.Y, box.Width, box.Height, bw, relief)
}

// AltBorderElement is the alt theme's border (ttkDefaultTheme.c, which
// despite its name implements "alt"): DrawBorder's corner shading.
type AltBorderElement struct{ BorderElement }

func NewAltBorderElementFactory(ctx *DrawContext) Element {
	return &AltBorderElement{BorderElement{ctx: ctx}}
}

func (e *AltBorderElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	relief := LookupRelief(e.ctx.Style, "-relief", state, option.ReliefFlat)
	if bw <= 0 || relief == option.ReliefFlat {
		return
	}
	border := draw.NewBorderFromPixel(bg)
	borderColor := LookupColor(e.ctx.Style, "-bordercolor", state, 0x414141)
	drawAltBorder(d, drawable, gc, border, borderColor, box, bw, relief)
}

type shade int

const (
	shFlat shade = iota
	shLite
	shDark
	shBrdr
)

// Top-left outer, top-left inner, bottom-right inner, bottom-right outer.
var altShadowColors = map[option.Relief][4]shade{
	option.ReliefGroove: {shDark, shLite, shDark, shLite},
	option.ReliefRaised: {shLite, shFlat, shDark, shBrdr},
	option.ReliefRidge:  {shLite, shDark, shLite, shDark},
	option.ReliefSolid:  {shBrdr, shBrdr, shBrdr, shBrdr},
	option.ReliefSunken: {shBrdr, shDark, shFlat, shLite},
}

var altThinShadowColors = map[option.Relief][2]shade{
	option.ReliefGroove: {shDark, shLite},
	option.ReliefRaised: {shLite, shDark},
	option.ReliefRidge:  {shLite, shDark},
	option.ReliefSolid:  {shBrdr, shBrdr},
	option.ReliefSunken: {shDark, shLite},
}

// drawCorner ports DrawCorner (ttkDefaultTheme.c, the alt theme).
func drawCorner(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, borderColor uint64, x, y, w, h int, bottomRight bool, c shade) {
	pixel := border.BgPixel
	switch c {
	case shLite:
		pixel = border.LightPixel
	case shDark:
		pixel = border.DarkPixel
	case shBrdr:
		pixel = borderColor
	}
	w--
	h--
	pts := []platform.Point{{X: int16(x), Y: int16(y + h)}, {X: int16(x), Y: int16(y)}, {X: int16(x + w), Y: int16(y)}}
	if bottomRight {
		pts[1] = platform.Point{X: int16(x + w), Y: int16(y + h)}
		// Measured against Tk under X11: the bottom-right corner leaves both
		// of its endpoints to the top-left corner's colour.
		if h > 0 && w > 0 {
			pts[0].X++
			pts[2].Y++
		}
	}
	d.SetForeground(gc, pixel)
	d.DrawLines(drawable, gc, pts, 0)
}

// drawAltBorder ports DrawBorder (ttkDefaultTheme.c): 1- and 2-pixel
// borders are drawn as corners from the shadow tables, wider ones Motif-style.
func drawAltBorder(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, borderColor uint64, b Box, bw int, relief option.Relief) {
	corner := func(x, y, w, h int, bottomRight bool, c shade) {
		drawCorner(d, drawable, gc, border, borderColor, x, y, w, h, bottomRight, c)
	}
	switch bw {
	case 2:
		sc := altShadowColors[relief]
		corner(b.X, b.Y, b.Width, b.Height, false, sc[0])
		corner(b.X+1, b.Y+1, b.Width-2, b.Height-2, false, sc[1])
		corner(b.X+1, b.Y+1, b.Width-2, b.Height-2, true, sc[2])
		corner(b.X, b.Y, b.Width, b.Height, true, sc[3])
	case 1:
		sc := altThinShadowColors[relief]
		corner(b.X, b.Y, b.Width, b.Height, false, sc[0])
		corner(b.X, b.Y, b.Width, b.Height, true, sc[1])
	default:
		draw.Draw3DRectangle(d, drawable, gc, border, b.X, b.Y, b.Width, b.Height, bw, relief)
	}
}

// --- PaddingElement ---

// PaddingElement provides spacing but draws nothing.
type PaddingElement struct {
	ctx *DrawContext
}

func NewPaddingElementFactory(ctx *DrawContext) Element {
	return &PaddingElement{ctx: ctx}
}

func (e *PaddingElement) Size(state State) (int, int, Padding) {
	p := LookupPadding(e.ctx.Style, "-padding", state, Padding{})
	relief := LookupRelief(e.ctx.Style, "-relief", state, option.ReliefFlat)
	shift := LookupInt(e.ctx.Style, "-shiftrelief", state, 0)
	return 0, 0, RelievePadding(p, relief, shift)
}

func (e *PaddingElement) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, Box, State) {
}

// --- FocusElement ---

// FocusElement draws a 1px focus ring when StateFocus is set.
type FocusElement struct {
	ctx *DrawContext
}

func NewFocusElementFactory(ctx *DrawContext) Element {
	return &FocusElement{ctx: ctx}
}

func (e *FocusElement) Size(state State) (int, int, Padding) {
	return 0, 0, UniformPadding(LookupInt(e.ctx.Style, "-focusthickness", state, 1))
}

// Draw ports FocusElementDraw/DrawFocusRing: a dotted ring by default, or
// a solid one -focusthickness wide with -focussolid.
func (e *FocusElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	if state&StateFocus == 0 {
		return
	}
	d.SetForeground(gc, LookupColor(e.ctx.Style, "-focuscolor", state, 0x000000))
	t := LookupInt(e.ctx.Style, "-focusthickness", state, 1)
	if LookupInt(e.ctx.Style, "-focussolid", state, 0) == 0 {
		drawDottedRect(d, drawable, gc, box.X, box.Y, box.Width, box.Height)
		return
	}
	t = max(t, 1)
	if box.Width >= 2*t && box.Height >= 2*t {
		d.FillRectangle(drawable, gc, box.X, box.Y, uint(box.Width), uint(t))
		d.FillRectangle(drawable, gc, box.X, box.Y+box.Height-t, uint(box.Width), uint(t))
		d.FillRectangle(drawable, gc, box.X, box.Y+t, uint(t), uint(box.Height-2*t))
		d.FillRectangle(drawable, gc, box.X+box.Width-t, box.Y+t, uint(t), uint(box.Height-2*t))
	}
}

// --- LabelElement ---

// LabelElement renders text and/or an image. Requires a TextProvider.
type LabelElement struct {
	ctx      *DrawContext
	provider TextProvider
}

// NewLabelElementFactory returns a factory that creates a LabelElement bound to a TextProvider.
func NewLabelElementFactory(provider TextProvider) ElementFactory {
	return func(ctx *DrawContext) Element {
		return &LabelElement{ctx: ctx, provider: provider}
	}
}

// textLines lays the text out like the text part of Tk's label element
// (ttkLabel.c TextSetup): Tk_ComputeTextLayout with -wraplength.
func (e *LabelElement) textLines(state State) (lines []string, tw, th int) {
	text := e.provider.GetText()
	f := e.provider.GetFont()
	if f == nil {
		return nil, 0, 0
	}
	if text == "" {
		// Tk_ComputeTextLayout lays out "" as one empty line; with an image
		// the text part is not set up at all (LabelSetup).
		if e.provider.GetImage() != nil {
			return nil, 0, 0
		}
		return nil, 0, f.Metrics().Linespace()
	}
	wrap := 0
	if e.ctx.Style != nil {
		wrap = LookupInt(e.ctx.Style, "-wraplength", state, 0)
	}
	lines = font.WrapLines(f, text, wrap)
	for _, l := range lines {
		tw = max(tw, f.MeasureString(l))
	}
	return lines, tw, len(lines) * f.Metrics().Linespace()
}

func (e *LabelElement) Size(state State) (int, int, Padding) {
	f := e.provider.GetFont()
	img := e.provider.GetImage()
	compound := e.provider.GetCompound()
	_, tw, th := e.textLines(state)

	// TextReqWidth: -width in average characters; negative is a minimum.
	if f != nil && e.ctx.Style != nil {
		if styleWidth := LookupInt(e.ctx.Style, "-width", state, 0); styleWidth != 0 {
			avgCharWidth := f.MeasureString("0")
			if styleWidth < 0 {
				tw = max(tw, -styleWidth*avgCharWidth)
			} else {
				tw = styleWidth * avgCharWidth
			}
		}
	}

	w, h := widget.CompoundSize(compound, img, tw, th)
	return w, h, Padding{}
}

// Draw anchors the content in the box with -anchor (Ttk_AnchorBox, element
// default "w") and draws multi-line text with -justify.
func (e *LabelElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	f := e.provider.GetFont()
	img := e.provider.GetImage()
	compound := e.provider.GetCompound()

	fg := LookupColor(e.ctx.Style, "-foreground", state, 0x000000)
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)
	anchor := option.AnchorW
	justify := option.JustifyLeft
	if e.ctx.Style != nil {
		if a, ok := e.ctx.Style.LookupAs[option.Anchor]("-anchor", state); ok {
			anchor = a
		}
		if j, ok := e.ctx.Style.LookupAs[option.Justify]("-justify", state); ok {
			justify = j
		}
	}

	lines, tw, th := e.textLines(state)
	contentW, contentH := widget.CompoundSize(compound, img, tw, th)
	cx, cy := widget.ComputeAnchor(anchor, box.Width, box.Height, 0, 0, 0, contentW, contentH)
	cx += box.X
	cy += box.Y

	hasImg := img != nil
	hasText := len(lines) > 0

	if hasImg && hasText && compound != widget.CompoundNone {
		drawCompound(d, drawable, gc, img, f, lines, justify, compound,
			cx, cy, contentW, contentH, tw, th, fg, bg, e.ctx)
	} else if hasImg {
		imgW := img.Width()
		imgH := img.Height()
		ix := cx + (contentW-imgW)/2
		iy := cy + (contentH-imgH)/2
		img.Draw(d, drawable, gc, e.ctx.Depth,
			0, 0, imgW, imgH, ix, iy, bg)
	} else if hasText {
		drawTextLines(d, drawable, f, lines, justify, cx, cy, tw, fg)
		if up, ok := e.provider.(interface{ GetUnderline() int }); ok {
			underlineChar(d, drawable, gc, f, lines, justify, cx, cy, tw, up.GetUnderline(), fg)
		}
	}
}

// underlineChar ports Tk_UnderlineTextLayout: underline character index u
// (counted across the laid-out lines) with the font's underline metrics.
func underlineChar(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	f font.Font, lines []string, justify option.Justify, x, y, tw, u int, fgPixel uint64) {
	if u < 0 {
		return
	}
	ls := f.Metrics().Linespace()
	for i, line := range lines {
		runes := []rune(line)
		if u >= len(runes) {
			u -= len(runes)
			continue
		}
		lx := x
		switch justify {
		case option.JustifyCenter:
			lx = x + (tw-f.MeasureString(line))/2
		case option.JustifyRight:
			lx = x + tw - f.MeasureString(line)
		}
		pos, h := font.Underline(f)
		ux := lx + f.MeasureString(string(runes[:u]))
		d.SetForeground(gc, fgPixel)
		d.FillRectangle(drawable, gc, ux, y+i*ls+f.Metrics().Ascent+pos,
			uint(f.MeasureString(string(runes[u]))), uint(h))
		return
	}
}

// drawTextLines draws laid-out lines inside a block tw wide, each line
// positioned by justify like Tk_DrawTextLayout.
func drawTextLines(d platform.DisplayServer, drawable platform.DrawableID, f font.Font,
	lines []string, justify option.Justify, x, y, tw int, fgPixel uint64) {
	ls := f.Metrics().Linespace()
	for i, line := range lines {
		lx := x
		switch justify {
		case option.JustifyCenter:
			lx = x + (tw-f.MeasureString(line))/2
		case option.JustifyRight:
			lx = x + tw - f.MeasureString(line)
		}
		drawText(d, drawable, f, line, lx, y+i*ls, fgPixel)
	}
}

func drawText(_ platform.DisplayServer, drawable platform.DrawableID, f font.Font, text string, x, y int, fgPixel uint64) {
	df, ok := f.(platform.DrawableFont)
	if !ok {
		return
	}
	m := f.Metrics()
	baseline := y + m.Ascent
	// Extract RGB from pixel.
	r := uint16((fgPixel>>16)&0xFF) * 257
	g := uint16((fgPixel>>8)&0xFF) * 257
	b := uint16((fgPixel)&0xFF) * 257
	df.DrawString(drawable, x, baseline, text, fgPixel, r, g, b)
}

func drawCompound(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	img widget.WidgetImage, f font.Font, lines []string, justify option.Justify,
	compound widget.Compound, cx, cy, contentW, contentH, tw, th int,
	fgPixel, bgPixel uint64, ctx *DrawContext) {

	imgW := img.Width()
	imgH := img.Height()

	var imgX, imgY, textX, textY int
	switch compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + 4
		textY = cy + (contentH-th)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-th)/2
		imgX = cx + tw + 4
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-tw)/2
		textY = cy + imgH + 4
	case widget.CompoundBottom:
		textX = cx + (contentW-tw)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + th + 4
	case widget.CompoundCenter:
		imgX = cx + (contentW-imgW)/2
		imgY = cy + (contentH-imgH)/2
		textX = cx + (contentW-tw)/2
		textY = cy + (contentH-th)/2
	}

	img.Draw(d, drawable, gc, ctx.Depth,
		0, 0, imgW, imgH, imgX, imgY, bgPixel)
	drawTextLines(d, drawable, f, lines, justify, textX, textY, tw, fgPixel)
}

// --- FieldElement ---

// FieldElement draws an inset field background for entry/combobox/spinbox widgets.
type FieldElement struct {
	ctx *DrawContext
}

func NewFieldElementFactory(ctx *DrawContext) Element {
	return &FieldElement{ctx: ctx}
}

// Size and Draw port FieldElementSize/FieldElementDraw (ttkElements.c):
// a sunken border in the -fieldbackground colour, widened to the focus ring.
func (e *FieldElement) Size(state State) (int, int, Padding) {
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	fw := LookupInt(e.ctx.Style, "-focuswidth", state, 2)
	if fw > 0 && bw < 2 {
		bw = fw
	}
	return 0, 0, UniformPadding(bw)
}

func (e *FieldElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	fieldBg := LookupColor(e.ctx.Style, "-fieldbackground", state, 0xffffff)
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	fw := LookupInt(e.ctx.Style, "-focuswidth", state, 2)
	border := draw.NewBorderFromPixel(fieldBg)
	if fw > 0 && state&StateFocus != 0 {
		focusColor := LookupColor(e.ctx.Style, "-focuscolor", state, 0x4a6984)
		if fw > 1 && box.Width >= 2 && box.Height >= 2 {
			drawFocusField(d, drawable, gc, box, focusColor, fieldBg)
			return
		}
		draw.Fill3DRectangle(d, drawable, gc, border, box.X, box.Y, box.Width, box.Height, bw, option.ReliefSunken)
		d.SetForeground(gc, focusColor)
		d.DrawRectangle(drawable, gc, box.X, box.Y, uint(box.Width-1), uint(box.Height-1))
		return
	}
	draw.Fill3DRectangle(d, drawable, gc, border, box.X, box.Y, box.Width, box.Height, bw, option.ReliefSunken)
}

// drawFocusField is the 2-pixel rounded focus ring both field elements draw.
func drawFocusField(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	box Box, focusColor, fieldBg uint64) {
	x1, y1 := box.X, box.Y
	x2, y2 := box.X+box.Width-1, box.Y+box.Height-1
	d.SetForeground(gc, focusColor)
	d.DrawLine(drawable, gc, x1+1, y1, x2-1, y1)
	d.DrawLine(drawable, gc, x1+1, y2, x2-1, y2)
	d.DrawLine(drawable, gc, x1, y1+1, x1, y2-1)
	d.DrawLine(drawable, gc, x2, y1+1, x2, y2-1)
	d.DrawRectangle(drawable, gc, x1+1, y1+1, uint(box.Width-3), uint(box.Height-3))
	d.SetForeground(gc, fieldBg)
	d.FillRectangle(drawable, gc, x1+2, y1+2, uint(box.Width-4), uint(box.Height-4))
}

// AltFieldElement is the alt theme's field (ttkDefaultTheme.c).
type AltFieldElement struct{ FieldElement }

func NewAltFieldElementFactory(ctx *DrawContext) Element {
	return &AltFieldElement{FieldElement{ctx: ctx}}
}

// Size and Draw port the alt theme's FieldElementSize/FieldElementDraw:
// a 2-pixel DrawFieldBorder in -fieldbackground shades.
func (e *AltFieldElement) Size(state State) (int, int, Padding) {
	return 0, 0, UniformPadding(2)
}

func (e *AltFieldElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	fieldBg := LookupColor(e.ctx.Style, "-fieldbackground", state, 0xffffff)
	borderColor := LookupColor(e.ctx.Style, "-bordercolor", state, 0x000000)
	fw := LookupInt(e.ctx.Style, "-focuswidth", state, 2)
	border := draw.NewBorderFromPixel(fieldBg)
	x1, y1 := box.X, box.Y
	focus := fw > 0 && state&StateFocus != 0
	focusColor := LookupColor(e.ctx.Style, "-focuscolor", state, 0x4a6984)

	if focus && fw > 1 && box.Width >= 2 && box.Height >= 2 {
		drawFocusField(d, drawable, gc, box, focusColor, fieldBg)
		return
	}
	d.SetForeground(gc, fieldBg)
	d.FillRectangle(drawable, gc, box.X, box.Y, uint(box.Width), uint(box.Height))
	drawCorner(d, drawable, gc, border, borderColor, box.X, box.Y, box.Width, box.Height, false, shDark)
	drawCorner(d, drawable, gc, border, borderColor, box.X+1, box.Y+1, box.Width-2, box.Height-2, false, shBrdr)
	drawCorner(d, drawable, gc, border, borderColor, box.X+1, box.Y+1, box.Width-2, box.Height-2, true, shLite)
	drawCorner(d, drawable, gc, border, borderColor, box.X, box.Y, box.Width, box.Height, true, shFlat)
	if focus {
		d.SetForeground(gc, focusColor)
		d.DrawRectangle(drawable, gc, x1, y1, uint(box.Width-1), uint(box.Height-1))
	}
}

// --- MenubuttonIndicatorElement ---

// MenubuttonIndicatorElement ports MenuIndicatorElement (ttkElements.c): a
// filled down arrow of -arrowsize inside -arrowpadding.
type MenubuttonIndicatorElement struct {
	ctx *DrawContext
}

func NewMenubuttonIndicatorElementFactory(ctx *DrawContext) Element {
	return &MenubuttonIndicatorElement{ctx: ctx}
}

func (e *MenubuttonIndicatorElement) Size(state State) (int, int, Padding) {
	h := LookupInt(e.ctx.Style, "-arrowsize", state, 5)
	p := LookupPadding(e.ctx.Style, "-arrowpadding", state, UniformPadding(3))
	return 2*h + 1 + p.Left + p.Right, h + 1 + p.Top + p.Bottom, Padding{}
}

func (e *MenubuttonIndicatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	h := LookupInt(e.ctx.Style, "-arrowsize", state, 5)
	w, ht := 2*h+1, h+1
	b := Box{box.X + (box.Width-w)/2, box.Y + (box.Height-ht)/2, w, ht}
	d.SetForeground(gc, LookupColor(e.ctx.Style, "-arrowcolor", state, 0x000000))
	// TtkFillArrow fills the triangle, then outlines it (closing the path)
	// and plots the third point, so the edges X leaves out are drawn too.
	pts := arrowDownPoints(b)
	d.FillPolygon(drawable, gc, pts, 2, 0)
	d.DrawLines(drawable, gc, append(pts, pts[0]), 0)
	d.DrawLine(drawable, gc, int(pts[2].X), int(pts[2].Y), int(pts[2].X), int(pts[2].Y))
}

// arrowUpPoints ports ArrowPoints (ttkDefaultTheme.c) for ARROW_UP.
func arrowUpPoints(b Box) []platform.Point {
	h := (b.Width - 1) / 2
	cx, cy := b.X+h, b.Y
	if b.Height <= h {
		h = b.Height - 1
	}
	return []platform.Point{
		{X: int16(cx), Y: int16(cy)},
		{X: int16(cx - h), Y: int16(cy + h)},
		{X: int16(cx + h), Y: int16(cy + h)},
	}
}

// arrowDownPoints ports ArrowPoints (ttkDefaultTheme.c) for ARROW_DOWN.
func arrowDownPoints(b Box) []platform.Point {
	h := (b.Width - 1) / 2
	cx, cy := b.X+h, b.Y+b.Height-1
	if b.Height <= h {
		h = b.Height - 1
	}
	return []platform.Point{
		{X: int16(cx), Y: int16(cy)},
		{X: int16(cx - h), Y: int16(cy - h)},
		{X: int16(cx + h), Y: int16(cy - h)},
	}
}

// --- SeparatorElement ---

// SeparatorElement draws a horizontal or vertical separator line.
type SeparatorElement struct {
	ctx    *DrawContext
	orient Orientation
}

// Orientation for separator.
type Orientation = option.Orient

// The values of -orient.
const (
	Horizontal = option.Horizontal
	Vertical   = option.Vertical
)

// NewSeparatorElementFactory returns a factory for separator elements.
func NewSeparatorElementFactory(orient Orientation) ElementFactory {
	return func(ctx *DrawContext) Element {
		return &SeparatorElement{ctx: ctx, orient: orient}
	}
}

func (e *SeparatorElement) Size(State) (int, int, Padding) {
	if e.orient == Horizontal {
		return 0, 2, Padding{}
	}
	return 2, 0, Padding{}
}

func (e *SeparatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)
	border := draw.NewBorderFromPixel(bg)

	if e.orient == Horizontal {
		// Draw two lines: dark on top, light below.
		y := box.Y + box.Height/2 - 1
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(drawable, gc, box.X, y, box.X+box.Width-1, y)
		d.SetForeground(gc, border.LightPixel)
		d.DrawLine(drawable, gc, box.X, y+1, box.X+box.Width-1, y+1)
	} else {
		x := box.X + box.Width/2 - 1
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(drawable, gc, x, box.Y, x, box.Y+box.Height-1)
		d.SetForeground(gc, border.LightPixel)
		d.DrawLine(drawable, gc, x+1, box.Y, x+1, box.Y+box.Height-1)
	}
}
