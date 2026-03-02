// Demo: TTK Menubuttons with different menu configurations.
// Ported from Tk's ttkmenu.tcl demo (simplified — no tearoff).
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("TTK Menubutton Demonstration"), takigo.Size(400, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	ttk.SetCurrentTheme("clam")

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("Below are themed menubuttons. Click each to\nopen a dropdown menu."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Selected: (none)"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	setStatus := func(s string) {
		statusLabel.Text = "Selected: " + s
		statusLabel.Display()
	}

	// File menubutton.
	fileMenu := menu.New(root, "filemenu", app)
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Open", func() { setStatus("File > Open") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Close", func() { setStatus("File > Close") })

	fileMB := ttk.NewMenubutton(root, "filemb", app,
		ttk.MenubuttonText("File"),
		ttk.MenubuttonMenu(fileMenu),
	)
	pack.Pack(fileMB.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	// Edit menubutton.
	editMenu := menu.New(root, "editmenu", app)
	editMenu.AddCommand("Cut", func() { setStatus("Edit > Cut") })
	editMenu.AddCommand("Copy", func() { setStatus("Edit > Copy") })
	editMenu.AddCommand("Paste", func() { setStatus("Edit > Paste") })

	editMB := ttk.NewMenubutton(root, "editmb", app,
		ttk.MenubuttonText("Edit"),
		ttk.MenubuttonMenu(editMenu),
	)
	pack.Pack(editMB.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

	// Help menubutton.
	helpMenu := menu.New(root, "helpmenu", app)
	helpMenu.AddCommand("About", func() { setStatus("Help > About") })
	helpMenu.AddCommand("Documentation", func() { setStatus("Help > Documentation") })

	helpMB := ttk.NewMenubutton(root, "helpmb", app,
		ttk.MenubuttonText("Help"),
		ttk.MenubuttonMenu(helpMenu),
	)
	pack.Pack(helpMB.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

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
	_ = fileMB
	_ = editMB
	_ = helpMB
	app.MainLoop()
}
