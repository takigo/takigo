// Package tkbutton holds what the label, button, checkbutton and
// radiobutton widgets share. It ports tk/generic/tkButton.c and
// tk/unix/tkUnixButton.c, which keep the four in one TkButton structure
// told apart by a type: the common fields, TkpComputeButtonGeometry and
// TkpDisplayButton.
package tkbutton

import (
	"strings"
	"unicode/utf8"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
)

// Type tells the four widgets apart (TYPE_LABEL ... TYPE_RADIO_BUTTON).
type Type int

const (
	TypeLabel Type = iota
	TypeButton
	TypeCheck
	TypeRadio
)

// DefaultState is the value of a button's -default option.
type DefaultState int

const (
	DefaultDisabled DefaultState = iota
	DefaultNormal
	DefaultActive
)

// Shared is the part of TkButton the four widgets have in common; each
// embeds it after widget.Base. The exported fields are the options they
// share.
type Shared struct {
	Text      string
	Anchor    option.Anchor
	Justify   option.Justify
	WrapLen   int // -wraplength in pixels (0 = no wrapping)
	Underline int // index of the character to underline (-1 = none)

	Img       widget.WidgetImage
	SelectImg widget.WidgetImage // -selectimage, shown while selected
	Compound  widget.Compound

	// WidthChars and HeightChars are Tk's -width and -height: characters
	// and lines of text, pixels when an image is shown (0 = natural size).
	WidthChars  int
	HeightChars int

	State widget.State

	// IndicatorOn and SelectColor are a check or radio button's
	// -indicatoron and -selectcolor.
	IndicatorOn bool
	SelectColor *color.ColorRef

	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef
	DisabledFg       *color.ColorRef

	// Default is a button's -default: room for (and, when active, a drawn)
	// default ring around it.
	Default DefaultState

	// OverRelief and OffRelief are -overrelief and -offrelief; a check or
	// radio button without an indicator shows OffRelief until selected.
	OverRelief option.Relief
	OffRelief  option.Relief

	HasFocus bool

	typ     Type
	pressed bool

	textWidth, textHeight             int
	indicatorSpace, indicatorDiameter int
	lines                             []string
	linesKey                          layoutKey
}

// layoutKey is what the line breaks of the text depend on.
type layoutKey struct {
	text string
	font font.Font
	wrap int
}

// Selection is what a check or radio button shows.
type Selection struct {
	Selected, Tristate bool
}

// Init gives s the option defaults of a widget of type t (the option
// table in tkButton.c).
func Init(s *Shared, t Type) {
	s.typ = t
	s.Anchor = option.AnchorCenter
	s.Justify = option.JustifyCenter
	s.Underline = -1
	switch t {
	case TypeButton:
		s.OverRelief, s.OffRelief = option.ReliefRaised, option.ReliefFlat
	case TypeCheck, TypeRadio:
		s.IndicatorOn = true
		s.OffRelief = option.ReliefRaised
	default:
	}
}

// Type returns the widget type s was initialised for.
func (s *Shared) Type() Type { return s.typ }

// Pressed reports whether button 1 is held down on the widget.
func (s *Shared) Pressed() bool { return s.pressed }

// SetPressed records whether button 1 is held down on the widget.
func (s *Shared) SetPressed(p bool) { s.pressed = p }

// TextSize returns the size of the laid-out text (butPtr->textWidth and
// textHeight), as of the last ComputeGeometry.
func (s *Shared) TextSize() (w, h int) { return s.textWidth, s.textHeight }

// IndicatorSpace returns the room left for the indicator, as of the last
// ComputeGeometry.
func (s *Shared) IndicatorSpace() int { return s.indicatorSpace }

// Inset is butPtr->inset: border, highlight and the default ring's room.
func Inset(b *widget.Base, s *Shared) int {
	inset := b.BorderWidth + b.HighlightWidth
	if s.Default != DefaultDisabled {
		inset += 5
	}
	return inset
}

// TextLines lays the text out as Tk_ComputeTextLayout does: a line per
// newline, wrapped at -wraplength when set. "" is one empty line.
func (s *Shared) TextLines(f font.Font) []string {
	k := layoutKey{s.Text, f, s.WrapLen}
	if s.linesKey == k && s.lines != nil {
		return s.lines
	}
	switch {
	case s.Text == "":
		s.lines = []string{""}
	case s.WrapLen <= 0 || f == nil:
		s.lines = strings.Split(s.Text, "\n")
	default:
		s.lines = font.WrapLines(f, s.Text, s.WrapLen)
	}
	s.linesKey = k
	return s.lines
}

