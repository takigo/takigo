// Demo: Toplevel window containing several labelframe widgets.
// Ported from Tk's labelframe.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Labelframe Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("labelframe"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Labelframes are used to group related widgets together.  The label may be either  plain text or another widget."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Demo area frame.
	body := frame.New(f, "f")
	pack.Pack(body, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Left labelframe: "Value" with radiobuttons 1-4.
	lfValue := labelframe.New(body, "f",
		labelframe.Text("Value"),
		labelframe.PadX("1.5p"),
		labelframe.PadY("1.5p"),
	)
	grid.Grid(lfValue, grid.Row(0), grid.Column(0), grid.PadX("2m"), grid.PadY("2m"))

	valueVar := widget.NewUnsetVariable[string]() // lfdummy starts unset
	for _, v := range []string{"1", "2", "3", "4"} {
		rb := radiobutton.New(lfValue, "b"+v,
			radiobutton.Text("This is value "+v),
			radiobutton.Value(v),
			radiobutton.Var(valueVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
		_ = rb
	}

	// Right labelframe: checkbutton as labelwidget controls enable/disable of options.
	lfOpts := labelframe.New(body, "f2",
		labelframe.PadX("1.5p"),
		labelframe.PadY("1.5p"),
	)
	grid.Grid(lfOpts, grid.Row(0), grid.Column(1), grid.PadX("2m"), grid.PadY("2m"))

	// Option checkbuttons.
	var optionCbs []*checkbutton.Checkbutton
	enableVar := widget.NewVariable("0")

	enableCb := checkbutton.New(lfOpts, "cb",
		checkbutton.Text("Use this option."),
		checkbutton.Var(enableVar),
		checkbutton.PadX(0),
	)
	// Use checkbutton as the labelwidget (like Tcl's -labelwidget).
	lfOpts.SetLabelWidget(enableCb)

	for i, s := range []string{"Option1", "Option2", "Option3"} {
		v := widget.NewVariable("0")
		cb := checkbutton.New(lfOpts, fmt.Sprintf("b%d", i),
			checkbutton.Text(s),
			checkbutton.Var(v),
		)
		pack.Pack(cb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
		optionCbs = append(optionCbs, cb)
	}

	// Enable/disable callback — toggle option checkbuttons.
	enableCb.Command = func() {
		for _, cb := range optionCbs {
			if enableVar.Get() == "1" {
				cb.State = widget.StateNormal
			} else {
				cb.State = widget.StateDisabled
			}
			cb.Display()
		}
	}

	// Initially disable all option checkbuttons.
	for _, cb := range optionCbs {
		cb.State = widget.StateDisabled
		cb.Display()
	}

	grid.ColumnConfigure(body, 0, grid.Weight(1))
	grid.ColumnConfigure(body, 1, grid.Weight(1))

	_ = enableCb
	_ = lfValue
	_ = lfOpts
	app.Run()
}
