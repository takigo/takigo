// Demo: TTK buttons, checkbuttons, radiobuttons, and separator.
// Ported from Tk's ttkbut.tcl demo.
package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Simple Ttk Widgets"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkbut"),
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
		label.Text("Ttk is the themed widget set. This is a Ttk themed label, "+
			"and below are four groups of Ttk widgets in Ttk labelframes. "+
			"The first group are buttons that set the current application "+
			"theme when pressed. The second group contains two sets of "+
			"checkbuttons, with a separator between the sets. The third "+
			"group has a collection of linked radiobuttons. The fourth "+
			"group has a collection of toggle switches."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	ttk.SetCurrentTheme("clam")

	// Container frame for the grid layout (matches Tcl's ttk::frame $w.f).
	container := ttk.NewFrame(f, "container")
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

	c1 := ttk.NewCheckbutton(chkFrame, "c1",
		ttk.CheckbuttonText("Cheese"),
		ttk.CheckbuttonVar(cheese),
	)
	c2 := ttk.NewCheckbutton(chkFrame, "c2",
		ttk.CheckbuttonText("Tomato"),
		ttk.CheckbuttonVar(tomato),
	)
	pack.Pack(c1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
	pack.Pack(c2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	sep := ttk.NewSeparator(chkFrame, "sep")
	pack.Pack(sep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	c3 := ttk.NewCheckbutton(chkFrame, "c3",
		ttk.CheckbuttonText("Basil"),
		ttk.CheckbuttonVar(basil),
	)
	c4 := ttk.NewCheckbutton(chkFrame, "c4",
		ttk.CheckbuttonText("Oregano"),
		ttk.CheckbuttonVar(oregano),
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
		r := ttk.NewRadiobutton(radFrame, fmt.Sprintf("r%d", i+1),
			ttk.RadiobuttonText(item.text),
			ttk.RadiobuttonValue(item.value),
			ttk.RadiobuttonVar(happiness),
		)
		pack.Pack(r, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadX("3p"), pack.PadY("1.5p"))
		_ = r
	}

	// -- Group 4: Toggle switches --
	togFrame := labelframe.New(container, "toggles",
		labelframe.Text("Toggleswitches"),
	)

	sw1var := widget.NewVariable(true)
	sw2var := widget.NewVariable(false)
	sw3var := widget.NewVariable(true)

	sw1 := ttk.NewToggleswitch(togFrame, "sw1",
		ttk.ToggleswitchText("Lights"),
		ttk.ToggleswitchVar(sw1var),
	)
	sw2 := ttk.NewToggleswitch(togFrame, "sw2",
		ttk.ToggleswitchText("Music"),
		ttk.ToggleswitchVar(sw2var),
	)
	sw3 := ttk.NewToggleswitch(togFrame, "sw3",
		ttk.ToggleswitchText("Alarm"),
		ttk.ToggleswitchVar(sw3var),
	)
	pack.Pack(sw1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
	pack.Pack(sw2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))
	pack.Pack(sw3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY("1.5p"))

	// -- Arrange the four groups in a grid row (matches Tcl's grid arrangement) --
	nwe := grid.StickN | grid.StickW | grid.StickE
	grid.Grid(btnFrame, grid.Row(0), grid.Column(0),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(chkFrame, grid.Row(0), grid.Column(1),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(radFrame, grid.Row(0), grid.Column(2),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(togFrame, grid.Row(0), grid.Column(3),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))

	// Equal column weights with uniform sizing.
	grid.ColumnConfigure(container, 0, grid.Weight(1), grid.Uniform("yes"))
	grid.ColumnConfigure(container, 1, grid.Weight(1), grid.Uniform("yes"))
	grid.ColumnConfigure(container, 2, grid.Weight(1), grid.Uniform("yes"))
	grid.ColumnConfigure(container, 3, grid.Weight(1), grid.Uniform("yes"))
	grid.RowConfigure(container, 0, grid.Weight(1))

	_ = c1
	_ = c2
	_ = c3
	_ = c4
	_ = sep
	_ = sw1
	_ = sw2
	_ = sw3
	_ = sw1var
	_ = sw2var
	_ = sw3var
	app.Run()
}
