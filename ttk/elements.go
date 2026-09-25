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

func (e *FocusElement) Size(State) (int, int, Padding) {
	return 0, 0, UniformPadding(1)
}

func (e *FocusElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	if state&StateFocus == 0 {
		return
	}
	focusColor := LookupColor(e.ctx.Style, "-focuscolor", state, 0x000000)
	d.SetForeground(gc, focusColor)
	d.DrawRectangle(drawable, gc, box.X, box.Y, uint(box.Width-1), uint(box.Height-1))
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
		if v, ok := e.ctx.Style.Lookup("-anchor", state); ok {
			if a, ok := v.(option.Anchor); ok {
				anchor = a
			}
		}
		if v, ok := e.ctx.Style.Lookup("-justify", state); ok {
			if j, ok := v.(option.Justify); ok {
				justify = j
			}
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
	r := uint16((fgPixel>>16)&0xFF) << 8
	g := uint16((fgPixel>>8)&0xFF) << 8
	b := uint16((fgPixel)&0xFF) << 8
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

func (e *FieldElement) Size(state State) (int, int, Padding) {
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	return 0, 0, UniformPadding(bw)
}

func (e *FieldElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	fieldBg := LookupColor(e.ctx.Style, "-fieldbackground", state, 0xffffff)
	bw := LookupInt(e.ctx.Style, "-borderwidth", state, 2)
	bg := LookupColor(e.ctx.Style, "-background", state, 0xd9d9d9)

	// Fill field background.
	d.SetForeground(gc, fieldBg)
	d.FillRectangle(drawable, gc, box.X, box.Y, uint(box.Width), uint(box.Height))

	// Sunken border.
	if bw > 0 {
		border := draw.NewBorderFromPixel(bg)
		draw.Draw3DRectangle(d, drawable, gc, border, box.X, box.Y, box.Width, box.Height, bw, option.ReliefSunken)
	}

	// Focus ring inside border.
	if state&StateFocus != 0 {
		focusColor := LookupColor(e.ctx.Style, "-focuscolor", state, 0x4a6984)
		d.SetForeground(gc, focusColor)
		d.DrawRectangle(drawable, gc, box.X+bw-1, box.Y+bw-1,
			uint(box.Width-2*bw+1), uint(box.Height-2*bw+1))
	}
}

// --- MenubuttonIndicatorElement ---

// MenubuttonIndicatorElement draws a small downward-pointing arrow for TMenubutton.
type MenubuttonIndicatorElement struct {
	ctx *DrawContext
}

func NewMenubuttonIndicatorElementFactory(ctx *DrawContext) Element {
	return &MenubuttonIndicatorElement{ctx: ctx}
}

func (e *MenubuttonIndicatorElement) Size(state State) (int, int, Padding) {
	// Compact arrow: 11px wide, fills height, with 4px left margin.
	return 11 + 4, 0, Padding{}
}

func (e *MenubuttonIndicatorElement) Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, box Box, state State) {
	fg := LookupColor(e.ctx.Style, "-foreground", state, 0x000000)
	if state&StateDisabled != 0 {
		fg = LookupColor(e.ctx.Style, "-foreground", StateDisabled, 0xa3a3a3)
	}

	// Draw a small centered downward triangle.
	aw := 7 // arrow width (odd for symmetric look)
	ah := 4 // arrow height
	ax := box.X + 4 + (box.Width-4-aw)/2
	ay := box.Y + (box.Height-ah)/2

	d.SetForeground(gc, fg)
	for i := range ah {
		x0 := ax + i
		x1 := ax + aw - 1 - i
		y := ay + i
		if x0 <= x1 {
			d.DrawLine(drawable, gc, x0, y, x1, y)
		}
	}
}

// --- SeparatorElement ---

// SeparatorElement draws a horizontal or vertical separator line.
type SeparatorElement struct {
	ctx    *DrawContext
	orient Orientation
}

// Orientation for separator.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
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
