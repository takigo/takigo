package dialog

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
)

func TestParseHexColorStrict(t *testing.T) {
	tests := []struct {
		in      string
		r, g, b int
		ok      bool
	}{
		{"#ff8000", 255, 128, 0, true},
		{"#FF8000", 255, 128, 0, true},
		{"  #ff8000 ", 255, 128, 0, true},
		{"#000000", 0, 0, 0, true},
		{"ff8000", 0, 0, 0, false},
		{"#ff800", 0, 0, 0, false},
		{"#ff80000", 0, 0, 0, false},
		{"#gg8000", 0, 0, 0, false},
		{"", 0, 0, 0, false},
	}
	for _, tt := range tests {
		r, g, b, ok := parseHexColorStrict(tt.in)
		if ok != tt.ok || (ok && (r != tt.r || g != tt.g || b != tt.b)) {
			t.Errorf("parseHexColorStrict(%q) = (%d, %d, %d, %v), want (%d, %d, %d, %v)",
				tt.in, r, g, b, ok, tt.r, tt.g, tt.b, tt.ok)
		}
	}
}

// TestColorChooserHexEntry checks that the hex entry drives the sliders
// and the preview, and that its colour is the one the dialog returns.
func TestColorChooserHexEntry(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := newColorChooser(app, "Choose Color", "#000000")
	t.Cleanup(func() { c.dialog.Close(ResultCancel) })

	// The validation of the entry applies a colour that is complete.
	c.hexEntry.SetText("#FF8000")
	if !c.hexEntry.ValidateCmd(c.hexEntry.GetText()) {
		t.Fatal("the hex entry rejected a valid colour")
	}
	if got := int(c.rScale.Get()); got != 255 {
		t.Errorf("red slider = %d, want 255", got)
	}
	if got := int(c.gScale.Get()); got != 128 {
		t.Errorf("green slider = %d, want 128", got)
	}
	if got := int(c.bScale.Get()); got != 0 {
		t.Errorf("blue slider = %d, want 0", got)
	}
	if got := c.color(); got != "#ff8000" {
		t.Errorf("color = %q, want %q", got, "#ff8000")
	}
	// The entry keeps the text the user typed.
	if got := c.hexEntry.GetText(); got != "#FF8000" {
		t.Errorf("hex entry = %q, want %q", got, "#FF8000")
	}

	// An incomplete colour leaves the sliders where they are.
	c.hexEntry.ValidateCmd("#ff80")
	if got := int(c.rScale.Get()); got != 255 {
		t.Errorf("red slider = %d after an invalid colour, want 255", got)
	}

	// Return in the entry applies the colour that is in it.
	c.hexEntry.SetText("#00ff00")
	app.Dispatcher().Dispatch(&event.Event{
		Type:   event.KeyPressType,
		Window: c.hexEntry.Window().PlatformID,
		KeySym: platform.XK_Return,
	})
	if got := int(c.gScale.Get()); got != 255 {
		t.Errorf("green slider = %d after Return, want 255", got)
	}
	if got := c.color(); got != "#00ff00" {
		t.Errorf("color = %q, want %q", got, "#00ff00")
	}
}