// ComputeGeometry ports TkpComputeButtonGeometry: it lays out the text,
// sizes the content from the image, the text and -width/-height, leaves
// room for the indicator, the padding, the border, the highlight and the
// default ring, and sets the window's request.
func ComputeGeometry(b *widget.Base, s *Shared) {
	inset := Inset(b, s)
	s.indicatorSpace, s.indicatorDiameter = 0, 0
	s.textWidth, s.textHeight = 0, 0

	var width, height, avg, linespace int
	haveImage := s.Img != nil
	if haveImage {
		width, height = s.Img.Width(), s.Img.Height()
	}
	haveText := false
	if (!haveImage || s.Compound != widget.CompoundNone) && b.Font != nil {
		linespace = b.Font.Metrics().Linespace()
		lines := s.TextLines(b.Font)
		s.textHeight = len(lines) * linespace
		for _, line := range lines {
			s.textWidth = max(s.textWidth, font.TextWidth(b.Font, line))
		}
		avg = b.Font.MeasureString("0")
		haveText = s.textWidth != 0 && s.textHeight != 0
	}
	indicator := s.typ >= TypeCheck && s.IndicatorOn
	indicatorFor := func(height int) {
		if !indicator {
			return
		}
		s.indicatorSpace = height
		if s.typ == TypeCheck {
			s.indicatorDiameter = 65 * height / 100
		} else {
			s.indicatorDiameter = 75 * height / 100
		}
	}

	switch {
	case s.Compound != widget.CompoundNone && haveImage && haveText:
		switch s.Compound {
		case widget.CompoundTop, widget.CompoundBottom:
			height += s.textHeight + b.PadY
			width = max(width, s.textWidth)
		case widget.CompoundLeft, widget.CompoundRight:
			width += s.textWidth + b.PadX
			height = max(height, s.textHeight)
		default:
			width = max(width, s.textWidth)
			height = max(height, s.textHeight)
		}
		if s.WidthChars > 0 {
			width = s.WidthChars
		}
		if s.HeightChars > 0 {
			height = s.HeightChars
		}
		indicatorFor(height)
		width += 2 * b.PadX
		height += 2 * b.PadY
	case haveImage:
		if s.WidthChars > 0 {
			width = s.WidthChars
		}
		if s.HeightChars > 0 {
			height = s.HeightChars
		}
		indicatorFor(height)
	default:
		width, height = s.textWidth, s.textHeight
		if s.WidthChars > 0 {
			width = s.WidthChars * avg
		}
		if s.HeightChars > 0 {
			height = s.HeightChars * linespace
		}
		if indicator {
			s.indicatorDiameter = linespace
			s.indicatorSpace = linespace + avg
		}
		width += 2 * b.PadX
		height += 2 * b.PadY
	}
	// Two extra pixels so a push button's content can shift by one for
	// the raised and sunken looks.
	if s.typ == TypeButton {
		width += 2
		height += 2
	}
	b.Win.ReqWidth = width + s.indicatorSpace + 2*inset
	b.Win.ReqHeight = height + 2*inset
}

