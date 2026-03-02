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
	app := demohelper.Setup("TTK Widgets", 400, 350,
		"TTK themed widgets: labels, buttons, separators.\nThey use the clam theme for modern appearance.")

	ttk.SetCurrentTheme("clam")

	statusLabel := label.New(app, "status",
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// TTK Frame.
	ttkFrame := ttk.NewFrame(app, "ttkframe",
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	pack.Pack(ttkFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(15), pack.PadY(10))

	// TTK Label.
	ttkLabel := ttk.NewLabel(ttkFrame, "ttklabel",
		ttk.LabelText("This is a TTK Label"),
	)
	pack.Pack(ttkLabel, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(8))

	// Separator.
	sep := ttk.NewSeparator(ttkFrame, "sep")
	pack.Pack(sep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(5))

	// TTK Buttons.
	for i, text := range []string{"TTK Button 1", "TTK Button 2", "TTK Button 3"} {
		idx := i + 1
		btnText := text
		btn := ttk.NewButton(ttkFrame, fmt.Sprintf("ttkbtn%d", idx),
			ttk.ButtonText(btnText),
			ttk.ButtonCommand(func() {
				setStatus(fmt.Sprintf("Clicked: %s", btnText))
			}),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
		_ = btn
	}

	_ = statusLabel
	_ = ttkLabel
	_ = sep
	app.Run()
}
