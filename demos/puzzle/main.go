// Demo: 15-puzzle game using a collection of buttons.
// Ported from Tk's puzzle.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("15-Puzzle Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("15-Puzzle"),
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
		label.Text("A 15-puzzle appears below as a collection of buttons.  Click "+
			"on any of the pieces next to the space, and that piece will "+
			"slide over the space.  Continue this until the pieces are "+
			"arranged in numerical order from upper-left to lower-right."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Puzzle frame (matches Tcl: 90p ≈ 120px, pady 1c padx 1c).
	puzzleFrame := frame.New(f, "frame",
		frame.Width(240),
		frame.Height(240),
		frame.BorderWidth(2),
		frame.Relief(option.ReliefSunken),
		frame.Background("#4a6984"),
	)
	pack.Pack(puzzleFrame, pack.SideOpt(pack.Top), pack.PadX("1c"), pack.PadY("1c"))

	// Initial tile order (scrambled).
	order := []int{3, 1, 6, 2, 5, 7, 15, 13, 4, 11, 8, 9, 14, 10, 12}

	// Track positions: piecePos[pieceNum] = {col, row}, spaceCol/Row for empty.
	type pos struct{ col, row int }
	piecePos := make(map[int]pos)
	var buttons []*button.Button
	btnMap := make(map[int]*button.Button)

	spaceCol, spaceRow := 3, 3 // Empty space starts at bottom-right.

	// Place a piece at grid position.
	placePiece := func(btn *window.Window, col, row int) {
		place.Place(btn,
			place.RelX(float64(col)*0.25),
			place.RelY(float64(row)*0.25),
			place.RelWidth(0.25),
			place.RelHeight(0.25),
		)
	}

	// Try to move a piece into the empty space.
	tryMove := func(num int) {
		p := piecePos[num]
		dx := math.Abs(float64(p.col - spaceCol))
		dy := math.Abs(float64(p.row - spaceRow))

		// Must be adjacent (not diagonal).
		if (dx == 1 && dy == 0) || (dx == 0 && dy == 1) {
			// Swap piece and space.
			oldSpace := pos{spaceCol, spaceRow}
			spaceCol, spaceRow = p.col, p.row
			piecePos[num] = oldSpace
			placePiece(btnMap[num].Window(), oldSpace.col, oldSpace.row)
		}
	}

	// Create the 15 pieces.
	idx := 0
	for row := range 4 {
		for col := range 4 {
			if row == 3 && col == 3 {
				break // Skip last position (empty space).
			}
			num := order[idx]
			idx++

			piecePos[num] = pos{col, row}
			n := num // capture
			btn := button.New(puzzleFrame, fmt.Sprintf("%d", num),
				button.Text(fmt.Sprintf("%d", num)),
				button.BorderWidth(0),
				button.Command(func() { tryMove(n) }),
			)
			btn.HighlightWidth = 0
			placePiece(btn.Window(), col, row)
			buttons = append(buttons, btn)
			btnMap[num] = btn
		}
	}

	_ = buttons
	app.Run()
}
