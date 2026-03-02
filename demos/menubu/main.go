// Demo: Menu buttons with different directions.
// Ported from Tk's menubu.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/widget/menubutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Menubutton Directions"), takigo.Size(500, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Four menu buttons showing menus in different directions.\nClick each to see the menu position."),
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

	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Create menu buttons with different directions.
	mbFrame := frame.New(root, "mbframe", app)
	pack.Pack(mbFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(20))

	directions := []struct {
		name string
		dir  menubutton.Direction
	}{
		{"Below", menubutton.Below},
		{"Above", menubutton.Above},
		{"Left", menubutton.Left},
		{"Right", menubutton.Right},
	}

	for _, d := range directions {
		m := menu.New(root, "menu_"+d.name, app)
		dirName := d.name
		m.AddCommand("Item 1", func() { setStatus(dirName + " > Item 1") })
		m.AddCommand("Item 2", func() { setStatus(dirName + " > Item 2") })
		m.AddCommand("Item 3", func() { setStatus(dirName + " > Item 3") })

		mb := menubutton.New(mbFrame.Window(), "mb_"+d.name, app,
			menubutton.Text(d.name),
			menubutton.MenuOpt(m),
			menubutton.DirectionOpt(d.dir),
		)
		pack.Pack(mb.Window(), pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(10))
		_ = mb
	}

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
	_ = statusLabel
	app.MainLoop()
}
