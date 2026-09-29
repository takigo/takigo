package spinbox

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entryutil"
)

// bindSpinbox registers all event handlers for the spinbox widget.
func bindSpinbox(s *Spinbox, app widget.AppContext) {
	w := s.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, s.handleExpose)

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, s.handleConfigure)

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, s.handleFocus)

	// Mouse: click to position cursor or press buttons.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, s.handleButtonPress)

	// Mouse release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, s.handleButtonRelease)

	// Mouse drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, s.handleMotion)

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, s.handleKeyPress)
	app.Dispatcher().Bind(w.PlatformID, event.VirtualMask, s.handleVirtual)
}

// handleVirtual handles the input method's virtual events as the entry
// does. library/spinbox.tcl has no such bindings, which leaves every
// intermediate composition behind in a Tk spinbox.
func (s *Spinbox) handleVirtual(ev *event.Event) {
	switch ev.Name {
	case event.IMEStart:
		s.imeMark = s.InsertPos
		return
	case event.IMEEnd:
		if s.imeMark < s.InsertPos && s.InsertPos <= len(s.text) {
			s.SelFirst, s.SelLast, s.SelAnchor = s.imeMark, s.InsertPos, s.imeMark
		}
	case event.IMEClear:
		first := min(s.imeMark, len(s.text))
		if first < s.InsertPos {
			if !s.tryEdit(string(s.text[:first]) + string(s.text[s.InsertPos:])) {
				return
			}
			s.DeleteChars(first, s.InsertPos-first)
		}
	case event.AccentBackspace:
		s.backspace()
	default:
		return
	}
	s.seeInsert()
	s.Display()
}

// backspace deletes the selection, or the character before the insertion
// cursor (tk::EntryBackspace).
func (s *Spinbox) backspace() {
	if s.SelFirst >= 0 {
		prospective := string(s.text[:s.SelFirst]) + string(s.text[s.SelLast:])
		if s.tryEdit(prospective) {
			s.DeleteSelection()
		}
	} else if s.InsertPos > 0 {
		prospective := string(s.text[:s.InsertPos-1]) + string(s.text[s.InsertPos:])
		if s.tryEdit(prospective) {
			s.DeleteChars(s.InsertPos-1, 1)
		}
	}
}

// handleExpose handles Exposure events.
func (s *Spinbox) handleExpose(ev *event.Event) {
	if ev.ExposeCount > 0 {
		return
	}
	s.Display()
}

