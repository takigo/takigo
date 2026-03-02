// Demo: Animated scrolling label.
// Ported from Tk's anilabel.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Animated Label"), takigo.Size(400, 200))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("A label with scrolling text animation."),
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

	// Animated label.
	scrollText := "    Welcome to Takigo — a pure Go port of the Tk toolkit!    "
	runes := []rune(scrollText)
	offset := 0

	aniLabel := label.New(root, "anilabel", app,
		label.Text(scrollText),
		label.Anchor(option.AnchorW),
		label.PadX(20), label.PadY(20),
	)
	pack.Pack(aniLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(20))

	// Animation loop.
	var animate func()
	animate = func() {
		offset = (offset + 1) % len(runes)
		visible := make([]rune, len(runes))
		for i := range runes {
			visible[i] = runes[(offset+i)%len(runes)]
		}
		aniLabel.Text = string(visible)
		aniLabel.Display()
		app.After(100*time.Millisecond, animate)
	}
	app.After(100*time.Millisecond, animate)

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
	_ = aniLabel
	app.MainLoop()
}
