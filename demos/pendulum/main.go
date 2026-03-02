// Demo: Pendulum physics simulation on canvas.
// Ported from Tk's pendulum.tcl demo.
package main

import (
	"math"
	"time"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
)

func main() {
	app := demohelper.Setup("Pendulum", 400, 450, "A simple pendulum physics simulation.")

	c := canvas.New(app, "pendulum",
		canvas.Background("#1a1a2e"),
		canvas.Width(350),
		canvas.Height(350),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Pendulum parameters.
	pivotX, pivotY := 175.0, 50.0
	length := 200.0
	bobRadius := 15.0
	theta := math.Pi / 4 // Initial angle.
	omega := 0.0         // Angular velocity.
	gravity := 9.8
	dt := 0.05
	damping := 0.998

	// Pivot point.
	c.CreateOval(pivotX-4, pivotY-4, pivotX+4, pivotY+4,
		canvas.FillColor("#aaaaaa"), canvas.OutlineColor("#888888"))

	// Animation loop.
	var animate func()
	animate = func() {
		// Physics update (simple pendulum).
		alpha := -gravity / length * math.Sin(theta)
		omega += alpha * dt
		omega *= damping
		theta += omega * dt

		bobX := pivotX + length*math.Sin(theta)
		bobY := pivotY + length*math.Cos(theta)

		// Redraw.
		c.Delete("rod")
		c.Delete("bob")
		c.Delete("trail")

		// Trail shadow.
		for i := 1; i <= 5; i++ {
			t := theta - omega*dt*float64(i)
			tx := pivotX + length*math.Sin(t)
			ty := pivotY + length*math.Cos(t)
			alpha := 50 + i*20
			_ = alpha
			c.CreateOval(tx-bobRadius*0.5, ty-bobRadius*0.5, tx+bobRadius*0.5, ty+bobRadius*0.5,
				canvas.FillColor("#333355"), canvas.Tags("trail"))
		}

		// Rod.
		c.CreateLine([]float64{pivotX, pivotY, bobX, bobY},
			canvas.OutlineColor("#cccccc"), canvas.OutlineWidth(2), canvas.Tags("rod"))

		// Bob.
		c.CreateOval(bobX-bobRadius, bobY-bobRadius, bobX+bobRadius, bobY+bobRadius,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("#c0392b"), canvas.OutlineWidth(2),
			canvas.Tags("bob"))

		app.After(20*time.Millisecond, animate)
	}
	app.After(20*time.Millisecond, animate)

	app.Run()
}
