// Demo: Menu bar with cascading submenus.
// Ported from Tk's menu.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("Menu Demo"), takigo.Size(500, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Click the menu buttons below to open menus.\nMenus support commands, separators, and cascaded submenus."),
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

	// Menu bar frame.
	menuBar := frame.New(root, "menubar", app,
		frame.Relief(option.ReliefRaised),
		frame.BorderWidth(1),
	)
	pack.Pack(menuBar.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// File menu.
	fileMenu := menu.New(root, "filemenu", app)
	fileMenu.AddCommand("New", func() { setStatus("File > New") })
	fileMenu.AddCommand("Open...", func() { setStatus("File > Open") })
	fileMenu.AddCommand("Save", func() { setStatus("File > Save") })
	fileMenu.AddCommand("Save As...", func() { setStatus("File > Save As") })
	fileMenu.AddSeparator()
	fileMenu.AddCommand("Quit", func() { app.Quit() })

	fileMb := menubutton.New(menuBar.Window(), "filemb", app,
		menubutton.Text("File"),
		menubutton.MenuOpt(fileMenu),
	)
	pack.Pack(fileMb.Window(), pack.SideOpt(pack.Left))

	// Edit menu with cascade.
	editMenu := menu.New(root, "editmenu", app)
	editMenu.AddCommand("Undo", func() { setStatus("Edit > Undo") })
	editMenu.AddCommand("Redo", func() { setStatus("Edit > Redo") })
	editMenu.AddSeparator()
	editMenu.AddCommand("Cut", func() { setStatus("Edit > Cut") })
	editMenu.AddCommand("Copy", func() { setStatus("Edit > Copy") })
	editMenu.AddCommand("Paste", func() { setStatus("Edit > Paste") })

	editMb := menubutton.New(menuBar.Window(), "editmb", app,
		menubutton.Text("Edit"),
		menubutton.MenuOpt(editMenu),
	)
	pack.Pack(editMb.Window(), pack.SideOpt(pack.Left))

	// Help menu.
	helpMenu := menu.New(root, "helpmenu", app)
	helpMenu.AddCommand("About", func() { setStatus("Help > About: Takigo Menu Demo") })
	helpMenu.AddCommand("Documentation", func() { setStatus("Help > Documentation") })

	helpMb := menubutton.New(menuBar.Window(), "helpmb", app,
		menubutton.Text("Help"),
		menubutton.MenuOpt(helpMenu),
	)
	pack.Pack(helpMb.Window(), pack.SideOpt(pack.Left))

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
	_ = fileMb
	_ = editMb
	_ = helpMb
	app.MainLoop()
}
