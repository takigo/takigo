package takigo_test

import (
	"errors"
	"testing"

	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
)

func TestUseThemeRethemesOneApp(t *testing.T) {
	app := testutil.NewTestApp(t)
	before := ttk.CurrentTheme().Name
	other := "clam"
	if before == "clam" {
		other = "default"
	}

	b := ttk.NewButton(app, "b", ttk.ButtonText("themed"))
	r := ttk.NewRadiobutton(app, "r", ttk.RadiobuttonText("choice"))
	pack.Pack(b)
	pack.Pack(r)
	app.UpdateIdleTasks()
	if b.Theme.Name != before {
		t.Fatalf("button theme = %q, want the default %q", b.Theme.Name, before)
	}

	if err := ttk.UseTheme(app, "no-such-theme"); !errors.Is(err, ttk.ErrUnknownTheme) {
		t.Errorf("UseTheme(unknown) = %v, want ErrUnknownTheme", err)
	}
	if err := ttk.UseTheme(app, other); err != nil {
		t.Fatal(err)
	}
	if b.Theme.Name != other || r.Theme.Name != other {
		t.Errorf("existing widgets have themes %q and %q, want %q", b.Theme.Name, r.Theme.Name, other)
	}
	if r.Win.ReqWidth <= 0 || r.Win.ReqHeight <= 0 {
		t.Errorf("radiobutton request after re-theming = %dx%d", r.Win.ReqWidth, r.Win.ReqHeight)
	}
	if n := ttk.NewLabel(app, "l"); n.Theme.Name != other {
		t.Errorf("new widget theme = %q, want %q", n.Theme.Name, other)
	}
	// The process default, which other Apps use, is untouched.
	if got := ttk.CurrentTheme().Name; got != before {
		t.Errorf("process default theme changed to %q", got)
	}
}