// Display ports TkpDisplayButton: the background for the state, the
// content (image, text or both), the indicator, then the border, the
// default ring and the focus highlight on top. It draws into the window's
// current drawable, which widget.Base.EventuallyRedraw points at a pixmap.
func Display(b *widget.Base, s *Shared, sel Selection) {
	w := b.Win
	if w.PlatformID == 0 {
		return
	}
	d := w.Display.Server
	gc := w.GC
	drawable := w.Drawable()

	// Colours and border for the state.
	border := b.Border
	var bgPixel uint64
	if b.Background != nil {
		bgPixel = b.Background.Pixel
	}
	var fg *color.ColorRef
	if b.Foreground != nil {
		fg = b.Foreground.Ref()
	}
	switch {
	case s.State == widget.StateDisabled && s.DisabledFg != nil:
		fg = s.DisabledFg
	case s.State == widget.StateActive:
		if s.ActiveForeground != nil {
			fg = s.ActiveForeground
		}
		if s.ActiveBackground != nil {
			bgPixel = s.ActiveBackground.Pixel
			border = borderOf(s.ActiveBackground)
		}
	default:
	}
	if sel.Selected && s.SelectColor != nil && !s.IndicatorOn && s.typ >= TypeCheck {
		bgPixel = s.SelectColor.Pixel
		border = borderOf(s.SelectColor)
	}
	if border == nil {
		border = draw.NewBorderFromPixel(bgPixel)
	}

	// A check or radio button without an indicator shows its selection as
	// its relief; a push button sinks while pressed (button.tcl).
	relief := b.Relief
	switch {
	case s.typ == TypeButton && s.pressed:
		relief = option.ReliefSunken
	case s.typ >= TypeCheck && !s.IndicatorOn:
		if sel.Selected {
			relief = option.ReliefSunken
		} else {
			relief = s.OffRelief
		}
	}

	d.SetForeground(gc, bgPixel)
	d.FillRectangle(drawable, gc, 0, 0, uint(w.Width), uint(w.Height))

	// Content: image and text, text or image, anchored in the window
	// beside the indicator's column.
	inset := Inset(b, s)
	img := s.Img
	if sel.Selected && s.SelectImg != nil {
		img = s.SelectImg
	}
	haveImage := s.Img != nil
	haveText := s.textWidth != 0 && s.textHeight != 0
	var imgW, imgH int
	if haveImage {
		imgW, imgH = s.Img.Width(), s.Img.Height()
	}
	drawImage := func(x, y int) {
		// Keep the coordinates inside the window (Tk bug 979239).
		width, height := min(imgW, w.Width), min(imgH, w.Height)
		x = max(min(x, w.Width-width), 0)
		y = max(min(y, w.Height-height), 0)
		img.Draw(d, drawable, gc, w.Depth, 0, 0, width, height, x, y, bgPixel)
	}
	var x, y int
	switch {
	case s.Compound != widget.CompoundNone && haveImage && haveText:
		var textX, textY, imgX, imgY, fullW, fullH int
		switch s.Compound {
		case widget.CompoundTop, widget.CompoundBottom:
			if s.Compound == widget.CompoundTop {
				textY = imgH + b.PadY
			} else {
				imgY = s.textHeight + b.PadY
			}
			fullH = imgH + s.textHeight + b.PadY
			fullW = max(imgW, s.textWidth)
			textX = (fullW - s.textWidth) / 2
			imgX = (fullW - imgW) / 2
		case widget.CompoundLeft, widget.CompoundRight:
			if s.Compound == widget.CompoundLeft {
				textX = imgW + b.PadX
			} else {
				imgX = s.textWidth + b.PadX
			}
			fullW = s.textWidth + b.PadX + imgW
			fullH = max(imgH, s.textHeight)
			textY = (fullH - s.textHeight) / 2
			imgY = (fullH - imgH) / 2
		default:
			fullW = max(imgW, s.textWidth)
			fullH = max(imgH, s.textHeight)
			textX = (fullW - s.textWidth) / 2
			imgX = (fullW - imgW) / 2
			textY = (fullH - s.textHeight) / 2
			imgY = (fullH - imgH) / 2
		}
		x, y = widget.ComputeAnchor(s.Anchor, w.Width, w.Height, inset, b.PadX, b.PadY, s.indicatorSpace+fullW, fullH)
		x += s.indicatorSpace
		x, y = s.shift(w.Width, w.Height, relief, x, y, imgW, imgH)
		drawImage(x+imgX, y+imgY)
		s.drawText(b, drawable, x+textX, y+textY, fg)
	case haveImage:
		x, y = widget.ComputeAnchor(s.Anchor, w.Width, w.Height, inset, 0, 0, s.indicatorSpace+imgW, imgH)
		x += s.indicatorSpace
		x, y = s.shift(w.Width, w.Height, relief, x, y, imgW, imgH)
		drawImage(x, y)
	default:
		x, y = widget.ComputeAnchor(s.Anchor, w.Width, w.Height, inset, b.PadX, b.PadY, s.indicatorSpace+s.textWidth, s.textHeight)
		x += s.indicatorSpace
		x, y = s.shift(w.Width, w.Height, relief, x, y, 0, 0)
		s.drawText(b, drawable, x, y, fg)
	}

	// The indicator, centred in its column and on the window's mid-line.
	if s.typ >= TypeCheck && s.IndicatorOn && s.indicatorDiameter > 2*b.BorderWidth {
		kind := draw.CheckIndicator
		if s.typ == TypeRadio {
			kind = draw.RadioIndicator
		}
		state := draw.IndicatorOff
		switch {
		case sel.Tristate:
			state = draw.IndicatorTristate
		case sel.Selected:
			state = draw.IndicatorOn
		}
		var selPixel, fgPixel, disPixel uint64 = 0xffffff, 0, 0xa3a3a3
		if s.SelectColor != nil {
			selPixel = s.SelectColor.Pixel
		}
		if b.Foreground != nil {
			fgPixel = b.Foreground.Pixel
		}
		if s.DisabledFg != nil {
			disPixel = s.DisabledFg.Pixel
		}
		draw.DrawCheckIndicator(d, drawable, gc, w.Depth,
			x-s.indicatorSpace/2, w.Height/2, kind,
			draw.NewBorderFromPixel(bgPixel), fgPixel, selPixel, disPixel,
			state, s.State == widget.StateDisabled)
	}

	// The border and the rings last, over any overflowing content. The
	// default ring is drawn in the highlight border's colour.
	if relief != option.ReliefFlat {
		in := b.HighlightWidth
		ring := b.Border
		if b.HighlightBackground != nil {
			ring = borderOf(b.HighlightBackground.Ref())
		}
		flatRing := func(in, bw int) {
			if ring == nil {
				return
			}
			d.SetForeground(gc, ring.BgPixel)
			for i := range bw {
				d.DrawRectangle(drawable, gc, in+i, in+i,
					uint(w.Width-2*(in+i)-1), uint(w.Height-2*(in+i)-1))
			}
		}
		switch s.Default {
		case DefaultActive:
			// 2px of space, a 1px sunken ring, 2px of space.
			flatRing(in, 2)
			in += 2
			draw.Draw3DRectangle(d, drawable, gc, ring, in, in, w.Width-2*in, w.Height-2*in, 1, option.ReliefSunken)
			in++
			flatRing(in, 2)
			in += 2
		case DefaultNormal:
			flatRing(0, 5)
			in += 5
		default:
		}
		draw.Draw3DRectangle(d, drawable, gc, border, in, in, w.Width-2*in, w.Height-2*in, b.BorderWidth, relief)
	}
	// The focus ring shrink-wraps the button, not the default ring's room.
	focusPad := 0
	if s.Default == DefaultNormal {
		focusPad = 5
	}
	b.DrawHighlightBorder(s.HasFocus, focusPad)
}

