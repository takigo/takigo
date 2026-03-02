// Demo: Color picker dialog.
// Ported from Tk's clrpick.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Color Picker", 400, 250,
		"Click the button to open the color chooser.\nThe chosen color is displayed below.")
	root, app := d.Root, d.App

	// Color display label.
	colorLabel := label.New(root, "colorlabel", app,
		label.Text("Selected: #3399ff"),
		label.Background("#3399ff"),
		label.Foreground("white"),
		label.PadX(20), label.PadY(20),
	)
	pack.Pack(colorLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	chooseBtn := button.New(root, "choose", app,
		button.Text("Choose Color..."),
		button.Command(func() {
			color, ok := dialog.ChooseColor(app,
				dialog.ColorParent(root),
				dialog.ColorInitial("#3399ff"),
			)
			if ok {
				colorLabel.Text = fmt.Sprintf("Selected: %s", color)
				c, err := app.ColorCache().Get(color)
				if err == nil {
					colorLabel.Background = c
					colorLabel.UpdateBorder()
				}
				colorLabel.Display()
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(chooseBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(30), pack.PadY(20))

	_ = colorLabel
	_ = chooseBtn
	d.Run()
}