// handleConfigure handles ConfigureNotify (resize) events.
func (s *Spinbox) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		s.Win.Width = ev.ConfigWidth
		s.Win.Height = ev.ConfigHeight
		s.computeGeometry()
		s.Display()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (s *Spinbox) handleFocus(ev *event.Event) {
	if ev.Type == event.FocusInType {
		s.HasFocus = true
		s.CursorOn = true
		s.Display()
	} else if ev.Type == event.FocusOutType {
		s.HasFocus = false
		s.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (s *Spinbox) handleButtonPress(ev *event.Event) {
	if ev.Button == 1 {
		app := s.App
		app.Server().SetInputFocus(s.Win.PlatformID, platform.RevertToParent, platform.CurrentTime)

		btn := s.hitButton(ev.X, ev.Y)
		if btn == "up" {
			s.pressedButton = "up"
			s.SpinUp()
			s.Display()
		} else if btn == "down" {
			s.pressedButton = "down"
			s.SpinDown()
			s.Display()
		} else {
			// Click in text area -- position cursor.
			s.ClearSelection()
			s.InsertPos = s.closestGap(ev.X)
			s.SelAnchor = s.InsertPos
			s.Display()
		}
	} else if ev.Button == 4 {
		// Mouse wheel up.
		s.SpinUp()
	} else if ev.Button == 5 {
		// Mouse wheel down.
		s.SpinDown()
	}
}

// handleButtonRelease handles mouse button release events.
func (s *Spinbox) handleButtonRelease(ev *event.Event) {
	if ev.Button == 1 {
		s.pressedButton = ""
		s.Display()
	}
}

// handleMotion handles mouse motion events (drag to select).
func (s *Spinbox) handleMotion(ev *event.Event) {
	if ev.State&platform.Button1Mask != 0 {
		// Only drag-select in text area.
		if s.hitButton(ev.X, ev.Y) != "" {
			return
		}
		pos := s.closestGap(ev.X)
		if pos < s.SelAnchor {
			s.SelFirst = pos
			s.SelLast = s.SelAnchor
		} else {
			s.SelFirst = s.SelAnchor
			s.SelLast = pos
		}
		s.InsertPos = pos
		s.seeInsert()
		s.Display()
	}
}

// handleKeyPress handles keyboard events.
func (s *Spinbox) handleKeyPress(ev *event.Event) {
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&(platform.ControlMask|platform.CommandMask) != 0

	switch ev.KeySym {
	case platform.XK_Up:
		s.SpinUp()
	case platform.XK_Down:
		s.SpinDown()

	case platform.XK_Left:
		if ctrl {
			moveCursor(s, entryutil.WordStart(s.text, s.InsertPos), shift)
		} else {
			moveCursor(s, s.InsertPos-1, shift)
		}
	case platform.XK_Right:
		if ctrl {
			moveCursor(s, entryutil.WordEnd(s.text, s.InsertPos), shift)
		} else {
			moveCursor(s, s.InsertPos+1, shift)
		}
	case platform.XK_Home:
		moveCursor(s, 0, shift)
	case platform.XK_End:
		moveCursor(s, len(s.text), shift)

	case platform.XK_BackSpace:
		s.backspace()
	case platform.XK_Delete:
		if s.SelFirst >= 0 {
			prospective := string(s.text[:s.SelFirst]) + string(s.text[s.SelLast:])
			if s.tryEdit(prospective) {
				s.DeleteSelection()
			}
		} else if s.InsertPos < len(s.text) {
			prospective := string(s.text[:s.InsertPos]) + string(s.text[s.InsertPos+1:])
			if s.tryEdit(prospective) {
				s.DeleteChars(s.InsertPos, 1)
			}
		}

	default:
		if ctrl {
			switch ev.KeySym {
			case platform.KeySym(0x0061): // XK_a
				s.SelectAll()
				s.Display()
			}
			return
		}
		// Insert printable characters.
		insertStr := ev.Str
		if insertStr == "" {
			if r := platform.KeySymToRune(ev.KeySym); r > 0 {
				insertStr = string(r)
			}
		}
		if insertStr != "" && insertStr[0] >= 32 {
			var prospective string
			if s.SelFirst >= 0 {
				prospective = string(s.text[:s.SelFirst]) + insertStr + string(s.text[s.SelLast:])
			} else {
				prospective = string(s.text[:s.InsertPos]) + insertStr + string(s.text[s.InsertPos:])
			}
			if s.tryEdit(prospective) {
				if s.SelFirst >= 0 {
					s.DeleteSelection()
				}
				s.InsertChars(s.InsertPos, insertStr)
			}
		}
	}
}

// moveCursor moves the cursor, optionally extending selection.
func moveCursor(s *Spinbox, newPos int, shift bool) {
	if newPos < 0 {
		newPos = 0
	}
	if newPos > len(s.text) {
		newPos = len(s.text)
	}

	if shift {
		if s.SelFirst < 0 {
			s.SelAnchor = s.InsertPos
		}
		if newPos < s.SelAnchor {
			s.SelFirst = newPos
			s.SelLast = s.SelAnchor
		} else {
			s.SelFirst = s.SelAnchor
			s.SelLast = newPos
		}
		if s.SelFirst == s.SelLast {
			s.ClearSelection()
		}
	} else {
		s.ClearSelection()
	}

	s.InsertPos = newPos
	s.seeInsert()
	s.Display()
}
