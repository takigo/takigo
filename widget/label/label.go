// Package label implements the label widget, which displays a text string
// or image. It ports the label-specific parts of tk/generic/tkButton.c.
package label

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Label displays a text string with optional border.
type Label struct {
	widget.Base

	Text      string
	Anchor    option.Anchor
	Justify   option.Justify
	WrapLen   int // wrap length in pixels (0 = no wrap)
	Underline int // index of character to underline (-1 = none)

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

// Width sets the requested width (in characters, approximately).
func Width(w int) LabelOption {
	return func(l *Label) { l.Win.ReqWidth = w }
}

// Height sets the requested height (in lines, approximately).
func Height(h int) LabelOption {
	return func(l *Label) { l.Win.ReqHeight = h }
}

// New creates a new Label widget as a child of parent.
func New(parent *window.Window, name string, app widget.AppContext, opts ...LabelOption) *Label {
	w := window.NewChildWindow(parent, name, 0, 0, 1, 1)
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

// computeGeometry computes the text size and sets the requested window size.
func (l *Label) computeGeometry() {
	if l.Font == nil {
		return
	}
	l.textWidth = l.Font.MeasureString(l.Text)
	m := l.Font.Metrics()
	l.textHeight = m.Linespace()

	inset := l.BorderWidth + l.HighlightWidth
	w := l.Win
	w.ReqWidth = l.textWidth + 2*l.PadX + 2*inset
	w.ReqHeight = l.textHeight + 2*l.PadY + 2*inset
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

	// Draw text.
	if l.Font != nil && l.Text != "" && l.Foreground != nil {
		inset := l.BorderWidth + l.HighlightWidth
		availW := w.Width - 2*inset - 2*l.PadX
		availH := w.Height - 2*inset - 2*l.PadY

		// Compute text position based on anchor.
		textX, textY := anchorText(l.Anchor,
			inset+l.PadX, inset+l.PadY,
			availW, availH,
			l.textWidth, l.textHeight)

		// Baseline is textY + ascent.
		m := l.Font.Metrics()
		baseline := textY + m.Ascent

		if xftFont, ok := l.Font.(*font.XftFont); ok {
			xftFont.DrawString(w.Drawable(), textX, baseline, l.Text,
				l.Foreground.Pixel, l.Foreground.Red, l.Foreground.Green, l.Foreground.Blue)
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
