package labelframe_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/checkbutton"
	"github.com/takigo/takigo/widget/labelframe"
)

// A labelframe draws a titled border around the widgets packed in it.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Labelframe"))
	if err != nil {
		log.Fatal(err)
	}
	lf := labelframe.New(app, "options", labelframe.Text("Options"), labelframe.PadX(8), labelframe.PadY(4))
	pack.Pack(lf, pack.PadX(10), pack.PadY(10), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	for _, name := range []string{"Wrap", "Bold", "Italic"} {
		c := checkbutton.New(lf, name, checkbutton.Text(name), checkbutton.BoolVar(widget.NewVariable(false)))
		pack.Pack(c, pack.SideOpt(pack.Top), pack.Anchor(option.AnchorW))
	}
	app.Run()
}
