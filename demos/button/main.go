// Demo: Buttons that change the window background color.
// Ported from Tk's button.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	d := demohelper.Setup("Button Demonstration", 400, 350,
		"Click any button to change the background color.\nThe color resets after 1.5 seconds.")
	root, app := d.Root, d.App

	defaultBg, _ := app.ColorCache().Get("#d9d9d9")

	// Color-changing function.
	changeColor := func(colorName string) {
		c, err := app.ColorCache().Get(colorName)
		if err != nil {
			return
		}
		root.BackgroundPixel = c.Pixel
		di := root.Display.XDisplay
		gc := root.GC
		di.SetForeground(gc, c.Pixel)
		di.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		di.Flush()

		// Reset after 1.5 seconds.
		app.After(1500*time.Millisecond, func() {
			root.BackgroundPixel = defaultBg.Pixel
			di.SetForeground(gc, defaultBg.Pixel)
			di.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
			di.Flush()
			// Redraw all children.
			pack.ArrangeContainer(root)
		})
	}

	// Color buttons.
	colors := []struct {
		text  string
		color string
	}{
		{"Peach Puff", "#ffdab9"},
		{"Light Blue", "#add8e6"},
		{"Bisque", "#ffe4c4"},
		{"Light Green", "#90ee90"},
	}

	for _, c := range colors {
		colorVal := c.color
		btn := button.New(root, "btn_"+c.text, app,
			button.Text(c.text),
			button.Command(func() { changeColor(colorVal) }),
			button.PadX(10),
			button.PadY(6),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.Expand(true), pack.PadX(20), pack.PadY(5))
	}

	d.Run()
}
