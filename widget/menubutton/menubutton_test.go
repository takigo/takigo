package menubutton_test

import (
	"testing"

	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/widget/menu"
	"github.com/takigo/takigo/widget/menubutton"
)

func TestMenubuttonGeometry(t *testing.T) {
	app := testutil.NewTestApp(t)
	short := menubutton.New(app, "s", menubutton.Text("File"))
	long := menubutton.New(app, "l", menubutton.Text("A longer caption"))
	if long.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("longer text is not wider: %d vs %d", long.Win.ReqWidth, short.Win.ReqWidth)
	}
	ind := menubutton.New(app, "i", menubutton.Text("File"), menubutton.IndicatorOnOpt(true))
	if ind.Win.ReqWidth <= short.Win.ReqWidth {
		t.Errorf("the indicator did not widen the button: %d vs %d", ind.Win.ReqWidth, short.Win.ReqWidth)
	}
	before := short.Win.ReqWidth
	short.SetText("File and more")
	if short.Win.ReqWidth <= before {
		t.Errorf("SetText did not grow the request: %d vs %d", short.Win.ReqWidth, before)
	}
}

func TestMenubuttonPostsItsMenu(t *testing.T) {
	app := testutil.NewTestApp(t)
	mb := menubutton.New(app, "mb", menubutton.Text("File"))
	m := menu.New(mb, "m")
	m.AddCommand("Open", func() {})
	if err := mb.Configure(menubutton.MenuOpt(m)); err != nil {
		t.Fatal(err)
	}
	pack.Pack(mb)
	app.UpdateIdleTasks()

	mb.PostMenu()
	if !m.IsPosted() {
		t.Fatal("PostMenu did not post the menu")
	}
	m.Unpost()
	if m.IsPosted() {
		t.Error("the menu is still posted after Unpost")
	}
}
