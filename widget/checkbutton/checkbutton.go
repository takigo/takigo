// Package checkbutton implements a toggle button with a square indicator.
// It ports the checkbutton-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package checkbutton

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/internal/tkbutton"
	"github.com/takigo/takigo/window"
)

// Checkbutton is a two-state (or three-state) toggle with an indicator.
type Checkbutton struct {
	widget.Base
	tkbutton.Shared

	Command func() // invoked on toggle

	// Variable linkage. The variable holds a string equal to OnValue, OffValue,
	// or TristateValue. The checkbutton renders a checkmark when the
	// variable equals OnValue, an empty box when OffValue, and a dash when
	// TristateValue. Mirrors Tk's `-variable -onvalue -offvalue -tristatevalue`.
	Variable      *widget.Variable[string]
	OnValue       string // value meaning "on" (default "1")
	OffValue      string // value meaning "off" (default "0")
	TristateValue string // Tk -tristatevalue (default "")
	unsub         func()
}

// CheckbuttonOption configures a Checkbutton.
type CheckbuttonOption func(*Checkbutton)

// Text sets the checkbutton text.
func Text(s string) CheckbuttonOption {
	return func(c *Checkbutton) { c.Text = s }
}

// Command sets the callback invoked when toggled.
func Command(fn func()) CheckbuttonOption {
	return func(c *Checkbutton) { c.Command = fn }
}

// Var links the checkbutton to a string variable. The variable's value
// drives the displayed state: matches OnValue for checked, OffValue for
// unchecked, and TristateValue for the indeterminate dash (when configured).
func Var(v *widget.Variable[string]) CheckbuttonOption {
	return func(c *Checkbutton) {
		if c.unsub != nil {
			c.unsub()
		}
		c.Variable = v
		c.unsub = v.OnChange(func(_, _ string) {
			c.Display()
		})
	}
}

// BoolVar links the checkbutton to a bool variable: true is checked. It is
// the typed alternative to Var, whose string variable Tk compares with
// -onvalue and -offvalue.
func BoolVar(v *widget.Variable[bool]) CheckbuttonOption {
	return func(c *Checkbutton) {
		if c.unsub != nil {
			c.unsub()
		}
		str := widget.NewVariable(c.OffValue)
		c.Variable = str
		toString := func() {
			if v.Get() {
				str.Set(c.OnValue)
			} else {
				str.Set(c.OffValue)
			}
		}
		toString()
		fromBool := v.OnChange(func(_, _ bool) { toString() })
		fromString := str.OnChange(func(_, now string) {
			v.Set(now == c.OnValue)
			c.Display()
		})
		c.unsub = func() {
			fromBool()
			fromString()
		}
	}
}

// OnValueOpt sets the value of the linked variable that renders the checkbutton
// as selected (checkmark). Default "1" — matches Tk's -onvalue default.
func OnValueOpt(v string) CheckbuttonOption {
	return func(c *Checkbutton) { c.OnValue = v }
}

// OffValueOpt sets the value of the linked variable that renders the checkbutton
// as unselected (empty box). Default "0" — matches Tk's -offvalue default.
func OffValueOpt(v string) CheckbuttonOption {
	return func(c *Checkbutton) { c.OffValue = v }
}

// TristateValueOpt sets the value of the linked variable that renders the
// checkbutton in the tri-state look. Mirrors Tk's -tristatevalue, whose
// default is "": an empty variable shows the tri-state look.
func TristateValueOpt(v string) CheckbuttonOption {
	return func(c *Checkbutton) { c.TristateValue = v }
}

// Background sets the background color.
func Background[C color.Spec](name C) CheckbuttonOption {
	return func(c *Checkbutton) { c.SetBackgroundColor(name) }
}

// Foreground sets the text color.
func Foreground[C color.Spec](name C) CheckbuttonOption {
	return func(c *Checkbutton) { c.SetForegroundColor(name) }
}

// FontOpt sets the font.
func FontOpt[F font.Spec](name F) CheckbuttonOption {
	return func(c *Checkbutton) { c.SetFont(name) }
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) CheckbuttonOption {
	return func(c *Checkbutton) { c.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadX = screenunit.ToPixels(p) }
}

// PadY sets vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadY = screenunit.ToPixels(p) }
}

// ImageOpt sets the image shown in the normal (unselected) state.
func ImageOpt(img widget.WidgetImage) CheckbuttonOption {
	return func(c *Checkbutton) { c.Img = img }
}

