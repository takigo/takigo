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

// justifyValues maps justification radio value → option.Justify.
var justifyValues = map[string]option.Justify{
	"left":   option.JustifyLeft,
	"center": option.JustifyCenter,
	"right":  option.JustifyRight,
}

func main() {
	app := demohelper.Setup("Listbox Demonstration (50 states)", 300, 400,
		"A listbox containing the 50 states is displayed below, along with a scrollbar. You can scan the list either using the scrollbar or by scanning. To scan, press button 2 in the widget and drag up or down.")

	// lb is declared below; the command closure captures the pointer.
	var lb *listbox.Listbox

	// "multi" is the tristatevalue: shown when selected items have mixed justifications.
	justVar := widget.NewVariable("left")

	justFrame := labelframe.New(app, "justif", labelframe.Text("Justification"))
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

	lb = listbox.New(lbFrame, "list",
		listbox.Items(states...),
		listbox.Height(12),
		listbox.SelectModeOpt(listbox.SelectExtended),
	)

	// Update justVar based on selection: "multi" if multiple items selected, else reflect current justify.
	lb.SelectCmd = func() {
		sel := lb.Selection()
		if len(sel) <= 1 {
			// Single or no selection: reflect actual listbox justify setting.
			switch lb.Justify {
			case option.JustifyCenter:
				justVar.Set("center")
			case option.JustifyRight:
				justVar.Set("right")
			default:
				justVar.Set("left")
			}
		} else {
			// Multiple items selected: show indeterminate state.
			justVar.Set("multi")
		}
	}

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
