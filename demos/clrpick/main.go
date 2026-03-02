// Demo: Color picker dialog.
// Ported from Tk's clrpick.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Color Picker"), takigo.Size(400, 250))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Click the button to open the color chooser.\nThe chosen color is displayed below."),
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

	// Color display label.
	colorLabel := label.New(root, "colorlabel", app,
		label.Text("Selected: #3399ff"),
		label.Background("#3399ff"),
		label.Foreground("white"),
		label.PadX(20), label.PadY(20),
	)
	pack.Pack(colorLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	chooseBtn := button.New(root, "choose", app,
		button.Text("Choose Color..."),
		button.Command(func() {
			color, ok := dialog.ChooseColor(app,
				dialog.ColorParent(root),
				dialog.ColorInitial("#3399ff"),
			)
			if ok {
				colorLabel.Text = fmt.Sprintf("Selected: %s", color)
				c, err := app.ColorCache().Get(color)
				if err == nil {
					colorLabel.Background = c
					colorLabel.UpdateBorder()
				}
				colorLabel.Display()
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(chooseBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(30), pack.PadY(20))

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
	_ = colorLabel
	_ = chooseBtn
	app.MainLoop()
}