// SelectImageOpt sets the image shown in the selected state.
func SelectImageOpt(img widget.WidgetImage) CheckbuttonOption {
	return func(c *Checkbutton) { c.SelectImg = img }
}

// IndicatorOnOpt sets whether the indicator (checkbox square) is drawn.
func IndicatorOnOpt(on bool) CheckbuttonOption {
	return func(c *Checkbutton) { c.IndicatorOn = on }
}

// SelectColor sets -selectcolor, the indicator's colour when selected.
func SelectColor[C color.Spec](name C) CheckbuttonOption {
	return func(c *Checkbutton) {
		if col, ok := c.LookupColor(name); ok {
			c.SelectColor = col.Ref()
		}
	}
}

// State sets -state (normal, active or disabled).
func State(st widget.State) CheckbuttonOption {
	return func(c *Checkbutton) { c.State = st }
}

// New creates a new Checkbutton widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...CheckbuttonOption) *Checkbutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

	c := &Checkbutton{OnValue: "1", OffValue: "0"}
	tkbutton.Init(&c.Shared, tkbutton.TypeCheck)
	widget.InitBase(&c.Base, w, app)
	c.SetDisplayProc(c.display)
	w.OnDestroy(c.Destroy)
	w.Class = "Checkbutton"

	// Checkbutton-specific defaults.
	c.BorderWidth = widget.DefBorderWidth
	c.Relief = option.ReliefFlat
	c.PadX = 1
	c.PadY = 1
	c.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveBackground); err == nil {
		c.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveForeground); err == nil {
		c.ActiveForeground = af.Ref()
	}

	// Disabled foreground.
	if df, err := app.ColorCache().Get(widget.PaletteFor(app).DisabledForeground); err == nil {
		c.DisabledFg = df.Ref()
	}

	// Select color (indicator fill when checked).
	if sc, err := app.ColorCache().Get(widget.PaletteFor(app).SelectColor); err == nil {
		c.SelectColor = sc.Ref()
	}

	for _, opt := range opts {
		opt(c)
	}

	// Default variable if none was provided. Created after options so
	// OffValue (set via OffValueOpt) is honoured for the initial value.
	if c.Variable == nil {
		c.Variable = widget.NewVariable(c.OffValue)
	}

	c.computeGeometry()

	if c.Background != nil {
		w.SetBackgroundPixel(c.Background.Pixel)
	}

	bindCheckbutton(c, app)

	return c
}

// Selected returns whether the checkbutton is currently selected
// (variable equals OnValue).
func (c *Checkbutton) Selected() bool {
	return c.Variable.Get() == c.OnValue
}

// tristate returns whether the checkbutton is in the indeterminate state
// (variable equals TristateValue, when configured).
// isTristate follows Tk's check order: the on-value wins, then a match with
// -tristatevalue (default "", so an empty variable shows the tri-state look).
func (c *Checkbutton) isTristate() bool {
	v := c.Variable.Get()
	return v != c.OnValue && v == c.TristateValue
}

// computeGeometry ports TkpComputeButtonGeometry for TYPE_CHECK_BUTTON.
func (c *Checkbutton) computeGeometry() { tkbutton.ComputeGeometry(&c.Base, &c.Shared) }

// display ports TkpDisplayButton for TYPE_CHECK_BUTTON.
func (c *Checkbutton) display() {
	tkbutton.Display(&c.Base, &c.Shared, tkbutton.Selection{Selected: c.Selected(), Tristate: c.isTristate()})
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (c *Checkbutton) Display() {
	c.EventuallyRedraw()
}

// Toggle flips the checkbutton state by alternating the linked variable
// between OnValue and OffValue. Mirrors Tcl's click behaviour for a
// checkbutton without a tri-state value.
func (c *Checkbutton) Toggle() {
	if c.State == widget.StateDisabled {
		return
	}
	if c.Selected() {
		c.Variable.Set(c.OffValue)
	} else {
		c.Variable.Set(c.OnValue)
	}
	c.Display()
	if c.Command != nil {
		c.Command()
	}
}

// Invoke is an alias for Toggle.
func (c *Checkbutton) Invoke() {
	c.Toggle()
}

// Configure applies options.
func (c *Checkbutton) Configure(opts ...CheckbuttonOption) error {
	return widget.Configure(c, opts, c.computeGeometry)
}

// Destroy cleans up the checkbutton.
func (c *Checkbutton) Destroy() {
	if c.Destroyed {
		return
	}
	c.Destroyed = true
	if c.unsub != nil {
		c.unsub()
	}
	window.DestroyWindow(c.Win)
}
