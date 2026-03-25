package spinbox

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entryutil"
)

func bindSpinbox(s *Spinbox, app widget.AppContext) {
	w := s.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			s.computeGeometry()
			s.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			s.HasFocus = true
			s.CursorOn = true
			s.Display()
		} else if ev.Type == event.FocusOutType {
			s.HasFocus = false
			s.Display()
		}
	})

	// Mouse: click to position cursor or press buttons.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			app.Server().SetInputFocus(w.PlatformID, platform.RevertToParent, platform.CurrentTime)

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
	})

	// Mouse release.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			s.pressedButton = ""
			s.Display()
		}
	})

	// Mouse drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
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
	})

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		shift := ev.State&platform.ShiftMask != 0
		ctrl := ev.State&(platform.ControlMask|platform.Mod2Mask) != 0 // Ctrl or Cmd (macOS)

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
	})
}

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
