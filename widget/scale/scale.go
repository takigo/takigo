// Package scale implements a scale (slider) widget.
// It ports tk/generic/tkScale.c.
package scale

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Orient specifies the scale orientation.
type Orient = option.Orient

const (
	Horizontal = option.Horizontal
	Vertical   = option.Vertical
)

// Scale is a slider widget that lets the user select a numeric value.
type Scale struct {
	widget.Base

	Orient       Orient
	Value        float64
	From         float64
	To           float64
	Resolution   float64
	ShowValue    bool
	Label        string
	SliderLength int
	Width        int // trough cross-axis width
	TickInterval float64
	// Length is -length: the long-axis size in pixels (default 100).
	Length  int
	Command func(float64)

	// Interaction state.
	dragging   bool
	dragOffset int
	active     bool // pointer over the slider (Tk's "active" state)
	focused    bool

	// Colors.
	TroughColor *color.ColorRef
}

// ScaleOption configures a Scale.
type ScaleOption func(*Scale)

func OrientOpt(o Orient) ScaleOption          { return func(s *Scale) { s.Orient = o } }
func FromOpt(v float64) ScaleOption           { return func(s *Scale) { s.From = v } }
func ToOpt(v float64) ScaleOption             { return func(s *Scale) { s.To = v } }
func ValueOpt(v float64) ScaleOption          { return func(s *Scale) { s.Value = v } }
func ResolutionOpt(v float64) ScaleOption     { return func(s *Scale) { s.Resolution = v } }
func ShowValueOpt(b bool) ScaleOption         { return func(s *Scale) { s.ShowValue = b } }
func LabelOpt(s string) ScaleOption           { return func(sc *Scale) { sc.Label = s } }
func SliderLengthOpt(n int) ScaleOption       { return func(s *Scale) { s.SliderLength = n } }
func WidthOpt(w int) ScaleOption              { return func(s *Scale) { s.Width = w } }
func TickIntervalOpt(v float64) ScaleOption   { return func(s *Scale) { s.TickInterval = v } }
func LengthOpt(n int) ScaleOption             { return func(s *Scale) { s.Length = n } }
func CommandOpt(fn func(float64)) ScaleOption { return func(s *Scale) { s.Command = fn } }

func Background[C color.Spec](name C) ScaleOption {
	return func(s *Scale) { s.SetBackgroundColor(name) }
}

// New creates a new Scale widget.
func New(parent widget.Caregiver, name string, opts ...ScaleOption) *Scale {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	s := &Scale{
		Orient:       Vertical, // DEF_SCALE_ORIENT
		From:         0,
		To:           100,
		Resolution:   1,
		ShowValue:    true,
		SliderLength: 30,
		Width:        15,
		Length:       100,
	}
	widget.InitBase(&s.Base, w, app)
	s.SetDisplayProc(s.display)
	w.Class = "Scale"
	w.Flags |= window.FlagFocusable
	s.BorderWidth = 1
	s.HighlightWidth = 1
	s.Relief = option.ReliefFlat

	if tc, err := app.ColorCache().Get("#c3c3c3"); err == nil {
		s.TroughColor = tc.Ref()
	}

	for _, opt := range opts {
		opt(s)
	}

	s.computeGeometry()

	if s.Background != nil {
		w.SetBackgroundPixel(s.Background.Pixel)
	}

	bindScale(s, app)
	return s
}

// Set sets the scale value.
func (s *Scale) Set(value float64) {
	value = s.roundToResolution(value)
	value = s.clampValue(value)
	if value == s.Value {
		return
	}
	s.Value = value
	if s.Command != nil {
		s.Command(s.Value)
	}
	s.Display()
}

// Get returns the current value.
func (s *Scale) Get() float64 {
	return s.Value
}

// roundToResolution ports TkRoundValueToResolution.
func (s *Scale) roundToResolution(v float64) float64 {
	return s.roundInterval(v-s.From) + s.From
}

