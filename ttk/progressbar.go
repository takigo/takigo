package ttk

import (
	"time"

	"github.com/msorc/takigo/internal/xlib"
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
func NewProgressbar(parent *window.Window, name string, app widget.AppContext, opts ...ProgressbarOption) *Progressbar {
	win := window.NewChildWindow(parent, name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	p := &Progressbar{
		Maximum:  100,
		Orient:   Horizontal,
		Length:   200,
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

	// Override requested size.
	if p.Orient == Horizontal {
		win.ReqWidth = p.Length
		win.ReqHeight = 20
	} else {
		win.ReqWidth = 20
		win.ReqHeight = p.Length
	}

	return p
}

// Display draws the progressbar.
func (p *Progressbar) Display() {
	if p.Destroyed {
		return
	}
	win := p.Win
	if win.XWindow == xlib.Window(0) {
		return
	}

	d := win.Display.XDisplay
	gc := win.GC
	width := win.Width
	height := win.Height

	if width <= 0 || height <= 0 {
		return
	}

	// Allocate or resize pixmap.
	if p.pixmap == xlib.Pixmap(0) || p.pixmapW != width || p.pixmapH != height {
		if p.pixmap != xlib.Pixmap(0) {
			d.FreePixmap(p.pixmap)
		}
		p.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		p.pixmapW = width
		p.pixmapH = height
	}

	pixDrawable := xlib.PixmapDrawable(p.pixmap)

	// Background.
	bg := LookupColor(p.Context.Style, "-background", p.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Trough (sunken area).
	troughColor := LookupColor(p.Context.Style, "-troughcolor", p.State, 0xc3c3c3)
	borderW := 1
	troughX := borderW
	troughY := borderW
	troughW := width - 2*borderW
	troughH := height - 2*borderW

	d.SetForeground(gc, troughColor)
	d.FillRectangle(pixDrawable, gc, troughX, troughY, uint(troughW), uint(troughH))

	// Draw trough border (sunken).
	d.SetForeground(gc, uint64(0x9e9a91)) // dark
	d.DrawLine(pixDrawable, gc, 0, 0, width-1, 0)
	d.DrawLine(pixDrawable, gc, 0, 0, 0, height-1)
	d.SetForeground(gc, uint64(0xffffff)) // light
	d.DrawLine(pixDrawable, gc, 0, height-1, width-1, height-1)
	d.DrawLine(pixDrawable, gc, width-1, 0, width-1, height-1)

	// Progress bar.
	barColor := LookupColor(p.Context.Style, "-barcolor", p.State, 0x4a6984)

	if p.Mode == ProgressDeterminate {
		if p.Maximum > 0 && p.Value > 0 {
			frac := p.Value / p.Maximum
			if frac > 1 {
				frac = 1
			}
			if p.Orient == Horizontal {
				barW := int(frac * float64(troughW))
				if barW > 0 {
					d.SetForeground(gc, barColor)
					d.FillRectangle(pixDrawable, gc, troughX, troughY, uint(barW), uint(troughH))
				}
			} else {
				barH := int(frac * float64(troughH))
				if barH > 0 {
					d.SetForeground(gc, barColor)
					d.FillRectangle(pixDrawable, gc, troughX, troughY+troughH-barH, uint(troughW), uint(barH))
				}
			}
		}
	} else {
		// Indeterminate: bouncing bar.
		barLen := troughW / 5
		if barLen < 20 {
			barLen = 20
		}
		maxPhase := troughW - barLen
		if maxPhase < 1 {
			maxPhase = 1
		}
		pos := p.phase % (maxPhase + 1)
		if p.Orient == Horizontal {
			d.SetForeground(gc, barColor)
			d.FillRectangle(pixDrawable, gc, troughX+pos, troughY, uint(barLen), uint(troughH))
		} else {
			d.SetForeground(gc, barColor)
			d.FillRectangle(pixDrawable, gc, troughX, troughY+pos, uint(troughW), uint(barLen))
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
		// Advance phase.
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
