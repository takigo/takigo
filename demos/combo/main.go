// Demo: TTK Combobox with editable, readonly, and disabled states.
// Ported from Tk's combo.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Combobox Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("combo"),
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
		label.Text("Three different combo-boxes are displayed below. You can add characters to the first one by pointing, clicking and typing, just as with an entry; pressing Return will cause the current value to be added to the list that is selectable from the drop-down list, and you can choose other values by pressing the Down key, using the arrow keys to pick another one, and pressing Return again. The second combo-box is fixed to a particular value, and cannot be modified at all. The third one only allows you to select values from its drop-down list of Australian cities."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	ttk.SetCurrentTheme("clam")

	// Inner frame (matches Tcl's ttk::frame $w.f).
	body := ttk.NewFrame(f, "body")
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	cities := []string{
		"Canberra", "Sydney", "Melbourne", "Perth",
		"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
	}

	// Editable combobox in labelframe (starts empty, no initial values).
	editFrame := labelframe.New(body, "c1", labelframe.Text("Fully Editable"))
	pack.Pack(editFrame, pack.SideOpt(pack.Top), pack.PadY("3p"), pack.PadX("7.5p"))

	editCombo := ttk.NewCombobox(editFrame, "c")
	pack.Pack(editCombo, pack.PadY("3p"), pack.PadX("7.5p"))

	// Disabled combobox in labelframe.
	disFrame := labelframe.New(body, "c2", labelframe.Text("Disabled"))
	pack.Pack(disFrame, pack.SideOpt(pack.Top), pack.PadY("3p"), pack.PadX("7.5p"))

	disCombo := ttk.NewCombobox(disFrame, "c",
		ttk.ComboboxText("unchangable"),
		ttk.ComboboxCbState(ttk.ComboDisabled),
	)
	pack.Pack(disCombo, pack.PadY("3p"), pack.PadX("7.5p"))

	// Readonly combobox in labelframe.
	roFrame := labelframe.New(body, "c3", labelframe.Text("Defined List Only"))
	pack.Pack(roFrame, pack.SideOpt(pack.Top), pack.PadY("3p"), pack.PadX("7.5p"))

	roCombo := ttk.NewCombobox(roFrame, "c",
		ttk.ComboboxValues(cities),
		ttk.ComboboxText("Sydney"),
		ttk.ComboboxCbState(ttk.ComboReadonly),
	)
	pack.Pack(roCombo, pack.PadY("3p"), pack.PadX("7.5p"))

	_ = editFrame
	_ = editCombo
	_ = disFrame
	_ = disCombo
	_ = roFrame
	_ = roCombo
	app.Run()
}
