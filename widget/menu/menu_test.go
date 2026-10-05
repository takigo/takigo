package menu_test

import (
	"testing"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/menu"
)

func TestMenuEntries(t *testing.T) {
	app := testutil.NewTestApp(t)
	m := menu.New(app, "m", menu.TearOffOpt(false))
	ran := ""
	m.AddCommand("Open", func() { ran += "o" })
	m.AddCommandAccel("Save", "Ctrl+S", func() { ran += "s" })
	m.AddSeparator()
	sub := menu.New(m, "sub", menu.TearOffOpt(false))
	sub.AddCommand("Inner", func() {})
	m.AddCascade("More", sub)
	m.AddCheckbutton("Wrap", true, func() {})
	m.AddRadiobutton("Left", false, func() {})

	entries := m.Entries()
	if len(entries) != 6 {
		t.Fatalf("%d entries, want 6", len(entries))
	}
	if len(sub.Entries()) != 1 {
		t.Errorf("submenu has %d entries, want 1", len(sub.Entries()))
	}

	m.PrepareGeometry()
	if m.Win.ReqWidth <= 0 || m.Win.ReqHeight <= 0 {
		t.Errorf("menu request = %dx%d", m.Win.ReqWidth, m.Win.ReqHeight)
	}
	one := menu.New(app, "one", menu.TearOffOpt(false))
	one.AddCommand("Open", func() {})
	one.PrepareGeometry()
	if one.Win.ReqHeight >= m.Win.ReqHeight {
		t.Errorf("six entries are not taller than one: %d vs %d", m.Win.ReqHeight, one.Win.ReqHeight)
	}
}

func TestMenuPostUnpost(t *testing.T) {
	app := testutil.NewTestApp(t)
	m := menu.New(app, "m", menu.TearOffOpt(false))
	m.AddCommand("Open", func() {})
	if m.IsPosted() {
		t.Fatal("a new menu is posted")
	}
	m.Post(10, 10)
	if !m.IsPosted() {
		t.Fatal("Post did not post")
	}
	m.Unpost()
	m.Unpost() // idempotent
	if m.IsPosted() {
		t.Error("still posted after Unpost")
	}
	m.Post(10, 10)
	m.Destroy()
	if m.IsPosted() {
		t.Error("a destroyed menu is still posted")
	}
}
