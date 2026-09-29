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

	// Minimum length as in Tk's default theme: two arrow boxes plus the
	// 8px minimum thumb (38px for the default 9p arrows).
	minLen := 2*s.sbWidth + 8
	if s.Orient == Vertical {
		win.ReqWidth = s.sbWidth
		win.ReqHeight = minLen
	} else {
		win.ReqWidth = minLen
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

// Default theme scrollbar metrics (tk/generic/ttk/ttkDefaultTheme.c): a 1px
// sunken trough border, square arrow boxes filling the rest of the
// thickness, and a minimum thumb of MIN_THUMB_SIZE plus its 1px border.
const (
	sbTroughBorder = 1
	sbMinThumb     = 8 + 2
)

// arrowBox returns the size of the square arrow boxes.
func (s *Scrollbar) arrowBox() int {
	thick := s.Win.Width
	if s.Orient == Horizontal {
		thick = s.Win.Height
	}
	return max(0, thick-2*sbTroughBorder)
}

// computeGeometry ports ScrollbarDoLayout (tk/generic/ttk/ttkScrollbar.c):
// the thumb moves over the trough minus its own minimum size.
func (s *Scrollbar) computeGeometry() {
	totalLen := s.Win.Height
	if s.Orient == Horizontal {
		totalLen = s.Win.Width
	}
	box := s.arrowBox()
	s.troughStart = sbTroughBorder + box
	s.troughEnd = max(totalLen-sbTroughBorder-box, s.troughStart)
	// A bar too short for the minimum thumb gets a thumb filling the
	// trough rather than a negative length.
	size := float64(max(s.troughEnd-s.troughStart-sbMinThumb, 0))
	s.thumbStart = s.troughStart + int(size*s.First)
	s.thumbEnd = min(s.troughStart+int(size*s.Last)+sbMinThumb, s.troughEnd)
	s.thumbStart = min(s.thumbStart, s.thumbEnd)
}

// Display draws the default theme's Horizontal/Vertical.Scrollbar layout:
// trough (TroughElement, 1px sunken), the two arrows (ArrowElement with a thin
// raised border and a padded triangle) and the thumb (thin raised).
func (s *Scrollbar) Display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == 0 || w.Width <= 0 || w.Height <= 0 {
		return
	}
	d := w.Display.Server
	gc := w.GC
	s.computeGeometry()

	troughColor := uint64(0xc3c3c3)
	bgColor := uint64(0xd9d9d9)
	arrowColor := uint64(0x000000)
	if s.First <= 0 && s.Last >= 1 {
		arrowColor = 0xa3a3a3 // disabled: colors(-disabledfg)
	}
	if s.Context != nil && s.Context.Style != nil {
		troughColor = LookupColor(s.Context.Style, "-troughcolor", s.State, troughColor)
		bgColor = LookupColor(s.Context.Style, "-background", s.State, bgColor)
	}
	trough := draw.NewBorderFromPixel(troughColor)
	border := draw.NewBorderFromPixel(bgColor)

	fill3DRectangle(d, w.Drawable(), gc, trough, Box{0, 0, w.Width, w.Height},
		sbTroughBorder, option.ReliefSunken)

	// rect maps (along, across, length, thickness) to window coordinates.
	rect := func(along, across, length, thick int) (int, int, int, int) {
		if s.Orient == Vertical {
			return across, along, thick, length
		}
		return along, across, length, thick
	}
	thinRaised := func(x, y, bw, bh int) {
		if bw <= 0 || bh <= 0 {
			return
		}
		d.SetForeground(gc, bgColor)
		d.FillRectangle(w.Drawable(), gc, x, y, uint(bw), uint(bh))
		// DrawBorder with borderWidth 1: thinShadowColors[raised] = LITE, DARK.
		d.SetForeground(gc, border.LightPixel)
		d.DrawLine(w.Drawable(), gc, x, y+bh-1, x, y)
		d.DrawLine(w.Drawable(), gc, x, y, x+bw-1, y)
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(w.Drawable(), gc, x, y+bh-1, x+bw-1, y+bh-1)
		d.DrawLine(w.Drawable(), gc, x+bw-1, y+bh-1, x+bw-1, y)
	}

	box := s.arrowBox()
	totalLen := w.Width
	dir1, dir2 := arrowLeft, arrowRight
	if s.Orient == Vertical {
		totalLen = w.Height
		dir1, dir2 = arrowUp, arrowDown
	}
	for i, along := range []int{sbTroughBorder, totalLen - sbTroughBorder - box} {
		x, y, bw, bh := rect(along, sbTroughBorder, box, box)
		thinRaised(x, y, bw, bh)
		dir := dir1
		if i == 1 {
			dir = dir2
		}
		drawArrowInBox(d, w.Drawable(), gc, arrowColor, x, y, bw, bh, dir)
	}
	x, y, bw, bh := rect(s.thumbStart, sbTroughBorder, s.thumbEnd-s.thumbStart, box)
	thinRaised(x, y, bw, bh)

	d.Flush()
}

// drawArrowInBox follows ArrowElementDraw: pad the box by ArrowPadding
// {3,3,4,4}, size the arrow with TtkArrowSize, centre it and fill it.
func drawArrowInBox(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	color uint64, x, y, w, h, dir int) {
	x, y, w, h = x+3, y+3, w-7, h-7
	var cx, cy int
	switch dir {
	case arrowUp, arrowDown:
		hh := w / 2
		cx, cy = 2*hh+1, hh+1
		if (h-cy)%2 == 1 {
			cy++
		}
	default:
		hh := h / 2
		cx, cy = hh+1, 2*hh+1
		if (w-cx)%2 == 1 {
			cx++
		}
	}
	// Ttk_AnchorBox(center): C integer division truncates toward zero.
	ax := x + (w-cx)/2
	ay := y + (h-cy)/2
	// Measured against Tk 9.1: down/right arrows sit one pixel further along.
	switch dir {
	case arrowDown:
		ay++
	case arrowRight:
		ax++
	}
	drawScrollArrow(d, drawable, gc, color, ax, ay, cx, cy, dir)
}

// hitTest returns which region a pixel coordinate falls in.
func (s *Scrollbar) hitTest(x, y int) sbRegion {
	var pos int
	if s.Orient == Vertical {
		pos = y
	} else {
		pos = x
	}

	if pos < s.troughStart {
		return sbArrow1
	}
	if pos >= s.troughEnd {
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

		troughLen := s.troughEnd - s.troughStart - sbMinThumb
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
	d.DrawLine(drawable, gc, int(points[2].X), int(points[2].Y), int(points[2].X), int(points[2].Y))
}
