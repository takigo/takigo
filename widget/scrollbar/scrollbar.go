// Package scrollbar implements a scrollbar widget.
// It ports tk/generic/tkScrollbar.c.
package scrollbar

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Orient specifies the scrollbar orientation.
type Orient int

const (
	Vertical Orient = iota
	Horizontal
)

// Scrollbar is a scrollbar widget.
type Scrollbar struct {
	widget.Base

	Orient Orient

	// Thumb position (0.0 to 1.0).
	First float64
	Last  float64

	// Pixel geometry of the trough and thumb.
	troughStart int // pixel start of trough (after arrow zone)
	troughEnd   int // pixel end of trough
	thumbStart  int // pixel start of thumb
	thumbEnd    int // pixel end of thumb
	arrowSize   int // arrow button size

	// Interaction.
	activeRegion region // what is under the pointer
	pressRegion  region // what was pressed
	dragOffset   int    // offset from thumb start during drag

	// Visual config.
	TroughColor *colorRef
	Width       int // scrollbar width (perpendicular to orient)
	ElementBW   int // element border width

	// Command callback: called when user interacts.
	Command func(args ...interface{})
}

type region int

const (
	regionNone region = iota
	regionArrow1
	regionArrow2
	regionThumb
	regionTroughBefore
	regionTroughAfter
)

type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// ScrollbarOption configures a Scrollbar.
type ScrollbarOption func(*Scrollbar)

// OrientOpt sets the scrollbar orientation.
func OrientOpt(o Orient) ScrollbarOption {
	return func(s *Scrollbar) { s.Orient = o }
}

// WidthOpt sets the scrollbar width.
func WidthOpt(w int) ScrollbarOption {
	return func(s *Scrollbar) { s.Width = w }
}

// CommandOpt sets the scroll command callback.
func CommandOpt(fn func(args ...interface{})) ScrollbarOption {
	return func(s *Scrollbar) { s.Command = fn }
}

// New creates a new Scrollbar widget.
func New(parent widget.Caregiver, name string, opts ...ScrollbarOption) *Scrollbar {
	app := parent.AppContext()
	s := &Scrollbar{
		Orient:    Vertical,
		First:     0,
		Last:      1,
		Width:     15,
		ElementBW: 1,
	}

	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)
	widget.InitBase(&s.Base, w, app)

	s.BorderWidth = 1
	s.Relief = option.ReliefSunken

	// Trough color.
	if tc, err := app.ColorCache().Get("#c3c3c3"); err == nil {
		s.TroughColor = &colorRef{tc.Pixel, tc.Red, tc.Green, tc.Blue}
	}

	for _, opt := range opts {
		opt(s)
	}

	// Set requested size.
	if s.Orient == Vertical {
		w.ReqWidth = s.Width + 2*s.BorderWidth
		w.ReqHeight = 100
	} else {
		w.ReqWidth = 100
		w.ReqHeight = s.Width + 2*s.BorderWidth
	}

	s.arrowSize = s.Width

	if s.Background != nil {
		w.BackgroundPixel = s.Background.Pixel
	}

	bindScrollbar(s, app)
	return s
}

// Set updates the thumb position. Called by the widget being scrolled.
func (s *Scrollbar) Set(first, last float64) {
	if first < 0 {
		first = 0
	}
	if last > 1 {
		last = 1
	}
	if first >= last {
		first = 0
		last = 1
	}
	s.First = first
	s.Last = last
	s.computeGeometry()
	s.Display()
}

// computeGeometry calculates thumb pixel positions.
func (s *Scrollbar) computeGeometry() {
	w := s.Win
	var totalLen int
	if s.Orient == Vertical {
		totalLen = w.Height
	} else {
		totalLen = w.Width
	}

	bw := s.BorderWidth
	s.troughStart = bw + s.arrowSize
	s.troughEnd = totalLen - bw - s.arrowSize
	troughLen := s.troughEnd - s.troughStart

	if troughLen < 1 {
		troughLen = 1
	}

	s.thumbStart = s.troughStart + int(s.First*float64(troughLen))
	s.thumbEnd = s.troughStart + int(s.Last*float64(troughLen))

	minThumb := s.Width / 2
	if minThumb < 6 {
		minThumb = 6
	}
	if s.thumbEnd-s.thumbStart < minThumb {
		s.thumbEnd = s.thumbStart + minThumb
		if s.thumbEnd > s.troughEnd {
			s.thumbEnd = s.troughEnd
			s.thumbStart = s.thumbEnd - minThumb
		}
	}
}

