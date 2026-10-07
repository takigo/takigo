package entryedit

import (
	"slices"
	"testing"
)

func newBuffer(s string) *Buffer {
	b := &Buffer{}
	b.Init()
	b.Set(s)
	return b
}

func TestInsertMovesWhatFollows(t *testing.T) {
	t.Parallel()
	b := newBuffer("hello world")
	b.InsertPos, b.LeftIndex = 6, 6
	b.SelectRange(6, 11)
	b.SelAnchor = 6
	if !b.Insert(5, []rune(",")) || b.Get() != "hello, world" {
		t.Fatalf("Insert: %q", b.Get())
	}
	if b.InsertPos != 7 || b.SelFirst != 7 || b.SelLast != 12 || b.SelAnchor != 7 || b.LeftIndex != 7 {
		t.Errorf("indexes after insert: cursor %d sel [%d,%d) anchor %d left %d, want all moved by one",
			b.InsertPos, b.SelFirst, b.SelLast, b.SelAnchor, b.LeftIndex)
	}
	if b.Insert(0, nil) {
		t.Error("inserting nothing reported a change")
	}
	b.Insert(99, []rune("!"))
	if b.Get() != "hello, world!" {
		t.Errorf("insert past the end: %q", b.Get())
	}
}

func TestDeletePullsIndexesBack(t *testing.T) {
	t.Parallel()
	b := newBuffer("abcdefgh")
	b.InsertPos = 6
	b.SelectRange(1, 7)
	b.SelAnchor = 7
	if !b.Delete(2, 3) || b.Get() != "abfgh" {
		t.Fatalf("Delete: %q", b.Get())
	}
	if b.InsertPos != 3 || b.SelFirst != 1 || b.SelLast != 4 || b.SelAnchor != 4 {
		t.Errorf("indexes after delete: cursor %d sel [%d,%d) anchor %d, want 3 [1,4) 4",
			b.InsertPos, b.SelFirst, b.SelLast, b.SelAnchor)
	}
	b.SelectRange(1, 2)
	b.Delete(1, 1)
	if b.HasSelection() {
		t.Error("a selection emptied by a delete stayed")
	}
	if b.Delete(99, 1) || b.Delete(0, 0) {
		t.Error("deleting nothing reported a change")
	}
	b.Delete(2, 99)
	if b.Get() != "af" {
		t.Errorf("delete past the end: %q", b.Get())
	}
}

func TestSelectionAndProspective(t *testing.T) {
	t.Parallel()
	b := newBuffer("abc")
	b.InsertPos = 1
	if got := b.Prospective("X"); got != "aXbc" {
		t.Errorf("Prospective at the cursor = %q", got)
	}
	b.SelectRange(0, 2)
	if got := b.Prospective("X"); got != "Xc" || b.SelectedText() != "ab" {
		t.Errorf("Prospective over the selection = %q, selected %q", got, b.SelectedText())
	}
	b.SelectRange(2, 1)
	if b.HasSelection() {
		t.Error("an empty range selected something")
	}
	b.SelectAll()
	if b.SelFirst != 0 || b.SelLast != 3 {
		t.Errorf("SelectAll = [%d,%d)", b.SelFirst, b.SelLast)
	}
	b.Set("")
	b.SelectAll()
	if b.HasSelection() {
		t.Error("SelectAll on empty text selected something")
	}
}

func TestMoveCursorExtendsFromAnchor(t *testing.T) {
	t.Parallel()
	b := newBuffer("abcdef")
	b.InsertPos = 2
	b.MoveCursor(4, true)
	if b.SelFirst != 2 || b.SelLast != 4 || b.InsertPos != 4 || b.SelAnchor != 2 {
		t.Errorf("shift-right: sel [%d,%d) cursor %d anchor %d", b.SelFirst, b.SelLast, b.InsertPos, b.SelAnchor)
	}
	b.MoveCursor(1, true)
	if b.SelFirst != 1 || b.SelLast != 2 {
		t.Errorf("shift-left past the anchor: sel [%d,%d)", b.SelFirst, b.SelLast)
	}
	b.MoveCursor(2, true)
	if b.HasSelection() {
		t.Error("extending back to the anchor kept a selection")
	}
	b.MoveCursor(99, false)
	if b.InsertPos != 6 || b.HasSelection() {
		t.Errorf("plain move: cursor %d selection %v", b.InsertPos, b.HasSelection())
	}
	b.MoveCursor(-5, false)
	if b.InsertPos != 0 {
		t.Errorf("cursor clamped to %d", b.InsertPos)
	}
}

func TestSetKeepsIndexesInside(t *testing.T) {
	t.Parallel()
	b := newBuffer("long text here")
	b.InsertPos, b.LeftIndex = 10, 5
	b.SelectAll()
	b.Set("ab")
	if b.InsertPos != 2 || b.LeftIndex != 2 || b.HasSelection() || !slices.Equal(b.Text, []rune("ab")) {
		t.Errorf("after Set: cursor %d left %d selection %v text %q", b.InsertPos, b.LeftIndex, b.HasSelection(), b.Get())
	}
}
