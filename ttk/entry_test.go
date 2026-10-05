package ttk

import (
	"testing"

	"github.com/takigo/takigo/ttk/entrytext"
)

// --- entrytext.Helper pure-logic tests (no display required) ---

func TestHelperInsertAt(t *testing.T) {
	h := entrytext.Helper{Text: []rune("hello")}
	if !h.InsertAt(2, []rune("XX")) {
		t.Fatal("InsertAt returned false")
	}
	if got := h.Get(); got != "heXXllo" {
		t.Errorf("Get = %q, want %q", got, "heXXllo")
	}
	if h.InsertPos != 4 {
		t.Errorf("InsertPos = %d, want 4", h.InsertPos)
	}
}

func TestHelperInsertReplacesSelection(t *testing.T) {
	h := entrytext.Helper{Text: []rune("hello"), InsertPos: 2, SelAnchor: 2, SelFirst: 2, SelLast: 4}
	if !h.InsertAt(2, []rune("Y")) {
		t.Fatal("InsertAt returned false")
	}
	// InsertAt first deletes the selection [2,4) = "ll", leaving "heo",
	// then inserts "Y" at InsertPos=2 → "heYo".
	if got := h.Get(); got != "heYo" {
		t.Errorf("Get = %q, want %q", got, "heYo")
	}
}

func TestHelperDeleteRange(t *testing.T) {
	h := entrytext.Helper{Text: []rune("hello"), InsertPos: 5}
	if !h.DeleteRange(1, 4) {
		t.Fatal("DeleteRange returned false")
	}
	if got := h.Get(); got != "ho" {
		t.Errorf("Get = %q, want %q", got, "ho")
	}
}

func TestHelperMoveCursorShift(t *testing.T) {
	h := entrytext.Helper{Text: []rune("hello"), InsertPos: 2}
	h.MoveCursor(4, -1, true)
	if !h.HasSelection() {
		t.Fatal("expected selection")
	}
	if h.SelFirst != 2 || h.SelLast != 4 {
		t.Errorf("selection = [%d,%d), want [2,4)", h.SelFirst, h.SelLast)
	}
	h.MoveCursor(3, -1, false)
	if h.HasSelection() {
		t.Errorf("expected no selection after non-shift move, got [%d,%d)", h.SelFirst, h.SelLast)
	}
	if h.InsertPos != 3 {
		t.Errorf("InsertPos = %d, want 3", h.InsertPos)
	}
}

func TestHelperSelectAll(t *testing.T) {
	h := entrytext.Helper{Text: []rune("hello"), InsertPos: 0}
	h.SelectAll()
	if h.SelFirst != 0 || h.SelLast != 5 {
		t.Errorf("selection = [%d,%d), want [0,5)", h.SelFirst, h.SelLast)
	}
}

func TestHelperValidateRejects(t *testing.T) {
	h := entrytext.Helper{
		Text: []rune("hello"),
		Validate: func(_ entrytext.ValidateReason, _ string) bool {
			return false
		},
	}
	if h.InsertAt(2, []rune("X")) {
		t.Fatal("InsertAt should have been rejected")
	}
	if got := h.Get(); got != "hello" {
		t.Errorf("text should be unchanged, got %q", got)
	}
}

func TestHelperEditableFalse(t *testing.T) {
	h := entrytext.Helper{
		Text:     []rune("hello"),
		Editable: func() bool { return false },
	}
	if h.InsertAt(0, []rune("X")) {
		t.Fatal("InsertAt should have been blocked by Editable")
	}
	if h.DeleteRange(0, 1) {
		t.Fatal("DeleteRange should have been blocked by Editable")
	}
}

func TestWordStartEnd(t *testing.T) {
	cases := []struct {
		in       string
		startPos int
		want     int
	}{
		// pos past the end of "hello" → WordStart jumps back to its beginning.
		{"hello world", 8, 6},
		{"hello world", 0, 0},
		// pos at start of "hello" → already at word start.
		{"  hello", 2, 0},
		// pos in middle of "foo_bar" → start of word is 0.
		{"foo_bar baz", 4, 0},
		// pos in middle of "baz" → WordStart goes back to start of "foo_bar".
		{"foo_bar baz", 8, 0},
		{"foo_bar baz", 3, 0},
		{"", 0, 0},
	}
	for _, c := range cases {
		got := entrytext.WordStart([]rune(c.in), c.startPos)
		if got != c.want {
			t.Errorf("WordStart(%q, %d) = %d, want %d", c.in, c.startPos, got, c.want)
		}
	}
}

func TestWordEnd(t *testing.T) {
	// WordEnd skips word chars then any trailing non-word chars, landing at
	// the start of the NEXT word (Emacs-style whitespace gobbling).
	cases := []struct {
		in    string
		pos   int
		wantS int
		wantE int
	}{
		{"hello world", 0, 0, 6},
		{"hello world", 3, 0, 6},
		// WordStart at pos=6 (a space) walks back to the start of the previous
		// word "hello" (index 0); WordEnd at pos=6 walks forward to the next
		// word "world" (index 11).
		{"hello world", 6, 0, 11},
		{"  hello", 0, 0, 2},
	}
	for _, c := range cases {
		s := entrytext.WordStart([]rune(c.in), c.pos)
		e := entrytext.WordEnd([]rune(c.in), c.pos)
		if s != c.wantS || e != c.wantE {
			t.Errorf("pos=%d in=%q: start=%d end=%d, want start=%d end=%d",
				c.pos, c.in, s, e, c.wantS, c.wantE)
		}
	}
}

func TestIsWordChar(t *testing.T) {
	for _, r := range []rune{'a', 'Z', '0', '_'} {
		if !entrytext.IsWordChar(r) {
			t.Errorf("IsWordChar(%q) false, want true", r)
		}
	}
	for _, r := range []rune{' ', '.', '-', '\n'} {
		if entrytext.IsWordChar(r) {
			t.Errorf("IsWordChar(%q) true, want false", r)
		}
	}
}

func TestParseValidate(t *testing.T) {
	cases := map[string]ValidateMode{
		"none":     ValidateNone,
		"key":      ValidateKey,
		"focus":    ValidateFocus,
		"focusin":  ValidateFocusIn,
		"focusout": ValidateFocusOut,
		"all":      ValidateAll,
		"":         ValidateNone,
		"bogus":    ValidateNone,
	}
	for in, want := range cases {
		if got := parseValidate(in); got != want {
			t.Errorf("parseValidate(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestParseEntryState(t *testing.T) {
	cases := map[string]EntryState{
		"normal":    EntryNormal,
		"disabled":  EntryDisabled,
		"readonly":  EntryReadonly,
		"":          EntryNormal,
		"something": EntryNormal,
	}
	for in, want := range cases {
		if got := parseEntryState(in); got != want {
			t.Errorf("parseEntryState(%q) = %d, want %d", in, got, want)
		}
	}
}
