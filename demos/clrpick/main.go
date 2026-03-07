// Demo: Color picker dialog.
// Ported from Tk's clrpick.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Color Picker"),
		takigo.Geometry("+300+300"),
		takigo.IconName("colors"),
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
		label.Text("Press the buttons below to choose the foreground and background colors for the widgets in this window."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Current colors tracked for initial values in subsequent dialogs.
	bgHex := "#d9d9d9"
	fgHex := "#000000"

	// "Set background color ..." button.
	backBtn := button.New(f, "back",
		button.Text("Set background color ..."),
	)

	// "Set foreground color ..." button.
	foreBtn := button.New(f, "fore",
		button.Text("Set foreground color ..."),
	)

	backBtn.Command = func() {
		color, ok := dialog.ChooseColor(app,
			dialog.ColorTitle("Choose a background color"),
			dialog.ColorInitial(bgHex),
		)
		if ok {
			bgHex = color
			// Apply background recursively to all widgets in the window.
			window.ApplyBackgroundRecursive(app.Root(), color)
		}
	}

	foreBtn.Command = func() {
		color, ok := dialog.ChooseColor(app,
			dialog.ColorTitle("Choose a foreground color"),
			dialog.ColorInitial(fgHex),
		)
		if ok {
			fgHex = color
			c, err := app.ColorCache().Get(color)
			if err == nil {
				// Apply foreground to both buttons.
				backBtn.Foreground = c
				backBtn.Display()
				foreBtn.Foreground = c
				foreBtn.Display()
			}
		}
	}

	pack.Pack(backBtn, pack.SideOpt(pack.Top), pack.Anchor(0), pack.PadY("2m"))
	pack.Pack(foreBtn, pack.SideOpt(pack.Top), pack.Anchor(0), pack.PadY("2m"))

	app.Run()
}
