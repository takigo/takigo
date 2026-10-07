package entry_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/entry"
)

func TestEntryEditing(t *testing.T) {
	app := testutil.NewTestApp(t)
	e := entry.New(app, "e", entry.Text("hello"))
	if e.GetText() != "hello" {
		t.Fatalf("GetText = %q", e.GetText())
	}
	e.Insert(5, " world")
	e.Insert(0, ">> ")
	if e.GetText() != ">> hello world" {
		t.Errorf("after inserts: %q", e.GetText())
	}
	e.Delete(0, 3)
	if e.GetText() != "hello world" {
		t.Errorf("after delete: %q", e.GetText())
	}

	e.SelectRange(0, 5)
	if e.SelectedText() != "hello" {
		t.Errorf("SelectedText = %q, want hello", e.SelectedText())
	}
	e.DeleteSelection()
	if e.GetText() != " world" || e.SelectedText() != "" {
		t.Errorf("after DeleteSelection: text %q, selection %q", e.GetText(), e.SelectedText())
	}
	e.SelectAll()
	if e.SelectedText() != " world" {
		t.Errorf("SelectAll selected %q", e.SelectedText())
	}
	e.ClearSelection()
	if e.SelectedText() != "" {
		t.Errorf("ClearSelection left %q", e.SelectedText())
	}

	e.SetText("héllo ✓")
	e.Insert(1, "X")
	if e.GetText() != "hXéllo ✓" {
		t.Errorf("indices are not in characters: %q", e.GetText())
	}
}

func TestEntryShowKeepsText(t *testing.T) {
	app := testutil.NewTestApp(t)
	e := entry.New(app, "e", entry.Show('*'), entry.Text("secret"))
	if e.GetText() != "secret" {
		t.Errorf("GetText with Show = %q, want the real text", e.GetText())
	}
}

func TestEntryWidthAndView(t *testing.T) {
	app := testutil.NewTestApp(t)
	narrow := entry.New(app, "n", entry.Width(5))
	wide := entry.New(app, "w", entry.Width(40))
	if wide.Win.ReqWidth <= narrow.Win.ReqWidth {
		t.Errorf("Width(40) is not wider than Width(5): %d vs %d", wide.Win.ReqWidth, narrow.Win.ReqWidth)
	}

	first, last := narrow.VisibleRange()
	if first != 0 || last != 1 {
		t.Errorf("empty entry shows %v..%v, want 0..1", first, last)
	}
}
