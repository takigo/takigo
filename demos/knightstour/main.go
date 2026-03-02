// Demo: Knight's tour visualization on canvas.
// Ported from Tk's knightstour.tcl demo.
package main

import (
	"fmt"
	"time"

	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

const boardSize = 6 // 6x6 for faster computation.

func main() {
	d := demohelper.Setup("Knight's Tour", 450, 500, fmt.Sprintf("Knight's tour on a %dx%d board.\nClick Start to begin the animation.", boardSize, boardSize))
	root, app := d.Root, d.App

	statusLabel := label.New(root, "status", app,
		label.Text("Move: 0"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(root, "board", app,
		canvas.Background("white"),
		canvas.Width(400),
		canvas.Height(400),
	)
	pack.Pack(c.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	cellSize := 400.0 / float64(boardSize)
	margin := 10.0

	// Draw chessboard.
	for row := range boardSize {
		for col := range boardSize {
			x1 := margin + float64(col)*cellSize
			y1 := margin + float64(row)*cellSize
			x2 := x1 + cellSize
			y2 := y1 + cellSize
			fill := "#f0d9b5"
			if (row+col)%2 == 1 {
				fill = "#b58863"
			}
			c.CreateRectangle(x1, y1, x2, y2,
				canvas.FillColor(fill), canvas.OutlineColor("#888888"))
		}
	}

	// Find knight's tour using Warnsdorff's heuristic.
	knightMoves := [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}

	findTour := func() [][2]int {
		visited := make([][]bool, boardSize)
		for i := range visited {
			visited[i] = make([]bool, boardSize)
		}

		degree := func(r, c int) int {
			count := 0
			for _, m := range knightMoves {
				nr, nc := r+m[0], c+m[1]
				if nr >= 0 && nr < boardSize && nc >= 0 && nc < boardSize && !visited[nr][nc] {
					count++
				}
			}
			return count
		}

		path := make([][2]int, 0, boardSize*boardSize)
		r, cc := 0, 0
		visited[r][cc] = true
		path = append(path, [2]int{r, cc})

		for len(path) < boardSize*boardSize {
			bestDeg := 9
			bestR, bestC := -1, -1
			for _, m := range knightMoves {
				nr, nc := r+m[0], cc+m[1]
				if nr >= 0 && nr < boardSize && nc >= 0 && nc < boardSize && !visited[nr][nc] {
					d := degree(nr, nc)
					if d < bestDeg {
						bestDeg = d
						bestR, bestC = nr, nc
					}
				}
			}
			if bestR < 0 {
				break
			}
			r, cc = bestR, bestC
			visited[r][cc] = true
			path = append(path, [2]int{r, cc})
		}
		return path
	}

	tour := findTour()

	// Start button.
	step := 0
	var animateFunc func()

	startBtn := button.New(root, "start", app,
		button.Text("Start"),
		button.PadX(10), button.PadY(4),
	)

	startBtn.Command = func() {
		step = 0
		c.Delete("knight")
		c.Delete("path")
		animateFunc()
	}

	animateFunc = func() {
		if step >= len(tour) {
			statusLabel.Text = fmt.Sprintf("Tour complete! %d moves.", len(tour))
			statusLabel.Display()
			return
		}

		pos := tour[step]
		cx := margin + (float64(pos[1])+0.5)*cellSize
		cy := margin + (float64(pos[0])+0.5)*cellSize

		// Draw path line from previous position.
		if step > 0 {
			prev := tour[step-1]
			px := margin + (float64(prev[1])+0.5)*cellSize
			py := margin + (float64(prev[0])+0.5)*cellSize
			c.CreateLine([]float64{px, py, cx, cy},
				canvas.OutlineColor("#3498db"), canvas.OutlineWidth(2), canvas.Tags("path"))
		}

		// Draw knight marker.
		c.Delete("knight")
		r := cellSize * 0.3
		c.CreateOval(cx-r, cy-r, cx+r, cy+r,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"), canvas.OutlineWidth(2),
			canvas.Tags("knight"))
		c.CreateText(cx, cy,
			canvas.TextOpt(fmt.Sprintf("%d", step+1)),
			canvas.FontOpt("Sans Bold 9"), canvas.TextColor("white"),
			canvas.AnchorOpt(option.AnchorCenter), canvas.Tags("knight"))

		statusLabel.Text = fmt.Sprintf("Move: %d / %d", step+1, len(tour))
		statusLabel.Display()

		step++
		app.After(300*time.Millisecond, animateFunc)
	}

	pack.Pack(startBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	_ = statusLabel
	_ = startBtn
	d.Run()
}
