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

// Radiobutton is a TTK themed mutually-exclusive selection button.
type Radiobutton struct {
	TtkWidget

	Text     string
	Font     font.Font
	Command  func()
	Variable *widget.Variable[string]
	Value    string

	unsub func()
}

// RadiobuttonOption configures a Radiobutton.
type RadiobuttonOption func(*Radiobutton)

// RadiobuttonText sets the button label.
func RadiobuttonText(s string) RadiobuttonOption {
	return func(r *Radiobutton) { r.Text = s }
}

// RadiobuttonValue sets the value this radiobutton represents.
func RadiobuttonValue(v string) RadiobuttonOption {
	return func(r *Radiobutton) { r.Value = v }
}

// RadiobuttonVar links the button to a string variable (shared across group).
func RadiobuttonVar(v *widget.Variable[string]) RadiobuttonOption {
	return func(r *Radiobutton) {
		if r.unsub != nil {
			r.unsub()
			r.unsub = nil
		}
		r.Variable = v
		r.unsub = v.OnChange(func(_, _ string) {
			if r.Variable.Get() == r.Value {
				r.State |= StateSelected
			} else {
				r.State &^= StateSelected
			}
			r.State &^= StateAlternate
			r.Display()
		})
	}
}

// RadiobuttonCommand sets the callback invoked on selection.
func RadiobuttonCommand(fn func()) RadiobuttonOption {
	return func(r *Radiobutton) { r.Command = fn }
}

// RadiobuttonAlternate sets the initial state to include StateAlternate.
// Matches Tk's behavior when the linked variable is unset.
func RadiobuttonAlternate() RadiobuttonOption {
	return func(r *Radiobutton) { r.State |= StateAlternate }
}

// NewRadiobutton creates a TTK themed radiobutton.
func NewRadiobutton(parent widget.Caregiver, name string, opts ...RadiobuttonOption) *Radiobutton {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	win.Flags |= window.FlagFocusable

	r := &Radiobutton{}
	r.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)
	r.Variable = widget.NewVariable("")

	InitTtkWidget(&r.TtkWidget, win, app, "TRadiobutton")
	r.reconfigure = func() { _ = r.Configure() }
	win.OnDestroy(r.Destroy)
	r.DisplayFunc = r.Display

	for _, opt := range opts {
		opt(r)
	}

	r.syncSelected()

	r.computeSize()
	bindTtkRadiobutton(r, app)
	return r
}

// syncSelected sets the selected state from -variable and -value.
func (r *Radiobutton) syncSelected() {
	if r.Variable != nil && r.Variable.Get() == r.Value {
		r.State |= StateSelected
		r.State &^= StateAlternate
	} else {
		r.State &^= StateSelected
	}
}

// Configure sets options after creation.
func (r *Radiobutton) Configure(opts ...RadiobuttonOption) error {
	return configure(&r.TtkWidget, r, opts, r.syncSelected, r.computeSize)
}

func (r *Radiobutton) computeSize() {
	textW, textH := 0, 0
	if r.Font != nil {
		textW = r.Font.MeasureString(r.Text)
		textH = r.Font.Metrics().Linespace()
	}
	var style *Style
	if r.Context != nil {
		style = r.Context.Style
	}
	r.Win.ReqWidth, r.Win.ReqHeight = newIndicatorLayout(style, r.State).reqSize(textW, textH)
}

// Selected returns whether this radiobutton is currently selected.
func (r *Radiobutton) Selected() bool {
	if r.Variable == nil {
		return false
	}
	return r.Variable.Get() == r.Value
}

