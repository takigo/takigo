// Demo: Ruler with tab stops on canvas.
// Ported from Tk's ruler.tcl demo (simplified).
package main

import (
	"fmt"
	"os"

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
	app, err := takigo.NewApp(takigo.Title("Ruler Demo"), takigo.Size(600, 250))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("A ruler with tick marks and tab stops.\nThe tab stops are represented by small triangles."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Canvas.
	c := canvas.New(root, "ruler", app,
		canvas.Background("#ffffee"),
		canvas.Width(560),
		canvas.Height(100),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	// Ruler background.
	c.CreateRectangle(20, 20, 540, 60,
		canvas.FillColor("#f0f0e0"), canvas.OutlineColor("black"), canvas.OutlineWidth(1))

	// Draw tick marks (1/8 inch marks at ~7.5px each for 72 DPI).
	rulerLeft := 20.0
	rulerTop := 20.0
	rulerBottom := 60.0
	ppi := 72.0 // pixels per inch

	for i := range int(7*8) + 1 { // 7 inches, 8 ticks per inch
		x := rulerLeft + float64(i)*ppi/8
		var tickLen float64
		switch {
		case i%8 == 0: // inch mark
			tickLen = rulerBottom - rulerTop
		case i%4 == 0: // half inch
			tickLen = (rulerBottom - rulerTop) * 0.6
		case i%2 == 0: // quarter inch
			tickLen = (rulerBottom - rulerTop) * 0.4
		default: // eighth inch
			tickLen = (rulerBottom - rulerTop) * 0.25
		}
		c.CreateLine([]float64{x, rulerBottom, x, rulerBottom - tickLen},
			canvas.OutlineColor("black"), canvas.OutlineWidth(1))

		// Inch labels.
		if i%8 == 0 && i > 0 {
			c.CreateText(x, rulerTop-5,
				canvas.TextOpt(fmt.Sprintf("%d", i/8)),
				canvas.FontOpt("Sans 8"), canvas.AnchorOpt(option.AnchorS))
		}
	}

	// Tab stop triangles.
	tabStops := []float64{1.0, 2.5, 4.0, 5.5} // inches
	for _, tab := range tabStops {
		x := rulerLeft + tab*ppi
		// Small downward-pointing triangle.
		c.CreatePolygon([]float64{x - 5, rulerBottom + 5, x + 5, rulerBottom + 5, x, rulerBottom + 15},
			canvas.FillColor("#cc0000"), canvas.OutlineColor("black"), canvas.OutlineWidth(1),
			canvas.Tags("tab"))
	}

	// Root event handlers.
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
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
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
