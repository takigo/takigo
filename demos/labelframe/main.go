// Demo: Labelframes with text labels containing check/radio buttons.
// Ported from Tk's labelframe.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app := demohelper.Setup("Labelframe Demonstration", 500, 300, "Labelframes are used to group related widgets together. The label may be either plain text or another widget.")

	// Demo area frame — packs at bottom fill both expand.
	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Left labelframe: "Value" with radiobuttons 1-4.
	lfValue := labelframe.New(f, "lf",
		labelframe.Text("Value"),
	)
	grid.Grid(lfValue, grid.Row(0), grid.Column(0), grid.PadX("2m"), grid.PadY("2m"))

	valueVar := widget.NewVariable("1")
	for _, v := range []string{"1", "2", "3", "4"} {
		rb := radiobutton.New(lfValue, "b"+v,
			radiobutton.Text("This is value "+v),
			radiobutton.Value(v),
			radiobutton.Var(valueVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
		_ = rb
	}

	// Right labelframe: "Options" with checkbuttons.
	// (Tcl uses a checkbutton as the labelwidget; we use plain text instead.)
	lfOpts := labelframe.New(f, "lf2",
		labelframe.Text("Options"),
	)
	grid.Grid(lfOpts, grid.Row(0), grid.Column(1), grid.PadX("2m"), grid.PadY("2m"))

	enableVar := widget.NewVariable(false)
	enableCb := checkbutton.New(lfOpts, "cb",
		checkbutton.Text("Use this option."),
		checkbutton.Var(enableVar),
	)
	pack.Pack(enableCb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	for _, s := range []string{"Option1", "Option2", "Option3"} {
		v := widget.NewVariable(false)
		cb := checkbutton.New(lfOpts, "opt_"+s,
			checkbutton.Text(s),
			checkbutton.Var(v),
		)
		pack.Pack(cb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
		_ = cb
	}

	grid.ColumnConfigure(f.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(f.Window(), 1, grid.SlotConfig{Weight: 1})

	_ = enableCb
	_ = lfValue
	_ = lfOpts
	app.Run()
}
