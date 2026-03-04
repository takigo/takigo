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
		pack.Expand(true))

	// Left canvas: pendulum visualization (matches Tcl: 240p x 150p ≈ 320x200).
	pendulumCanvas := canvas.New(container, "pendulum",
		canvas.Background("white"),
		canvas.Width(320),
		canvas.Height(200),
		canvas.BorderWidthOpt(2),
		canvas.ReliefOpt(option.ReliefSunken),
	)
	pack.Pack(pendulumCanvas, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Right canvas: phase space graph.
	phaseCanvas := canvas.New(container, "phase",
		canvas.Background("white"),
		canvas.Width(320),
		canvas.Height(200),
		canvas.BorderWidthOpt(2),
		canvas.ReliefOpt(option.ReliefSunken),
	)
	pack.Pack(phaseCanvas, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Pendulum parameters (scaled to match Tcl: 120p pivot height, 111p length).
	pivotX, pivotY := 160.0, 24.0
	length := 148.0 // 111p ≈ 148px
	bobRadius := 16.0
	theta := math.Pi / 4 // Initial angle (radians).
	omega := 0.0         // Angular velocity.
	gravity := 9.8
	dt := 0.05
	damping := 0.998

	// Phase space parameters.
	phaseCX := 160.0 // Center of phase canvas.
	phaseCY := 100.0
	phaseScaleX := 80.0  // Pixels per radian for angle.
	phaseScaleY := 8.0   // Pixels per (rad/s) for angular velocity.
	const maxPhasePoints = 500

	// Phase space trail: ring buffer of (x,y) canvas coordinates.
	type phasePoint struct{ x, y float64 }
	phaseTrail := make([]phasePoint, 0, maxPhasePoints)

	// Draw phase space axes (static).
	// Vertical axis (angular velocity).
	phaseCanvas.CreateLine([]float64{phaseCX, 190, phaseCX, 5},
		canvas.OutlineColor("grey75"), canvas.OutlineWidth(1), canvas.Tags("y_axis"))
	// Horizontal axis (angle).
	phaseCanvas.CreateLine([]float64{5, phaseCY, 315, phaseCY},
		canvas.OutlineColor("grey75"), canvas.OutlineWidth(1), canvas.Tags("x_axis"))
	// Axis labels.
	phaseCanvas.CreateText(phaseCX-3, 4,
		canvas.TextOpt("δθ"), canvas.TextColor("black"),
		canvas.AnchorOpt(option.AnchorE), canvas.Tags("label_dtheta"))
	phaseCanvas.CreateText(315, phaseCY+3,
		canvas.TextOpt("θ"), canvas.TextColor("black"),
		canvas.AnchorOpt(option.AnchorE), canvas.Tags("label_theta"))

	// Pivot point on pendulum canvas.
	pendulumCanvas.CreateOval(pivotX-4, pivotY-3, pivotX+4, pivotY+4,
		canvas.FillColor("grey50"), canvas.Tags("pivot"))
	// Plate line.
	pendulumCanvas.CreateLine([]float64{0, pivotY - 6, 320, pivotY - 6},
		canvas.OutlineColor("grey50"), canvas.OutlineWidth(2), canvas.Tags("plate"))

	// Instruction text.
	pendulumCanvas.CreateText(4, 4,
		canvas.TextOpt("Click to Adjust Bob Start Position"), canvas.TextColor("black"),
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

		// (no trail shadow — Tcl doesn't have one)

		// Rod.
		pendulumCanvas.CreateLine([]float64{pivotX, pivotY, bobX, bobY},
			canvas.OutlineColor("black"), canvas.OutlineWidth(3), canvas.Tags("rod"))

		// Bob.
		pendulumCanvas.CreateOval(bobX-bobRadius, bobY-bobRadius,
			bobX+bobRadius, bobY+bobRadius,
			canvas.FillColor("yellow"), canvas.OutlineColor("black"),
			canvas.OutlineWidth(1), canvas.Tags("bob"))
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

		// Draw trail with 10 grey levels (matching Tcl's grey0..grey90).
		n := len(phaseTrail)
		if n < 2 {
			return
		}
		for level := 0; level <= 90; level += 10 {
			// Each level covers a bucket of the oldest (grey0) to newest (grey90) points.
			startFrac := float64(level) / 100.0
			endFrac := float64(level+10) / 100.0
			start := int(startFrac * float64(n))
			end := int(endFrac * float64(n))
			if start >= n-1 {
				break
			}
			if end > n {
				end = n
			}
			if end-start < 2 {
				continue
			}
			col := fmt.Sprintf("grey%d", level)
			pts := make([]float64, 0, (end-start)*2)
			for i := start; i < end; i++ {
				pts = append(pts, phaseTrail[i].x, phaseTrail[i].y)
			}
			if len(pts) >= 4 {
				phaseCanvas.CreateLine(pts, canvas.OutlineColor(col),
					canvas.OutlineWidth(1), canvas.Tags("trail"))
			}
		}
	}

	// Click-and-drag to reposition pendulum bob.
	disp := app.Dispatcher()
	disp.Bind(pendulumCanvas.Win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
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

	disp.Bind(pendulumCanvas.Win.PlatformID, event.MotionMask, func(ev *event.Event) {
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

	disp.Bind(pendulumCanvas.Win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
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
