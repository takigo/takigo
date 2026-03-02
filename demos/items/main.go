// Demo: All canvas item types showcase.
// Ported from Tk's items.tcl demo.
package main

import (
	"math"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
)

func main() {
	d := demohelper.Setup("Canvas Items", 700, 550,
		"A showcase of all canvas item types: rectangles, ovals,\nlines, polygons, arcs, and text.")
	root, app := d.Root, d.App

	// Canvas.
	c := canvas.New(root, "items", app,
		canvas.Background("white"),
		canvas.Width(660),
		canvas.Height(420),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Section 1: Rectangles.
	c.CreateText(110, 15, canvas.TextOpt("Rectangles"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateRectangle(20, 30, 100, 90,
		canvas.FillColor("#4a86c8"), canvas.OutlineColor("black"), canvas.OutlineWidth(2))

	c.CreateRectangle(120, 30, 200, 90,
		canvas.FillColor("#e8a835"), canvas.OutlineColor("black"), canvas.OutlineWidth(1))

	c.CreateRectangle(20, 100, 200, 130,
		canvas.FillColor("#6bb86b"), canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(3))

	// Section 2: Ovals.
	c.CreateText(330, 15, canvas.TextOpt("Ovals"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateOval(240, 30, 340, 90,
		canvas.FillColor("#c85a5a"), canvas.OutlineColor("black"), canvas.OutlineWidth(2))

	c.CreateOval(360, 30, 420, 130,
		canvas.FillColor("#9b59b6"), canvas.OutlineColor("black"), canvas.OutlineWidth(1))

	// Section 3: Lines.
	c.CreateText(550, 15, canvas.TextOpt("Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateLine([]float64{460, 30, 640, 30},
		canvas.OutlineColor("red"), canvas.OutlineWidth(3))

	c.CreateLine([]float64{460, 50, 550, 90, 640, 50},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2))

	c.CreateLine([]float64{460, 110, 550, 70, 640, 110},
		canvas.OutlineColor("darkgreen"), canvas.OutlineWidth(2), canvas.Smooth(true))

	c.CreateLine([]float64{460, 130, 640, 130},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.Arrow(canvas.ArrowBoth))

	// Section 4: Polygons.
	c.CreateText(110, 155, canvas.TextOpt("Polygons"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	// Triangle.
	c.CreatePolygon([]float64{60, 170, 20, 260, 100, 260},
		canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"), canvas.OutlineWidth(2))

	// Pentagon.
	cx, cy, r := 160.0, 220.0, 40.0
	penta := make([]float64, 10)
	for i := range 5 {
		angle := math.Pi/2 + float64(i)*2*math.Pi/5
		penta[i*2] = cx + r*math.Cos(angle)
		penta[i*2+1] = cy - r*math.Sin(angle)
	}
	c.CreatePolygon(penta,
		canvas.FillColor("#3498db"), canvas.OutlineColor("navy"), canvas.OutlineWidth(2))

	// Section 5: Arcs.
	c.CreateText(330, 155, canvas.TextOpt("Arcs"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateArc(240, 170, 340, 270,
		canvas.FillColor("#f39c12"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
		canvas.StartAngle(0), canvas.Extent(120), canvas.ArcStyleOpt(canvas.ArcStylePieslice))

	c.CreateArc(350, 170, 430, 270,
		canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(3),
		canvas.StartAngle(30), canvas.Extent(270), canvas.ArcStyleOpt(canvas.ArcStyleArc))

	// Section 6: Text.
	c.CreateText(550, 155, canvas.TextOpt("Text Items"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateText(550, 190, canvas.TextOpt("Default font"),
		canvas.TextColor("black"), canvas.AnchorOpt(option.AnchorCenter))

	c.CreateText(550, 220, canvas.TextOpt("Bold text"),
		canvas.FontOpt("Sans Bold 14"), canvas.TextColor("navy"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateText(550, 250, canvas.TextOpt("Italic text"),
		canvas.FontOpt("Sans Italic 12"), canvas.TextColor("darkred"),
		canvas.AnchorOpt(option.AnchorCenter))

	// Section 7: Dashed lines.
	c.CreateText(330, 290, canvas.TextOpt("Dashed Lines"), canvas.FontOpt("Sans Bold 11"),
		canvas.AnchorOpt(option.AnchorCenter))

	c.CreateLine([]float64{20, 310, 300, 310},
		canvas.OutlineColor("black"), canvas.OutlineWidth(2), canvas.Dash(6, 4))

	c.CreateLine([]float64{20, 330, 300, 330},
		canvas.OutlineColor("blue"), canvas.OutlineWidth(2), canvas.Dash(12, 4, 4, 4))

	c.CreateRectangle(340, 300, 640, 380,
		canvas.OutlineColor("gray50"), canvas.OutlineWidth(2), canvas.Dash(8, 4))

	c.CreateOval(380, 305, 600, 375,
		canvas.FillColor("#eaf2f8"), canvas.OutlineColor("#2980b9"), canvas.OutlineWidth(2))

	d.Run()
}
