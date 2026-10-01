// Package message implements the message widget, which displays multi-line
// text with automatic word-wrapping based on an aspect ratio.
// It ports tk/generic/tkMessage.c.
package message

import (
	"strings"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Message displays multi-line text with automatic word wrapping based on
// an aspect ratio (100*width/height). Unlike Label, Message automatically
// determines line breaks to achieve the desired aspect ratio.
type Message struct {
	widget.Base

	Text    string
	Anchor  option.Anchor
	Justify option.Justify
	Aspect  int // desired aspect ratio: 100*width/height (default 150)
	Width   int // explicit width in pixels (0 = auto from aspect)

	// Computed layout.
	lines     []string
	msgWidth  int // computed text width
	msgHeight int // computed text height
}

// MessageOption configures a Message.
type MessageOption func(*Message)

// Text sets the message text.
func Text(s string) MessageOption {
	return func(m *Message) { m.Text = s }
}

// Background sets the background color.
func Background(name string) MessageOption {
	return func(m *Message) { m.SetBackgroundName(name) }
}

// Foreground sets the text color.
func Foreground(name string) MessageOption {
	return func(m *Message) { m.SetForegroundName(name) }
}

// FontOpt sets the font.
func FontOpt(name string) MessageOption {
	return func(m *Message) { m.SetFontName(name) }
}

// BorderWidth sets the border width.
func BorderWidth[L screenunit.Length](w L) MessageOption {
	return func(m *Message) { m.BorderWidth = screenunit.ToPixels(w) }
}

// Relief sets the border relief.
func Relief(r option.Relief) MessageOption {
	return func(m *Message) { m.Relief = r }
}

// Anchor sets the text anchor within the widget.
func Anchor(a option.Anchor) MessageOption {
	return func(m *Message) { m.Anchor = a }
}

// JustifyOpt sets the text justification for lines.
func JustifyOpt(j option.Justify) MessageOption {
	return func(m *Message) { m.Justify = j }
}

// Aspect sets the desired aspect ratio (100*width/height).
// Default is 150 (width is 1.5x height).
func Aspect(a int) MessageOption {
	return func(m *Message) { m.Aspect = a }
}

// WidthOpt sets the explicit width in pixels (overrides aspect ratio).
// Accepts int (pixels), float64, or string with unit suffix.
// Set to 0 (default) to use aspect-ratio-based auto-width.
func WidthOpt[L screenunit.Length](w L) MessageOption {
	return func(m *Message) { m.Width = screenunit.ToPixels(w) }
}

// PadX sets horizontal padding.
func PadX[L screenunit.Length](p L) MessageOption {
	return func(m *Message) { m.PadX = screenunit.ToPixels(p) }
}

// PadY sets vertical padding.
func PadY[L screenunit.Length](p L) MessageOption {
	return func(m *Message) { m.PadY = screenunit.ToPixels(p) }
}

// HighlightWidth sets the focus highlight border width.
func HighlightWidth(w int) MessageOption {
	return func(m *Message) { m.HighlightWidth = w }
}

// New creates a new Message widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...MessageOption) *Message {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	m := &Message{
		Anchor:  option.AnchorCenter,
		Justify: option.JustifyLeft,
		Aspect:  150,
	}
	widget.InitBase(&m.Base, w, app)
	m.SetDisplayProc(m.display)
	w.Class = "Message"

	// Message-specific defaults (from tkUnixDefault.h).
	m.BorderWidth = 1
	m.Relief = option.ReliefFlat

	// Default padding based on font metrics (like Tk).
	if m.Font != nil {
		metrics := m.Font.Metrics()
		m.PadX = metrics.Ascent / 2
		m.PadY = metrics.Ascent / 4
	}

	for _, opt := range opts {
		opt(m)
	}

	m.computeGeometry()

	if m.Background != nil {
		w.SetBackgroundPixel(m.Background.Pixel)
	}

	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		m.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			m.Display()
		}
	})

	return m
}

// wrapText wraps text to fit within maxWidth pixels, returning lines and the
// actual width used.
func (m *Message) wrapText(maxWidth int) ([]string, int, int) {
	if m.Text == "" || m.Font == nil {
		return nil, 0, 0
	}

	paragraphs := strings.Split(m.Text, "\n")
	var result []string
	actualWidth := 0

	metrics := m.Font.Metrics()

	// Break decisions add word widths, as Tk_ComputeTextLayout measures
	// chunk by chunk, instead of re-measuring the growing line per word;
	// each emitted line is measured once for the real width.
	spaceW := m.Font.MeasureString(" ")
	emit := func(line string) {
		result = append(result, line)
		actualWidth = max(actualWidth, m.Font.MeasureString(line))
	}
	for _, para := range paragraphs {
		words := strings.Fields(para)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		start, curW := 0, m.Font.MeasureString(words[0])
		for i := 1; i < len(words); i++ {
			ww := m.Font.MeasureString(words[i])
			if maxWidth > 0 && curW+spaceW+ww > maxWidth {
				emit(strings.Join(words[start:i], " "))
				start, curW = i, ww
			} else {
				curW += spaceW + ww
			}
		}
		emit(strings.Join(words[start:], " "))
	}

	height := len(result) * metrics.Linespace()
	return result, actualWidth, height
}

