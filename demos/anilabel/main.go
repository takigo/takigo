// Demo: Animated scrolling label.
// Ported from Tk's anilabel.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Animated Label", 400, 200,
		"A label with scrolling text animation.")
	root, app := d.Root, d.App

	// Animated label.
	scrollText := "    Welcome to Takigo — a pure Go port of the Tk toolkit!    "
	runes := []rune(scrollText)
	offset := 0

	aniLabel := label.New(root, "anilabel", app,
		label.Text(scrollText),
		label.Anchor(option.AnchorW),
		label.PadX(20), label.PadY(20),
	)
	pack.Pack(aniLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(20))

	// Animation loop.
	var animate func()
	animate = func() {
		offset = (offset + 1) % len(runes)
		visible := make([]rune, len(runes))
		for i := range runes {
			visible[i] = runes[(offset+i)%len(runes)]
		}
		aniLabel.Text = string(visible)
		aniLabel.Display()
		app.After(100*time.Millisecond, animate)
	}
	app.After(100*time.Millisecond, animate)

	_ = aniLabel
	d.Run()
}
