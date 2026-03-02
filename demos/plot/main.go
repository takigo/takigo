// Demo: 2D plot with draggable data points.
// Ported from Tk's plot.tcl demo.
package main

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
)

func main() {
	d := demohelper.Setup("2D Plot", 550, 450,
		"A 2D data plot. Drag the data points with the mouse\nto see them move along the line.")
	app := d.App

	// Canvas.
	c := canvas.New(app, "plot",
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(350),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Plot area dimensions.
	const (
		plotLeft   = 60
		plotRight  = 480
		plotTop    = 40
		plotBottom = 310
	)

	// Draw axes.
	c.CreateLine([]float64{plotLeft, plotBottom, plotRight, plotBottom},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2))
	c.CreateLine([]float64{plotLeft, plotBottom, plotLeft, plotTop},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2))

	// Axis labels.
	c.CreateText(float64(plotLeft+plotRight)/2, float64(plotBottom+25),
		canvas.TextOpt("X Axis"), canvas.AnchorOpt(option.AnchorCenter))
	c.CreateText(float64(plotLeft-35), float64(plotTop+plotBottom)/2,
		canvas.TextOpt("Y"), canvas.AnchorOpt(option.AnchorCenter))

	// Tick marks and labels.
	for i := range 6 {
		x := plotLeft + i*(plotRight-plotLeft)/5
		c.CreateLine([]float64{float64(x), float64(plotBottom), float64(x), float64(plotBottom + 5)},
			canvas.OutlineColor("black"))
		c.CreateText(float64(x), float64(plotBottom+12),
			canvas.TextOpt(fmt.Sprintf("%d", i*20)),
			canvas.AnchorOpt(option.AnchorCenter), canvas.FontOpt("Sans 8"))
	}
	for i := range 6 {
		y := plotBottom - i*(plotBottom-plotTop)/5
		c.CreateLine([]float64{float64(plotLeft - 5), float64(y), float64(plotLeft), float64(y)},
			canvas.OutlineColor("black"))
		c.CreateText(float64(plotLeft-10), float64(y),
			canvas.TextOpt(fmt.Sprintf("%d", i*20)),
			canvas.AnchorOpt(option.AnchorE), canvas.FontOpt("Sans 8"))
	}

	// Data points and line.
	dataX := []float64{0, 20, 40, 60, 80, 100}
	dataY := []float64{10, 45, 30, 70, 55, 90}

	// Convert data to pixel coordinates.
	toPixelX := func(v float64) float64 {
		return plotLeft + v/100*(plotRight-plotLeft)
	}
	toPixelY := func(v float64) float64 {
		return plotBottom - v/100*(plotBottom-plotTop)
	}

	// Draw connecting line.
	lineCoords := make([]float64, len(dataX)*2)
	for i := range dataX {
		lineCoords[i*2] = toPixelX(dataX[i])
		lineCoords[i*2+1] = toPixelY(dataY[i])
	}
	c.CreateLine(lineCoords,
		canvas.OutlineColor("#3498db"), canvas.OutlineWidth(2),
		canvas.Tags("dataline"))

	// Draw data points as small filled circles.
	ptSize := 5.0
	for i := range dataX {
		px := toPixelX(dataX[i])
		py := toPixelY(dataY[i])
		c.CreateOval(px-ptSize, py-ptSize, px+ptSize, py+ptSize,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("point"))
	}

	// Draw grid lines.
	for i := 1; i < 5; i++ {
		x := plotLeft + i*(plotRight-plotLeft)/5
		c.CreateLine([]float64{float64(x), float64(plotTop), float64(x), float64(plotBottom)},
			canvas.OutlineColor("#e0e0e0"), canvas.OutlineWidth(1))
		y := plotBottom - i*(plotBottom-plotTop)/5
		c.CreateLine([]float64{float64(plotLeft), float64(y), float64(plotRight), float64(y)},
			canvas.OutlineColor("#e0e0e0"), canvas.OutlineWidth(1))
	}

	// Sine curve overlay.
	sineCoords := make([]float64, 0, 202)
	for i := range 101 {
		x := float64(i)
		y := 50 + 40*math.Sin(x*math.Pi/50)
		sineCoords = append(sineCoords, toPixelX(x), toPixelY(y))
	}
	c.CreateLine(sineCoords,
		canvas.OutlineColor("#27ae60"), canvas.OutlineWidth(1), canvas.Smooth(true))

	d.Run()
}
