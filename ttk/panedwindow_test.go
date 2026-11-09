package ttk_test

import (
	"slices"
	"testing"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
)

func TestPanedwindowLayout(t *testing.T) {
	app := testutil.NewTestApp(t)
	pw := ttk.NewPanedwindow(app, "pw", ttk.PanedwindowOrient(ttk.Horizontal))
	a := ttk.NewFrame(pw, "a", ttk.FrameWidth(50), ttk.FrameHeight(40))
	b := ttk.NewFrame(pw, "b", ttk.FrameWidth(50), ttk.FrameHeight(40))
	c := ttk.NewFrame(pw, "c", ttk.FrameWidth(50), ttk.FrameHeight(40))
	for _, f := range []*ttk.Frame{a, b, c} {
		if err := pw.Add(f); err != nil {
			t.Fatal(err)
		}
	}
	if err := pw.Pane(b, ttk.PaneWeight(1)); err != nil {
		t.Fatal(err)
	}
	if err := pw.Add(app.Root()); err == nil {
		t.Error("adding a window that is not a child succeeded")
	}
	sash := screenunit.Pt(3.75).Pixels()
	place.Place(pw, place.X(0), place.Y(0), place.Width(150+2*sash+60), place.Height(40))

	var widths []int
	var s0, moved, shoved, pushed, hit, miss, dragged int
	var panes []string
	app.After(100*time.Millisecond, func() {
		defer app.Quit()
		for _, f := range []*ttk.Frame{a, b, c} {
			widths = append(widths, f.Win.Width)
		}
		s0, _ = pw.SashPos(0)
		moved, _ = pw.SetSashPos(0, 20)
		// Moving the second sash onto the first shoves the first up.
		shoved, _ = pw.SetSashPos(1, 10)
		pushed, _ = pw.SashPos(0)
		hit = pw.Identify(shoved+1, 5)
		miss = pw.Identify(shoved+sash+30, 5)
		// panedwindow.tcl: press on sash 1, drag 15 pixels right, release.
		disp := app.Dispatcher()
		id := pw.Win.PlatformID
		disp.Dispatch(&event.Event{Type: event.ButtonPressType, Window: id, Button: 1, X: shoved + 1, Y: 5})
		disp.Dispatch(&event.Event{Type: event.MotionType, Window: id, X: shoved + 16, Y: 5})
		disp.Dispatch(&event.Event{Type: event.ButtonReleaseType, Window: id, Button: 1, X: shoved + 16, Y: 5})
		dragged, _ = pw.SashPos(1)
		pw.Forget(c)
		for _, w := range pw.Panes() {
			panes = append(panes, w.Name)
		}
	})
	app.MainLoop()

	// The weighted middle pane takes the 60 extra pixels.
	if !slices.Equal(widths, []int{50, 110, 50}) {
		t.Errorf("pane widths = %v, want [50 110 50]", widths)
	}
	if s0 != 50 || moved != 20 {
		t.Errorf("sash 0 at %d, moved to %d; want 50 and 20", s0, moved)
	}
	if shoved != 10 || pushed != 10-sash {
		t.Errorf("sash 1 moved to %d with sash 0 at %d, want 10 and %d (ShoveUp)", shoved, pushed, 10-sash)
	}
	if hit != 1 || miss != -1 {
		t.Errorf("Identify = %d and %d, want 1 and -1", hit, miss)
	}
	if dragged != shoved+15 {
		t.Errorf("dragged sash 1 to %d, want %d", dragged, shoved+15)
	}
	if !slices.Equal(panes, []string{"a", "b"}) {
		t.Errorf("panes after Forget = %v", panes)
	}
}

func TestPanedwindowSashFromTheme(t *testing.T) {
	app := testutil.NewTestApp(t)
	if err := ttk.UseTheme(app, "clam"); err != nil {
		t.Skip("no clam theme:", err)
	}
	pw := ttk.NewPanedwindow(app, "pw", ttk.PanedwindowOrient(ttk.Vertical))
	a := ttk.NewFrame(pw, "a", ttk.FrameWidth(30), ttk.FrameHeight(30))
	b := ttk.NewFrame(pw, "b", ttk.FrameWidth(30), ttk.FrameHeight(30))
	_ = pw.Add(a)
	_ = pw.Add(b)
	app.UpdateIdleTasks()
	app.UpdateIdleTasks()
	// clamTheme.tcl: Sash -sashthickness 4.5p.
	if want := 60 + screenunit.Pt(4.5).Pixels(); pw.Win.ReqHeight != want {
		t.Errorf("clam panedwindow request height = %d, want %d", pw.Win.ReqHeight, want)
	}
}
