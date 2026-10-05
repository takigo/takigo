package listbox

import (
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/window"
)

type charFont struct{}

func (charFont) Attrs() font.Attributes     { return font.Attributes{} }
func (charFont) Metrics() font.Metrics      { return font.Metrics{Ascent: 8, Descent: 2, MaxWidth: 5} }
func (charFont) MeasureString(s string) int { return 5 * len(s) }
func (charFont) Close()                     {}

func newTestListbox(items ...string) *Listbox {
	lb := &Listbox{items: items, activeIndex: -1, selected: map[int]bool{}}
	lb.Win = &window.Window{Width: 100, Height: 100}
	lb.Font = charFont{}
	return lb
}

func TestInsertDeleteRemapsIndices(t *testing.T) {
	red := &color.ColorRef{Pixel: 0xff0000}
	lb := newTestListbox("a", "b", "c", "d")
	lb.itemBg = map[int]*color.ColorRef{2: red}
	lb.selected[3] = true
	lb.selAnchor, lb.activeIndex = 2, 3

	lb.Insert(1, "x", "y")
	if lb.itemBg[4] != red || lb.itemBg[2] != nil {
		t.Errorf("itemBg after insert = %v, want red at 4", lb.itemBg)
	}
	if !lb.selected[5] || lb.selAnchor != 4 || lb.activeIndex != 5 {
		t.Errorf("after insert: selected=%v anchor=%d active=%d, want 5/4/5", lb.selected, lb.selAnchor, lb.activeIndex)
	}

	lb.Delete(4, 4)
	if len(lb.itemBg) != 0 {
		t.Errorf("itemBg kept a deleted item's colour: %v", lb.itemBg)
	}
	if !lb.selected[4] || lb.selAnchor != 4 || lb.activeIndex != 4 {
		t.Errorf("after delete: selected=%v anchor=%d active=%d, want 4/4/4", lb.selected, lb.selAnchor, lb.activeIndex)
	}
}

func TestMaxWidthTracksEdits(t *testing.T) {
	lb := newTestListbox("ab", "abcd")
	if got := lb.maxWidth(); got != 20 {
		t.Fatalf("maxWidth = %d, want 20", got)
	}
	lb.Insert(0, "abcdef")
	if got := lb.maxWidth(); got != 30 {
		t.Errorf("maxWidth after insert = %d, want 30", got)
	}
	lb.Delete(0, 0)
	if got := lb.maxWidth(); got != 20 {
		t.Errorf("maxWidth after deleting the widest = %d, want 20", got)
	}
}
