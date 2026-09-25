// Package radiobutton implements a mutually-exclusive selection button with
// a circle indicator. It ports the radiobutton-specific parts of
// tk/generic/tkButton.c and library/button.tcl.
package radiobutton

import (
	"log"

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
	TristateValue string          // Tk -tristatevalue (default ""): variable==TristateValue shows the tri-state look
	SelectColor   *color.ColorRef // indicator fill when selected

	// Image (displayed instead of text when set).
	Img widget.WidgetImage

	// Active colors.
	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef

	// Disabled foreground (used when State == StateDisabled).
	DisabledFg *color.ColorRef

	// WidthChars sets the requested width in characters of the default font (Tk's -width).
	WidthChars int

	textWidth      int
	textHeight     int
	indicatorSpace int // Tk butPtr->indicatorSpace
	pressed        bool
	HasFocus       bool
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
		if err != nil {
			log.Printf("radiobutton: failed to get color %q: %v", name, err)
			return
		}
		r.Background = col
		r.UpdateBorder()
	}
}

// Foreground sets the text color.
func Foreground(name string) RadiobuttonOption {
	return func(r *Radiobutton) {
		col, err := r.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("radiobutton: failed to get color %q: %v", name, err)
			return
		}
		r.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) RadiobuttonOption {
	return func(r *Radiobutton) {
		f, err := r.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("radiobutton: failed to get font %q: %v", name, err)
			return
		}
		r.Font = f
	}
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

// Width sets the requested width in characters of the default font.
// Matches Tk's `-width` option for text-only buttons.
func Width(n int) RadiobuttonOption {
	return func(r *Radiobutton) { r.WidthChars = n }
}

// --- Ttk-compatible aliases (prefix with Radiobutton) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.RadiobuttonText, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// RadiobuttonText is an alias for Text.
var RadiobuttonText = Text

// RadiobuttonValue is an alias for Value.
var RadiobuttonValue = Value

// RadiobuttonCommand is an alias for Command.
var RadiobuttonCommand = Command

// RadiobuttonVar is an alias for Var.
var RadiobuttonVar = Var

// RadiobuttonBackground is an alias for Background.
var RadiobuttonBackground = Background

// RadiobuttonForeground is an alias for Foreground.
var RadiobuttonForeground = Foreground

// RadiobuttonFontOpt is an alias for FontOpt.
var RadiobuttonFontOpt = FontOpt

// RadiobuttonAnchor is an alias for Anchor.
var RadiobuttonAnchor = Anchor

// RadiobuttonTristateValueOpt is an alias for TristateValueOpt.
var RadiobuttonTristateValueOpt = TristateValueOpt

// RadiobuttonIndicatorOnOpt is an alias for IndicatorOnOpt.
var RadiobuttonIndicatorOnOpt = IndicatorOnOpt

// RadiobuttonPadX is an alias for PadX.
var RadiobuttonPadX = PadX

// RadiobuttonPadY is an alias for PadY.
var RadiobuttonPadY = PadY

// RadiobuttonImageOpt is an alias for ImageOpt.
var RadiobuttonImageOpt = ImageOpt

// RadiobuttonWidth is an alias for Width.
var RadiobuttonWidth = Width

