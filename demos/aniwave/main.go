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
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Animated Waveform"), takigo.Size(550, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("An animated sine wave on a canvas."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

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

	// Root events.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})
	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	app.MainLoop()
}
