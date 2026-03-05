// Demo: Two text widgets sharing the same logical document (true peering).
// Ported from Tk's textpeer.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Widget Peering Demonstration", 700, 500,
		"A demonstration of the text peer facility. The two text widgets "+
			"below are peers of each other; they display and edit the same "+
			"underlying document. Note that editing in one peer immediately "+
			"updates the other. Each peer can show a different part of the "+
			"document and has its own insert cursor and selection.")

	// Inner frame for grid layout.
	w := frame.New(app, "w")
	pack.Pack(w, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Initial content.
	content := "This is a coupled pair of text widgets; they are peers to " +
		"each other. They share the same underlying data model, so " +
		"editing in one immediately updates the other. Each peer can " +
		"show a different location in the document, and each has its " +
		"own insert cursor and selection.\n\nTry editing in either widget!"

	// peerWidgets holds all current peer TextWidgets.
	var peerWidgets []*text.TextWidget
	peerCount := 0

	// doc will be set after creating the first widget.
	var sharedDoc *text.Document

	// removePeer removes the peer at peerIdx from the frame and list.
	var removePeer func(tw *text.TextWidget)

	// addPeer adds a new peer row using the shared document.
	var addPeer func()

	addPeer = func() {
		peerCount++
		idx := peerCount

		var tw *text.TextWidget
		if sharedDoc == nil {
			// Primary widget — owns the document.
			tw = text.New(w, fmt.Sprintf("text%d", idx),
				text.Height(10),
				text.WrapModeOpt(text.WrapWord),
			)
			sharedDoc = tw.Doc()
			tw.Insert("1.0", content)
		} else {
			// Peer widget — shares the document.
			tw = text.NewPeer(sharedDoc, w, fmt.Sprintf("text%d", idx),
				text.Height(10),
				text.WrapModeOpt(text.WrapWord),
			)
		}
		peerWidgets = append(peerWidgets, tw)

		row := (idx - 1) * 2

		sb := ttk.NewScrollbar(w, fmt.Sprintf("sb%d", idx),
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
		tw.YScrollCmd = func(first, last float64) { sb.Set(first, last) }

		makeBtn := button.New(w, fmt.Sprintf("clone%d", idx),
			button.Text("Make Peer"),
			button.Command(addPeer),
		)
		deleteBtn := button.New(w, fmt.Sprintf("kill%d", idx),
			button.Text("Delete Peer"),
			button.Command(func() { removePeer(tw) }),
		)

		grid.Grid(tw, grid.Row(row), grid.Column(0), grid.RowSpan(2),
			grid.Sticky(grid.NSEW))
		grid.Grid(sb, grid.Row(row), grid.Column(1), grid.RowSpan(2),
			grid.Sticky(grid.NS))
		grid.Grid(makeBtn, grid.Row(row), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))
		grid.Grid(deleteBtn, grid.Row(row+1), grid.Column(2),
			grid.Sticky(grid.StickN+grid.EW))
	}

	removePeer = func(tw *text.TextWidget) {
		// Don't remove if only one peer left.
		if len(peerWidgets) <= 1 {
			return
		}
		// Remove from list.
		for i, p := range peerWidgets {
			if p == tw {
				peerWidgets = append(peerWidgets[:i], peerWidgets[i+1:]...)
				break
			}
		}
		// Destroy the widget's window (unmap + free resources).
		app.Server().UnmapWindow(tw.Win.PlatformID)
	}

	// Create two initial peers.
	addPeer()
	addPeer()

	grid.ColumnConfigure(w.Window(), 0, grid.SlotConfig{Weight: 1})

	app.Run()
}