// roundInterval ports TkRoundIntervalToResolution.
func (s *Scale) roundInterval(d float64) float64 {
	r := s.Resolution
	if r <= 0 {
		return d
	}
	tick := math.Floor(d / r)
	rounded := r * tick
	if rem := d - rounded; rem < 0 {
		if rem <= -r/2 {
			rounded = (tick - 1) * r
		}
	} else if rem >= r/2 {
		rounded = (tick + 1) * r
	}
	return rounded
}

func (s *Scale) clampValue(v float64) float64 {
	lo, hi := s.From, s.To
	if lo > hi {
		lo, hi = hi, lo
	}
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// spacing is SPACING in tkScale.h.
const spacing = 2

func (s *Scale) inset() int { return s.HighlightWidth + s.BorderWidth }

func (s *Scale) fontHeight() int {
	if s.Font == nil {
		return 0
	}
	return s.Font.Metrics().Linespace() + spacing
}

// pixelRange is the travel of the slider centre (TkScaleValueToPixel).
func (s *Scale) pixelRange() int {
	size := s.Win.Width
	if s.Orient == Vertical {
		size = s.Win.Height
	}
	return size - s.SliderLength - 2*s.inset() - 2*s.BorderWidth
}

// valueToPixel ports TkScaleValueToPixel: the slider centre for value.
func (s *Scale) valueToPixel(value float64) int {
	pr := s.pixelRange()
	y := 0
	if vr := s.To - s.From; vr != 0 {
		y = int(math.Floor((value-s.From)*float64(pr)/vr + 0.5))
		y = max(0, min(pr, y))
	}
	return y + s.SliderLength/2 + s.inset() + s.BorderWidth
}

// pixelToValue ports TkScalePixelToValue.
func (s *Scale) pixelToValue(pixel int) float64 {
	pr := s.pixelRange()
	if pr <= 0 {
		return s.Value
	}
	f := float64(pixel-(s.SliderLength/2+s.inset()+s.BorderWidth)) / float64(pr)
	f = max(0, min(1, f))
	return s.roundToResolution(s.From + f*(s.To-s.From))
}

// hitTest ports TkpScaleElement: "slider", "trough1", "trough2" or "".
func (s *Scale) hitTest(x, y int) string {
	g := s.layout()
	pos := x
	if s.Orient == Vertical {
		if x < g.troughX || x >= g.troughX+2*s.BorderWidth+s.Width ||
			y < s.inset() || y >= s.Win.Height-s.inset() {
			return ""
		}
		pos = y
	} else if y < g.troughY || y >= g.troughY+2*s.BorderWidth+s.Width ||
		x < s.inset() || x >= s.Win.Width-s.inset() {
		return ""
	}
	first := s.valueToPixel(s.Value) - s.SliderLength/2
	switch {
	case pos < first:
		return "trough1"
	case pos < first+s.SliderLength:
		return "slider"
	}
	return "trough2"
}

// scaleLayout holds ComputeScaleGeometry's positions.
type scaleLayout struct {
	labelY, valueY, troughY, tickY           int // horizontal
	tickRightX, valueRightX, troughX, labelX int // vertical
}

// layout ports ComputeScaleGeometry; it also returns the requested size.
func (s *Scale) layout() scaleLayout {
	g, _, _ := s.geometry()
	return g
}

func (s *Scale) geometry() (g scaleLayout, reqW, reqH int) {
	inset := s.inset()
	fh := s.fontHeight()
	if s.Orient == Horizontal {
		y, extra := inset, 0
		if s.Label != "" {
			g.labelY = y + spacing
			y += fh
			extra = spacing
		}
		if s.ShowValue {
			g.valueY = y + spacing
			y += fh
			extra = spacing
		} else {
			g.valueY = y
		}
		y += extra
		g.troughY = y
		y += s.Width + 2*s.BorderWidth
		if s.TickInterval != 0 {
			g.tickY = y + spacing
			y += fh + spacing
		}
		return g, s.Length + 2*inset, y + inset
	}
	valuePixels, tickPixels, ascent := 0, 0, 0
	if s.Font != nil {
		valuePixels = max(s.Font.MeasureString(s.formatValue(s.From)), s.Font.MeasureString(s.formatValue(s.To)))
		tickPixels = max(s.Font.MeasureString(s.formatTick(s.From)), s.Font.MeasureString(s.formatTick(s.To)))
		ascent = s.Font.Metrics().Ascent
	}
	x := inset
	switch {
	case s.TickInterval != 0 && s.ShowValue:
		g.tickRightX = x + spacing + tickPixels
		g.valueRightX = g.tickRightX + valuePixels + ascent/2
		x = g.valueRightX + spacing
	case s.TickInterval != 0:
		g.tickRightX = x + spacing + tickPixels
		g.valueRightX = g.tickRightX
		x = g.tickRightX + spacing
	case s.ShowValue:
		g.tickRightX = x
		g.valueRightX = x + spacing + valuePixels
		x = g.valueRightX + spacing
	default:
		g.tickRightX, g.valueRightX = x, x
	}
	g.troughX = x
	x += 2*s.BorderWidth + s.Width
	if s.Label != "" && s.Font != nil {
		g.labelX = x + ascent/2
		x = g.labelX + ascent/2 + s.Font.MeasureString(s.Label)
	}
	return g, x + inset, s.Length + 2*inset
}

func (s *Scale) computeGeometry() {
	_, w, h := s.geometry()
	s.Win.ReqWidth, s.Win.ReqHeight = w, h
	s.Win.InternalBorderLeft = s.inset()
	s.Win.InternalBorderRight = s.inset()
	s.Win.InternalBorderTop = s.inset()
	s.Win.InternalBorderBottom = s.inset()
}

// formatValue and formatTick stand in for Tk's ComputeFormat digits.
func (s *Scale) formatValue(v float64) string {
	if s.Resolution >= 1 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

func (s *Scale) formatTick(v float64) string { return s.formatValue(v) }

// ticks returns the tick values, thinned so the labels do not overlap as
// in DisplayHorizontalScale/DisplayVerticalScale.
func (s *Scale) ticks() []float64 {
	// ConfigureScale rounds -tickinterval to the resolution and points it
	// from -from towards -to; an interval below half the resolution rounds
	// to zero and draws no ticks.
	interval := s.roundInterval(s.TickInterval)
	if interval == 0 || s.Font == nil {
		return nil
	}
	if (interval < 0) != (s.To-s.From < 0) {
		interval = -interval
	}
	n := math.Abs((s.To - s.From) / interval)
	var maxTicks float64
	if s.Orient == Horizontal {
		maxTicks = float64(s.Win.Width) / float64(max(1, s.Font.MeasureString(s.formatTick(s.From))))
	} else {
		maxTicks = float64(s.Win.Height) / float64(max(1, s.fontHeight()))
	}
	if n > maxTicks {
		interval *= n / maxTicks
	}
	var out []float64
	for v := s.From; ; v += interval {
		v = s.roundToResolution(v)
		if (s.To >= s.From && v > s.To) || (s.To < s.From && v < s.To) {
			break
		}
		out = append(out, v)
	}
	return out
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (s *Scale) Display() {
	s.EventuallyRedraw()
}

// display ports TkpDisplayScale (tk/unix/tkUnixScale.c).
func (s *Scale) display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) || w.Width <= 0 || w.Height <= 0 {
		return
	}
	d := w.Display.Server
	gc := w.GC
	pix := w.Drawable()

	bg := uint64(0xd9d9d9)
	if s.Background != nil {
		bg = s.Background.Pixel
	}
	bgBorder := draw.NewBorderFromPixel(bg)
	d.SetForeground(gc, bg)
	d.FillRectangle(pix, gc, 0, 0, uint(w.Width), uint(w.Height))

	g := s.layout()
	fg := uint64(0)
	if s.Foreground != nil {
		fg = s.Foreground.Pixel
	}
	var m fontMetrics
	df, _ := s.Font.(platform.DrawableFont)
	if s.Font != nil {
		fm := s.Font.Metrics()
		m = fontMetrics{fm.Ascent, fm.Descent}
	}
	drawText := func(x, y int, str string) {
		if df != nil {
			r, gg, b := uint16(fg>>16&0xff)<<8, uint16(fg>>8&0xff)<<8, uint16(fg&0xff)<<8
			df.DrawString(pix, x, y, str, fg, r, gg, b)
		}
	}
	inset := s.inset()
	bw := s.BorderWidth
	sliderBorder := bgBorder
	if s.active {
		sliderBorder = draw.NewBorderFromPixel(s.activeBg())
	}
	shadow := max(1, bw/2)
	trough := uint64(0xc3c3c3)
	if s.TroughColor != nil {
		trough = s.TroughColor.Pixel
	}

	if s.Orient == Horizontal {
		// DisplayHorizontalValue: centred on the value's pixel, kept inside.
		hval := func(v float64, top int, str string) {
			if s.Font == nil {
				return
			}
			width := s.Font.MeasureString(str)
			x := max(s.valueToPixel(v)-width/2, inset+spacing)
			if x+width >= w.Width-inset {
				x = w.Width - inset - spacing - width
			}
			drawText(x, top+m.ascent, str)
		}
		for _, t := range s.ticks() {
			hval(t, g.tickY, s.formatTick(t))
		}
		if s.ShowValue {
			hval(s.Value, g.valueY, s.formatValue(s.Value))
		}
		y := g.troughY
		draw.Draw3DRectangle(d, pix, gc, bgBorder, inset, y, w.Width-2*inset, s.Width+2*bw, bw, option.ReliefSunken)
		d.SetForeground(gc, trough)
		d.FillRectangle(pix, gc, inset+bw, y+bw, uint(w.Width-2*inset-2*bw), uint(s.Width))
		half := s.SliderLength / 2
		x := s.valueToPixel(s.Value) - half
		y += bw
		height := s.Width
		draw.Draw3DRectangle(d, pix, gc, sliderBorder, x, y, 2*half, height, shadow, option.ReliefRaised)
		x += shadow
		y += shadow
		half -= shadow
		height -= 2 * shadow
		draw.Fill3DRectangle(d, pix, gc, sliderBorder, x, y, half, height, shadow, option.ReliefRaised)
		draw.Fill3DRectangle(d, pix, gc, sliderBorder, x+half, y, half, height, shadow, option.ReliefRaised)
		if s.Label != "" {
			drawText(inset+m.ascent/2, g.labelY+m.ascent, s.Label)
		}
	} else {
		// DisplayVerticalValue: right-aligned, centred on the value's pixel.
		vval := func(v float64, right int, str string) {
			if s.Font == nil {
				return
			}
			y := s.valueToPixel(v) + m.ascent/2
			width := s.Font.MeasureString(str)
			if y-m.ascent < inset+spacing {
				y = inset + spacing + m.ascent
			}
			if y+m.descent > w.Height-inset-spacing {
				y = w.Height - inset - spacing - m.descent
			}
			drawText(right-width, y, str)
		}
		for _, t := range s.ticks() {
			vval(t, g.tickRightX, s.formatTick(t))
		}
		if s.ShowValue {
			vval(s.Value, g.valueRightX, s.formatValue(s.Value))
		}
		draw.Draw3DRectangle(d, pix, gc, bgBorder, g.troughX, inset, s.Width+2*bw, w.Height-2*inset, bw, option.ReliefSunken)
		d.SetForeground(gc, trough)
		d.FillRectangle(pix, gc, g.troughX+bw, inset+bw, uint(s.Width), uint(w.Height-2*inset-2*bw))
		half := s.SliderLength / 2
		width := s.Width
		x := g.troughX + bw
		y := s.valueToPixel(s.Value) - half
		draw.Draw3DRectangle(d, pix, gc, sliderBorder, x, y, width, 2*half, shadow, option.ReliefRaised)
		x += shadow
		y += shadow
		width -= 2 * shadow
		half -= shadow
		draw.Fill3DRectangle(d, pix, gc, sliderBorder, x, y, width, half, shadow, option.ReliefRaised)
		draw.Fill3DRectangle(d, pix, gc, sliderBorder, x, y+half, width, half, shadow, option.ReliefRaised)
		if s.Label != "" {
			drawText(g.labelX, inset+3*m.ascent/2, s.Label)
		}
	}

	hl := s.HighlightWidth
	if s.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, pix, gc, bgBorder, hl, hl, w.Width-2*hl, w.Height-2*hl, bw, s.Relief)
	}
	if hl > 0 {
		pixel := bg
		if s.focused && s.HighlightColor != nil {
			pixel = s.HighlightColor.Pixel
		} else if s.HighlightBackground != nil {
			pixel = s.HighlightBackground.Pixel
		}
		d.SetForeground(gc, pixel)
		for i := range hl {
			d.DrawRectangle(pix, gc, i, i, uint(w.Width-1-2*i), uint(w.Height-1-2*i))
		}
	}
}