// Display renders the radiobutton.
func (r *Radiobutton) Display() {
	if r.Destroyed {
		return
	}
	win := r.Win
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
	if r.Context != nil && r.Context.Style != nil {
		bgColor = LookupColor(r.Context.Style, "-background", r.State, bgColor)
		fgColor = LookupColor(r.Context.Style, "-foreground", r.State, fgColor)
	}

	// Background.
	d.SetForeground(gc, bgColor)
	d.FillRectangle(win.Drawable(), gc, 0, 0, uint(width), uint(height))

	// Determine indicator style from style options.
	hasClamStyle := false
	var upperBorderColor, lowerBorderColor uint64
	if r.Context != nil && r.Context.Style != nil {
		upperBorderColor, hasClamStyle = r.Context.Style.LookupAs[uint64]("-upperbordercolor", r.State)
		lowerBorderColor, _ = r.Context.Style.LookupAs[uint64]("-lowerbordercolor", r.State)
	}

	selected := r.Selected()

	// Indicator (circle).
	textH := 0
	if r.Font != nil {
		textH = r.Font.Metrics().Linespace()
	}
	var layoutStyle *Style
	if r.Context != nil {
		layoutStyle = r.Context.Style
	}
	lay := newIndicatorLayout(layoutStyle, r.State)
	indX, indY, labelX, labelY := lay.place(height, textH)
	indSize := lay.size

	// Indicator fill and color.
	indFill := uint64(0xffffff)
	indColor := uint64(0xffffff) // default to white (matches Tcl default theme)
	if r.Context != nil && r.Context.Style != nil {
		indFill = LookupColor(r.Context.Style, "-indicatorbackground", r.State, indFill)
		indColor = LookupColor(r.Context.Style, "-indicatorforeground", r.State, indColor)
	}
	if r.State&StateDisabled != 0 {
		indFill = bgColor
	}

	// Check for classic Motif indicator style: -indicatorrelief set in style.
	hasClassicStyle := false
	classicRelief := option.ReliefFlat
	if r.Context != nil && r.Context.Style != nil {
		if rel, ok := r.Context.Style.LookupAs[option.Relief]("-indicatorrelief", r.State); ok {
			classicRelief, hasClassicStyle = rel, true
		}
	}

	if hasClassicStyle {
		// Classic Motif style: 3D raised/sunken diamond; relief encodes selected state.
		border := draw.NewBorderFromPixel(bgColor)
		radius := indSize / 2
		pts := []draw.Point{
			{X: indX, Y: indY + radius},
			{X: indX + radius, Y: indY + indSize - 1},
			{X: indX + indSize - 1, Y: indY + radius},
			{X: indX + radius, Y: indY},
		}
		d.SetForeground(gc, indFill)
		draw.FillPolygon(d, win.Drawable(), gc, pts)
		if r.State&StateDisabled == 0 {
			drawDiamond3DRadio(d, win.Drawable(), gc, border, pts, 2, classicRelief)
		}
	} else if hasClamStyle {
		// Clam style: flat two-color circle border.
		// Outer circle with gradient effect (upper-left = darkest, lower-right = dark).
		d.SetForeground(gc, upperBorderColor)
		d.FillArc(win.Drawable(), gc, indX, indY, uint(indSize), uint(indSize), 0, 360*64)

		// Draw lower-right half in lighter color for gradient effect.
		d.SetForeground(gc, lowerBorderColor)
		d.FillArc(win.Drawable(), gc, indX, indY, uint(indSize), uint(indSize), 225*64, 180*64)

		// Inner fill circle (1px border).
		d.SetForeground(gc, indFill)
		d.FillArc(win.Drawable(), gc, indX+1, indY+1, uint(indSize-2), uint(indSize-2), 0, 360*64)

		// Inner dot when selected, horizontal line when alternate.
		if selected {
			dotSize := indSize / 2
			dotX := indX + (indSize-dotSize)/2
			dotY := indY + (indSize-dotSize)/2
			d.SetForeground(gc, fgColor)
			d.FillArc(win.Drawable(), gc, dotX, dotY, uint(dotSize), uint(dotSize), 0, 360*64)
		} else if r.State&StateAlternate != 0 {
			lineY := indY + indSize/2
			d.SetForeground(gc, fgColor)
			d.DrawLine(win.Drawable(), gc, indX+3, lineY, indX+indSize-4, lineY)
			d.DrawLine(win.Drawable(), gc, indX+3, lineY+1, indX+indSize-4, lineY+1)
		}
	} else {
		// IndicatorElementDraw (ttkElements.c): -bordercolor is not set by
		// the default theme, so the element default #888888 applies.
		st := draw.IndicatorOff
		if r.State&StateAlternate != 0 {
			st = draw.IndicatorTristate
		} else if selected {
			st = draw.IndicatorOn
		}
		indBg := LookupColor(r.Context.Style, "-indicatorbackground", r.State, 0xffffff)
		draw.DrawTtkIndicator(d, win.Drawable(), gc, win.Depth, indX, indY, indSize, true,
			st, indBg, indColor, 0x888888, bgColor)
	}

	// Text label.
	if r.Font != nil && r.Text != "" {
		if df, ok := r.Font.(platform.DrawableFont); ok {
			m := r.Font.Metrics()
			textX := labelX
			textY := labelY + m.Ascent
			rv := uint16((fgColor>>16)&0xFF) * 257
			gv := uint16((fgColor>>8)&0xFF) * 257
			bv := uint16((fgColor)&0xFF) * 257
			df.DrawString(win.Drawable(), textX, textY, r.Text, fgColor, rv, gv, bv)
		}
	}

	// Focus ring.
	if r.State&StateFocus != 0 {
		focusColor := uint64(0x000000)
		if r.Context != nil && r.Context.Style != nil {
			focusColor = LookupColor(r.Context.Style, "-focuscolor", r.State, focusColor)
		}
		d.SetForeground(gc, focusColor)
		// The focus element surrounds the label only.
		textW := 0
		if r.Font != nil {
			textW = r.Font.MeasureString(r.Text)
		}
		d.DrawRectangle(win.Drawable(), gc, labelX-focusThickness, labelY-focusThickness,
			uint(textW+2*focusThickness-1), uint(textH+2*focusThickness-1))
	}

	d.Flush()
}

