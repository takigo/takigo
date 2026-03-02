// Demo: Entry widgets with password mode and various constraints.
// Ported from Tk's entry3.tcl demo (simplified — no validation callbacks).
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
	app, err := takigo.NewApp(takigo.Title("Validated Entries"), takigo.Size(500, 300))
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
		label.Text("Four different entries are shown, each with\ndifferent constraints. Use Tab to move between them."),
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

	// Row 0: Integer entry.
	l1 := label.New(formFrame.Window(), "l1", app,
		label.Text("Integer:"),
		label.Anchor(option.AnchorE),
	)
	e1 := entry.New(formFrame.Window(), "e1", app,
		entry.Text("12345"),
		entry.Width(20),
	)
	grid.Grid(l1.Window(), grid.Row(0), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e1.Window(), grid.Row(0), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 1: Short entry (max ~10 chars concept).
	l2 := label.New(formFrame.Window(), "l2", app,
		label.Text("Short text:"),
		label.Anchor(option.AnchorE),
	)
	e2 := entry.New(formFrame.Window(), "e2", app,
		entry.Width(10),
		entry.Placeholder("Max 10 chars"),
	)
	grid.Grid(l2.Window(), grid.Row(1), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e2.Window(), grid.Row(1), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 2: Phone number.
	l3 := label.New(formFrame.Window(), "l3", app,
		label.Text("Phone:"),
		label.Anchor(option.AnchorE),
	)
	e3 := entry.New(formFrame.Window(), "e3", app,
		entry.Text("1-(555)-123-4567"),
		entry.Width(20),
	)
	grid.Grid(l3.Window(), grid.Row(2), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e3.Window(), grid.Row(2), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

	// Row 3: Password entry.
	l4 := label.New(formFrame.Window(), "l4", app,
		label.Text("Password:"),
		label.Anchor(option.AnchorE),
	)
	e4 := entry.New(formFrame.Window(), "e4", app,
		entry.Show('*'),
		entry.Width(20),
		entry.Placeholder("Enter password"),
	)
	grid.Grid(l4.Window(), grid.Row(3), grid.Column(0), grid.Sticky(grid.StickE), grid.PadX(5), grid.PadY(5))
	grid.Grid(e4.Window(), grid.Row(3), grid.Column(1), grid.Sticky(grid.EW), grid.PadX(5), grid.PadY(5))

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
	_ = l1
	_ = l2
	_ = l3
	_ = l4
	_ = e1
	_ = e2
	_ = e3
	_ = e4
	app.MainLoop()
}
