// Demo: Animated canvas waveform.
// Ported from Tk's aniwave.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

// scaledCoords returns a copy of coords with all values multiplied by the
// display scaling factor, matching Tcl's "$w.c scale wave 0 0 $sf $sf".
func scaledCoords(coords []float64) []float64 {
	sf := float64(screenunit.ScalingPct()) / 100.0
	out := make([]float64, len(coords))
	for i, v := range coords {
		out[i] = v * sf
	}
	return out
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Animated Wave Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("aniwave"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This demonstration contains a canvas widget with a line item inside it. The animation routines work by adjusting the coordinates list of the line; a trace on a variable is used so updates to the variable result in a change of position of the line."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "c",
		canvas.Background("black"),
		canvas.Width(screenunit.Pt(225).Pixels()),
		canvas.Height(screenunit.Pt(150).Pixels()),
	)
	pack.Pack(c, pack.PadX(screenunit.Pt(7.5)), pack.PadY(screenunit.Pt(7.5)), pack.Expand(true))

	// Build initial wave coordinates matching Tcl:
	// x from -10 to 300 step 5, each y=100, then spike at end (305,0) (310,200).
	var waveCoords []float64
	for x := -10; x <= 300; x += 5 {
		waveCoords = append(waveCoords, float64(x), 100)
	}
	waveCoords = append(waveCoords, 305, 0, 310, 200)

	waveID := c.CreateLine(scaledCoords(waveCoords),
		canvas.OutlineColor("green"), canvas.OutlineWidth(screenunit.Pt(0.75).Pixels()), canvas.Smooth(true),
		canvas.Tags("wave"))

	direction := "left"

	// basicMotion shifts y-values one position left or right through the array.
	basicMotion := func() {
		n := len(waveCoords)
		old := make([]float64, n)
		copy(old, waveCoords)
		for i := 1; i < n; i += 2 {
			if direction == "left" {
				if i+2 >= n {
					waveCoords[i] = old[1]
				} else {
					waveCoords[i] = old[i+2]
				}
			} else {
				if i-2 < 0 {
					waveCoords[i] = old[n-1]
				} else {
					waveCoords[i] = old[i-2]
				}
			}
		}
	}

	// reverser detects when peak moves off-screen and reverses direction.
	reverser := func() {
		if waveCoords[1] < 10 {
			direction = "right"
		} else if waveCoords[len(waveCoords)-1] < 10 {
			direction = "left"
		}
	}

	// Animation loop: shift coordinates and update canvas.
	var move func()
	move = func() {
		basicMotion()
		reverser()
		c.SetItemCoords(waveID, scaledCoords(waveCoords))
		app.After(10*time.Millisecond, move)
	}
	move() // aniwave.tcl starts with one step, then reschedules

	app.Run()
}
