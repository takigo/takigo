package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Scrollbar is a TTK themed scrollbar widget.
type Scrollbar struct {
	TtkWidget

	Orient Orientation

	// Thumb position (0.0 to 1.0).
	First float64
	Last  float64

	// Command callback: called when user interacts.
	Command func(args ...any)

	// Pixel geometry.
	troughStart int
	troughEnd   int
	thumbStart  int
	thumbEnd    int
	sbWidth     int // scrollbar width (perpendicular to orient)

	// Interaction.
	pressRegion sbRegion
	dragOffset  int
}

type sbRegion int

const (
	sbNone sbRegion = iota
	sbArrow1
	sbArrow2
	sbThumb
	sbTroughBefore
	sbTroughAfter
)

// ScrollbarOption configures a Scrollbar.
type ScrollbarOption func(*Scrollbar)

// ScrollbarOrientOpt sets the orientation.
func ScrollbarOrientOpt(o Orientation) ScrollbarOption {
	return func(s *Scrollbar) { s.Orient = o }
}

// ScrollbarCommandOpt sets the scroll command callback.
func ScrollbarCommandOpt(fn func(args ...any)) ScrollbarOption {
	return func(s *Scrollbar) { s.Command = fn }
}

// NewScrollbar creates a themed scrollbar widget.
func NewScrollbar(parent widget.Caregiver, name string, opts ...ScrollbarOption) *Scrollbar {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	s := &Scrollbar{
		Orient:  Vertical,
		First:   0,
		Last:    1,
		sbWidth: 15,
	}

	for _, opt := range opts {
		opt(s)
	}

	styleName := "Vertical.TScrollbar"
	if s.Orient == Horizontal {
		styleName = "Horizontal.TScrollbar"
	}

	InitTtkWidget(&s.TtkWidget, win, app, styleName)
	s.DisplayFunc = s.Display

	if s.Orient == Vertical {
		win.ReqWidth = s.sbWidth
		win.ReqHeight = 100
	} else {
		win.ReqWidth = 100
		win.ReqHeight = s.sbWidth
	}

	bindTtkHover(&s.TtkWidget, app)
	bindTtkScrollbar(s, app)
	return s
}

// Set updates the thumb position. Called by the scrolled widget.
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

// computeGeometry calculates pixel positions of trough and thumb.
func (s *Scrollbar) computeGeometry() {
	w := s.Win
	var totalLen int
	if s.Orient == Vertical {
		totalLen = w.Height
	} else {
		totalLen = w.Width
	}

	arrowSize := s.sbWidth
	s.troughStart = arrowSize
	s.troughEnd = totalLen - arrowSize
	troughLen := s.troughEnd - s.troughStart
	if troughLen < 1 {
		troughLen = 1
	}

	s.thumbStart = s.troughStart + int(s.First*float64(troughLen))
	s.thumbEnd = s.troughStart + int(s.Last*float64(troughLen))

	minThumb := max(s.sbWidth/2, 6)
	if s.thumbEnd-s.thumbStart < minThumb {
		s.thumbEnd = s.thumbStart + minThumb
		if s.thumbEnd > s.troughEnd {
			s.thumbEnd = s.troughEnd
			s.thumbStart = s.thumbEnd - minThumb
		}
	}
}

// Display renders the scrollbar with custom drawing (overrides layout-based TtkWidget.Display).
func (s *Scrollbar) Display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == 0 {
		return
	}

	d := w.Display.Server
	gc := w.GC
	width := w.Width
	height := w.Height
	if width <= 0 || height <= 0 {
		return
	}

	s.computeGeometry()

	// Look up theme colors (fall back to defaults if no theme loaded).
	troughColor := uint64(0xc3c3c3)
	bgColor := uint64(0xd9d9d9)
	if s.Context != nil && s.Context.Style != nil {
		troughColor = LookupColor(s.Context.Style, "-troughcolor", s.State, troughColor)
		bgColor = LookupColor(s.Context.Style, "-background", s.State, bgColor)
	}

	border := draw.NewBorderFromPixel(bgColor)
	arrowSize := s.sbWidth

	// Fill trough background.
	d.SetForeground(gc, troughColor)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(width), uint(height))

	// Arrow padding (matches Tk's ArrowPadding = {3,3,3,3}).
	arrowPad := 3

	if s.Orient == Vertical {
		// Arrow 1 (up).
		d.SetForeground(gc, bgColor)
		d.FillRectangle(w.Drawable(), gc, 0, 0, uint(width), uint(arrowSize))
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			0, 0, width, arrowSize, 1, option.ReliefRaised)
		drawScrollArrow(d, w.Drawable(), gc, 0x000000,
			arrowPad, arrowPad, width-2*arrowPad, arrowSize-2*arrowPad, arrowUp)

		// Arrow 2 (down).
		d.SetForeground(gc, bgColor)
		d.FillRectangle(w.Drawable(), gc, 0, height-arrowSize, uint(width), uint(arrowSize))
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			0, height-arrowSize, width, arrowSize, 1, option.ReliefRaised)
		drawScrollArrow(d, w.Drawable(), gc, 0x000000,
			arrowPad, height-arrowSize+arrowPad, width-2*arrowPad, arrowSize-2*arrowPad, arrowDown)

		// Thumb.
		if s.thumbEnd > s.thumbStart {
			d.SetForeground(gc, bgColor)
			d.FillRectangle(w.Drawable(), gc, 0, s.thumbStart,
				uint(width), uint(s.thumbEnd-s.thumbStart))
			draw.Draw3DRectangle(d, w.Drawable(), gc, border,
				0, s.thumbStart, width, s.thumbEnd-s.thumbStart, 1, option.ReliefRaised)
		}
	} else {
		// Arrow 1 (left).
		d.SetForeground(gc, bgColor)
		d.FillRectangle(w.Drawable(), gc, 0, 0, uint(arrowSize), uint(height))
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			0, 0, arrowSize, height, 1, option.ReliefRaised)
		drawScrollArrow(d, w.Drawable(), gc, 0x000000,
			arrowPad, arrowPad, arrowSize-2*arrowPad, height-2*arrowPad, arrowLeft)

		// Arrow 2 (right).
		d.SetForeground(gc, bgColor)
		d.FillRectangle(w.Drawable(), gc, width-arrowSize, 0, uint(arrowSize), uint(height))
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			width-arrowSize, 0, arrowSize, height, 1, option.ReliefRaised)
		drawScrollArrow(d, w.Drawable(), gc, 0x000000,
			width-arrowSize+arrowPad, arrowPad, arrowSize-2*arrowPad, height-2*arrowPad, arrowRight)

		// Thumb.
		if s.thumbEnd > s.thumbStart {
			d.SetForeground(gc, bgColor)
			d.FillRectangle(w.Drawable(), gc, s.thumbStart, 0,
				uint(s.thumbEnd-s.thumbStart), uint(height))
			draw.Draw3DRectangle(d, w.Drawable(), gc, border,
				s.thumbStart, 0, s.thumbEnd-s.thumbStart, height, 1, option.ReliefRaised)
		}
	}

	d.Flush()
}

