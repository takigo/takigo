// Demo: TTK frame with paned windows.
// Ported from Tk's ttkpane.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	d := demohelper.Setup("TTK + Paned Windows", 500, 350,
		"TTK frames inside a paned window.\nDrag the sash to resize.")
	app := d.App

	ttk.SetCurrentTheme("clam")

	pw := panedwindow.New(app, "pw",
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Left pane: TTK frame with label.
	leftFrame := ttk.NewFrame(pw, "left",
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	leftLabel := ttk.NewLabel(leftFrame, "leftlbl",
		ttk.LabelText("TTK Frame (Left Pane)"),
	)
	pack.Pack(leftLabel, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))
	pw.Add(leftFrame.Window(), 150)

	// Right pane: TTK frame with button.
	rightFrame := ttk.NewFrame(pw, "right",
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	rightLabel := ttk.NewLabel(rightFrame, "rightlbl",
		ttk.LabelText("TTK Frame (Right Pane)"),
	)
	pack.Pack(rightLabel, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	ttkBtn := ttk.NewButton(rightFrame, "ttkbtn",
		ttk.ButtonText("TTK Button"),
		ttk.ButtonCommand(func() { fmt.Println("TTK button clicked!") }),
	)
	pack.Pack(ttkBtn, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
	pw.Add(rightFrame.Window(), 150)

	_ = leftLabel
	_ = rightLabel
	_ = ttkBtn
	d.Run()
}
