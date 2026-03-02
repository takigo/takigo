// Demo: Animated canvas waveform.
// Ported from Tk's aniwave.tcl demo.
package main

import (
	"fmt"
	"math"
	"time"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
)

func main() {
	d := demohelper.Setup("Animated Waveform", 550, 350, "An animated sine wave on a canvas.")
	root, app := d.Root, d.App

	c := canvas.New(root, "wave", app,
		canvas.Background("black"),
		canvas.Width(500),
		canvas.Height(250),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Initial wave.
	phase := 0.0
	wavePoints := 100

	makeWaveCoords := func(p float64) []float64 {
		coords := make([]float64, wavePoints*2)
		for i := range wavePoints {
			x := float64(i) * 500 / float64(wavePoints-1)
			y := 125 + 80*math.Sin(2*math.Pi*float64(i)/float64(wavePoints)+p)
			coords[i*2] = x
			coords[i*2+1] = y
		}
		return coords
	}

	waveID := c.CreateLine(makeWaveCoords(0),
		canvas.OutlineColor("#00ff00"), canvas.OutlineWidth(2), canvas.Smooth(true))

	// Second wave (different color/phase).
	wave2ID := c.CreateLine(makeWaveCoords(math.Pi/3),
		canvas.OutlineColor("#ff6600"), canvas.OutlineWidth(2), canvas.Smooth(true))

	// Animation loop.
	var animate func()
	animate = func() {
		phase += 0.1
		c.SetItemCoords(fmt.Sprintf("%d", waveID), makeWaveCoords(phase))
		c.SetItemCoords(fmt.Sprintf("%d", wave2ID), makeWaveCoords(phase+math.Pi/3))
		app.After(33*time.Millisecond, animate)
	}
	app.After(33*time.Millisecond, animate)

	d.Run()
}
