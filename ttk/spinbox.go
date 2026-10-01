package ttk

import (
	"fmt"
	"math"
	"strconv"

	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk/entrytext"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Spinbox is a themed spinbox widget — an entry with up/down buttons.
type Spinbox struct {
	TtkWidget

	// Text state.
	edit entrytext.Helper

	// Range mode.
	From      float64
	To        float64
	Increment float64
	Format    string
	Wrap      bool

	// Values mode (overrides range if non-empty).
	Values      []string
	valuesIndex int

	// Callback.
	Command func(value string)

	// Font.
	Font font.Font

	// Layout metrics.
	insetX      int
	insetY      int
	buttonWidth int
	avgWidth    int
	prefWidth   int

	// Interaction state.
	pressedButton string // "up", "down", or ""
	hasFocus      bool
	cursorOn      bool

	// Validation.
	Validate    string
	ValidateCmd func(string) bool
}

// SpinboxOption configures a TTK Spinbox.
type SpinboxOption func(*Spinbox)

// SpinboxFrom sets the minimum value for range mode.
func SpinboxFrom(v float64) SpinboxOption { return func(s *Spinbox) { s.From = v } }

// SpinboxTo sets the maximum value for range mode.
func SpinboxTo(v float64) SpinboxOption { return func(s *Spinbox) { s.To = v } }

// SpinboxIncrement sets the step size for range mode.
func SpinboxIncrement(v float64) SpinboxOption { return func(s *Spinbox) { s.Increment = v } }

// SpinboxFormat sets the printf format for numeric display.
func SpinboxFormat(f string) SpinboxOption { return func(s *Spinbox) { s.Format = f } }

// SpinboxWrap enables wrapping at bounds.
func SpinboxWrap(b bool) SpinboxOption { return func(s *Spinbox) { s.Wrap = b } }

// SpinboxValues sets the list of values for values mode.
func SpinboxValues(v []string) SpinboxOption { return func(s *Spinbox) { s.Values = v } }

// SpinboxCommand sets the value change callback.
func SpinboxCommand(fn func(string)) SpinboxOption { return func(s *Spinbox) { s.Command = fn } }

// SpinboxWidth sets the preferred width in characters.
func SpinboxWidth(w int) SpinboxOption { return func(s *Spinbox) { s.prefWidth = w } }

// SpinboxValidate sets the validation mode ("key", "all", etc.).
func SpinboxValidate(v string) SpinboxOption { return func(s *Spinbox) { s.Validate = v } }

// SpinboxValidateCmd sets the validation callback.
func SpinboxValidateCmd(fn func(string) bool) SpinboxOption {
	return func(s *Spinbox) { s.ValidateCmd = fn }
}

// NewSpinbox creates a themed spinbox widget.
func NewSpinbox(parent widget.Caregiver, name string, opts ...SpinboxOption) *Spinbox {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	s := &Spinbox{
		prefWidth:   10,
		cursorOn:    true,
		From:        0,
		To:          100,
		Increment:   1,
		Format:      "%.0f",
		buttonWidth: 16,
		insetX:      4,
		insetY:      2,
	}

	s.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	s.edit = entrytext.Helper{
		Font:  s.Font,
		TextX: s.insetX + 2,
		App:   app,
		Win:   win,
		Redraw: func() {
			s.Display()
		},
		Editable: func() bool { return s.State&(StateDisabled|StateReadonly) == 0 },
		Validate: s.tryValidate,
	}

	InitTtkWidget(&s.TtkWidget, win, app, "TSpinbox")
	s.reconfigure = func() { _ = s.Configure() }
	s.DisplayFunc = s.Display

	for _, opt := range opts {
		opt(s)
	}

	// Set initial value.
	if len(s.Values) > 0 {
		s.edit.Text = []rune(s.Values[0])
	} else {
		s.edit.Text = []rune(s.formatValue(s.From))
	}
	s.edit.InsertPos = len(s.edit.Text)
	s.edit.SelFirst = -1
	s.edit.SelLast = -1

	s.computeGeometry()

	win.Flags |= window.FlagFocusable
	win.SetCursor(cursor.XTerm)
	bindSpinbox(s, app)

	return s
}

// Configure sets options after creation.
func (s *Spinbox) Configure(opts ...SpinboxOption) error {
	return configure(&s.TtkWidget, s, opts, nil, s.computeGeometry)
}

// Get returns the current text value.
func (s *Spinbox) Get() string {
	return s.edit.Get()
}

// Set sets the text value.
func (s *Spinbox) Set(text string) {
	s.edit.Set(text)
	if s.edit.InsertPos > len(s.edit.Text) {
		s.edit.InsertPos = len(s.edit.Text)
	}
	s.computeGeometry()
}

// SpinUp increments the value.
func (s *Spinbox) SpinUp() {
	if len(s.Values) > 0 {
		// Find current value in list.
		cur := string(s.edit.Text)
		s.valuesIndex = -1
		for i, v := range s.Values {
			if v == cur {
				s.valuesIndex = i
				break
			}
		}
		s.valuesIndex++
		if s.valuesIndex >= len(s.Values) {
			if s.Wrap {
				s.valuesIndex = 0
			} else {
				s.valuesIndex = len(s.Values) - 1
			}
		}
		s.Set(s.Values[s.valuesIndex])
	} else {
		val := s.currentNumericValue()
		val += s.Increment
		lo, hi := s.From, s.To
		if lo > hi {
			lo, hi = hi, lo
		}
		if val > hi {
			if s.Wrap {
				val = lo
			} else {
				val = hi
			}
		}
		s.Set(s.formatValue(val))
	}
	s.selectAll()
	s.Display()
	s.fireCommand()
}

// SpinDown decrements the value.
func (s *Spinbox) SpinDown() {
	if len(s.Values) > 0 {
		cur := string(s.edit.Text)
		s.valuesIndex = -1
		for i, v := range s.Values {
			if v == cur {
				s.valuesIndex = i
				break
			}
		}
		s.valuesIndex--
		if s.valuesIndex < 0 {
			if s.Wrap {
				s.valuesIndex = len(s.Values) - 1
			} else {
				s.valuesIndex = 0
			}
		}
		s.Set(s.Values[s.valuesIndex])
	} else {
		val := s.currentNumericValue()
		val -= s.Increment
		lo, hi := s.From, s.To
		if lo > hi {
			lo, hi = hi, lo
		}
		if val < lo {
			if s.Wrap {
				val = hi
			} else {
				val = lo
			}
		}
		s.Set(s.formatValue(val))
	}
	s.selectAll()
	s.Display()
	s.fireCommand()
}

func (s *Spinbox) currentNumericValue() float64 {
	val, err := strconv.ParseFloat(string(s.edit.Text), 64)
	if err != nil {
		return s.From
	}
	return val
}

func (s *Spinbox) formatValue(v float64) string {
	if s.Increment >= 1 {
		return fmt.Sprintf("%.0f", math.Round(v))
	}
	return fmt.Sprintf(s.Format, v)
}

func (s *Spinbox) fireCommand() {
	if s.Command != nil {
		s.Command(string(s.edit.Text))
	}
}

func (s *Spinbox) selectAll() {
	s.edit.SelectAll()
}

func (s *Spinbox) tryValidate(_ entrytext.ValidateReason, prospective string) bool {
	if s.ValidateCmd == nil {
		return true
	}
	v := s.Validate
	if v != "key" && v != "all" {
		return true
	}
	return s.ValidateCmd(prospective)
}

// fieldPad ports FieldElementSize: -borderwidth widened to -focuswidth.
func (s *Spinbox) fieldPad() int {
	bw := LookupInt(s.Context.Style, "-borderwidth", s.State, 2)
	fw := LookupInt(s.Context.Style, "-focuswidth", s.State, 2)
	if fw > 0 && bw < 2 {
		bw = fw
	}
	return bw
}

// arrowSize ports BoxArrowElementSize for the up/down arrows.
func (s *Spinbox) arrowSize() (int, int) {
	pad := 3 * screenunit.ScalingPct() / 100
	size := LookupInt(s.Context.Style, "-arrowsize", s.State, 14) - 2*pad +
		2*((screenunit.ScalingPct()+50)/100)
	return 2*(size/2) + 1 + 2*pad, size/2 + 1 + 2*pad
}

// computeGeometry ports the SpinboxLayout size: the field around the
// right-packed arrow pair and the -padding'd textarea of -width "0"s.
func (s *Spinbox) computeGeometry() {
	if s.Font == nil || s.Context == nil {
		return
	}
	m := s.Font.Metrics()
	s.avgWidth = max(1, s.Font.MeasureString("0"))
	fp := s.fieldPad()
	aw, ah := s.arrowSize()
	p := LookupPadding(s.Context.Style, "-padding", s.State, Padding{})
	s.buttonWidth = aw
	s.insetX = fp + p.Left
	s.edit.TextX = s.insetX
	win := s.Win
	win.ReqWidth = 2*fp + aw + p.Left + p.Right + s.prefWidth*s.avgWidth
	win.ReqHeight = 2*fp + max(2*ah, m.Linespace()+p.Top+p.Bottom)
}

// arrowBoxes returns the up and down arrow parcels: the "null" group is
// packed right and centred vertically, up at its top, down at its bottom.
func (s *Spinbox) arrowBoxes() (Box, Box) {
	fp := s.fieldPad()
	aw, ah := s.arrowSize()
	x := s.Win.Width - fp - aw
	y := fp + (s.Win.Height-2*fp-2*ah)/2
	return Box{x, y, aw, ah}, Box{x, y + ah, aw, ah}
}

// hitButton returns "up", "down", or "" based on click position.
func (s *Spinbox) hitButton(x, y int) string {
	up, down := s.arrowBoxes()
	in := func(b Box) bool { return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height }
	switch {
	case in(up):
		return "up"
	case in(down):
		return "down"
	}
	return ""
}

func (s *Spinbox) closestGap(x int) int { return s.edit.ClosestGap(x) }

// Display draws the themed spinbox.
func (s *Spinbox) Display() {
	if s.Destroyed {
		return
	}
	win := s.Win
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

	// Double buffer.
	if s.pixmap == 0 || s.pixmapW != width || s.pixmapH != height {
		if s.pixmap != 0 {
			d.FreePixmap(s.pixmap)
		}
		s.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		s.pixmapW = width
		s.pixmapH = height
	}
	if s.pixmap == 0 {
		return
	}

	pixDrawable := platform.PixmapDrawable(s.pixmap)

	st := s.Context.Style
	bg := LookupColor(st, "-background", s.State, 0xd9d9d9)
	fg := LookupColor(st, "-foreground", s.State, 0x000000)

	// Spinbox.field (FieldElementDraw, ttkElements.c).
	fieldBg := LookupColor(st, "-fieldbackground", s.State, 0xffffff)
	draw.Fill3DRectangle(d, pixDrawable, gc, draw.NewBorderFromPixel(fieldBg),
		0, 0, width, height, LookupInt(st, "-borderwidth", s.State, 2), option.ReliefSunken)
	if LookupInt(st, "-focuswidth", s.State, 2) > 0 && s.hasFocus {
		d.SetForeground(gc, LookupColor(st, "-focuscolor", s.State, 0x4a6984))
		d.DrawRectangle(pixDrawable, gc, 0, 0, uint(width-1), uint(height-1))
	}

	// Spinbox.uparrow / Spinbox.downarrow (BoxArrowElementDraw).
	up, down := s.arrowBoxes()
	btnLeft := up.X
	border := draw.NewBorderFromPixel(bg)
	arrowColor := LookupColor(st, "-arrowcolor", s.State, 0x000000)
	pad := 3 * screenunit.ScalingPct() / 100
	for i, ab := range []Box{up, down} {
		relief := LookupRelief(st, "-relief", s.State, option.ReliefRaised)
		if (i == 0 && s.pressedButton == "up") || (i == 1 && s.pressedButton == "down") {
			relief = option.ReliefSunken
		}
		draw.Fill3DRectangle(d, pixDrawable, gc, border, ab.X, ab.Y, ab.Width, ab.Height,
			LookupInt(st, "-borderwidth", s.State, 1), relief)
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(pixDrawable, gc, ab.X, ab.Y+1, ab.X, ab.Y+ab.Height-1)
		ib := Box{ab.X + pad, ab.Y + pad, ab.Width - 2*pad, ab.Height - 2*pad}
		cx, cy := 2*(ib.Width/2)+1, ib.Width/2+1
		if (ib.Height-cy)%2 == 1 {
			cy++
		}
		b := Box{ib.X + (ib.Width-cx)/2, ib.Y + (ib.Height-cy)/2, cx, cy}
		pts := arrowDownPoints(b)
		if i == 0 {
			pts = arrowUpPoints(b)
		}
		d.SetForeground(gc, arrowColor)
		d.FillPolygon(pixDrawable, gc, pts, 2, 0)
		d.DrawLines(pixDrawable, gc, append(pts, pts[0]), 0)
		d.DrawLine(pixDrawable, gc, int(pts[2].X), int(pts[2].Y), int(pts[2].X), int(pts[2].Y))
	}

	// Draw text.
	if s.Font != nil && len(s.edit.Text) > 0 {
		m := s.Font.Metrics()
		textX := s.insetX
		textY := (height-m.Linespace())/2 + m.Ascent

		// Selection highlight.
		if s.hasFocus && s.edit.HasSelection() {
			selStartX := textX + s.Font.MeasureString(string(s.edit.Text[:sbClamp(s.edit.SelFirst, len(s.edit.Text))]))
			selEndX := textX + s.Font.MeasureString(string(s.edit.Text[:sbClamp(s.edit.SelLast, len(s.edit.Text))]))
			rightEdge := btnLeft - s.insetX
			if selStartX < textX {
				selStartX = textX
			}
			if selEndX > rightEdge {
				selEndX = rightEdge
			}
			if selEndX > selStartX {
				d.SetForeground(gc, uint64(0x4a6984))
				d.FillRectangle(pixDrawable, gc, selStartX, (height-m.Linespace())/2,
					uint(selEndX-selStartX), uint(m.Linespace()))
			}
		}

		if df, ok := s.Font.(platform.DrawableFont); ok {
			r := uint16((fg>>16)&0xFF) * 257
			g := uint16((fg>>8)&0xFF) * 257
			b := uint16((fg)&0xFF) * 257
			df.DrawString(pixDrawable, textX, textY, string(s.edit.Text), fg, r, g, b)
		}

		// Insert cursor.
		if s.hasFocus && s.cursorOn {
			cursorX := textX + s.Font.MeasureString(string(s.edit.Text[:sbClamp(s.edit.InsertPos, len(s.edit.Text))]))
			rightEdge := btnLeft - s.insetX
			if cursorX >= textX && cursorX < rightEdge {
				d.SetForeground(gc, uint64(0x000000))
				d.FillRectangle(pixDrawable, gc, cursorX, (height-m.Linespace())/2, 2, uint(m.Linespace()))
			}
		}
	}

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

func sbClamp(idx, max int) int {
	if idx < 0 {
		return 0
	}
	if idx > max {
		return max
	}
	return idx
}

func bindSpinbox(s *Spinbox, app widget.AppContext) {
	win := s.Win

	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		s.ChangeState(StateHover|StateActive, 0)
	})

	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		s.ChangeState(0, StateHover|StateActive)
	})

	// Button press.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if s.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			widget.Focus(app, win)

			btn := s.hitButton(ev.X, ev.Y)
			if btn == "up" {
				s.pressedButton = "up"
				s.SpinUp()
			} else if btn == "down" {
				s.pressedButton = "down"
				s.SpinDown()
			} else {
				// Click in text area — position cursor.
				s.edit.SelAnchor = s.closestGap(ev.X)
				s.edit.InsertPos = s.edit.SelAnchor
				s.edit.ClearSelection()
				s.Display()
			}
		} else if ev.Button == 4 {
			s.SpinUp()
		} else if ev.Button == 5 {
			s.SpinDown()
		}
	})

	// Button release.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			s.pressedButton = ""
			s.Display()
		}
	})

	// Motion — drag select in text area.
	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask != 0 {
			if s.hitButton(ev.X, ev.Y) != "" {
				return
			}
			s.edit.MoveCursor(s.closestGap(ev.X), s.edit.SelAnchor, true)
		}
	})

	// Focus.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			s.hasFocus = true
			s.cursorOn = true
			s.ChangeState(StateFocus, 0)
		} else if ev.Type == event.FocusOutType {
			s.hasFocus = false
			s.ChangeState(0, StateFocus)
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if s.State&StateDisabled != 0 {
			return
		}

		switch ev.KeySym {
		case platform.XK_Up:
			s.SpinUp()
		case platform.XK_Down:
			s.SpinDown()
		case platform.XK_Left, platform.XK_Right, platform.XK_Home, platform.XK_End:
			s.edit.HandleNavKey(ev)
		case platform.XK_BackSpace, platform.XK_Delete, platform.XK_Insert:
			s.edit.HandleEditKey(ev)
		default:
			if ev.State&platform.ControlMask != 0 {
				s.edit.HandleCtrlKey(ev)
				return
			}
			s.edit.HandleKey(ev)
		}
	})
	app.Dispatcher().Bind(win.PlatformID, event.VirtualMask, func(ev *event.Event) {
		if s.State&StateDisabled == 0 {
			s.edit.HandleVirtual(ev)
		}
	})
}
