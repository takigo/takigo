// Demo: System tray icon.
// Ported from Tk's systray.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/systray"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

func main() {
	app := demohelper.Setup("System Tray Demonstration", 400, 200,
		"This demonstration showcases the system tray commands. Running this demo creates the systray icon. Clicking the buttons below modifies and destroys the icon and displays the notification.")

	statusLabel := label.New(app, "status",
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	var tray *systray.TrayIcon
	modified := false

	// Labelframe with Create / Modify / Destroy buttons.
	lf := labelframe.New(app, "trayframe",
		labelframe.Text("Tray Icon"),
	)

	createBtn := button.New(lf, "create",
		button.Text("Create"),
		button.Command(func() {
			if tray != nil {
				setStatus("Systray icon already exists")
				return
			}
			var err error
			tray, err = systray.New(app, app.Display(),
				systray.TrayTooltip("Takigo Demo"),
				systray.TrayClickHandler(func() {
					setStatus("Tray icon clicked!")
				}),
			)
			if err != nil {
				setStatus(fmt.Sprintf("Tray error: %v", err))
				tray = nil
			} else {
				modified = false
				setStatus("Tray icon created")
			}
		}),
		button.PadX(8), button.PadY(4),
	)

	modifyBtn := button.New(lf, "modify",
		button.Text("Modify"),
		button.Command(func() {
			if tray == nil {
				setStatus("Please create systray icon first")
				return
			}
			if !modified {
				tray.SetTooltip("Modified tooltip")
				setStatus("Tray icon modified (tooltip changed)")
				modified = true
			} else {
				tray.SetTooltip("Takigo Demo")
				setStatus("Tray icon restored (tooltip reset)")
				modified = false
			}
		}),
		button.PadX(8), button.PadY(4),
	)

	destroyBtn := button.New(lf, "destroy",
		button.Text("Destroy"),
		button.Command(func() {
			if tray == nil {
				setStatus("Systray icon was already destroyed")
				return
			}
			tray.Destroy()
			tray = nil
			modified = false
			setStatus("Tray icon destroyed")
		}),
		button.PadX(8), button.PadY(4),
	)

	pack.Pack(createBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX(4), pack.PadY(4))
	pack.Pack(modifyBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX(4), pack.PadY(4))
	pack.Pack(destroyBtn, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillX), pack.PadX(4), pack.PadY(4))

	pack.Pack(lf, pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(10))

	_ = statusLabel
	app.Run()
}
