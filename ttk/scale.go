package ttk

import (
	"time"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Scale is ttk::scale (tk/generic/ttk/ttkScale.c with the default theme's
// trough and slider elements from ttkElements.c, bindings from
// library/ttk/scale.tcl).
type Scale struct {
	TtkWidget

	Orient   Orientation
	From, To float64
	Value    float64
	Length   int
	Command  func(float64)
	Variable *widget.Variable[float64]

	dragging  bool
	overThumb bool
	repeatGen int
	unsub     func()
}

// ScaleOption configures a Scale.
type ScaleOption func(*Scale)

// ScaleOrient sets -orient.
func ScaleOrient(o Orientation) ScaleOption { return func(s *Scale) { s.Orient = o } }

// ScaleFrom sets -from.
func ScaleFrom(v float64) ScaleOption { return func(s *Scale) { s.From = v } }

// ScaleTo sets -to.
func ScaleTo(v float64) ScaleOption { return func(s *Scale) { s.To = v } }

// ScaleValue sets -value.
func ScaleValue(v float64) ScaleOption { return func(s *Scale) { s.Value = v } }

// ScaleLength sets -length (a Tk distance).
func ScaleLength(v any) ScaleOption { return func(s *Scale) { s.Length = screenunit.Px(v) } }

// ScaleCommand sets -command; it receives the new value.
func ScaleCommand(fn func(float64)) ScaleOption { return func(s *Scale) { s.Command = fn } }

// ScaleVariable links -variable.
func ScaleVariable(v *widget.Variable[float64]) ScaleOption {
	return func(s *Scale) { s.Variable = v }
}

// NewScale creates a themed scale.
func NewScale(parent widget.Caregiver, name string, opts ...ScaleOption) *Scale {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)
	win.Flags |= window.FlagFocusable

	s := &Scale{To: 1, Length: 100}
	s.DisplayFunc = s.Display
	for _, opt := range opts {
		opt(s)
	}
	style := "Horizontal.TScale"
	if s.Orient == Vertical {
		style = "Vertical.TScale"
	}
	InitTtkWidget(&s.TtkWidget, win, app, style)
	win.Class = "TScale"
	if s.Variable != nil {
		s.Value = s.Variable.Get()
		s.unsub = s.Variable.OnChange(func(_, v float64) {
			s.Value = v
			s.Display()
		})
	}
	s.requestSize()
	bindScale(s, app)
	return s
}

func (s *Scale) sliderDim() int { return 16 * screenunit.ScalingPct() / 100 }

// inset is the focus element's -focusthickness plus -padding.
func (s *Scale) inset() Padding {
	ft := 1
	var pad Padding
	if s.Context != nil {
		ft = LookupInt(s.Context.Style, "-focusthickness", 0, 1)
		pad = LookupPadding(s.Context.Style, "-padding", 0, Padding{})
	}
	return Padding{Left: pad.Left + ft, Top: pad.Top + ft, Right: pad.Right + ft, Bottom: pad.Bottom + ft}
}

// requestSize ports ScaleSize: the layout size, stretched to -length along
// the long axis.
func (s *Scale) requestSize() {
	in := s.inset()
	dim := s.sliderDim()
	w, h := dim+in.Left+in.Right, dim+in.Top+in.Bottom
	if s.Orient == Vertical {
		h = max(h, s.Length)
	} else {
		w = max(w, s.Length)
	}
	s.Win.ReqWidth, s.Win.ReqHeight = w, h
}

func (s *Scale) troughBox() Box {
	in := s.inset()
	return Box{in.Left, in.Top, s.Win.Width - in.Left - in.Right, s.Win.Height - in.Top - in.Bottom}
}

func (s *Scale) fraction(v float64) float64 {
	if s.From == s.To {
		return 1
	}
	f := (v - s.From) / (s.To - s.From)
	return max(0, min(1, f))
}

// sliderBox ports ScaleDoLayout: the slider packed at the trough's start,
// centred across it, and shifted by fraction * range.
func (s *Scale) sliderBox() Box {
	t := s.troughBox()
	dim := s.sliderDim()
	f := s.fraction(s.Value)
	if s.Orient == Vertical {
		return Box{t.X + (t.Width-dim)/2, t.Y + int(f*float64(t.Height-dim)), dim, dim}
	}
	return Box{t.X + int(f*float64(t.Width-dim)), t.Y + (t.Height-dim)/2, dim, dim}
}

// PointToValue ports PointToValue.
func (s *Scale) PointToValue(x, y int) float64 {
	t := s.troughBox()
	dim := s.sliderDim()
	var f float64
	if s.Orient == Vertical {
		if t.Height-dim <= 0 {
			return s.Value
		}
		f = float64(y-(t.Y+dim/2)) / float64(t.Height-dim)
	} else {
		if t.Width-dim <= 0 {
			return s.Value
		}
		f = float64(x-(t.X+dim/2)) / float64(t.Width-dim)
	}
	f = max(0, min(1, f))
	return s.From + f*(s.To-s.From)
}

// Set ports "$scale set": clamp, redisplay, update -variable, run -command.
func (s *Scale) Set(v float64) {
	if s.State&StateDisabled != 0 {
		return
	}
	lo, hi := s.From, s.To
	if lo > hi {
		lo, hi = hi, lo
	}
	s.Value = max(lo, min(hi, v))
	s.Display()
	if s.Variable != nil {
		s.Variable.Set(s.Value)
	}
	if s.Command != nil {
		s.Command(s.Value)
	}
}

// Get returns the current value.
func (s *Scale) Get() float64 { return s.Value }

func (s *Scale) increment(delta float64) {
	if s.From > s.To {
		delta = -delta
	}
	s.Set(s.Value + delta)
}

// Display draws the trough groove (TroughElementDraw with -groovewidth),
// the inner-colour fill up to the slider centre and the slider image.
func (s *Scale) Display() {
	win := s.Win
	if s.Destroyed || win.PlatformID == 0 || s.Context == nil {
		return
	}
	width, height := win.Width, win.Height
	if width <= 0 || height <= 0 {
		return
	}
	d := win.Display.Server
	gc := win.GC
	if s.pixmap == 0 || s.pixmapW != width || s.pixmapH != height {
		if s.pixmap != 0 {
			d.FreePixmap(s.pixmap)
		}
		s.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		s.pixmapW, s.pixmapH = width, height
	}
	if s.pixmap == 0 {
		return
	}
	pix := platform.PixmapDrawable(s.pixmap)
	st := s.Context.Style
	bg := LookupColor(st, "-background", s.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pix, gc, 0, 0, uint(width), uint(height))

	t := s.troughBox()
	groove := LookupInt(st, "-groovewidth", s.State, screenunit.Px("3p"))
	bw := LookupInt(st, "-troughborderwidth", s.State, 1)
	g := t
	if groove > 0 && groove < t.Height && groove < t.Width {
		if s.Orient == Vertical {
			g.X += (t.Width - groove) / 2
			g.Width = groove
		} else {
			g.Y += (t.Height - groove) / 2
			g.Height = groove
		}
	}
	trough := draw.NewBorderFromPixel(LookupColor(st, "-troughcolor", s.State, 0xc3c3c3))
	fill3DRectangle(d, pix, gc, trough, g, bw, option.ReliefSunken)

	sb := s.sliderBox()
	dim := s.sliderDim()
	inner := LookupColor(st, "-innercolor", s.State, 0x4a6984)
	st2 := s.State
	if s.overThumb {
		st2 |= StateHover
	}
	outer := LookupColor(st, "-outercolor", st2, 0xffffff)
	border := LookupColor(st, "-bordercolor", s.State, 0xc3c3c3)
	d.SetForeground(gc, inner)
	if s.Orient == Vertical {
		if h := sb.Y + dim/2 - (g.Y + bw); h > 0 {
			d.FillRectangle(pix, gc, g.X+bw, g.Y+bw, uint(g.Width-2*bw), uint(h))
		}
	} else if w := sb.X + dim/2 - (g.X + bw); w > 0 {
		d.FillRectangle(pix, gc, g.X+bw, g.Y+bw, uint(w), uint(g.Height-2*bw))
	}
	if sb.X >= 0 && sb.Y >= 0 && sb.X+dim <= width && sb.Y+dim <= height {
		draw.DrawTtkSlider(d, pix, gc, win.Depth, sb.X, sb.Y, dim, inner, outer, border, bg)
	}

	if s.State&StateFocus != 0 {
		d.SetForeground(gc, LookupColor(st, "-focuscolor", s.State, 0x000000))
		drawDottedRect(d, pix, gc, 0, 0, width, height)
	}
	d.CopyArea(pix, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// fill3DRectangle ports ttkElements.c's Fill3DRectangle, whose 1-pixel
// raised/sunken case gives the full south and east edges to the second
// shade.
func fill3DRectangle(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	border *draw.Border, b Box, bw int, relief option.Relief) {
	if bw == 1 && b.Width >= 2 && b.Height >= 2 &&
		(relief == option.ReliefRaised || relief == option.ReliefSunken) {
		x1, x2, y1, y2 := b.X, b.X+b.Width-1, b.Y, b.Y+b.Height-1
		d.SetForeground(gc, border.BgPixel)
		d.FillRectangle(drawable, gc, b.X+1, b.Y+1, uint(b.Width-2), uint(b.Height-2))
		nw, se := border.DarkPixel, border.LightPixel
		if relief == option.ReliefRaised {
			nw, se = se, nw
		}
		d.SetForeground(gc, nw)
		d.DrawLine(drawable, gc, x1, y1, x2-1, y1)
		d.DrawLine(drawable, gc, x1, y1, x1, y2-1)
		d.SetForeground(gc, se)
		d.DrawLine(drawable, gc, x1, y2, x2, y2)
		d.DrawLine(drawable, gc, x2, y1, x2, y2)
		return
	}
	draw.Fill3DRectangle(d, drawable, gc, border, b.X, b.Y, b.Width, b.Height, bw, relief)
}

// drawDottedRect approximates TtkDrawFocusRing's dotted rectangle.
func drawDottedRect(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID, x, y, w, h int) {
	for i := 0; i < w; i += 2 {
		d.DrawLine(drawable, gc, x+i, y, x+i, y)
		d.DrawLine(drawable, gc, x+i, y+h-1, x+i, y+h-1)
	}
	for i := 0; i < h; i += 2 {
		d.DrawLine(drawable, gc, x, y+i, x, y+i)
		d.DrawLine(drawable, gc, x+w-1, y+i, x+w-1, y+i)
	}
}

func (s *Scale) inSlider(x, y int) bool {
	b := s.sliderBox()
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// repeat ports ttk::Repeatedly: run fn now, again after 300 ms and then
// every 100 ms until the button is released.
func (s *Scale) repeat(app widget.AppContext, fn func()) {
	s.repeatGen++
	gen := s.repeatGen
	fn()
	var tick func()
	tick = func() {
		if gen != s.repeatGen || s.Destroyed {
			return
		}
		fn()
		app.After(100*time.Millisecond, tick)
	}
	app.After(300*time.Millisecond, tick)
}

func bindScale(s *Scale, app widget.AppContext) {
	win := s.Win
	disp := app.Dispatcher()
	disp.Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Type != event.ButtonPressType {
			return
		}
		app.Server().SetInputFocus(win.PlatformID, platform.RevertToParent, platform.CurrentTime)
		s.dragging = false
		switch ev.Button {
		case 1:
			if s.inSlider(ev.X, ev.Y) {
				s.dragging = true
				return
			}
			inc := 1.0
			if (s.PointToValue(ev.X, ev.Y) <= s.Value) != (s.From > s.To) {
				inc = -1
			}
			s.repeat(app, func() { s.increment(inc) })
		case 2, 3:
			s.Set(s.PointToValue(ev.X, ev.Y))
			s.dragging = true
		}
	})
	disp.Bind(win.PlatformID, event.MotionMask|event.LeaveMask, func(ev *event.Event) {
		over := ev.Type == event.MotionType && s.inSlider(ev.X, ev.Y)
		if over != s.overThumb {
			s.overThumb = over
			s.Display()
		}
		if ev.Type == event.MotionType && s.dragging {
			s.Set(s.PointToValue(ev.X, ev.Y))
		}
	})
	disp.Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Type == event.ButtonReleaseType {
			s.dragging = false
			s.repeatGen++
		}
	})
	disp.Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.Type != event.KeyPressType {
			return
		}
		switch ev.KeySym {
		case platform.XK_Left, platform.XK_Up:
			s.increment(-1)
		case platform.XK_Right, platform.XK_Down:
			s.increment(1)
		case platform.XK_Home:
			s.Set(s.From)
		case platform.XK_End:
			s.Set(s.To)
		}
	})
}

// Destroy unlinks the variable and destroys the window.
func (s *Scale) Destroy() {
	if s.unsub != nil {
		s.unsub()
		s.unsub = nil
	}
	s.TtkWidget.Destroy()
}
