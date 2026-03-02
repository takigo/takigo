// Demo: Canvas and text display layout.
// Ported from Tk's print.tcl demo (printing not available).
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
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Canvas & Text Print Demo"), takigo.Size(700, 500))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("This demo shows a canvas with shapes and a text widget\nwith sample content, as used in the Tk print demo."),
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

	infoLabel := label.New(root, "info", app,
		label.Text("Note: Printing is not available on this platform."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.Foreground("#666666"),
	)
	pack.Pack(infoLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(3))

	// Content area: canvas left, text right.
	contentFrame := frame.New(root, "content", app)
	pack.Pack(contentFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	// Canvas with shapes.
	c := canvas.New(contentFrame.Window(), "canv", app,
		canvas.Background("white"),
		canvas.Width(300), canvas.Height(350),
		canvas.BorderWidthOpt(2), canvas.ReliefOpt(option.ReliefSunken),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5))

	// Draw shapes on canvas.
	c.CreateRectangle(20, 20, 140, 80, canvas.FillColor("#4488cc"), canvas.OutlineColor("black"))
	c.CreateOval(160, 20, 280, 100, canvas.FillColor("#cc4444"), canvas.OutlineColor("black"))
	c.CreateRectangle(20, 120, 280, 200, canvas.FillColor("#44aa44"), canvas.OutlineColor("black"))
	c.CreateOval(60, 220, 240, 340, canvas.FillColor("#cc8844"), canvas.OutlineColor("black"))
	c.CreateText(150, 105, canvas.TextOpt("Shapes Demo"), canvas.TextColor("black"))

	// Text widget with sample content.
	tw := text.New(contentFrame.Window(), "txt", app,
		text.Width(40), text.Height(20),
		text.WrapModeOpt(text.WrapWord),
		text.BorderWidthOpt(2),
	)
	pack.Pack(tw.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5))

	tw.Insert("1.0", `This is a sample text widget that would normally be printed alongside the canvas in Tk's print demo.

The original Tk demo demonstrates the "tk print" command which sends canvas and text widget contents to the system printer.

Since printing is a platform-specific feature that relies on the native print dialog, it is not implemented in the Go port.

However, this demo still shows the canvas with geometric shapes on the left and this text widget on the right, demonstrating the layout used in the original demo.

Features shown:
- Canvas rectangles and ovals
- Canvas text items
- Fill and outline colors
- Text widget with word wrapping
- Side-by-side layout with pack geometry`)

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
	_ = infoLabel
	app.MainLoop()
}
