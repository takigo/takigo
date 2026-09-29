package ttk

import (
	"log"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Checkbutton is a TTK themed toggle button with a checkbox indicator.
type Checkbutton struct {
	TtkWidget

	Text     string
	Font     font.Font
	Command  func()
	Variable *widget.Variable[bool]

	selected bool
	unsub    func()
	// layoutMode is set for styles such as Toolbutton whose layout has no
	// indicator: the widget is drawn from the style's layout like a button.
	layoutMode bool
}

// GetText implements TextProvider.
func (c *Checkbutton) GetText() string { return c.Text }

// GetFont implements TextProvider.
func (c *Checkbutton) GetFont() font.Font { return c.Font }

// GetImage implements TextProvider.
func (c *Checkbutton) GetImage() widget.WidgetImage { return nil }

// GetCompound implements TextProvider.
func (c *Checkbutton) GetCompound() widget.Compound { return widget.CompoundNone }

// CheckbuttonStyleOpt sets -style, e.g. "Toolbutton".
func CheckbuttonStyleOpt(name string) CheckbuttonOption {
	return func(c *Checkbutton) { c.StyleName = name }
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
		if err != nil {
			log.Printf("ttk.checkbutton: failed to get font %q: %v", name, err)
			return
		}
		c.Font = f
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

// CheckbuttonAlternate sets the initial state to include StateAlternate.
// Matches Tk's behavior when the linked variable is unset.
func CheckbuttonAlternate() CheckbuttonOption {
	return func(c *Checkbutton) { c.State |= StateAlternate }
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
	win.OnDestroy(c.Destroy)
	c.DisplayFunc = c.Display

	for _, opt := range opts {
		opt(c)
	}

	// Sync initial state from variable.
	if c.Variable != nil {
		c.selected = c.Variable.Get()
		if c.selected {
			c.State |= StateSelected
			c.State &^= StateAlternate
		}
		c.unsub = c.Variable.OnChange(func(_, _ bool) {
			c.selected = c.Variable.Get()
			if c.selected {
				c.State |= StateSelected
			} else {
				c.State &^= StateSelected
			}
			c.State &^= StateAlternate
			c.Display()
		})
	}

	if c.StyleName != "TCheckbutton" && c.Theme != nil {
		if tmpl := c.Theme.GetLayout(c.StyleName); tmpl != nil {
			c.setStyle(c.Theme.ResolveStyle(c.StyleName))
			c.LabelFactory = NewLabelElementFactory(c)
			c.Layout = newLayoutWithLabel(tmpl, c.Theme, c.Context, c.Context.Style, c.LabelFactory)
			c.layoutMode = true
			bindTtkHover(&c.TtkWidget, app)
		}
	}

	// Compute initial size.
	c.computeSize()

	bindTtkCheckbutton(c, app)
	return c
}

func (c *Checkbutton) computeSize() {
	if c.layoutMode {
		c.Win.ReqWidth, c.Win.ReqHeight = c.Layout.Size(c.State)
		return
	}
	textW, textH := 0, 0
	if c.Font != nil {
		textW = c.Font.MeasureString(c.Text)
		textH = c.Font.Metrics().Linespace()
	}
	var style *Style
	if c.Context != nil {
		style = c.Context.Style
	}
	c.Win.ReqWidth, c.Win.ReqHeight = newIndicatorLayout(style, c.State).reqSize(textW, textH)
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
	if c.layoutMode {
		c.TtkWidget.Display()
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
	hasClamStyle := false
	var upperBorderColor, lowerBorderColor uint64
	if c.Context != nil && c.Context.Style != nil {
		upperBorderColor, hasClamStyle = c.Context.Style.LookupAs[uint64]("-upperbordercolor", c.State)
		lowerBorderColor, _ = c.Context.Style.LookupAs[uint64]("-lowerbordercolor", c.State)
	}

	// Indicator (checkbox square).
	textH := 0
	if c.Font != nil {
		textH = c.Font.Metrics().Linespace()
	}
	var layoutStyle *Style
	if c.Context != nil {
		layoutStyle = c.Context.Style
	}
	lay := newIndicatorLayout(layoutStyle, c.State)
	indX, indY, labelX, labelY := lay.place(height, textH)
	indSize := lay.size

	// Indicator fill and color.
	indFill := uint64(0xffffff)
	indColor := uint64(0xffffff) // default to white (matches Tcl default theme)
	if c.Context != nil && c.Context.Style != nil {
		indFill = LookupColor(c.Context.Style, "-indicatorbackground", c.State, indFill)
		indColor = LookupColor(c.Context.Style, "-indicatorforeground", c.State, indColor)
	}
	if c.State&StateDisabled != 0 {
		indFill = bgColor
	}

	// Check for classic Motif indicator style: -indicatorrelief set in style.
	hasClassicStyle := false
	classicRelief := option.ReliefFlat
	if c.Context != nil && c.Context.Style != nil {
		if r, ok := c.Context.Style.LookupAs[option.Relief]("-indicatorrelief", c.State); ok {
			classicRelief, hasClassicStyle = r, true
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

		// X mark when selected, horizontal line when alternate.
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
		} else if c.State&StateAlternate != 0 {
			lineY := indY + indSize/2
			d.SetForeground(gc, fgColor)
			d.DrawLine(win.Drawable(), gc, indX+3, lineY, indX+indSize-4, lineY)
			d.DrawLine(win.Drawable(), gc, indX+3, lineY+1, indX+indSize-4, lineY+1)
		}
	} else {
		// IndicatorElementDraw (ttkElements.c): -bordercolor is not set by
		// the default theme, so the element default #888888 applies.
		st := draw.IndicatorOff
		if c.State&StateAlternate != 0 {
			st = draw.IndicatorTristate
		} else if c.selected {
			st = draw.IndicatorOn
		}
		indBg := LookupColor(c.Context.Style, "-indicatorbackground", c.State, 0xffffff)
		draw.DrawTtkIndicator(d, win.Drawable(), gc, win.Depth, indX, indY, indSize, false,
			st, indBg, indColor, 0x888888, bgColor)
	}

	// Text label.
	if c.Font != nil && c.Text != "" {
		if df, ok := c.Font.(platform.DrawableFont); ok {
			m := c.Font.Metrics()
			textX := labelX
			textY := labelY + m.Ascent
			r := uint16((fgColor>>16)&0xFF) * 257
			g := uint16((fgColor>>8)&0xFF) * 257
			b := uint16((fgColor)&0xFF) * 257
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
		// The focus element surrounds the label only.
		textW := 0
		if c.Font != nil {
			textW = c.Font.MeasureString(c.Text)
		}
		d.DrawRectangle(win.Drawable(), gc, labelX-focusThickness, labelY-focusThickness,
			uint(textW+2*focusThickness-1), uint(textH+2*focusThickness-1))
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
	c.State &^= StateAlternate
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

	// Enter → set hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		c.ChangeState(StateHover|StateActive, 0)
	})

	// Leave → clear hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		c.ChangeState(0, StateHover|StateActive|StatePressed)
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
