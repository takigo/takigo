package takigo_test

import (
	"testing"

	takigo "github.com/takigo/takigo"

	"github.com/takigo/takigo/appearance"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/darktheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/label"
)

func newAppWith(t *testing.T, opts ...takigo.AppOption) *takigo.App {
	t.Helper()
	testutil.RequireDisplay(t)
	app, err := takigo.NewApp(opts...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(app.Destroy)
	return app
}

func TestDarkAppearance(t *testing.T) {
	app := newAppWith(t, takigo.UseAppearance(appearance.Dark))
	if !widget.PaletteFor(app).Dark {
		t.Fatal("the App did not get the dark palette")
	}
	l := label.New(app, "l", label.Text("MMMM"), label.PadX(10), label.PadY(10))
	e := entry.New(app, "e")
	c := canvas.New(app, "c", canvas.Width(40), canvas.Height(20))
	line := c.CreateLine([]float64{0, 10, 40, 10})
	pack.Pack(l)
	pack.Pack(e)
	pack.Pack(c)

	if l.Background.Pixel != 0x2e2e2e || l.Foreground.Pixel != 0xe6e6e6 {
		t.Errorf("label colours = %06x on %06x, want e6e6e6 on 2e2e2e", l.Foreground.Pixel, l.Background.Pixel)
	}
	if e.Background.Pixel != 0x1e1e1e {
		t.Errorf("entry background = %06x, want the dark field colour 1e1e1e", e.Background.Pixel)
	}
	// A themed widget follows without being asked.
	if b := ttk.NewButton(app, "tb"); b.Theme.Name != "dark" {
		t.Errorf("ttk theme = %q, want dark", b.Theme.Name)
	}
	_ = line

	img := testutil.Grab(t, app, l.Win)
	if c := img.NRGBAAt(2, 2); luma(c) > 80 {
		t.Errorf("the label's background is painted %v, which is not dark", c)
	}
	light := 0
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if luma(img.NRGBAAt(x, y)) > 160 {
				light++
			}
		}
	}
	if light < 20 {
		t.Errorf("only %d light pixels: the label's text is not light on dark", light)
	}
}

func TestLightIsTheDefault(t *testing.T) {
	app := newAppWith(t)
	if widget.PaletteFor(app) != widget.LightPalette {
		t.Error("a plain App does not have Tk's light palette")
	}
	l := label.New(app, "l", label.Text("x"))
	if l.Background.Pixel != 0xd9d9d9 || l.Foreground.Pixel != 0 {
		t.Errorf("label colours = %06x on %06x, want Tk's 000000 on d9d9d9", l.Foreground.Pixel, l.Background.Pixel)
	}
	if b := ttk.NewButton(app, "tb"); b.Theme.Name == "dark" {
		t.Error("a light App got the dark ttk theme")
	}
}

func TestFollowSystemAppearance(t *testing.T) {
	t.Setenv("TAKIGO_APPEARANCE", "dark")
	if app := newAppWith(t, takigo.FollowSystemAppearance()); !widget.PaletteFor(app).Dark {
		t.Error("FollowSystemAppearance ignored a dark desktop")
	}
}

func TestCanvasInkFollowsPalette(t *testing.T) {
	app := newAppWith(t, takigo.UseAppearance(appearance.Dark))
	c := canvas.New(app, "c", canvas.Width(60), canvas.Height(20), canvas.BorderWidthOpt(0), canvas.HighlightWidthOpt(0))
	pack.Pack(c)
	c.CreateLine([]float64{0, 10, 60, 10}, canvas.OutlineWidth(3))
	img := testutil.Grab(t, app, c.Win)
	if on, off := luma(img.NRGBAAt(30, 10)), luma(img.NRGBAAt(30, 3)); on < 180 || off > 80 {
		t.Errorf("line luma %d on background luma %d: a default item is not visible on a dark canvas", on, off)
	}
}
