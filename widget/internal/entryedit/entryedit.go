// Package entryedit holds the editable text of the classic entry and
// spinbox widgets, which tk/generic/tkEntry.c serves with one structure:
// the text, the insertion cursor, the selection and the scroll index, the
// edits that keep them consistent (EntryInsert, EntryDelete) and the key
// bindings of library/entry.tcl.
package entryedit

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/textedit"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Buffer is the text of an entry with its cursor, selection and scroll
// position, in runes. Widgets embed it; its fields are their state.
type Buffer struct {
	Text      []rune
	InsertPos int // rune index of the insertion cursor (0..len(Text))

	// Selection, the half-open interval [SelFirst, SelLast); SelFirst < 0
	// is none. SelAnchor is the end that stays put while it extends.
	SelFirst  int
	SelLast   int
	SelAnchor int

	LeftIndex int // rune index of the leftmost visible character
	ImeMark   int // insert position when the input method began composing
}

// Init clears the selection; a zero Buffer has one at 0.
func (b *Buffer) Init() {
	b.SelFirst, b.SelLast = -1, -1
}

// Get returns the text.
func (b *Buffer) Get() string { return string(b.Text) }

// Set replaces the text, keeping the cursor and scroll index inside it
// and dropping the selection.
func (b *Buffer) Set(s string) {
	b.Text = []rune(s)
	b.InsertPos = min(b.InsertPos, len(b.Text))
	b.LeftIndex = min(b.LeftIndex, len(b.Text))
	b.ClearSelection()
}

// Insert puts runes before index and moves the cursor, selection and
// scroll index that follow it (EntryInsert). It reports whether the text
// changed.
func (b *Buffer) Insert(index int, runes []rune) bool {
	count := len(runes)
	if count == 0 {
		return false
	}
	index = textedit.ClampIdx(index, len(b.Text))
	text := make([]rune, 0, len(b.Text)+count)
	text = append(text, b.Text[:index]...)
	text = append(text, runes...)
	text = append(text, b.Text[index:]...)
	b.Text = text
	if b.InsertPos >= index {
		b.InsertPos += count
	}
	if b.SelFirst >= index {
		b.SelFirst += count
	}
	if b.SelLast > index {
		b.SelLast += count
	}
	if b.SelAnchor >= index {
		b.SelAnchor += count
	}
	if b.LeftIndex > index {
		b.LeftIndex += count
	}
	return true
}

// Delete removes count runes from index and pulls the cursor, selection
// and scroll index back over them (EntryDelete). It reports whether the
// text changed.
func (b *Buffer) Delete(index, count int) bool {
	if count <= 0 || len(b.Text) == 0 || index >= len(b.Text) {
		return false
	}
	index = max(index, 0)
	count = min(count, len(b.Text)-index)
	b.Text = append(b.Text[:index], b.Text[index+count:]...)
	adjust := func(idx *int) {
		switch {
		case *idx < 0:
		case *idx >= index+count:
			*idx -= count
		case *idx >= index:
			*idx = index
		}
	}
	adjust(&b.InsertPos)
	adjust(&b.SelFirst)
	adjust(&b.SelLast)
	adjust(&b.SelAnchor)
	adjust(&b.LeftIndex)
	if b.SelFirst >= 0 && b.SelLast <= b.SelFirst {
		b.ClearSelection()
	}
	return true
}

// HasSelection reports whether some text is selected.
func (b *Buffer) HasSelection() bool { return b.SelFirst >= 0 && b.SelLast > b.SelFirst }

// ClearSelection drops the selection.
func (b *Buffer) ClearSelection() { b.SelFirst, b.SelLast = -1, -1 }

// SelectRange selects [first, last), clamped to the text; an empty range
// clears the selection.
func (b *Buffer) SelectRange(first, last int) {
	first = max(first, 0)
	last = min(last, len(b.Text))
	if first >= last {
		b.ClearSelection()
		return
	}
	b.SelFirst, b.SelLast = first, last
}

// SelectAll selects the whole text.
func (b *Buffer) SelectAll() {
	if len(b.Text) > 0 {
		b.SelectRange(0, len(b.Text))
	}
}

// SelectedText returns the selected text.
func (b *Buffer) SelectedText() string {
	if !b.HasSelection() {
		return ""
	}
	return string(b.Text[b.SelFirst:b.SelLast])
}

