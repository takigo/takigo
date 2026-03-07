// Demo: Knight's tour visualization on canvas.
// Ported from Tk's knightstour.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scale"
)

const boardSize = 8

var knightMoves = [][2]int{{-2, -1}, {-2, 1}, {-1, -2}, {-1, 2}, {1, -2}, {1, 2}, {2, -1}, {2, 1}}

func degree(visited [][]bool, r, c int) int {
	count := 0
	for _, m := range knightMoves {
		nr, nc := r+m[0], c+m[1]
		if nr >= 0 && nr < boardSize && nc >= 0 && nc < boardSize && !visited[nr][nc] {
			count++
		}
	}
	return count
}

// edgeDistance returns the minimum distance of (r,c) from any board edge.
// Lower = closer to edge. Used as secondary tiebreaker (Edgemost heuristic).
func edgeDistance(r, c int) int {
	d := r
	if boardSize-1-r < d {
		d = boardSize - 1 - r
	}
	if c < d {
		d = c
	}
	if boardSize-1-c < d {
		d = boardSize - 1 - c
	}
	return d
}

// findTour finds a knight's tour starting at (startR, startC) using
// Warnsdorff's heuristic with Edgemost tiebreaking.
func findTour(startR, startC int) [][2]int {
	visited := make([][]bool, boardSize)
	for i := range visited {
		visited[i] = make([]bool, boardSize)
	}

	path := make([][2]int, 0, boardSize*boardSize)
	r, cc := startR, startC
	visited[r][cc] = true
	path = append(path, [2]int{r, cc})

	for len(path) < boardSize*boardSize {
		bestDeg := 9
		bestEdge := boardSize
		bestR, bestC := -1, -1
		for _, m := range knightMoves {
			nr, nc := r+m[0], cc+m[1]
			if nr >= 0 && nr < boardSize && nc >= 0 && nc < boardSize && !visited[nr][nc] {
				d := degree(visited, nr, nc)
				e := edgeDistance(nr, nc)
				if d < bestDeg || (d == bestDeg && e < bestEdge) {
					bestDeg = d
					bestEdge = e
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

func main() {
	app, err := takigo.NewApp(takigo.Title("Knight's Tour"),
		takigo.Geometry("+300+300"),
		takigo.IconName("knightstour"),
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
		label.Text(fmt.Sprintf("Knight's tour on a %dx%d board. Click a square to set the starting position, then click Start.", boardSize, boardSize)),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	statusLabel := label.New(f, "status",
		label.Text("Click a square to set start, then Start."),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	c := canvas.New(f, "board",
		canvas.Background("white"),
		canvas.Width(400),
		canvas.Height(400),
	)
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	const cellSize = 400.0 / float64(boardSize)
	const margin = 0.0

	// Draw chessboard.
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

	// State.
	startRow, startCol := 0, 0
	tour := findTour(startRow, startCol)
	step := 0
	running := false
	delayMs := 300.0
	var animateFunc func()

	// Highlight starting square.
	highlightStart := func() {
		c.Delete("startmark")
		x := margin + (float64(startCol)+0.15)*cellSize
		y := margin + (float64(startRow)+0.15)*cellSize
		r := cellSize * 0.7
		c.CreateOval(x, y, x+r, y+r,
			canvas.FillColor("#2ecc71"), canvas.OutlineColor("black"),
			canvas.OutlineWidth(2), canvas.Tags("startmark"))
	}
	highlightStart()

	// Button bar.
	btnFrame := frame.New(f, "buttons")
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

	// Repeat checkbutton.
	repeatVar := widget.NewVariable(false)
	repeatChk := checkbutton.New(btnFrame, "repeat",
		checkbutton.Text("Repeat"),
		checkbutton.Var(repeatVar),
	)
	pack.Pack(repeatChk, pack.SideOpt(pack.Left), pack.PadX(4))

	// Speed slider.
	speedScale := scale.New(f, "speed",
		scale.OrientOpt(scale.Horizontal),
		scale.FromOpt(50),
		scale.ToOpt(800),
		scale.ValueOpt(delayMs),
		scale.ResolutionOpt(10),
		scale.LabelOpt("Delay (ms)"),
		scale.ShowValueOpt(true),
	)
	speedScale.Command = func(v float64) { delayMs = v }
	pack.Pack(speedScale, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(10), pack.PadY(2))

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
			if repeatVar.Get() {
				// Restart from same square.
				step = 0
				c.Delete("knight")
				c.Delete("path")
				tour = findTour(startRow, startCol)
				running = true
				updateButtons()
				animateFunc()
			} else {
				statusLabel.Text = fmt.Sprintf("Tour complete! %d moves.", len(tour))
				statusLabel.Display()
			}
			return
		}

		pos := tour[step]
		cx := margin + (float64(pos[1])+0.5)*cellSize
		cy := margin + (float64(pos[0])+0.5)*cellSize

		if step > 0 {
			prev := tour[step-1]
			px := margin + (float64(prev[1])+0.5)*cellSize
			py := margin + (float64(prev[0])+0.5)*cellSize
			c.CreateLine([]float64{px, py, cx, cy},
				canvas.OutlineColor("#3498db"), canvas.OutlineWidth(2),
				canvas.Tags("path"))
		}

		c.Delete("knight")
		r := cellSize * 0.3
		c.CreateOval(cx-r, cy-r, cx+r, cy+r,
			canvas.FillColor("#e74c3c"), canvas.OutlineColor("black"),
			canvas.OutlineWidth(2), canvas.Tags("knight"))
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
		c.Delete("knight")
		c.Delete("path")
		c.Delete("startmark")
		step = 0
		tour = findTour(startRow, startCol)
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
		highlightStart()
		updateButtons()
		statusLabel.Text = "Click a square to set start, then Start."
		statusLabel.Display()
	}

	// Click on canvas to set start square (only when not running).
	app.Dispatcher().Bind(c.Win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 || running {
			return
		}
		col := int(float64(ev.X) / cellSize)
		row := int(float64(ev.Y) / cellSize)
		if row < 0 || row >= boardSize || col < 0 || col >= boardSize {
			return
		}
		startRow = row
		startCol = col
		step = 0
		tour = findTour(startRow, startCol)
		c.Delete("knight")
		c.Delete("path")
		highlightStart()
		statusLabel.Text = fmt.Sprintf("Start: %c%d — click Start to run.",
			rune('a'+col), boardSize-row)
		statusLabel.Display()
	})

	app.Run()
}
