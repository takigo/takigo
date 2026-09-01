// Demo: This demonstration script shows an example with a horizontal scale.
// Ported from Tk's hscale.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Horizontal Scale Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("hscale"),
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
		label.Text("An arrow and a horizontal scale are displayed below.  If you click or drag mouse button 1 in the scale, you can change the length of the arrow."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.PadX(".5c"))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Inner frame with border (matches Tcl's `frame -borderwidth 7.5p`).
	bw := screenunit.Px("7.5p")
	fr := frame.New(f, "frame",
		frame.BorderWidth(bw),
	)
	fr.SetInternalBorder(bw, bw, bw, bw)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Canvas for arrow display.
	c := canvas.New(fr, "canvas",
		canvas.Width(screenunit.Px("37.5p")),
		canvas.Height(screenunit.Px("37.5p")),
		canvas.BorderWidthOpt(0),
		canvas.HighlightWidthOpt(0),
		canvas.Background("#d9d9d9"),
	)

	// Create initial polygon and line items with dummy coords.
	c.CreatePolygon([]float64{0, 0, 1, 1, 2, 2},
		canvas.FillColor("deepskyblue"), canvas.Tags("poly"))
	c.CreateLine([]float64{0, 0, 1, 1, 2, 2, 0, 0},
		canvas.OutlineColor("black"), canvas.Tags("line"))

	// setWidth mirrors the Tk setWidth proc.
	setWidth := func(value float64) {
		width := value + 21
		x2 := width - 30
		if x2 < 21 {
			x2 = 21
		}
		coords := []float64{
			20, 15, 20, 35, x2, 35, x2, 45, width, 25, x2, 5, x2, 15, 20, 15,
		}
		c.SetItemCoords("poly", coords)
		c.SetItemCoords("line", coords)

		if sf := screenunit.ScalingFactor(); sf != 1.0 {
			c.Scale("poly", 0, 0, sf, sf)
			c.Scale("line", 0, 0, sf, sf)
		}
	}

	// Horizontal scale: 0-250, matching Tcl's -tickinterval 50 -length 213p.
	sc := scale.New(fr, "scale",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(0),
		scale.ToOpt(250),
		scale.ValueOpt(75),
		scale.TickIntervalOpt(50),
		scale.LengthOpt(screenunit.Px("213p")),
		scale.CommandOpt(func(v float64) {
			setWidth(v)
		}),
	)

	pack.Pack(c, pack.SideOpt(pack.Top), pack.Expand(true), pack.Anchor(option.AnchorS), pack.FillOpt(pack.FillX), pack.PadX("12p"))
	pack.Pack(sc, pack.SideOpt(pack.Bottom), pack.Expand(true), pack.Anchor(option.AnchorN))

	// Set initial arrow.
	setWidth(75)

	app.Run()
}
