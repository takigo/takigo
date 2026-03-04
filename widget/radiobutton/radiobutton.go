// Package radiobutton implements a mutually-exclusive selection button with
// a circle indicator. It ports the radiobutton-specific parts of
// tk/generic/tkButton.c and library/button.tcl.
package radiobutton

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Radiobutton is a mutually-exclusive selection widget with a circle indicator.
type Radiobutton struct {
	widget.Base

	Text    string
	Command func() // invoked on selection
	Anchor  option.Anchor
	State   widget.State
	Value   string // the value this radio represents

	// Variable linkage (shared across radio group).
	Variable *widget.Variable[string]
	unsub    func()

	// Indicator.
	IndicatorOn bool
	SelectColor *colorRef // indicator fill when selected

	// Active colors.
	ActiveBackground *colorRef
	ActiveForeground *colorRef

	textWidth  int
	textHeight int
	pressed    bool
}

type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// Option configures a Radiobutton.
type Option func(*Radiobutton)

// Text sets the radiobutton text.
func Text(s string) Option {
	return func(r *Radiobutton) { r.Text = s }
}

// Value sets the value this radio represents.
func Value(v string) Option {
	return func(r *Radiobutton) { r.Value = v }
}

// Command sets the callback invoked when selected.
func Command(fn func()) Option {
	return func(r *Radiobutton) { r.Command = fn }
}

// Var links the radiobutton to a string variable (shared across group).
func Var(v *widget.Variable[string]) Option {
	return func(r *Radiobutton) {
		if r.unsub != nil {
			r.unsub()
		}
		r.Variable = v
		r.unsub = v.OnChange(func(_, _ string) {
			r.Display()
		})
	}
}

