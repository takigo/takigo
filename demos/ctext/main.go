// Demo: Canvas text items at various positions and anchors.
// Ported from Tk's ctext.tcl demo (simplified — no inline editing).
package main

import (
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
)

func main() {
	d := demohelper.Setup("Canvas Text Demo", 550, 400,
		"Canvas text items rendered at various positions\nwith different anchors, fonts, and colors.")
	root, app := d.Root, d.App

	// Canvas.
	c := canvas.New(root, "canvas", app,
		canvas.Background("white"),
		canvas.Width(500),
		canvas.Height(300),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Title text.
	c.CreateText(250, 30,
		canvas.TextOpt("Canvas Text Demo"),
		canvas.FontOpt("Sans Bold 18"),
		canvas.TextColor("navy"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

	// Multi-line centered text.
	c.CreateText(250, 80,
		canvas.TextOpt("This text is rendered directly on the canvas.\nMultiple lines are supported via newlines."),
		canvas.FontOpt("Sans 12"),
		canvas.TextColor("black"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

	// Left-anchored text.
	c.CreateText(50, 150,
		canvas.TextOpt("Left-anchored text\nat position (50, 150)"),
		canvas.FontOpt("Sans Italic 11"),
		canvas.TextColor("darkred"),
		canvas.AnchorOpt(option.AnchorNW),
	)

	// Right-anchored text.
	c.CreateText(450, 150,
		canvas.TextOpt("Right-anchored text\nat position (450, 150)"),
		canvas.FontOpt("Sans 11"),
		canvas.TextColor("darkblue"),
		canvas.AnchorOpt(option.AnchorNE),
	)

	// Center paragraph.
	c.CreateText(250, 230,
		canvas.TextOpt("Center-anchored text with a longer paragraph.\nThis demonstrates how canvas text wraps and displays\nmultiple lines of content."),
		canvas.FontOpt("Serif 11"),
		canvas.TextColor("darkgreen"),
		canvas.AnchorOpt(option.AnchorCenter),
	)

	d.Run()
}
