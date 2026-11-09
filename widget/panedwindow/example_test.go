package panedwindow_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/panedwindow"
)

// A paned window lays its panes side by side, with a sash between them the
// user drags; panes are children of the paned window added with Add.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Panedwindow"))
	if err != nil {
		log.Fatal(err)
	}
	pw := panedwindow.New(app, "pane",
		panedwindow.OrientOpt(panedwindow.Horizontal),
		panedwindow.SashWidthOpt(4),
	)
	pack.Pack(pw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	left := label.New(pw, "left", label.Text("left"), label.Background("light yellow"))
	right := label.New(pw, "right", label.Text("right"), label.Background("light cyan"))
	pw.Add(left.Window(), 50)
	pw.Add(right.Window(), 50)
	pw.SetStretch(right.Window(), panedwindow.StretchAlways)
	app.Run()
}
