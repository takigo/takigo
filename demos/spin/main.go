// Demo: Spinbox widgets with different value modes.
// Ported from Tk's spin.tcl demo.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Spinbox Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("spin"),
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
		label.Text("Three different spin-boxes are displayed below. You can add characters by pointing, clicking and typing. The normal Motif editing characters are supported, along with many Emacs bindings. For example, Backspace and Control-h delete the character to the left of the insertion cursor and Delete and Control-d delete the chararacter to the right of the insertion cursor. For values that are too large to fit in the window all at once, you can scan through the value by dragging with mouse button2 pressed. Note that the first spin-box will only permit you to type in integers, and the third selects from a list of Australian cities."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	padX := screenunit.Px("7.5p")
	padY := screenunit.Px("3p")

	s1 := spinbox.New(f, "s1",
		spinbox.FromOpt(1),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(1),
		spinbox.WidthOpt(10),
		spinbox.ValidateOpt("key"),
		spinbox.ValidateCmdOpt(func(s string) bool {
			if s == "" {
				return true
			}
			_, err := strconv.Atoi(s)
			return err == nil
		}),
	)
	s1.SetText("1")
	pack.Pack(s1, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	s2 := spinbox.New(f, "s2",
		spinbox.FromOpt(0),
		spinbox.ToOpt(3),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%05.2f"),
		spinbox.WidthOpt(10),
	)
	s2.SetText("00.00")
	pack.Pack(s2, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	s3 := spinbox.New(f, "s3",
		spinbox.ValuesOpt([]string{
			"Canberra", "Sydney", "Melbourne", "Perth",
			"Adelaide", "Brisbane", "Hobart", "Darwin", "Alice Springs",
		}),
		spinbox.WidthOpt(10),
	)
	s3.SetText("Canberra")
	pack.Pack(s3, pack.SideOpt(pack.Top), pack.PadX(padX), pack.PadY(padY))

	_ = s1
	_ = s2
	_ = s3
	app.Run()
}
