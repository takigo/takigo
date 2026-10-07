package ttk

import (
	"math"
	"time"

	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// ProgressMode is the mode of a progressbar.
type ProgressMode int

const (
	// ProgressDeterminate shows a bar proportional to value/maximum.
	ProgressDeterminate ProgressMode = iota
	// ProgressIndeterminate shows a bouncing bar.
	ProgressIndeterminate
)

// Progressbar is a themed progress indicator widget.
type Progressbar struct {
	TtkWidget
	Mode    ProgressMode
	Value   float64
	Maximum float64
	Orient  Orientation
	Length  int // requested length in pixels

	animating  bool
	cancelTick func() bool // cancels the pending tick, like progress.tcl's Timers
}

// ProgressbarOption configures a Progressbar.
type ProgressbarOption func(*Progressbar)

// ProgressbarMode sets the mode.
func ProgressbarMode(m ProgressMode) ProgressbarOption {
	return func(p *Progressbar) { p.Mode = m }
}

// ProgressbarValue sets the initial value.
func ProgressbarValue(v float64) ProgressbarOption {
	return func(p *Progressbar) { p.Value = v }
}

// ProgressbarMaximum sets the maximum value.
func ProgressbarMaximum(v float64) ProgressbarOption {
	return func(p *Progressbar) { p.Maximum = v }
}

// ProgressbarOrient sets the orientation.
func ProgressbarOrient(o Orientation) ProgressbarOption {
	return func(p *Progressbar) { p.Orient = o }
}

// ProgressbarLength sets the requested length.
func ProgressbarLength(l int) ProgressbarOption {
	return func(p *Progressbar) { p.Length = l }
}

// NewProgressbar creates a themed progressbar widget.
func NewProgressbar(parent widget.Caregiver, name string, opts ...ProgressbarOption) *Progressbar {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	p := &Progressbar{
		Maximum: 100,
		Orient:  Horizontal,
		Length:  screenunit.Pt(75).Pixels(), // ttkProgress.c -length default
	}

	for _, opt := range opts {
		opt(p)
	}

	InitTtkWidget(&p.TtkWidget, win, app, p.orientStyle())
	p.reconfigure = func() { _ = p.Configure() }
	win.OnDestroy(p.Destroy)
	p.DisplayFunc = p.Display
	p.requestSize()

	return p
}

func (p *Progressbar) orientStyle() string {
	if p.Orient == Vertical {
		return "Vertical.TProgressbar"
	}
	return "Horizontal.TProgressbar"
}

// requestSize requests -length along the bar; across it the pbar
// -thickness inside the trough's 1px sunken border.
func (p *Progressbar) requestSize() {
	thick := LookupInt(p.Context.Style, "-thickness", p.State, screenunit.Pt(3).Pixels()) + 2*pbTroughBorder
	if p.Orient == Horizontal {
		p.Win.ReqWidth, p.Win.ReqHeight = p.Length, thick
	} else {
		p.Win.ReqWidth, p.Win.ReqHeight = thick, p.Length
	}
}

// Configure sets options after creation.
func (p *Progressbar) Configure(opts ...ProgressbarOption) error {
	return configure(&p.TtkWidget, p, opts, func() { p.StyleName = p.orientStyle() }, p.requestSize)
}

// pbTroughBorder is the default theme trough's sunken border width.
const pbTroughBorder = 1

// Display draws the progressbar.
func (p *Progressbar) Display() {
	if p.Destroyed {
		return
	}
	win := p.Win
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

	pixDrawable := p.backBuffer(width, height)
	if pixDrawable == 0 {
		return
	}

	// Background.
	bg := LookupColor(p.Context.Style, "-background", p.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Trough: filled with -troughcolor inside a sunken border of its shades.
	troughColor := LookupColor(p.Context.Style, "-troughcolor", p.State, 0xc3c3c3)
	bw := pbTroughBorder
	troughX, troughY := bw, bw
	troughW, troughH := width-2*bw, height-2*bw
	fill3DRectangle(d, pixDrawable, gc, draw.NewBorderFromPixel(troughColor),
		Box{0, 0, width, height}, bw, option.ReliefSunken)

	// Bar (pbar element) in -barcolor: the value fraction of the trough, or a
	// -barsize long block in indeterminate mode.
	barColor := LookupColor(p.Context.Style, "-barcolor", p.State, 0x4a6984)
	d.SetForeground(gc, barColor)
	length := troughW
	if p.Orient == Vertical {
		length = troughH
	}
	var start, size int
	if p.Mode == ProgressDeterminate {
		if p.Maximum > 0 && p.Value > 0 {
			size = int(min(p.Value/p.Maximum, 1) * float64(length))
		}
	} else {
		size = min(LookupInt(p.Context.Style, "-barsize", p.State, screenunit.Pt(22.5).Pixels()), length)
		// ProgressbarIndeterminateLayout: value/maximum bounces over 0..2.
		f := 0.0
		if p.Maximum != 0 {
			f = math.Mod(math.Abs(p.Value/p.Maximum), 2)
		}
		if f > 1 {
			f = 2 - f
		}
		start = int(f * float64(length-size))
	}
	if size > 0 {
		if p.Orient == Horizontal {
			d.FillRectangle(pixDrawable, gc, troughX+start, troughY, uint(size), uint(troughH))
		} else {
			d.FillRectangle(pixDrawable, gc, troughX, troughY+troughH-start-size, uint(troughW), uint(size))
		}
	}

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// SetValue sets the progress value and redisplays.
func (p *Progressbar) SetValue(v float64) {
	p.Value = v
	p.Display()
}

// Step ports ProgressbarStepCommand: add amount to the value, wrapping at
// -maximum in determinate mode.
func (p *Progressbar) Step(amount float64) {
	p.Value += amount
	if p.Mode == ProgressDeterminate && p.Maximum != 0 {
		p.Value = math.Mod(p.Value, p.Maximum)
	}
	p.Display()
}

// Start ports ttk::progressbar::start (library/ttk/progress.tcl): step by 1
// now and then every interval.
func (p *Progressbar) Start(interval time.Duration) {
	if p.animating {
		return
	}
	p.animating = true

	var tick func()
	tick = func() {
		if p.Destroyed {
			return
		}
		p.cancelTick = p.App.After(interval, tick)
		p.Step(1)
	}
	tick()
}

// Stop stops the animation.
func (p *Progressbar) Stop() {
	if !p.animating {
		return
	}
	p.animating = false
	p.cancelTick()
}

// Destroy frees resources.
func (p *Progressbar) Destroy() {
	if p.animating {
		p.Stop()
	}
	p.TtkWidget.Destroy()
}
