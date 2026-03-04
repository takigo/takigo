// Demo: US states listbox with vertical scrollbar.
// Ported from Tk's states.tcl demo.
package main

import (
	"strings"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/listbox"
	"github.com/msorc/takigo/widget/radiobutton"
	"github.com/msorc/takigo/ttk"
)

func main() {
	app := demohelper.Setup("Listbox Demonstration (50 states)", 300, 400,
		"A listbox containing the 50 states is displayed below, along with a scrollbar. You can scan the list either using the scrollbar or by scanning. To scan, press button 2 in the widget and drag up or down.")

	// Justification group (matches Tcl's labelframe $w.justif).
	justVar := widget.NewVariable("left")
	justFrame := labelframe.New(app, "justif", labelframe.Text("Justification"))
	for _, c := range []string{"Left", "Center", "Right"} {
		rb := radiobutton.New(justFrame, strings.ToLower(c),
			radiobutton.Text(c),
			radiobutton.Value(strings.ToLower(c)),
			radiobutton.Var(justVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Left), pack.PadY("1.5p"), pack.FillOpt(pack.FillX))
	}
	pack.Pack(justFrame, pack.SideOpt(pack.Top))

	// Listbox frame with border.
	lbFrame := frame.New(app, "frame",
		frame.BorderWidth(19), // .5c ≈ 19px
	)
	pack.Pack(lbFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillY), pack.Expand(true))

	states := []string{
		"Alabama", "Alaska", "Arizona", "Arkansas", "California",
		"Colorado", "Connecticut", "Delaware", "Florida", "Georgia",
		"Hawaii", "Idaho", "Illinois", "Indiana", "Iowa",
		"Kansas", "Kentucky", "Louisiana", "Maine", "Maryland",
		"Massachusetts", "Michigan", "Minnesota", "Mississippi", "Missouri",
		"Montana", "Nebraska", "Nevada", "New Hampshire", "New Jersey",
		"New Mexico", "New York", "North Carolina", "North Dakota", "Ohio",
		"Oklahoma", "Oregon", "Pennsylvania", "Rhode Island", "South Carolina",
		"South Dakota", "Tennessee", "Texas", "Utah", "Vermont",
		"Virginia", "Washington", "West Virginia", "Wisconsin", "Wyoming",
	}

	lb := listbox.New(lbFrame, "list",
		listbox.Items(states...),
		listbox.Height(12),
	)

	yscroll := ttk.NewScrollbar(lbFrame, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						lb.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					lb.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	_ = justVar
	_ = justFrame
	app.Run()
}
