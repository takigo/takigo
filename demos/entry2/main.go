// Demo: Entry widgets with scrollbars.
// Ported from Tk's entry2.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Entry Demonstration (with scrollbars)"),
		takigo.Geometry("+300+300"),
		takigo.IconName("entry2"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(5)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Three different entries are displayed below, with a scrollbar for each entry.  You can add characters by pointing, clicking and typing.  The normal Motif editing characters are supported, along with many Emacs bindings.  For example, Backspace and Control-h delete the character to the left of the insertion cursor and Delete and Control-d delete the chararacter to the right of the insertion cursor.  For entries that are too large to fit in the window all at once, you can scan through the entries with the scrollbars, or by dragging with the middle mouse button pressed."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	fr := frame.New(f, "frame",
		frame.BorderWidth(10), // 7.5p ≈ 10px
	)
	pack.Pack(fr, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.Expand(true))

	e1 := entry.New(fr, "e1")
	s1 := ttk.NewScrollbar(fr, "s1",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						e1.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					e1.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	e1.ScrollCmd = func(first, last float64) {
		s1.Set(first, last)
	}

	spacer1 := frame.New(fr, "spacer1", frame.Width(20), frame.Height(10)) // 15p x 7.5p

	e2 := entry.New(fr, "e2")
	s2 := ttk.NewScrollbar(fr, "s2",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						e2.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					e2.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	e2.ScrollCmd = func(first, last float64) {
		s2.Set(first, last)
	}

	spacer2 := frame.New(fr, "spacer2", frame.Width(20), frame.Height(10)) // 15p x 7.5p

	e3 := entry.New(fr, "e3",
		entry.Placeholder("Enter text here"),
		entry.PlaceholderForeground("gray75"),
	)
	s3 := ttk.NewScrollbar(fr, "s3",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						e3.XViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					e3.XViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	e3.ScrollCmd = func(first, last float64) {
		s3.Set(first, last)
	}

	e1.SetText("Initial value")
	e2.SetText("This entry contains a long value, much too long " +
		"to fit in the window at one time, so long in fact " +
		"that you'll have to scan or scroll to see the end.")

	pack.Pack(e1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(s1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(spacer1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(e2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(s2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(spacer2, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(e3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))
	pack.Pack(s3, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	app.Run()
}
