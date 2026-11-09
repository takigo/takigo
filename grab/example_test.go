package grab_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/toplevel"
)

// A local grab sends the App's pointer and keyboard events to one window
// and its descendants only, which makes a toplevel modal.
func Example() {
	app, err := takigo.NewApp(takigo.Title("Grab"))
	if err != nil {
		log.Fatal(err)
	}
	open := button.New(app, "open", button.Text("Modal..."), button.Command(func() {
		dlg := toplevel.New(app, "dlg", toplevel.Title("Modal"), toplevel.TransientFor(app))
		closeDlg := func() {
			app.GrabManager().Release()
			dlg.Destroy()
		}
		pack.Pack(button.New(dlg, "ok", button.Text("OK"), button.Command(closeDlg)), pack.PadX(30), pack.PadY(20))
		dlg.OnClose(closeDlg)
		app.GrabManager().Set(dlg.Window(), false)
	}))
	pack.Pack(open, pack.PadX(20), pack.PadY(20))
	app.Run()
}
