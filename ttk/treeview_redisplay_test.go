package ttk

import (
	"bytes"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/internal/testutil"
)

func TestTreeviewInsertDuplicateID(t *testing.T) {
	testutil.RequireDisplay(t)
	var log bytes.Buffer
	app, err := takigo.NewApp(takigo.Title("test"), takigo.Size(200, 150),
		takigo.WithLogger(slog.New(slog.NewTextHandler(&log, nil))))
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	t.Cleanup(app.Destroy)

	tv := NewTreeview(app, "tv")
	if got := tv.Insert("", -1, ItemID("a")); got != "a" {
		t.Fatalf("Insert(a) = %q", got)
	}
	if log.Len() != 0 {
		t.Errorf("unexpected log output: %s", log.String())
	}
	if got := tv.Insert("", -1, ItemID("a"), ItemText("again")); got != "" {
		t.Errorf("duplicate Insert(a) = %q, want empty", got)
	}
	if !strings.Contains(log.String(), "Item a already exists") {
		t.Errorf("duplicate Insert logged %q", log.String())
	}
	if got := tv.Children(""); !slices.Equal(got, []string{"a"}) {
		t.Errorf("Children = %v, want [a]", got)
	}
	if tv.Item("a").Text != "" {
		t.Errorf("duplicate Insert overwrote the item: text %q", tv.Item("a").Text)
	}
	tv.Delete("a")
	if len(tv.Children("")) != 0 || tv.Exists("a") {
		t.Errorf("after Delete(a): Children = %v, Exists = %v", tv.Children(""), tv.Exists("a"))
	}
}

func TestTreeviewGeneratedIDs(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := NewTreeview(app, "tv")
	if got := tv.Insert("", -1); got != "I001" {
		t.Errorf("first generated ID = %q, want I001", got)
	}
	// A taken ID is skipped, and IDs are hexadecimal like "I%03X".
	tv.Insert("", -1, ItemID("I002"))
	if got := tv.Insert("", -1); got != "I003" {
		t.Errorf("generated ID after explicit I002 = %q, want I003", got)
	}
	for range 6 {
		tv.Insert("", -1)
	}
	if got := tv.Insert("", -1); got != "I00A" {
		t.Errorf("tenth generated ID = %q, want I00A", got)
	}
	seen := map[string]bool{}
	for _, id := range tv.Children("") {
		if seen[id] {
			t.Errorf("ID %s appears twice in %v", id, tv.Children(""))
		}
		seen[id] = true
	}
}

// Mutators redraw at idle, so many calls coalesce into one redisplay.
func TestTreeviewMutatorsCoalesceRedisplay(t *testing.T) {
	app := testutil.NewTestApp(t)
	tv := NewTreeview(app, "tv", TreeviewColumns("a", "b"))
	id := tv.Insert("", -1, ItemID("x"))
	tv.Insert("", -1, ItemID("y"))

	mutators := []struct {
		name string
		fn   func()
	}{
		{"SetItemText", func() { tv.SetItemText(id, "t") }},
		{"SetItemValues", func() { tv.SetItemValues(id, "1", "2") }},
		{"SetCellValue", func() { tv.SetCellValue(id, "a", "v") }},
		{"SelectionSet", func() { tv.SelectionSet("y") }},
		{"SelectionAdd", func() { tv.SelectionAdd("x") }},
		{"SelectionToggle", func() { tv.SelectionToggle("x") }},
		{"SelectionRemove", func() { tv.SelectionRemove("y") }},
		{"SetFocus", func() { tv.SetFocus(id) }},
		{"ColumnConfigure", func() { tv.ColumnConfigure("a", ColWidth(80)) }},
		{"HeadingConfigure", func() { tv.HeadingConfigure("a", HeadText("A")) }},
		{"TagConfigure", func() { tv.TagConfigure("t", TagBackground(0xff0000)) }},
		{"TagAdd", func() { tv.TagAdd("t", id) }},
		{"TagRemove", func() { tv.TagRemove("t", id) }},
		{"TagDelete", func() { tv.TagDelete("t") }},
		{"SetSortIndicator", func() { tv.SetSortIndicator("a", false) }},
		{"ClearSortIndicator", func() { tv.ClearSortIndicator() }},
		{"SetStripe", func() { tv.SetStripe(true) }},
		{"SetColSeparator", func() { tv.SetColSeparator("a", true) }},
		{"YView", func() { tv.YView(1) }},
	}
	for _, m := range mutators {
		app.UpdateIdleTasks()
		if tv.redisplayPending {
			t.Fatalf("%s: redisplay still pending after idle", m.name)
		}
		m.fn()
		if !tv.redisplayPending {
			t.Errorf("%s: no redisplay scheduled", m.name)
		}
	}

	app.UpdateIdleTasks()
	queued := 0
	app.DoWhenIdle(func() { queued++ })
	for i := range 500 {
		tv.SetCellValue(id, "a", strings.Repeat("x", i%7))
	}
	app.UpdateIdleTasks()
	if queued != 1 || tv.redisplayPending {
		t.Errorf("after 500 SetCellValue: queued marker ran %d times, pending = %v", queued, tv.redisplayPending)
	}
}
