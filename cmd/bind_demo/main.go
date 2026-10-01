// Phase 12 demo: binding engine with per-widget, class, all bindings,
// virtual events, and double-click detection.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(
		takigo.Title("Bind Engine Demo"),
		takigo.Size(500, 400),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	eng := app.Bind()

	// Status label to show binding events.
	statusFrame := frame.New(app, "statusframe")
	pack.Pack(statusFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	statusLabel := label.New(statusFrame, "status",
		label.Text("Status: Ready"))
	pack.Pack(statusLabel, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX), pack.Expand(true))

	setStatus := func(msg string) {
		statusLabel.Configure(label.Text(msg))
	}

	// Two buttons to demonstrate per-widget bindings.
	btnFrame := frame.New(app, "btnframe")
	pack.Pack(btnFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(10))

	btn1 := button.New(btnFrame, "btn1",
		button.Text("Button 1 (click me)"),
		button.Command(func() { setStatus("Button 1 command") }))
	pack.Pack(btn1, pack.SideOpt(pack.Left), pack.PadX(10))

	btn2 := button.New(btnFrame, "btn2",
		button.Text("Button 2 (click me)"),
		button.Command(func() { setStatus("Button 2 command") }))
	pack.Pack(btn2, pack.SideOpt(pack.Left), pack.PadX(10))

	// Register windows with the bind engine.
	eng.RegisterWindow(root, "Toplevel")
	eng.RegisterWindow(statusFrame.Window(), "Frame")
	eng.RegisterWindow(statusLabel.Window(), "Label")
	eng.RegisterWindow(btnFrame.Window(), "Frame")
	eng.RegisterWindow(btn1.Window(), "Button")
	eng.RegisterWindow(btn2.Window(), "Button")

	// --- Per-widget binding ---
	eng.BindWindow(btn1, "<Enter>", func(*bind.EventData) bool {
		setStatus("Mouse entered Button 1")
		return false
	})
	eng.BindWindow(btn1, "<Leave>", func(*bind.EventData) bool {
		setStatus("Mouse left Button 1")
		return false
	})

	// --- Class binding ---
	eng.Bind("Button", "<Button-3>", func(*bind.EventData) bool {
		setStatus("Right-click on any Button (class binding)")
		return false
	})

	// --- "all" binding ---
	eng.Bind("all", "<Motion>", func(ed *bind.EventData) bool {
		setStatus(fmt.Sprintf("Motion at (%d, %d)", ed.Event.X, ed.Event.Y))
		return false
	})

	// --- Double-click binding ---
	eng.Bind("Button", "<Double-Button-1>", func(*bind.EventData) bool {
		setStatus("Double-click on a Button!")
		return true // break: don't also fire single-click class binding
	})

	// --- Virtual event binding ---
	eng.Bind("all", "<<Copy>>", func(*bind.EventData) bool {
		setStatus("Virtual event: <<Copy>> (Ctrl+C)")
		return false
	})
	eng.Bind("all", "<<Paste>>", func(*bind.EventData) bool {
		setStatus("Virtual event: <<Paste>> (Ctrl+V)")
		return false
	})
	eng.Bind("all", "<<SelectAll>>", func(*bind.EventData) bool {
		setStatus("Virtual event: <<SelectAll>> (Ctrl+A)")
		return false
	})

	// --- Key binding with break ---
	eng.Bind(root.PathName, "<Key-q>", func(*bind.EventData) bool {
		setStatus("'q' pressed on root — quitting...")
		app.Quit()
		return true
	})
	eng.Bind("all", "<Key-Escape>", func(*bind.EventData) bool {
		setStatus("Escape pressed — quitting...")
		app.Quit()
		return true
	})

	// Info label.
	infoLabel := label.New(app, "info",
		label.Text("Try: hover buttons, right-click, double-click, Ctrl+C/V/A, press 'q' or Esc"))
	pack.Pack(infoLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(10))
	eng.RegisterWindow(infoLabel.Window(), "Label")

	fmt.Println("Bind demo running. Press 'q' or Escape to quit.")
	app.Run()
}
