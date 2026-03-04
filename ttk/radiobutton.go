package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

const ttkRadioIndicatorSize = 13

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
			r.Display()
		})
	}
}

// RadiobuttonCommand sets the callback invoked on selection.
func RadiobuttonCommand(fn func()) RadiobuttonOption {
	return func(r *Radiobutton) { r.Command = fn }
}

// NewRadiobutton creates a TTK themed radiobutton.
func NewRadiobutton(parent widget.Caregiver, name string, opts ...RadiobuttonOption) *Radiobutton {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	r := &Radiobutton{}
	r.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)
	r.Variable = widget.NewVariable("")

	InitTtkWidget(&r.TtkWidget, win, app, "TRadiobutton")
	// No layout registered for TRadiobutton → TtkWidget.Display() is a no-op.

	for _, opt := range opts {
		opt(r)
	}

	r.computeSize()
	bindTtkRadiobutton(r, app)
	return r
}

func (r *Radiobutton) computeSize() {
	w := r.Win
	indW := ttkRadioIndicatorSize + 4
	textW, textH := 0, 0
	if r.Font != nil && r.Text != "" {
		textW = r.Font.MeasureString(r.Text)
		textH = r.Font.Metrics().Linespace()
	}
	h := max(textH, ttkRadioIndicatorSize)
	w.ReqWidth = indW + textW + 8
	w.ReqHeight = h + 6
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

	border := draw.NewBorderFromPixel(bgColor)

	selected := r.Selected()

	// Indicator (circle).
	indX := 3
	indSize := ttkRadioIndicatorSize
	indY := (height - indSize) / 2

	// White fill for indicator circle.
	d.SetForeground(gc, uint64(0xffffff))
	d.FillArc(win.Drawable(), gc, indX, indY, uint(indSize), uint(indSize), 0, 360*64)

	// Border arcs around circle.
	d.SetForeground(gc, border.DarkPixel)
	d.DrawArc(win.Drawable(), gc, indX, indY, uint(indSize-1), uint(indSize-1), 45*64, 180*64)
	d.SetForeground(gc, border.LightPixel)
	d.DrawArc(win.Drawable(), gc, indX, indY, uint(indSize-1), uint(indSize-1), 225*64, 180*64)

	// Inner dot when selected.
	if selected {
		dotSize := indSize - 6
		dotX := indX + 3
		dotY := indY + 3
		d.SetForeground(gc, fgColor)
		d.FillArc(win.Drawable(), gc, dotX, dotY, uint(dotSize), uint(dotSize), 0, 360*64)
	}

	// Text label.
	if r.Font != nil && r.Text != "" {
		if df, ok := r.Font.(platform.DrawableFont); ok {
			m := r.Font.Metrics()
			textX := indX + indSize + 4
			textY := (height-m.Linespace())/2 + m.Ascent
			rv := uint16((fgColor>>16)&0xFF) << 8
			gv := uint16((fgColor>>8)&0xFF) << 8
			bv := uint16((fgColor)&0xFF) << 8
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
		d.DrawRectangle(win.Drawable(), gc, 0, 0, uint(width-1), uint(height-1))
	}

	d.Flush()
}

// Select selects this radiobutton (sets the variable to this button's value).
func (r *Radiobutton) Select() {
	if r.State&StateDisabled != 0 {
		return
	}
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

	// Expose — override bindTtkCommon.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		r.Display()
	})

	// Configure — override bindTtkCommon.
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			r.Display()
		}
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

