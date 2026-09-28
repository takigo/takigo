package text

import (
	"fmt"
	"strings"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
)

// bindText registers all event handlers for the text widget.
func bindText(t *TextWidget, app widget.AppContext) {
	w := t.Win

	// Expose.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, t.handleExpose)

	// Configure (resize).
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, t.handleConfigure)

	// Focus.
	app.Dispatcher().Bind(w.PlatformID, event.FocusChangeMask, t.handleFocus)

	// Mouse: click to position cursor.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, t.handleButtonPress)

	// Mouse: drag to select.
	app.Dispatcher().Bind(w.PlatformID, event.MotionMask, t.handleMotion)

	// Leave: clear tag hover state.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, t.handleLeave)

	// Keyboard.
	app.Dispatcher().Bind(w.PlatformID, event.KeyPressMask, t.handleKeyPress)
	app.Dispatcher().Bind(w.PlatformID, event.VirtualMask, t.handleVirtual)
}

// handleVirtual handles the input method's virtual events like the Text
// bindings in library/text.tcl: the text being composed is inserted as
// usual and underlined with the IMEmarkedtext tag, and deleted when the
// input method replaces it.
func (t *TextWidget) handleVirtual(ev *event.Event) {
	if t.readOnly {
		return
	}
	insert := t.doc.Marks["insert"].Pos
	switch ev.Name {
	case event.IMEStart:
		t.imeMark = insert
	case event.IMEEnd:
		t.TagAdd("IMEmarkedtext", indexString(t.imeMark), "insert")
		t.TagConfigure("IMEmarkedtext", TagUnderline(true))
	case event.IMEClear:
		t.Delete(indexString(t.imeMark), "insert")
		t.seeInsert()
	case event.AccentBackspace:
		t.Delete(indexString(Backward(insert, 1, t.doc)), "insert")
		t.seeInsert()
	}
}

func indexString(i Index) string { return fmt.Sprintf("%d.%d", i.Line, i.Char) }

// handleExpose handles Exposure events.
func (t *TextWidget) handleExpose(ev *event.Event) {
	if ev.ExposeCount > 0 {
		return
	}
	t.Display()
}

// handleConfigure handles ConfigureNotify (resize) events.
func (t *TextWidget) handleConfigure(ev *event.Event) {
	if ev.Type == event.ConfigureType {
		t.Win.Width = ev.ConfigWidth
		t.Win.Height = ev.ConfigHeight
		t.inset = t.BorderWidth + t.HighlightWidth
		t.notifyYScrollbar()
		t.Display()
	}
}

// handleFocus handles FocusIn/FocusOut events.
func (t *TextWidget) handleFocus(ev *event.Event) {
	if ev.Type == event.FocusInType {
		t.hasFocus = true
		t.cursorOn = true
		t.Display()
	} else if ev.Type == event.FocusOutType {
		t.hasFocus = false
		t.Display()
	}
}

