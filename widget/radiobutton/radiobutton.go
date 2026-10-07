// Package radiobutton implements a mutually-exclusive selection button with
// a circle indicator. It ports the radiobutton-specific parts of
// tk/generic/tkButton.c and library/button.tcl.
package radiobutton

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/internal/tkbutton"
	"github.com/takigo/takigo/window"
)

// Radiobutton is one choice of a group sharing a variable.
type Radiobutton struct {
	widget.Base
	tkbutton.Shared

	Command func() // invoked on selection
	Value   string // the value this radio represents

	// Variable linkage (shared across radio group).
	Variable *widget.Variable[string]
	unsub    func()

	untraced      bool   // created its variable; see Var
	TristateValue string // Tk -tristatevalue (default ""): variable==TristateValue shows the tri-state look
}

// RadiobuttonOption configures a Radiobutton.
type RadiobuttonOption func(*Radiobutton)

// Text sets the radiobutton text.
func Text(s string) RadiobuttonOption {
	return func(r *Radiobutton) { r.Text = s }
}

// Value sets the value this radio represents.
func Value(v string) RadiobuttonOption {
	return func(r *Radiobutton) { r.Value = v }
}

// Command sets the callback invoked when selected.
func Command(fn func()) RadiobuttonOption {
	return func(r *Radiobutton) { r.Command = fn }
}

// Var links the radiobutton to a string variable (shared across group).
func Var(v *widget.Variable[string]) RadiobuttonOption {
	return func(r *Radiobutton) {
		if r.unsub != nil {
			r.unsub()
		}
		r.Variable = v
		// ConfigureButton creates a missing variable as "" before tracing
		// it, so this button is not tri-stated until the variable changes.
		r.untraced = !v.IsSet()
		if r.untraced {
			v.Set("")
		}
		r.unsub = v.OnChange(func(_, _ string) {
			r.untraced = false
			r.Display()
		})
	}
}

// Background sets the background color.
func Background[C color.Spec](name C) RadiobuttonOption {
	return func(r *Radiobutton) { r.SetBackgroundColor(name) }
}

// Foreground sets the text color.
func Foreground[C color.Spec](name C) RadiobuttonOption {
	return func(r *Radiobutton) { r.SetForegroundColor(name) }
}

// FontOpt sets the font.
func FontOpt[F font.Spec](name F) RadiobuttonOption {
	return func(r *Radiobutton) { r.SetFont(name) }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) RadiobuttonOption {
	return func(r *Radiobutton) { r.Anchor = a }
}

// TristateValueOpt sets Tk's -tristatevalue: when the linked variable equals
// it (and not -value), the radiobutton shows the tri-state look. The default
// is "", so an empty variable shows the tri-state look.
func TristateValueOpt(v string) RadiobuttonOption {
	return func(r *Radiobutton) { r.TristateValue = v }
}

// IndicatorOnOpt sets whether the circle indicator is shown.
// When false the button renders like a toggle button: raised when
// unselected, sunken when selected (matching Tk's -indicatoron 0).
func IndicatorOnOpt(on bool) RadiobuttonOption {
	return func(r *Radiobutton) { r.IndicatorOn = on }
}

// PadX sets horizontal padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) RadiobuttonOption {
	return func(r *Radiobutton) { r.PadX = screenunit.ToPixels(p) }
}

// ImageOpt sets the image displayed by the radiobutton (replaces text).
func ImageOpt(img widget.WidgetImage) RadiobuttonOption {
	return func(r *Radiobutton) { r.Img = img }
}

// PadY sets vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) RadiobuttonOption {
	return func(r *Radiobutton) { r.PadY = screenunit.ToPixels(p) }
}

// Width sets the requested width in characters of the default font.
// Matches Tk's `-width` option for text-only buttons.
func Width(n int) RadiobuttonOption {
	return func(r *Radiobutton) { r.WidthChars = n }
}

// New creates a new Radiobutton widget.
func New(parent widget.Caregiver, name string, opts ...RadiobuttonOption) *Radiobutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

	r := &Radiobutton{}
	tkbutton.Init(&r.Shared, tkbutton.TypeRadio)
	widget.InitBase(&r.Base, w, app)
	r.SetDisplayProc(r.display)
	w.OnDestroy(r.Destroy)
	w.Class = "Radiobutton"

	// Radiobutton-specific defaults.
	r.BorderWidth = widget.DefBorderWidth
	r.Relief = option.ReliefFlat
	r.PadX = 1
	r.PadY = 1
	r.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveBackground); err == nil {
		r.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveForeground); err == nil {
		r.ActiveForeground = af.Ref()
	}

	// Disabled foreground.
	if df, err := app.ColorCache().Get(widget.PaletteFor(app).DisabledForeground); err == nil {
		r.DisabledFg = df.Ref()
	}

	// Select color.
	if sc, err := app.ColorCache().Get(widget.PaletteFor(app).SelectColor); err == nil {
		r.SelectColor = sc.Ref()
	}

	// Default variable.
	r.Variable = widget.NewVariable("")

	for _, opt := range opts {
		opt(r)
	}

	r.computeGeometry()

	if r.Background != nil {
		w.SetBackgroundPixel(r.Background.Pixel)
	}

	bindRadiobutton(r, app)

	return r
}

// Selected returns whether this radiobutton is currently selected.
func (r *Radiobutton) Selected() bool {
	return r.Variable.Get() == r.Value
}

// computeGeometry ports TkpComputeButtonGeometry for TYPE_RADIO_BUTTON.
func (r *Radiobutton) computeGeometry() { tkbutton.ComputeGeometry(&r.Base, &r.Shared) }

// display ports TkpDisplayButton for TYPE_RADIO_BUTTON. The -value match
// wins, then -tristatevalue (default "").
func (r *Radiobutton) display() {
	selected := r.Selected()
	tristate := !selected && !r.untraced && r.Variable.Get() == r.TristateValue
	tkbutton.Display(&r.Base, &r.Shared, tkbutton.Selection{Selected: selected, Tristate: tristate})
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (r *Radiobutton) Display() {
	r.EventuallyRedraw()
}

// Select selects this radiobutton (sets the variable to this button's value).
func (r *Radiobutton) Select() {
	if r.State == widget.StateDisabled {
		return
	}
	r.Variable.Set(r.Value)
	r.Display()
	if r.Command != nil {
		r.Command()
	}
}

// Invoke is an alias for Select.
func (r *Radiobutton) Invoke() {
	r.Select()
}

// Configure applies options.
func (r *Radiobutton) Configure(opts ...RadiobuttonOption) error {
	return widget.Configure(r, opts, r.computeGeometry)
}

// Destroy cleans up the radiobutton.
func (r *Radiobutton) Destroy() {
	if r.Destroyed {
		return
	}
	r.Destroyed = true
	if r.unsub != nil {
		r.unsub()
	}
	window.DestroyWindow(r.Win)
}
