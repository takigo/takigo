// Demo: Ruler with tab stops on canvas.
// Ported from Tk's ruler.tcl demo (simplified).
package main

import (
	"fmt"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
)

func main() {
	d := demohelper.Setup("Ruler Demo", 600, 250,
		"A ruler with tick marks and tab stops.\nThe tab stops are represented by small triangles.")
	root, app := d.Root, d.App

	// Canvas.
	c := canvas.New(root, "ruler", app,
		canvas.Background("#ffffee"),
		canvas.Width(560),
		canvas.Height(100),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	// Ruler background.
	c.CreateRectangle(20, 20, 540, 60,
		canvas.FillColor("#f0f0e0"), canvas.OutlineColor("black"), canvas.OutlineWidth(1))

	// Draw tick marks (1/8 inch marks at ~7.5px each for 72 DPI).
	rulerLeft := 20.0
	rulerTop := 20.0
	rulerBottom := 60.0
	ppi := 72.0 // pixels per inch

	for i := range int(7*8) + 1 { // 7 inches, 8 ticks per inch
		x := rulerLeft + float64(i)*ppi/8
		var tickLen float64
		switch {
		case i%8 == 0: // inch mark
			tickLen = rulerBottom - rulerTop
		case i%4 == 0: // half inch
			tickLen = (rulerBottom - rulerTop) * 0.6
		case i%2 == 0: // quarter inch
			tickLen = (rulerBottom - rulerTop) * 0.4
		default: // eighth inch
			tickLen = (rulerBottom - rulerTop) * 0.25
		}
		c.CreateLine([]float64{x, rulerBottom, x, rulerBottom - tickLen},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))

		// Inch labels.
		if i%8 == 0 && i > 0 {
			c.CreateText(x, rulerTop-5,
				canvas.TextOpt(fmt.Sprintf("%d", i/8)),
				canvas.FontOpt("Sans 8"), canvas.AnchorOpt(option.AnchorS))
		}
	}

	// Tab stop triangles.
	tabStops := []float64{1.0, 2.5, 4.0, 5.5} // inches
	for _, tab := range tabStops {
		x := rulerLeft + tab*ppi
		// Small downward-pointing triangle.
		c.CreatePolygon([]float64{x - 5, rulerBottom + 5, x + 5, rulerBottom + 5, x, rulerBottom + 15},
			canvas.FillColor("#cc0000"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("tab"))
	}

	d.Run()
}
