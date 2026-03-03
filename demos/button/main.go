// Demo: Buttons that change the window background color.
// Ported from Tk's button.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	app := demohelper.Setup("Button Demonstration", 400, 350,
		"If you click on any of the four buttons below, the background "+
			"of the button area will change to the color indicated in the "+
			"button. You can press Tab to move among the buttons, then "+
			"press Space to invoke the current button.")
	root := app.Window()

	// Color-changing function.
	changeColor := func(colorName string) {
		c, err := app.ColorCache().Get(colorName)
		if err != nil {
			return
		}
		root.BackgroundPixel = c.Pixel
		di := root.Display.Server
		gc := root.GC
		di.SetForeground(gc, c.Pixel)
		di.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		di.Flush()
		// Redraw all children.
		pack.ArrangeContainer(root)
	}

	// Color buttons (match Tk's button.tcl).
	colors := []struct {
		text  string
		color string
	}{
		{"Peach Puff", "#ffdab9"},
		{"Light Blue", "#add8e6"},
		{"Sea Green", "#2e8b57"},
		{"Yellow", "#ffff00"},
	}

	for _, c := range colors {
		colorVal := c.color
		btn := button.New(app, "btn_"+c.text,
			button.Text(c.text),
			button.Command(func() { changeColor(colorVal) }),
			button.PadX(10),
			button.PadY(6),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.Expand(true), pack.PadX(20), pack.PadY(5))
	}

	app.Run()
}
