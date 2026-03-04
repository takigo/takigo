// Demo: Two text widgets sharing the same logical document.
// Ported from Tk's textpeer.tcl demo (peering not implemented; static layout).
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Widget Peering Demonstration", 700, 500,
		"This window demonstrates two text widgets that would be peers in Tk. "+
			"They have the same underlying data model, but can show different locations, "+
			"have different current edit locations, and have different selections.")

	// Inner frame for grid layout (demohelper uses pack in app.Window()).
	w := frame.New(app, "w")
	pack.Pack(w, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	content := "This is a coupled pair of text widgets; they are peers to " +
		"each other. They have the same underlying data model, but " +
		"can show different locations, have different current edit " +
		"locations, and have different selections. You can also " +
		"create additional peers of any of these text widgets using " +
		"the Make Peer button beside the text widget to clone, and " +
		"delete a particular peer widget using the Delete Peer button."

	makeRow := func(idx int) {
		row := idx * 2

		tw := text.New(w, "text"+fmt.Sprint(idx),
			text.Height(10),
			text.WrapModeOpt(text.WrapWord),
		)
		sb := scrollbar.New(w, "sb"+fmt.Sprint(idx),
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
		tw.YScrollCmd = func(first, last float64) { sb.Set(first, last) }

		makeBtn := button.New(w, "clone"+fmt.Sprint(idx),
			button.Text("Make Peer"),
		)
		deleteBtn := button.New(w, "kill"+fmt.Sprint(idx),
			button.Text("Delete Peer"),
		)

		grid.Grid(tw, grid.Row(row), grid.Column(0), grid.RowSpan(2),
			grid.Sticky(grid.NSEW))
		grid.Grid(sb, grid.Row(row), grid.Column(1), grid.RowSpan(2),
			grid.Sticky(grid.NSEW))
		grid.Grid(makeBtn, grid.Row(row), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))
		grid.Grid(deleteBtn, grid.Row(row+1), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))

		tw.Insert("1.0", content)
	}

	makeRow(1)
	makeRow(2)

	grid.ColumnConfigure(w.Window(), 0, grid.SlotConfig{Weight: 1})

	app.Run()
}
