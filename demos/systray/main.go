// Demo: System tray icon.
// Ported from Tk's systray.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/systray"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("System Tray Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("systray"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This demonstration showcases the system tray commands. Running this demo creates the systray icon. Clicking the buttons below modifies and destroys the icon and displays the notification."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Context menu for right-click on tray icon (matches Tcl's button3 handler).
	iconMenu := menu.New(app, "iconmenu")
	iconMenu.AddCommand("Status", func() {
		dialog.ShowMessage(app, dialog.MsgTitle("Systray"), dialog.MsgMessage("Systray icon is active"))
	})
	iconMenu.AddCommand("Exit", func() {
		app.Quit()
	})

	var tray *systray.TrayIcon
	modified := false

	create := func() {
		if tray != nil {
			dialog.ShowMessage(app, dialog.MsgTitle("Systray"), dialog.MsgMessage("Systray icon already exists"))
			return
		}
		var err error
		tray, err = systray.New(app, app.Display(),
			systray.TrayTooltip("Takigo Demo"),
			systray.TrayClickHandler(func() {}),
			systray.TrayRightClickHandler(func(x, y int) {
				iconMenu.Post(x, y)
			}),
		)
		if err != nil {
			tray = nil
		}
		modified = false
	}

	// Labelframe with Create / Modify / Destroy buttons.
	lf := labelframe.New(f, "lf", labelframe.Text("Tray Icon"))

	createBtn := button.New(lf, "b0",
		button.Text("Create"),
		button.Command(func() { create() }),
	)

	modifyBtn := button.New(lf, "b1",
		button.Text("Modify"),
		button.Command(func() {
			if tray == nil {
				dialog.ShowMessage(app, dialog.MsgTitle("Systray"), dialog.MsgMessage("Please create systray icon first"))
				return
			}
			if !modified {
				tray.SetTooltip("Modified text")
				modified = true
			} else {
				tray.SetTooltip("Takigo Demo")
				modified = false
			}
		}),
	)

	destroyBtn := button.New(lf, "b2",
		button.Text("Destroy"),
		button.Command(func() {
			if tray == nil {
				dialog.ShowMessage(app, dialog.MsgTitle("Systray"), dialog.MsgMessage("Systray icon was already destroyed"))
				return
			}
			tray.Destroy()
			tray = nil
			modified = false
		}),
	)

	pack.Pack(createBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX("3p"), pack.PadY("3p"))
	pack.Pack(modifyBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX("3p"), pack.PadY("3p"))
	pack.Pack(destroyBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX("3p"), pack.PadY("3p"))

	notifyBtn := button.New(f, "b3",
		button.Text("Display Notification"),
		button.Command(func() {
			if tray == nil {
				dialog.ShowMessage(app, dialog.MsgTitle("Systray"), dialog.MsgMessage("Please create systray icon first"))
				return
			}
			dialog.ShowMessage(app, dialog.MsgTitle("Alert"), dialog.MsgMessage("This is an alert"))
		}),
	)

	pack.Pack(lf, pack.FillOpt(pack.FillX), pack.PadX("3p"), pack.PadY("3p"))
	pack.Pack(notifyBtn, pack.FillOpt(pack.FillX), pack.PadX("3p"), pack.PadY("3p"))

	// Auto-create systray icon at startup (matching Tcl's `create` call at end of script).
	app.After(0, func() { create() })

	app.Run()
}
