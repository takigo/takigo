package menu_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/menu"
)

func TestMenuEntryKinds(t *testing.T) {
	app := testutil.NewTestApp(t)
	m := menu.New(app, "m", menu.TearOffOpt(false))
	m.AddCommand("Open", func() {})
	m.AddSeparator()
	m.AddCheckbutton("Wrap", true, func() {})
	m.AddRadiobutton("Left", false, func() {})
	sub := menu.New(m, "sub", menu.TearOffOpt(false))
	m.AddCascade("More", sub)
	want := []menu.EntryType{menu.Command, menu.Separator, menu.Checkbutton, menu.Radiobutton, menu.Cascade}
	entries := m.Entries()
	if len(entries) != len(want) {
		t.Fatalf("%d entries, want %d", len(entries), len(want))
	}
	for i, e := range entries {
		if e.Type != want[i] {
			t.Errorf("entry %d is type %v, want %v", i, e.Type, want[i])
		}
	}
	if !entries[2].Checked || entries[3].Checked {
		t.Errorf("checked flags: %v %v", entries[2].Checked, entries[3].Checked)
	}
	if entries[4].SubMenu != sub {
		t.Error("cascade does not carry its submenu")
	}
	if m.IsPosted() {
		t.Error("a new menu is posted")
	}
}

func TestMenubarLaysOutCascades(t *testing.T) {
	app := testutil.NewTestApp(t)
	mb := menu.NewMenubar(app, "menubar")
	file := menu.New(mb, "file", menu.TearOffOpt(false))
	file.AddCommand("Quit", func() {})
	mb.AddCascade("File", 0, file)
	edit := menu.New(mb, "edit", menu.TearOffOpt(false))
	mb.AddCascade("Edit", 0, edit)
	if mb.Win.ReqHeight <= 0 || mb.Win.ReqWidth <= 0 {
		t.Errorf("menubar requests %dx%d", mb.Win.ReqWidth, mb.Win.ReqHeight)
	}
	if app.Window().Menubar != mb.Win {
		t.Error("the toplevel does not know its menubar")
	}
	// The toplevel's content is pushed below the bar.
	if app.Window().InternalBorderTop < mb.Win.ReqHeight {
		t.Errorf("internal border top %d, bar height %d", app.Window().InternalBorderTop, mb.Win.ReqHeight)
	}
}
