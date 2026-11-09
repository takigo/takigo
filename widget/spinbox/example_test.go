package spinbox_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/spinbox"
)

// A numeric spinbox steps between -from and -to by -increment.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Spinbox"))
	if err != nil {
		log.Fatal(err)
	}
	s := spinbox.New(app, "count",
		spinbox.FromOpt(0),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%.1f"),
		spinbox.WidthOpt(8),
		spinbox.CommandOpt(func(v string) { log.Println("value:", v) }),
	)
	pack.Pack(s, pack.PadX(10), pack.PadY(10))
	app.Run()
}

// With -values the spinbox cycles through a fixed list.
func ExampleValuesOpt() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	s := spinbox.New(app, "day",
		spinbox.ValuesOpt([]string{"Mon", "Tue", "Wed", "Thu", "Fri"}),
		spinbox.WrapOpt(true),
		spinbox.ReadOnly(true),
	)
	pack.Pack(s)
	s.SpinUp()
	log.Println("day:", s.GetText())
	app.Run()
}