// Display draws the scrollbar.
func (s *Scrollbar) Display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	s.computeGeometry()

	// Trough background.
	if s.TroughColor != nil {
		d.SetForeground(gc, s.TroughColor.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Outer border.
	if s.Border != nil && s.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			0, 0, w.Width, w.Height, s.BorderWidth, s.Relief)
	}

	if s.Border == nil || s.Background == nil {
		d.Flush()
		return
	}

	bw := s.BorderWidth

	if s.Orient == Vertical {
		sbWidth := w.Width - 2*bw

		// Arrow 1 (up).
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			bw, bw, sbWidth, s.arrowSize, s.ElementBW, option.ReliefRaised)
		// Arrow 2 (down).
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			bw, w.Height-bw-s.arrowSize, sbWidth, s.arrowSize, s.ElementBW, option.ReliefRaised)
		// Thumb.
		if s.thumbEnd > s.thumbStart {
			d.SetForeground(gc, s.Background.Pixel)
			d.FillRectangle(w.Drawable(), gc, bw, s.thumbStart,
				uint(sbWidth), uint(s.thumbEnd-s.thumbStart))
			draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
				bw, s.thumbStart, sbWidth, s.thumbEnd-s.thumbStart,
				s.ElementBW, option.ReliefRaised)
		}
	} else {
		sbHeight := w.Height - 2*bw

		// Arrow 1 (left).
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			bw, bw, s.arrowSize, sbHeight, s.ElementBW, option.ReliefRaised)
		// Arrow 2 (right).
		draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
			w.Width-bw-s.arrowSize, bw, s.arrowSize, sbHeight, s.ElementBW, option.ReliefRaised)
		// Thumb.
		if s.thumbEnd > s.thumbStart {
			d.SetForeground(gc, s.Background.Pixel)
			d.FillRectangle(w.Drawable(), gc, s.thumbStart, bw,
				uint(s.thumbEnd-s.thumbStart), uint(sbHeight))
			draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
				s.thumbStart, bw, s.thumbEnd-s.thumbStart, sbHeight,
				s.ElementBW, option.ReliefRaised)
		}
	}

	d.Flush()
}

// hitTest returns which region a pixel coordinate falls in.
func (s *Scrollbar) hitTest(x, y int) region {
	var pos int
	if s.Orient == Vertical {
		pos = y
	} else {
		pos = x
	}

	bw := s.BorderWidth
	var totalLen int
	if s.Orient == Vertical {
		totalLen = s.Win.Height
	} else {
		totalLen = s.Win.Width
	}

	if pos < bw+s.arrowSize {
		return regionArrow1
	}
	if pos >= totalLen-bw-s.arrowSize {
		return regionArrow2
	}
	if pos >= s.thumbStart && pos < s.thumbEnd {
		return regionThumb
	}
	if pos < s.thumbStart {
		return regionTroughBefore
	}
	return regionTroughAfter
}

// Configure applies options.
func (s *Scrollbar) Configure(opts ...option.Option) {
	option.Apply(s, opts)
	s.computeGeometry()
	s.Display()
}

// Destroy cleans up the scrollbar.
func (s *Scrollbar) Destroy() {
	if s.Destroyed {
		return
	}
	s.Destroyed = true
	window.DestroyWindow(s.Win)
}

func bindScrollbar(s *Scrollbar, app widget.AppContext) {
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
			s.computeGeometry()
			s.Display()
		}
	})

	// Button press.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		// Mouse wheel: Button 4 (up) and Button 5 (down).
		if ev.Button == 4 {
			if s.Command != nil {
				s.Command("scroll", -3, "units")
			}
			return
		}
		if ev.Button == 5 {
			if s.Command != nil {
				s.Command("scroll", 3, "units")
			}
			return
		}
		if ev.Button != 1 {
			return
		}
		rgn := s.hitTest(ev.X, ev.Y)
		s.pressRegion = rgn

		switch rgn {
		case regionArrow1:
			if s.Command != nil {
				s.Command("scroll", -1, "units")
			}
		case regionArrow2:
			if s.Command != nil {
				s.Command("scroll", 1, "units")
			}
		case regionTroughBefore:
			if s.Command != nil {
				s.Command("scroll", -1, "pages")
			}
		case regionTroughAfter:
			if s.Command != nil {
				s.Command("scroll", 1, "pages")
			}
		case regionThumb:
			if s.Orient == Vertical {
				s.dragOffset = ev.Y - s.thumbStart
			} else {
				s.dragOffset = ev.X - s.thumbStart
			}
		}
	})

	// Button release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		s.pressRegion = regionNone
	})

	// Motion (drag thumb).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		if s.pressRegion != regionThumb {
			return
		}

		var pos int
		if s.Orient == Vertical {
			pos = ev.Y
		} else {
			pos = ev.X
		}

		troughLen := s.troughEnd - s.troughStart
		if troughLen < 1 {
			return
		}

		newThumbPos := pos - s.dragOffset
		fraction := float64(newThumbPos-s.troughStart) / float64(troughLen)
		if fraction < 0 {
			fraction = 0
		}
		if fraction > 1 {
			fraction = 1
		}

		if s.Command != nil {
			s.Command("moveto", fraction)
		}
	})
}
