package scale_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/scale"
)

// A scale's command receives the value whenever the slider moves.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Scale"))
	if err != nil {
		log.Fatal(err)
	}
	shown := label.New(app, "shown", label.Text("50%"))
	s := scale.New(app, "volume",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(100),
		scale.ResolutionOpt(5),
		scale.TickIntervalOpt(25),
		scale.LengthOpt(200),
		scale.LabelOpt("Volume"),
		scale.CommandOpt(func(v float64) { shown.Configure(label.Text(fmt.Sprintf("%.0f%%", v))) }),
	)
	s.Set(50)
	pack.Pack(s, pack.PadX(10), pack.PadY(4))
	pack.Pack(shown)
	app.Run()
}
