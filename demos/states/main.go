// Demo: US states listbox with vertical scrollbar.
// Ported from Tk's states.tcl demo.
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
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("US States"), takigo.Size(300, 400))
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
		label.Text("A listbox containing the 50 US states.\nSelect a state and press Dismiss to exit."),
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

	// Listbox frame.
	lbFrame := frame.New(root, "lbframe", app)
	pack.Pack(lbFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	states := []string{
		"Alabama", "Alaska", "Arizona", "Arkansas", "California",
		"Colorado", "Connecticut", "Delaware", "Florida", "Georgia",
		"Hawaii", "Idaho", "Illinois", "Indiana", "Iowa",
		"Kansas", "Kentucky", "Louisiana", "Maine", "Maryland",
		"Massachusetts", "Michigan", "Minnesota", "Mississippi", "Missouri",
		"Montana", "Nebraska", "Nevada", "New Hampshire", "New Jersey",
		"New Mexico", "New York", "North Carolina", "North Dakota", "Ohio",
		"Oklahoma", "Oregon", "Pennsylvania", "Rhode Island", "South Carolina",
		"South Dakota", "Tennessee", "Texas", "Utah", "Vermont",
		"Virginia", "Washington", "West Virginia", "Wisconsin", "Wyoming",
	}

	lb := listbox.New(lbFrame.Window(), "states", app,
		listbox.Items(states...),
		listbox.Height(12),
	)

	yscroll := scrollbar.New(lbFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

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
