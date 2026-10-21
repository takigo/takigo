package entry_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/entry"
)

// TestEntryOptions checks the options that were missing: the fields of the
// entry could not be set before (Tk's -insertwidth, -selectbackground,
// -selectforeground, -insertbackground, -justify and -state).
func TestEntryOptions(t *testing.T) {
	app := testutil.NewTestApp(t)
	e := entry.New(app, "e",
		entry.InsertWidth(4),
		entry.SelectBackground("red"),
		entry.SelectForeground("blue"),
		entry.InsertBackground("green"),
		entry.Anchor(option.AnchorE),
		entry.Justify(option.JustifyRight),
		entry.State(widget.StateNormal),
	)
	if e.InsertWidth != 4 {
		t.Errorf("InsertWidth = %d, want 4", e.InsertWidth)
	}
	if e.SelBg == nil {
		t.Error("SelectBackground did not set the selection background")
	}
	if e.SelFg == nil {
		t.Error("SelectForeground did not set the selection foreground")
	}
	if e.InsertBg == nil {
		t.Error("InsertBackground did not set the insert background")
	}
	if e.Anchor != option.AnchorE {
		t.Errorf("Anchor = %v, want e", e.Anchor)
	}
	if e.Justify != option.JustifyRight {
		t.Errorf("Justify = %v, want right", e.Justify)
	}

	// The same options work at runtime.
	if err := e.Configure(entry.InsertWidth(6), entry.Justify(option.JustifyCenter)); err != nil {
		t.Fatalf("Configure: %v", err)
	}
	if e.InsertWidth != 6 {
		t.Errorf("InsertWidth = %d after Configure, want 6", e.InsertWidth)
	}
	if e.Justify != option.JustifyCenter {
		t.Errorf("Justify = %v after Configure, want center", e.Justify)
	}
}

// TestEntryState checks Tk's -state and -readonly: a disabled entry cannot
// be edited, a read-only one can be selected but not edited
// (tk/generic/tkEntry.c: the insert and delete commands need the normal
// state).
func TestEntryState(t *testing.T) {
	app := testutil.NewTestApp(t)

	normal := entry.New(app, "n", entry.Text("abc"))
	normal.Insert(0, "x")
	if normal.GetText() != "xabc" {
		t.Errorf("a normal entry was not edited: %q", normal.GetText())
	}

	disabled := entry.New(app, "d", entry.Text("abc"), entry.State(widget.StateDisabled))
	disabled.Insert(0, "x")
	disabled.Delete(0, 1)
	if disabled.GetText() != "abc" {
		t.Errorf("a disabled entry was edited: %q", disabled.GetText())
	}
	if disabled.State != widget.StateDisabled {
		t.Errorf("State = %v, want disabled", disabled.State)
	}

	readonly := entry.New(app, "r", entry.Text("abc"), entry.ReadOnly(true))
	readonly.Insert(0, "x")
	readonly.Delete(0, 1)
	if readonly.GetText() != "abc" {
		t.Errorf("a read-only entry was edited: %q", readonly.GetText())
	}
	// A read-only entry can still be selected and copied.
	readonly.SelectRange(0, 2)
	if readonly.SelectedText() != "ab" {
		t.Errorf("a read-only entry selected %q, want %q", readonly.SelectedText(), "ab")
	}
}
