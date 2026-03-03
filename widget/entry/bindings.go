// Entry event bindings, porting tk/library/entry.tcl.
package entry

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

func bindEntry(e *Entry, app widget.AppContext) {
	w := e.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		e.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			e.computeGeometry()
			e.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
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
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			// Request X11 input focus so key events come to this window.
			app.Server().SetInputFocus(w.PlatformID, platform.RevertToParent, platform.CurrentTime)
			e.ClearSelection()
			e.InsertPos = e.closestGap(ev.X)
			e.SelAnchor = e.InsertPos
			e.Display()
		}
	})

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if ev.State&platform.Button1Mask != 0 {
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
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		shift := ev.State&platform.ShiftMask != 0
		ctrl := ev.State&platform.ControlMask != 0

		switch ev.KeySym {
		case platform.XK_Left:
			if ctrl {
				newPos := wordStart(e.text, e.InsertPos)
				moveCursor(e, newPos, shift)
			} else {
				moveCursor(e, e.InsertPos-1, shift)
			}

		case platform.XK_Right:
			if ctrl {
				newPos := wordEnd(e.text, e.InsertPos)
				moveCursor(e, newPos, shift)
			} else {
				moveCursor(e, e.InsertPos+1, shift)
			}

		case platform.XK_Home:
			moveCursor(e, 0, shift)

		case platform.XK_End:
			moveCursor(e, len(e.text), shift)

		case platform.XK_BackSpace:
			if e.SelFirst >= 0 {
				e.DeleteSelection()
			} else if e.InsertPos > 0 {
				e.DeleteChars(e.InsertPos-1, 1)
			}

		case platform.XK_Delete:
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
				if r := platform.KeySymToRune(ev.KeySym); r > 0 {
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
	case platform.KeySym(0x0061): // XK_a
		// Select all.
		e.SelectAll()
		e.Display()

	case platform.KeySym(0x006b): // XK_k
		// Kill to end of line.
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, len(e.text)-e.InsertPos)
		}

	case platform.KeySym(0x0064): // XK_d
		// Delete forward.
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, 1)
		}
	}
}
