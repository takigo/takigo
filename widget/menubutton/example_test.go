package menubutton_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/menu"
	"github.com/takigo/takigo/widget/menubutton"
)

// A menubutton posts its menu, a child of the menubutton, when pressed.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Menubutton"))
	if err != nil {
		log.Fatal(err)
	}
	mb := menubutton.New(app, "file",
		menubutton.Text("File"),
		menubutton.UnderlineOpt(0),
		menubutton.Relief(option.ReliefRaised),
	)
	m := menu.New(mb, "m")
	m.AddCommand("New", func() { log.Println("new") })
	m.AddCommand("Quit", app.Quit)
	mb.Configure(menubutton.MenuOpt(m))
	pack.Pack(mb, pack.PadX(20), pack.PadY(20))
	app.Run()
}

// An option menu shows the chosen entry as its text.
func ExampleOptionMenuOpt() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	mb := menubutton.New(app, "size",
		menubutton.Text("medium"),
		menubutton.OptionMenuOpt(true),
		menubutton.IndicatorOnOpt(true),
	)
	m := menu.New(mb, "m")
	for _, v := range []string{"small", "medium", "large"} {
		m.AddCommand(v, func() { mb.SetText(v) })
	}
	mb.Configure(menubutton.MenuOpt(m))
	pack.Pack(mb, pack.PadX(20), pack.PadY(20))
	app.Run()
}
