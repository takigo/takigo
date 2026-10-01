// Demo: Themed Spinboxes with integer, float, and string values.
// Ported from Tk's ttkspin.tcl demo.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Themed Spinbox Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkspin"),
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
		ttk.LabelText("Three different themed spin-boxes are displayed below.  You can add characters by pointing, clicking and typing.  The normal Motif editing characters are supported, along with many Emacs bindings.  For example, Backspace and Control-h delete the character to the left of the insertion cursor and Delete and Control-d delete the chararacter to the right of the insertion cursor.  For values that are too large to fit in the window all at once, you can scan through the value by dragging with mouse button2 pressed.  Note that the first spin-box will only permit you to type in integers, and the third selects from a list of Australian cities."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Integer spinbox (1-10).
	inner := ttk.NewFrame(f, "f")
	pack.Pack(inner, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	s1 := ttk.NewSpinbox(inner, "s1",
		ttk.SpinboxFrom(1),
		ttk.SpinboxTo(10),
		ttk.SpinboxWidth(10),
		ttk.SpinboxValidate("key"),
		ttk.SpinboxValidateCmd(func(s string) bool {
			if s == "" {
				return true
			}
			_, err := strconv.Atoi(s)
			return err == nil
		}),
	)
	s1.Set("1")

	// Float spinbox (0-3 step 0.5).
	s2 := ttk.NewSpinbox(inner, "s2",
		ttk.SpinboxFrom(0),
		ttk.SpinboxTo(3),
		ttk.SpinboxIncrement(0.5),
		ttk.SpinboxFormat("%05.2f"),
		ttk.SpinboxWidth(10),
	)
	s2.Set("00.00")

	// Values spinbox (Australian cities).
	s3 := ttk.NewSpinbox(inner, "s3",
		ttk.SpinboxValues([]string{
			"Canberra", "Sydney", "Melbourne", "Perth",
			"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
		}),
		ttk.SpinboxWidth(10),
	)
	s3.Set("Canberra")

	pack.Pack(geometry.Group{s1, s2, s3}, pack.SideOpt(pack.Top),
		pack.PadY(screenunit.Pt(3)), pack.PadX(screenunit.Pt(7.5)))

	app.Run()
}
