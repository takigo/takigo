// Demo: Simple form with labeled entries using grid layout.
// Ported from Tk's form.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Form Entry"), takigo.Size(450, 280))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("A simple form with five fields. Use Tab\nto move between fields."),
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

	// Form grid.
	formFrame := frame.New(root, "form", app)
	pack.Pack(formFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	fields := []string{"Name:", "Address:", "City:", "State:", "Phone:"}
	entries := make([]*entry.Entry, len(fields))

	for i, fieldName := range fields {
		l := label.New(formFrame.Window(), fmt.Sprintf("l%d", i), app,
			label.Text(fieldName),
			label.Anchor(option.AnchorE),
		)
		e := entry.New(formFrame.Window(), fmt.Sprintf("e%d", i), app,
			entry.Width(30),
		)
		grid.Grid(l.Window(), grid.Row(i), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(4))
		grid.Grid(e.Window(), grid.Row(i), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(4))
		entries[i] = e
		_ = l
	}

	grid.ColumnConfigure(formFrame.Window(), 1, grid.SlotConfig{Weight: 1})

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
	_ = focusMgr
	_ = entries
	app.MainLoop()
}
