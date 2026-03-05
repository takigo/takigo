// Package scale implements a scale (slider) widget.
// It ports tk/generic/tkScale.c.
package scale

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Orient specifies the scale orientation.
type Orient int

const (
	Horizontal Orient = iota
	Vertical
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
	// Length sets the preferred length along the primary axis in pixels.
	// 0 means use the default (200px).
	Length  int
	Command func(float64)

	// Interaction state.
	dragging   bool
	dragOffset int

	// Colors.
	TroughColor *colorRef
}

type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// ScaleOption configures a Scale.
type ScaleOption func(*Scale)

func OrientOpt(o Orient) ScaleOption    { return func(s *Scale) { s.Orient = o } }
func FromOpt(v float64) ScaleOption     { return func(s *Scale) { s.From = v } }
func ToOpt(v float64) ScaleOption       { return func(s *Scale) { s.To = v } }
func ValueOpt(v float64) ScaleOption    { return func(s *Scale) { s.Value = v } }
func ResolutionOpt(v float64) ScaleOption { return func(s *Scale) { s.Resolution = v } }
func ShowValueOpt(b bool) ScaleOption   { return func(s *Scale) { s.ShowValue = b } }
func LabelOpt(s string) ScaleOption     { return func(sc *Scale) { sc.Label = s } }
func SliderLengthOpt(n int) ScaleOption { return func(s *Scale) { s.SliderLength = n } }
func WidthOpt(w int) ScaleOption        { return func(s *Scale) { s.Width = w } }
func TickIntervalOpt(v float64) ScaleOption { return func(s *Scale) { s.TickInterval = v } }
func LengthOpt(n int) ScaleOption           { return func(s *Scale) { s.Length = n } }
func CommandOpt(fn func(float64)) ScaleOption { return func(s *Scale) { s.Command = fn } }

func Background(name string) ScaleOption {
	return func(s *Scale) {
		col, err := s.App.ColorCache().Get(name)
		if err == nil {
			s.Background = col
			s.UpdateBorder()
		}
	}
}

