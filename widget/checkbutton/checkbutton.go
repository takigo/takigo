// Package checkbutton implements a toggle button with a square indicator.
// It ports the checkbutton-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package checkbutton

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

// Checkbutton is a toggle widget with a square indicator and text label.
type Checkbutton struct {
	widget.Base

	Text    string
	Command func() // invoked on toggle
	Anchor  option.Anchor
	State   widget.State

	// Variable linkage. The variable holds a string equal to OnValue, OffValue,
	// or TristateValue. The checkbutton renders a checkmark when the
	// variable equals OnValue, an empty box when OffValue, and a dash when
	// TristateValue. Mirrors Tk's `-variable -onvalue -offvalue -tristatevalue`.
	Variable      *widget.Variable[string]
	OnValue       string // value meaning "on" (default "1")
	OffValue      string // value meaning "off" (default "0")
	TristateValue string // Tk -tristatevalue (default "")
	unsub         func()

	// Indicator.
	IndicatorOn bool            // whether to draw the indicator (default true)
	SelectColor *color.ColorRef // indicator fill color when selected

	// Images (selectimage shown when checked; image shown otherwise).
	Img       widget.WidgetImage
	SelectImg widget.WidgetImage

	// Active colors (used on hover).
	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef

	// Disabled foreground (used when State == StateDisabled).
	DisabledFg *color.ColorRef

	textWidth      int
	textHeight     int
	indicatorSpace int // Tk butPtr->indicatorSpace
	pressed        bool
	HasFocus       bool
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
func Background(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("checkbutton: failed to get color %q: %v", name, err)
			return
		}
		c.Background = col
		c.UpdateBorder()
	}
}

// Foreground sets the text color.
func Foreground(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("checkbutton: failed to get color %q: %v", name, err)
			return
		}
		c.Foreground = col
	}
}

// FontOpt sets the font.
func FontOpt(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		f, err := c.App.FontRegistry().Get(name)
		if err != nil {
			log.Printf("checkbutton: failed to get font %q: %v", name, err)
			return
		}
		c.Font = f
	}
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) CheckbuttonOption {
	return func(c *Checkbutton) { c.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadX = screenunit.PxOr(p, c.PadX) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadY = screenunit.PxOr(p, c.PadY) }
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

// --- Ttk-compatible aliases (prefix with Checkbutton) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.CheckbuttonText, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// CheckbuttonText is an alias for Text.
var CheckbuttonText = Text

// CheckbuttonCommand is an alias for Command.
var CheckbuttonCommand = Command

// CheckbuttonVar is an alias for Var.
var CheckbuttonVar = Var

// CheckbuttonOnValueOpt is an alias for OnValueOpt.
var CheckbuttonOnValueOpt = OnValueOpt

// CheckbuttonOffValueOpt is an alias for OffValueOpt.
var CheckbuttonOffValueOpt = OffValueOpt

// CheckbuttonTristateValueOpt is an alias for TristateValueOpt.
var CheckbuttonTristateValueOpt = TristateValueOpt

// CheckbuttonBackground is an alias for Background.
var CheckbuttonBackground = Background

// CheckbuttonForeground is an alias for Foreground.
var CheckbuttonForeground = Foreground

// CheckbuttonFontOpt is an alias for FontOpt.
var CheckbuttonFontOpt = FontOpt

// CheckbuttonAnchor is an alias for Anchor.
var CheckbuttonAnchor = Anchor

// CheckbuttonPadX is an alias for PadX.
var CheckbuttonPadX = PadX

// CheckbuttonPadY is an alias for PadY.
var CheckbuttonPadY = PadY

// CheckbuttonImageOpt is an alias for ImageOpt.
var CheckbuttonImageOpt = ImageOpt

// CheckbuttonSelectImageOpt is an alias for SelectImageOpt.
var CheckbuttonSelectImageOpt = SelectImageOpt

// CheckbuttonIndicatorOnOpt is an alias for IndicatorOnOpt.
var CheckbuttonIndicatorOnOpt = IndicatorOnOpt

// New creates a new Checkbutton widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...CheckbuttonOption) *Checkbutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

	c := &Checkbutton{
		Anchor:      option.AnchorCenter,
		IndicatorOn: true,
		OnValue:     "1",
		OffValue:    "0",
	}
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
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		c.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		c.ActiveForeground = af.Ref()
	}

	// Disabled foreground.
	if df, err := app.ColorCache().Get(widget.DefDisabledForeground); err == nil {
		c.DisabledFg = df.Ref()
	}

	// Select color (indicator fill when checked).
	if sc, err := app.ColorCache().Get(widget.DefSelectColor); err == nil {
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
		w.BackgroundPixel = c.Background.Pixel
	}

	bindCheckbutton(c, app)

	return c
}

// computeGeometry ports TkpComputeButtonGeometry (tk/unix/tkUnixButton.c)
// for check/radio buttons: the indicator gets its own column
// (indicatorSpace) left of the content.
func (c *Checkbutton) computeGeometry() {
	c.textWidth, c.textHeight = 0, 0
	avg := 0
	if c.Font != nil {
		c.textWidth = c.Font.MeasureString(c.Text)
		c.textHeight = c.Font.Metrics().Linespace()
		avg = c.Font.MeasureString("0")
	}

	inset := c.BorderWidth + c.HighlightWidth
	// TkpComputeButtonGeometry sizes from -image alone; -selectimage is
	// drawn in the same space.
	img := c.Img
	var width, height int
	c.indicatorSpace = 0
	if img != nil {
		width, height = img.Width(), img.Height()
		if c.IndicatorOn {
			c.indicatorSpace = height
		}
	} else {
		width, height = c.textWidth, c.textHeight
		if c.IndicatorOn {
			c.indicatorSpace = c.textHeight + avg
		}
		width += 2 * c.PadX
		height += 2 * c.PadY
	}

	w := c.Win
	w.ReqWidth = width + c.indicatorSpace + 2*inset
	w.ReqHeight = height + 2*inset
}

