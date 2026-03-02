// Demo: Horizontal paned window with colored panes.
// Ported from Tk's paned1.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	app := demohelper.Setup("Horizontal Panes", 500, 300,
		"A horizontal paned window. Drag the sash to resize panes.")

	// Paned window.
	pw := panedwindow.New(app, "panes",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Three colored panes.
	colors := []struct {
		name, color, text string
	}{
		{"left", "#e74c3c", "Left\n(Red)"},
		{"center", "#3498db", "Center\n(Blue)"},
		{"right", "#2ecc71", "Right\n(Green)"},
	}

	for _, c := range colors {
		f := frame.New(pw, c.name, frame.Background(c.color))
		l := label.New(f, c.name+"_label",
			label.Text(c.text),
			label.Background(c.color),
			label.Foreground("white"),
		)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
			pack.Expand(true), pack.PadX(10), pack.PadY(10))
		pw.Add(f.Window(), 80)
		_ = l
	}

	app.Run()
}
