package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

const ttkIndicatorSize = 13

// Checkbutton is a TTK themed toggle button with a checkbox indicator.
type Checkbutton struct {
	TtkWidget

	Text     string
	Font     font.Font
	Command  func()
	Variable *widget.Variable[bool]

	selected bool
	unsub    func()
}

// CheckbuttonOption configures a Checkbutton.
type CheckbuttonOption func(*Checkbutton)

// CheckbuttonText sets the button label.
func CheckbuttonText(s string) CheckbuttonOption {
	return func(c *Checkbutton) { c.Text = s }
}

// CheckbuttonFont sets the font.
func CheckbuttonFont(name string) CheckbuttonOption {
	return func(c *Checkbutton) {
		f, err := c.App.FontRegistry().Get(name)
		if err == nil {
			c.Font = f
		}
	}
}

// CheckbuttonCommand sets the callback invoked on toggle.
func CheckbuttonCommand(fn func()) CheckbuttonOption {
	return func(c *Checkbutton) { c.Command = fn }
}

// CheckbuttonVar links the button to a bool variable.
func CheckbuttonVar(v *widget.Variable[bool]) CheckbuttonOption {
	return func(c *Checkbutton) { c.Variable = v }
}

// NewCheckbutton creates a TTK themed checkbutton.
func NewCheckbutton(parent widget.Caregiver, name string, opts ...CheckbuttonOption) *Checkbutton {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	win.Flags |= window.FlagFocusable

	c := &Checkbutton{}
	c.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&c.TtkWidget, win, app, "TCheckbutton")
	c.DisplayFunc = c.Display

	for _, opt := range opts {
		opt(c)
	}

	// Sync initial state from variable.
	if c.Variable != nil {
		c.selected = c.Variable.Get()
		if c.selected {
			c.State |= StateSelected
		}
		c.unsub = c.Variable.OnChange(func(_, _ bool) {
			c.selected = c.Variable.Get()
			if c.selected {
				c.State |= StateSelected
			} else {
				c.State &^= StateSelected
			}
			c.Display()
		})
	}

	// Compute initial size.
	c.computeSize()

	bindTtkCheckbutton(c, app)
	return c
}

func (c *Checkbutton) computeSize() {
	w := c.Win
	indSize := ttkIndicatorSize
	if c.Context != nil && c.Context.Style != nil {
		if sz := LookupInt(c.Context.Style, "-indicatorsize", c.State, 0); sz > 0 {
			indSize = sz
		}
	}
	indW := indSize + 4
	textW, textH := 0, 0
	if c.Font != nil && c.Text != "" {
		textW = c.Font.MeasureString(c.Text)
		textH = c.Font.Metrics().Linespace()
	}
	h := max(textH, indSize)
	w.ReqWidth = indW + textW + 8
	w.ReqHeight = h + 6
}