// Prospective returns the text as it would be with the selection, or the
// gap at the cursor, replaced by s: what -validatecommand is asked about.
func (b *Buffer) Prospective(s string) string {
	if b.HasSelection() {
		return string(b.Text[:b.SelFirst]) + s + string(b.Text[b.SelLast:])
	}
	return string(b.Text[:b.InsertPos]) + s + string(b.Text[b.InsertPos:])
}

// MoveCursor puts the cursor at pos, clamped to the text; with shift the
// selection extends from its anchor (or the old cursor) to it.
func (b *Buffer) MoveCursor(pos int, shift bool) {
	pos = textedit.ClampIdx(pos, len(b.Text))
	if !shift {
		b.ClearSelection()
		b.InsertPos = pos
		return
	}
	if b.SelFirst < 0 {
		b.SelAnchor = b.InsertPos
	}
	b.ExtendTo(pos)
}

// ExtendTo selects from the anchor to pos and puts the cursor at pos
// (a drag, or a shifted cursor key).
func (b *Buffer) ExtendTo(pos int) {
	pos = textedit.ClampIdx(pos, len(b.Text))
	if pos < b.SelAnchor {
		b.SelFirst, b.SelLast = pos, b.SelAnchor
	} else {
		b.SelFirst, b.SelLast = b.SelAnchor, pos
	}
	if b.SelFirst == b.SelLast {
		b.ClearSelection()
	}
	b.InsertPos = pos
}

// Editor binds a Buffer to the widget that shows it, so the key bindings
// of library/entry.tcl can edit through the widget's own methods (which
// re-lay out, scroll and redraw) and validate as the widget does.
type Editor struct {
	Buf *Buffer
	App widget.AppContext
	Win *window.Window

	// TryEdit reports whether the text may become prospective
	// (-validate key, -validatecommand).
	TryEdit func(prospective string) bool
	// Insert and Delete are the widget's InsertChars and DeleteChars.
	Insert func(index int, s string)
	Delete func(index, count int)
	// Moved runs after the cursor or the selection changed: the widget
	// scrolls the cursor into view and redraws.
	Moved func()
}

// MoveCursor moves the cursor (extending the selection with shift) and
// lets the widget scroll and redraw.
func (e *Editor) MoveCursor(pos int, shift bool) {
	e.Buf.MoveCursor(pos, shift)
	e.Moved()
}

// InsertText replaces the selection, or inserts at the cursor, when the
// widget's validation allows it.
func (e *Editor) InsertText(s string) {
	if s == "" || !e.TryEdit(e.Buf.Prospective(s)) {
		return
	}
	if e.Buf.HasSelection() {
		e.Delete(e.Buf.SelFirst, e.Buf.SelLast-e.Buf.SelFirst)
	}
	e.Insert(e.Buf.InsertPos, s)
}

// DeleteSelection deletes the selection when validation allows it and
// reports whether it did (or there was none).
func (e *Editor) DeleteSelection() bool {
	if !e.Buf.HasSelection() {
		return false
	}
	if !e.TryEdit(e.Buf.Prospective("")) {
		return true
	}
	e.Delete(e.Buf.SelFirst, e.Buf.SelLast-e.Buf.SelFirst)
	return true
}

// deleteAt deletes count runes at index when validation allows it.
func (e *Editor) deleteAt(index, count int) {
	b := e.Buf
	if index < 0 || index >= len(b.Text) {
		return
	}
	prospective := string(b.Text[:index]) + string(b.Text[min(index+count, len(b.Text)):])
	if e.TryEdit(prospective) {
		e.Delete(index, count)
	}
}

// Backspace ports tk::EntryBackspace: delete the selection, or the
// character before the cursor.
func (e *Editor) Backspace() {
	if !e.DeleteSelection() {
		e.deleteAt(e.Buf.InsertPos-1, 1)
	}
}

// DeleteForward deletes the selection, or the character after the cursor.
func (e *Editor) DeleteForward() {
	if !e.DeleteSelection() {
		e.deleteAt(e.Buf.InsertPos, 1)
	}
}

// Copy puts the selection on the clipboard.
func (e *Editor) Copy(time platform.Timestamp) {
	if e.Buf.HasSelection() {
		e.App.Clipboard().Set(e.Win.PlatformID, e.Buf.SelectedText(), time)
	}
}