// Select selects this radiobutton (sets the variable to this button's value).
func (r *Radiobutton) Select() {
	if r.State&StateDisabled != 0 {
		return
	}
	r.State |= StateSelected
	r.State &^= StateAlternate
	if r.Variable != nil {
		r.Variable.Set(r.Value)
	}
	r.Display()
	if r.Command != nil {
		r.Command()
	}
}

func bindTtkRadiobutton(r *Radiobutton, app widget.AppContext) {
	win := r.Win

	// Enter → set hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		r.ChangeState(StateHover|StateActive, 0)
	})

	// Leave → clear hover state and redraw.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		r.ChangeState(0, StateHover|StateActive|StatePressed)
	})

	// Button1 press → +StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			r.ChangeState(StatePressed, 0)
		}
	})

	// Button1 release → select if inside.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			wasPressed := r.State&StatePressed != 0
			r.ChangeState(0, StatePressed)
			if wasPressed && ev.X >= 0 && ev.X < win.Width && ev.Y >= 0 && ev.Y < win.Height {
				r.Select()
			}
		}
	})

	// Space key → select.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_space {
			r.Select()
		}
	})
}

// drawDiamond3DRadio draws 3D shaded edges for a diamond indicator.
// Points order: left, bottom, right, top.
func drawDiamond3DRadio(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, pts []draw.Point, bw int, relief option.Relief) {
	var lightPx, darkPx uint64
	switch relief {
	case option.ReliefRaised, option.ReliefRidge:
		lightPx = border.LightPixel
		darkPx = border.DarkPixel
	case option.ReliefSunken, option.ReliefGroove:
		lightPx = border.DarkPixel
		darkPx = border.LightPixel
	default:
		lightPx = border.BgPixel
		darkPx = border.BgPixel
	}
	left, bottom, right, top := pts[0], pts[1], pts[2], pts[3]
	for i := range bw {
		// Upper half: top→left and top→right (light for raised).
		d.SetForeground(gc, lightPx)
		d.DrawLine(drawable, gc, top.X, top.Y+i, left.X+i, left.Y)
		d.DrawLine(drawable, gc, top.X, top.Y+i, right.X-i, right.Y)
		// Lower half: bottom→left and bottom→right (dark for raised).
		d.SetForeground(gc, darkPx)
		d.DrawLine(drawable, gc, bottom.X, bottom.Y-i, left.X+i, left.Y)
		d.DrawLine(drawable, gc, bottom.X, bottom.Y-i, right.X-i, right.Y)
	}
}

// Destroy cleans up the radiobutton, unsubscribing from any linked Variable.
func (r *Radiobutton) Destroy() {
	if r.unsub != nil {
		r.unsub()
		r.unsub = nil
	}
	r.TtkWidget.Destroy()
}
