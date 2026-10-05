//go:build linux || freebsd || openbsd || netbsd

package takigo_test

import (
	"strconv"
	"sync"
	"testing"

	takigo "github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/internal/xlib"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/menu"
)

// Apps on their own goroutines share no UI state: each lays out classic and
// ttk widgets with all three geometry managers, posts a menu and reads
// pixels back while the others do the same. Run under -race; before the
// managers' tables moved onto the windows this died with concurrent map
// writes.
func TestAppsRunConcurrently(t *testing.T) {
	testutil.RequireDisplay(t)
	// clam inherits from the default theme, so its styles get fallbacks.
	theme := ttk.CurrentTheme().Name
	ttk.SetCurrentTheme("clam")
	defer ttk.SetCurrentTheme(theme)
	testutil.Settle()
	before := xlib.ErrorCount()
	for range 3 {
		testutil.Settle()
		var wg sync.WaitGroup
		for i := range 3 {
			wg.Go(func() { runBusyApp(t, i) })
		}
		wg.Wait()
	}
	if n := xlib.ErrorCount() - before; n != 0 {
		t.Errorf("%d X errors while Apps ran side by side, want 0", n)
	}
}

func runBusyApp(t *testing.T, n int) {
	a, err := takigo.NewApp(takigo.Title("app"+strconv.Itoa(n)), takigo.Size(240, 200))
	if err != nil {
		t.Error(err)
		return
	}
	outer := frame.New(a, "outer")
	pack.Pack(geometry.Group{outer}, pack.Expand(true))
	gridded := frame.New(outer, "gridded")
	placed := frame.New(outer, "placed", frame.Width(120), frame.Height(60))
	pack.Pack(geometry.Group{gridded, placed}, pack.PadX(screenunit.Mm(1)))

	lbl := label.New(gridded, "l", label.Text("classic"))
	btn := button.New(gridded, "b", button.Text("button"))
	tlbl := ttk.NewLabel(gridded, "tl", ttk.LabelText("themed"))
	tbtn := ttk.NewButton(gridded, "tb", ttk.ButtonText("themed"))
	grid.Grid(geometry.Group{lbl, btn}, grid.Row(0))
	grid.Grid(geometry.Group{tlbl, tbtn}, grid.Row(1))
	inner := label.New(placed, "in", label.Text("placed"))
	place.Place(inner, place.X(4), place.Y(4))

	m := menu.New(a, "m")
	m.AddCommand("one", func() {})

	srv := a.Server()
	root := platform.WindowDrawable(a.Root().PlatformID)
	rounds := 0
	var step func()
	step = func() {
		text := strconv.Itoa(rounds)
		lbl.Configure(label.Text(text))
		// A style no theme defines is created in the shared theme on demand.
		tbtn.Configure(ttk.ButtonText(text), ttk.ButtonStyleOpt("App"+strconv.Itoa(n)+"Round"+text+".TButton"))
		place.Place(inner, place.X(4+rounds), place.Y(4))
		pack.Forget(placed)
		pack.Pack(geometry.Group{placed}, pack.PadX(screenunit.Mm(1)))
		a.UpdateIdleTasks()

		m.Post(10, 10)
		m.Unpost()

		pm := srv.CreatePixmap(root, 8, 8, uint(a.Root().Depth))
		if px := srv.GetImageRGBA(platform.PixmapDrawable(pm), 0, 0, 8, 8); px == nil {
			t.Error("GetImageRGBA of a pixmap failed")
		}
		srv.FreePixmap(pm)
		// An unmapped window cannot be read (BadMatch): the error is
		// trapped, not reported.
		if px := srv.GetImageRGBA(platform.WindowDrawable(m.Window().PlatformID), 0, 0, 4, 4); px != nil {
			t.Error("GetImageRGBA of an unmapped window returned pixels")
		}

		if rounds++; rounds < 25 {
			a.DoWhenIdle(step)
		} else {
			a.Quit()
		}
	}
	a.DoWhenIdle(step)
	a.Run()
}

// The error trap around XGetImage is process-wide. Used from two displays
// at once it must still trap each display's own error, and leave the error
// handler reporting errors afterwards.
func TestErrorTrapAcrossDisplays(t *testing.T) {
	testutil.RequireDisplay(t)
	testutil.Settle()
	var apps [2]*takigo.App
	for i := range apps {
		a, err := takigo.NewApp(takigo.Size(20, 20))
		if err != nil {
			t.Fatal(err)
		}
		defer a.Destroy()
		apps[i] = a
	}
	before := xlib.ErrorCount()
	var wg sync.WaitGroup
	for _, a := range apps {
		wg.Go(func() {
			// The root is not mapped before MainLoop: BadMatch.
			d := platform.WindowDrawable(a.Root().PlatformID)
			for range 300 {
				if px := a.Server().GetImageRGBA(d, 0, 0, 4, 4); px != nil {
					t.Error("GetImageRGBA of an unmapped window returned pixels")
					return
				}
			}
		})
	}
	wg.Wait()
	if n := xlib.ErrorCount() - before; n != 0 {
		t.Errorf("%d trapped errors were reported, want 0", n)
	}

	apps[0].Server().FreePixmap(platform.PixmapID(0x7fffffff)) // BadPixmap
	apps[0].Server().Sync(false)
	if n := xlib.ErrorCount() - before; n != 1 {
		t.Errorf("an X error after the traps was reported %d times, want 1", n)
	}
}
