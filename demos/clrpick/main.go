// Demo: Color picker dialog.
// Ported from Tk's clrpick.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	app := demohelper.Setup("Color Picker", 400, 250,
		"Press the buttons below to choose the foreground and background colors for the widgets in this window.")

	// Current colors tracked for initial values in subsequent dialogs.
	bgHex := "#d9d9d9"
	fgHex := "#000000"

	// "Set background color ..." button.
	backBtn := button.New(app, "back",
		button.Text("Set background color ..."),
		button.PadX(15), button.PadY(8),
	)

	// "Set foreground color ..." button.
	foreBtn := button.New(app, "fore",
		button.Text("Set foreground color ..."),
		button.PadX(15), button.PadY(8),
	)

	backBtn.Command = func() {
		color, ok := dialog.ChooseColor(app,
			dialog.ColorTitle("Choose a background color"),
			dialog.ColorInitial(bgHex),
		)
		if ok {
			bgHex = color
			c, err := app.ColorCache().Get(color)
			if err == nil {
				// Apply background to both buttons.
				backBtn.Background = c
				backBtn.UpdateBorder()
				backBtn.Display()
				foreBtn.Background = c
				foreBtn.UpdateBorder()
				foreBtn.Display()

				// Apply to root window background.
				root := app.Root()
				root.BackgroundPixel = c.Pixel
			}
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

	pack.Pack(backBtn, pack.SideOpt(pack.Top), pack.PadY(5))
	pack.Pack(foreBtn, pack.SideOpt(pack.Top), pack.PadY(5))

	app.Run()
}
