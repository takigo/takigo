// Package checkbutton implements a toggle button with a square indicator.
// It ports the checkbutton-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package checkbutton

import (
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

	// Variable linkage.
	Variable *widget.Variable[bool]
	unsub    func()

	// Indicator.
	IndicatorOn   bool      // whether to draw the indicator (default true)
	Indeterminate bool      // shows a dash (partial/tri-state) instead of a checkmark
	SelectColor   *colorRef // indicator fill color when selected

	// Active colors (used on hover).
	ActiveBackground *colorRef
	ActiveForeground *colorRef

	textWidth  int
	textHeight int
	pressed    bool
}

// colorRef holds a resolved color reference.
type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// Option configures a Checkbutton.
type Option func(*Checkbutton)

// Text sets the checkbutton text.
func Text(s string) Option {
	return func(c *Checkbutton) { c.Text = s }
}

// Command sets the callback invoked when toggled.
func Command(fn func()) Option {
	return func(c *Checkbutton) { c.Command = fn }
}

// Var links the checkbutton to a boolean variable.
func Var(v *widget.Variable[bool]) Option {
	return func(c *Checkbutton) {
		if c.unsub != nil {
			c.unsub()
		}
		c.Variable = v
		c.unsub = v.OnChange(func(_, _ bool) {
			c.Display()
		})
	}
}

