// Demo: Paned window that separates two windows horizontally.
// Ported from Tk's paned1.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Horizontal Paned Window Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("paned1"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("The sash between the two coloured windows below can be used to divide the area between them.  Use the left mouse button to resize without redrawing by just moving the sash, and use the middle mouse button to resize opaquely (always redrawing the windows in each position.)"),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Paned window.
	pw := panedwindow.New(f, "pane")
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("2m"), pack.PadY("1.5p"))

	// Two colored panes (match Tk: yellow left, cyan right).
	leftLabel := label.New(pw, "left",
		label.Text("This is the\nleft side"),
		label.Background("yellow"),
		label.Foreground("black"),
	)
	rightLabel := label.New(pw, "right",
		label.Text("This is the\nright side"),
		label.Background("cyan"),
		label.Foreground("black"),
	)
	pw.Add(leftLabel.Window(), 150)
	pw.Add(rightLabel.Window(), 150)

	app.Run()
}
