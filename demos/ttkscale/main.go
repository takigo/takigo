// Demo: Horizontal scale with label feedback.
// Ported from Tk's ttkscale.tcl demo.
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
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Themed Scale Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkscale"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("3.5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("A label tied to a horizontal scale is displayed below. If you click or drag mouse button 1 in the scale, you can change the contents of the label; a callback command is used to couple the slider to both the text and the coloring of the label."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.PadX(".5c"))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Rainbow color list — use X11 named colors matching Tcl.
	colorList := []string{"Red", "Orange", "Yellow", "Green", "Blue", "Violet"}

	// Inner frame with border (matches Tcl's `ttk::frame $w.frame -borderwidth 7.5p`).
	fr := frame.New(f, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Color label display — packed first (label before scale, matching Tcl).
	valueLabel := label.New(fr, "label")

	// Command callback to update label text and color from scale value.
	updateLabel := func(v float64) {
		idx := int(v)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(colorList) {
			idx = len(colorList) - 1
		}
		c := colorList[idx]
		valueLabel.Text = fmt.Sprintf("Color: %s", c)
		col, err := app.ColorCache().Get(c)
		if err == nil {
			valueLabel.Foreground = col
		}
		valueLabel.Display()
	}

	sc := scale.New(fr, "scale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(5),
		scale.CommandOpt(updateLabel),
	)
	// Trigger initial label text (matches Tcl's "$w.frame.scale set 0").
	updateLabel(0)
	pack.Pack(valueLabel)
	pack.Pack(sc)

	app.Run()
}
