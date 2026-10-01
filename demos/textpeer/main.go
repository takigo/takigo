// Demo: A pair of text widgets that can edit a single logical buffer.
// Ported from Tk's textpeer.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Widget Peering Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("textpeer"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	count := 0

	// Define a widget that we peer from; it won't ever be shown.
	count++
	first := text.New(f, fmt.Sprintf("text%d", count))
	first.Insert("end", "This is a coupled pair of text widgets; they are peers to "+
		"each other. They have the same underlying data model, but "+
		"can show different locations, have different current edit "+
		"locations, and have different selections. You can also "+
		"create additional peers of any of these text widgets using "+
		"the Make Peer button beside the text widget to clone, and "+
		"delete a particular peer widget using the Delete Peer "+
		"button.")
	sharedDoc := first.Doc()

	// peerWidgets holds references for cleanup.
	type peerInfo struct {
		tw   *text.TextWidget
		sb   *ttk.Scrollbar
		make *button.Button
		kill *button.Button
	}
	peers := make(map[int]*peerInfo)

	var makeClone func()

	killClone := func(idx int) {
		if p, ok := peers[idx]; ok {
			app.Server().UnmapWindow(p.tw.Win.PlatformID)
			app.Server().UnmapWindow(p.sb.Win.PlatformID)
			app.Server().UnmapWindow(p.make.Win.PlatformID)
			app.Server().UnmapWindow(p.kill.Win.PlatformID)
			delete(peers, idx)
		}
	}

	makeClone = func() {
		count++
		idx := count

		tw := text.NewPeer(sharedDoc, f, fmt.Sprintf("text%d", idx),
			text.Height(10),
			text.WrapModeOpt(text.WrapWord),
		)

		sb := ttk.NewScrollbar(f, fmt.Sprintf("sb%d", idx),
			ttk.ScrollbarOrientOpt(ttk.Vertical),
			ttk.ScrollbarCommandOpt(widget.ScrollY(tw)),
		)
		tw.YScrollCmd = func(first, last float64) { sb.Set(first, last) }

		makeBtn := button.New(f, fmt.Sprintf("clone%d", idx),
			button.Text("Make Peer"),
			button.Command(makeClone),
		)
		killBtn := button.New(f, fmt.Sprintf("kill%d", idx),
			button.Text("Delete Peer"),
			button.Command(func() { killClone(idx) }),
		)

		peers[idx] = &peerInfo{tw: tw, sb: sb, make: makeBtn, kill: killBtn}

		row := idx * 2
		grid.Grid(tw, grid.Row(row), grid.Column(0), grid.RowSpan(2),
			grid.Sticky(grid.NSEW))
		grid.Grid(sb, grid.Row(row), grid.Column(1), grid.RowSpan(2),
			grid.Sticky(grid.NSEW))
		grid.Grid(makeBtn, grid.Row(row), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))
		grid.Grid(killBtn, grid.Row(row+1), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))
		grid.RowConfigure(f, row+1, grid.Weight(1))
	}

	// Create two initial peers.
	makeClone()
	makeClone()

	// Destroy the hidden first text widget.
	app.Server().UnmapWindow(first.Win.PlatformID)

	// See Code / Dismiss buttons.
	btns := demohelper.AddSeeDismiss(f)
	grid.Grid(btns, grid.Row(5000), grid.Column(0), grid.ColumnSpan(3),
		grid.Sticky(grid.EW))

	grid.ColumnConfigure(f, 0, grid.Weight(1))

	app.Run()
}