// New creates a new Scale widget.
func New(parent widget.Caregiver, name string, opts ...ScaleOption) *Scale {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	s := &Scale{
		Orient:       Horizontal,
		From:         0,
		To:           100,
		Resolution:   1,
		ShowValue:    true,
		SliderLength: 30,
		Width:        15,
	}
	widget.InitBase(&s.Base, w, app)
	s.BorderWidth = 1
	s.Relief = option.ReliefFlat

	if tc, err := app.ColorCache().Get("#c3c3c3"); err == nil {
		s.TroughColor = &colorRef{tc.Pixel, tc.Red, tc.Green, tc.Blue}
	}

	for _, opt := range opts {
		opt(s)
	}

	s.computeGeometry()

	if s.Background != nil {
		w.BackgroundPixel = s.Background.Pixel
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

func (s *Scale) roundToResolution(v float64) float64 {
	if s.Resolution <= 0 {
		return v
	}
	return math.Round(v/s.Resolution) * s.Resolution
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

// pixelRange returns the available pixel range for the slider.
func (s *Scale) pixelRange() (start, length int) {
	w := s.Win
	inset := s.BorderWidth + 2
	if s.Orient == Horizontal {
		start = inset
		length = w.Width - 2*inset - s.SliderLength
	} else {
		start = inset
		length = w.Height - 2*inset - s.SliderLength
	}
	if length < 1 {
		length = 1
	}
	return
}

// valueToPixel converts a value to a pixel position.
func (s *Scale) valueToPixel(value float64) int {
	start, pxRange := s.pixelRange()
	vRange := s.To - s.From
	if vRange == 0 {
		return start
	}
	fraction := (value - s.From) / vRange
	return start + int(fraction*float64(pxRange))
}

// pixelToValue converts a pixel position to a value.
func (s *Scale) pixelToValue(pixel int) float64 {
	start, pxRange := s.pixelRange()
	if pxRange <= 0 {
		return s.From
	}
	fraction := float64(pixel-start) / float64(pxRange)
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	return s.From + fraction*(s.To-s.From)
}

// hitTest returns "slider", "trough1", or "trough2" for a pixel position.
func (s *Scale) hitTest(x, y int) string {
	var pos int
	if s.Orient == Horizontal {
		pos = x
	} else {
		pos = y
	}
	sliderPos := s.valueToPixel(s.Value)
	if pos >= sliderPos && pos < sliderPos+s.SliderLength {
		return "slider"
	}
	if pos < sliderPos {
		return "trough1"
	}
	return "trough2"
}

func (s *Scale) computeGeometry() {
	w := s.Win
	preferredLength := 200
	if s.Length > 0 {
		preferredLength = s.Length
	}
	if s.Orient == Horizontal {
		w.ReqWidth = preferredLength
		w.ReqHeight = s.Width + 4
		if s.ShowValue {
			if s.Font != nil {
				m := s.Font.Metrics()
				w.ReqHeight += m.Linespace() + 2
			}
		}
		if s.TickInterval > 0 && s.Font != nil {
			m := s.Font.Metrics()
			w.ReqHeight += 5 + m.Linespace() + 2 // tick line + label
		}
		if s.Label != "" && s.Font != nil {
			m := s.Font.Metrics()
			w.ReqHeight += m.Linespace() + 2
		}
	} else {
		w.ReqWidth = s.Width + 4
		w.ReqHeight = preferredLength
		if s.ShowValue {
			if s.Font != nil {
				valStr := s.formatValue(s.To)
				valW := s.Font.MeasureString(valStr)
				w.ReqWidth += valW + 4
			}
		}
		if s.TickInterval > 0 && s.Font != nil {
			valStr := s.formatValue(s.To)
			valW := s.Font.MeasureString(valStr)
			extra := 5 + valW + 4 // tick line + label
			if extra > w.ReqWidth-s.Width-4 {
				w.ReqWidth = s.Width + 4 + extra
			}
		}
		if s.Label != "" && s.Font != nil {
			lblW := s.Font.MeasureString(s.Label)
			if lblW+4 > w.ReqWidth {
				w.ReqWidth = lblW + 4
			}
		}
	}
}

func (s *Scale) formatValue(v float64) string {
	if s.Resolution >= 1 {
		return fmt.Sprintf("%.0f", v)
	}
	return fmt.Sprintf("%.1f", v)
}

// troughRect returns the trough rectangle position.
func (s *Scale) troughRect() (x, y, w, h int) {
	win := s.Win
	inset := s.BorderWidth
	if s.Orient == Horizontal {
		ty := inset
		if s.ShowValue && s.Font != nil {
			m := s.Font.Metrics()
			ty += m.Linespace() + 2
		}
		return inset, ty, win.Width - 2*inset, s.Width
	}
	tx := inset
	return tx, inset, s.Width, win.Height - 2*inset
}

// Display draws the scale.
func (s *Scale) Display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if s.Background != nil {
		d.SetForeground(gc, s.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Trough.
	tx, ty, tw, th := s.troughRect()
	if s.TroughColor != nil {
		d.SetForeground(gc, s.TroughColor.Pixel)
		d.FillRectangle(w.Drawable(), gc, tx, ty, uint(tw), uint(th))
	}
	if s.Border != nil {
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			tx, ty, tw, th, 1, option.ReliefSunken)
	}

	// Slider.
	sliderPos := s.valueToPixel(s.Value)
	if s.Orient == Horizontal {
		sliderX := sliderPos
		sliderY := ty
		sliderW := s.SliderLength
		sliderH := th
		if s.Background != nil {
			d.SetForeground(gc, s.Background.Pixel)
			d.FillRectangle(w.Drawable(), gc, sliderX, sliderY,
				uint(sliderW), uint(sliderH))
		}
		if s.Border != nil {
			draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
				sliderX, sliderY, sliderW, sliderH, 2, option.ReliefRaised)
		}
	} else {
		sliderX := tx
		sliderY := sliderPos
		sliderW := tw
		sliderH := s.SliderLength
		if s.Background != nil {
			d.SetForeground(gc, s.Background.Pixel)
			d.FillRectangle(w.Drawable(), gc, sliderX, sliderY,
				uint(sliderW), uint(sliderH))
		}
		if s.Border != nil {
			draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
				sliderX, sliderY, sliderW, sliderH, 2, option.ReliefRaised)
		}
	}

	// Tick marks and labels.
	if s.TickInterval > 0 && s.Font != nil && s.Foreground != nil {
		if df, ok := s.Font.(platform.DrawableFont); ok {
			m := s.Font.Metrics()
			_, ty, tw2, th := s.troughRect()
			vRange := s.To - s.From
			if vRange != 0 && s.TickInterval > 0 {
				_, pxRange := s.pixelRange()
				pxStart, _ := s.pixelRange()
				if s.Orient == Horizontal {
					tickY := ty + th + 2
					for tv := s.From; tv <= s.To+s.TickInterval*0.001; tv += s.TickInterval {
						px := pxStart + int((tv-s.From)/vRange*float64(pxRange)) + s.SliderLength/2
						d.SetForeground(gc, s.Foreground.Pixel)
						d.DrawLine(w.Drawable(), gc, px, tickY, px, tickY+4)
						label := s.formatValue(tv)
						lw := s.Font.MeasureString(label)
						df.DrawString(w.Drawable(), px-lw/2, tickY+5+m.Ascent, label,
							s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
					}
				} else {
					tickX := s.BorderWidth + tw2 + 2
					for tv := s.From; tv <= s.To+s.TickInterval*0.001; tv += s.TickInterval {
						py := pxStart + int((tv-s.From)/vRange*float64(pxRange)) + s.SliderLength/2
						d.SetForeground(gc, s.Foreground.Pixel)
						d.DrawLine(w.Drawable(), gc, tickX, py, tickX+4, py)
						label := s.formatValue(tv)
						df.DrawString(w.Drawable(), tickX+6, py+m.Ascent/2, label,
							s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
					}
				}
			}
		}
	}

	// Value text.
	if s.ShowValue && s.Font != nil && s.Foreground != nil {
		valStr := s.formatValue(s.Value)
		if df, ok := s.Font.(platform.DrawableFont); ok {
			m := s.Font.Metrics()
			if s.Orient == Horizontal {
				valW := s.Font.MeasureString(valStr)
				vx := sliderPos + s.SliderLength/2 - valW/2
				vy := s.BorderWidth + m.Ascent
				df.DrawString(w.Drawable(), vx, vy, valStr,
					s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
			} else {
				vx := tx + tw + 4
				vy := sliderPos + s.SliderLength/2 + m.Ascent/2
				df.DrawString(w.Drawable(), vx, vy, valStr,
					s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
			}
		}
	}

	// Label text.
	if s.Label != "" && s.Font != nil && s.Foreground != nil {
		if df, ok := s.Font.(platform.DrawableFont); ok {
			m := s.Font.Metrics()
			if s.Orient == Horizontal {
				_, _, _, troughH := s.troughRect()
				ly := s.BorderWidth + troughH
				if s.ShowValue {
					ly += m.Linespace() + 2
				}
				ly += m.Ascent + 2
				df.DrawString(w.Drawable(), s.BorderWidth+4, ly, s.Label,
					s.Foreground.Pixel, s.Foreground.Red, s.Foreground.Green, s.Foreground.Blue)
			}
		}
	}

	d.Flush()
}

// Configure applies options.
func (s *Scale) Configure(opts ...option.Option) {
	option.Apply(s, opts)
	s.UpdateBorder()
	s.computeGeometry()
	if s.Background != nil {
		s.Win.BackgroundPixel = s.Background.Pixel
	}
	s.Display()
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

	// Motion (drag).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if !s.dragging {
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
