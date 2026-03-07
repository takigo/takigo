// Demo: Rube Goldberg machine — complex canvas animation.
// Ported from Tk's goldberg.tcl demo (greatly simplified).
package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Goldberg Machine"),
		takigo.Geometry("+300+300"),
		takigo.IconName("goldberg"),
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
		label.Text("A simplified Rube Goldberg machine animation. Click Start to begin the chain reaction.\nA ball rolls down ramps, bounces off platforms, and lands in a funnel to hit the target star."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "goldberg",
		canvas.Background("#2c3e50"),
		canvas.Width(600),
		canvas.Height(400),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Draw static scenery.
	drawScene := func() {
		c.Delete("scene")

		// Ramp 1 (top left, slopes down-right).
		c.CreateLine([]float64{50, 80, 200, 130},
			canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4), canvas.Tags("scene"))
		// Ramp supports.
		c.CreateLine([]float64{50, 80, 50, 135},
			canvas.OutlineColor("#7f8c8d"), canvas.OutlineWidth(2), canvas.Tags("scene"))
		c.CreateLine([]float64{125, 105, 125, 135},
			canvas.OutlineColor("#7f8c8d"), canvas.OutlineWidth(2), canvas.Tags("scene"))

		// Platform 1 (catches ball from ramp 1).
		c.CreateRectangle(190, 130, 230, 140,
			canvas.FillColor("#7f8c8d"), canvas.OutlineColor("#7f8c8d"), canvas.Tags("scene"))

		// Ramp 2 (middle, slopes down-left).
		c.CreateLine([]float64{220, 140, 100, 210},
			canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4), canvas.Tags("scene"))
		// Ramp 2 supports.
		c.CreateLine([]float64{160, 175, 160, 215},
			canvas.OutlineColor("#7f8c8d"), canvas.OutlineWidth(2), canvas.Tags("scene"))

		// Platform 2 (catches ball from ramp 2).
		c.CreateRectangle(80, 210, 130, 220,
			canvas.FillColor("#7f8c8d"), canvas.OutlineColor("#7f8c8d"), canvas.Tags("scene"))

		// Ramp 3 (bottom, slopes down-right).
		c.CreateLine([]float64{100, 220, 350, 300},
			canvas.OutlineColor("#95a5a6"), canvas.OutlineWidth(4), canvas.Tags("scene"))
		// Ramp 3 supports.
		c.CreateLine([]float64{175, 245, 175, 305},
			canvas.OutlineColor("#7f8c8d"), canvas.OutlineWidth(2), canvas.Tags("scene"))
		c.CreateLine([]float64{260, 275, 260, 305},
			canvas.OutlineColor("#7f8c8d"), canvas.OutlineWidth(2), canvas.Tags("scene"))

		// Funnel at bottom right.
		c.CreatePolygon([]float64{340, 300, 420, 300, 400, 350, 360, 350},
			canvas.FillColor("#e67e22"), canvas.OutlineColor("#d35400"),
			canvas.OutlineWidth(2), canvas.Tags("scene"))
		// Funnel tube.
		c.CreateRectangle(368, 350, 392, 365,
			canvas.FillColor("#d35400"), canvas.OutlineColor("#c0392b"), canvas.Tags("scene"))

		// Target star at bottom.
		starCoords := make([]float64, 20)
		sx, sy := 380.0, 385.0
		for i := range 10 {
			angle := float64(i)*math.Pi/5 - math.Pi/2
			r := 18.0
			if i%2 == 1 {
				r = 8.0
			}
			starCoords[i*2] = sx + r*math.Cos(angle)
			starCoords[i*2+1] = sy + r*math.Sin(angle)
		}
		c.CreatePolygon(starCoords,
			canvas.FillColor("#f1c40f"), canvas.OutlineColor("#f39c12"),
			canvas.OutlineWidth(1), canvas.Tags("star"))

		// Decorative gears (static circles).
		for _, g := range [][3]float64{{450, 100, 25}, {490, 140, 18}, {520, 80, 15}} {
			gx, gy, gr := g[0], g[1], g[2]
			c.CreateOval(gx-gr, gy-gr, gx+gr, gy+gr,
				canvas.OutlineColor("#5d6d7e"), canvas.OutlineWidth(3), canvas.Tags("scene"))
			// Gear spokes.
			for s := 0; s < 4; s++ {
				a := float64(s) * math.Pi / 4
				c.CreateLine([]float64{
					gx - gr*0.7*math.Cos(a), gy - gr*0.7*math.Sin(a),
					gx + gr*0.7*math.Cos(a), gy + gr*0.7*math.Sin(a),
				}, canvas.OutlineColor("#5d6d7e"), canvas.OutlineWidth(2), canvas.Tags("scene"))
			}
		}

		// Title.
		c.CreateText(300, 25,
			canvas.TextOpt("Rube Goldberg Machine"),
			canvas.FontOpt("Sans Bold 16"), canvas.TextColor("#ecf0f1"),
			canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("scene"))

		// "Ready" indicator.
		c.CreateText(300, 390,
			canvas.TextOpt("Press Start to begin"),
			canvas.FontOpt("Sans 10"), canvas.TextColor("#7f8c8d"),
			canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("status"))
	}

	drawScene()

	// Animation path: ball follows connected segments along the ramps.
	ballRadius := 8.0
	type segment struct {
		x1, y1, x2, y2 float64
		steps           int
	}
	// Segments are carefully connected: each starts where the previous ended.
	path := []segment{
		{50, 72, 200, 122, 40},   // Roll down ramp 1.
		{200, 122, 210, 132, 8},  // Short drop to platform 1.
		{220, 132, 100, 202, 40}, // Roll down ramp 2.
		{100, 202, 105, 212, 8},  // Short drop to platform 2.
		{100, 212, 350, 292, 50}, // Roll down ramp 3.
		{350, 292, 380, 345, 15}, // Drop into funnel.
		{380, 345, 380, 365, 10}, // Fall through funnel tube.
	}

	segIdx := 0
	segStep := 0
	running := false

	// Flash counter for end animation.
	flashCount := 0

	var animate func()
	animate = func() {
		if segIdx >= len(path) {
			// Final: flash the star and show completion text.
			c.Delete("ball")

			if flashCount < 10 {
				color := "#f1c40f"
				outline := "#f39c12"
				if flashCount%2 == 1 {
					color = "#e74c3c"
					outline = "#c0392b"
				}
				c.ItemConfigure("star", canvas.FillColor(color), canvas.OutlineColor(outline))

				flashCount++
				app.After(150*time.Millisecond, animate)
			} else {
				// Restore star color and show done message.
				c.ItemConfigure("star", canvas.FillColor("#2ecc71"), canvas.OutlineColor("#27ae60"))
				c.Delete("status")
				c.CreateText(300, 390,
					canvas.TextOpt("Done! Press Reset to try again."),
					canvas.FontOpt("Sans 10"), canvas.TextColor("#2ecc71"),
					canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("status"))
				running = false
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

		// Update status with segment info.
		c.Delete("status")
		segName := ""
		switch {
		case segIdx < 1:
			segName = "Rolling down ramp 1..."
		case segIdx < 2:
			segName = "Dropping to platform..."
		case segIdx < 3:
			segName = "Rolling down ramp 2..."
		case segIdx < 4:
			segName = "Dropping to platform..."
		case segIdx < 5:
			segName = "Rolling down ramp 3..."
		case segIdx < 7:
			segName = "Falling into funnel..."
		default:
			segName = "Almost there..."
		}
		c.CreateText(300, 390,
			canvas.TextOpt(fmt.Sprintf("Step %d: %s", segIdx+1, segName)),
			canvas.FontOpt("Sans 10"), canvas.TextColor("#bdc3c7"),
			canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("status"))

		app.After(30*time.Millisecond, animate)
	}

	reset := func() {
		running = false
		segIdx = 0
		segStep = 0
		flashCount = 0
		c.Delete("ball")
		c.Delete("star")
		c.Delete("status")
		c.Delete("scene")
		drawScene()
	}

	// Button bar.
	btnBar := frame.New(f, "btnbar")
	pack.Pack(btnBar, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	startBtn := button.New(btnBar, "start",
		button.Text("Start"),
		button.Command(func() {
			if running {
				return
			}
			reset()
			running = true
			c.Delete("status")
			c.CreateText(300, 390,
				canvas.TextOpt("Starting chain reaction..."),
				canvas.FontOpt("Sans 10"), canvas.TextColor("#3498db"),
				canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("status"))
			app.After(500*time.Millisecond, animate)
		}),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(startBtn, pack.SideOpt(pack.Left), pack.PadX(5))

	resetBtn := button.New(btnBar, "reset",
		button.Text("Reset"),
		button.Command(func() {
			reset()
		}),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(resetBtn, pack.SideOpt(pack.Left), pack.PadX(5))

	app.Run()
}
