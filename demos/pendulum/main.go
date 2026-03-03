// Demo: Pendulum physics simulation on canvas with phase space graph.
// Ported from Tk's pendulum.tcl demo.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
)

func main() {
	app := demohelper.Setup("Pendulum Simulation", 820, 500, "This demonstration shows how animations can be linked to simulations of physical systems. In the left canvas is a graphical representation of a simple pendulum, and in the right canvas is a graph of the phase space of the system (angle vs angular velocity). The pendulum bob may be repositioned by clicking and dragging anywhere on the left canvas.")

	// Container frame to hold both canvases side by side.
	container := frame.New(app, "container")
	pack.Pack(container, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	// Left canvas: pendulum visualization.
	pendulumCanvas := canvas.New(container, "pendulum",
		canvas.Background("#1a1a2e"),
		canvas.Width(370),
		canvas.Height(370),
	)
	pack.Pack(pendulumCanvas, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	// Right canvas: phase space graph.
	phaseCanvas := canvas.New(container, "phase",
		canvas.Background("#0d0d1a"),
		canvas.Width(370),
		canvas.Height(370),
	)
	pack.Pack(phaseCanvas, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	// Pendulum parameters.
	pivotX, pivotY := 185.0, 50.0
	length := 200.0
	bobRadius := 15.0
	theta := math.Pi / 4 // Initial angle (radians).
	omega := 0.0         // Angular velocity.
	gravity := 9.8
	dt := 0.05
	damping := 0.998

	// Phase space parameters.
	phaseCX := 185.0 // Center of phase canvas.
	phaseCY := 185.0
	phaseScaleX := 80.0  // Pixels per radian for angle.
	phaseScaleY := 8.0   // Pixels per (rad/s) for angular velocity.
	const maxPhasePoints = 500

	// Phase space trail: ring buffer of (x,y) canvas coordinates.
	type phasePoint struct{ x, y float64 }
	phaseTrail := make([]phasePoint, 0, maxPhasePoints)

	// Draw phase space axes (static).
	// Vertical axis (angular velocity).
	phaseCanvas.CreateLine([]float64{phaseCX, 10, phaseCX, 360},
		canvas.OutlineColor("#444466"), canvas.OutlineWidth(1), canvas.Tags("axis"))
	// Horizontal axis (angle).
	phaseCanvas.CreateLine([]float64{10, phaseCY, 360, phaseCY},
		canvas.OutlineColor("#444466"), canvas.OutlineWidth(1), canvas.Tags("axis"))
	// Axis labels.
	phaseCanvas.CreateText(355, phaseCY+15,
		canvas.TextOpt("θ"), canvas.TextColor("#888888"),
		canvas.AnchorOpt(option.AnchorE), canvas.Tags("axis"))
	phaseCanvas.CreateText(phaseCX+15, 15,
		canvas.TextOpt("dθ/dt"), canvas.TextColor("#888888"),
		canvas.AnchorOpt(option.AnchorW), canvas.Tags("axis"))

	// Pivot point on pendulum canvas.
	pendulumCanvas.CreateOval(pivotX-4, pivotY-4, pivotX+4, pivotY+4,
		canvas.FillColor("#aaaaaa"), canvas.OutlineColor("#888888"))
	// Plate line.
	pendulumCanvas.CreateLine([]float64{0, pivotY, 370, pivotY},
		canvas.OutlineColor("#555555"), canvas.OutlineWidth(1), canvas.Tags("plate"))

	// Instruction text.
	pendulumCanvas.CreateText(5, 5,
		canvas.TextOpt("Click to adjust bob"), canvas.TextColor("#666688"),
		canvas.AnchorOpt(option.AnchorNW), canvas.Tags("instr"))

	// dragging tracks whether the user is dragging the bob.
	dragging := false
	running := true

	// showPendulum repositions the pendulum from angle or from click position.
	showPendulum := func() {
		bobX := pivotX + length*math.Sin(theta)
		bobY := pivotY + length*math.Cos(theta)

		pendulumCanvas.Delete("rod")
		pendulumCanvas.Delete("bob")
		pendulumCanvas.Delete("trail")

		// Trail shadow.
		for i := 1; i <= 5; i++ {
			t := theta - omega*dt*float64(i)
			tx := pivotX + length*math.Sin(t)
			ty := pivotY + length*math.Cos(t)
			pendulumCanvas.CreateOval(tx-bobRadius*0.5, ty-bobRadius*0.5,
				tx+bobRadius*0.5, ty+bobRadius*0.5,
				canvas.FillColor("#333355"), canvas.Tags("trail"))
		}

		// Rod.
		pendulumCanvas.CreateLine([]float64{pivotX, pivotY, bobX, bobY},
			canvas.OutlineColor("#cccccc"), canvas.OutlineWidth(2), canvas.Tags("rod"))

		// Bob.
		pendulumCanvas.CreateOval(bobX-bobRadius, bobY-bobRadius,
			bobX+bobRadius, bobY+bobRadius,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("#c0392b"),
			canvas.OutlineWidth(2), canvas.Tags("bob"))
	}

	// showPhase draws the phase space trail.
	showPhase := func() {
		// Compute phase space canvas coordinates.
		px := phaseCX + theta*phaseScaleX
		py := phaseCY - omega*phaseScaleY

		phaseTrail = append(phaseTrail, phasePoint{px, py})
		if len(phaseTrail) > maxPhasePoints {
			phaseTrail = phaseTrail[1:]
		}

		// Remove old trail lines.
		phaseCanvas.Delete("trail")

		// Draw trail with fading colors — newer segments are brighter.
		n := len(phaseTrail)
		if n < 2 {
			return
		}

		// Draw segments in batches for performance: use a few color levels.
		const colorLevels = 10
		batchSize := n / colorLevels
		if batchSize < 2 {
			batchSize = 2
		}

		for i := 1; i < n; i++ {
			// Determine brightness: 0.0 (oldest) to 1.0 (newest).
			frac := float64(i) / float64(n)
			r := int(40 + frac*200)
			g := int(40 + frac*80)
			b := int(80 + frac*175)
			if r > 255 {
				r = 255
			}
			if g > 255 {
				g = 255
			}
			if b > 255 {
				b = 255
			}
			col := fmt.Sprintf("#%02x%02x%02x", r, g, b)

			// Only draw every Nth segment for older parts to save performance.
			skip := 1
			if frac < 0.3 {
				skip = 4
			} else if frac < 0.6 {
				skip = 2
			}
			if i%skip != 0 && i != n-1 {
				continue
			}

			prev := i - skip
			if prev < 0 {
				prev = 0
			}
			phaseCanvas.CreateLine(
				[]float64{phaseTrail[prev].x, phaseTrail[prev].y,
					phaseTrail[i].x, phaseTrail[i].y},
				canvas.OutlineColor(col), canvas.OutlineWidth(2), canvas.Tags("trail"))
		}

		// Draw current position dot.
		phaseCanvas.CreateOval(px-3, py-3, px+3, py+3,
			canvas.FillColor("#ff6644"), canvas.OutlineColor("#ff6644"), canvas.Tags("trail"))
	}

	// Click-and-drag to reposition pendulum bob.
	disp := app.Dispatcher()
	disp.Bind(pendulumCanvas.Win.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		// Scroll events (button 4-7) should not trigger drag.
		if ev.Button >= 4 {
			return
		}
		dragging = true
		running = false
		x := float64(ev.X)
		y := float64(ev.Y)
		dx := x - pivotX
		dy := y - pivotY
		length = math.Hypot(dx, dy)
		if length < 20 {
			length = 20
		}
		theta = math.Atan2(dx, dy)
		omega = 0
		// Clear phase trail on drag.
		phaseTrail = phaseTrail[:0]
		showPendulum()
	})

	disp.Bind(pendulumCanvas.Win.XWindow, event.MotionMask, func(ev *event.Event) {
		if !dragging {
			return
		}
		x := float64(ev.X)
		y := float64(ev.Y)
		dx := x - pivotX
		dy := y - pivotY
		length = math.Hypot(dx, dy)
		if length < 20 {
			length = 20
		}
		theta = math.Atan2(dx, dy)
		omega = 0
		showPendulum()
	})

	disp.Bind(pendulumCanvas.Win.XWindow, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button >= 4 {
			return
		}
		dragging = false
		running = true
	})

	// Animation loop.
	var animate func()
	animate = func() {
		if running && !dragging {
			// Physics update (simple pendulum).
			alpha := -gravity / length * math.Sin(theta)
			omega += alpha * dt
			omega *= damping
			theta += omega * dt

			showPendulum()
			showPhase()
		}

		app.After(20*time.Millisecond, animate)
	}
	app.After(100*time.Millisecond, animate)

	showPendulum()
	app.Run()
}
