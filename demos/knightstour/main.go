// Demo: Calculate a Knight's tour of a chessboard.
// Ported from Tk's knightstour.tcl demo (Warnsdorff's rule with the
// Edgemost tie-break).
package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"time"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/text"
)

// tclRand ports Tcl's rand() (tclBasic.c ExprRandFunc), a Park-Miller
// generator, so a seeded run picks the same squares as the Tcl demo.
type tclRand struct{ seed int64 }

func newTclRand(seed int64) *tclRand {
	s := seed & 0x7fffffff
	if s == 0 || s == 0x7fffffff {
		s ^= 123459876
	}
	return &tclRand{s}
}

func (r *tclRand) float() float64 {
	const ia, im, iq, ir = 16807, 2147483647, 127773, 2836
	tmp := r.seed / iq
	r.seed = ia*(r.seed-tmp*iq) - ir*tmp
	if r.seed < 0 {
		r.seed += im
	}
	return float64(r.seed) / im
}

func validMoves(square int) []int {
	var moves []int
	for _, p := range [][2]int{{-1, -2}, {-2, -1}, {-2, 1}, {-1, 2}, {1, 2}, {2, 1}, {2, -1}, {1, -2}} {
		col, row := square%8+p[0], square/8+p[1]
		if row >= 0 && row < 8 && col >= 0 && col < 8 {
			moves = append(moves, row*8+col)
		}
	}
	return moves
}

func edgemost(a, b int) int {
	f := func(v int) int { return 3 - int(math.Abs(3.5-float64(v))) }
	if f(a%8)*f(a/8) < f(b%8)*f(b/8) {
		return a
	}
	return b
}

func squareName(sq int) string { return fmt.Sprintf("%c%d", 'a'+sq%8, sq/8+1) }

