package label_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/label"
)

func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Label"))
	if err != nil {
		log.Fatal(err)
	}
	l := label.New(app, "msg",
		label.Text("A label wraps long text at -wraplength and justifies the lines."),
		label.WrapLength(screenunit.In(3)),
		label.JustifyOpt(option.JustifyLeft),
		label.Relief(option.ReliefGroove),
		label.BorderWidth(2),
	)
	pack.Pack(l, pack.PadX(10), pack.PadY(10))
	app.Run()
}
