// Package radiobutton implements a mutually-exclusive selection button with
// a circle indicator. It ports the radiobutton-specific parts of
// tk/generic/tkButton.c and library/button.tcl.
package radiobutton

import (
	"github.com/msorc/takigo/color"
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
	IndicatorOn   bool
	TristateValue string          // if non-empty and variable==TristateValue, show indeterminate dash
	SelectColor   *color.ColorRef // indicator fill when selected

	// Image (displayed instead of text when set).
	Img widget.WidgetImage

	// Active colors.
	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef

	textWidth  int
	textHeight int
	pressed    bool
	HasFocus   bool
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
		r.unsub = v.OnChange(func(_, _ string) {
			r.Display()
		})
	}
}

// Background sets the background color.
func Background(name string) RadiobuttonOption {
	return func(r *Radiobutton) {
		col, err := r.App.ColorCache().Get(name)
		if err == nil {
			r.Background = col
			r.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) RadiobuttonOption {
	return func(r *Radiobutton) {
		col, err := r.App.ColorCache().Get(name)
		if err == nil {
			r.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) RadiobuttonOption {
	return func(r *Radiobutton) {
		f, err := r.App.FontRegistry().Get(name)
		if err == nil {
			r.Font = f
		}
	}
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) RadiobuttonOption {
	return func(r *Radiobutton) { r.Anchor = a }
}

// TristateValueOpt sets the value that triggers an indeterminate (dash) display.
// When the linked variable equals this value, the radiobutton shows a horizontal
// dash instead of the dot, indicating a mixed/indeterminate state.
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
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) RadiobuttonOption {
	return func(r *Radiobutton) { r.PadX = screenunit.Px(p) }
}

// ImageOpt sets the image displayed by the radiobutton (replaces text).
func ImageOpt(img widget.WidgetImage) RadiobuttonOption {
	return func(r *Radiobutton) { r.Img = img }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) RadiobuttonOption {
	return func(r *Radiobutton) { r.PadY = screenunit.Px(p) }
}

// indicatorSize is the diameter of the circle indicator.
const indicatorSize = 13

// New creates a new Radiobutton widget.
func New(parent widget.Caregiver, name string, opts ...RadiobuttonOption) *Radiobutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

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
		r.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		r.ActiveForeground = af.Ref()
	}

	// Select color.
	if sc, err := app.ColorCache().Get("#b03060"); err == nil {
		r.SelectColor = sc.Ref()
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

	bw := r.BorderWidth
	if !r.IndicatorOn {
		bw = 2 // button-mode always uses 2px border
	}
	inset := bw + r.HighlightWidth

	var contentW, contentH int
	if r.Img != nil {
		contentW = r.Img.Width()
		contentH = r.Img.Height()
	} else {
		contentW = r.textWidth
		contentH = r.textHeight
	}
	if r.IndicatorOn && r.Img == nil {
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
	tristate := r.TristateValue != "" && r.Variable.Get() == r.TristateValue

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *color.ColorRef
	if r.Background != nil {
		bgPixel = r.Background.Pixel
	}
	if r.Foreground != nil {
		fgCol = r.Foreground.Ref()
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

	// Draw border (inset by highlight width so highlight ring is outermost).
	hlw := r.HighlightWidth
	if !r.IndicatorOn {
		// When indicator is off, render as a toggle button: raised or sunken.
		btnRelief := option.ReliefRaised
		if selected {
			btnRelief = option.ReliefSunken
		}
		bw := 2
		border := r.Border
		if border == nil {
			border = draw.NewBorderFromPixel(bgPixel)
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, border, hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, bw, btnRelief)
	} else if r.Border != nil && r.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, r.Border,
			hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, r.BorderWidth, r.Relief)
	}

	inset := r.BorderWidth + r.HighlightWidth
	if !r.IndicatorOn {
		inset = 2 + r.HighlightWidth
	}
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

		if tristate && fgCol != nil {
			// Indeterminate: draw a horizontal dash inside the circle.
			midY := indY + indicatorSize/2
			d.SetForeground(gc, fgCol.Pixel)
			d.DrawLine(w.Drawable(), gc, indX+3, midY, indX+indicatorSize-4, midY)
			d.DrawLine(w.Drawable(), gc, indX+3, midY+1, indX+indicatorSize-4, midY+1)
		} else if selected && fgCol != nil {
			// Draw dot when selected.
			dotSize := indicatorSize - 6
			dotX := indX + 3
			dotY := indY + 3
			d.SetForeground(gc, fgCol.Pixel)
			d.FillArc(w.Drawable(), gc, dotX, dotY, uint(dotSize), uint(dotSize), 0, 360*64)
		}
	}

	// Draw image (if set) or text.
	if r.Img != nil {
		imgW := r.Img.Width()
		imgH := r.Img.Height()
		imgX := frameX + (availW-imgW)/2
		imgY := frameY + (availH-imgH)/2
		if photo, ok := r.Img.(interface {
			Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
				depth int, imgX, imgY, w, h, dstX, dstY int, bgPixel uint64)
		}); ok {
			photo.Draw(d, w.Drawable(), gc, w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)
		}
	} else if r.Font != nil && r.Text != "" && fgCol != nil {
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

	r.DrawHighlightBorder(r.HasFocus, 0)

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
