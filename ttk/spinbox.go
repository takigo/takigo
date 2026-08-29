package ttk

import (
	"fmt"
	"math"
	"strconv"

	"github.com/msorc/takigo/cursor"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Spinbox is a themed spinbox widget — an entry with up/down buttons.
type Spinbox struct {
	TtkWidget

	// Text state.
	text      []rune
	insertPos int
	selFirst  int
	selLast   int
	selAnchor int
	leftIndex int

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
		selFirst:    -1,
		selLast:     -1,
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

	InitTtkWidget(&s.TtkWidget, win, app, "TSpinbox")
	s.DisplayFunc = s.Display

	for _, opt := range opts {
		opt(s)
	}

	// Set initial value.
	if len(s.Values) > 0 {
		s.text = []rune(s.Values[0])
	} else {
		s.text = []rune(s.formatValue(s.From))
	}

	s.computeGeometry()

	win.Flags |= window.FlagFocusable
	win.SetCursor(uint(cursor.XTerm))
	bindSpinbox(s, app)

	return s
}

// Get returns the current text value.
func (s *Spinbox) Get() string {
	return string(s.text)
}

// Set sets the text value.
func (s *Spinbox) Set(text string) {
	s.text = []rune(text)
	if s.insertPos > len(s.text) {
		s.insertPos = len(s.text)
	}
	s.selFirst = -1
	s.selLast = -1
	s.computeGeometry()
	s.Display()
}

// SpinUp increments the value.
func (s *Spinbox) SpinUp() {
	if len(s.Values) > 0 {
		// Find current value in list.
		cur := string(s.text)
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
		cur := string(s.text)
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
	val, err := strconv.ParseFloat(string(s.text), 64)
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
		s.Command(string(s.text))
	}
}

func (s *Spinbox) selectAll() {
	if len(s.text) > 0 {
		s.selFirst = 0
		s.selLast = len(s.text)
	}
}

func (s *Spinbox) tryEdit(prospective string) bool {
	if s.ValidateCmd == nil {
		return true
	}
	v := s.Validate
	if v != "key" && v != "all" {
		return true
	}
	return s.ValidateCmd(prospective)
}

func (s *Spinbox) computeGeometry() {
	if s.Font == nil {
		return
	}
	m := s.Font.Metrics()
	s.avgWidth = s.Font.MeasureString("0")
	if s.avgWidth < 1 {
		s.avgWidth = 1
	}

	win := s.Win
	win.ReqWidth = s.prefWidth*s.avgWidth + 2*s.insetX + s.buttonWidth + 4
	win.ReqHeight = m.Linespace() + 2*s.insetY + 4
}

// hitButton returns "up", "down", or "" based on click position.
func (s *Spinbox) hitButton(x, y int) string {
	btnLeft := s.Win.Width - s.buttonWidth
	if x < btnLeft {
		return ""
	}
	midY := s.Win.Height / 2
	if y < midY {
		return "up"
	}
	return "down"
}

func (s *Spinbox) closestGap(x int) int {
	if s.Font == nil || len(s.text) == 0 {
		return 0
	}
	textX := s.insetX + 2
	xInText := x - textX
	if xInText <= 0 {
		return 0
	}
	// Binary search for closest character boundary.
	lo, hi := 0, len(s.text)
	for lo < hi {
		mid := (lo + hi) / 2
		w := s.Font.MeasureString(string(s.text[:mid+1]))
		if w <= xInText {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

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

	bg := LookupColor(s.Context.Style, "-background", s.State, 0xd9d9d9)
	fg := LookupColor(s.Context.Style, "-foreground", s.State, 0x000000)

	// Entry field background (white for editable).
	fieldBg := uint64(0xffffff)
	if s.State&StateDisabled != 0 {
		fieldBg = bg
	}

	// Fill entry area.
	d.SetForeground(gc, fieldBg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Draw border.
	d.SetForeground(gc, uint64(0x9e9a91))
	d.DrawLine(pixDrawable, gc, 0, 0, width-1, 0)
	d.DrawLine(pixDrawable, gc, 0, 0, 0, height-1)
	border := draw.NewBorderFromPixel(bg)
	d.SetForeground(gc, border.LightPixel)
	d.DrawLine(pixDrawable, gc, 0, height-1, width-1, height-1)
	d.DrawLine(pixDrawable, gc, width-1, 0, width-1, height-1)

	// Button area.
	btnLeft := width - s.buttonWidth
	midY := height / 2

	// Button background.
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, btnLeft, 1, uint(s.buttonWidth-1), uint(height-2))

	// Separator line.
	d.SetForeground(gc, uint64(0x9e9a91))
	d.DrawLine(pixDrawable, gc, btnLeft, 1, btnLeft, height-2)

	// Horizontal divider between up and down.
	d.DrawLine(pixDrawable, gc, btnLeft+1, midY, width-2, midY)

	// Up/down button relief.
	if s.pressedButton == "up" {
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(pixDrawable, gc, btnLeft+1, 1, width-2, 1)
		d.DrawLine(pixDrawable, gc, btnLeft+1, 1, btnLeft+1, midY-1)
	} else if s.pressedButton == "down" {
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(pixDrawable, gc, btnLeft+1, midY+1, width-2, midY+1)
		d.DrawLine(pixDrawable, gc, btnLeft+1, midY+1, btnLeft+1, height-2)
	}

	// Draw arrows.
	d.SetForeground(gc, fg)
	cx := btnLeft + s.buttonWidth/2
	// Up arrow.
	upCy := midY / 2
	for row := 0; row < 3; row++ {
		d.DrawLine(pixDrawable, gc, cx-row, upCy-1+row, cx+row, upCy-1+row)
	}
	// Down arrow.
	downCy := midY + (height-midY)/2
	for row := 0; row < 3; row++ {
		d.DrawLine(pixDrawable, gc, cx-row, downCy+1-row, cx+row, downCy+1-row)
	}

	// Draw text.
	if s.Font != nil && len(s.text) > 0 {
		m := s.Font.Metrics()
		textX := s.insetX + 2
		textY := (height-m.Linespace())/2 + m.Ascent

		// Selection highlight.
		if s.hasFocus && s.selFirst >= 0 && s.selLast > s.selFirst {
			selStartX := textX + s.Font.MeasureString(string(s.text[:sbClamp(s.selFirst, len(s.text))]))
			selEndX := textX + s.Font.MeasureString(string(s.text[:sbClamp(s.selLast, len(s.text))]))
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
			r := uint16((fg>>16)&0xFF) << 8
			g := uint16((fg>>8)&0xFF) << 8
			b := uint16((fg)&0xFF) << 8
			df.DrawString(pixDrawable, textX, textY, string(s.text), fg, r, g, b)
		}

		// Insert cursor.
		if s.hasFocus && s.cursorOn {
			cursorX := textX + s.Font.MeasureString(string(s.text[:sbClamp(s.insertPos, len(s.text))]))
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

	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			s.Display()
		}
	})

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
			app.Server().SetInputFocus(win.PlatformID, platform.RevertToParent, platform.CurrentTime)

			btn := s.hitButton(ev.X, ev.Y)
			if btn == "up" {
				s.pressedButton = "up"
				s.SpinUp()
			} else if btn == "down" {
				s.pressedButton = "down"
				s.SpinDown()
			} else {
				// Click in text area — position cursor.
				s.selFirst = -1
				s.selLast = -1
				s.insertPos = s.closestGap(ev.X)
				s.selAnchor = s.insertPos
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
			pos := s.closestGap(ev.X)
			if pos < s.selAnchor {
				s.selFirst = pos
				s.selLast = s.selAnchor
			} else {
				s.selFirst = s.selAnchor
				s.selLast = pos
			}
			s.insertPos = pos
			s.Display()
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
		shift := ev.State&platform.ShiftMask != 0
		ctrl := ev.State&platform.ControlMask != 0

		switch ev.KeySym {
		case platform.XK_Up:
			s.SpinUp()
		case platform.XK_Down:
			s.SpinDown()

		case platform.XK_Left:
			if ctrl {
				sbMoveCursor(s, sbWordStart(s.text, s.insertPos), shift)
			} else {
				sbMoveCursor(s, s.insertPos-1, shift)
			}
		case platform.XK_Right:
			if ctrl {
				sbMoveCursor(s, sbWordEnd(s.text, s.insertPos), shift)
			} else {
				sbMoveCursor(s, s.insertPos+1, shift)
			}
		case platform.XK_Home:
			sbMoveCursor(s, 0, shift)
		case platform.XK_End:
			sbMoveCursor(s, len(s.text), shift)

		case platform.XK_BackSpace:
			if s.selFirst >= 0 {
				prospective := string(s.text[:s.selFirst]) + string(s.text[s.selLast:])
				if s.tryEdit(prospective) {
					sbDeleteSelection(s)
				}
			} else if s.insertPos > 0 {
				prospective := string(s.text[:s.insertPos-1]) + string(s.text[s.insertPos:])
				if s.tryEdit(prospective) {
					s.text = append(s.text[:s.insertPos-1], s.text[s.insertPos:]...)
					s.insertPos--
					s.Display()
				}
			}
		case platform.XK_Delete:
			if s.selFirst >= 0 {
				prospective := string(s.text[:s.selFirst]) + string(s.text[s.selLast:])
				if s.tryEdit(prospective) {
					sbDeleteSelection(s)
				}
			} else if s.insertPos < len(s.text) {
				prospective := string(s.text[:s.insertPos]) + string(s.text[s.insertPos+1:])
				if s.tryEdit(prospective) {
					s.text = append(s.text[:s.insertPos], s.text[s.insertPos+1:]...)
					s.Display()
				}
			}

		default:
			if ctrl {
				if ev.KeySym == platform.KeySym(0x0061) { // XK_a
					s.selectAll()
					s.Display()
				}
				return
			}
			insertStr := ev.Str
			if insertStr == "" {
				if r := platform.KeySymToRune(ev.KeySym); r > 0 {
					insertStr = string(r)
				}
			}
			if insertStr != "" && insertStr[0] >= 32 {
				var prospective string
				if s.selFirst >= 0 {
					prospective = string(s.text[:s.selFirst]) + insertStr + string(s.text[s.selLast:])
				} else {
					prospective = string(s.text[:s.insertPos]) + insertStr + string(s.text[s.insertPos:])
				}
				if s.tryEdit(prospective) {
					if s.selFirst >= 0 {
						sbDeleteSelection(s)
					}
					runes := []rune(insertStr)
					newText := make([]rune, 0, len(s.text)+len(runes))
					newText = append(newText, s.text[:s.insertPos]...)
					newText = append(newText, runes...)
					newText = append(newText, s.text[s.insertPos:]...)
					s.text = newText
					s.insertPos += len(runes)
					s.Display()
				}
			}
		}
	})
}

func sbDeleteSelection(s *Spinbox) {
	if s.selFirst < 0 {
		return
	}
	s.text = append(s.text[:s.selFirst], s.text[s.selLast:]...)
	s.insertPos = s.selFirst
	s.selFirst = -1
	s.selLast = -1
	s.Display()
}

func sbMoveCursor(s *Spinbox, newPos int, shift bool) {
	if newPos < 0 {
		newPos = 0
	}
	if newPos > len(s.text) {
		newPos = len(s.text)
	}

	if shift {
		if s.selFirst < 0 {
			s.selAnchor = s.insertPos
		}
		if newPos < s.selAnchor {
			s.selFirst = newPos
			s.selLast = s.selAnchor
		} else {
			s.selFirst = s.selAnchor
			s.selLast = newPos
		}
		if s.selFirst == s.selLast {
			s.selFirst = -1
			s.selLast = -1
		}
	} else {
		s.selFirst = -1
		s.selLast = -1
	}

	s.insertPos = newPos
	s.Display()
}

func sbWordStart(text []rune, pos int) int {
	if pos <= 0 {
		return 0
	}
	if pos > len(text) {
		pos = len(text)
	}
	i := pos - 1
	for i > 0 && !sbIsWordChar(text[i]) {
		i--
	}
	for i > 0 && sbIsWordChar(text[i-1]) {
		i--
	}
	return i
}

func sbWordEnd(text []rune, pos int) int {
	if pos >= len(text) {
		return len(text)
	}
	i := pos
	for i < len(text) && sbIsWordChar(text[i]) {
		i++
	}
	for i < len(text) && !sbIsWordChar(text[i]) {
		i++
	}
	return i
}

func sbIsWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}