// Display renders the checkbutton with custom drawing.
func (c *Checkbutton) Display() {
	if c.Destroyed {
		return
	}
	win := c.Win
	if win.PlatformID == 0 {
		return
	}

	d := win.Display.Server
	gc := win.GC
	width := win.Width
	height := win.Height
	if width <= 0 || height <= 0 {
		return
	}

	bgColor := uint64(0xd9d9d9)
	fgColor := uint64(0x000000)
	if c.Context != nil && c.Context.Style != nil {
		bgColor = LookupColor(c.Context.Style, "-background", c.State, bgColor)
		fgColor = LookupColor(c.Context.Style, "-foreground", c.State, fgColor)
	}

	// Background.
	d.SetForeground(gc, bgColor)
	d.FillRectangle(win.Drawable(), gc, 0, 0, uint(width), uint(height))

	// Determine indicator style from style options.
	indSize := ttkIndicatorSize
	hasClamStyle := false
	var upperBorderColor, lowerBorderColor uint64
	if c.Context != nil && c.Context.Style != nil {
		if sz := LookupInt(c.Context.Style, "-indicatorsize", c.State, 0); sz > 0 {
			indSize = sz
		}
		if v, ok := c.Context.Style.Lookup("-upperbordercolor", c.State); ok {
			if uc, ok2 := v.(uint64); ok2 {
				upperBorderColor = uc
				hasClamStyle = true
			}
		}
		if v, ok := c.Context.Style.Lookup("-lowerbordercolor", c.State); ok {
			if lc, ok2 := v.(uint64); ok2 {
				lowerBorderColor = lc
			}
		}
	}

	// Indicator (checkbox square).
	indX := 3
	indY := (height - indSize) / 2

	// Indicator fill and color.
	indFill := uint64(0xffffff)
	indColor := fgColor
	if c.Context != nil && c.Context.Style != nil {
		indFill = LookupColor(c.Context.Style, "-indicatorbackground", c.State, indFill)
		indColor = LookupColor(c.Context.Style, "-indicatorcolor", c.State, fgColor)
	}
	if c.State&StateDisabled != 0 {
		indFill = bgColor
	}

	// Check for classic Motif indicator style: -indicatorrelief set in style.
	hasClassicStyle := false
	classicRelief := option.ReliefFlat
	if c.Context != nil && c.Context.Style != nil {
		if v, ok := c.Context.Style.Lookup("-indicatorrelief", c.State); ok {
			if r, ok2 := v.(option.Relief); ok2 {
				classicRelief = r
				hasClassicStyle = true
			}
		}
	}

	if hasClassicStyle {
		// Classic Motif style: 3D raised/sunken square; relief encodes on/off state.
		border := draw.NewBorderFromPixel(bgColor)
		d.SetForeground(gc, indFill)
		d.FillRectangle(win.Drawable(), gc, indX, indY, uint(indSize), uint(indSize))
		if c.State&StateDisabled == 0 {
			draw.Draw3DRectangle(d, win.Drawable(), gc, border,
				indX, indY, indSize, indSize, 2, classicRelief)
		}
	} else if hasClamStyle {
		// Clam style: 1px flat border (upper-left = darkest, lower-right = dark), white fill.
		d.SetForeground(gc, indFill)
		d.FillRectangle(win.Drawable(), gc, indX+1, indY+1,
			uint(indSize-2), uint(indSize-2))

		// Top + left border line.
		d.SetForeground(gc, upperBorderColor)
		d.DrawLine(win.Drawable(), gc, indX, indY, indX+indSize-1, indY)
		d.DrawLine(win.Drawable(), gc, indX, indY, indX, indY+indSize-1)

		// Bottom + right border line.
		d.SetForeground(gc, lowerBorderColor)
		d.DrawLine(win.Drawable(), gc, indX, indY+indSize-1, indX+indSize-1, indY+indSize-1)
		d.DrawLine(win.Drawable(), gc, indX+indSize-1, indY, indX+indSize-1, indY+indSize-1)

		// X mark when selected.
		if c.selected {
			d.SetForeground(gc, fgColor)
			// Cross from (5,5) to (11,11) and (11,5) to (5,11) in 16px space.
			// Scale relative to indSize.
			x0 := indX + indSize*5/16
			y0 := indY + indSize*5/16
			x1 := indX + indSize*11/16
			y1 := indY + indSize*11/16
			// Draw thick X (2px wide).
			d.DrawLine(win.Drawable(), gc, x0, y0, x1, y1)
			d.DrawLine(win.Drawable(), gc, x0+1, y0, x1+1, y1)
			d.DrawLine(win.Drawable(), gc, x0, y0+1, x1, y1+1)
			d.DrawLine(win.Drawable(), gc, x1, y0, x0, y1)
			d.DrawLine(win.Drawable(), gc, x1-1, y0, x0-1, y1)
			d.DrawLine(win.Drawable(), gc, x1, y0+1, x0, y1+1)
		}
	} else {
		// Default style: 3D sunken border, checkmark.
		border := draw.NewBorderFromPixel(bgColor)
		d.SetForeground(gc, indFill)
		d.FillRectangle(win.Drawable(), gc, indX+2, indY+2,
			uint(indSize-4), uint(indSize-4))

		// Sunken border around checkbox (flat when disabled).
		if c.State&StateDisabled == 0 {
			draw.Draw3DRectangle(d, win.Drawable(), gc, border,
				indX, indY, indSize, indSize, 2, option.ReliefSunken)
		}

		// Checkmark when selected.
		if c.selected {
			d.SetForeground(gc, indColor)
			cx := indX + 3
			cy := indY + indSize/2
			d.DrawLine(win.Drawable(), gc, cx, cy, cx+2, cy+3)
			d.DrawLine(win.Drawable(), gc, cx+1, cy, cx+3, cy+3)
			d.DrawLine(win.Drawable(), gc, cx+2, cy+3, cx+7, cy-2)
			d.DrawLine(win.Drawable(), gc, cx+3, cy+3, cx+8, cy-2)
		}
	}

	// Text label.
	if c.Font != nil && c.Text != "" {
		if df, ok := c.Font.(platform.DrawableFont); ok {
			m := c.Font.Metrics()
			textX := indX + indSize + 4
			textY := (height-m.Linespace())/2 + m.Ascent
			r := uint16((fgColor>>16)&0xFF) << 8
			g := uint16((fgColor>>8)&0xFF) << 8
			b := uint16((fgColor)&0xFF) << 8
			df.DrawString(win.Drawable(), textX, textY, c.Text, fgColor, r, g, b)
		}
	}

	// Focus ring.
	if c.State&StateFocus != 0 {
		focusColor := uint64(0x000000)
		if c.Context != nil && c.Context.Style != nil {
			focusColor = LookupColor(c.Context.Style, "-focuscolor", c.State, focusColor)
		}
		d.SetForeground(gc, focusColor)
		d.DrawRectangle(win.Drawable(), gc, 0, 0, uint(width-1), uint(height-1))
	}

	d.Flush()
}

// Toggle toggles the checkbutton state.
func (c *Checkbutton) Toggle() {
	if c.State&StateDisabled != 0 {
		return
	}
	c.selected = !c.selected
	if c.selected {
		c.State |= StateSelected
	} else {
		c.State &^= StateSelected
	}
	if c.Variable != nil {
		c.Variable.Set(c.selected)
	}
	c.Display()
	if c.Command != nil {
		c.Command()
	}
}

func bindTtkCheckbutton(c *Checkbutton, app widget.AppContext) {
	win := c.Win

	// Expose — override bindTtkCommon.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.Display()
	})

	// Configure — override bindTtkCommon.
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			c.Display()
		}
	})

	// Enter → set hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		c.ChangeState(StateHover|StateActive, 0)
		c.Display()
	})

	// Leave → clear hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		c.ChangeState(0, StateHover|StateActive|StatePressed)
		c.Display()
	})

	// Focus → redraw to show/hide focus ring.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		c.Display()
	})

	// Button1 press → +StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			c.ChangeState(StatePressed, 0)
		}
	})

	// Button1 release → toggle if inside.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			wasPressed := c.State&StatePressed != 0
			c.ChangeState(0, StatePressed)
			if wasPressed && ev.X >= 0 && ev.X < win.Width && ev.Y >= 0 && ev.Y < win.Height {
				c.Toggle()
			}
		}
	})

	// Space key → toggle.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			c.Toggle()
		}
	})
}

// Destroy cleans up the checkbutton, unsubscribing from any linked Variable.
func (c *Checkbutton) Destroy() {
	if c.unsub != nil {
		c.unsub()
		c.unsub = nil
	}
	c.TtkWidget.Destroy()
}
