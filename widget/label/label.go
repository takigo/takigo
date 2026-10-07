// Package label implements the label widget, which displays a text string
// or image. It ports the label-specific parts of tk/generic/tkButton.c.
package label

import (
	gocolor "image/color"

	"github.com/takigo/takigo/bitmap"
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/internal/tkbutton"
	"github.com/takigo/takigo/window"
)

// Label displays text, an image, or both.
type Label struct {
	widget.Base
	tkbutton.Shared

	// Disabled draws the text in the disabled foreground.
	Disabled bool

	// TextVariable linkage: when set, the variable's value overrides Text.
	TextVar *widget.Variable[string]
	unsub   func()
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

	l := &Label{}
	tkbutton.Init(&l.Shared, tkbutton.TypeLabel)
	widget.InitBase(&l.Base, w, app)
	if df, err := app.ColorCache().Get(widget.PaletteFor(app).DisabledForeground); err == nil {
		l.DisabledFg = df.Ref()
	}
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

// computeGeometry ports TkpComputeButtonGeometry for TYPE_LABEL.
func (l *Label) computeGeometry() { tkbutton.ComputeGeometry(&l.Base, &l.Shared) }

// display ports TkpDisplayButton for TYPE_LABEL.
func (l *Label) display() {
	l.State = widget.StateNormal
	if l.Disabled {
		l.State = widget.StateDisabled
	}
	tkbutton.Display(&l.Base, &l.Shared, tkbutton.Selection{})
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (l *Label) Display() {
	l.EventuallyRedraw()
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
