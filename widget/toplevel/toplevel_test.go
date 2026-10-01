package toplevel_test

import (
	"testing"

	"github.com/msorc/takigo/internal/testutil"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/toplevel"
)

func TestToplevelLifecycle(t *testing.T) {
	app := testutil.NewTestApp(t)
	top := toplevel.New(app, "top", toplevel.Title("Second window"), toplevel.Geometry("300x200"))
	if !top.Win.IsTopLevel() {
		t.Fatal("the window is not a toplevel")
	}
	if top.Win.PathName != ".top" || app.Lookup(".top") != top.Win {
		t.Errorf("path %q, Lookup = %v", top.Win.PathName, app.Lookup(".top"))
	}
	child := label.New(top, "l", label.Text("inside"))
	if child.Win.PathName != ".top.l" {
		t.Errorf("child path = %q", child.Win.PathName)
	}

	top.Show()
	app.UpdateIdleTasks()
	top.Hide()

	if err := top.Configure(toplevel.Geometry("not a geometry")); err == nil {
		t.Error("Configure with a bad geometry returned no error")
	}

	top.Destroy()
	if !child.Win.IsDestroyed() {
		t.Error("destroying the toplevel left its child alive")
	}
	if app.Lookup(".top") != nil {
		t.Error("the destroyed toplevel is still found by path")
	}
}
