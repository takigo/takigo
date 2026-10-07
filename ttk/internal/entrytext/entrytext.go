// Package entrytext provides shared text-editing logic used by ttk entry,
// combobox, and spinbox widgets. It ports the cursor/selection/word-movement
// helpers from tk/generic/tkEntry.c in pure-Go form so the three widgets
// stay consistent.
//
// All methods are safe to call from the event-loop goroutine only.
package entrytext

import (
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/internal/textedit"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Helper holds the editable-text state shared by ttk entry / combobox /
// spinbox. Owners embed or reference it and forward their own events into
// its methods. Mutating methods redraw by calling Redraw (set by the owner).
type Helper struct {
	Text      []rune
	InsertPos int

	// Selection. SelFirst < 0 means no selection; otherwise SelFirst <= SelLast.
	SelFirst  int
	SelLast   int
	SelAnchor int

	Font  font.Font
	TextX int // x-coord (in window) where the text area starts (after padding)
	// XOffset is the pixel width of the text scrolled off the left edge.
	XOffset int

	App widget.AppContext
	Win *window.Window

	// Redraw is called after any change that affects display.
	Redraw func()

	// Editable reports whether text can be modified.
	Editable func() bool

	// Validate, if non-nil, is called before applying edits. Return false to
	// reject the edit.
	Validate func(reason ValidateReason, newValue string) bool

	imeMark int // insert position when the input method began composing
}

// ValidateReason mirrors tk/generic/tkEntry.c VREASON.
type ValidateReason int

const (
	ValidateKey      ValidateReason = iota // character insert / delete
	ValidateFocusIn                        // focus entered
	ValidateFocusOut                       // focus left
	ValidateForced                         // forced via $w validate
)

// HasSelection reports whether a non-empty selection is currently active.
func (h *Helper) HasSelection() bool {
	return h.SelFirst >= 0 && h.SelLast > h.SelFirst
}

// ClearSelection drops the selection but keeps InsertPos.
func (h *Helper) ClearSelection() {
	h.SelFirst = -1
	h.SelLast = -1
}

// DeleteSelection removes the selected text (if any) and collapses the cursor.
// Returns true if anything was deleted.
func (h *Helper) DeleteSelection() bool {
	if !h.HasSelection() {
		return false
	}
	return h.DeleteRange(h.SelFirst, h.SelLast)
}

// DeleteRange deletes [first, last) and shifts InsertPos if needed.
func (h *Helper) DeleteRange(first, last int) bool {
	if first >= last {
		return false
	}
	if first < 0 {
		first = 0
	}
	if last > len(h.Text) {
		last = len(h.Text)
	}
	if !h.allowEdit() {
		return false
	}
	if !h.runValidate(ValidateKey, string(append(append([]rune{}, h.Text[:first]...), h.Text[last:]...))) {
		return false
	}
	h.Text = append(h.Text[:first], h.Text[last:]...)
	if h.InsertPos > first {
		if h.InsertPos > last {
			h.InsertPos -= last - first
		} else {
			h.InsertPos = first
		}
	}
	h.ClearSelection()
	h.redraw()
	return true
}

// InsertAt inserts runes at pos, honouring any current selection and validation.
func (h *Helper) InsertAt(pos int, runes []rune) bool {
	if len(runes) == 0 {
		return false
	}
	if !h.allowEdit() {
		return false
	}
	pos = max(0, min(pos, len(h.Text)))
	cut, end := pos, pos
	if h.HasSelection() {
		cut, end = h.SelFirst, h.SelLast
	}
	out := make([]rune, 0, len(h.Text)-(end-cut)+len(runes))
	out = append(out, h.Text[:cut]...)
	out = append(out, runes...)
	out = append(out, h.Text[end:]...)
	if !h.runValidate(ValidateKey, string(out)) {
		return false
	}
	h.Text = out
	h.InsertPos = cut + len(runes)
	h.ClearSelection()
	h.redraw()
	return true
}

// Set replaces the text wholesale. Runs validation if configured.
func (h *Helper) Set(s string) {
	if !h.runValidate(ValidateKey, s) {
		return
	}
	h.Text = []rune(s)
	h.InsertPos = len(h.Text)
	h.ClearSelection()
	h.redraw()
}

// Get returns the current text.
func (h *Helper) Get() string {
	return string(h.Text)
}

// MoveCursor moves the insertion cursor to newPos. With shift, the existing
// selection anchor is used to extend the selection (matching Emacs/Tk).
func (h *Helper) MoveCursor(newPos, anchor int, shift bool) {
	if newPos < 0 {
		newPos = 0
	}
	if newPos > len(h.Text) {
		newPos = len(h.Text)
	}
	if shift {
		if !h.HasSelection() {
			h.SelAnchor = anchor
			if h.SelAnchor < 0 {
				h.SelAnchor = h.InsertPos
			}
		}
		if newPos < h.SelAnchor {
			h.SelFirst = newPos
			h.SelLast = h.SelAnchor
		} else {
			h.SelFirst = h.SelAnchor
			h.SelLast = newPos
		}
		if h.SelFirst == h.SelLast {
			h.ClearSelection()
		}
	} else {
		h.ClearSelection()
	}
	h.InsertPos = newPos
	h.redraw()
}

// SelectAll selects the entire text.
func (h *Helper) SelectAll() {
	if len(h.Text) == 0 {
		return
	}
	h.SelAnchor = 0
	h.SelFirst = 0
	h.SelLast = len(h.Text)
	h.InsertPos = len(h.Text)
	h.redraw()
}

// ClosestGap returns the rune index whose boundary (between idx-1 and idx) is
// nearest to pixel x relative to the widget window. Used by mouse press/drag.
func (h *Helper) ClosestGap(x int) int {
	if h.Font == nil || len(h.Text) == 0 {
		return 0
	}
	xInText := x - h.TextX + h.XOffset
	if xInText <= 0 {
		return 0
	}
	idx := textedit.RuneIndexAtPixel(h.Font, h.Text, xInText)
	if idx >= len(h.Text) {
		return len(h.Text)
	}
	charStart := h.Font.MeasureString(string(h.Text[:idx]))
	charEnd := h.Font.MeasureString(string(h.Text[:idx+1]))
	if xInText >= (charStart+charEnd)/2 {
		return idx + 1
	}
	return idx
}

// WordStart returns the rune index of the start of the word at or before pos.
func (h *Helper) WordStart(pos int) int { return textedit.WordStart(h.Text, pos) }

// WordEnd returns the rune index past the end of the word at or after pos.
func (h *Helper) WordEnd(pos int) int { return textedit.WordEnd(h.Text, pos) }

// ClipboardSet puts the current selection on the clipboard.
func (h *Helper) ClipboardSet(time platform.Timestamp) {
	if !h.HasSelection() {
		return
	}
	h.App.Clipboard().Set(h.Win.PlatformID, string(h.Text[h.SelFirst:h.SelLast]), time)
}

// ClipboardGet pastes from the clipboard at InsertPos.
func (h *Helper) ClipboardGet(time platform.Timestamp) {
	h.App.Clipboard().Get(h.Win.PlatformID, time, func(text string) {
		if text == "" {
			return
		}
		h.InsertAt(h.InsertPos, []rune(text))
	})
}

// HandleKey processes a printable character key event (no shift/ctrl modifier).
// Returns true if the event was consumed.
func (h *Helper) HandleKey(ev *event.Event) bool {
	if ev.State&platform.ControlMask != 0 {
		return false
	}
	s := ev.Str
	if s == "" {
		if r := platform.KeySymToRune(ev.KeySym); r > 0 {
			s = string(r)
		}
	}
	if s == "" || s[0] < 32 {
		return false
	}
	h.InsertAt(h.InsertPos, []rune(s))
	return true
}

// HandleCtrlKey dispatches Ctrl+key combos. Returns true if consumed.
func (h *Helper) HandleCtrlKey(ev *event.Event) bool {
	if ev.State&platform.ControlMask == 0 {
		return false
	}
	anchor := h.SelAnchor
	if anchor < 0 {
		anchor = h.InsertPos
	}
	switch ev.KeySym {
	case platform.XK_a:
		h.SelectAll()
		return true
	case platform.XK_e:
		h.MoveCursor(len(h.Text), anchor, false)
		return true
	case platform.XK_b:
		h.MoveCursor(h.InsertPos-1, anchor, false)
		return true
	case platform.XK_f:
		h.MoveCursor(h.InsertPos+1, anchor, false)
		return true
	case platform.XK_c:
		h.ClipboardSet(ev.Time)
		return true
	case platform.XK_x:
		h.ClipboardSet(ev.Time)
		h.DeleteSelection()
		h.redraw()
		return true
	case platform.XK_v:
		h.ClipboardGet(ev.Time)
		return true
	case platform.XK_w:
		h.ClipboardSet(ev.Time)
		h.DeleteSelection()
		h.redraw()
		return true
	case platform.XK_k:
		if h.InsertPos < len(h.Text) {
			h.DeleteRange(h.InsertPos, len(h.Text))
		}
		return true
	case platform.XK_d:
		if h.InsertPos < len(h.Text) {
			h.DeleteRange(h.InsertPos, h.InsertPos+1)
		}
		return true
	}
	return false
}

// HandleNavKey handles arrow / Home / End keys. Returns true if consumed.
func (h *Helper) HandleNavKey(ev *event.Event) bool {
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&platform.ControlMask != 0
	anchor := h.SelAnchor
	if anchor < 0 {
		anchor = h.InsertPos
	}
	switch ev.KeySym {
	case platform.XK_Left:
		if ctrl {
			h.MoveCursor(h.WordStart(h.InsertPos), anchor, shift)
		} else {
			h.MoveCursor(h.InsertPos-1, anchor, shift)
		}
		return true
	case platform.XK_Right:
		if ctrl {
			h.MoveCursor(h.WordEnd(h.InsertPos), anchor, shift)
		} else {
			h.MoveCursor(h.InsertPos+1, anchor, shift)
		}
		return true
	case platform.XK_Home:
		h.MoveCursor(0, anchor, shift)
		return true
	case platform.XK_End:
		h.MoveCursor(len(h.Text), anchor, shift)
		return true
	}
	return false
}

// HandleEditKey handles BackSpace / Delete / Insert. Returns true if consumed.
func (h *Helper) HandleEditKey(ev *event.Event) bool {
	shift := ev.State&platform.ShiftMask != 0
	ctrl := ev.State&platform.ControlMask != 0
	switch ev.KeySym {
	case platform.XK_BackSpace:
		if h.HasSelection() {
			h.DeleteSelection()
		} else if h.InsertPos > 0 {
			h.DeleteRange(h.InsertPos-1, h.InsertPos)
		}
		h.redraw()
		return true
	case platform.XK_Delete:
		if shift && h.HasSelection() {
			h.ClipboardSet(ev.Time)
			h.DeleteSelection()
		} else if h.HasSelection() {
			h.DeleteSelection()
		} else if h.InsertPos < len(h.Text) {
			h.DeleteRange(h.InsertPos, h.InsertPos+1)
		}
		h.redraw()
		return true
	case platform.XK_Insert:
		if ctrl && h.HasSelection() {
			h.ClipboardSet(ev.Time)
		} else if shift {
			h.ClipboardGet(ev.Time)
		}
		return true
	}
	return false
}

// HandleVirtual handles the input method's virtual events like the TEntry
// bindings in library/ttk/entry.tcl: the text being composed is inserted as
// usual and shown selected, and deleted when the input method replaces it.
func (h *Helper) HandleVirtual(ev *event.Event) {
	switch ev.Name {
	case event.IMEStart:
		h.imeMark = h.InsertPos
	case event.IMEEnd:
		if h.imeMark < h.InsertPos && h.InsertPos <= len(h.Text) {
			h.SelFirst, h.SelLast, h.SelAnchor = h.imeMark, h.InsertPos, h.imeMark
			h.redraw()
		}
	case event.IMEClear:
		h.DeleteRange(min(h.imeMark, len(h.Text)), h.InsertPos)
	case event.AccentBackspace:
		h.HandleEditKey(&event.Event{KeySym: platform.XK_BackSpace})
	}
}

// -------- internal --------

func (h *Helper) redraw() {
	if h.Redraw != nil {
		h.Redraw()
	}
}

func (h *Helper) allowEdit() bool {
	if h.Editable != nil {
		return h.Editable()
	}
	return true
}

func (h *Helper) runValidate(reason ValidateReason, newValue string) bool {
	if h.Validate == nil {
		return true
	}
	return h.Validate(reason, newValue)
}
