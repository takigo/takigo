// Demo: A busy window that blocks input to a window while work is done
// (Tk's "tk busy"). No Tk demo of the same name exists.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/busy"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Busy Window Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("busy"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("The button below holds the window busy for three seconds: the cursor becomes a watch and the entry and the buttons ignore the mouse and the keyboard until the time is up."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	status := label.New(f, "status", label.Text("Ready"))
	pack.Pack(status, pack.SideOpt(pack.Bottom), pack.PadY(screenunit.Pt(3)))

	e := entry.New(f, "entry", entry.Width(30))
	e.SetText("Try typing here while busy")
	pack.Pack(e, pack.SideOpt(pack.Top), pack.PadY(screenunit.Pt(3)))

	var hold *busy.BusyWin
	b := button.New(f, "busy",
		button.Text("Busy for 3 seconds"),
		button.Command(func() {
			if hold != nil {
				return
			}
			status.Configure(label.Text("Busy..."))
			hold = busy.Hold(app, app.Window())
			app.After(3*time.Second, func() {
				hold.Release()
				hold = nil
				status.Configure(label.Text("Released"))
			})
		}),
	)
	pack.Pack(b, pack.SideOpt(pack.Top), pack.PadY(screenunit.Pt(3)))

	app.Run()
}
