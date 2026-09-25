package ttk

import (
	"time"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
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

	// Animation state for indeterminate mode.
	phase     int
	phaseDir  int // +1 or -1
	animating bool
	stopChan  chan struct{}
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
		Maximum:  100,
		Orient:   Horizontal,
		Length:   screenunit.Px("75p"), // ttkProgress.c -length default
		phaseDir: 1,
	}

	for _, opt := range opts {
		opt(p)
	}

	styleName := "Horizontal.TProgressbar"
	if p.Orient == Vertical {
		styleName = "Vertical.TProgressbar"
	}

	InitTtkWidget(&p.TtkWidget, win, app, styleName)
	p.DisplayFunc = p.Display

	// -length along the bar; across it the pbar -thickness inside the
	// trough's 1px sunken border.
	thick := LookupInt(p.Context.Style, "-thickness", p.State, screenunit.Px("3p")) + 2*pbTroughBorder
	if p.Orient == Horizontal {
		win.ReqWidth, win.ReqHeight = p.Length, thick
	} else {
		win.ReqWidth, win.ReqHeight = thick, p.Length
	}

	return p
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

	// Allocate or resize pixmap.
	if p.pixmap == 0 || p.pixmapW != width || p.pixmapH != height {
		if p.pixmap != 0 {
			d.FreePixmap(p.pixmap)
		}
		p.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		p.pixmapW = width
		p.pixmapH = height
	}
	if p.pixmap == 0 {
		return
	}

	pixDrawable := platform.PixmapDrawable(p.pixmap)

	// Background.
	bg := LookupColor(p.Context.Style, "-background", p.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Trough: filled with -troughcolor inside a sunken border of its shades.
	troughColor := LookupColor(p.Context.Style, "-troughcolor", p.State, 0xc3c3c3)
	bw := pbTroughBorder
	troughX, troughY := bw, bw
	troughW, troughH := width-2*bw, height-2*bw
	d.SetForeground(gc, troughColor)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))
	draw.Draw3DRectangle(d, pixDrawable, gc, draw.NewBorderFromPixel(troughColor),
		0, 0, width, height, bw, option.ReliefSunken)

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
		size = min(LookupInt(p.Context.Style, "-barsize", p.State, screenunit.Px("22.5p")), length)
		if maxPhase := length - size; maxPhase > 0 {
			start = p.phase % (maxPhase + 1)
		}
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

// Start begins the animation for indeterminate mode.
func (p *Progressbar) Start(interval time.Duration) {
	if p.animating {
		return
	}
	p.animating = true
	p.stopChan = make(chan struct{})

	var tick func()
	tick = func() {
		select {
		case <-p.stopChan:
			return
		default:
		}
		if !p.animating || p.Destroyed {
			return
		}
		if p.Mode == ProgressDeterminate {
			// Advance value, wrapping at maximum.
			p.Value += 1
			if p.Value >= p.Maximum {
				p.Value = 0
			}
		} else {
			// Advance phase for bouncing bar.
			troughW := p.Win.Width - 2
			barLen := troughW / 5
			if barLen < 20 {
				barLen = 20
			}
			maxPhase := troughW - barLen
			if maxPhase < 1 {
				maxPhase = 1
			}

			p.phase += p.phaseDir * 3
			if p.phase >= maxPhase {
				p.phase = maxPhase
				p.phaseDir = -1
			} else if p.phase <= 0 {
				p.phase = 0
				p.phaseDir = 1
			}
		}
		p.Display()
		p.App.After(interval, tick)
	}
	p.App.After(interval, tick)
}

// Stop stops the animation.
func (p *Progressbar) Stop() {
	if !p.animating {
		return
	}
	p.animating = false
	close(p.stopChan)
}

// Destroy frees resources.
func (p *Progressbar) Destroy() {
	if p.animating {
		p.Stop()
	}
	p.TtkWidget.Destroy()
}
