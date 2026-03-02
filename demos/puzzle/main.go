// Demo: 15-puzzle game using buttons and place geometry.
// Ported from Tk's puzzle.tcl demo.
package main

import (
	"fmt"
	"math"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/window"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("15-Puzzle"), takigo.Size(340, 400))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("A 15-puzzle. Click on a piece next to the\nempty space to slide it into the space."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button at bottom.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

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

	// Root event handlers.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = buttons
	app.MainLoop()
}
