// Demo: Toplevel window containing several labelframe widgets.
// Ported from Tk's labelframe.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/checkbutton"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/labelframe"
	"github.com/takigo/takigo/widget/radiobutton"
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
		label.WrapLength(screenunit.In(4)),
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
		labelframe.PadX(screenunit.Pt(1.5)),
		labelframe.PadY(screenunit.Pt(1.5)),
	)
	grid.Grid(lfValue, grid.Row(0), grid.Column(0), grid.PadX(screenunit.Mm(2)), grid.PadY(screenunit.Mm(2)))

	valueVar := widget.NewUnsetVariable[string]() // lfdummy starts unset
	for _, v := range []string{"1", "2", "3", "4"} {
		rb := radiobutton.New(lfValue, "b"+v,
			radiobutton.Text("This is value "+v),
			radiobutton.Value(v),
			radiobutton.Var(valueVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(screenunit.Pt(1.5)))
		_ = rb
	}

	// Right labelframe: checkbutton as labelwidget controls enable/disable of options.
	lfOpts := labelframe.New(body, "f2",
		labelframe.PadX(screenunit.Pt(1.5)),
		labelframe.PadY(screenunit.Pt(1.5)),
	)
	grid.Grid(lfOpts, grid.Row(0), grid.Column(1), grid.PadX(screenunit.Mm(2)), grid.PadY(screenunit.Mm(2)))

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
		pack.Pack(cb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(screenunit.Pt(1.5)))
		optionCbs = append(optionCbs, cb)
	}

	// Enable/disable callback — toggle option checkbuttons.
	enableCb.Command = func() {
		for _, cb := range optionCbs {
			if enableVar.Get() == "1" {
				cb.Configure(checkbutton.State(widget.StateNormal))
			} else {
				cb.Configure(checkbutton.State(widget.StateDisabled))
			}
		}
	}

	// Initially disable all option checkbuttons.
	for _, cb := range optionCbs {
		cb.Configure(checkbutton.State(widget.StateDisabled))
	}

	grid.ColumnConfigure(body, 0, grid.Weight(1))
	grid.ColumnConfigure(body, 1, grid.Weight(1))

	_ = enableCb
	_ = lfValue
	_ = lfOpts
	app.Run()
}
