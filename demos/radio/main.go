// Demo: Radiobutton groups for selecting point size, color, and alignment.
// Ported from Tk's radio.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Radiobutton Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("radio"),
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
		label.Text("Three groups of radiobuttons are displayed below. If you click on a button then the button will become selected exclusively among all the buttons in its group. A variable is associated with each group to indicate which of the group's buttons is selected."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Variables.
	sizeVar := widget.NewVariable("12")
	colorVar := widget.NewVariable("red")
	alignVar := widget.NewVariable("left")

	// Inner frame for grid layout.
	body := frame.New(f, "body")
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Point Size group — spans 2 rows (matches Tcl: -rowspan 2).
	sizeFrame := labelframe.New(body, "left", labelframe.Text("Point Size"))
	grid.Grid(sizeFrame, grid.Row(0), grid.Column(0), grid.RowSpan(2),
		grid.PadX(".5c"), grid.PadY(".5c"))

	for _, s := range []string{"10", "12", "14", "18", "24"} {
		rb := radiobutton.New(sizeFrame, "size_"+s,
			radiobutton.Text("Point Size "+s),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY("1.5p"),
			pack.Anchor(option.AnchorW), pack.FillOpt(pack.FillX))
	}

	// Color group — spans 2 rows.
	colorFrame := labelframe.New(body, "mid", labelframe.Text("Color"))
	grid.Grid(colorFrame, grid.Row(0), grid.Column(1), grid.RowSpan(2),
		grid.PadX(".5c"), grid.PadY(".5c"))

	for _, c := range []string{"Red", "Green", "Blue", "Yellow", "Orange", "Purple"} {
		rb := radiobutton.New(colorFrame, "color_"+strings.ToLower(c),
			radiobutton.Text(c),
			radiobutton.Value(strings.ToLower(c)),
			radiobutton.Var(colorVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY("1.5p"), pack.FillOpt(pack.FillX))
	}

	// Alignment group — compass grid layout.
	alignFrame := labelframe.New(body, "right", labelframe.Text("Alignment"))
	grid.Grid(alignFrame, grid.Row(0), grid.Column(2),
		grid.PadX(".5c"), grid.PadY(".5c"))

	// Center label — compound changes on alignment selection.
	centerLabel := label.New(alignFrame, "l", label.Text("Label"))

	// Update center label compound when alignment changes.
	alignVar.OnChange(func(_, v string) {
		switch v {
		case "top":
			centerLabel.Compound = widget.CompoundTop
		case "left":
			centerLabel.Compound = widget.CompoundLeft
		case "right":
			centerLabel.Compound = widget.CompoundRight
		case "bottom":
			centerLabel.Compound = widget.CompoundBottom
		}
		centerLabel.Display()
	})

	for _, a := range []struct {
		text, value string
		row, col    int
	}{
		{"Top", "top", 0, 1},
		{"Left", "left", 1, 0},
		{"Right", "right", 1, 2},
		{"Bottom", "bottom", 2, 1},
	} {
		rb := radiobutton.New(alignFrame, a.value,
			radiobutton.Text(a.text),
			radiobutton.Value(a.value),
			radiobutton.Var(alignVar),
			radiobutton.IndicatorOnOpt(false),
		)
		grid.Grid(rb, grid.Row(a.row), grid.Column(a.col))
	}
	grid.Grid(centerLabel, grid.Row(1), grid.Column(1))

	// Tristate button (column 2, row 1 — below alignment group).
	tristateBtn := button.New(body, "tristate",
		button.Text("Tristate"),
		button.Command(func() {
			sizeVar.Set("multi")
			colorVar.Set("multi")
		}),
	)
	grid.Grid(tristateBtn, grid.Row(1), grid.Column(2),
		grid.PadX(".5c"), grid.PadY(".5c"))

	_ = sizeVar
	_ = colorVar
	_ = alignVar
	_ = sizeFrame
	_ = colorFrame
	_ = alignFrame
	_ = centerLabel
	_ = tristateBtn
	app.Run()
}
