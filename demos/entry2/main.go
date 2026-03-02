// Demo: Entry widgets with horizontal scrollbars.
// Ported from Tk's entry2.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Entry Demonstration (with scrollbars)"), takigo.Size(450, 320))
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
		label.Text("Three entry widgets with scrollbars are displayed below.\nYou can add characters by pointing, clicking and typing."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button at bottom.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Helper: create entry+scrollbar pair.
	makeEntryWithScroll := func(parent *frame.Frame, name, text string) *entry.Entry {
		ef := frame.New(parent.Window(), name+"_frame", app)
		pack.Pack(ef.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX(20), pack.PadY(5))

		e := entry.New(ef.Window(), name, app)
		sb := scrollbar.New(ef.Window(), name+"_sb", app,
			scrollbar.OrientOpt(scrollbar.Horizontal),
			scrollbar.CommandOpt(func(args ...any) {
				if len(args) < 1 {
					return
				}
				switch args[0] {
				case "moveto":
					if len(args) >= 2 {
						if f, ok := args[1].(float64); ok {
							e.XViewMoveTo(f)
						}
					}
				case "scroll":
					if len(args) >= 3 {
						n, _ := args[1].(int)
						unit, _ := args[2].(string)
						e.XViewScroll(n, unit == "pages")
					}
				}
			}),
		)

		e.SetText(text)
		e.ScrollCmd = func(first, last float64) {
			sb.Set(first, last)
		}

		pack.Pack(e.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
		pack.Pack(sb.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

		return e
	}

	// Container frame.
	container := frame.New(root, "container", app)
	pack.Pack(container.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Three entry+scrollbar pairs.
	e1 := makeEntryWithScroll(container, "e1", "Initial value")
	e2 := makeEntryWithScroll(container, "e2",
		"This entry contains a long value, much too long to fit in the window at one time, and thus you can use the scrollbar to see the rest.")
	e3 := makeEntryWithScroll(container, "e3", "")

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
	_ = e1
	_ = e2
	_ = e3
	app.MainLoop()
}