func main() {
	app, err := takigo.NewApp(takigo.Title("Knight's Tour"),
		takigo.Geometry("+300+300"),
		takigo.IconName("knightstour"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// The screenshot scripts run the Tcl side with srand(1).
	rnd := newTclRand(time.Now().UnixNano())
	if os.Getenv("TAKIGO_FREEZE_TIMERS") == "1" {
		// Like the wrapper's "expr {srand(1)}", which returns (and so
		// consumes) the first number.
		rnd = newTclRand(1)
		rnd.float()
	}

	dlg := frame.New(app, "f")
	pack.Pack(dlg, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	f := ttk.NewFrame(dlg, "f")
	pt := screenunit.Distance.Float
	c := canvas.New(f, "c", canvas.Width(screenunit.Pt(192)), canvas.Height(screenunit.Pt(192)))
	txt := text.New(f, "txt", text.Width(12), text.Height(1),
		text.PadXOpt(screenunit.Pt(3).Pixels()), text.FontOpt(font.TkFixedFont))
	vs := ttk.NewScrollbar(f, "vs", ttk.ScrollbarOrient(ttk.Vertical),
		ttk.ScrollbarCommand(widget.ScrollY(txt)))
	txt.YScrollCmd = vs.Set

	speed := widget.NewVariable(1400.0)
	delay := 2000 - 1400
	continuous := widget.NewVariable(false)

	tf := ttk.NewFrame(dlg, "tf")
	ls := ttk.NewLabel(tf, "ls", ttk.LabelText("Speed"))
	sc := ttk.NewScale(tf, "sc", ttk.ScaleFrom(0), ttk.ScaleTo(1992),
		ttk.ScaleVariable(speed),
		ttk.ScaleCommand(func(v float64) { delay = 2000 - int(v) }))
	cc := ttk.NewCheckbutton(tf, "cc", ttk.CheckbuttonText("Repeat"),
		ttk.CheckbuttonVar(continuous))

	var squares [64]canvas.ItemID
	var visited []int
	var initial int
	var b1 *ttk.Button
	var tour func(square int)
	gen := 0 // cancels pending moves, like "after cancel $aid"

	sq := 0
	for row := 7; row >= 0; row-- {
		for col := range 8 {
			fill, dfill := "bisque", "bisque3"
			if (col&1)^(row&1) != 0 {
				fill, dfill = "tan3", "tan4"
			}
			squares[sq] = c.CreateRectangle(
				pt(screenunit.Pt(float64(col*24+3))), pt(screenunit.Pt(float64(row*24+3))),
				pt(screenunit.Pt(float64(col*24+24))), pt(screenunit.Pt(float64(row*24+24))),
				canvas.FillColor(fill), canvas.DisabledFill(dfill),
				canvas.OutlineWidth(screenunit.Pt(1.5).Pixels()),
				canvas.StateOpt(canvas.ItemStateDisabled), canvas.OutlineColor("black"))
			sq++
		}
	}
	// On X11 the demo draws the knight as a polygon.
	knight := c.CreatePolygon([]float64{
		2, 25, 24, 25, 21, 19, 20, 8, 14, 0, 10, 0, 0, 13, 0, 16,
		2, 17, 4, 14, 5, 15, 3, 17, 5, 17, 9, 14, 10, 15, 5, 21,
	}, canvas.Tags("knight"), canvas.FillColor("black"), canvas.ActiveFill("#600000"))
	_ = knight
	scaleFactor := float64(screenunit.ScalingPct()) / 100
	c.Scale("knight", 0, 0, scaleFactor, scaleFactor)
	moveKnightTo := func(square int) {
		xy := c.ItemCoords(fmt.Sprint(squares[square]))
		c.MoveTo("knight", xy[0], xy[1])
	}
	moveKnightTo(int(rnd.float() * 64))

	var dragging []int
	c.BindItem("knight", event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		c.DeleteTag("selected", "all")
		c.AddTag("selected", "current")
		dragging = []int{ev.X, ev.Y}
	})
	c.BindItem("knight", event.MotionMask, func(ev *event.Event) {
		if dragging != nil {
			c.Move("selected", float64(ev.X-dragging[0]), float64(ev.Y-dragging[1]))
			dragging = []int{ev.X, ev.Y}
		}
	})
	c.BindItem("knight", event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		id := c.FindClosest(float64(ev.X), float64(ev.Y), 0, fmt.Sprint(knight))
		xy := c.ItemCoords(fmt.Sprint(id))
		c.MoveTo("selected", xy[0], xy[1])
		c.DeleteTag("selected", "all")
		dragging = nil
	})

	visitedHas := func(s int) bool {
		return slices.Contains(visited, s)
	}
	checkSquare := func(s int) int {
		n := 0
		for _, t := range validMoves(s) {
			if !visitedHas(t) {
				n++
			}
		}
		return n
	}
	next := func(s int) int {
		minimum, nextSq := 9, -1
		for _, t := range validMoves(s) {
			if visitedHas(t) {
				continue
			}
			if n := checkSquare(t); n < minimum {
				minimum, nextSq = n, t
			} else if n == minimum {
				nextSq = edgemost(nextSq, t)
			}
		}
		return nextSq
	}
	setOutline := func(s int, color string) {
		c.ItemConfigure(fmt.Sprint(squares[s]), canvas.StateOpt(canvas.ItemStateNormal), canvas.OutlineColor(color))
	}
	var movePiece func(g, last, square int)
	movePiece = func(g, last, square int) {
		if g != gen {
			return
		}
		txt.Insert("end", fmt.Sprintf("%2d. %s .. %s\n", len(visited), squareName(last), squareName(square)))
		txt.See("end")
		setOutline(last, "black")
		setOutline(square, "red")
		moveKnightTo(square)
		visited = append(visited, square)
		if n := next(square); n != -1 {
			app.After(time.Duration(delay)*time.Millisecond, func() { movePiece(g, square, n) })
			return
		}
		b1.ChangeState(0, ttk.StateDisabled)
		switch {
		case len(visited) != 64:
			txt.Insert("end", "FAILED!")
		case initial == square:
			txt.Insert("end", "Closed tour!")
		default:
			txt.Insert("end", "Success")
			if continuous.Get() {
				app.After(time.Duration(delay*2)*time.Millisecond, func() { tour(int(rnd.float() * 64)) })
			}
		}
	}
	tour = func(square int) {
		visited = nil
		txt.Delete("1.0", "end")
		b1.ChangeState(ttk.StateDisabled, 0)
		for _, id := range squares {
			c.ItemConfigure(fmt.Sprint(id), canvas.StateOpt(canvas.ItemStateDisabled), canvas.OutlineColor("black"))
		}
		if square < 0 {
			xy := c.ItemCoords("knight")
			id := c.FindClosest(xy[0], xy[1], 0, fmt.Sprint(knight))
			for i, s := range squares {
				if s == id {
					square = i
				}
			}
		}
		initial = square
		gen++
		g := gen
		app.DoWhenIdle(func() { movePiece(g, square, square) })
	}

	b1 = ttk.NewButton(tf, "b1", ttk.ButtonText("Start"), ttk.ButtonCommand(func() { tour(-1) }))
	// Exit (b2) exists but is only packed outside the widget demo.
	ttk.NewButton(tf, "b2", ttk.ButtonText("Exit"), ttk.ButtonCommand(app.Quit))

	grid.Grid(c, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(txt, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(vs, grid.Row(0), grid.Column(2), grid.Sticky(grid.NSEW))
	grid.RowConfigure(f, 0, grid.Weight(1))
	grid.ColumnConfigure(f, 1, grid.Weight(1))
	grid.Grid(f, grid.Row(0), grid.Column(0), grid.ColumnSpan(6), grid.Sticky(grid.NSEW))

	right := []pack.PackOption{pack.SideOpt(pack.Right), pack.PadX(screenunit.Pt(1.5)), pack.PadY(screenunit.Pt(1.5))}
	pack.Pack(b1, right...)
	pack.Pack(cc, right...)
	pack.Pack(sc, right...)
	pack.Pack(ls, right...)
	grid.Grid(tf, grid.Row(1), grid.Column(0), grid.ColumnSpan(6), grid.Sticky(grid.EW))
	btns := demohelper.AddSeeDismiss(dlg)
	grid.Grid(btns, grid.Row(2), grid.Column(0), grid.ColumnSpan(6), grid.Sticky(grid.EW))
	grid.RowConfigure(dlg, 0, grid.Weight(1))
	grid.ColumnConfigure(dlg, 0, grid.Weight(1))

	app.Run()
}