// activeImage returns the image to display: as in TkpDisplayButton,
// -selectimage replaces -image while selected, and only when -image is set.
func (c *Checkbutton) activeImage() widget.WidgetImage {
	if c.Img != nil && c.Selected() && c.SelectImg != nil {
		return c.SelectImg
	}
	return c.Img
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

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (c *Checkbutton) Display() {
	c.EventuallyRedraw()
}

// display draws the checkbutton.
func (c *Checkbutton) display() {
	if c.Destroyed {
		return
	}
	w := c.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	selected := c.Selected()
	tristate := c.isTristate()

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *color.ColorRef
	if c.Background != nil {
		bgPixel = c.Background.Pixel
	}
	if c.Foreground != nil {
		fgCol = c.Foreground.Ref()
	}

	if c.State == widget.StateActive && c.ActiveBackground != nil {
		bgPixel = c.ActiveBackground.Pixel
	}
	if c.State == widget.StateActive && c.ActiveForeground != nil {
		fgCol = c.ActiveForeground
	}
	if c.State == widget.StateDisabled && c.DisabledFg != nil {
		fgCol = c.DisabledFg
	}

	// In toggle mode, the SelectColor fills an inner "indicator" rectangle
	// around the image (matching Tk's -indicatoron 0 -selectcolor behaviour,
	// where the selectcolor area sits inside the widget border instead of
	// flooding the whole widget). Keep the widget background as the default
	// bg so the rest of the widget keeps its normal appearance.
	widgetBg := bgPixel
	selectPixel := uint64(0)
	hasSelectFill := false
	if !c.IndicatorOn && selected && c.SelectColor != nil {
		selectPixel = c.SelectColor.Pixel
		hasSelectFill = true
	}

	// Fill background with the widget's own background colour.
	d.SetForeground(gc, widgetBg)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border (inset by highlight width so highlight ring is outermost).
	hlw := c.HighlightWidth
	if !c.IndicatorOn {
		// Toggle button mode: raised or sunken relief based on selection.
		btnRelief := option.ReliefRaised
		if selected {
			btnRelief = option.ReliefSunken
		}
		border := c.Border
		if border == nil {
			border = draw.NewBorderFromPixel(widgetBg)
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, border, hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, c.BorderWidth, btnRelief)
	} else if c.Border != nil && c.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, c.Border,
			hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, c.BorderWidth, c.Relief)
	}

	// Fill the SelectColor indicator area inside the bezel (matches Tk's
	// -indicatoron 0 -selectcolor rendering: a small colour rectangle that
	// hugs the image).
	if hasSelectFill && c.Img != nil {
		// Inset by 1 pixel so the SelectColor area sits just inside the
		// widget edge, leaving room for a visible bezel frame around it.
		innerInset := 1
		if innerInset*2 < w.Width && innerInset*2 < w.Height {
			d.SetForeground(gc, selectPixel)
			d.FillRectangle(w.Drawable(), gc,
				innerInset, innerInset,
				uint(w.Width-2*innerInset), uint(w.Height-2*innerInset))
		}
		// Image transparent pixels show the widget's normal background
		// (not the selectcolor), so the bitmap stays readable on top of
		// the selectcolor frame.
	} else if hasSelectFill {
		// No image: selectcolor becomes the full background (pushbutton mode).
		d.SetForeground(gc, selectPixel)
		d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))
		bgPixel = selectPixel
	}

	inset := c.BorderWidth + c.HighlightWidth
	img := c.activeImage()
	var contentX, contentY int
	if img != nil {
		contentX, contentY = widget.ComputeAnchor(c.Anchor, w.Width, w.Height, inset, 0, 0, c.indicatorSpace+img.Width(), img.Height())
	} else {
		contentX, contentY = widget.ComputeAnchor(c.Anchor, w.Width, w.Height, inset, c.PadX, c.PadY, c.indicatorSpace+c.textWidth, c.textHeight)
	}
	contentX += c.indicatorSpace

	if c.IndicatorOn {
		// Tk centres the indicator in its column and on the window's mid-line.
		state := draw.IndicatorOff
		switch {
		case tristate:
			state = draw.IndicatorTristate
		case selected:
			state = draw.IndicatorOn
		}
		selPixel := uint64(0xffffff)
		if c.SelectColor != nil {
			selPixel = c.SelectColor.Pixel
		}
		var fgPixel, disPixel uint64 = 0, 0xa3a3a3
		if c.Foreground != nil {
			fgPixel = c.Foreground.Pixel
		}
		if c.DisabledFg != nil {
			disPixel = c.DisabledFg.Pixel
		}
		draw.DrawCheckIndicator(d, w.Drawable(), gc, w.Depth,
			contentX-c.indicatorSpace/2, w.Height/2, draw.CheckIndicator,
			draw.NewBorderFromPixel(bgPixel), fgPixel, selPixel, disPixel,
			state, c.State == widget.StateDisabled)
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
	} else if c.Font != nil && c.Text != "" && fgCol != nil {
		textX, textY := contentX, contentY
		m := c.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := c.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, c.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	c.DrawHighlightBorder(c.HasFocus, 0)

	d.Flush()
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
func (c *Checkbutton) Configure(opts ...option.Option) {
	option.Apply(c, opts)
	c.UpdateBorder()
	c.computeGeometry()
	if c.Background != nil {
		c.Win.BackgroundPixel = c.Background.Pixel
	}
	c.Display()
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
