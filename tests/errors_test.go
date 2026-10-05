package takigo_test

import (
	"errors"
	"testing"

	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/checkbutton"
	"github.com/takigo/takigo/widget/label"
)

func TestConfigureReturnsOptionErrors(t *testing.T) {
	app := testutil.NewTestApp(t)
	b := button.New(app, "b", button.Text("ok"), button.Background("red"))

	err := b.Configure(button.Background("no-such-colour"), button.Text("still applied"), button.FontOpt("nofont{"))
	if !errors.Is(err, color.ErrUnknown) {
		t.Errorf("Configure error = %v, want color.ErrUnknown", err)
	}
	if b.Text != "still applied" {
		t.Errorf("Text = %q: an option after a failed one was not applied", b.Text)
	}
	if b.Background == nil || b.Background.Red != 0xffff || b.Background.Green != 0 {
		t.Errorf("Background = %+v, want the previous red", b.Background)
	}
	if err := b.Configure(button.Background(color.RGB(0, 128, 0))); err != nil {
		t.Errorf("Configure(RGB) = %v", err)
	}
	if b.Background.Green>>8 != 128 {
		t.Errorf("Background = %+v, want green 128", b.Background)
	}

	l := ttk.NewLabel(app, "l", ttk.LabelText("x"))
	if err := l.Configure(ttk.LabelFont(font.Attributes{Family: "Helvetica", Size: 11})); err != nil {
		t.Errorf("ttk Configure(font attributes) = %v", err)
	}
}

func TestGeometryManagersReturnErrors(t *testing.T) {
	app := testutil.NewTestApp(t)
	a := label.New(app, "a")
	b := label.New(app, "b")

	if err := grid.Grid(a, grid.Row(-1)); !errors.Is(err, grid.ErrBadIndex) {
		t.Errorf("Grid(row -1) = %v, want grid.ErrBadIndex", err)
	}
	if err := pack.Pack(a, pack.After(b)); !errors.Is(err, pack.ErrNotPacked) {
		t.Errorf("Pack(after unpacked) = %v, want pack.ErrNotPacked", err)
	}
	if err := pack.Pack(b); err != nil {
		t.Errorf("Pack = %v", err)
	}
}

func TestCheckbuttonBoolVar(t *testing.T) {
	app := testutil.NewTestApp(t)
	on := widget.NewVariable(true)
	c := checkbutton.New(app, "c", checkbutton.Text("x"), checkbutton.BoolVar(on))
	if !c.Selected() {
		t.Fatal("checkbutton not selected for a true variable")
	}
	c.Toggle()
	if on.Get() || c.Selected() {
		t.Errorf("after Toggle: variable = %v, selected = %v; want false, false", on.Get(), c.Selected())
	}
	on.Set(true)
	if !c.Selected() {
		t.Error("setting the variable did not select the checkbutton")
	}
	c.Destroy()
	on.Set(false) // must not reach the destroyed widget
}
