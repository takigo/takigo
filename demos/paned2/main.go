// Demo: Vertical paned window with text and listbox.
// Ported from Tk's paned2.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/panedwindow"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	d := demohelper.Setup("Vertical Panes", 500, 450,
		"A vertical paned window with a listbox and text widget.\nDrag the sash between them to resize.")
	root, app := d.Root, d.App

	// Vertical paned window.
	pw := panedwindow.New(root, "vpanes", app,
		panedwindow.OrientOpt(panedwindow.Vertical),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Top pane: listbox.
	topFrame := frame.New(pw.Window(), "top", app)
	lb := listbox.New(topFrame.Window(), "filelist", app,
		listbox.Items(
			"main.go", "canvas.go", "widget.go", "event.go", "window.go",
			"label.go", "button.go", "entry.go", "text.go", "scrollbar.go",
			"menu.go", "dialog.go", "scale.go", "listbox.go", "frame.go",
		),
		listbox.Height(6),
	)
	pack.Pack(lb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(topFrame.Window(), 100)

	// Bottom pane: text widget.
	bottomFrame := frame.New(pw.Window(), "bottom", app)
	tw := text.New(bottomFrame.Window(), "preview", app,
		text.Width(50),
		text.Height(10),
		text.WrapModeOpt(text.WrapWord),
	)
	tw.Insert("1.0", "Select a file from the list above to preview its contents.\n\nThis is a vertical paned window demo showing a listbox\nin the top pane and a text widget in the bottom pane.\n\nDrag the horizontal sash to resize the panes.")
	pack.Pack(tw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	pw.Add(bottomFrame.Window(), 100)

	_ = lb
	d.Run()
}
