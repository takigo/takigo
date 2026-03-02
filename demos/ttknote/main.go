// Demo: TTK Notebook with multiple tabbed pages.
// Ported from Tk's ttknote.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Notebook Demonstration", 500, 350, "A notebook widget with three tabs. Click each tab\nto switch between pages.")
	app := d.App

	ttk.SetCurrentTheme("clam")

	// Notebook.
	nb := ttk.NewNotebook(app, "nb")
	pack.Pack(nb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(15), pack.PadY(10))

	// Tab 1: Description.
	page1 := frame.New(nb, "page1")
	descLabel := label.New(page1, "desc",
		label.Text("This is the first tab.\n\nNotebook widgets allow you to\norganize content into tabs.\nClick on a tab to view its content."),
		label.Anchor(option.AnchorNW),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(descLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	nb.Add(page1.Window(), "Description")

	// Tab 2: Buttons.
	page2 := frame.New(nb, "page2")

	statusLabel := label.New(page2, "status2",
		label.Text("Click a button:"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	for i, text := range []string{"Button A", "Button B", "Button C"} {
		btnText := text
		btn := ttk.NewButton(page2, fmt.Sprintf("btn%d", i),
			ttk.ButtonText(btnText),
			ttk.ButtonCommand(func() {
				statusLabel.Text = "Clicked: " + btnText
				statusLabel.Display()
			}),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(3))
		_ = btn
	}
	nb.Add(page2.Window(), "Buttons")

	// Tab 3: Labels.
	page3 := frame.New(nb, "page3")
	for _, text := range []string{"Label One", "Label Two", "Label Three"} {
		l := ttk.NewLabel(page3, "l_"+text,
			ttk.LabelText(text),
		)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
		_ = l
	}
	nb.Add(page3.Window(), "Labels")

	_ = descLabel
	_ = statusLabel
	d.Run()
}
