// Package label implements the label widget, which displays a text string
// or image. It ports the label-specific parts of tk/generic/tkButton.c.
package label

import (
	gocolor "image/color"
	"strings"

	"github.com/takigo/takigo/bitmap"
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
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

	// Disabled state — when true, text is drawn in gray.
	Disabled bool

	// TextVariable linkage — when set, the variable's value overrides Text.
	TextVar *widget.Variable[string]
	unsub   func()

	// WidthChars/HeightChars are Tk's -width/-height: characters and lines
	// for text, pixels when an image is shown. 0 = natural size.
	WidthChars  int
	HeightChars int

	textWidth  int
	textHeight int

	// lines caches textLines for linesKey.
	lines    []string
	linesKey labelLayoutKey
}

// LabelOption configures a Label.
type LabelOption func(*Label)

// Text sets the label text.
func Text(s string) LabelOption {
	return func(l *Label) { l.Text = s }
}

// TextVariable links the label's text to a string variable.
// When the variable changes, the label text updates automatically.
func TextVariable(v *widget.Variable[string]) LabelOption {
	return func(l *Label) {
		if l.unsub != nil {
			l.unsub()
		}
		l.TextVar = v
		l.Text = v.Get()
		l.unsub = v.OnChange(func(_, newVal string) {
			l.Text = newVal
			l.computeGeometry()
			l.Display()
		})
	}
}

// Background sets the background color.
func Background[C color.Spec](name C) LabelOption {
	return func(l *Label) { l.SetBackgroundColor(name) }
}

// Foreground sets the text color.
func Foreground[C color.Spec](name C) LabelOption {
	return func(l *Label) { l.SetForegroundColor(name) }
}

// FontOpt sets the font.
func FontOpt[F font.Spec](name F) LabelOption {
	return func(l *Label) { l.SetFont(name) }
}

// BorderWidth sets the border width.
func BorderWidth[L screenunit.Length](w L) LabelOption {
	return func(l *Label) { l.BorderWidth = screenunit.ToPixels(w) }
}

// Relief sets the border relief.
func Relief(r option.Relief) LabelOption {
	return func(l *Label) { l.Relief = r }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) LabelOption {
	return func(l *Label) { l.Anchor = a }
}

// JustifyOpt sets the text justification for multi-line labels.
func JustifyOpt(j option.Justify) LabelOption {
	return func(l *Label) { l.Justify = j }
}

// PadX sets horizontal padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) LabelOption {
	return func(l *Label) { l.PadX = screenunit.ToPixels(p) }
}

// PadY sets vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) LabelOption {
	return func(l *Label) { l.PadY = screenunit.ToPixels(p) }
}

// ImageOpt sets the image to display.
func ImageOpt(img widget.WidgetImage) LabelOption {
	return func(l *Label) { l.Img = img }
}

// CompoundOpt sets how text and image are combined.
func CompoundOpt(c widget.Compound) LabelOption {
	return func(l *Label) { l.Compound = c }
}

// Bitmap sets a built-in bitmap by name (e.g. "questhead", "error", "info").
// The bitmap is rendered in the label's foreground color on a transparent background.
func Bitmap(name string) LabelOption {
	return func(l *Label) {
		fg := gocolor.RGBA{0, 0, 0, 255}
		if l.Foreground != nil {
			fg = gocolor.RGBA{
				uint8(l.Foreground.Red >> 8),
				uint8(l.Foreground.Green >> 8),
				uint8(l.Foreground.Blue >> 8),
				255,
			}
		}
		bg := gocolor.RGBA{0, 0, 0, 0} // transparent
		photo := bitmap.Get(name, fg, bg)
		if photo != nil {
			l.Img = photo
		}
	}
}

// Width sets -width: characters for text, pixels when an image is shown.
func Width(w int) LabelOption {
	return func(l *Label) { l.WidthChars = w }
}

// Height sets -height: lines for text, pixels when an image is shown.
func Height(h int) LabelOption {
	return func(l *Label) { l.HeightChars = h }
}

// WrapLength sets the maximum line width for text wrapping.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
// Set to 0 (default) to disable wrapping.
func WrapLength[L screenunit.Length](w L) LabelOption {
	return func(l *Label) { l.WrapLen = screenunit.ToPixels(w) }
}

// New creates a new Label widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...LabelOption) *Label {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	l := &Label{
		Anchor:    option.AnchorCenter,
		Justify:   option.JustifyCenter, // DEF_BUTTON_JUSTIFY
		Underline: -1,
	}
	widget.InitBase(&l.Base, w, app)
	l.SetDisplayProc(l.display)
	w.OnDestroy(l.Destroy)
	w.Class = "Label"

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
		w.SetBackgroundPixel(l.Background.Pixel)
	}

	// Bind events.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		l.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			l.Display()
		}
	})

	return l
}

