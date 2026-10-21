package menu_test

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget/menu"
)

// TestMenuLetterNavigationNeedsUnderline checks the letter navigation of
// ::tk::TraverseWithinMenu (tk/library/menu.tcl:1028): only an underlined
// character matches, and the entries are scanned from the first one.
func TestMenuLetterNavigationNeedsUnderline(t *testing.T) {
	app := testutil.NewTestApp(t)
	m := menu.New(app, "m", menu.TearOffOpt(false))

	ran := ""
	m.AddCommand("Open", func() { ran += "o" })      // no underline
	m.AddCommandUL("Save", 1, func() { ran += "s" }) // the underlined "a"
	m.AddCommandUL("Quit", 0, func() { ran += "q" }) // the underlined "Q"
	m.Post(10, 10)
	defer m.Unpost()

	press := func(ks platform.KeySym) {
		app.Dispatcher().Dispatch(&event.Event{
			Type: event.KeyPressType, Window: m.Window().PlatformID, KeySym: ks,
		})
	}

	// "Open" has no underline, so "o" must not activate it.
	press(platform.KeySym('o'))
	if ran != "" {
		t.Errorf("typing %q ran %q, want nothing: an entry without an underline never matches", "o", ran)
	}

	// The underlined "a" of "Save" matches.
	press(platform.XK_a)
	if ran != "s" {
		t.Errorf("typing %q ran %q, want %q", "a", ran, "s")
	}

	// The scan starts at the first entry, so the underlined "Q" of "Quit"
	// is found even though a later entry was activated before.
	ran = ""
	press(platform.XK_q)
	if ran != "q" {
		t.Errorf("typing %q ran %q, want %q", "q", ran, "q")
	}
}