type fontMetrics struct{ ascent, descent int }

func (s *Scale) activeBg() uint64 {
	if c, err := s.App.ColorCache().Get(widget.DefActiveBackground); err == nil {
		return c.Pixel
	}
	return 0xececec
}

// Configure applies options.
func (s *Scale) Configure(opts ...ScaleOption) error {
	return widget.Configure(s, opts, s.computeGeometry)
}

// Destroy cleans up the scale.
func (s *Scale) Destroy() {
	if s.Destroyed {
		return
	}
	s.Destroyed = true
	window.DestroyWindow(s.Win)
}

func bindScale(s *Scale, app widget.AppContext) {
	w := s.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	// Configure.
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			s.Display()
		}
	})

	// Button press.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		// Mouse wheel: Button 4 (up/left) and Button 5 (down/right).
		if ev.Button == 4 {
			step := s.Resolution
			if step <= 0 {
				step = 1
			}
			inc := step
			if s.From > s.To {
				inc = -step
			}
			s.Set(s.Value + inc)
			return
		}
		if ev.Button == 5 {
			step := s.Resolution
			if step <= 0 {
				step = 1
			}
			inc := step
			if s.From > s.To {
				inc = -step
			}
			s.Set(s.Value - inc)
			return
		}
		if ev.Button != 1 {
			return
		}
		hit := s.hitTest(ev.X, ev.Y)
		switch hit {
		case "slider":
			s.dragging = true
			sliderPos := s.valueToPixel(s.Value)
			if s.Orient == Horizontal {
				s.dragOffset = ev.X - sliderPos
			} else {
				s.dragOffset = ev.Y - sliderPos
			}
		case "trough1":
			step := s.Resolution
			if step <= 0 {
				step = 1
			}
			if s.From <= s.To {
				s.Set(s.Value - step*10)
			} else {
				s.Set(s.Value + step*10)
			}
		case "trough2":
			step := s.Resolution
			if step <= 0 {
				step = 1
			}
			if s.From <= s.To {
				s.Set(s.Value + step*10)
			} else {
				s.Set(s.Value - step*10)
			}
		}
	})

	// Button release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		s.dragging = false
	})

	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		s.focused = ev.Type == event.FocusInType
		s.Display()
	})

	// Motion: scale.tcl's tk::ScaleActivate, then drag.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask|event.LeaveMask, func(ev *event.Event) {
		active := ev.Type == event.MotionType && (s.dragging || s.hitTest(ev.X, ev.Y) == "slider")
		if active != s.active {
			s.active = active
			s.Display()
		}
		if ev.Type != event.MotionType || !s.dragging {
			return
		}
		var pos int
		if s.Orient == Horizontal {
			pos = ev.X - s.dragOffset
		} else {
			pos = ev.Y - s.dragOffset
		}
		value := s.pixelToValue(pos)
		s.Set(value)
	})

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		step := s.Resolution
		if step <= 0 {
			step = 1
		}
		inc := step
		if s.From > s.To {
			inc = -step
		}

		switch ev.KeySym {
		case platform.XK_Left, platform.XK_Down:
			s.Set(s.Value - inc)
		case platform.XK_Right, platform.XK_Up:
			s.Set(s.Value + inc)
		case platform.XK_Home:
			s.Set(s.From)
		case platform.XK_End:
			s.Set(s.To)
		}
	})
}
