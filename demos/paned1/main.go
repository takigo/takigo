// Demo: Horizontal paned window with colored panes.
// Ported from Tk's paned1.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	app := demohelper.Setup("Horizontal Paned Window Demonstration", 500, 300,
		"The sash between the two coloured windows below can be used to divide the area between them. Use the left mouse button to resize by moving the sash.")

	// Paned window.
	pw := panedwindow.New(app, "panes",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

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

	_ = leftLabel
	_ = rightLabel

	app.Run()
}
