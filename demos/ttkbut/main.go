// Demo: TTK buttons, checkbuttons, radiobuttons, and separator.
// Ported from Tk's ttkbut.tcl demo.
package main

import (
	"fmt"
	"sort"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app := demohelper.Setup("Simple Ttk Widgets", 600, 400,
		"Ttk is the themed widget set. This is a Ttk themed label, "+
			"and below are four groups of Ttk widgets in Ttk labelframes. "+
			"The first group are buttons that set the current application "+
			"theme when pressed. The second group contains two sets of "+
			"checkbuttons, with a separator between the sets. The third "+
			"group has a collection of linked radiobuttons.")

	ttk.SetCurrentTheme("clam")

	// Container frame for the grid layout (matches Tcl's ttk::frame $w.f).
	container := ttk.NewFrame(app, "container")
	pack.Pack(container, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// -- Group 1: Buttons (theme switchers) --
	btnFrame := labelframe.New(container, "buttons",
		labelframe.Text("Buttons"),
	)

	themes := ttk.ThemeNames()
	sort.Strings(themes)
	for i, theme := range themes {
		themeName := theme
		btn := ttk.NewButton(btnFrame, fmt.Sprintf("theme%d", i),
			ttk.ButtonText(themeName),
			ttk.ButtonCommand(func() {
				ttk.SetCurrentTheme(themeName)
			}),
		)
		pack.Pack(btn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadY("1.5p"))
		_ = btn
	}

	// -- Group 2: Checkbuttons --
	chkFrame := labelframe.New(container, "checks",
		labelframe.Text("Checkbuttons"),
	)

	cheese := widget.NewVariable(false)
	tomato := widget.NewVariable(false)
	basil := widget.NewVariable(false)
	oregano := widget.NewVariable(false)

	c1 := checkbutton.New(chkFrame, "c1",
		checkbutton.Text("Cheese"),
		checkbutton.Var(cheese),
	)
	c2 := checkbutton.New(chkFrame, "c2",
		checkbutton.Text("Tomato"),
		checkbutton.Var(tomato),
	)
	pack.Pack(c1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
	pack.Pack(c2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	sep := ttk.NewSeparator(chkFrame, "sep")
	pack.Pack(sep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	c3 := checkbutton.New(chkFrame, "c3",
		checkbutton.Text("Basil"),
		checkbutton.Var(basil),
	)
	c4 := checkbutton.New(chkFrame, "c4",
		checkbutton.Text("Oregano"),
		checkbutton.Var(oregano),
	)
	pack.Pack(c3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
	pack.Pack(c4, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	// -- Group 3: Radiobuttons --
	radFrame := labelframe.New(container, "radios",
		labelframe.Text("Radiobuttons"),
	)

	happiness := widget.NewVariable("great")
	for i, item := range []struct{ text, value string }{
		{"Great", "great"},
		{"Good", "good"},
		{"OK", "ok"},
		{"Poor", "poor"},
		{"Awful", "awful"},
	} {
		r := radiobutton.New(radFrame, fmt.Sprintf("r%d", i+1),
			radiobutton.Text(item.text),
			radiobutton.Value(item.value),
			radiobutton.Var(happiness),
		)
		pack.Pack(r, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX("3p"), pack.PadY("1.5p"))
		_ = r
	}

	// -- Arrange the three groups in a grid row (matches Tcl's grid arrangement) --
	// Note: Tcl has a 4th toggleswitch group; without it, buttons don't need rowspan 2.
	nwe := grid.StickN | grid.StickW | grid.StickE
	grid.Grid(btnFrame, grid.Row(0), grid.Column(0),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(chkFrame, grid.Row(0), grid.Column(1),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(radFrame, grid.Row(0), grid.Column(2),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))

	// Equal column weights (Tcl uses -uniform yes; not yet supported in grid).
	grid.ColumnConfigure(container.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(container.Window(), 1, grid.SlotConfig{Weight: 1})
	grid.ColumnConfigure(container.Window(), 2, grid.SlotConfig{Weight: 1})
	grid.RowConfigure(container.Window(), 0, grid.SlotConfig{Weight: 1})

	_ = c1
	_ = c2
	_ = c3
	_ = c4
	_ = sep
	app.Run()
}
