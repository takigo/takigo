package checkbutton_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/checkbutton"
)

// A checkbutton linked to a Variable[bool] follows and drives it.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Checkbutton"))
	if err != nil {
		log.Fatal(err)
	}
	bold := widget.NewVariable(false)
	bold.OnChange(func(_, on bool) { log.Println("bold:", on) })
	c := checkbutton.New(app, "bold",
		checkbutton.Text("Bold"),
		checkbutton.BoolVar(bold),
	)
	pack.Pack(c)
	app.Run()
}
