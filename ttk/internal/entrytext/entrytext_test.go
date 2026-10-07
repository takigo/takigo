package entrytext

import (
	"testing"

	"github.com/takigo/takigo/font"
)

// fixedFont is 10px per rune.
type fixedFont struct{}

func (fixedFont) Attrs() font.Attributes     { return font.Attributes{} }
func (fixedFont) Metrics() font.Metrics      { return font.Metrics{} }
func (fixedFont) MeasureString(s string) int { return 10 * len([]rune(s)) }
func (fixedFont) Close()                     {}

func newHelper(s string) (*Helper, *int) {
	redraws := 0
	h := &Helper{Font: fixedFont{}, Redraw: func() { redraws++ }}
	h.ClearSelection()
	h.Set(s)
	return h, &redraws
}

func TestInsertAtReplacesSelection(t *testing.T) {
	t.Parallel()
	h, redraws := newHelper("abcdef")
	h.SelFirst, h.SelLast = 1, 4
	if !h.InsertAt(0, []rune("XY")) || h.Get() != "aXYef" || h.InsertPos != 3 || h.HasSelection() {
		t.Errorf("InsertAt over a selection: %q cursor %d selection %v", h.Get(), h.InsertPos, h.HasSelection())
	}
	if h.InsertAt(2, nil) {
		t.Error("inserting nothing reported a change")
	}
	if *redraws != 2 {
		t.Errorf("%d redraws, want 2 (Set and the insert)", *redraws)
	}
}

func TestValidationAndEditableGate(t *testing.T) {
	t.Parallel()
	h, _ := newHelper("abc")
	var asked []string
	h.Validate = func(_ ValidateReason, v string) bool { asked = append(asked, v); return v != "xabc" }
	if h.InsertAt(0, []rune("x")) || h.Get() != "abc" {
		t.Errorf("a rejected insert changed the text to %q", h.Get())
	}
	if !h.InsertAt(3, []rune("d")) || h.Get() != "abcd" {
		t.Errorf("an accepted insert gave %q", h.Get())
	}
	if len(asked) != 2 || asked[0] != "xabc" || asked[1] != "abcd" {
		t.Errorf("validation was asked about %q", asked)
	}
	h.Editable = func() bool { return false }
	if h.DeleteRange(0, 1) || h.InsertAt(0, []rune("z")) || h.Get() != "abcd" {
		t.Errorf("a read-only helper changed its text to %q", h.Get())
	}
}

func TestDeleteRangeMovesCursor(t *testing.T) {
	t.Parallel()
	h, _ := newHelper("abcdefgh")
	h.InsertPos = 6
	if !h.DeleteRange(2, 4) || h.Get() != "abefgh" || h.InsertPos != 4 {
		t.Errorf("delete before the cursor: %q cursor %d", h.Get(), h.InsertPos)
	}
	h.InsertPos = 3
	h.DeleteRange(2, 5)
	if h.InsertPos != 2 {
		t.Errorf("delete around the cursor left it at %d, want 2", h.InsertPos)
	}
	if h.DeleteRange(3, 3) || h.DeleteRange(-1, -1) {
		t.Error("an empty range reported a change")
	}
}

func TestMoveCursorAndSelectAll(t *testing.T) {
	t.Parallel()
	h, _ := newHelper("abcdef")
	h.InsertPos = 2
	h.MoveCursor(5, -1, true)
	if h.SelFirst != 2 || h.SelLast != 5 || h.InsertPos != 5 {
		t.Errorf("shifted move: sel [%d,%d) cursor %d", h.SelFirst, h.SelLast, h.InsertPos)
	}
	h.MoveCursor(2, h.SelAnchor, true)
	if h.HasSelection() {
		t.Error("moving back to the anchor kept a selection")
	}
	h.MoveCursor(99, -1, false)
	if h.InsertPos != 6 {
		t.Errorf("cursor clamped to %d", h.InsertPos)
	}
	h.SelectAll()
	if h.SelFirst != 0 || h.SelLast != 6 || h.InsertPos != 6 {
		t.Errorf("SelectAll: [%d,%d) cursor %d", h.SelFirst, h.SelLast, h.InsertPos)
	}
	if !h.DeleteSelection() || h.Get() != "" {
		t.Errorf("DeleteSelection left %q", h.Get())
	}
}

func TestClosestGap(t *testing.T) {
	t.Parallel()
	h, _ := newHelper("abcd")
	h.TextX, h.XOffset = 5, 0
	for _, tt := range []struct{ x, want int }{{0, 0}, {5, 0}, {9, 0}, {10, 1}, {14, 1}, {20, 2}, {44, 4}, {99, 4}} {
		if got := h.ClosestGap(tt.x); got != tt.want {
			t.Errorf("ClosestGap(%d) = %d, want %d", tt.x, got, tt.want)
		}
	}
	h.XOffset = 10 // scrolled one character
	if got := h.ClosestGap(5); got != 1 {
		t.Errorf("ClosestGap at the text start while scrolled = %d, want 1", got)
	}
	if got := h.WordStart(3); got != 0 {
		t.Errorf("WordStart = %d", got)
	}
}
