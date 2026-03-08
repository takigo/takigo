// Demo: Animated canvas waveform.
// Ported from Tk's aniwave.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Animated Wave"),
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
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This demonstration contains a canvas widget with a line item inside it. The animation routines work by adjusting the coordinates list of the line."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "wave",
		canvas.Background("black"),
		canvas.Width(300),
		canvas.Height(200),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("7.5p"), pack.PadY("7.5p"))

	// Initial wave.
	phase := 0.0
	wavePoints := 100

	makeWaveCoords := func(p float64) []float64 {
		coords := make([]float64, wavePoints*2)
		for i := range wavePoints {
			x := float64(i) * 300 / float64(wavePoints-1)
			y := 100 + 60*math.Sin(2*math.Pi*float64(i)/float64(wavePoints)+p)
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

	// Animation state.
	running := true

	// Animation loop.
	var animate func()
	animate = func() {
		if !running {
			return
		}
		phase += 0.1
		c.SetItemCoords(fmt.Sprintf("%d", waveID), makeWaveCoords(phase))
		c.SetItemCoords(fmt.Sprintf("%d", wave2ID), makeWaveCoords(phase+math.Pi/3))
		app.After(33*time.Millisecond, animate)
	}
	app.After(33*time.Millisecond, animate)

	// Pause/Resume toggle button.
	pauseBtn := button.New(f, "pause",
		button.Text("Pause"),
	)
	pauseBtn.Command = func() {
		if running {
			running = false
			pauseBtn.Text = "Resume"
			pauseBtn.Display()
		} else {
			running = true
			pauseBtn.Text = "Pause"
			pauseBtn.Display()
			animate()
		}
	}
	pack.Pack(pauseBtn, pack.SideOpt(pack.Top), pack.PadY("3p"))

	app.Run()
}