// Cut copies the selection and deletes it.
func (e *Editor) Cut(time platform.Timestamp) {
	if e.Buf.HasSelection() {
		e.Copy(time)
		e.DeleteSelection()
	}
}

// Paste inserts the clipboard's text when it arrives.
func (e *Editor) Paste(time platform.Timestamp) {
	e.App.Clipboard().Get(e.Win.PlatformID, time, e.InsertText)
}

// HandleKey ports the key bindings of library/entry.tcl: cursor movement
// (by character and, with Control, by word), editing keys, the Emacs and
// clipboard Control bindings, and typed characters. It reports whether it
// handled the key.
func (e *Editor) HandleKey(ev *event.Event) bool {
	b := e.Buf
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&(platform.ControlMask|platform.CommandMask) != 0
	switch ev.KeySym {
	case platform.XK_Left:
		if ctrl {
			e.MoveCursor(textedit.WordStart(b.Text, b.InsertPos), shift)
		} else {
			e.MoveCursor(b.InsertPos-1, shift)
		}
	case platform.XK_Right:
		if ctrl {
			e.MoveCursor(textedit.WordEnd(b.Text, b.InsertPos), shift)
		} else {
			e.MoveCursor(b.InsertPos+1, shift)
		}
	case platform.XK_Home:
		e.MoveCursor(0, shift)
	case platform.XK_End:
		e.MoveCursor(len(b.Text), shift)
	case platform.XK_BackSpace:
		e.Backspace()
	case platform.XK_Delete:
		if shift && b.HasSelection() {
			e.Cut(ev.Time)
		} else {
			e.DeleteForward()
		}
	case platform.XK_Insert:
		switch {
		case ctrl:
			e.Copy(ev.Time)
		case shift:
			e.Paste(ev.Time)
		}
	default:
		if ctrl {
			return e.handleCtrlKey(ev)
		}
		// What %A would give: the composed string, or the keysym's
		// character; modifiers and function keys have none.
		s := ev.Str
		if s == "" {
			if r := platform.KeySymToRune(ev.KeySym); r > 0 {
				s = string(r)
			}
		}
		if s == "" || s[0] < 32 {
			return false
		}
		e.InsertText(s)
	}
	return true
}

// handleCtrlKey ports the Control bindings: Emacs movement and deletion,
// and the clipboard.
func (e *Editor) handleCtrlKey(ev *event.Event) bool {
	b := e.Buf
	switch ev.KeySym {
	case platform.XK_a:
		e.MoveCursor(0, false)
	case platform.XK_e:
		e.MoveCursor(len(b.Text), false)
	case platform.XK_b:
		e.MoveCursor(b.InsertPos-1, false)
	case platform.XK_f:
		e.MoveCursor(b.InsertPos+1, false)
	case platform.XK_c:
		e.Copy(ev.Time)
	case platform.XK_x, platform.XK_w:
		e.Cut(ev.Time)
	case platform.XK_v:
		e.Paste(ev.Time)
	case platform.XK_k:
		e.deleteAt(b.InsertPos, len(b.Text)-b.InsertPos)
	case platform.XK_d:
		e.deleteAt(b.InsertPos, 1)
	case platform.XK_h:
		e.Backspace()
	case platform.KeySym('/'):
		b.SelectAll()
		e.Moved()
	case platform.KeySym('\\'):
		b.ClearSelection()
		e.Moved()
	default:
		return false
	}
	return true
}

// HandleVirtual handles the input method's virtual events like the Entry
// bindings in library/entry.tcl: the text being composed is inserted as
// usual and shown selected, and deleted when the input method replaces it.
func (e *Editor) HandleVirtual(ev *event.Event) {
	b := e.Buf
	switch ev.Name {
	case event.IMEStart:
		b.ImeMark = b.InsertPos
		return
	case event.IMEEnd:
		b.SelectRange(b.ImeMark, b.InsertPos)
		b.SelAnchor = b.ImeMark
	case event.IMEClear:
		if first := min(b.ImeMark, len(b.Text)); first < b.InsertPos {
			e.deleteAt(first, b.InsertPos-first)
		}
	case event.AccentBackspace:
		e.Backspace()
	default:
		return
	}
	e.Moved()
}
