// Entry event bindings, porting tk/library/entry.tcl.
package entry

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
)

func bindEntry(e *Entry, app widget.AppContext) {
	w := e.Win

	// Expose.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		e.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			e.computeGeometry()
			e.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.XWindow, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			e.HasFocus = true
			e.CursorOn = true
			e.Display()
		} else if ev.Type == event.FocusOutType {
			e.HasFocus = false
			e.Display()
		}
	})

	// Mouse: click to position cursor and take focus.
	app.Dispatcher().Bind(w.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			// Request X11 input focus so key events come to this window.
			app.DisplayPtr().SetInputFocus(w.XWindow, xlib.RevertToParent, xlib.CurrentTime)
			e.ClearSelection()
			e.InsertPos = e.closestGap(ev.X)
			e.SelAnchor = e.InsertPos
			e.Display()
		}
	})

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.XWindow, event.MotionMask, func(ev *event.Event) {
		if ev.State&xlib.Button1Mask != 0 {
			pos := e.closestGap(ev.X)
			if pos < e.SelAnchor {
				e.SelFirst = pos
				e.SelLast = e.SelAnchor
			} else {
				e.SelFirst = e.SelAnchor
				e.SelLast = pos
			}
			e.InsertPos = pos
			e.seeInsert()
			e.Display()
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(w.XWindow, event.KeyPressMask, func(ev *event.Event) {
		shift := ev.State&xlib.ShiftMask != 0
		ctrl := ev.State&xlib.ControlMask != 0

		switch ev.KeySym {
		case xlib.XK_Left:
			if ctrl {
				newPos := wordStart(e.text, e.InsertPos)
				moveCursor(e, newPos, shift)
			} else {
				moveCursor(e, e.InsertPos-1, shift)
			}

		case xlib.XK_Right:
			if ctrl {
				newPos := wordEnd(e.text, e.InsertPos)
				moveCursor(e, newPos, shift)
			} else {
				moveCursor(e, e.InsertPos+1, shift)
			}

		case xlib.XK_Home:
			moveCursor(e, 0, shift)

		case xlib.XK_End:
			moveCursor(e, len(e.text), shift)

		case xlib.XK_BackSpace:
			if e.SelFirst >= 0 {
				e.DeleteSelection()
			} else if e.InsertPos > 0 {
				e.DeleteChars(e.InsertPos-1, 1)
			}

		case xlib.XK_Delete:
			if e.SelFirst >= 0 {
				e.DeleteSelection()
			} else if e.InsertPos < len(e.text) {
				e.DeleteChars(e.InsertPos, 1)
			}

		default:
			if ctrl {
				handleCtrlKey(e, ev)
				return
			}
			// Insert printable characters.
			// First try ev.Str (from XLookupString), then fall back
			// to keysym-to-unicode conversion for non-Latin layouts.
			insertStr := ev.Str
			if insertStr == "" {
				if r := xlib.KeySymToRune(ev.KeySym); r > 0 {
					insertStr = string(r)
				}
			}
			if insertStr != "" && insertStr[0] >= 32 {
				if e.SelFirst >= 0 {
					e.DeleteSelection()
				}
				e.InsertChars(e.InsertPos, insertStr)
			}
		}
	})
}

// moveCursor moves the cursor, optionally extending selection.
func moveCursor(e *Entry, newPos int, shift bool) {
	if newPos < 0 {
		newPos = 0
	}
	if newPos > len(e.text) {
		newPos = len(e.text)
	}

	if shift {
		// Extend selection.
		if e.SelFirst < 0 {
			e.SelAnchor = e.InsertPos
		}
		if newPos < e.SelAnchor {
			e.SelFirst = newPos
			e.SelLast = e.SelAnchor
		} else {
			e.SelFirst = e.SelAnchor
			e.SelLast = newPos
		}
		if e.SelFirst == e.SelLast {
			e.ClearSelection()
		}
	} else {
		e.ClearSelection()
	}

	e.InsertPos = newPos
	e.seeInsert()
	e.Display()
}

// handleCtrlKey handles control key combinations.
func handleCtrlKey(e *Entry, ev *event.Event) {
	switch ev.KeySym {
	case xlib.XK_a:
		// Select all.
		e.SelectAll()
		e.Display()

	case xlib.XK_k:
		// Kill to end of line.
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, len(e.text)-e.InsertPos)
		}

	case xlib.XK_d:
		// Delete forward.
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, 1)
		}
	}
}
