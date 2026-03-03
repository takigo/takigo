// Demo: TTK Combobox with editable, readonly, and disabled states.
// Ported from Tk's combo.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

func main() {
	app := demohelper.Setup("Combobox Demonstration", 450, 400,
		"Three different combo-boxes are displayed below. You can add characters to the first one by pointing, clicking and typing, just as with an entry; pressing Return will cause the current value to be added to the drop-down list. The second combo-box is fixed to a particular value, and cannot be modified at all. The third one only allows you to select values from its drop-down list of Australian cities.")

	ttk.SetCurrentTheme("clam")

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("Selection: (none)"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	cities := []string{
		"Canberra", "Sydney", "Melbourne", "Perth",
		"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
	}

	// Editable combobox in labelframe.
	editFrame := labelframe.New(app, "c1", labelframe.Text("Fully Editable"))
	pack.Pack(editFrame, pack.SideOpt(pack.Top), pack.PadY(3), pack.PadX(8))

	editCombo := ttk.NewCombobox(editFrame, "c",
		ttk.ComboboxValues(cities),
		ttk.ComboboxText("Canberra"),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(editCombo, pack.PadY(3), pack.PadX(8))

	// Disabled combobox in labelframe.
	disFrame := labelframe.New(app, "c2", labelframe.Text("Disabled"))
	pack.Pack(disFrame, pack.SideOpt(pack.Top), pack.PadY(3), pack.PadX(8))

	disCombo := ttk.NewCombobox(disFrame, "c",
		ttk.ComboboxValues(cities),
		ttk.ComboboxText("Melbourne"),
		ttk.ComboboxCbState(ttk.ComboDisabled),
	)
	pack.Pack(disCombo, pack.PadY(3), pack.PadX(8))

	// Readonly combobox in labelframe.
	roFrame := labelframe.New(app, "c3", labelframe.Text("Defined List Only"))
	pack.Pack(roFrame, pack.SideOpt(pack.Top), pack.PadY(3), pack.PadX(8))

	roCombo := ttk.NewCombobox(roFrame, "c",
		ttk.ComboboxValues(cities),
		ttk.ComboboxText("Sydney"),
		ttk.ComboboxCbState(ttk.ComboReadonly),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(roCombo, pack.PadY(3), pack.PadX(8))

	_ = statusLabel
	_ = editFrame
	_ = editCombo
	_ = disFrame
	_ = disCombo
	_ = roFrame
	_ = roCombo
	app.Run()
}
