package spinbox

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
)

func bindSpinbox(s *Spinbox, app widget.AppContext) {
	w := s.Win

	// Expose.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			s.computeGeometry()
			s.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.XWindow, event.FocusChangeMask, func(ev *event.Event) {
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
	app.Dispatcher().Bind(w.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			app.DisplayPtr().SetInputFocus(w.XWindow, xlib.RevertToParent, xlib.CurrentTime)

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
				// Click in text area — position cursor.
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
	app.Dispatcher().Bind(w.XWindow, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			s.pressedButton = ""
			s.Display()
		}
	})

	// Mouse drag to select.
	app.Dispatcher().Bind(w.XWindow, event.MotionMask, func(ev *event.Event) {
		if ev.State&xlib.Button1Mask != 0 {
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
	app.Dispatcher().Bind(w.XWindow, event.KeyPressMask, func(ev *event.Event) {
		shift := ev.State&xlib.ShiftMask != 0
		ctrl := ev.State&xlib.ControlMask != 0

		switch ev.KeySym {
		case xlib.XK_Up:
			s.SpinUp()
		case xlib.XK_Down:
			s.SpinDown()

		case xlib.XK_Left:
			if ctrl {
				moveCursor(s, wordStart(s.text, s.InsertPos), shift)
			} else {
				moveCursor(s, s.InsertPos-1, shift)
			}
		case xlib.XK_Right:
			if ctrl {
				moveCursor(s, wordEnd(s.text, s.InsertPos), shift)
			} else {
				moveCursor(s, s.InsertPos+1, shift)
			}
		case xlib.XK_Home:
			moveCursor(s, 0, shift)
		case xlib.XK_End:
			moveCursor(s, len(s.text), shift)

		case xlib.XK_BackSpace:
			if s.SelFirst >= 0 {
				s.DeleteSelection()
			} else if s.InsertPos > 0 {
				s.DeleteChars(s.InsertPos-1, 1)
			}
		case xlib.XK_Delete:
			if s.SelFirst >= 0 {
				s.DeleteSelection()
			} else if s.InsertPos < len(s.text) {
				s.DeleteChars(s.InsertPos, 1)
			}

		default:
			if ctrl {
				switch ev.KeySym {
				case xlib.XK_a:
					s.SelectAll()
					s.Display()
				}
				return
			}
			// Insert printable characters.
			insertStr := ev.Str
			if insertStr == "" {
				if r := xlib.KeySymToRune(ev.KeySym); r > 0 {
					insertStr = string(r)
				}
			}
			if insertStr != "" && insertStr[0] >= 32 {
				if s.SelFirst >= 0 {
					s.DeleteSelection()
				}
				s.InsertChars(s.InsertPos, insertStr)
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