// textLines returns the text split into display lines, applying WrapLen if
// set. The result is kept until the text, font or wrap length changes, as
// Tk keeps the label's textLayout between ComputeLabelGeometry and display.
func (l *Label) textLines() []string {
	if l.Text == "" {
		return nil
	}
	k := labelLayoutKey{l.Text, l.Font, l.WrapLen}
	if l.linesKey == k && l.lines != nil {
		return l.lines
	}
	if l.WrapLen <= 0 || l.Font == nil {
		l.lines = strings.Split(l.Text, "\n")
	} else {
		l.lines = font.WrapLines(l.Font, l.Text, l.WrapLen)
	}
	l.linesKey = k
	return l.lines
}

// labelLayoutKey is what a label's line breaks depend on.
type labelLayoutKey struct {
	text string
	font font.Font
	wrap int
}

// computeGeometry computes the text/image size and sets the requested window size.
func (l *Label) computeGeometry() {
	// Measure text, handling multiline and wraplength.
	if l.Font != nil && l.Text != "" {
		m := l.Font.Metrics()
		lines := l.textLines()
		l.textHeight = len(lines) * m.Linespace()
		// Measure actual width of wrapped/unwrapped text lines.
		// Tk's Tk_ComputeTextLayout returns the actual text width,
		// not the wrapLength, so the label only requests what it needs.
		l.textWidth = 0
		for _, line := range lines {
			w := font.TextWidth(l.Font, line)
			if w > l.textWidth {
				l.textWidth = w
			}
		}
	} else {
		// Tk_ComputeTextLayout lays out "" as one empty line.
		l.textWidth = 0
		l.textHeight = 0
		if l.Font != nil {
			l.textHeight = l.Font.Metrics().Linespace()
		}
	}

	// The rest ports TkpComputeButtonGeometry (tk/unix/tkUnixButton.c) for
	// TYPE_LABEL: no padding around an image-only label, padX/padY as the
	// compound gap, -width/-height in chars/lines for text.
	haveText := l.textWidth != 0 && l.textHeight != 0
	var width, height int
	switch {
	case l.Img != nil && l.Compound != widget.CompoundNone && haveText:
		width, height = l.compoundSize()
		if l.WidthChars > 0 {
			width = l.WidthChars
		}
		if l.HeightChars > 0 {
			height = l.HeightChars
		}
		width += 2 * l.PadX
		height += 2 * l.PadY
	case l.Img != nil:
		width, height = l.Img.Width(), l.Img.Height()
		if l.WidthChars > 0 {
			width = l.WidthChars
		}
		if l.HeightChars > 0 {
			height = l.HeightChars
		}
	default:
		width, height = l.textWidth, l.textHeight
		if l.Font != nil {
			if l.WidthChars > 0 {
				width = l.WidthChars * l.Font.MeasureString("0")
			}
			if l.HeightChars > 0 {
				height = l.HeightChars * l.Font.Metrics().Linespace()
			}
		}
		width += 2 * l.PadX
		height += 2 * l.PadY
	}

	inset := l.BorderWidth + l.HighlightWidth
	w := l.Win
	w.ReqWidth = width + 2*inset
	w.ReqHeight = height + 2*inset
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (l *Label) Display() {
	l.EventuallyRedraw()
}

// display draws the label.
func (l *Label) display() {
	if l.Destroyed {
		return
	}
	w := l.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
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

	hasImg := l.Img != nil
	hasText := l.Font != nil && l.Text != "" && l.Foreground != nil

	// Resolve effective foreground color (gray when disabled).
	fgPixel := uint64(0)
	var fgR, fgG, fgB uint16
	if l.Foreground != nil {
		fgPixel = l.Foreground.Pixel
		fgR, fgG, fgB = l.Foreground.Red, l.Foreground.Green, l.Foreground.Blue
	}
	if l.Disabled {
		fgPixel, fgR, fgG, fgB = widget.DisabledColor(l.App)
	}

	bgPixel := uint64(0)
	if l.Background != nil {
		bgPixel = l.Background.Pixel
	}

	if hasImg && hasText && l.Compound != widget.CompoundNone {
		drawCompound(l, w, bgPixel, fgPixel, fgR, fgG, fgB)
	} else if hasImg {
		// Image only: Tk anchors it inside the inset, ignoring padx/pady.
		imgW := l.Img.Width()
		imgH := l.Img.Height()
		ix, iy := widget.ComputeAnchor(l.Anchor, w.Width, w.Height, inset, 0, 0, imgW, imgH)
		l.Img.Draw(w.Display.Server, w.Drawable(), gc,
			w.Depth, 0, 0, imgW, imgH, ix, iy, bgPixel)
	} else if hasText {
		// Text only — handle multiline (with optional wraplength).
		textX, textY := widget.ComputeAnchor(l.Anchor, w.Width, w.Height, inset, l.PadX, l.PadY, l.textWidth, l.textHeight)
		m := l.Font.Metrics()
		if df, ok := l.Font.(platform.DrawableFont); ok {
			lines := l.textLines()
			for i, line := range lines {
				baseline := textY + m.Ascent + i*m.Linespace()
				lx := textX
				// Apply justify for multiline.
				if len(lines) > 1 {
					lw := font.TextWidth(l.Font, line)
					switch l.Justify {
					case option.JustifyCenter:
						lx = textX + (l.textWidth-lw)/2
					case option.JustifyRight:
						lx = textX + l.textWidth - lw
					default:
					}
				}
				for _, seg := range font.Segments(l.Font, line) {
					df.DrawString(w.Drawable(), lx+seg.X, baseline, seg.Text,
						fgPixel, fgR, fgG, fgB)
				}
			}
		}
	}

}

// compoundSize computes the image+text block size; the gap is padX/padY,
// as in TkpComputeButtonGeometry.
func (l *Label) compoundSize() (int, int) {
	imgW, imgH := l.Img.Width(), l.Img.Height()
	switch l.Compound {
	case widget.CompoundLeft, widget.CompoundRight:
		return imgW + l.PadX + l.textWidth, max(imgH, l.textHeight)
	case widget.CompoundTop, widget.CompoundBottom:
		return max(imgW, l.textWidth), imgH + l.PadY + l.textHeight
	case widget.CompoundCenter:
		return max(imgW, l.textWidth), max(imgH, l.textHeight)
	default:
		return imgW, imgH
	}
}

// drawCompound draws image and text in compound mode.
func drawCompound(l *Label, w *window.Window,
	bgPixel uint64,
	fgPixel uint64, fgR, fgG, fgB uint16) {

	imgW := l.Img.Width()
	imgH := l.Img.Height()
	contentW, contentH := l.compoundSize()

	// Anchor the content block.
	inset := l.BorderWidth + l.HighlightWidth
	cx, cy := widget.ComputeAnchor(l.Anchor, w.Width, w.Height, inset, l.PadX, l.PadY, contentW, contentH)

	var imgX, imgY, textX, textY int
	switch l.Compound {
	case widget.CompoundLeft:
		imgX = cx
		imgY = cy + (contentH-imgH)/2
		textX = cx + imgW + l.PadX
		textY = cy + (contentH-l.textHeight)/2
	case widget.CompoundRight:
		textX = cx
		textY = cy + (contentH-l.textHeight)/2
		imgX = cx + l.textWidth + l.PadX
		imgY = cy + (contentH-imgH)/2
	case widget.CompoundTop:
		imgX = cx + (contentW-imgW)/2
		imgY = cy
		textX = cx + (contentW-l.textWidth)/2
		textY = cy + imgH + l.PadY
	case widget.CompoundBottom:
		textX = cx + (contentW-l.textWidth)/2
		textY = cy
		imgX = cx + (contentW-imgW)/2
		imgY = cy + l.textHeight + l.PadY
	case widget.CompoundCenter:
		imgX = cx + (contentW-imgW)/2
		imgY = cy + (contentH-imgH)/2
		textX = cx + (contentW-l.textWidth)/2
		textY = cy + (contentH-l.textHeight)/2
	default:
	}

	// Draw image.
	l.Img.Draw(w.Display.Server, w.Drawable(), w.GC,
		w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)

	// Draw text.
	if l.Font != nil && l.Foreground != nil {
		m := l.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := l.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, l.Text,
				fgPixel, fgR, fgG, fgB)
		}
	}
}

// Configure applies options to the label.
func (l *Label) Configure(opts ...LabelOption) error {
	return widget.Configure(l, opts, l.computeGeometry)
}

// SetImage sets or clears the label image at runtime, recomputing geometry
// and requesting re-layout from the geometry manager.
func (l *Label) SetImage(img widget.WidgetImage) {
	l.Configure(ImageOpt(img))
}

// Destroy cleans up the label, unsubscribing from any linked TextVariable.
func (l *Label) Destroy() {
	if l.Destroyed {
		return
	}
	l.Destroyed = true
	if l.unsub != nil {
		l.unsub()
		l.unsub = nil
	}
	window.DestroyWindow(l.Win)
}
