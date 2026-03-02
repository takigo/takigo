// Package button implements the button widget with press/hover interaction.
// It ports the button-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package button

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Button is an interactive widget that invokes a command when clicked.
type Button struct {
	widget.Base

	Text      string
	Command   func() // invoked on click
	Anchor    option.Anchor
	Justify   option.Justify
	Underline int // index of char to underline (-1 = none)

	// Image support.
	Img      widget.WidgetImage
	Compound widget.Compound

	// State.
	State     widget.State
	OverRelief option.Relief // relief when mouse is over button
	OffRelief  option.Relief // relief when not pressed

	// Active colors (used on hover).
	ActiveBackground *colorRef
	ActiveForeground *colorRef

	textWidth  int
	textHeight int
	pressed    bool // button1 is held down
}

// colorRef holds a resolved color reference.
type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// ButtonOption configures a Button.
type ButtonOption func(*Button)

// Text sets the button text.
func Text(s string) ButtonOption {
	return func(b *Button) { b.Text = s }
}

// Command sets the callback invoked when the button is clicked.
func Command(fn func()) ButtonOption {
	return func(b *Button) { b.Command = fn }
}

// Background sets the background color.
func Background(name string) ButtonOption {
	return func(b *Button) {
		col, err := b.App.ColorCache().Get(name)
		if err == nil {
			b.Background = col
			b.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) ButtonOption {
	return func(b *Button) {
		col, err := b.App.ColorCache().Get(name)
		if err == nil {
			b.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) ButtonOption {
	return func(b *Button) {
		f, err := b.App.FontRegistry().Get(name)
		if err == nil {
			b.Font = f
		}
	}
}

// ImageOpt sets the image to display.
func ImageOpt(img widget.WidgetImage) ButtonOption {
	return func(b *Button) { b.Img = img }
}

// CompoundOpt sets how text and image are combined.
func CompoundOpt(c widget.Compound) ButtonOption {
	return func(b *Button) { b.Compound = c }
}

// BorderWidth sets the border width.
func BorderWidth(w int) ButtonOption {
	return func(b *Button) { b.BorderWidth = w }
}

// ReliefOpt sets the border relief.
func ReliefOpt(r option.Relief) ButtonOption {
	return func(b *Button) { b.Relief = r }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) ButtonOption {
	return func(b *Button) { b.Anchor = a }
}

// PadX sets horizontal padding.
func PadX(p int) ButtonOption {
	return func(b *Button) { b.PadX = p }
}

// PadY sets vertical padding.
func PadY(p int) ButtonOption {
	return func(b *Button) { b.PadY = p }
}

// New creates a new Button widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...ButtonOption) *Button {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	b := &Button{
		Anchor:    option.AnchorCenter,
		Justify:   option.JustifyCenter,
		Underline: -1,
		OverRelief: option.ReliefRaised,
		OffRelief:  option.ReliefFlat,
	}
	widget.InitBase(&b.Base, w, app)

	// Button-specific defaults.
	b.BorderWidth = widget.DefBorderWidth
	b.Relief = option.ReliefRaised
	b.PadX = 3
	b.PadY = 1
	b.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		b.ActiveBackground = &colorRef{ac.Pixel, ac.Red, ac.Green, ac.Blue}
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		b.ActiveForeground = &colorRef{af.Pixel, af.Red, af.Green, af.Blue}
	}

	for _, opt := range opts {
		opt(b)
	}

	// Compute geometry.
	b.computeGeometry()

	// Update X window background.
	if b.Background != nil {
		w.BackgroundPixel = b.Background.Pixel
	}

	// Bind events.
	bindButton(b, app)

	return b
}

// computeGeometry computes text/image size and sets requested window size.
func (b *Button) computeGeometry() {
	if b.Font != nil && b.Text != "" {
		b.textWidth = b.Font.MeasureString(b.Text)
		m := b.Font.Metrics()
		b.textHeight = m.Linespace()
	} else {
		b.textWidth = 0
		b.textHeight = 0
	}

	contentW, contentH := compoundSize(b.Compound, b.Img, b.textWidth, b.textHeight)

	inset := b.BorderWidth + b.HighlightWidth
	w := b.Win
	w.ReqWidth = contentW + 2*b.PadX + 2*inset
	w.ReqHeight = contentH + 2*b.PadY + 2*inset
}

// Display draws the button.
func (b *Button) Display() {
	if b.Destroyed {
		return
	}
	w := b.Win
	if w.XWindow == xlib.Window(0) {
		return
	}

	d := w.Display.XDisplay
	gc := w.GC

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *colorRef
	if b.Background != nil {
		bgPixel = b.Background.Pixel
	}
	if b.Foreground != nil {
		fgCol = &colorRef{b.Foreground.Pixel, b.Foreground.Red, b.Foreground.Green, b.Foreground.Blue}
	}

	if b.State == widget.StateActive && b.ActiveBackground != nil {
		bgPixel = b.ActiveBackground.Pixel
	}
	if b.State == widget.StateActive && b.ActiveForeground != nil {
		fgCol = b.ActiveForeground
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Determine relief.
	relief := b.Relief
	if b.pressed {
		relief = option.ReliefSunken
	}

	// Draw border.
	border := b.Border
	if b.State == widget.StateActive && b.Background != nil {
		// Recompute border from active background for proper shading.
		if b.ActiveBackground != nil {
			border = draw.NewBorder(b.ActiveBackground.Red, b.ActiveBackground.Green, b.ActiveBackground.Blue)
		}
	}
	if border != nil && b.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			0, 0, w.Width, w.Height, b.BorderWidth, relief)
	}

	// Draw content (image and/or text).
	inset := b.BorderWidth + b.HighlightWidth
	availW := w.Width - 2*inset - 2*b.PadX
	availH := w.Height - 2*inset - 2*b.PadY
	frameX := inset + b.PadX
	frameY := inset + b.PadY

	// Shift content 1px down-right when pressed (Tk behavior).
	pressOff := 0
	if b.pressed {
		pressOff = 1
	}

	hasImg := b.Img != nil
	hasText := b.Font != nil && b.Text != "" && fgCol != nil

	if hasImg && hasText && b.Compound != widget.CompoundNone {
		drawCompoundButton(b, w, frameX, frameY, availW, availH, bgPixel, fgCol, pressOff)
	} else if hasImg {
		imgW := b.Img.Width()
		imgH := b.Img.Height()
		ix, iy := anchorText(b.Anchor, frameX, frameY, availW, availH, imgW, imgH)
		b.Img.Draw(w.Display.XDisplay, w.Drawable(), gc,
			w.Visual, w.Depth, 0, 0, imgW, imgH, ix+pressOff, iy+pressOff, bgPixel)
	} else if hasText {
		textX, textY := anchorText(b.Anchor, frameX, frameY,
			availW, availH, b.textWidth, b.textHeight)
		textX += pressOff
		textY += pressOff
		m := b.Font.Metrics()
		baseline := textY + m.Ascent
		if xftFont, ok := b.Font.(*font.XftFont); ok {
			xftFont.DrawString(w.Drawable(), textX, baseline, b.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	d.Flush()
}

// anchorText computes x,y for text within a frame.
func anchorText(a option.Anchor, frameX, frameY, frameW, frameH, textW, textH int) (int, int) {
	var x, y int
	switch a {
	case option.AnchorNW:
		x, y = frameX, frameY
	case option.AnchorN:
		x, y = frameX+(frameW-textW)/2, frameY
	case option.AnchorNE:
		x, y = frameX+frameW-textW, frameY
	case option.AnchorW:
		x, y = frameX, frameY+(frameH-textH)/2
	case option.AnchorCenter:
		x, y = frameX+(frameW-textW)/2, frameY+(frameH-textH)/2
	case option.AnchorE:
		x, y = frameX+frameW-textW, frameY+(frameH-textH)/2
	case option.AnchorSW:
		x, y = frameX, frameY+frameH-textH
	case option.AnchorS:
		x, y = frameX+(frameW-textW)/2, frameY+frameH-textH
	case option.AnchorSE:
		x, y = frameX+frameW-textW, frameY+frameH-textH
	}
	return x, y
}

// compoundSize computes the total content size for a compound image+text layout.
func compoundSize(c widget.Compound, img widget.WidgetImage, textW, textH int) (int, int) {
	if img == nil {
		return textW, textH
	}
	imgW := img.Width()
	imgH := img.Height()

	if textW == 0 && textH == 0 {
		return imgW, imgH
	}

	switch c {
	case widget.CompoundLeft, widget.CompoundRight:
		return imgW + 4 + textW, max(imgH, textH)
	case widget.CompoundTop, widget.CompoundBottom:
		return max(imgW, textW), imgH + 4 + textH
	case widget.CompoundCenter:
		return max(imgW, textW), max(imgH, textH)
	default:
		return imgW, imgH
	}
}

// drawCompoundButton draws image and text in compound mode for a button.
func drawCompoundButton(b *Button, w *window.Window,
	frameX, frameY, availW, availH int, bgPixel uint64,
	fgCol *colorRef, pressOff int) {

	imgW := b.Img.Width()
	imgH := b.Img.Height()
	contentW, contentH := compoundSize(b.Compound, b.Img, b.textWidth, b.textHeight)

	cx, cy := anchorText(b.Anchor, frameX, frameY, availW, availH, contentW, contentH)
	cx += pressOff
	cy += pressOff

	var imgX, imgY, textX, textY int
	switch b.Compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + 4
		textY = cy + (contentH-b.textHeight)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-b.textHeight)/2
		imgX = cx + b.textWidth + 4
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-b.textWidth)/2
		textY = cy + imgH + 4
	case widget.CompoundBottom:
		textX = cx + (contentW-b.textWidth)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + b.textHeight + 4
	case widget.CompoundCenter:
		imgX = cx + (contentW-imgW)/2
		imgY = cy + (contentH-imgH)/2
		textX = cx + (contentW-b.textWidth)/2
		textY = cy + (contentH-b.textHeight)/2
	}

	// Draw image.
	b.Img.Draw(w.Display.XDisplay, w.Drawable(), w.GC,
		w.Visual, w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)

	// Draw text.
	if b.Font != nil && fgCol != nil {
		m := b.Font.Metrics()
		baseline := textY + m.Ascent
		if xftFont, ok := b.Font.(*font.XftFont); ok {
			xftFont.DrawString(w.Drawable(), textX, baseline, b.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}
}

// Invoke executes the button's command.
func (b *Button) Invoke() {
	if b.State == widget.StateDisabled {
		return
	}
	if b.Command != nil {
		b.Command()
	}
}

// Configure applies options to the button.
func (b *Button) Configure(opts ...option.Option) {
	option.Apply(b, opts)
	b.UpdateBorder()
	b.computeGeometry()
	if b.Background != nil {
		b.Win.BackgroundPixel = b.Background.Pixel
	}
	b.Display()
}

// Destroy cleans up the button.
func (b *Button) Destroy() {
	if b.Destroyed {
		return
	}
	b.Destroyed = true
	window.DestroyWindow(b.Win)
}
