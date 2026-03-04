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
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
)

const boardSize = 8 // 8x8 to match Tk's knightstour.tcl.

func main() {
	app := demohelper.Setup("Knight's Tour", 450, 530, fmt.Sprintf("Knight's tour on a %dx%d board.\nClick Start to begin the animation.", boardSize, boardSize))

	statusLabel := label.New(app, "status",
		label.Text("Move: 0"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(app, "board",
		canvas.Background("white"),
		canvas.Width(400),
		canvas.Height(400),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	cellSize := 400.0 / float64(boardSize)
	margin := 10.0

	// Draw chessboard squares (tagged so we can redraw).
	drawBoard := func() {
		c.Delete("board")
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
					canvas.FillColor(fill), canvas.OutlineColor("#888888"),
					canvas.Tags("board"))
			}
		}
	}
	drawBoard()

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

	// Animation state.
	step := 0
	running := false
	delayMs := 300.0 // milliseconds between steps
	var animateFunc func()

	// Button bar.
	btnFrame := frame.New(app, "buttons")
	pack.Pack(btnFrame, pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))

	startBtn := button.New(btnFrame, "start",
		button.Text("Start"),
		button.PadX(10), button.PadY(4),
	)
	stopBtn := button.New(btnFrame, "stop",
		button.Text("Stop"),
		button.PadX(10), button.PadY(4),
	)
	resetBtn := button.New(btnFrame, "reset",
		button.Text("Reset"),
		button.PadX(10), button.PadY(4),
	)
	stopBtn.State = widget.StateDisabled
	stopBtn.Display()

	pack.Pack(startBtn, pack.SideOpt(pack.Left), pack.PadX(4))
	pack.Pack(stopBtn, pack.SideOpt(pack.Left), pack.PadX(4))
	pack.Pack(resetBtn, pack.SideOpt(pack.Left), pack.PadX(4))

	// Speed slider.
	speedScale := scale.New(app, "speed",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(50),
		scale.ToOpt(800),
		scale.ValueOpt(delayMs),
		scale.ResolutionOpt(10),
		scale.LabelOpt("Delay (ms)"),
		scale.ShowValueOpt(true),
	)
	speedScale.Command = func(v float64) {
		delayMs = v
	}
	pack.Pack(speedScale, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(10), pack.PadY(2))

	// Update button states based on running/stopped.
	updateButtons := func() {
		if running {
			startBtn.State = widget.StateDisabled
			stopBtn.State = widget.StateNormal
		} else {
			startBtn.State = widget.StateNormal
			stopBtn.State = widget.StateDisabled
		}
		startBtn.Display()
		stopBtn.Display()
	}

	animateFunc = func() {
		if !running {
			return
		}
		if step >= len(tour) {
			running = false
			updateButtons()
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
		app.After(time.Duration(delayMs)*time.Millisecond, animateFunc)
	}

	startBtn.Command = func() {
		if running {
			return
		}
		// If tour was completed or never started from 0, restart from current step.
		if step == 0 {
			c.Delete("knight")
			c.Delete("path")
		}
		running = true
		updateButtons()
		animateFunc()
	}

	stopBtn.Command = func() {
		running = false
		updateButtons()
		statusLabel.Text = fmt.Sprintf("Stopped at move %d / %d", step, len(tour))
		statusLabel.Display()
	}

	resetBtn.Command = func() {
		running = false
		step = 0
		c.Delete("knight")
		c.Delete("path")
		updateButtons()
		statusLabel.Text = "Move: 0"
		statusLabel.Display()
	}

	_ = statusLabel
	_ = startBtn
	_ = stopBtn
	_ = resetBtn
	app.Run()
}
