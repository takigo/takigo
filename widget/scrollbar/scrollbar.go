// Package scrollbar implements a scrollbar widget.
// It ports tk/generic/tkScrollbar.c.
package scrollbar

import (
	"github.com/msorc/takigo/color"
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
	TroughColor *color.ColorRef
	Width       int // scrollbar width (perpendicular to orient)
	ElementBW   int // element border width

	// Command callback: called when user interacts.
	Command func(args ...any)
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
func CommandOpt(fn func(args ...any)) ScrollbarOption {
	return func(s *Scrollbar) { s.Command = fn }
}

// --- Ttk-compatible aliases (prefix with Scrollbar) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// ScrollbarOrientOpt is an alias for OrientOpt.
var ScrollbarOrientOpt = OrientOpt

// ScrollbarWidthOpt is an alias for WidthOpt.
var ScrollbarWidthOpt = WidthOpt

// ScrollbarCommandOpt is an alias for CommandOpt.
var ScrollbarCommandOpt = CommandOpt

// New creates a new Scrollbar widget.
func New(parent widget.Caregiver, name string, opts ...ScrollbarOption) *Scrollbar {
	app := parent.AppContext()
	s := &Scrollbar{
		Orient:    Vertical,
		First:     0,
		Last:      1,
		Width:     11, // DEF_SCROLLBAR_WIDTH
		ElementBW: 1,
	}

	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)
	widget.InitBase(&s.Base, w, app)
	s.SetDisplayProc(s.display)
	w.Class = "Scrollbar"

	s.BorderWidth = 1
	s.Relief = option.ReliefSunken

	// Trough color.
	if tc, err := app.ColorCache().Get("#c3c3c3"); err == nil {
		s.TroughColor = tc.Ref()
	}

	for _, opt := range opts {
		opt(s)
	}

	// TkpComputeScrollbarGeometry (tk/unix/tkUnixScrlbr.c): the thickness is
	// -width plus the inset; the length leaves room for both arrows, whose
	// length is thickness - 2*inset + 1.
	inset := s.BorderWidth + s.HighlightWidth
	s.arrowSize = s.Width + 1
	thick := s.Width + 2*inset
	length := 2 * (s.arrowSize + s.BorderWidth + inset)
	if s.Orient == Vertical {
		w.ReqWidth, w.ReqHeight = thick, length
	} else {
		w.ReqWidth, w.ReqHeight = length, thick
	}

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

// computeGeometry ports the slider part of TkpComputeScrollbarGeometry
// (tk/unix/tkUnixScrlbr.c).
func (s *Scrollbar) computeGeometry() {
	w := s.Win
	inset := s.BorderWidth + s.HighlightWidth
	thick, totalLen := w.Width, w.Height
	if s.Orient == Horizontal {
		thick, totalLen = w.Height, w.Width
	}
	const minSliderLength = 5 // MIN_SLIDER_LENGTH
	s.arrowSize = thick - 2*inset + 1
	fieldLength := max(0, totalLen-2*(s.arrowSize+inset))
	first := int(float64(fieldLength) * s.First)
	last := int(float64(fieldLength) * s.Last)
	first = max(0, min(first, fieldLength-minSliderLength))
	last = min(max(last, first+minSliderLength), fieldLength)
	s.troughStart = s.arrowSize + inset
	s.troughEnd = s.troughStart + fieldLength
	s.thumbStart = first + s.troughStart
	s.thumbEnd = last + s.troughStart
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (s *Scrollbar) Display() {
	s.EventuallyRedraw()
}

// display ports TkpDisplayScrollbar: highlight ring, outer border, trough,
// 3D triangle arrows and the slider, the active element drawn with the
// active background.
func (s *Scrollbar) display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) || s.Border == nil || s.Background == nil {
		return
	}
	d := w.Display.Server
	gc := w.GC
	s.computeGeometry()

	hl := s.HighlightWidth
	inset := s.BorderWidth + hl
	width := w.Width - 2*inset
	if s.Orient == Horizontal {
		width = w.Height - 2*inset
	}
	ebw := s.ElementBW
	if ebw < 0 {
		ebw = s.BorderWidth
	}

	d.SetForeground(gc, s.Background.Pixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))
	s.DrawHighlightBorder(false, 0)
	draw.Draw3DRectangle(d, w.Drawable(), gc, s.Border,
		hl, hl, w.Width-2*hl, w.Height-2*hl, s.BorderWidth, s.Relief)
	if s.TroughColor != nil {
		d.SetForeground(gc, s.TroughColor.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, inset, inset,
		uint(max(0, w.Width-2*inset)), uint(max(0, w.Height-2*inset)))

	activeBorder := s.Border
	if ac, err := s.App.ColorCache().Get(widget.DefActiveBackground); err == nil {
		activeBorder = draw.NewBorder(ac.Red, ac.Green, ac.Blue)
	}
	borderFor := func(r region) *draw.Border {
		if s.activeRegion == r {
			return activeBorder
		}
		return s.Border
	}
	pt := func(x, y int) platform.Point { return platform.Point{X: int16(x), Y: int16(y)} }
	al := s.arrowSize
	var top, bottom []platform.Point
	if s.Orient == Vertical {
		top = []platform.Point{pt(inset-1, al+inset-1), pt(width+inset, al+inset-1), pt(width/2+inset, inset-1)}
		y0 := w.Height - al - inset + 1
		bottom = []platform.Point{pt(inset, y0), pt(width/2+inset, w.Height-inset), pt(width+inset, y0)}
	} else {
		top = []platform.Point{pt(al+inset-1, inset-1), pt(inset, width/2+inset), pt(al+inset-1, width+inset)}
		x0 := w.Width - al - inset + 1
		bottom = []platform.Point{pt(x0, inset-1), pt(x0, width+inset), pt(w.Width-inset, width/2+inset)}
	}
	draw.Fill3DPolygon(d, w.Drawable(), gc, borderFor(regionArrow1), top, ebw, option.ReliefRaised)
	draw.Fill3DPolygon(d, w.Drawable(), gc, borderFor(regionArrow2), bottom, ebw, option.ReliefRaised)

	if s.Orient == Vertical {
		draw.Fill3DRectangle(d, w.Drawable(), gc, borderFor(regionThumb),
			inset, s.thumbStart, width, s.thumbEnd-s.thumbStart, ebw, option.ReliefRaised)
	} else {
		draw.Fill3DRectangle(d, w.Drawable(), gc, borderFor(regionThumb),
			s.thumbStart, inset, s.thumbEnd-s.thumbStart, width, ebw, option.ReliefRaised)
	}
}

// hitTest returns which region a pixel coordinate falls in.
func (s *Scrollbar) hitTest(x, y int) region {
	var pos int
	if s.Orient == Vertical {
		pos = y
	} else {
		pos = x
	}

	if pos < s.troughStart {
		return regionArrow1
	}
	if pos >= s.troughEnd {
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
