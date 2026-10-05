// Demo: This demonstration script prompts the user to select a color.
// Ported from Tk's clrpick.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Color Selection Dialog"),
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
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Press the buttons below to choose the foreground and background colors for the widgets in this window."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
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
			// Apply foreground to both buttons.
			backBtn.Configure(button.Foreground(color))
			foreBtn.Configure(button.Foreground(color))
		}
	}

	pack.Pack(geometry.Group{backBtn, foreBtn}, pack.SideOpt(pack.Top), pack.Anchor(option.AnchorCenter), pack.PadY(screenunit.Mm(2)))

	app.Run()
}
