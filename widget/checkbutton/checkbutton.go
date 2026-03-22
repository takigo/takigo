// Package checkbutton implements a toggle button with a square indicator.
// It ports the checkbutton-specific parts of tk/generic/tkButton.c and
// library/button.tcl.
package checkbutton

import (
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

	// Variable linkage.
	Variable *widget.Variable[bool]
	unsub    func()

	// Indicator.
	IndicatorOn   bool            // whether to draw the indicator (default true)
	Indeterminate bool            // shows a dash (partial/tri-state) instead of a checkmark
	SelectColor   *color.ColorRef // indicator fill color when selected

	// Images (selectimage shown when checked; image shown otherwise).
	Img       widget.WidgetImage
	SelectImg widget.WidgetImage

	// Active colors (used on hover).
	ActiveBackground *color.ColorRef
	ActiveForeground *color.ColorRef

	textWidth  int
	textHeight int
	pressed    bool
	HasFocus   bool
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

// Var links the checkbutton to a boolean variable.
func Var(v *widget.Variable[bool]) CheckbuttonOption {
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
func Background(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err == nil {
			c.Background = col
			c.UpdateBorder()
		}
	}
}

// Foreground sets the text color.
func Foreground(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		col, err := c.App.ColorCache().Get(name)
		if err == nil {
			c.Foreground = col
		}
	}
}

// FontOpt sets the font.
func FontOpt(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		f, err := c.App.FontRegistry().Get(name)
		if err == nil {
			c.Font = f
		}
	}
}

// Anchor sets the text anchor.
func Anchor(a option.Anchor) CheckbuttonOption {
	return func(c *Checkbutton) { c.Anchor = a }
}

// PadX sets horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadX = screenunit.Px(p) }
}

// PadY sets vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) CheckbuttonOption {
	return func(c *Checkbutton) { c.PadY = screenunit.Px(p) }
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

// indicatorSize is the side length of the square indicator.
const indicatorSize = 13

// New creates a new Checkbutton widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...CheckbuttonOption) *Checkbutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	w.Flags |= window.FlagFocusable

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
		c.ActiveBackground = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		c.ActiveForeground = af.Ref()
	}

	// Select color (indicator fill when checked).
	if sc, err := app.ColorCache().Get("#b03060"); err == nil {
		c.SelectColor = sc.Ref()
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

	// Determine content size from image or text.
	img := c.activeImage()
	inset := c.BorderWidth + c.HighlightWidth
	var contentW, contentH int
	if img != nil {
		contentW = img.Width()
		contentH = img.Height()
	} else {
		contentW = c.textWidth
		contentH = c.textHeight
	}
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

// activeImage returns the image to display based on current state.
func (c *Checkbutton) activeImage() widget.WidgetImage {
	selected := c.Variable.Get()
	if selected && c.SelectImg != nil {
		return c.SelectImg
	}
	return c.Img
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

	// In toggle mode, use select color as background when selected.
	if !c.IndicatorOn && selected && c.SelectColor != nil {
		bgPixel = c.SelectColor.Pixel
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw border (inset by highlight width so highlight ring is outermost).
	hlw := c.HighlightWidth
	if !c.IndicatorOn {
		// Toggle button mode: raised or sunken relief based on selection.
		btnRelief := option.ReliefRaised
		if selected {
			btnRelief = option.ReliefSunken
		}
		bw := 2
		border := c.Border
		if border == nil {
			border = draw.NewBorderFromPixel(bgPixel)
		}
		draw.Draw3DRectangle(d, w.Drawable(), gc, border, hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, bw, btnRelief)
	} else if c.Border != nil && c.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, c.Border,
			hlw, hlw, w.Width-2*hlw, w.Height-2*hlw, c.BorderWidth, c.Relief)
	}

	inset := c.BorderWidth + c.HighlightWidth
	if !c.IndicatorOn {
		inset = 2 + c.HighlightWidth
	}
	availW := max(0, w.Width-2*inset-2*c.PadX)
	availH := max(0, w.Height-2*inset-2*c.PadY)
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

	// Draw image (if set) or text.
	img := c.activeImage()
	if img != nil {
		imgW := img.Width()
		imgH := img.Height()
		imgX := frameX + indW + (availW-indW-imgW)/2
		imgY := frameY + (availH-imgH)/2
		if photo, ok := img.(interface {
			Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
				depth int, imgX, imgY, w, h, dstX, dstY int, bgPixel uint64)
		}); ok {
			photo.Draw(d, w.Drawable(), gc, w.Depth, 0, 0, imgW, imgH, imgX, imgY, bgPixel)
		}
	} else if c.Font != nil && c.Text != "" && fgCol != nil {
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

	c.DrawHighlightBorder(c.HasFocus, 0)

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
