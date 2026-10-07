// Demo: Several combobox widgets.
// Ported from Tk's combo.tcl demo.
package main

import (
	"fmt"
	"os"
	"slices"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
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

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength(screenunit.In(5)),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelText("Three different combo-boxes are displayed below. You can add characters to the first one by pointing, clicking and typing, just as with an entry; pressing Return will cause the current value to be added to the list that is selectable from the drop-down list, and you can choose other values by pressing the Down key, using the arrow keys to pick another one, and pressing Return again. The second combo-box is fixed to a particular value, and cannot be modified at all. The third one only allows you to select values from its drop-down list of Australian cities."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	firstValue := widget.NewVariable[string]("")
	secondValue := widget.NewVariable[string]("unchangable") //nolint:misspell // Tk's combo.tcl spells it so
	ozCity := widget.NewVariable[string]("Sydney")
	vars := []demohelper.NamedVar{
		{Name: "firstValue", Var: firstValue},
		{Name: "secondValue", Var: secondValue},
		{Name: "ozCity", Var: ozCity},
	}
	btns := demohelper.AddSeeDismissWithVars(f, vars)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	body := ttk.NewFrame(f, "f")
	pack.Pack(body, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	cities := []string{
		"Canberra", "Sydney", "Melbourne", "Perth",
		"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
	}

	editFrame := ttk.NewLabelframe(body, "c1", ttk.LabelframeText("Fully Editable"))
	pack.Pack(editFrame, pack.SideOpt(pack.Top), pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	editCombo := ttk.NewCombobox(editFrame, "c",
		ttk.ComboboxPlaceholder("Enter text here"),
	)
	pack.Pack(editCombo, pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	disFrame := ttk.NewLabelframe(body, "c2", ttk.LabelframeText("Disabled"))
	pack.Pack(disFrame, pack.SideOpt(pack.Top), pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	disCombo := ttk.NewCombobox(disFrame, "c",
		ttk.ComboboxText(secondValue.Get()),
		ttk.ComboboxState(ttk.FieldDisabled),
	)
	pack.Pack(disCombo, pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	roFrame := ttk.NewLabelframe(body, "c3", ttk.LabelframeText("Defined List Only"))
	pack.Pack(roFrame, pack.SideOpt(pack.Top), pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	roCombo := ttk.NewCombobox(roFrame, "c",
		ttk.ComboboxValues(cities),
		ttk.ComboboxText(ozCity.Get()),
		ttk.ComboboxState(ttk.FieldReadonly),
	)
	pack.Pack(roCombo, pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	app.Dispatcher().Bind(editCombo.Win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Return {
			cur := editCombo.Get()
			if cur == "" {
				return
			}
			if slices.Contains(editCombo.Values, cur) {
				return
			}
			editCombo.Values = append(editCombo.Values, cur)
		}
	})

	_ = disCombo
	_ = roCombo
	app.Run()
}
