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
	// No layout registered for TCheckbutton → TtkWidget.Display() is a no-op.

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
	indW := ttkIndicatorSize + 4
	textW, textH := 0, 0
	if c.Font != nil && c.Text != "" {
		textW = c.Font.MeasureString(c.Text)
		textH = c.Font.Metrics().Linespace()
	}
	h := max(textH, ttkIndicatorSize)
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

	border := draw.NewBorderFromPixel(bgColor)

	// Indicator (checkbox square).
	indX := 3
	indSize := ttkIndicatorSize
	indY := (height - indSize) / 2

	// White fill for checkbox.
	d.SetForeground(gc, uint64(0xffffff))
	d.FillRectangle(win.Drawable(), gc, indX+2, indY+2,
		uint(indSize-4), uint(indSize-4))

	// Sunken border around checkbox.
	draw.Draw3DRectangle(d, win.Drawable(), gc, border,
		indX, indY, indSize, indSize, 2, option.ReliefSunken)

	// Checkmark when selected.
	if c.selected {
		d.SetForeground(gc, fgColor)
		cx := indX + 3
		cy := indY + indSize/2
		d.DrawLine(win.Drawable(), gc, cx, cy, cx+2, cy+3)
		d.DrawLine(win.Drawable(), gc, cx+1, cy, cx+3, cy+3)
		d.DrawLine(win.Drawable(), gc, cx+2, cy+3, cx+7, cy-2)
		d.DrawLine(win.Drawable(), gc, cx+3, cy+3, cx+8, cy-2)
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
