// Demo: Simple Ttk widgets, such as labels, labelframes, buttons, checkbuttons,
// radiobuttons, a separator and a toggleswitch.
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
		label.Text("Ttk is the new Tk themed widget set. This is a Ttk themed label, "+
			"and below are four groups of Ttk widgets in Ttk labelframes. "+
			"The first group are all buttons that set the current application "+
			"theme when pressed. The second group contains two sets of "+
			"checkbuttons, with a separator widget between the sets. The third "+
			"group has a collection of linked radiobuttons. Finally, the "+
			"toggleswitch in the fourth labelframe controls whether all the "+
			"themed widgets in this toplevel, except that labelframe and its "+
			"children, are in the disabled state."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Container frame for the grid layout.
	container := ttk.NewFrame(f, "container")
	pack.Pack(container, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	// Collect all TTK widgets for enable/disable toggling.
	var ttkWidgets []*ttk.TtkWidget

	// -- Group 1: Buttons (theme switchers) --
	btnFrame := labelframe.New(container, "buttons",
		labelframe.Text("Buttons"),
	)

	themes := ttk.ThemeNames()
	sort.Strings(themes)
	for _, theme := range themes {
		themeName := theme
		btn := ttk.NewButton(btnFrame, themeName,
			ttk.ButtonText(themeName),
			ttk.ButtonCommand(func() {
				ttk.SetCurrentTheme(themeName)
			}),
		)
		pack.Pack(btn, pack.PadY("1.5p"))
		ttkWidgets = append(ttkWidgets, &btn.TtkWidget)
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
	ttkWidgets = append(ttkWidgets, &c1.TtkWidget, &c2.TtkWidget)

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
	ttkWidgets = append(ttkWidgets, &c3.TtkWidget, &c4.TtkWidget)

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
		ttkWidgets = append(ttkWidgets, &r.TtkWidget)
	}

	// -- Group 4: Toggleswitch (enable/disable all widgets) --
	togFrame := labelframe.New(container, "toggle",
		labelframe.Text("Toggleswitch"),
	)

	enabled := widget.NewVariable(true)

	togLabel := ttk.NewLabel(togFrame, "l",
		ttk.LabelText("Enable/disable widgets"),
	)
	togSwitch := ttk.NewToggleswitch(togFrame, "sw",
		ttk.ToggleswitchVar(enabled),
		ttk.ToggleswitchCommand(func() {
			for _, w := range ttkWidgets {
				if enabled.Get() {
					w.State &^= ttk.StateDisabled
				} else {
					w.State |= ttk.StateDisabled
				}
				w.Display()
			}
		}),
	)
	pack.Pack(togSwitch, pack.SideOpt(pack.Right), pack.PadX("3p"), pack.PadY("1.5p"))
	pack.Pack(togLabel, pack.SideOpt(pack.Left), pack.PadX("3p"), pack.PadY("1.5p"))

	// -- Grid layout: buttons span 2 rows; toggle in row 1 cols 1-2 --
	nwe := grid.StickN | grid.StickW | grid.StickE
	grid.Grid(btnFrame, grid.Row(0), grid.Column(0), grid.RowSpan(2),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(chkFrame, grid.Row(0), grid.Column(1),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(radFrame, grid.Row(0), grid.Column(2),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))
	grid.Grid(togFrame, grid.Row(1), grid.Column(1), grid.ColumnSpan(2),
		grid.Sticky(nwe), grid.PadX("3p"), grid.PadY("1.5p"))

	// Equal column weights with uniform sizing.
	grid.ColumnConfigure(container, 0, grid.Weight(1), grid.Uniform("yes"))
	grid.ColumnConfigure(container, 1, grid.Weight(1), grid.Uniform("yes"))
	grid.ColumnConfigure(container, 2, grid.Weight(1), grid.Uniform("yes"))
	grid.RowConfigure(container, 1, grid.Weight(1))

	_ = sep
	_ = togLabel
	_ = togSwitch
	app.Run()
}