// Background sets the background color.
func Background(name string) Option {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err == nil {
			c.Background = col
			c.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) Option {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err == nil {
			c.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) Option {
	return func(c *Checkbutton) {
		f, err := c.App.FontRegistry().Get(name)
		if err == nil {
			c.Font = f
		}
	}
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) Option {
	return func(c *Checkbutton) { c.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) Option {
	return func(c *Checkbutton) { c.PadX = screenunit.Px(p) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) Option {
	return func(c *Checkbutton) { c.PadY = screenunit.Px(p) }
}

// indicatorSize is the side length of the square indicator.
const indicatorSize = 13

// New creates a new Checkbutton widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...Option) *Checkbutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	c := &Checkbutton{
		Anchor:      option.AnchorW,
		IndicatorOn: true,
	}
	widget.InitBase(&c.Base, w, app)

	// Checkbutton-specific defaults.
	c.BorderWidth = 0
	c.Relief = option.ReliefFlat
	c.PadX = 1
	c.PadY = 1
	c.HighlightWidth = 1

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		c.ActiveBackground = &colorRef{ac.Pixel, ac.Red, ac.Green, ac.Blue}
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		c.ActiveForeground = &colorRef{af.Pixel, af.Red, af.Green, af.Blue}
	}

	// Select color (indicator fill when checked).
	if sc, err := app.ColorCache().Get("#b03060"); err == nil {
		c.SelectColor = &colorRef{sc.Pixel, sc.Red, sc.Green, sc.Blue}
	}

	// Default variable.
	c.Variable = widget.NewVariable(false)

	for _, opt := range opts {
		opt(c)
	}

	c.computeGeometry()

	if c.Background != nil {
		w.BackgroundPixel = c.Background.Pixel
	}

	bindCheckbutton(c, app)

	return c
}

func (c *Checkbutton) computeGeometry() {
	if c.Font != nil && c.Text != "" {
		c.textWidth = c.Font.MeasureString(c.Text)
		m := c.Font.Metrics()
		c.textHeight = m.Linespace()
	} else {
		c.textWidth = 0
		c.textHeight = 0
	}

	inset := c.BorderWidth + c.HighlightWidth
	contentW := c.textWidth
	contentH := c.textHeight
	if c.IndicatorOn {
		contentW += indicatorSize + 4 // indicator + gap
		if indicatorSize > contentH {
			contentH = indicatorSize
		}
	}

	w := c.Win
	w.ReqWidth = contentW + 2*c.PadX + 2*inset
	w.ReqHeight = contentH + 2*c.PadY + 2*inset
}

// Selected returns whether the checkbutton is currently selected.
func (c *Checkbutton) Selected() bool {
	return c.Variable.Get()
}

// Display draws the checkbutton.
func (c *Checkbutton) Display() {
	if c.Destroyed {
		return
	}
	w := c.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	selected := c.Variable.Get()

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *colorRef
	if c.Background != nil {
		bgPixel = c.Background.Pixel
	}
	if c.Foreground != nil {
		fgCol = &colorRef{c.Foreground.Pixel, c.Foreground.Red, c.Foreground.Green, c.Foreground.Blue}
	}

	if c.State == widget.StateActive && c.ActiveBackground != nil {
		bgPixel = c.ActiveBackground.Pixel
	}
	if c.State == widget.StateActive && c.ActiveForeground != nil {
		fgCol = c.ActiveForeground
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border.
	if c.Border != nil && c.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, c.Border,
			0, 0, w.Width, w.Height, c.BorderWidth, c.Relief)
	}

	inset := c.BorderWidth + c.HighlightWidth
	availW := w.Width - 2*inset - 2*c.PadX
	availH := w.Height - 2*inset - 2*c.PadY
	frameX := inset + c.PadX
	frameY := inset + c.PadY

	indW := 0
	if c.IndicatorOn {
		indW = indicatorSize + 4

		// Draw square indicator.
		indX := frameX
		indY := frameY + (availH-indicatorSize)/2

		// Sunken border for indicator box.
		indBorder := c.Border
		if indBorder == nil {
			indBorder = draw.NewBorderFromPixel(bgPixel)
		}

		// Fill indicator.
		if (selected || c.Indeterminate) && c.SelectColor != nil {
			d.SetForeground(gc, c.SelectColor.Pixel)
		} else {
			d.SetForeground(gc, uint64(0xffffff)) // white background
		}
		d.FillRectangle(w.Drawable(), gc, indX+2, indY+2,
			uint(indicatorSize-4), uint(indicatorSize-4))

		// Draw sunken border around indicator.
		draw.Draw3DRectangle(d, w.Drawable(), gc, indBorder,
			indX, indY, indicatorSize, indicatorSize, 2, option.ReliefSunken)

		if c.Indeterminate && fgCol != nil {
			// Draw a horizontal dash for the indeterminate/partial state.
			d.SetForeground(gc, fgCol.Pixel)
			midY := indY + indicatorSize/2
			d.DrawLine(w.Drawable(), gc, indX+3, midY, indX+indicatorSize-4, midY)
			d.DrawLine(w.Drawable(), gc, indX+3, midY+1, indX+indicatorSize-4, midY+1)
		} else if selected && fgCol != nil {
			// Draw checkmark when selected.
			d.SetForeground(gc, fgCol.Pixel)
			cx := indX + 3
			cy := indY + indicatorSize/2
			d.DrawLine(w.Drawable(), gc, cx, cy, cx+2, cy+3)
			d.DrawLine(w.Drawable(), gc, cx+1, cy, cx+3, cy+3)
			d.DrawLine(w.Drawable(), gc, cx+2, cy+3, cx+7, cy-2)
			d.DrawLine(w.Drawable(), gc, cx+3, cy+3, cx+8, cy-2)
		}
	}

	// Draw text.
	if c.Font != nil && c.Text != "" && fgCol != nil {
		textX := frameX + indW
		textY := frameY + (availH-c.textHeight)/2
		// Apply anchor for remaining space.
		remainW := availW - indW
		switch c.Anchor {
		case option.AnchorCenter, option.AnchorN, option.AnchorS:
			textX += (remainW - c.textWidth) / 2
		case option.AnchorE, option.AnchorNE, option.AnchorSE:
			textX += remainW - c.textWidth
		}
		m := c.Font.Metrics()
		baseline := textY + m.Ascent
		if df, ok := c.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, baseline, c.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	d.Flush()
}

// SetIndeterminate sets the indeterminate (partial tri-state) display flag and redraws.
func (c *Checkbutton) SetIndeterminate(v bool) {
	c.Indeterminate = v
	c.Display()
}

// Toggle flips the checkbutton state.
func (c *Checkbutton) Toggle() {
	if c.State == widget.StateDisabled {
		return
	}
	c.Variable.Set(!c.Variable.Get())
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
