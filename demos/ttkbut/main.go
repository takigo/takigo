// Demo: TTK buttons, labels, and separator.
// Ported from Tk's ttkbut.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("TTK Widgets", 400, 350,
		"TTK themed widgets: labels, buttons, separators.\nThey use the clam theme for modern appearance.")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// TTK Frame.
	ttkFrame := ttk.NewFrame(root, "ttkframe", app,
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	pack.Pack(ttkFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(15), pack.PadY(10))

	// TTK Label.
	ttkLabel := ttk.NewLabel(ttkFrame.Window(), "ttklabel", app,
		ttk.LabelText("This is a TTK Label"),
	)
	pack.Pack(ttkLabel.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(8))

	// Separator.
	sep := ttk.NewSeparator(ttkFrame.Window(), "sep", app)
	pack.Pack(sep.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(5))

	// TTK Buttons.
	for i, text := range []string{"TTK Button 1", "TTK Button 2", "TTK Button 3"} {
		idx := i + 1
		btnText := text
		btn := ttk.NewButton(ttkFrame.Window(), fmt.Sprintf("ttkbtn%d", idx), app,
			ttk.ButtonText(btnText),
			ttk.ButtonCommand(func() {
				setStatus(fmt.Sprintf("Clicked: %s", btnText))
			}),
		)
		pack.Pack(btn.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
		_ = btn
	}

	_ = statusLabel
	_ = ttkLabel
	_ = sep
	d.Run()
}
