// Demo: Radiobutton groups for selecting point size, color, and alignment.
// Ported from Tk's radio.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry"
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
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Three groups of radiobuttons are displayed below. If you click on a button then the button will become selected exclusively among all the buttons in its group. A variable is associated with each group to indicate which of the group's buttons is selected."),
	)
	grid.Grid(msg, grid.Row(0), grid.Column(0), grid.ColumnSpan(3), grid.Sticky(grid.NSEW))

	vars := make(demohelper.DemoVars[string])

	btns := demohelper.AddVarsSeeDismiss(f, &vars)
	grid.Grid(btns, grid.Row(3), grid.Column(0), grid.ColumnSpan(3), grid.Sticky(grid.EW))

	// Variables
	sizeVar := widget.NewVariable("12")
	colorVar := widget.NewVariable("red")
	alignVar := widget.NewVariable("top")
	vars["size"] = sizeVar
	vars["color"] = colorVar
	vars["align"] = alignVar

	left := labelframe.New(f, "left",
		labelframe.Text("Point Size"),
		labelframe.PadX("1.5p"),
		labelframe.PadY("1.5p"),
	)
	mid := labelframe.New(f, "mid",
		labelframe.Text("Color"),
		labelframe.PadX("1.5p"),
		labelframe.PadY("1.5p"),
	)
	right := labelframe.New(f, "right",
		labelframe.Text("Alignment"),
		labelframe.PadX("1.5p"),
		labelframe.PadY("1.5p"),
	)
	tristate := button.New(f, "tristate",
		button.Text("Tristate"),
		button.PadX("1.5p"),
		button.PadY("1.5p"),
		button.Command(func() {
			sizeVar.Set("multi")
			colorVar.Set("multi")
		}),
	)
	grid.Grid(left, grid.Column(0), grid.Row(1),
		grid.PadX(".5c"), grid.PadY(".5c"),
		grid.RowSpan(2),
	)
	grid.Grid(mid, grid.Column(1), grid.Row(1),
		grid.PadX(".5c"), grid.PadY(".5c"),
		grid.RowSpan(2),
	)
	grid.Grid(right, grid.Column(2), grid.Row(1),
		grid.PadX(".5c"), grid.PadY(".5c"),
	)
	grid.Grid(tristate, grid.Column(2), grid.Row(2),
		grid.PadX(".5c"), grid.PadY(".5c"),
	)

	for _, s := range []string{"10", "12", "14", "18", "24"} {
		rb := radiobutton.New(left, "size_"+s,
			radiobutton.Text("Point Size "+s),
			radiobutton.Value(s),
			radiobutton.Var(sizeVar),
			radiobutton.TristateValueOpt("multi"),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY("1.5p"),
			pack.Anchor(option.AnchorW), pack.FillOpt(pack.FillX))
	}

	for _, c := range []string{"Red", "Green", "Blue", "Yellow", "Orange", "Purple"} {
		colorName := strings.ToLower(c)
		rb := radiobutton.New(mid, "color_"+colorName,
			radiobutton.Text(c),
			radiobutton.Value(colorName),
			radiobutton.Var(colorVar),
			radiobutton.TristateValueOpt("multi"),
			radiobutton.Command(func() {
				col, err := app.ColorCache().Get(colorName)
				if err == nil {
					mid.Foreground = col
					mid.Display()
				}
			}),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY("1.5p"), pack.FillOpt(pack.FillX))
	}

	l := label.New(right, "l", label.Text("Label"),
		label.Bitmap("questhead"),
		label.CompoundOpt(widget.CompoundTop))
	// $w.right.l configure -width [winfo reqwidth $w.right.l] -compound top
	// $w.right.l configure -height [winfo reqheight $w.right.l]

	// Update center label compound when alignment changes.
	alignVar.OnChange(func(_, v string) {
		switch v {
		case "top":
			l.Compound = widget.CompoundTop
		case "left":
			l.Compound = widget.CompoundLeft
		case "right":
			l.Compound = widget.CompoundRight
		case "bottom":
			l.Compound = widget.CompoundBottom
		}
		l.Display()
	})

	rightButtons := make(map[string]*radiobutton.Radiobutton)
	for _, a := range []struct {
		text, value string
		row, col    int
	}{
		{"Top", "top", 0, 1},
		{"Left", "left", 1, 0},
		{"Right", "right", 1, 2},
		{"Bottom", "bottom", 2, 1},
	} {
		rb := radiobutton.New(right, a.value,
			radiobutton.Text(a.text),
			radiobutton.Value(a.value),
			radiobutton.Var(alignVar),
			radiobutton.IndicatorOnOpt(false),
		)
		rightButtons[a.value] = rb
	}
	grid.Grid(geometry.Group{grid.Relative(grid.RelEmpty), rightButtons["top"]})
	grid.Grid(geometry.Group{rightButtons["left"], l, rightButtons["right"]})
	grid.Grid(geometry.Group{grid.Relative(grid.RelEmpty), rightButtons["bottom"]})

	app.Run()
}