// New creates a new Radiobutton widget.
func New(parent widget.Caregiver, name string, opts ...RadiobuttonOption) *Radiobutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

	r := &Radiobutton{
		Anchor:      option.AnchorCenter,
		IndicatorOn: true,
	}
	widget.InitBase(&r.Base, w, app)
	w.Class = "Radiobutton"

	// Radiobutton-specific defaults.
	r.BorderWidth = widget.DefBorderWidth
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

	// Disabled foreground.
	if df, err := app.ColorCache().Get(widget.DefDisabledForeground); err == nil {
		r.DisabledFg = df.Ref()
	}

	// Select color.
	if sc, err := app.ColorCache().Get(widget.DefSelectColor); err == nil {
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

// computeGeometry ports TkpComputeButtonGeometry (tk/unix/tkUnixButton.c)
// for check/radio buttons: the indicator gets its own column
// (indicatorSpace) left of the content.
func (r *Radiobutton) computeGeometry() {
	r.textWidth, r.textHeight = 0, 0
	avg := 0
	if r.Font != nil {
		r.textWidth = r.Font.MeasureString(r.Text)
		r.textHeight = r.Font.Metrics().Linespace()
		avg = r.Font.MeasureString("0")
	}

	inset := r.BorderWidth + r.HighlightWidth
	img := r.Img
	var width, height int
	r.indicatorSpace = 0
	if img != nil {
		width, height = img.Width(), img.Height()
		if r.IndicatorOn {
			r.indicatorSpace = height
		}
	} else {
		width, height = r.textWidth, r.textHeight
		if r.WidthChars > 0 {
			width = r.WidthChars * avg
		}
		if r.IndicatorOn {
			r.indicatorSpace = r.textHeight + avg
		}
		width += 2 * r.PadX
		height += 2 * r.PadY
	}

	w := r.Win
	w.ReqWidth = width + r.indicatorSpace + 2*inset
	w.ReqHeight = height + 2*inset
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
	// Tk: the -value match wins, then -tristatevalue (default "").
	tristate := !selected && r.Variable.Get() == r.TristateValue

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
	if r.State == widget.StateDisabled && r.DisabledFg != nil {
		fgCol = r.DisabledFg
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border (inset by highlight width so highlight ring is outermost).
	hlw := r.HighlightWidth
	if !r.IndicatorOn {
		// TkpDisplayButton: selected -> sunken on -selectcolor, otherwise
		// -offrelief (raised by default).
		btnRelief := option.ReliefRaised
		border := r.Border
		if border == nil {
			border = draw.NewBorderFromPixel(bgPixel)
		}
		if selected {
			btnRelief = option.ReliefSunken
			if r.SelectColor != nil {
				border = draw.NewBorderFromPixel(r.SelectColor.Pixel)
				d.SetForeground(gc, border.BgPixel)
				d.FillRectangle(w.Drawable(), gc, hlw, hlw, uint(w.Width-2*hlw), uint(w.Height-2*hlw))
			}
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, border, hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, r.BorderWidth, btnRelief)
	} else if r.Border != nil && r.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, r.Border,
			hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, r.BorderWidth, r.Relief)
	}

	inset := r.BorderWidth + r.HighlightWidth
	img := r.Img
	var contentX, contentY int
	if img != nil {
		contentX, contentY = widget.ComputeAnchor(r.Anchor, w.Width, w.Height, inset, 0, 0, r.indicatorSpace+img.Width(), img.Height())
	} else {
		contentX, contentY = widget.ComputeAnchor(r.Anchor, w.Width, w.Height, inset, r.PadX, r.PadY, r.indicatorSpace+r.textWidth, r.textHeight)
	}
	contentX += r.indicatorSpace

	if r.IndicatorOn {
		// Tk centres the indicator in its column and on the window's mid-line.
		state := draw.IndicatorOff
		switch {
		case tristate:
			state = draw.IndicatorTristate
		case selected:
			state = draw.IndicatorOn
		}
		selPixel := uint64(0xffffff)
		if r.SelectColor != nil {
			selPixel = r.SelectColor.Pixel
		}
		var fgPixel, disPixel uint64 = 0, 0xa3a3a3
		if r.Foreground != nil {
			fgPixel = r.Foreground.Pixel
		}
		if r.DisabledFg != nil {
			disPixel = r.DisabledFg.Pixel
		}
		draw.DrawCheckIndicator(d, w.Drawable(), gc, w.Depth,
			contentX-r.indicatorSpace/2, w.Height/2, draw.RadioIndicator,
			draw.NewBorderFromPixel(bgPixel), fgPixel, selPixel, disPixel,
			state, r.State == widget.StateDisabled)
	}

	// Draw image (if set) or text.
	if img != nil {
		imgW := img.Width()
		imgH := img.Height()
		imgX, imgY := contentX, contentY
		if photo, ok := img.(interface {
			Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
				depth int, imgX, imgY, w, h, dstX, dstY int, bgPixel uint64)
		}); ok {
			photo.Draw(d, w.Drawable(), gc, w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)
		}
	} else if r.Font != nil && r.Text != "" && fgCol != nil {
		textX, textY := contentX, contentY
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
