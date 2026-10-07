package button_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
)

func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Button"))
	if err != nil {
		log.Fatal(err)
	}
	b := button.New(app, "quit",
		button.Text("Quit"),
		button.Command(app.Quit),
	)
	pack.Pack(b, pack.PadX(10), pack.PadY(10))
	app.Run()
}

func ExampleButton_Configure() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	b := button.New(app, "b", button.Text("Start"))
	pack.Pack(b)
	b.Configure(button.Command(func() {
		b.Configure(button.Text("Running"), button.ReliefOpt(0))
	}))
	app.Run()
}
