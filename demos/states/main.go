// Demo: US states listbox with vertical scrollbar.
// Ported from Tk's states.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/labelframe"
	"github.com/takigo/takigo/widget/listbox"
	"github.com/takigo/takigo/widget/radiobutton"
)

// justifyValues maps justification radio value -> option.Justify.
var justifyValues = map[string]option.Justify{
	"left":   option.JustifyLeft,
	"center": option.JustifyCenter,
	"right":  option.JustifyRight,
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Listbox Demonstration (50 states)"),
		takigo.Geometry("+300+300"),
		takigo.IconName("states"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("A listbox containing the 50 states is displayed below, along with a scrollbar.  You can scan the list either using the scrollbar or by scanning.  To scan, press button 2 in the widget and drag up or down."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	// lb is declared below; the command closure captures the pointer.
	var lb *listbox.Listbox

	// "multi" is the tristatevalue: shown when selected items have mixed justifications.
	justVar := widget.NewUnsetVariable[string]()

	justFrame := labelframe.New(f, "justif", labelframe.Text("Justification"))
	for _, c := range []string{"Left", "Center", "Right"} {
		val := strings.ToLower(c)
		j := justifyValues[val]
		rb := radiobutton.New(justFrame, val,
			radiobutton.Text(c),
			radiobutton.Value(val),
			radiobutton.Var(justVar),
			radiobutton.TristateValueOpt("multi"),
			radiobutton.Anchor(option.AnchorW),
			radiobutton.Command(func() {
				if lb != nil && justVar.Get() != "multi" {
					lb.SetJustify(j)
				}
			}),
		)
		pack.Pack(rb, pack.SideOpt(pack.Left), pack.PadY(screenunit.Pt(1.5)), pack.FillOpt(pack.FillX))
	}
	pack.Pack(justFrame, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Listbox frame with border.
	lbFrame := frame.New(f, "frame",
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

	lb = listbox.New(lbFrame, "list",
		listbox.Items(states...),
		listbox.Height(12),
	)

	yscroll := ttk.NewScrollbar(lbFrame, "scroll",
		ttk.ScrollbarOrient(ttk.Vertical),
		ttk.ScrollbarCommand(widget.ScrollY(lb)),
	)
	lb.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := lb.YVisibleRange()
	yscroll.Set(first, last)

	app.Run()
}
