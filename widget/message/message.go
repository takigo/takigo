// Package message implements the message widget, which displays multi-line
// text with automatic word-wrapping based on an aspect ratio.
// It ports tk/generic/tkMessage.c.
package message

import (
	"log"
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
	return func(m *Message) {
		col, err := m.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("message: failed to get color %q: %v", name, err)
			return
		}
		m.Background = col
		m.UpdateBorder()
	}
}

// Foreground sets the text color.
func Foreground(name string) MessageOption {
	return func(m *Message) {
		col, err := m.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("message: failed to get color %q: %v", name, err)
			return
		}
		m.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) MessageOption {
	return func(m *Message) {
		f, err := m.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("message: failed to get font %q: %v", name, err)
			return
		}
		m.Font = f
	}
}

// BorderWidth sets the border width.
func BorderWidth(w int) MessageOption {
	return func(m *Message) { m.BorderWidth = w }
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
func WidthOpt(w any) MessageOption {
	return func(m *Message) { m.Width = screenunit.Px(w) }
}

// PadX sets horizontal padding.
func PadX(p any) MessageOption {
	return func(m *Message) { m.PadX = screenunit.Px(p) }
}

// PadY sets vertical padding.
func PadY(p any) MessageOption {
	return func(m *Message) { m.PadY = screenunit.Px(p) }
}

// HighlightWidth sets the focus highlight border width.
func HighlightWidth(w int) MessageOption {
	return func(m *Message) { m.HighlightWidth = w }
}

// --- Ttk-compatible aliases (prefix with Message) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// MessageText is an alias for Text.
var MessageText = Text

// MessageBackground is an alias for Background.
var MessageBackground = Background

// MessageForeground is an alias for Foreground.
var MessageForeground = Foreground

// MessageFontOpt is an alias for FontOpt.
var MessageFontOpt = FontOpt

// MessageBorderWidth is an alias for BorderWidth.
var MessageBorderWidth = BorderWidth

// MessageRelief is an alias for Relief.
var MessageRelief = Relief

// MessageAnchor is an alias for Anchor.
var MessageAnchor = Anchor

// MessageJustifyOpt is an alias for JustifyOpt.
var MessageJustifyOpt = JustifyOpt

// MessageAspect is an alias for Aspect.
var MessageAspect = Aspect

// MessageWidthOpt is an alias for WidthOpt.
var MessageWidthOpt = WidthOpt

// MessagePadX is an alias for PadX.
var MessagePadX = PadX

// MessagePadY is an alias for PadY.
var MessagePadY = PadY

// MessageHighlightWidth is an alias for HighlightWidth.
var MessageHighlightWidth = HighlightWidth

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
		w.BackgroundPixel = m.Background.Pixel
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

	for _, para := range paragraphs {
		words := strings.Fields(para)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}
		current := words[0]
		for _, word := range words[1:] {
			candidate := current + " " + word
			if maxWidth > 0 && m.Font.MeasureString(candidate) > maxWidth {
				result = append(result, current)
				w := m.Font.MeasureString(current)
				if w > actualWidth {
					actualWidth = w
				}
				current = word
			} else {
				current = candidate
			}
		}
		result = append(result, current)
		w := m.Font.MeasureString(current)
		if w > actualWidth {
			actualWidth = w
		}
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

// Display draws the message widget.
func (m *Message) Display() {
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
		d.Flush()
		return
	}

	inset := m.BorderWidth + m.HighlightWidth
	availW := max(0, w.Width-2*inset-2*m.PadX)
	availH := max(0, w.Height-2*inset-2*m.PadY)
	frameX := inset + m.PadX
	frameY := inset + m.PadY

	// Anchor the text block within the available area.
	textX, textY := widget.AnchorText(m.Anchor, frameX, frameY, availW, availH, m.msgWidth, m.msgHeight)

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

	d.Flush()
}

// Configure applies options to the message.
func (m *Message) Configure(opts ...option.Option) {
	option.Apply(m, opts)
	m.UpdateBorder()
	m.computeGeometry()
	if m.Background != nil {
		m.Win.BackgroundPixel = m.Background.Pixel
	}
	m.Display()
}

// Destroy cleans up the message widget.
func (m *Message) Destroy() {
	if m.Destroyed {
		return
	}
	m.Destroyed = true
	window.DestroyWindow(m.Win)
}
