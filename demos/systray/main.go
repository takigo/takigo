// Demo: System tray icon.
// Ported from Tk's systray.tcl demo.
package main

import (
	"fmt"
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/systray"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("System Tray Demo", 400, 200,
		"Click the button to add a system tray icon.\nThe icon will be removed after 10 seconds.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

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

	trayBtn := button.New(root, "traybtn", app,
		button.Text("Add Tray Icon"),
		button.Command(func() {
			tray, err := systray.New(app, app.Display(),
				systray.TrayTooltip("Takigo Demo"),
				systray.TrayClickHandler(func() {
					setStatus("Tray icon clicked!")
				}),
			)
			if err != nil {
				setStatus(fmt.Sprintf("Tray error: %v", err))
			} else {
				setStatus("Tray icon added (removes in 10s)")
				app.After(10*time.Second, func() {
					tray.Destroy()
					setStatus("Tray icon removed")
				})
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(trayBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(30), pack.PadY(20))

	_ = statusLabel
	_ = trayBtn
	d.Run()
}
