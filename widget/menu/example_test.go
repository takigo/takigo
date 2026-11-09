package menu_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/bind"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/menu"
)

// A menubar along the top of the main window holds a cascade per menu.
func ExampleNewMenubar() {
	app, err := takigo.NewApp(takigo.Title("Menu"))
	if err != nil {
		log.Fatal(err)
	}
	bar := menu.NewMenubar(app, "menubar")

	file := menu.New(app, "file")
	file.AddCommandAccel("Open...", "Ctrl+O", func() { log.Println("open") })
	file.AddSeparator()
	file.AddCommandUL("Quit", 0, app.Quit)
	bar.AddCascade("File", 0, file)

	view := menu.New(app, "view")
	view.AddCheckbutton("Status bar", true, func() { log.Println("toggle status bar") })
	bar.AddCascade("View", 0, view)

	pack.Pack(label.New(app, "body", label.Text("content")), pack.PadX(40), pack.PadY(40))
	app.Run()
}

// A menu without a menubar pops up where Post puts it, here at the pointer
// on a right click.
func ExampleMenu_Post() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	popup := menu.New(app, "popup", menu.TearOffOpt(false))
	popup.AddCommand("Copy", func() { log.Println("copy") })
	popup.AddCommand("Paste", func() { log.Println("paste") })
	l := label.New(app, "l", label.Text("Right-click me"))
	pack.Pack(l, pack.PadX(40), pack.PadY(40))
	app.Bind().BindWindow(l, "<Button-3>", func(ev *bind.EventData) bool {
		popup.Post(ev.Event.RootX, ev.Event.RootY)
		return true
	})
	app.Run()
}
