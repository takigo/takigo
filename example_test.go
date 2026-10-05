package takigo_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/label"
)

// The examples have no "Output:" comment: they open a window, so go test
// compiles them without running them.

func Example() {
	app, err := takigo.NewApp(takigo.Title("Hello"))
	if err != nil {
		log.Fatal(err)
	}
	b := button.New(app, "hello",
		button.Text("Hello, world"),
		button.Command(app.Quit),
	)
	pack.Pack(b, pack.PadX(20), pack.PadY(20))
	app.Run()
}

func Example_grid() {
	app, err := takigo.NewApp(takigo.Title("Form"))
	if err != nil {
		log.Fatal(err)
	}
	nameLabel := label.New(app, "nameLabel", label.Text("Name:"))
	name := entry.New(app, "name", entry.Width(30))
	ok := button.New(app, "ok", button.Text("OK"), button.Command(app.Quit))

	grid.Grid(nameLabel, grid.Row(0), grid.Column(0), grid.Sticky(grid.StickE))
	grid.Grid(name, grid.Row(0), grid.Column(1), grid.Sticky(grid.EW))
	grid.Grid(ok, grid.Row(1), grid.Column(1), grid.Sticky(grid.StickE))
	grid.ColumnConfigure(app, 1, grid.Weight(1))
	app.Run()
}

func Example_configure() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	clicks := 0
	var b *button.Button
	b = button.New(app, "counter", button.Text("Clicked 0 times"), button.Command(func() {
		clicks++
		b.Configure(button.Text(fmt.Sprintf("Clicked %d times", clicks)))
	}))
	pack.Pack(b)
	app.Run()
}

func Example_canvas() {
	app, err := takigo.NewApp(takigo.Title("Canvas"))
	if err != nil {
		log.Fatal(err)
	}
	c := canvas.New(app, "c", canvas.Width(300), canvas.Height(200), canvas.Background("white"))
	pack.Pack(c, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	c.CreateRectangle(20, 20, 140, 100, canvas.FillColor("light blue"), canvas.OutlineColor("navy"))
	c.CreateOval(160, 20, 280, 100, canvas.FillColor("gold"))
	c.CreateLine([]float64{20, 150, 150, 120, 280, 150}, canvas.Smooth(true))
	c.CreateText(150, 180, canvas.TextOpt("Hello"), canvas.AnchorOpt(option.AnchorCenter))
	app.Run()
}

func Example_ttk() {
	app, err := takigo.NewApp(takigo.Title("Themed"))
	if err != nil {
		log.Fatal(err)
	}
	l := ttk.NewLabel(app, "l", ttk.LabelText("A themed label"))
	b := ttk.NewButton(app, "b", ttk.ButtonText("Quit"), ttk.ButtonCommand(app.Quit))
	pack.Pack(l, pack.PadY(4))
	pack.Pack(b, pack.PadY(4))
	app.Run()
}

func Example_dialog() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	b := button.New(app, "quit", button.Text("Quit..."), button.Command(func() {
		res := dialog.ShowMessage(app,
			dialog.MsgTitle("Quit"),
			dialog.MsgMessage("Really quit?"),
			dialog.MsgType(dialog.MsgQuestion),
			dialog.MsgButtons(dialog.BtnYesNo),
		)
		if res == dialog.ResultYes {
			app.Quit()
		}
	}))
	pack.Pack(b)
	app.Run()
}

// A goroutine hands its result to the event loop with RunOnMain.
func ExampleApp_RunOnMain() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	status := label.New(app, "status", label.Text("working..."))
	pack.Pack(status)
	go func() {
		result := "done"
		app.RunOnMain(func() { status.Configure(label.Text(result)) })
	}()
	app.Run()
}
