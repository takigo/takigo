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
)

func main() {
	d := demohelper.Setup("Combobox Demonstration", 450, 400,
		"Three comboboxes are shown below: editable,\nreadonly, and disabled. Click the arrow to see the\ndropdown list.")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Selection: (none)"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	countries := []string{
		"Australia", "Canada", "France", "Germany",
		"Japan", "United Kingdom", "United States",
	}

	// Editable combobox.
	editLabel := label.New(root, "editlabel", app,
		label.Text("Editable:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(editLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	editCombo := ttk.NewCombobox(root, "editcombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("Australia"),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(editCombo, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Readonly combobox.
	roLabel := label.New(root, "rolabel", app,
		label.Text("Readonly:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(roLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	roCombo := ttk.NewCombobox(root, "rocombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("Canada"),
		ttk.ComboboxCbState(ttk.ComboReadonly),
		ttk.ComboboxCommand(func(v string) {
			statusLabel.Text = "Selection: " + v
			statusLabel.Display()
		}),
	)
	pack.Pack(roCombo, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Disabled combobox.
	disLabel := label.New(root, "dislabel", app,
		label.Text("Disabled:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(disLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	disCombo := ttk.NewCombobox(root, "discombo", app,
		ttk.ComboboxValues(countries),
		ttk.ComboboxText("France"),
		ttk.ComboboxCbState(ttk.ComboDisabled),
	)
	pack.Pack(disCombo, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	_ = statusLabel
	_ = editLabel
	_ = editCombo
	_ = roLabel
	_ = roCombo
	_ = disLabel
	_ = disCombo
	d.Run()
}
