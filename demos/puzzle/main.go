// Demo: 15-puzzle game using buttons and place geometry.
// Ported from Tk's puzzle.tcl demo.
package main

import (
	"fmt"
	"math"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/window"
)

func main() {
	d := demohelper.Setup("15-Puzzle", 340, 400,
		"A 15-puzzle. Click on a piece next to the\nempty space to slide it into the space.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	// Puzzle frame.
	puzzleFrame := frame.New(root, "puzzle", app,
		frame.Width(240),
		frame.Height(240),
		frame.BorderWidth(2),
		frame.Relief(option.ReliefSunken),
		frame.Background("#4a6984"),
	)
	pack.Pack(puzzleFrame.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10))

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
			btn := button.New(puzzleFrame.Window(), fmt.Sprintf("p%d", num), app,
				button.Text(fmt.Sprintf("%d", num)),
				button.Command(func() { tryMove(n) }),
				button.PadX(2),
				button.PadY(2),
			)
			placePiece(btn.Window(), col, row)
			buttons = append(buttons, btn)
			btnMap[num] = btn
		}
	}

	_ = buttons
	d.Run()
}
