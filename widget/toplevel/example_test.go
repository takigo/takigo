package toplevel_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/toplevel"
)

// A toplevel is a second window; widgets created in it are packed into it.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Main"))
	if err != nil {
		log.Fatal(err)
	}
	var about *toplevel.Toplevel
	open := button.New(app, "open", button.Text("About..."), button.Command(func() {
		if about != nil {
			about.Show()
			return
		}
		about = toplevel.New(app, "about",
			toplevel.Title("About"),
			toplevel.TransientFor(app),
			toplevel.Resizable(false, false),
		)
		pack.Pack(label.New(about, "msg", label.Text("takigo example")), pack.PadX(20), pack.PadY(10))
		pack.Pack(button.New(about, "close", button.Text("Close"), button.Command(about.Hide)), pack.PadY(10))
		about.OnClose(about.Hide)
	}))
	pack.Pack(open, pack.PadX(20), pack.PadY(20))
	app.Run()
}