// hitTest returns which region a pixel coordinate falls in.
func (s *Scrollbar) hitTest(x, y int) sbRegion {
	var pos int
	if s.Orient == Vertical {
		pos = y
	} else {
		pos = x
	}

	var totalLen int
	if s.Orient == Vertical {
		totalLen = s.Win.Height
	} else {
		totalLen = s.Win.Width
	}

	arrowSize := s.sbWidth
	if pos < arrowSize {
		return sbArrow1
	}
	if pos >= totalLen-arrowSize {
		return sbArrow2
	}
	if pos >= s.thumbStart && pos < s.thumbEnd {
		return sbThumb
	}
	if pos < s.thumbStart {
		return sbTroughBefore
	}
	return sbTroughAfter
}

func bindTtkScrollbar(s *Scrollbar, app widget.AppContext) {
	w := s.Win

	// Expose — override bindTtkCommon's binding.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	// Configure — override bindTtkCommon's binding.
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
		// Mouse wheel.
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
		case sbArrow1:
			if s.Command != nil {
				s.Command("scroll", -1, "units")
			}
		case sbArrow2:
			if s.Command != nil {
				s.Command("scroll", 1, "units")
			}
		case sbTroughBefore:
			if s.Command != nil {
				s.Command("scroll", -1, "pages")
			}
		case sbTroughAfter:
			if s.Command != nil {
				s.Command("scroll", 1, "pages")
			}
		case sbThumb:
			if s.Orient == Vertical {
				s.dragOffset = ev.Y - s.thumbStart
			} else {
				s.dragOffset = ev.X - s.thumbStart
			}
		}
	})

	// Button release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		s.pressRegion = sbNone
	})

	// Motion (drag thumb).
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask == 0 {
			return
		}
		if s.pressRegion != sbThumb {
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

// Arrow directions for scrollbar arrows.
const (
	arrowUp = iota
	arrowDown
	arrowLeft
	arrowRight
)

// drawScrollArrow draws a filled triangle arrow inside the given box.
// Matches Tk's ArrowPoints + TtkFillArrow algorithm.
func drawScrollArrow(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	fgColor uint64, bx, by, bw, bh, direction int) {

	if bw <= 0 || bh <= 0 {
		return
	}

	var points [4]platform.Point
	switch direction {
	case arrowUp:
		h := (bw - 1) / 2
		cx := bx + h
		cy := by
		if bh <= h {
			h = bh - 1
		}
		points[0] = platform.Point{X: int16(cx), Y: int16(cy)}
		points[1] = platform.Point{X: int16(cx - h), Y: int16(cy + h)}
		points[2] = platform.Point{X: int16(cx + h), Y: int16(cy + h)}
	case arrowDown:
		h := (bw - 1) / 2
		cx := bx + h
		cy := by + bh - 1
		if bh <= h {
			h = bh - 1
		}
		points[0] = platform.Point{X: int16(cx), Y: int16(cy)}
		points[1] = platform.Point{X: int16(cx - h), Y: int16(cy - h)}
		points[2] = platform.Point{X: int16(cx + h), Y: int16(cy - h)}
	case arrowLeft:
		h := (bh - 1) / 2
		cx := bx
		cy := by + h
		if bw <= h {
			h = bw - 1
		}
		points[0] = platform.Point{X: int16(cx), Y: int16(cy)}
		points[1] = platform.Point{X: int16(cx + h), Y: int16(cy - h)}
		points[2] = platform.Point{X: int16(cx + h), Y: int16(cy + h)}
	case arrowRight:
		h := (bh - 1) / 2
		cx := bx + bw - 1
		cy := by + h
		if bw <= h {
			h = bw - 1
		}
		points[0] = platform.Point{X: int16(cx), Y: int16(cy)}
		points[1] = platform.Point{X: int16(cx - h), Y: int16(cy - h)}
		points[2] = platform.Point{X: int16(cx - h), Y: int16(cy + h)}
	}
	points[3] = points[0]

	d.SetForeground(gc, fgColor)
	d.FillPolygon(drawable, gc, points[:3], platform.PolygonConvex, platform.CoordModeOrigin)
	d.DrawLines(drawable, gc, points[:4], platform.CoordModeOrigin)
}
