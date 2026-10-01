package takigo_test

import (
	"errors"
	"testing"

	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/darktheme"
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

func TestUseSystemThemeAndDarkPalette(t *testing.T) {
	app := testutil.NewTestApp(t)
	f := ttk.NewFrame(app, "f", ttk.FrameWidth(60), ttk.FrameHeight(40))
	pack.Pack(f)

	t.Setenv("TAKIGO_APPEARANCE", "dark")
	name, err := ttk.UseSystemTheme(app, "clam", "dark")
	if err != nil || name != "dark" || f.Theme.Name != "dark" {
		t.Fatalf("UseSystemTheme = %q, %v; frame theme %q", name, err, f.Theme.Name)
	}
	if app.Appearance().String() != "dark" {
		t.Errorf("App.Appearance() = %v", app.Appearance())
	}
	img := testutil.Grab(t, app, f.Win)
	if c := img.NRGBAAt(30, 20); luma(c) > 80 {
		t.Errorf("a dark-themed frame painted %v, which is not dark", c)
	}

	t.Setenv("TAKIGO_APPEARANCE", "light")
	if name, err := ttk.UseSystemTheme(app, "clam", "dark"); err != nil || name != "clam" || f.Theme.Name != "clam" {
		t.Errorf("light: UseSystemTheme = %q, %v; frame theme %q", name, err, f.Theme.Name)
	}
}