// Background sets the background color.
func Background(name string) Option {
	return func(r *Radiobutton) {
		col, err := r.App.ColorCache().Get(name)
		if err == nil {
			r.Background = col
			r.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) Option {
	return func(r *Radiobutton) {
		col, err := r.App.ColorCache().Get(name)
		if err == nil {
			r.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) Option {
	return func(r *Radiobutton) {
		f, err := r.App.FontRegistry().Get(name)
		if err == nil {
			r.Font = f
		}
	}
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) Option {
	return func(r *Radiobutton) { r.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) Option {
	return func(r *Radiobutton) { r.PadX = screenunit.Px(p) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) Option {
	return func(r *Radiobutton) { r.PadY = screenunit.Px(p) }
}

// indicatorSize is the diameter of the circle indicator.
const indicatorSize = 13

// New creates a new Radiobutton widget.
func New(parent widget.Caregiver, name string, opts ...Option) *Radiobutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	r := &Radiobutton{
		Anchor:      option.AnchorW,
		IndicatorOn: true,
	}
	widget.InitBase(&r.Base, w, app)

	// Radiobutton-specific defaults.
	r.BorderWidth = 0
	r.Relief = option.ReliefFlat
	r.PadX = 1
	r.PadY = 1
	r.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		r.ActiveBackground = &colorRef{ac.Pixel, ac.Red, ac.Green, ac.Blue}
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		r.ActiveForeground = &colorRef{af.Pixel, af.Red, af.Green, af.Blue}
	}

	// Select color.
	if sc, err := app.ColorCache().Get("#b03060"); err == nil {
		r.SelectColor = &colorRef{sc.Pixel, sc.Red, sc.Green, sc.Blue}
	}

	// Default variable.
	r.Variable = widget.NewVariable("")

	for _, opt := range opts {
		opt(r)
	}

	r.computeGeometry()

	if r.Background != nil {
		w.BackgroundPixel = r.Background.Pixel
	}

	bindRadiobutton(r, app)

	return r
}

func (r *Radiobutton) computeGeometry() {
	if r.Font != nil && r.Text != "" {
		r.textWidth = r.Font.MeasureString(r.Text)
		m := r.Font.Metrics()
		r.textHeight = m.Linespace()
	} else {
		r.textWidth = 0
		r.textHeight = 0
	}

	inset := r.BorderWidth + r.HighlightWidth
	contentW := r.textWidth
	contentH := r.textHeight
	if r.IndicatorOn {
		contentW += indicatorSize + 4
		if indicatorSize > contentH {
			contentH = indicatorSize
		}
	}

	w := r.Win
	w.ReqWidth = contentW + 2*r.PadX + 2*inset
	w.ReqHeight = contentH + 2*r.PadY + 2*inset
}

// Selected returns whether this radiobutton is currently selected.
func (r *Radiobutton) Selected() bool {
	return r.Variable.Get() == r.Value
}

// Display draws the radiobutton.
func (r *Radiobutton) Display() {
	if r.Destroyed {
		return
	}
	w := r.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	selected := r.Selected()

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *colorRef
	if r.Background != nil {
		bgPixel = r.Background.Pixel
	}
	if r.Foreground != nil {
		fgCol = &colorRef{r.Foreground.Pixel, r.Foreground.Red, r.Foreground.Green, r.Foreground.Blue}
	}

	if r.State == widget.StateActive && r.ActiveBackground != nil {
		bgPixel = r.ActiveBackground.Pixel
	}
	if r.State == widget.StateActive && r.ActiveForeground != nil {
		fgCol = r.ActiveForeground
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border.
	if r.Border != nil && r.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, r.Border,
			0, 0, w.Width, w.Height, r.BorderWidth, r.Relief)
	}

	inset := r.BorderWidth + r.HighlightWidth
	availW := w.Width - 2*inset - 2*r.PadX
	availH := w.Height - 2*inset - 2*r.PadY
	frameX := inset + r.PadX
	frameY := inset + r.PadY

	indW := 0
	if r.IndicatorOn {
		indW = indicatorSize + 4

		// Draw circle indicator.
		indX := frameX
		indY := frameY + (availH-indicatorSize)/2

		// Fill circle: white background or select color.
		if selected && r.SelectColor != nil {
			d.SetForeground(gc, r.SelectColor.Pixel)
		} else {
			d.SetForeground(gc, uint64(0xffffff))
		}
		// FillArc uses 64ths of a degree; full circle = 0 to 360*64.
		d.FillArc(w.Drawable(), gc, indX, indY, uint(indicatorSize), uint(indicatorSize), 0, 360*64)

		// Draw circle border (dark outer ring).
		indBorder := r.Border
		if indBorder == nil {
			indBorder = draw.NewBorderFromPixel(bgPixel)
		}
		// Dark outer arc (top-left shadow).
		d.SetForeground(gc, indBorder.DarkPixel)
		d.DrawArc(w.Drawable(), gc, indX, indY, uint(indicatorSize-1), uint(indicatorSize-1), 45*64, 180*64)
		// Light inner arc (bottom-right highlight).
		d.SetForeground(gc, indBorder.LightPixel)
		d.DrawArc(w.Drawable(), gc, indX, indY, uint(indicatorSize-1), uint(indicatorSize-1), 225*64, 180*64)

		// Draw dot when selected.
		if selected && fgCol != nil {
			dotSize := indicatorSize - 6
			dotX := indX + 3
			dotY := indY + 3
			d.SetForeground(gc, fgCol.Pixel)
			d.FillArc(w.Drawable(), gc, dotX, dotY, uint(dotSize), uint(dotSize), 0, 360*64)
		}
	}

	// Draw text.
	if r.Font != nil && r.Text != "" && fgCol != nil {
		textX := frameX + indW
		textY := frameY + (availH-r.textHeight)/2
		remainW := availW - indW
		switch r.Anchor {
		case option.AnchorCenter, option.AnchorN, option.AnchorS:
			textX += (remainW - r.textWidth) / 2
		case option.AnchorE, option.AnchorNE, option.AnchorSE:
			textX += remainW - r.textWidth
		}
		m := r.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := r.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, r.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	d.Flush()
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
func (r *Radiobutton) Configure(opts ...option.Option) {
	option.Apply(r, opts)
	r.UpdateBorder()
	r.computeGeometry()
	if r.Background != nil {
		r.Win.BackgroundPixel = r.Background.Pixel
	}
	r.Display()
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
