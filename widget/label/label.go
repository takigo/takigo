// Package label implements the label widget, which displays a text string
// or image. It ports the label-specific parts of tk/generic/tkButton.c.
package label

import (
	"strings"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Label displays a text string, image, or both with optional border.
type Label struct {
	widget.Base

	Text      string
	Anchor    option.Anchor
	Justify   option.Justify
	WrapLen   int // wrap length in pixels (0 = no wrap)
	Underline int // index of character to underline (-1 = none)

	// Image support.
	Img      widget.WidgetImage
	Compound widget.Compound

	textWidth  int
	textHeight int
}

// LabelOption configures a Label.
type LabelOption func(*Label)

// Text sets the label text.
func Text(s string) LabelOption {
	return func(l *Label) { l.Text = s }
}

// Background sets the background color.
func Background(name string) LabelOption {
	return func(l *Label) {
		col, err := l.App.ColorCache().Get(name)
		if err == nil {
			l.Background = col
			l.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) LabelOption {
	return func(l *Label) {
		col, err := l.App.ColorCache().Get(name)
		if err == nil {
			l.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) LabelOption {
	return func(l *Label) {
		f, err := l.App.FontRegistry().Get(name)
		if err == nil {
			l.Font = f
		}
	}
}

// BorderWidth sets the border width.
func BorderWidth(w int) LabelOption {
	return func(l *Label) { l.BorderWidth = w }
}

// Relief sets the border relief.
func Relief(r option.Relief) LabelOption {
	return func(l *Label) { l.Relief = r }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) LabelOption {
	return func(l *Label) { l.Anchor = a }
}

// PadX sets horizontal padding.
func PadX(p int) LabelOption {
	return func(l *Label) { l.PadX = p }
}

// PadY sets vertical padding.
func PadY(p int) LabelOption {
	return func(l *Label) { l.PadY = p }
}

// ImageOpt sets the image to display.
func ImageOpt(img widget.WidgetImage) LabelOption {
	return func(l *Label) { l.Img = img }
}

// CompoundOpt sets how text and image are combined.
func CompoundOpt(c widget.Compound) LabelOption {
	return func(l *Label) { l.Compound = c }
}

// Width sets the requested width (in characters, approximately).
func Width(w int) LabelOption {
	return func(l *Label) { l.Win.ReqWidth = w }
}

// Height sets the requested height (in lines, approximately).
func Height(h int) LabelOption {
	return func(l *Label) { l.Win.ReqHeight = h }
}

// New creates a new Label widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...LabelOption) *Label {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	l := &Label{
		Anchor:    option.AnchorCenter,
		Justify:   option.JustifyLeft,
		Underline: -1,
	}
	widget.InitBase(&l.Base, w, app)

	// Label-specific defaults from tkUnixDefault.h.
	l.BorderWidth = 1
	l.Relief = option.ReliefFlat
	l.PadX = 1
	l.PadY = 1

	for _, opt := range opts {
		opt(l)
	}

	// Compute geometry.
	l.computeGeometry()

	// Update X window background.
	if l.Background != nil {
		w.BackgroundPixel = l.Background.Pixel
	}

	// Bind events.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		l.Display()
	})

	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			l.Display()
		}
	})

	return l
}

// computeGeometry computes the text/image size and sets the requested window size.
func (l *Label) computeGeometry() {
	// Measure text, handling multiline (newline-separated).
	if l.Font != nil && l.Text != "" {
		m := l.Font.Metrics()
		lines := strings.Split(l.Text, "\n")
		l.textHeight = len(lines) * m.Linespace()
		l.textWidth = 0
		for _, line := range lines {
			w := l.Font.MeasureString(line)
			if w > l.textWidth {
				l.textWidth = w
			}
		}
	} else {
		l.textWidth = 0
		l.textHeight = 0
	}

	contentW, contentH := compoundSize(l.Compound, l.Img, l.textWidth, l.textHeight)

	inset := l.BorderWidth + l.HighlightWidth
	w := l.Win
	w.ReqWidth = contentW + 2*l.PadX + 2*inset
	w.ReqHeight = contentH + 2*l.PadY + 2*inset
}

