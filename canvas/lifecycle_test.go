package canvas

import (
	"slices"
	"strconv"
	"testing"

	"github.com/msorc/takigo/event"
)

func TestItemConfigureTagsUpdatesIndex(t *testing.T) {
	c := newBenchCanvas()
	id := c.createItem(newRectOvalItem("rectangle", 0, 0, 10, 10, c), []ItemOption{Tags("old")})
	sid := strconv.FormatInt(id, 10)
	if err := c.ItemConfigure(sid, Tags("new")); err != nil {
		t.Fatal(err)
	}
	if got := c.FindWithTag("old"); len(got) != 0 {
		t.Errorf("FindWithTag(old) = %v after retagging, want none", got)
	}
	if got := c.FindWithTag("new"); !slices.Equal(got, []int64{id}) {
		t.Errorf("FindWithTag(new) = %v, want [%d]", got, id)
	}
}

func TestDeleteDropsItemBindingsAndFocus(t *testing.T) {
	c := newBenchCanvas()
	id := benchScene(c, 1)[0]
	sid := strconv.FormatInt(id, 10)
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
