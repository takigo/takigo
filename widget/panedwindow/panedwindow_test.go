package panedwindow_test

import (
	"testing"

	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/panedwindow"
)

func TestPanedWindowPanes(t *testing.T) {
	app := testutil.NewTestApp(t)
	pw := panedwindow.New(app, "pw")
	a := frame.New(pw, "a", frame.Width(100), frame.Height(60))
	b := frame.New(pw, "b", frame.Width(150), frame.Height(80))

	pw.Add(a.Win, 0)
	pw.Add(b.Win, 0)
	if panes := pw.Panes(); len(panes) != 2 || panes[0] != a.Win || panes[1] != b.Win {
		t.Fatalf("Panes = %v", panes)
	}
	app.UpdateIdleTasks()
	// Side by side: the request covers both widths and the taller pane.
	if pw.Win.ReqWidth < 250 || pw.Win.ReqHeight < 80 {
		t.Errorf("request = %dx%d, want at least 250x80", pw.Win.ReqWidth, pw.Win.ReqHeight)
	}

	pw.Remove(a.Win)
	if panes := pw.Panes(); len(panes) != 1 || panes[0] != b.Win {
		t.Errorf("after Remove: %v", panes)
	}

	// Destroying a pane's window takes it out of the paned window.
	b.Destroy()
	if n := len(pw.Panes()); n != 0 {
		t.Errorf("%d panes after destroying the last one", n)
	}
}

func TestPanedWindowVertical(t *testing.T) {
	app := testutil.NewTestApp(t)
	pw := panedwindow.New(app, "pw", panedwindow.OrientOpt(panedwindow.Vertical))
	a := frame.New(pw, "a", frame.Width(100), frame.Height(60))
	b := frame.New(pw, "b", frame.Width(150), frame.Height(80))
	pw.Add(a.Win, 0)
	pw.Add(b.Win, 0)
	app.UpdateIdleTasks()
	if pw.Win.ReqHeight < 140 || pw.Win.ReqWidth < 150 {
		t.Errorf("stacked request = %dx%d, want at least 150x140", pw.Win.ReqWidth, pw.Win.ReqHeight)
	}
}