// Display draws the label.
func (l *Label) Display() {
	if l.Destroyed {
		return
	}
	w := l.Win
	if w.XWindow == xlib.Window(0) {
		return
	}

	d := w.Display.XDisplay
	gc := w.GC

	// Fill background.
	if l.Background != nil {
		d.SetForeground(gc, l.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border.
	if l.Border != nil && l.BorderWidth > 0 && l.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, w.Drawable(), gc, l.Border,
			0, 0, w.Width, w.Height, l.BorderWidth, l.Relief)
	}

	// Draw content (image and/or text).
	inset := l.BorderWidth + l.HighlightWidth
	availW := w.Width - 2*inset - 2*l.PadX
	availH := w.Height - 2*inset - 2*l.PadY
	frameX := inset + l.PadX
	frameY := inset + l.PadY

	hasImg := l.Img != nil
	hasText := l.Font != nil && l.Text != "" && l.Foreground != nil

	bgPixel := uint64(0)
	if l.Background != nil {
		bgPixel = l.Background.Pixel
	}

	if hasImg && hasText && l.Compound != widget.CompoundNone {
		drawCompound(l, d, w, frameX, frameY, availW, availH, bgPixel)
	} else if hasImg {
		// Image only.
		imgW := l.Img.Width()
		imgH := l.Img.Height()
		ix, iy := anchorText(l.Anchor, frameX, frameY, availW, availH, imgW, imgH)
		l.Img.Draw(w.Display.XDisplay, w.Drawable(), gc,
			w.Visual, w.Depth, 0, 0, imgW, imgH, ix, iy, bgPixel)
	} else if hasText {
		// Text only — handle multiline.
		textX, textY := anchorText(l.Anchor, frameX, frameY,
			availW, availH, l.textWidth, l.textHeight)
		m := l.Font.Metrics()
		if xftFont, ok := l.Font.(*font.XftFont); ok {
			lines := strings.Split(l.Text, "\n")
			for i, line := range lines {
				baseline := textY + m.Ascent + i*m.Linespace()
				lx := textX
				// Apply justify for multiline.
				if len(lines) > 1 {
					lw := l.Font.MeasureString(line)
					switch l.Justify {
					case option.JustifyCenter:
						lx = textX + (l.textWidth-lw)/2
					case option.JustifyRight:
						lx = textX + l.textWidth - lw
					}
				}
				xftFont.DrawString(w.Drawable(), lx, baseline, line,
					l.Foreground.Pixel, l.Foreground.Red, l.Foreground.Green, l.Foreground.Blue)
			}
		}
	}

	d.Flush()
}

// anchorText computes the x,y position for text within a frame.
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
		w := imgW + 4 + textW // 4px gap
		return w, max(imgH, textH)
	case widget.CompoundTop, widget.CompoundBottom:
		return max(imgW, textW), imgH + 4 + textH
	case widget.CompoundCenter:
		return max(imgW, textW), max(imgH, textH)
	default: // CompoundNone — show image only when both present
		return imgW, imgH
	}
}

// drawCompound draws image and text in compound mode.
func drawCompound(l *Label, _ *xlib.Display, w *window.Window,
	frameX, frameY, availW, availH int, bgPixel uint64) {

	imgW := l.Img.Width()
	imgH := l.Img.Height()
	contentW, contentH := compoundSize(l.Compound, l.Img, l.textWidth, l.textHeight)

	// Anchor the content block.
	cx, cy := anchorText(l.Anchor, frameX, frameY, availW, availH, contentW, contentH)

	var imgX, imgY, textX, textY int
	switch l.Compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + 4
		textY = cy + (contentH-l.textHeight)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-l.textHeight)/2
		imgX = cx + l.textWidth + 4
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-l.textWidth)/2
		textY = cy + imgH + 4
	case widget.CompoundBottom:
		textX = cx + (contentW-l.textWidth)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + l.textHeight + 4
	case widget.CompoundCenter:
		imgX = cx + (contentW-imgW)/2
		imgY = cy + (contentH-imgH)/2
		textX = cx + (contentW-l.textWidth)/2
		textY = cy + (contentH-l.textHeight)/2
	}

	// Draw image.
	l.Img.Draw(w.Display.XDisplay, w.Drawable(), w.GC,
		w.Visual, w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)

	// Draw text.
	if l.Font != nil && l.Foreground != nil {
		m := l.Font.Metrics()
		baseline := textY + m.Ascent
		if xftFont, ok := l.Font.(*font.XftFont); ok {
			xftFont.DrawString(w.Drawable(), textX, baseline, l.Text,
				l.Foreground.Pixel, l.Foreground.Red, l.Foreground.Green, l.Foreground.Blue)
		}
	}
}

// Configure applies options to the label.
func (l *Label) Configure(opts ...option.Option) {
	option.Apply(l, opts)
	l.UpdateBorder()
	l.computeGeometry()
	if l.Background != nil {
		l.Win.BackgroundPixel = l.Background.Pixel
	}
	l.Display()
}

// Destroy cleans up the label.
func (l *Label) Destroy() {
	if l.Destroyed {
		return
	}
	l.Destroyed = true
	window.DestroyWindow(l.Win)
}
