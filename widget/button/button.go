// Package button implements the button widget with press/hover interaction.
// It ports the button-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package button

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/internal/tkbutton"
	"github.com/takigo/takigo/window"
)

// Button is an interactive widget that invokes a command when clicked.
type Button struct {
	widget.Base
	tkbutton.Shared

	Command func() // invoked on click
}

// DefaultState is the value of Tk's -default option.
type DefaultState = tkbutton.DefaultState

const (
	DefaultDisabled = tkbutton.DefaultDisabled
	DefaultNormal   = tkbutton.DefaultNormal
	DefaultActive   = tkbutton.DefaultActive
)

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
func Background[C color.Spec](name C) ButtonOption {
	return func(b *Button) { b.SetBackgroundColor(name) }
}

// Foreground sets the text color.
func Foreground[C color.Spec](name C) ButtonOption {
	return func(b *Button) { b.SetForegroundColor(name) }
}

// FontOpt sets the font.
func FontOpt[F font.Spec](name F) ButtonOption {
	return func(b *Button) { b.SetFont(name) }
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
func BorderWidth[L screenunit.Length](w L) ButtonOption {
	return func(b *Button) { b.BorderWidth = screenunit.ToPixels(w) }
}

// Default sets -default: DefaultNormal leaves 5px for a default ring,
// DefaultActive also draws it (TkpDisplayButton).
func Default(state DefaultState) ButtonOption {
	return func(b *Button) { b.Default = state }
}

// HighlightThickness sets -highlightthickness (width of the focus ring).
func HighlightThickness[L screenunit.Length](w L) ButtonOption {
	return func(b *Button) { b.HighlightWidth = screenunit.ToPixels(w) }
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
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) ButtonOption {
	return func(b *Button) { b.PadX = screenunit.ToPixels(p) }
}

// PadY sets vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) ButtonOption {
	return func(b *Button) { b.PadY = screenunit.ToPixels(p) }
}

// Width sets the requested button width in characters (like Tk's -width option).
// When > 0, the button is at least this many "0" characters wide.
func Width(n int) ButtonOption {
	return func(b *Button) { b.WidthChars = n }
}

// New creates a new Button widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...ButtonOption) *Button {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	b := &Button{}
	tkbutton.Init(&b.Shared, tkbutton.TypeButton)
	widget.InitBase(&b.Base, w, app)
	b.SetDisplayProc(b.display)
	w.Class = "Button"

	// Button-specific defaults (Tk: padx=3m, pady=1m, borderwidth=1, highlightthickness=1).
	b.BorderWidth = widget.DefBorderWidth
	b.Relief = option.ReliefRaised
	b.PadX = screenunit.Mm(3).Pixels()
	b.PadY = screenunit.Mm(1).Pixels()
	b.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveBackground); err == nil {
		b.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveForeground); err == nil {
		b.ActiveForeground = af.Ref()
	}
	if df, err := app.ColorCache().Get(widget.PaletteFor(app).DisabledForeground); err == nil {
		b.DisabledFg = df.Ref()
	}

	// Buttons are focusable via Tab traversal.
	w.Flags |= window.FlagFocusable

	for _, opt := range opts {
		opt(b)
	}

	// Compute geometry.
	b.computeGeometry()

	// Update X window background.
	if b.Background != nil {
		w.SetBackgroundPixel(b.Background.Pixel)
	}

	// Bind events.
	bindButton(b, app)

	// FocusIn/FocusOut — track focus state and redraw highlight.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			b.HasFocus = true
		} else {
			b.HasFocus = false
		}
		b.Display()
	})

	return b
}

// computeGeometry ports TkpComputeButtonGeometry for TYPE_BUTTON.
func (b *Button) computeGeometry() { tkbutton.ComputeGeometry(&b.Base, &b.Shared) }

// display ports TkpDisplayButton for TYPE_BUTTON.
func (b *Button) display() { tkbutton.Display(&b.Base, &b.Shared, tkbutton.Selection{}) }

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (b *Button) Display() {
	b.EventuallyRedraw()
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

// SetText updates the button's label text and refreshes its size and display.
func (b *Button) SetText(s string) {
	b.Configure(Text(s))
}

// Configure applies options to the button.
func (b *Button) Configure(opts ...ButtonOption) error {
	return widget.Configure(b, opts, b.computeGeometry)
}

// Destroy cleans up the button.
func (b *Button) Destroy() {
	if b.Destroyed {
		return
	}
	b.Destroyed = true
	window.DestroyWindow(b.Win)
}
