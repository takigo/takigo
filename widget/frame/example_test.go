package frame_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
)

// A frame groups widgets so they are laid out together.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Frame"))
	if err != nil {
		log.Fatal(err)
	}
	toolbar := frame.New(app, "toolbar", frame.Relief(option.ReliefRaised), frame.BorderWidth(1))
	pack.Pack(toolbar, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	for _, name := range []string{"open", "save", "close"} {
		b := button.New(toolbar, name, button.Text(name))
		pack.Pack(b, pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))
	}
	body := frame.New(app, "body", frame.Width(300), frame.Height(200), frame.Background("white"))
	pack.Pack(body, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	app.Run()
}
