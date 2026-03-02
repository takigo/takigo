// Demo: Rube Goldberg machine — complex canvas animation.
// Ported from Tk's goldberg.tcl demo (greatly simplified).
package main

import (
	"math"
	"time"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	d := demohelper.Setup("Goldberg Machine", 600, 500, "A simplified Rube Goldberg machine.\nClick Start to begin the chain reaction.")
	app := d.App

	c := canvas.New(app, "goldberg",
		canvas.Background("#2c3e50"),
		canvas.Width(560),
		canvas.Height(380),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Draw static elements.
	// Ramp 1 (top left).
	c.CreateLine([]float64{50, 80, 200, 130},
		canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4))
	// Ramp 2 (middle right).
	c.CreateLine([]float64{350, 150, 200, 200},
		canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4))
	// Ramp 3 (bottom left).
	c.CreateLine([]float64{100, 250, 300, 300},
		canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4))

	// Platforms.
	c.CreateRectangle(180, 195, 220, 205,
		canvas.FillColor("#7f8c8d"), canvas.OutlineColor("#7f8c8d"))
	c.CreateRectangle(280, 295, 320, 305,
		canvas.FillColor("#7f8c8d"), canvas.OutlineColor("#7f8c8d"))

	// Funnel at bottom.
	c.CreatePolygon([]float64{350, 300, 400, 300, 385, 350, 365, 350},
		canvas.FillColor("#e67e22"), canvas.OutlineColor("#d35400"), canvas.OutlineWidth(2))

	// Target (star shape at bottom).
	starCoords := make([]float64, 20)
	sx, sy := 375.0, 370.0
	for i := range 10 {
		angle := float64(i)*math.Pi/5 - math.Pi/2
		r := 15.0
		if i%2 == 1 {
			r = 7.0
		}
		starCoords[i*2] = sx + r*math.Cos(angle)
		starCoords[i*2+1] = sy + r*math.Sin(angle)
	}
	c.CreatePolygon(starCoords,
		canvas.FillColor("#f1c40f"), canvas.OutlineColor("#f39c12"), canvas.OutlineWidth(1),
		canvas.Tags("star"))

	// Title.
	c.CreateText(280, 25,
		canvas.TextOpt("Rube Goldberg Machine"),
		canvas.FontOpt("Sans Bold 16"), canvas.TextColor("#ecf0f1"),
		canvas.AnchorOpt(option.AnchorCenter))

	// Ball animation.
	ballRadius := 8.0
	type segment struct {
		x1, y1, x2, y2 float64
		steps           int
	}
	path := []segment{
		{50, 72, 200, 122, 40},   // Roll down ramp 1.
		{200, 122, 200, 187, 15}, // Fall to platform 1.
		{350, 142, 200, 192, 40}, // Roll down ramp 2 (reversed).
		{200, 192, 200, 242, 15}, // Fall.
		{100, 242, 300, 292, 40}, // Roll down ramp 3.
		{300, 292, 375, 340, 20}, // Fall into funnel.
	}

	step := 0
	segIdx := 0
	segStep := 0

	var animate func()
	animate = func() {
		if segIdx >= len(path) {
			// Final: flash the star.
			c.Delete("ball")
			for i := range 5 {
				color := "#f1c40f"
				if i%2 == 1 {
					color = "#e74c3c"
				}
				_ = color
			}
			return
		}

		seg := path[segIdx]
		t := float64(segStep) / float64(seg.steps)
		bx := seg.x1 + (seg.x2-seg.x1)*t
		by := seg.y1 + (seg.y2-seg.y1)*t

		c.Delete("ball")
		c.CreateOval(bx-ballRadius, by-ballRadius, bx+ballRadius, by+ballRadius,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(2),
			canvas.Tags("ball"))

		segStep++
		if segStep > seg.steps {
			segIdx++
			segStep = 0
		}
		step++

		app.After(30*time.Millisecond, animate)
	}

	startBtn := button.New(app, "start",
		button.Text("Start"),
		button.Command(func() {
			step = 0
			segIdx = 0
			segStep = 0
			c.Delete("ball")
			animate()
		}),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(startBtn, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	_ = startBtn
	d.Run()
}
