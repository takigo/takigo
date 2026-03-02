// Demo: Font chooser dialog.
// Ported from Tk's fontchoose.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Font Chooser", 450, 250,
		"Click the button to open the font chooser.\nThe selected font description is shown below.")
	root, app := d.Root, d.App

	// Font display label.
	fontLabel := label.New(root, "fontlabel", app,
		label.Text("Selected: (none)"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(fontLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	// Preview label.
	previewLabel := label.New(root, "preview", app,
		label.Text("The quick brown fox jumps over the lazy dog."),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(previewLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	chooseBtn := button.New(root, "choose", app,
		button.Text("Choose Font..."),
		button.Command(func() {
			fontDesc, ok := dialog.ChooseFont(app,
				dialog.FontParent(root),
			)
			if ok {
				fontLabel.Text = fmt.Sprintf("Selected: %s", fontDesc)
				fontLabel.Display()
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(chooseBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(30), pack.PadY(20))

	_ = fontLabel
	_ = previewLabel
	_ = chooseBtn
	d.Run()
}
