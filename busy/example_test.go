package busy_test

import (
	"log"
	"time"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/busy"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/label"
)

// Hold blocks the mouse and keyboard over a window, here the main window,
// while a goroutine works; RunOnMain releases it on the event loop.
func ExampleHold() {
	app, err := takigo.NewApp(takigo.Title("Busy"))
	if err != nil {
		log.Fatal(err)
	}
	status := label.New(app, "status", label.Text("idle"))
	start := button.New(app, "start", button.Text("Work"), button.Command(func() {
		hold := busy.Hold(app, app.Window())
		status.Configure(label.Text("working..."))
		go func() {
			time.Sleep(2 * time.Second)
			app.RunOnMain(func() {
				hold.Release()
				status.Configure(label.Text("done"))
			})
		}()
	}))
	pack.Pack(start, pack.PadX(20), pack.PadY(10))
	pack.Pack(status, pack.PadY(10))
	app.Run()
}