// computeGeometry computes the text layout and sets the requested window size.
// Implements Tk's aspect-ratio-based iterative width convergence algorithm.
func (m *Message) computeGeometry() {
	if m.Font == nil || m.Text == "" {
		m.lines = nil
		m.msgWidth = 0
		m.msgHeight = 0
		inset := m.BorderWidth + m.HighlightWidth
		m.Win.ReqWidth = 2*m.PadX + 2*inset
		m.Win.ReqHeight = 2*m.PadY + 2*inset
		return
	}

	inset := m.BorderWidth + m.HighlightWidth

	if m.Width > 0 {
		// Explicit width specified — wrap to that width.
		m.lines, m.msgWidth, m.msgHeight = m.wrapText(m.Width)
	} else {
		// Aspect ratio algorithm from Tk's ComputeMessageGeometry.
		// Start with a wide width and narrow it until the aspect ratio
		// falls within the desired range.
		maxWidth := m.Win.Display.Server.ScreenWidth(m.Win.Display.Screen) / 2
		lines, w, h := m.wrapText(maxWidth)

		if h > 0 {
			aspect := m.Aspect
			lowerBound := aspect - aspect/10
			upperBound := aspect + aspect/10

			// Iteratively narrow the width.
			inc := maxWidth / 2
			for inc > 1 {
				currentAspect := 100 * w / h
				if currentAspect < lowerBound {
					break
				}
				tryWidth := w - inc
				if tryWidth <= 0 {
					inc /= 2
					continue
				}
				newLines, newW, newH := m.wrapText(tryWidth)
				if newH > 0 {
					newAspect := 100 * newW / newH
					if newAspect < lowerBound {
						inc /= 2
						continue
					}
					if newAspect <= upperBound {
						lines, w, h = newLines, newW, newH
						break
					}
					lines, w, h = newLines, newW, newH
				}
				inc /= 2
			}
		}

		m.lines = lines
		m.msgWidth = w
		m.msgHeight = h
	}

	m.Win.ReqWidth = m.msgWidth + 2*m.PadX + 2*inset
	m.Win.ReqHeight = m.msgHeight + 2*m.PadY + 2*inset
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (m *Message) Display() {
	m.EventuallyRedraw()
}

// display draws the message widget.
func (m *Message) display() {
	if m.Destroyed {
		return
	}
	w := m.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Fill background.
	if m.Background != nil {
		d.SetForeground(gc, m.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border.
	if m.Border != nil && m.BorderWidth > 0 && m.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, w.Drawable(), gc, m.Border,
			0, 0, w.Width, w.Height, m.BorderWidth, m.Relief)
	}

	// Draw text.
	if m.Font == nil || len(m.lines) == 0 || m.Foreground == nil {
		return
	}

	inset := m.BorderWidth + m.HighlightWidth
	// Anchor the text block within the available area.
	textX, textY := widget.ComputeAnchor(m.Anchor, w.Width, w.Height, inset, m.PadX, m.PadY, m.msgWidth, m.msgHeight)

	metrics := m.Font.Metrics()
	if df, ok := m.Font.(platform.DrawableFont); ok {
		for i, line := range m.lines {
			baseline := textY + metrics.Ascent + i*metrics.Linespace()
			lx := textX
			// Apply justify for each line.
			if len(m.lines) > 1 {
				lw := m.Font.MeasureString(line)
				switch m.Justify {
				case option.JustifyCenter:
					lx = textX + (m.msgWidth-lw)/2
				case option.JustifyRight:
					lx = textX + m.msgWidth - lw
				}
			}
			df.DrawString(w.Drawable(), lx, baseline, line,
				m.Foreground.Pixel, m.Foreground.Red, m.Foreground.Green, m.Foreground.Blue)
		}
	}

}

// Configure applies options to the message.
func (m *Message) Configure(opts ...MessageOption) {
	widget.Configure(m, opts, m.computeGeometry)
}

// Destroy cleans up the message widget.
func (m *Message) Destroy() {
	if m.Destroyed {
		return
	}
	m.Destroyed = true
	window.DestroyWindow(m.Win)
}