// shift ports ShiftByOffset: a push button's content moves down and right
// as its relief sinks, so it appears to move with the button.
func (s *Shared) shift(winW, winH int, relief option.Relief, x, y, width, height int) (int, int) {
	if s.typ != TypeButton || relief == option.ReliefRaised {
		return x, y
	}
	sx, sy := 1, 1
	if relief == option.ReliefSunken {
		sx, sy = 2, 2
	}
	if relief != option.ReliefRidge {
		// One pixel back when the padding is even, or the content lands
		// too far right and down.
		if (winW-width)%2 == 0 {
			sx--
		}
		if (winH-height)%2 == 0 {
			sy--
		}
	}
	return x + sx, y + sy
}

// drawText draws the laid-out lines with their top-left corner at (x, y),
// justified within the text's width (Tk_DrawTextLayout), and underlines
// the -underline character (Tk_UnderlineTextLayout).
func (s *Shared) drawText(b *widget.Base, drawable platform.DrawableID, x, y int, fg *color.ColorRef) {
	if b.Font == nil || fg == nil {
		return
	}
	df, ok := b.Font.(platform.DrawableFont)
	if !ok {
		return
	}
	w := b.Win
	d := w.Display.Server
	m := b.Font.Metrics()
	lines := s.TextLines(b.Font)
	underline := s.Underline
	for i, line := range lines {
		baseline := y + m.Ascent + i*m.Linespace()
		lx := x
		if len(lines) > 1 {
			lw := font.TextWidth(b.Font, line)
			switch s.Justify {
			case option.JustifyCenter:
				lx += (s.textWidth - lw) / 2
			case option.JustifyRight:
				lx += s.textWidth - lw
			default:
			}
		}
		for _, seg := range font.Segments(b.Font, line) {
			df.DrawString(drawable, lx+seg.X, baseline, seg.Text, fg.Pixel, fg.Red, fg.Green, fg.Blue)
		}
		if n := utf8.RuneCountInString(line); underline >= 0 && underline < n {
			ch := []rune(line)
			ux := lx + font.TextWidth(b.Font, string(ch[:underline]))
			uw := b.Font.MeasureString(string(ch[underline]))
			pos, height := font.Underline(b.Font)
			d.SetForeground(w.GC, fg.Pixel)
			d.FillRectangle(drawable, w.GC, ux, baseline+pos, uint(uw), uint(height))
			underline = -1
		} else if underline >= 0 {
			underline -= n + 1 // the newline counts as a character
		}
	}
}

// borderOf makes a 3D border shaded from c.
func borderOf(c *color.ColorRef) *draw.Border {
	return draw.NewBorder(c.Red, c.Green, c.Blue)
}
