package canvas

import (
	"slices"
	"testing"

	"github.com/takigo/takigo/event"
)

func TestItemConfigureTagsUpdatesIndex(t *testing.T) {
	c := newBenchCanvas()
	id := c.createItem(newRectOvalItem("rectangle", 0, 0, 10, 10, c), []ItemOption{Tags("old")})
	sid := id
	if err := c.ItemConfigure(sid, Tags("new")); err != nil {
		t.Fatal(err)
	}
	if got := c.FindWithTag("old"); len(got) != 0 {
		t.Errorf("FindWithTag(old) = %v after retagging, want none", got)
	}
	if got := c.FindWithTag("new"); !slices.Equal(got, []ItemID{id}) {
		t.Errorf("FindWithTag(new) = %v, want [%d]", got, id)
	}
}

func TestDeleteDropsItemBindingsAndFocus(t *testing.T) {
	c := newBenchCanvas()
	id := benchScene(c, 1)[0]
	sid := id
	c.BindItem(sid, event.ButtonPressMask, func(*event.Event) {})
	c.focusItemID = id
	c.Delete(sid)
	if _, ok := c.idBindings[id]; ok {
		t.Error("idBindings kept the deleted item's handlers")
	}
	if c.focusItemID != 0 {
		t.Errorf("focusItemID = %d after deleting the focus item, want 0", c.focusItemID)
	}
}

func TestTextCursorLine(t *testing.T) {
	c := newBenchCanvas()
	tests := []struct {
		text      string
		wrap, pos int
		line, col int
	}{
		{"ab\ncd", 0, 4, 1, 1},
		{"ab\ncd", 0, 2, 0, 2},
		{"ab\ncd", 0, 3, 1, 0},
		{"ab\n\ncd", 0, 3, 1, 0},
		// 7px per rune, wrap at 21px: "abc" / "def"; the space is dropped.
		{"abc def", 21, 5, 1, 1},
		{"abc def", 21, 3, 0, 3},
	}
	for _, tt := range tests {
		ti := benchText(c, 0, 0, tt.text)
		ti.wrapLength = tt.wrap
		ti.updateBBox()
		ti.cursorPos = tt.pos
		if l, col := ti.cursorLine(); l != tt.line || col != tt.col {
			t.Errorf("%q wrap %d cursor %d: line %d col %d, want %d %d (lines %q)",
				tt.text, tt.wrap, tt.pos, l, col, tt.line, tt.col, ti.lay.lines)
		}
	}
}

func TestArcPointDistance(t *testing.T) {
	c := newBenchCanvas()
	// Quarter arc of a 100x100 circle centred at (50,50), from 0° to 90°:
	// the curve runs from (100,50) up to (50,0).
	a := newArcItem(0, 0, 100, 100, c)
	a.start, a.extent = 0, 90
	a.style = ArcStyleArc
	a.outlineWidth = 1

	if d := a.PointDistance(50+50*0.7071, 50-50*0.7071); d > 0.6 {
		t.Errorf("point on the arc at 45°: distance %.2f, want ~0", d)
	}
	if d := a.PointDistance(20, 80); d < 20 {
		t.Errorf("point across the circle from the arc: distance %.2f, want far", d)
	}
	a.style = ArcStylePieslice
	a.fill = a.outline
	if d := a.PointDistance(60, 40); d != 0 {
		t.Errorf("point inside a filled pieslice: distance %.2f, want 0", d)
	}
	if got := a.AreaOverlap(0, 60, 40, 100); got != -1 {
		t.Errorf("rectangle in the missing quadrant: AreaOverlap = %d, want -1", got)
	}
	if got := a.AreaOverlap(-10, -10, 110, 110); got != 1 {
		t.Errorf("rectangle around the arc: AreaOverlap = %d, want 1", got)
	}
}

func TestDeleteKeepsDisplayOrder(t *testing.T) {
	c := newBenchCanvas()
	ids := benchScene(c, 10)
	for _, i := range []int{1, 4, 5, 9} {
		c.Delete(ids[i])
	}
	c.Raise(ids[0])
	want := []ItemID{ids[2], ids[3], ids[6], ids[7], ids[8], ids[0]}
	if got := c.FindWithTag("all"); !slices.Equal(got, want) {
		t.Errorf("display list after deletes and a raise = %v, want %v", got, want)
	}
}

// Text item indexes count characters, as Tk's, so editing never splits a
// multi-byte character; "dchars" is inclusive of its last index.
func TestTextItemCharIndexes(t *testing.T) {
	c := newBenchCanvas()
	ti := benchText(c, 0, 0, "añb")
	ti.InsertText(2, "€")
	if ti.text != "añ€b" || ti.CursorIndex() != 3 {
		t.Errorf("after InsertText(2): %q cursor %d, want %q 3", ti.text, ti.CursorIndex(), "añ€b")
	}
	ti.SetCursorPos(2)
	if ti.cursorPos != len("añ") {
		t.Errorf("SetCursorPos(2) = byte %d, want %d", ti.cursorPos, len("añ"))
	}
	ti.DeleteChars(1, 3)
	if ti.text != "ab" || ti.CursorIndex() != 1 {
		t.Errorf("after DeleteChars(1, 3): %q cursor %d, want %q 1", ti.text, ti.CursorIndex(), "ab")
	}
	ti.SetCursorPos(99)
	if ti.CursorIndex() != 2 || ti.CharCount() != 2 {
		t.Errorf("cursor %d count %d after clamping, want 2 2", ti.CursorIndex(), ti.CharCount())
	}
	for _, tt := range []struct {
		index string
		want  int
	}{{"end", 2}, {"insert", 2}, {"insert-1", 1}, {"1", 1}, {"x", 0}} {
		if got := ti.parseIndex(tt.index); got != tt.want {
			t.Errorf("parseIndex(%q) = %d, want %d", tt.index, got, tt.want)
		}
	}

	id := c.createItem(benchText(c, 0, 0, "héllo"), nil)
	c.Dchars(id, "1", "2")
	if got := c.resolve(id)[0].item.(*TextItem).text; got != "hlo" {
		t.Errorf("Dchars(1, 2) left %q, want %q", got, "hlo")
	}
}
