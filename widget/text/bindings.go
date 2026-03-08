package text

import (
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

func bindText(t *TextWidget, app widget.AppContext) {
	w := t.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		t.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			t.inset = t.BorderWidth + t.HighlightWidth + 1
			t.notifyYScrollbar()
			t.Display()
		}
	})

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			t.hasFocus = true
			t.cursorOn = true
			t.Display()
		} else if ev.Type == event.FocusOutType {
			t.hasFocus = false
			t.Display()
		}
	})

	// Mouse: click to position cursor.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		switch ev.Button {
		case 1:
			app.Server().SetInputFocus(w.PlatformID, platform.RevertToParent, platform.CurrentTime)
			idx := t.indexFromPixel(ev.X, ev.Y)
			// Fire tag Button-1 bindings before modifying selection.
			if len(t.tagBindings) > 0 {
				for tag := range t.tagsAtIndex(idx) {
					t.fireTagHandlers(tag, "<Button-1>")
				}
			}
			t.clearSelection()
			t.doc.MarkSet("insert", idx)
			t.selAnchor = idx
			t.Display()
		case 4: // mouse wheel up
			t.scrollByDisplayLines(-3)
			t.clampScrollPosition()
			t.notifyYScrollbar()
			t.Display()
		case 5: // mouse wheel down
			t.scrollByDisplayLines(3)
			t.clampScrollPosition()
			t.notifyYScrollbar()
			t.Display()
		}
	})

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, func(ev *event.Event) {
		if len(t.tagBindings) > 0 {
			idx := t.indexFromPixel(ev.X, ev.Y)
			t.updateTagHover(t.tagsAtIndex(idx))
		}
		if ev.State&platform.Button1Mask != 0 {
			idx := t.indexFromPixel(ev.X, ev.Y)
			t.updateSelection(idx)
			t.doc.MarkSet("insert", idx)
			t.Display()
		}
	})

	// Leave: clear tag hover state.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if len(t.tagBindings) > 0 {
			t.updateTagHover(make(map[string]bool))
		}
	})

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		shift := ev.State&platform.ShiftMask != 0
		ctrl := ev.State&platform.ControlMask != 0

		switch ev.KeySym {
		case platform.XK_Left:
			if ctrl {
				newPos := WordStart(t.doc.Marks["insert"].Pos, t.doc)
				moveCursor(t, newPos, shift)
			} else {
				pos := t.doc.Marks["insert"].Pos
				moveCursor(t, Backward(pos, 1, t.doc), shift)
			}

		case platform.XK_Right:
			if ctrl {
				newPos := WordEnd(t.doc.Marks["insert"].Pos, t.doc)
				moveCursor(t, newPos, shift)
			} else {
				pos := t.doc.Marks["insert"].Pos
				moveCursor(t, Forward(pos, 1, t.doc), shift)
			}

		case platform.XK_Up:
			pos := t.doc.Marks["insert"].Pos
			moveCursor(t, UpLine(pos, t.doc), shift)

		case platform.XK_Down:
			pos := t.doc.Marks["insert"].Pos
			moveCursor(t, DownLine(pos, t.doc), shift)

		case platform.XK_Home:
			if ctrl {
				moveCursor(t, Index{1, 0}, shift)
			} else {
				pos := t.doc.Marks["insert"].Pos
				moveCursor(t, LineStart(pos.Line), shift)
			}

		case platform.XK_End:
			if ctrl {
				moveCursor(t, t.doc.EndIndex(), shift)
			} else {
				pos := t.doc.Marks["insert"].Pos
				moveCursor(t, LineEnd(pos.Line, t.doc), shift)
			}

		case platform.XK_Prior: // PageUp
			visLines := (t.Win.Height - 2*t.inset) / t.lineHeight()
			if visLines < 1 {
				visLines = 1
			}
			pos := t.doc.Marks["insert"].Pos
			for i := 0; i < visLines; i++ {
				pos = UpLine(pos, t.doc)
			}
			moveCursor(t, pos, shift)
			t.scrollByDisplayLines(-visLines)
			t.clampScrollPosition()
			t.notifyYScrollbar()

		case platform.XK_Next: // PageDown
			visLines := (t.Win.Height - 2*t.inset) / t.lineHeight()
			if visLines < 1 {
				visLines = 1
			}
			pos := t.doc.Marks["insert"].Pos
			for i := 0; i < visLines; i++ {
				pos = DownLine(pos, t.doc)
			}
			moveCursor(t, pos, shift)
			t.scrollByDisplayLines(visLines)
			t.clampScrollPosition()
			t.notifyYScrollbar()

		case platform.XK_Return:
			if t.readOnly {
				return
			}
			t.undoStack.Separator()
			t.deleteSelection()
			insertAt := t.doc.Marks["insert"].Pos
			endIdx := t.doc.Insert(insertAt, "\n")
			if t.undoEnabled {
				t.undoStack.RecordInsert(insertAt, endIdx, "\n")
			}
			t.doc.MarkSet("insert", endIdx)
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()

		case platform.XK_BackSpace:
			if t.readOnly {
				return
			}
			if t.deleteSelection() {
				t.seeInsert()
				t.notifyYScrollbar()
				t.Display()
			} else {
				pos := t.doc.Marks["insert"].Pos
				if pos.Line > 1 || pos.Char > 0 {
					prevPos := Backward(pos, 1, t.doc)
					text := t.doc.Get(prevPos, pos)
					t.doc.Delete(prevPos, pos)
					if t.undoEnabled {
						t.undoStack.RecordDelete(prevPos, pos, text)
					}
					t.doc.MarkSet("insert", prevPos)
					t.seeInsert()
					t.notifyYScrollbar()
					t.Display()
				}
			}

		case platform.XK_Delete:
			if t.readOnly {
				return
			}
			if t.deleteSelection() {
				t.seeInsert()
				t.notifyYScrollbar()
				t.Display()
			} else {
				pos := t.doc.Marks["insert"].Pos
				endPos := t.doc.EndIndex()
				if Compare(pos, endPos) < 0 {
					nextPos := Forward(pos, 1, t.doc)
					text := t.doc.Get(pos, nextPos)
					t.doc.Delete(pos, nextPos)
					if t.undoEnabled {
						t.undoStack.RecordDelete(pos, nextPos, text)
					}
					t.seeInsert()
					t.notifyYScrollbar()
					t.Display()
				}
			}

		case platform.XK_Tab:
			if t.readOnly {
				return
			}
			t.deleteSelection()
			insertAt := t.doc.Marks["insert"].Pos
			spaces := tabWidth - (insertAt.Char % tabWidth)
			tabStr := strings.Repeat(" ", spaces)
			endIdx := t.doc.Insert(insertAt, tabStr)
			if t.undoEnabled {
				t.undoStack.RecordInsert(insertAt, endIdx, tabStr)
			}
			t.doc.MarkSet("insert", endIdx)
			t.seeInsert()
			t.Display()

		default:
			if ctrl {
				if !t.readOnly {
					handleCtrlKey(t, ev)
				}
				return
			}
			if t.readOnly {
				return
			}
			// Insert printable characters.
			insertStr := ev.Str
			if insertStr == "" {
				if r := platform.KeySymToRune(ev.KeySym); r > 0 {
					insertStr = string(r)
				}
			}
			if insertStr != "" && len(insertStr) > 0 && insertStr[0] >= 32 {
				t.deleteSelection()
				insertAt := t.doc.Marks["insert"].Pos
				endIdx := t.doc.Insert(insertAt, insertStr)
				if t.undoEnabled {
					t.undoStack.RecordInsert(insertAt, endIdx, insertStr)
				}
				t.doc.MarkSet("insert", endIdx)
				t.seeInsert()
				t.notifyYScrollbar()
				t.Display()
			}
		}
	})
}

// moveCursor moves the insert cursor, optionally extending selection.
func moveCursor(t *TextWidget, newPos Index, shift bool) {
	newPos = Clamp(newPos, t.doc)

	if shift {
		if !t.HasSelection() {
			t.selAnchor = t.doc.Marks["insert"].Pos
		}
		t.updateSelection(newPos)
	} else {
		t.clearSelection()
	}

	t.doc.MarkSet("insert", newPos)
	t.seeInsert()
	t.Display()
}

// handleCtrlKey handles Ctrl key combinations.
func handleCtrlKey(t *TextWidget, ev *event.Event) {
	switch ev.KeySym {
	case platform.XK_a:
		t.SelectAll()
		t.Display()
	case platform.XK_z:
		t.Edit("undo")
	case platform.XK_y:
		t.Edit("redo")
	}
}
