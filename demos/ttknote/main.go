// Demo: TTK Notebook with multiple tabbed pages.
// Ported from Tk's ttknote.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Ttk Notebook Widget", 500, 350, "One of the Ttk widgets is the notebook widget, which provides a set of tabs that allow the selection of a group of panels, each with distinct content. Not only can the tabs be selected with the mouse, but they can also be switched between using Ctrl+Tab when the notebook page heading itself is selected. Note that the second tab is disabled, and cannot be selected.")

	ttk.SetCurrentTheme("clam")

	// Notebook.
	nb := ttk.NewNotebook(app, "nb")
	pack.Pack(nb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(15), pack.PadY(10))

	// Tab 1: Description with "Neat!" button.
	page1 := frame.New(nb, "page1")
	descLabel := label.New(page1, "desc",
		label.Text("Ttk is the new Tk themed widget set. One of the widgets\nit includes is the notebook widget, which provides a set\nof tabs that allow the selection of a group of panels,\neach with distinct content. They are a feature of many\nmodern user interfaces. Not only can the tabs be selected\nwith the mouse, but they can also be switched between\nusing Ctrl+Tab when the notebook page heading itself is\nselected. Note that the second tab is disabled, and\ncannot be selected."),
		label.Anchor(option.AnchorNW),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(descLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	neatLabel := label.New(page1, "neat",
		label.Text(""),
		label.Anchor(option.AnchorW),
		label.PadX(10),
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
	pack.Pack(neatBtn, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))
	pack.Pack(neatLabel, pack.SideOpt(pack.Left), pack.PadX(10), pack.PadY(5))
	nb.Add(page1.Window(), "Description")

	// Tab 2: Disabled tab.
	page2 := frame.New(nb, "page2")
	nb.Add(page2.Window(), "Disabled")
	nb.SetTabState(1, ttk.StateDisabled)

	// Tab 3: Text editor with scrollbar.
	page3 := frame.New(nb, "page3")

	tw := text.New(page3, "editor",
		text.Width(40),
		text.Height(10),
		text.WrapModeOpt(text.WrapChar),
	)

	yscroll := scrollbar.New(page3, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
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

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	nb.Add(page3.Window(), "Text Editor")

	app.Run()
}
