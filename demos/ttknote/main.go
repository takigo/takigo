// Demo: TTK Notebook with multiple tabbed pages.
// Ported from Tk's ttknote.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Ttk Notebook Widget"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttknote"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("One of the Ttk widgets is the notebook widget, which provides a set of tabs that allow the selection of a group of panels, each with distinct content. Not only can the tabs be selected with the mouse, but they can also be switched between using Ctrl+Tab when the notebook page heading itself is selected. Note that the second tab is disabled, and cannot be selected."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	ttk.SetCurrentTheme("clam")

	// Notebook.
	nb := ttk.NewNotebook(f, "nb")
	pack.Pack(nb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX("1.5p"), pack.PadY("3p"))

	// Tab 1: Description with "Neat!" button (ttk::frame, grid layout).
	page1 := ttk.NewFrame(nb, "page1")
	descLabel := label.New(page1, "desc",
		label.Text("Ttk is the new Tk themed widget set. One of the widgets "+
			"it includes is the notebook widget, which provides a set "+
			"of tabs that allow the selection of a group of panels, "+
			"each with distinct content. They are a feature of many "+
			"modern user interfaces. Not only can the tabs be selected "+
			"with the mouse, but they can also be switched between "+
			"using Ctrl+Tab when the notebook page heading itself is "+
			"selected. Note that the second tab is disabled, and "+
			"cannot be selected."),
		label.WrapLength("4i"),
	)
	grid.Grid(descLabel, grid.Row(0), grid.Column(0), grid.ColumnSpan(2),
		grid.Sticky(grid.StickN+grid.EW), grid.PadY("1.5p"))

	neatLabel := label.New(page1, "neat",
		label.Text(""),
	)

	neatBtn := ttk.NewButton(page1, "neatbtn",
		ttk.ButtonText("Neat!"),
		ttk.ButtonCommand(func() {
			neatLabel.Text = "Yeah, I know..."
			neatLabel.Display()
			app.After(500*time.Millisecond, func() {
				neatLabel.Text = ""
				neatLabel.Display()
			})
		}),
	)
	grid.Grid(neatBtn, grid.Row(1), grid.Column(0), grid.PadY("1.5p"))
	grid.Grid(neatLabel, grid.Row(1), grid.Column(1), grid.PadY("1.5p"))
	grid.RowConfigure(page1, 1, grid.Weight(1))
	grid.ColumnConfigure(page1, 0, grid.Weight(1))
	grid.ColumnConfigure(page1, 1, grid.Weight(1))

	nb.Add(page1.Window(), "Description")
	nb.SetTabUnderline(0, 0) // Alt+D → Description tab

	// Tab 2: Disabled tab (ttk::frame).
	page2 := ttk.NewFrame(nb, "page2")
	nb.Add(page2.Window(), "Disabled")
	nb.SetTabState(1, ttk.StateDisabled)

	// Tab 3: Text editor with scrollbar (ttk::frame).
	page3 := ttk.NewFrame(nb, "page3")

	tw := text.New(page3, "editor",
		text.Width(40),
		text.Height(10),
		text.WrapModeOpt(text.WrapChar),
	)

	yscroll := ttk.NewScrollbar(page3, "yscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY),
		pack.PadX("1.5p"), pack.PadY("1.5p"))
	pack.Pack(tw, pack.FillOpt(pack.FillBoth), pack.Expand(true),
		pack.PadX("1.5p"), pack.PadY("1.5p"))
	nb.Add(page3.Window(), "Text Editor")
	nb.SetTabUnderline(2, 0) // Alt+T → Text Editor tab

	_ = descLabel
	_ = neatLabel
	_ = neatBtn
	_ = yscroll
	_ = tw
	app.Run()
}