// handleButtonPress handles mouse button press events.
func (t *TextWidget) handleButtonPress(ev *event.Event) {
	switch ev.Button {
	case 1:
		t.App.Server().SetInputFocus(t.Win.PlatformID, platform.RevertToParent, platform.CurrentTime)
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
		t.lastDragIdx = idx
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
}

// handleMotion handles mouse motion events (drag to select, tag hover).
func (t *TextWidget) handleMotion(ev *event.Event) {
	if len(t.tagBindings) > 0 {
		idx := t.indexFromPixel(ev.X, ev.Y)
		t.updateTagHover(t.tagsAtIndex(idx))
	}
	if ev.State&platform.Button1Mask != 0 {
		idx := t.indexFromPixel(ev.X, ev.Y)
		if idx == t.lastDragIdx {
			return
		}
		t.lastDragIdx = idx
		t.updateSelection(idx)
		t.doc.MarkSet("insert", idx)
		t.Display()
	}
}

// handleLeave handles Leave events (clear tag hover).
func (t *TextWidget) handleLeave(ev *event.Event) {
	if len(t.tagBindings) > 0 {
		t.updateTagHover(make(map[string]bool))
	}
}

// handleKeyPress handles keyboard events.
func (t *TextWidget) handleKeyPress(ev *event.Event) {
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&(platform.ControlMask|platform.Mod2Mask) != 0 // Ctrl or Cmd (macOS)

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
		visLines := max((t.Win.Height-2*t.insetY)/t.lineHeight(), 1)
		pos := t.doc.Marks["insert"].Pos
		for range visLines {
			pos = UpLine(pos, t.doc)
		}
		moveCursor(t, pos, shift)
		t.scrollByDisplayLines(-visLines)
		t.clampScrollPosition()
		t.notifyYScrollbar()

	case platform.XK_Next: // PageDown
		visLines := max((t.Win.Height-2*t.insetY)/t.lineHeight(), 1)
		pos := t.doc.Marks["insert"].Pos
		for range visLines {
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

	case platform.XK_Insert:
		// Ctrl+Insert: copy; Shift+Insert: paste.
		if ctrl {
			if sel := t.GetSelection(); sel != "" {
				t.App.Clipboard().Set(t.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			}
		} else if shift {
			if t.readOnly {
				return
			}
			t.App.Clipboard().Get(t.Win.PlatformID, platform.Timestamp(ev.Time), func(text string) {
				if text == "" {
					return
				}
				t.undoStack.Separator()
				t.deleteSelection()
				insertAt := t.doc.Marks["insert"].Pos
				endIdx := t.doc.Insert(insertAt, text)
				if t.undoEnabled {
					t.undoStack.RecordInsert(insertAt, endIdx, text)
				}
				t.doc.MarkSet("insert", endIdx)
				t.seeInsert()
				t.notifyYScrollbar()
				t.Display()
			})
		}

	case platform.XK_Delete:
		// Shift+Delete: cut selection.
		if shift && !t.readOnly {
			if sel := t.GetSelection(); sel != "" {
				t.App.Clipboard().Set(t.Win.PlatformID, sel, platform.Timestamp(ev.Time))
				t.deleteSelection()
				t.seeInsert()
				t.notifyYScrollbar()
				t.Display()
			}
			return
		}
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
			handleCtrlKey(t, ev)
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
	// --- Navigation (work in read-only mode too) ---
	case platform.XK_a: // Ctrl+A: move to line start (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, LineStart(pos.Line), ev.State&platform.ShiftMask != 0)
	case platform.XK_e: // Ctrl+E: move to line end (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, LineEnd(pos.Line, t.doc), ev.State&platform.ShiftMask != 0)
	case platform.XK_b: // Ctrl+B: move back one char (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, Backward(pos, 1, t.doc), ev.State&platform.ShiftMask != 0)
	case platform.XK_f: // Ctrl+F: move forward one char (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, Forward(pos, 1, t.doc), ev.State&platform.ShiftMask != 0)
	case platform.XK_p: // Ctrl+P: move up one line (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, UpLine(pos, t.doc), ev.State&platform.ShiftMask != 0)
	case platform.XK_n: // Ctrl+N: move down one line (Emacs)
		pos := t.doc.Marks["insert"].Pos
		moveCursor(t, DownLine(pos, t.doc), ev.State&platform.ShiftMask != 0)

	// --- Clipboard (work in read-only mode for copy) ---
	case platform.XK_c: // Ctrl+C: copy selection
		if sel := t.GetSelection(); sel != "" {
			t.App.Clipboard().Set(t.Win.PlatformID, sel, platform.Timestamp(ev.Time))
		}
	case platform.XK_x: // Ctrl+X: cut selection
		if t.readOnly {
			return
		}
		if sel := t.GetSelection(); sel != "" {
			t.App.Clipboard().Set(t.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			t.deleteSelection()
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()
		}
	case platform.XK_v: // Ctrl+V: paste from clipboard
		if t.readOnly {
			return
		}
		t.App.Clipboard().Get(t.Win.PlatformID, platform.Timestamp(ev.Time), func(text string) {
			if text == "" {
				return
			}
			t.undoStack.Separator()
			t.deleteSelection()
			insertAt := t.doc.Marks["insert"].Pos
			endIdx := t.doc.Insert(insertAt, text)
			if t.undoEnabled {
				t.undoStack.RecordInsert(insertAt, endIdx, text)
			}
			t.doc.MarkSet("insert", endIdx)
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()
		})

	// --- Editing (blocked in read-only mode) ---
	case platform.XK_d: // Ctrl+D: delete char forward (Emacs)
		if t.readOnly {
			return
		}
		pos := t.doc.Marks["insert"].Pos
		if Compare(pos, t.doc.EndIndex()) < 0 {
			nextPos := Forward(pos, 1, t.doc)
			text := t.doc.Get(pos, nextPos)
			t.doc.Delete(pos, nextPos)
			if t.undoEnabled {
				t.undoStack.RecordDelete(pos, nextPos, text)
			}
			t.notifyYScrollbar()
			t.Display()
		}
	case platform.XK_k: // Ctrl+K: kill to end of line (Emacs)
		if t.readOnly {
			return
		}
		pos := t.doc.Marks["insert"].Pos
		lineEnd := LineEnd(pos.Line, t.doc)
		if Compare(pos, lineEnd) == 0 {
			// At end of line: delete the newline joining with next line.
			if pos.Line < t.doc.LineCount() {
				nextPos := Index{Line: pos.Line + 1, Char: 0}
				text := t.doc.Get(pos, nextPos)
				t.doc.Delete(pos, nextPos)
				if t.undoEnabled {
					t.undoStack.RecordDelete(pos, nextPos, text)
				}
			}
		} else {
			text := t.doc.Get(pos, lineEnd)
			t.doc.Delete(pos, lineEnd)
			if t.undoEnabled {
				t.undoStack.RecordDelete(pos, lineEnd, text)
			}
		}
		t.notifyYScrollbar()
		t.Display()
	case platform.XK_w: // Ctrl+W: cut selection (Emacs kill-region)
		if t.readOnly {
			return
		}
		if sel := t.GetSelection(); sel != "" {
			t.App.Clipboard().Set(t.Win.PlatformID, sel, platform.Timestamp(ev.Time))
			t.deleteSelection()
			t.seeInsert()
			t.notifyYScrollbar()
			t.Display()
		}
	case platform.XK_z: // Ctrl+Z: undo
		t.Edit("undo")
	case platform.XK_y: // Ctrl+Y: redo
		t.Edit("redo")
	}
}
