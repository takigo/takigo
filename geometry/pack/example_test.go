package pack_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
)

// Pack arranges children along a side of their parent, in packing order;
// a geometry.Group packs several with the same options.
func ExamplePack() {
	app, err := takigo.NewApp(takigo.Title("pack"))
	if err != nil {
		log.Fatal(err)
	}
	body := label.New(app, "body", label.Text("content"))
	pack.Pack(body, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	buttons := frame.New(app, "buttons")
	pack.Pack(buttons, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))
	ok := button.New(buttons, "ok", button.Text("OK"), button.Command(app.Quit))
	cancel := button.New(buttons, "cancel", button.Text("Cancel"), button.Command(app.Quit))
	pack.Pack(geometry.Group{cancel, ok}, pack.SideOpt(pack.Right), pack.PadX(4), pack.PadY(4))
	app.Run()
}
