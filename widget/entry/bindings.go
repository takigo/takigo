package entry

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/entryutil"
)

// bindEntry registers all event handlers for the entry widget.
func bindEntry(e *Entry, app widget.AppContext) {
	w := e.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, e.handleExpose)

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, e.handleConfigure)

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, e.handleFocus)

	// Mouse: click to position cursor and take focus.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, e.handleButtonPress)

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, e.handleMotion)

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, e.handleKeyPress)
}

// handleExpose handles Exposure events.
func (e *Entry) handleExpose(ev *event.Event) {
	if ev.ExposeCount > 0 {
		return
	}
	e.Display()
}

// handleConfigure handles ConfigureNotify (resize) events.
func (e *Entry) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		e.Win.Width = ev.ConfigWidth
		e.Win.Height = ev.ConfigHeight
		e.computeGeometry()
		e.Display()
		// Tk's EntryUpdateScrollbar runs on every redisplay, so the
		// scrollbar learns the real visible fraction once laid out.
		e.notifyScrollbar()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (e *Entry) handleFocus(ev *event.Event) {
	if ev.Type == event.FocusInType {
		e.HasFocus = true
		e.CursorOn = true
		e.tryFocusValidate("focusin")
		e.Display()
	} else if ev.Type == event.FocusOutType {
		e.HasFocus = false
		e.tryFocusValidate("focusout")
		e.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (e *Entry) handleButtonPress(ev *event.Event) {
	if ev.Button == 1 {
		// Request X11 input focus so key events come to this window.
		e.App.Server().SetInputFocus(e.Win.PlatformID, platform.RevertToParent, platform.CurrentTime)
		e.ClearSelection()
		e.InsertPos = e.closestGap(ev.X)
		e.SelAnchor = e.InsertPos
		e.Display()
	}
}

// handleMotion handles mouse motion events (drag to select).
func (e *Entry) handleMotion(ev *event.Event) {
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
}

// handleKeyPress handles keyboard events.
func (e *Entry) handleKeyPress(ev *event.Event) {
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&(platform.ControlMask|platform.Mod2Mask) != 0 // Ctrl or Cmd (macOS)

	switch ev.KeySym {
	case platform.XK_Left:
		if ctrl {
			newPos := entryutil.WordStart(e.text, e.InsertPos)
			moveCursor(e, newPos, shift)
		} else {
			moveCursor(e, e.InsertPos-1, shift)
		}

	case platform.XK_Right:
		if ctrl {
			newPos := entryutil.WordEnd(e.text, e.InsertPos)
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
			prospective := string(e.text[:e.SelFirst]) + string(e.text[e.SelLast:])
			if e.tryEdit(prospective) {
				e.DeleteSelection()
			}
		} else if e.InsertPos > 0 {
			prospective := string(e.text[:e.InsertPos-1]) + string(e.text[e.InsertPos:])
			if e.tryEdit(prospective) {
				e.DeleteChars(e.InsertPos-1, 1)
			}
		}

	case platform.XK_Insert:
		// Ctrl+Insert: copy; Shift+Insert: paste.
		if ctrl {
			if e.SelFirst >= 0 {
				sel := string(e.text[e.SelFirst:e.SelLast])
				e.App.Clipboard().Set(e.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			}
		} else if shift {
			e.App.Clipboard().Get(e.Win.PlatformID, platform.Timestamp(ev.Time), func(text string) {
				if text == "" {
					return
				}
				var prospective string
				if e.SelFirst >= 0 {
					prospective = string(e.text[:e.SelFirst]) + text + string(e.text[e.SelLast:])
				} else {
					prospective = string(e.text[:e.InsertPos]) + text + string(e.text[e.InsertPos:])
				}
				if e.tryEdit(prospective) {
					if e.SelFirst >= 0 {
						e.DeleteSelection()
					}
					e.InsertChars(e.InsertPos, text)
				}
			})
		}

	case platform.XK_Delete:
		// Shift+Delete: cut selection.
		if shift && e.SelFirst >= 0 {
			sel := string(e.text[e.SelFirst:e.SelLast])
			e.App.Clipboard().Set(e.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			prospective := string(e.text[:e.SelFirst]) + string(e.text[e.SelLast:])
			if e.tryEdit(prospective) {
				e.DeleteSelection()
			}
			return
		}
		if e.SelFirst >= 0 {
			prospective := string(e.text[:e.SelFirst]) + string(e.text[e.SelLast:])
			if e.tryEdit(prospective) {
				e.DeleteSelection()
			}
		} else if e.InsertPos < len(e.text) {
			prospective := string(e.text[:e.InsertPos]) + string(e.text[e.InsertPos+1:])
			if e.tryEdit(prospective) {
				e.DeleteChars(e.InsertPos, 1)
			}
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
			// Compute prospective value accounting for any selection deletion.
			var prospective string
			if e.SelFirst >= 0 {
				prospective = string(e.text[:e.SelFirst]) + insertStr + string(e.text[e.SelLast:])
			} else {
				prospective = string(e.text[:e.InsertPos]) + insertStr + string(e.text[e.InsertPos:])
			}
			if e.tryEdit(prospective) {
				if e.SelFirst >= 0 {
					e.DeleteSelection()
				}
				e.InsertChars(e.InsertPos, insertStr)
			}
		}
	}
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
	case platform.XK_a: // Ctrl+A: move to start of field (Emacs)
		moveCursor(e, 0, false)

	case platform.XK_e: // Ctrl+E: move to end of field (Emacs)
		moveCursor(e, len(e.text), false)

	case platform.XK_b: // Ctrl+B: move back one char (Emacs)
		moveCursor(e, e.InsertPos-1, false)

	case platform.XK_f: // Ctrl+F: move forward one char (Emacs)
		moveCursor(e, e.InsertPos+1, false)

	case platform.XK_c: // Ctrl+C: copy selection
		if e.SelFirst >= 0 {
			sel := string(e.text[e.SelFirst:e.SelLast])
			e.App.Clipboard().Set(e.Win.PlatformID, sel, platform.Timestamp(ev.Time))
		}

	case platform.XK_x: // Ctrl+X: cut selection
		if e.SelFirst >= 0 {
			sel := string(e.text[e.SelFirst:e.SelLast])
			e.App.Clipboard().Set(e.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			prospective := string(e.text[:e.SelFirst]) + string(e.text[e.SelLast:])
			if e.tryEdit(prospective) {
				e.DeleteSelection()
			}
		}

	case platform.XK_v: // Ctrl+V: paste from clipboard
		e.App.Clipboard().Get(e.Win.PlatformID, platform.Timestamp(ev.Time), func(text string) {
			if text == "" {
				return
			}
			var prospective string
			if e.SelFirst >= 0 {
				prospective = string(e.text[:e.SelFirst]) + text + string(e.text[e.SelLast:])
			} else {
				prospective = string(e.text[:e.InsertPos]) + text + string(e.text[e.InsertPos:])
			}
			if e.tryEdit(prospective) {
				if e.SelFirst >= 0 {
					e.DeleteSelection()
				}
				e.InsertChars(e.InsertPos, text)
			}
		})

	case platform.XK_w: // Ctrl+W: cut selection (Emacs kill-region)
		if e.SelFirst >= 0 {
			sel := string(e.text[e.SelFirst:e.SelLast])
			e.App.Clipboard().Set(e.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			prospective := string(e.text[:e.SelFirst]) + string(e.text[e.SelLast:])
			if e.tryEdit(prospective) {
				e.DeleteSelection()
			}
		}

	case platform.XK_k: // Ctrl+K: kill to end of field
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, len(e.text)-e.InsertPos)
		}

	case platform.XK_d: // Ctrl+D: delete char forward
		if e.InsertPos < len(e.text) {
			e.DeleteChars(e.InsertPos, 1)
		}
	}
}
