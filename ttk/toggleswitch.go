package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// toggleswitchW is the width of the trough in pixels (matches Tcl Toggleswitch1 rendered size ~57px).
const toggleswitchW = 57

// toggleswitchH is the height of the trough in pixels (matches Tcl Toggleswitch1 rendered size ~29px).
const toggleswitchH = 29

// Toggleswitch is a TTK sliding on/off toggle switch.
type Toggleswitch struct {
	TtkWidget

	Text     string
	Font     font.Font
	Command  func()
	Variable *widget.Variable[bool]

	selected bool
	unsub    func()
}

// ToggleswitchOption configures a Toggleswitch.
type ToggleswitchOption func(*Toggleswitch)

// ToggleswitchText sets the label text.
func ToggleswitchText(s string) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Text = s }
}

// ToggleswitchVar links the switch to a bool variable.
func ToggleswitchVar(v *widget.Variable[bool]) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Variable = v }
}

// ToggleswitchCommand sets the callback invoked on toggle.
func ToggleswitchCommand(fn func()) ToggleswitchOption {
	return func(t *Toggleswitch) { t.Command = fn }
}

// NewToggleswitch creates a TTK toggle switch widget.
func NewToggleswitch(parent widget.Caregiver, name string, opts ...ToggleswitchOption) *Toggleswitch {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	win.Class = "Toggleswitch"
	ts := &Toggleswitch{}
	ts.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&ts.TtkWidget, win, app, "TCheckbutton")
	ts.DisplayFunc = ts.Display

	for _, opt := range opts {
		opt(ts)
	}

	// Sync initial state from variable.
	if ts.Variable != nil {
		ts.selected = ts.Variable.Get()
		if ts.selected {
			ts.State |= StateSelected
		}
		ts.unsub = ts.Variable.OnChange(func(_, _ bool) {
			ts.selected = ts.Variable.Get()
			if ts.selected {
				ts.State |= StateSelected
			} else {
				ts.State &^= StateSelected
			}
			ts.Display()
		})
	}

	ts.computeSize()
	bindTtkHover(&ts.TtkWidget, app)
	bindToggleswitch(ts, app)
	return ts
}

func (ts *Toggleswitch) computeSize() {
	w := ts.Win
	textW, textH := 0, 0
	if ts.Font != nil && ts.Text != "" {
		textW = ts.Font.MeasureString(ts.Text)
		textH = ts.Font.Metrics().Linespace()
	}
	h := max(textH, toggleswitchH)
	gap := 0
	if ts.Text != "" {
		gap = 6
	}
	w.ReqWidth = toggleswitchW + gap + textW + 6
	w.ReqHeight = h + 6
}

// Display draws the toggle switch.
func (ts *Toggleswitch) Display() {
	if ts.Destroyed {
		return
	}
	win := ts.Win
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
	if ts.Context != nil && ts.Context.Style != nil {
		bgColor = LookupColor(ts.Context.Style, "-background", ts.State, bgColor)
	}

	// Window background.
	d.SetForeground(gc, bgColor)
	d.FillRectangle(win.Drawable(), gc, 0, 0, uint(width), uint(height))

	// Trough position (vertically centered).
	troughX := 3
	troughY := (height - toggleswitchH) / 2
	troughW := toggleswitchW
	troughH := toggleswitchH
	radius := troughH / 2

	// Trough color: dark navy when on, gray when off.
	var troughColor uint64
	if ts.selected {
		troughColor = 0x4a6984
		if ts.State&StateDisabled != 0 {
			troughColor = 0x90a0b0
		}
	} else {
		troughColor = 0xaaaaaa
		if ts.State&StateDisabled != 0 {
			troughColor = 0xcccccc
		}
	}

	// Draw rounded trough using filled rectangles + circles.
	d.SetForeground(gc, troughColor)
	// Center rectangle.
	d.FillRectangle(win.Drawable(), gc, troughX+radius, troughY, uint(troughW-2*radius), uint(troughH))
	// Left circle.
	d.FillArc(win.Drawable(), gc, troughX, troughY, uint(troughH), uint(troughH), 0, 360*64)
	// Right circle.
	d.FillArc(win.Drawable(), gc, troughX+troughW-troughH, troughY, uint(troughH), uint(troughH), 0, 360*64)

	// Thumb (white circle).
	thumbDiam := troughH - 4
	thumbY := troughY + 2
	var thumbX int
	if ts.selected {
		thumbX = troughX + troughW - thumbDiam - 2
	} else {
		thumbX = troughX + 2
	}

	d.SetForeground(gc, uint64(0xffffff))
	d.FillArc(win.Drawable(), gc, thumbX, thumbY, uint(thumbDiam), uint(thumbDiam), 0, 360*64)

	// Text label.
	if ts.Font != nil && ts.Text != "" {
		if df, ok := ts.Font.(platform.DrawableFont); ok {
			fgColor := uint64(0x000000)
			if ts.Context != nil && ts.Context.Style != nil {
				fgColor = LookupColor(ts.Context.Style, "-foreground", ts.State, fgColor)
			}
			m := ts.Font.Metrics()
			textX := troughX + troughW + 6
			textY := (height-m.Linespace())/2 + m.Ascent
			r := uint16((fgColor>>16)&0xFF) << 8
			g := uint16((fgColor>>8)&0xFF) << 8
			b := uint16((fgColor)&0xFF) << 8
			df.DrawString(win.Drawable(), textX, textY, ts.Text, fgColor, r, g, b)
		}
	}

	// Focus ring.
	if ts.State&StateFocus != 0 {
		d.SetForeground(gc, uint64(0x000000))
		d.DrawRectangle(win.Drawable(), gc, 0, 0, uint(width-1), uint(height-1))
	}

	d.Flush()
}

// Toggle flips the switch state.
func (ts *Toggleswitch) Toggle() {
	if ts.State&StateDisabled != 0 {
		return
	}
	ts.selected = !ts.selected
	if ts.selected {
		ts.State |= StateSelected
	} else {
		ts.State &^= StateSelected
	}
	if ts.Variable != nil {
		ts.Variable.Set(ts.selected)
	}
	ts.Display()
	if ts.Command != nil {
		ts.Command()
	}
}

// Get returns the current on/off state.
func (ts *Toggleswitch) Get() bool {
	return ts.selected
}

func bindToggleswitch(ts *Toggleswitch, app widget.AppContext) {
	win := ts.Win

	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		ts.Display()
	})

	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			ts.Display()
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			ts.ChangeState(StatePressed, 0)
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			wasPressed := ts.State&StatePressed != 0
			ts.ChangeState(0, StatePressed)
			if wasPressed && ev.X >= 0 && ev.X < win.Width && ev.Y >= 0 && ev.Y < win.Height {
				ts.Toggle()
			}
		}
	})
}

// Destroy cleans up the toggleswitch, unsubscribing from any linked Variable.
func (ts *Toggleswitch) Destroy() {
	if ts.unsub != nil {
		ts.unsub()
		ts.unsub = nil
	}
	ts.TtkWidget.Destroy()
}
