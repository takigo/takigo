// Demo: Vertical paned window with text and listbox.
// Ported from Tk's paned2.tcl demo.
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
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Vertical Panes"), takigo.Size(500, 450))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("A vertical paned window with a listbox and text widget.\nDrag the sash between them to resize."),
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

	// Vertical paned window.
	pw := panedwindow.New(root, "vpanes", app,
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	pack.Pack(pw.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Top pane: listbox.
	topFrame := frame.New(pw.Window(), "top", app)
	lb := listbox.New(topFrame.Window(), "filelist", app,
		listbox.Items(
			"main.go", "canvas.go", "widget.go", "event.go", "window.go",
			"label.go", "button.go", "entry.go", "text.go", "scrollbar.go",
			"menu.go", "dialog.go", "scale.go", "listbox.go", "frame.go",
		),
		listbox.Height(6),
	)
	pack.Pack(lb.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(topFrame.Window(), 100)

	// Bottom pane: text widget.
	bottomFrame := frame.New(pw.Window(), "bottom", app)
	tw := text.New(bottomFrame.Window(), "preview", app,
		text.Width(50),
		text.Height(10),
		text.WrapModeOpt(text.WrapWord),
	)
	tw.Insert("1.0", "Select a file from the list above to preview its contents.\n\nThis is a vertical paned window demo showing a listbox\nin the top pane and a text widget in the bottom pane.\n\nDrag the horizontal sash to resize the panes.")
	pack.Pack(tw.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(bottomFrame.Window(), 100)

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
	_ = lb
	app.MainLoop()
}
