// Demo: Themed Entry widgets — plain, placeholder, password, textvariable,
// and integer-only validation.
// Ported from Tk's entry1.tcl / entry3.tcl demos but using ttk::entry.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Themed Entry Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkentry"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	ttk.SetCurrentTheme("clam")

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.Text("Five themed entries are displayed below. You can edit, select, "+
			"and copy/paste using the standard keyboard bindings. The third entry "+
			"masks its contents (password). The fourth is bound to a string variable "+
			"and updates a label as you type. The fifth accepts only integers; "+
			"non-digit input is rejected."),
		label.JustifyOpt(option.JustifyLeft),
		label.WrapLength(screenunit.In(5)),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// 1. Plain entry with initial value.
	e1 := ttk.NewEntry(f, "e1",
		ttk.EntryText("Initial value"),
		ttk.EntryWidth(40),
	)
	pack.Pack(e1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)), pack.PadY(screenunit.Pt(3)))

	// 2. Placeholder entry (no initial text).
	e2 := ttk.NewEntry(f, "e2",
		ttk.EntryPlaceholder("Type here..."),
		ttk.EntryWidth(40),
	)
	pack.Pack(e2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)), pack.PadY(screenunit.Pt(3)))

	// 3. Password entry.
	e3 := ttk.NewEntry(f, "e3",
		ttk.EntryShow('*'),
		ttk.EntryWidth(40),
	)
	pack.Pack(e3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)), pack.PadY(screenunit.Pt(3)))

	// 4. Textvariable-bound entry that updates a label as the user types.
	statusVar := widget.NewVariable[string]("")
	e4 := ttk.NewEntry(f, "e4",
		ttk.EntryTextVariable(statusVar),
		ttk.EntryWidth(40),
	)
	pack.Pack(e4, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)), pack.PadY(screenunit.Pt(3)))

	liveLabel := label.New(f, "live", label.Text("(live)"))
	pack.Pack(liveLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)))
	statusVar.OnChange(func(_, val string) {
		liveLabel.Configure(label.Text("(live) " + val))
	})

	// 5. Integer-only entry with -validate "key".
	e5 := ttk.NewEntry(f, "e5",
		ttk.EntryText("42"),
		ttk.EntryWidth(20),
		ttk.EntryValidate(ttk.ValidateKey),
		ttk.EntryValidateCmd(func(s string) bool {
			if s == "" {
				return true
			}
			_, err := strconv.Atoi(s)
			return err == nil
		}),
	)
	pack.Pack(e5, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(5)), pack.PadY(screenunit.Pt(3)))

	app.Run()
}
