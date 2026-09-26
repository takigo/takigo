// Demo: Rube Goldberg machine — complex canvas animation.
// Ported from Tk's goldberg.tcl demo by Keith Vetter.
package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/canvas"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/window"
)

// Animation modes.
const (
	mStart = 0
	mGo    = 1
	mPause = 2
	mSStep = 3
	mBStep = 4
	mDone  = 5
)

// Colors keyed by name matching the Tcl C(...) array.
var colors = map[string]string{
	"fg":  "black",
	"bg":  "cornflowerblue",
	"0":   "white",
	"1a":  "darkgreen",
	"1b":  "yellow",
	"2":   "red",
	"3a":  "green",
	"3b":  "darkblue",
	"4":   "black",
	"5a":  "brown",
	"5b":  "white",
	"6":   "magenta",
	"7":   "green",
	"8":   "black",
	"9":   "blue4",
	"10a": "white",
	"10b": "cyan",
	"11a": "yellow",
	"11b": "mediumblue",
	"12":  "tan2",
	"13a": "yellow",
	"13b": "red",
	"14":  "white",
	"15a": "green",
	"15b": "yellow",
	"16":  "gray65",
	"17":  "#A65353",
	"18":  "black",
	"19":  "gray50",
	"20":  "cyan",
	"21":  "gray65",
	"22":  "cyan",
	"23a": "blue",
	"23b": "red",
	"23c": "yellow",
	"24a": "red",
	"24b": "white",
	"24c": "black",
	"26":  "white",
}

// Delays per speed setting (speed 1..10 → delay in ms).
var delays = map[int]int{
	1: 500, 2: 400, 3: 300, 4: 200, 5: 150,
	6: 100, 7: 80, 8: 50, 9: 20, 10: 10,
}

// XY6 gumball positions keyed by string (matching Tcl's array keys).
var xy6 = map[string][2]float64{
	"-1": {366, 207}, "-2": {349, 204}, "-3": {359, 193}, "-4": {375, 192},
	"-5": {340, 190}, "-6": {349, 177}, "-7": {366, 177}, "-8": {380, 176},
	"-9": {332, 172}, "-10": {342, 161}, "-11": {357, 164}, "-12": {372, 163},
	"-13": {381, 149}, "-14": {364, 151}, "-15": {349, 146}, "-16": {333, 148},
	"0": {357, 219},

	"1": {359, 261}, "2": {359, 291}, "3": {359, 318},
	"4": {361, 324}, "5": {365, 329}, "6": {367, 334},
	"7": {367, 340}, "8": {366, 346}, "9": {364, 350},
	"10": {361, 355}, "11": {359, 370}, "12": {359, 391},

	"13,0": {360, 456}, "13,1": {376, 456}, "13,2": {346, 456}, "13,3": {330, 456},
	"13,4": {353, 444}, "13,5": {368, 443}, "13,6": {339, 442}, "13,7": {359, 431},
	"13,8": {380, 437}, "13,9": {345, 428}, "13,10": {328, 434}, "13,11": {373, 424},
	"13,12": {331, 420}, "13,13": {360, 417}, "13,14": {345, 412}, "13,15": {376, 410},
	"13,16": {360, 403},
}

// pos holds an animation position with optional rotation and trigger flag.
type pos struct {
	x, y float64
	beta float64 // rotation angle in degrees, 0 if none
	flag string  // "x" or "y" trigger flags, "" if none
}

// goldberg holds all state for the Rube Goldberg animation.
type goldberg struct {
	app *takigo.App
	c   *canvas.Canvas
	sf  float64 // scaleFactor (0.75 at standard DPI)

	mode    int
	speed   int
	cnt     int
	message string
	active  []int
	step    map[int]int // per-animation step counters (-1 means unset)
	pause   bool

	move25Start time.Time
}

// scl scales a flat list of coordinates by the scale factor.
// Only x,y pairs (indices 0,1 then 2,3 etc.) are scaled.
func (g *goldberg) scl(coords []float64) []float64 {
	out := make([]float64, len(coords))
	for i, v := range coords {
		out[i] = math.Round(v * g.sf)
	}
	return out
}

// sclCoords scales a flat list treating every pair as x,y.
// This is the same as scl for flat coordinate lists.
func (g *goldberg) sclCoords(coords []float64) []float64 {
	return g.scl(coords)
}

// sclPos scales a slice of pos values (x,y only, beta/flag preserved).
func (g *goldberg) sclPos(positions []pos) []pos {
	out := make([]pos, len(positions))
	for i, p := range positions {
		out[i] = pos{
			x:    math.Round(p.x * g.sf),
			y:    math.Round(p.y * g.sf),
			beta: p.beta,
			flag: p.flag,
		}
	}
	return out
}

// box returns bounding box coordinates {x-r, y-r, x+r, y+r}.
func (g *goldberg) box(x, y, r float64) (float64, float64, float64, float64) {
	return x - r, y - r, x + r, y + r
}

// rotateC rotates point (x,y) around (ox,oy) by beta degrees.
func (g *goldberg) rotateC(x, y, ox, oy, beta float64) (float64, float64) {
	x -= ox
	y -= oy
	rad := beta * math.Pi / 180.0
	cs := math.Cos(rad)
	sn := math.Sin(rad)
	xx := x*cs - y*sn
	yy := x*sn + y*cs
	return xx + ox, yy + oy
}

// rotateItem rotates all coordinates of a canvas item around (ox,oy) by beta degrees.
func (g *goldberg) rotateItem(tag string, ox, oy, beta float64) {
	coords := g.c.ItemCoords(tag)
	if len(coords) < 2 {
		return
	}
	newCoords := make([]float64, len(coords))
	for i := 0; i < len(coords)-1; i += 2 {
		newCoords[i], newCoords[i+1] = g.rotateC(coords[i], coords[i+1], ox, oy, beta)
	}
	g.c.SetItemCoords(tag, newCoords)
}

// moveAbs moves a tagged item so its centroid is at (x,y).
func (g *goldberg) moveAbs(tag string, x, y float64) {
	ox, oy := g.centroid(tag)
	dx := x - ox
	dy := y - oy
	g.c.Move(tag, dx, dy)
}

// centroid returns the center of the bounding box of a tagged item.
func (g *goldberg) centroid(tag string) (float64, float64) {
	return g.anchor(tag, "c")
}

// anchor returns a point on the bounding box of a tagged item.
// where can contain n/s for y, e/w for x, c for center (default).
func (g *goldberg) anchor(tag, where string) (float64, float64) {
	x1, y1, x2, y2 := g.c.BBox(tag)
	fx1, fy1 := float64(x1), float64(y1)
	fx2, fy2 := float64(x2), float64(y2)

	var x, y float64
	if strings.Contains(where, "n") {
		y = fy1
	} else if strings.Contains(where, "s") {
		y = fy2
	} else {
		y = (fy1 + fy2) / 2.0
	}
	if strings.Contains(where, "w") {
		x = fx1
	} else if strings.Contains(where, "e") {
		x = fx2
	} else {
		x = (fx1 + fx2) / 2.0
	}
	return x, y
}

// sine creates a sine-wave line from (x0,y0) to (x1,y1) with given amplitude
// and frequency (wavelength). Returns the canvas item ID.
func (g *goldberg) sine(x0, y0, x1, y1, amp, freq float64, opts ...canvas.ItemOption) int64 {
	step := 2.0
	var xy []float64
	if y0 == y1 { // Horizontal
		for x := x0; x <= x1; x += step {
			beta := (x - x0) * 2 * math.Pi / freq
			y := y0 + amp*math.Sin(beta)
			xy = append(xy, x, y)
		}
	} else { // Vertical
		for y := y0; y <= y1; y += step {
			beta := (y - y0) * 2 * math.Pi / freq
			x := x0 + amp*math.Sin(beta)
			xy = append(xy, x, y)
		}
	}
	if len(xy) < 4 {
		// Need at least 2 points for a line.
		xy = []float64{x0, y0, x1, y1}
	}
	return g.c.CreateLine(xy, opts...)
}

// roundRect returns polygon coordinates for a rounded rectangle.
// The input is a bounding box (x0,y0,x3,y3) and a corner radius in pixels.
func (g *goldberg) roundRect(x0, y0, x3, y3, radius float64) []float64 {
	d := 2 * radius
	maxr := 0.75
	if d > maxr*(x3-x0) {
		d = maxr * (x3 - x0)
	}
	if d > maxr*(y3-y0) {
		d = maxr * (y3 - y0)
	}

	x1 := x0 + d
	x2 := x3 - d
	y1 := y0 + d
	y2 := y3 - d

	return []float64{
		x0, y0, x1, y0, x2, y0, x3, y0, x3, y1, x3, y2,
		x3, y3, x2, y3, x1, y3, x0, y3, x0, y2, x0, y1,
	}
}

// roundPoly2 computes 3 control points for a rounded corner at vertex (x1,y1)
// with edges from (x0,y0) and to (x2,y2). radius is in pixels.
func roundPoly2(x0, y0, x1, y1, x2, y2, radius float64) []float64 {
	d := 2 * radius
	maxr := 0.75

	v1x := x0 - x1
	v1y := y0 - y1
	v2x := x2 - x1
	v2y := y2 - y1

	vlen1 := math.Sqrt(v1x*v1x + v1y*v1y)
	vlen2 := math.Sqrt(v2x*v2x + v2y*v2y)

	if vlen1 == 0 || vlen2 == 0 {
		return []float64{x1, y1, x1, y1, x1, y1}
	}

	if d > maxr*vlen1 {
		d = maxr * vlen1
	}
	if d > maxr*vlen2 {
		d = maxr * vlen2
	}

	return []float64{
		x1 + d*v1x/vlen1, y1 + d*v1y/vlen1,
		x1, y1,
		x1 + d*v2x/vlen2, y1 + d*v2y/vlen2,
	}
}

// roundPoly creates a polygon with rounded corners. xy is a flat coordinate
// list, radii has one radius per vertex. Returns the canvas polygon item ID.
func (g *goldberg) roundPoly(xy []float64, radii []float64, opts ...canvas.ItemOption) int64 {
	n := len(xy) / 2 // number of vertices
	if n < 3 || len(radii) != n {
		return 0
	}

	var knots []float64

	// Previous vertex (wrap around).
	x0 := xy[(n-1)*2]
	y0 := xy[(n-1)*2+1]
	x1 := xy[0]
	y1 := xy[1]

	// Append the first vertex again at the end for lookahead.
	ext := make([]float64, len(xy)+2)
	copy(ext, xy)
	ext[len(xy)] = xy[0]
	ext[len(xy)+1] = xy[1]

	for i := 0; i < n; i++ {
		r := radii[i]
		x2 := ext[i*2+2]
		y2 := ext[i*2+3]

		pts := roundPoly2(x0, y0, x1, y1, x2, y2, r)
		knots = append(knots, pts...)

		x0, y0 = x1, y1
		x1, y1 = x2, y2
	}

	opts = append([]canvas.ItemOption{canvas.Smooth(true)}, opts...)
	return g.c.CreatePolygon(knots, opts...)
}

// sparkle creates a starburst of white lines centered at (cx,cy) with the given tag.
func (g *goldberg) sparkle(cx, cy float64, tag string) {
	// Sparkle endpoint offsets (relative to center 271,304 in original coords).
	endpoints := []float64{
		299, 283, 298, 302, 295, 314, 271, 331,
		239, 310, 242, 292, 256, 274, 281, 273,
	}
	for i := 0; i < len(endpoints); i += 2 {
		g.c.CreateLine([]float64{271, 304, endpoints[i], endpoints[i+1]},
			canvas.OutlineColor("white"), canvas.OutlineWidth(3),
			canvas.Tags(tag))
	}
	g.moveAbs(tag, cx, cy)
}

// h2o draws 3 parallel sine waves for water effect.
func (g *goldberg) h2o(y, freq float64) {
	color := colors["20"]
	g.c.Delete("I20")

	x := math.Round(208 * g.sf)
	y0 := math.Round(428 * g.sf)

	g.sine(x, y0, x, y, 4, freq,
		canvas.Tags("I20", "I20s"), canvas.OutlineWidth(3),
		canvas.OutlineColor(color), canvas.Smooth(true))

	// Copy coords for the two offset waves.
	coords := g.c.ItemCoords("I20s")
	if len(coords) >= 4 {
		coordsA := make([]float64, len(coords))
		copy(coordsA, coords)
		g.c.CreateLine(coordsA, canvas.OutlineWidth(3),
			canvas.OutlineColor(color), canvas.Smooth(true),
			canvas.Tags("I20", "I20a"))
		g.c.Move("I20a", 6, 0)

		coordsB := make([]float64, len(coords))
		copy(coordsB, coords)
		g.c.CreateLine(coordsB, canvas.OutlineWidth(3),
			canvas.OutlineColor(color), canvas.Smooth(true),
			canvas.Tags("I20", "I20b"))
		g.c.Move("I20b", 12, 0)
	}
}

// getStep returns the current step for animation `who`, auto-incrementing.
// If the step map entry is -1 (unset), it initializes to 0.
// Otherwise it increments and returns the new value.
func (g *goldberg) getStep(who int) int {
	v, ok := g.step[who]
	if !ok || v == -1 {
		g.step[who] = 0
		return 0
	}
	g.step[who] = v + 1
	return v + 1
}

// setStep forces the step counter for animation `who` to a specific value.
func (g *goldberg) setStep(who, val int) {
	g.step[who] = val
}

// resetStep clears all step counters and resets the global counter.
func (g *goldberg) resetStep() {
	g.cnt = 0
	for k := range g.step {
		g.step[k] = -1
	}
}

// start sets the animation mode to go.
func (g *goldberg) start() {
	g.mode = mGo
}

// draw0 draws the "START HERE!" text and arrow.
func (g *goldberg) draw0() {
	color := colors["0"]

	g.c.CreateText(699, 119,
		canvas.TextOpt("START HERE!"), canvas.TextColor(color),
		canvas.AnchorOpt(option.AnchorE),
		canvas.Tags("I0", "I0_0"),
		canvas.FontOpt("Times 12 italic bold"))

	g.c.CreateLine([]float64{719, 119, 763, 119},
		canvas.Tags("I0", "I0_1"), canvas.OutlineColor(color),
		canvas.OutlineWidth(5), canvas.Arrow(canvas.ArrowLast),
		canvas.ArrowShape(18, 18, 5))

	g.c.BindItem("I0", event.ButtonPressMask, func(ev *event.Event) {
		g.start()
	})
}

// draw1 draws the ball track with green channel and yellow ball.
func (g *goldberg) draw1() {
	color := colors["1a"]
	color2 := colors["1b"]

	g.c.CreatePolygon([]float64{844, 133, 800, 133, 800, 346, 820, 346, 820, 168, 844, 168, 844, 133},
		canvas.OutlineWidth(3), canvas.FillColor(color), canvas.OutlineNone())

	g.c.CreatePolygon([]float64{771, 133, 685, 133, 685, 168, 751, 168, 751, 346, 771, 346, 771, 133},
		canvas.OutlineWidth(3), canvas.FillColor(color), canvas.OutlineNone())

	x1, y1, x2, y2 := g.box(812, 122, 9)
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.Tags("I1"), canvas.FillColor(color2))
	g.c.BindItem("I1", event.ButtonPressMask, func(ev *event.Event) {
		g.start()
	})
}

// draw2 draws the lever/match assembly.
func (g *goldberg) draw2() {
	color := colors["2"]
	fg := colors["fg"]

	// Fulcrum
	g.c.CreatePolygon([]float64{750, 369, 740, 392, 760, 392},
		canvas.FillColor(fg), canvas.OutlineColor(fg))

	// Strike box
	g.c.CreateRectangle(628, 335, 660, 383,
		canvas.FillNone(), canvas.OutlineColor(fg), canvas.Tags("StrikeBox"))

	// Lever
	g.c.CreateLine([]float64{702, 366, 798, 366},
		canvas.OutlineColor(fg), canvas.OutlineWidth(7), canvas.Tags("I2_0"))

	// R strap
	g.c.CreateLine([]float64{712, 363, 712, 355},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I2_1"))

	// L strap
	g.c.CreateLine([]float64{705, 363, 705, 355},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I2_2"))

	// Match stick
	g.c.CreateLine([]float64{679, 356, 679, 360, 717, 360, 717, 356, 679, 356},
		canvas.OutlineColor(fg), canvas.Tags("I2_3"))

	// Match head
	g.c.CreatePolygon([]float64{671, 352, 677.4, 353.9, 680, 358.5, 677.4, 363.1, 671, 365, 664.6, 363.1, 662, 358.5, 664.6, 353.9},
		canvas.FillColor(color), canvas.OutlineColor(color), canvas.Tags("I2_4"))
}

// draw3 draws the pulley system with weight.
func (g *goldberg) draw3() {
	color := colors["3a"]
	color2 := colors["3b"]
	fg := colors["fg"]

	// 3 Pulleys
	pulleys := []float64{602, 296, 577, 174, 518, 174}
	for i := 0; i < len(pulleys); i += 2 {
		x, y := pulleys[i], pulleys[i+1]
		x1, y1, x2, y2 := g.box(x, y, 13)
		g.c.CreateOval(x1, y1, x2, y2,
			canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
		x1, y1, x2, y2 = g.box(x, y, 2)
		g.c.CreateOval(x1, y1, x2, y2,
			canvas.FillColor(fg), canvas.OutlineColor(fg))
	}

	// Wall to flame
	g.c.CreateLine([]float64{750, 309, 670, 309},
		canvas.Tags("I3_s"), canvas.OutlineWidth(3), canvas.OutlineColor(fg), canvas.Smooth(true))

	// Flame to pulley 1
	g.c.CreateLine([]float64{670, 309, 650, 309},
		canvas.Tags("I3_0"), canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Flame to pulley 1 (continued)
	g.c.CreateLine([]float64{650, 309, 600, 309},
		canvas.Tags("I3_1"), canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Pulley 1 half way to 2
	g.c.CreateLine([]float64{589, 296, 589, 235},
		canvas.Tags("I3_2"), canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Pulley 1 other half to 2
	g.c.CreateLine([]float64{589, 235, 589, 174},
		canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Across the top
	g.c.CreateLine([]float64{577, 161, 518, 161},
		canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Down to weight
	g.c.CreateLine([]float64{505, 174, 505, 205},
		canvas.Tags("I3_w"), canvas.OutlineWidth(3), canvas.OutlineColor(fg))

	// Weight: 2 circles + rectangle
	wx1, wy1 := 515.0, 207.0
	wx2, wy2 := 495.0, 207.0
	bx1, by1, bx2, by2 := g.box(wx1, wy1, 6)
	g.c.CreateOval(bx1, by1, bx2, by2, canvas.Tags("I3_"), canvas.FillColor(color2), canvas.OutlineColor(color2))
	bx1, by1, bx2, by2 = g.box(wx2, wy2, 6)
	g.c.CreateOval(bx1, by1, bx2, by2, canvas.Tags("I3_"), canvas.FillColor(color2), canvas.OutlineColor(color2))
	g.c.CreateRectangle(wx1, wy1-6, wx2, wy2+6, canvas.Tags("I3_"), canvas.FillColor(color2), canvas.OutlineColor(color2))

	// Rounded rectangle weight body
	rr := g.roundRect(492, 220, 518, 263, 15)
	g.c.CreatePolygon(rr, canvas.Smooth(true), canvas.Tags("I3_"), canvas.FillColor(color2), canvas.OutlineColor(color2))

	// Weight top bar
	g.c.CreateLine([]float64{500, 217, 511, 217},
		canvas.Tags("I3_"), canvas.OutlineColor(color2), canvas.OutlineWidth(10))

	// Bottom weight target
	g.c.CreateLine([]float64{502, 393, 522, 393, 522, 465},
		canvas.Tags("I3__"), canvas.OutlineColor(fg),
		canvas.JoinStyleOpt(platform.JoinMiter), canvas.OutlineWidth(10))
}

// draw4 draws the cage/grid.
func (g *goldberg) draw4() {
	color := colors["4"]
	x0, y0, x1, y1 := 527.0, 356.0, 611.0, 464.0

	// Horizontal bars
	for y := y0; y <= y1; y += 12 {
		g.c.CreateLine([]float64{x0, y, x1, y},
			canvas.OutlineColor(color), canvas.OutlineWidth(1))
	}
	// Vertical bars
	for x := x0; x <= x1; x += 12 {
		g.c.CreateLine([]float64{x, y0, x, y1},
			canvas.OutlineColor(color), canvas.OutlineWidth(1))
	}

	// Swing gate
	g.c.CreateLine([]float64{518, 464, 518, 428},
		canvas.Tags("I4"), canvas.OutlineColor(color), canvas.OutlineWidth(3))
}

// draw5 draws the mouse and its tunnel.
func (g *goldberg) draw5() {
	color := colors["5a"]
	color2 := colors["5b"]
	fg := colors["fg"]
	bg := colors["bg"]

	// Tunnel
	g.c.CreatePolygon([]float64{377, 248, 410, 248, 410, 465, 518, 465,
		518, 428, 451, 428, 451, 212, 377, 212},
		canvas.FillColor(color2), canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Mouse body
	g.c.CreatePolygon([]float64{534.5, 445.5, 541, 440, 552, 436, 560, 436, 569, 440, 574, 446, 575, 452, 574, 454,
		566, 456, 554, 456, 545, 456, 537, 454, 530, 452},
		canvas.Tags("I5", "I5_0"), canvas.FillColor(color))

	// Tail
	g.c.CreateLine([]float64{573, 452, 592, 458, 601, 460, 613, 456},
		canvas.Tags("I5", "I5_1"), canvas.OutlineColor(color), canvas.Smooth(true), canvas.OutlineWidth(3))

	// Eye
	g.c.CreatePolygon([]float64{540, 444, 541, 445, 541, 447, 540, 448, 538, 447, 538, 445},
		canvas.Tags("I5", "I5_2"), canvas.FillColor(bg), canvas.OutlineNone(), canvas.Smooth(true))

	// Front leg
	g.c.CreateLine([]float64{538, 454, 535, 461},
		canvas.Tags("I5", "I5_3"), canvas.OutlineColor(color), canvas.OutlineWidth(2))

	// Back leg
	g.c.CreateLine([]float64{566, 455, 569, 462},
		canvas.Tags("I5", "I5_4"), canvas.OutlineColor(color), canvas.OutlineWidth(2))

	// 2nd front leg
	g.c.CreateLine([]float64{544, 455, 545, 460},
		canvas.Tags("I5", "I5_5"), canvas.OutlineColor(color), canvas.OutlineWidth(2))

	// 2nd back leg
	g.c.CreateLine([]float64{560, 455, 558, 460},
		canvas.Tags("I5", "I5_6"), canvas.OutlineColor(color), canvas.OutlineWidth(2))
}

// draw6 draws the gumball machine with rotor, chute, and balls.
func (g *goldberg) draw6() {
	color := colors["6"]
	fg := colors["fg"]

	// Ball holder (rounded rect)
	rr := g.roundRect(324, 130, 391, 204, 10)
	g.c.CreatePolygon(rr, canvas.Smooth(true),
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.FillColor(color))

	// Below the ball holder
	g.c.CreateRectangle(339, 204, 376, 253,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.FillColor(color), canvas.Tags("I6c"))

	// Rotor fill
	x1, y1, x2, y2 := g.box(346, 339, 28)
	g.c.CreateOval(x1, y1, x2, y2, canvas.FillColor(color))
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(2),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(80), canvas.Extent(205))
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(2),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(-41), canvas.Extent(85))

	// Center of rotor
	x1, y1, x2, y2 = g.box(346, 339, 15)
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.FillColor(fg), canvas.Tags("I6m"))

	// Top drop to rotor
	xy := []float64{352, 312, 352, 254, 368, 254, 368, 322}
	g.c.CreatePolygon(xy, canvas.FillColor(color), canvas.OutlineNone())
	g.c.CreateLine(xy, canvas.OutlineColor(fg), canvas.OutlineWidth(2))

	// Poke bottom hole
	g.c.CreateRectangle(353, 240, 367, 300,
		canvas.FillColor(color), canvas.OutlineNone())

	// Poke another hole
	g.c.CreateRectangle(341, 190, 375, 210,
		canvas.FillColor(color), canvas.OutlineNone())

	// Bottom chute
	xy = []float64{368, 356, 368, 403, 389, 403, 389, 464, 320, 464, 320, 403, 352, 403, 352, 366}
	g.c.CreatePolygon(xy, canvas.FillColor(color), canvas.OutlineNone(), canvas.OutlineWidth(2))
	g.c.CreateLine(xy, canvas.OutlineColor(fg), canvas.OutlineWidth(2))

	// On/off rotor
	x1, y1, x2, y2 = g.box(275, 342, 7)
	g.c.CreateOval(x1, y1, x2, y2, canvas.OutlineColor(fg), canvas.FillColor(fg))

	// Fan belt top
	g.c.CreateLine([]float64{276, 334, 342, 325},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Fan belt bottom
	g.c.CreateLine([]float64{276, 349, 342, 353},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// What the mouse pushes
	g.c.CreateLine([]float64{337, 212, 337, 247},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I6_"))
	g.c.CreateLine([]float64{392, 212, 392, 247},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I6_"))
	g.c.CreateLine([]float64{337, 230, 392, 230},
		canvas.OutlineColor(fg), canvas.OutlineWidth(7), canvas.Tags("I6_"))

	// All the balls
	ballColors := []string{"red", "cyan", "orange", "green", "blue", "darkblue"}
	// Repeat 3 times for 18 total (we use 17)
	var allColors []string
	for i := 0; i < 3; i++ {
		allColors = append(allColors, ballColors...)
	}
	for i := 0; i < 17; i++ {
		loc := fmt.Sprintf("%d", -1*i)
		pos := xy6[loc]
		bx1, by1, bx2, by2 := g.box(pos[0], pos[1], 5)
		bc := allColors[i]
		g.c.CreateOval(bx1, by1, bx2, by2,
			canvas.FillColor(bc), canvas.OutlineColor(bc),
			canvas.Tags(fmt.Sprintf("I6_b%d", i)))
	}

	g.draw6a(12, false)
}

// draw6a draws the motor wheel spokes at a given angle.
func (g *goldberg) draw6a(beta float64, scale bool) {
	g.c.Delete("I6_0")

	var ox, oy float64
	if scale {
		s := g.scl([]float64{346, 339})
		ox, oy = s[0], s[1]
	} else {
		ox, oy = 346, 339
	}

	for i := 0; i < 4; i++ {
		b := beta + float64(i)*45
		var x, y float64
		if scale {
			s := g.scl([]float64{28})
			x, y = g.rotateC(s[0], 0, 0, 0, b)
		} else {
			x, y = g.rotateC(28, 0, 0, 0, b)
		}
		g.c.CreateLine([]float64{ox + x, oy + y, ox - x, oy - y},
			canvas.Tags("I6_0"), canvas.OutlineColor(colors["fg"]), canvas.OutlineWidth(2))
	}
}

// draw7 draws the on/off switch box.
func (g *goldberg) draw7() {
	fg := colors["fg"]
	color := colors["7"]

	// Box
	g.c.CreateRectangle(198, 306, 277, 374,
		canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.FillColor(color), canvas.Tags("I7z"))
	g.c.Lower("I7z")

	// Arrow/pointer
	g.c.CreateLine([]float64{275, 343, 230, 349},
		canvas.Tags("I7"), canvas.OutlineColor(fg), canvas.Arrow(canvas.ArrowLast),
		canvas.ArrowShape(23, 23, 8), canvas.OutlineWidth(6))

	// On button
	x1, y1, x2, y2 := g.box(225, 324, 3)
	g.c.CreateOval(x1, y1, x2, y2, canvas.FillColor(fg), canvas.OutlineColor(fg))
	g.c.CreateText(218, 323,
		canvas.TextOpt("on"), canvas.AnchorOpt(option.AnchorE),
		canvas.TextColor(fg), canvas.FontOpt("Times 8"))

	// Off button
	x1, y1, x2, y2 = g.box(225, 350, 3)
	g.c.CreateOval(x1, y1, x2, y2, canvas.FillColor(fg), canvas.OutlineColor(fg))
	g.c.CreateText(218, 349,
		canvas.TextOpt("off"), canvas.AnchorOpt(option.AnchorE),
		canvas.TextColor(fg), canvas.FontOpt("Times 8"))
}

// draw8 draws the spring/coil.
func (g *goldberg) draw8() {
	g.sine(271, 248, 271, 306, 5, 8,
		canvas.Tags("I8_s"), canvas.OutlineColor(colors["8"]), canvas.OutlineWidth(3))
}

// draw9 draws the fan/balloon.
func (g *goldberg) draw9() {
	color := colors["9"]
	fg := colors["fg"]

	g.c.CreateOval(266, 194, 310, 220,
		canvas.OutlineColor(color), canvas.FillColor(color))

	g.c.CreateOval(280, 209, 296, 248,
		canvas.OutlineColor(color), canvas.FillColor(color))

	g.c.CreatePolygon([]float64{288, 249, 252, 249, 260, 240, 280, 234, 296, 234, 316, 240, 324, 249, 288, 249},
		canvas.FillColor(color), canvas.Smooth(true))

	// Spinner
	g.c.CreatePolygon([]float64{248, 205, 265, 214, 264, 205, 265, 196},
		canvas.FillColor(color))

	// Fan blades (large)
	g.c.CreateOval(255, 206, 265, 234,
		canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I9_0"))
	g.c.CreateOval(255, 176, 265, 204,
		canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I9_0"))

	// Fan blades (small)
	g.c.CreateOval(255, 206, 265, 220,
		canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.Tags("I9_1"))
	g.c.CreateOval(255, 190, 265, 204,
		canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.Tags("I9_1"))
}

// draw10 draws the sailboat with waves.
func (g *goldberg) draw10() {
	color := colors["10a"]
	color2 := colors["10b"]
	fg := colors["fg"]
	bg := colors["bg"]

	// Sail
	g.c.CreatePolygon([]float64{191, 230, 233, 230, 233, 178, 191, 178},
		canvas.FillColor(color), canvas.OutlineWidth(3), canvas.OutlineColor(fg), canvas.Tags("I10"))

	// Front arc
	x1, y1, x2, y2 := g.box(209, 204, 31)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.FillColor(color), canvas.ArcStyleOpt(canvas.ArcStylePieslice),
		canvas.StartAngle(120), canvas.Extent(120), canvas.Tags("I10"))
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(120), canvas.Extent(120), canvas.Tags("I10"))

	// Back arc
	x1, y1, x2, y2 = g.box(249, 204, 31)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.FillColor(bg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(120), canvas.Extent(120), canvas.Tags("I10"))
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(120), canvas.Extent(120), canvas.Tags("I10"))

	// Mast
	g.c.CreateLine([]float64{200, 171, 200, 249},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I10"))

	// Bow sprit
	g.c.CreateLine([]float64{159, 234, 182, 234},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I10"))

	// Hull
	g.c.CreateLine([]float64{180, 234, 180, 251, 220, 251},
		canvas.OutlineColor(fg), canvas.OutlineWidth(6), canvas.Tags("I10"))

	// Waves (sine)
	g.sine(92, 255, 221, 255, 2, 25,
		canvas.OutlineColor(color2), canvas.OutlineWidth(1), canvas.Tags("I10w"))

	// Water: get sine coords, trim first 2 and last 2 values, close polygon
	waveCoords := g.c.ItemCoords("I10w")
	var waterXY []float64
	if len(waveCoords) > 8 {
		waterXY = append(waterXY, waveCoords[4:len(waveCoords)-4]...)
	} else {
		waterXY = append(waterXY, waveCoords...)
	}
	waterXY = append(waterXY, 222, 266, 222, 277, 99, 277)
	g.c.CreatePolygon(waterXY, canvas.FillColor(color2), canvas.OutlineColor(color2))

	// Water bottom
	g.c.CreateLine([]float64{222, 266, 222, 277, 97, 277, 97, 266},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Right curve
	x1, y1, x2, y2 = g.box(239, 262, 17)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(95), canvas.Extent(103))

	// Left curve
	x1, y1, x2, y2 = g.box(76, 266, 21)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(190))
}

// draw11 draws the loop-the-loop tube.
func (g *goldberg) draw11() {
	color := colors["11a"]
	color2 := colors["11b"]
	fg := colors["fg"]
	bg := colors["bg"]

	// Color the down tube
	g.c.CreateRectangle(23, 264, 55, 591,
		canvas.FillColor(color))

	// Color the outer loop
	x1, y1, x2, y2 := g.box(71, 460, 48)
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.FillColor(color))

	// Top right side
	g.c.CreateLine([]float64{55, 264, 55, 458},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Bottom right side
	g.c.CreateLine([]float64{55, 504, 55, 591},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Outer loop arc
	x1, y1, x2, y2 = g.box(71, 460, 48)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(110), canvas.Extent(-290),
		canvas.Tags("I11i"))

	// Inner loop
	x1, y1, x2, y2 = g.box(71, 460, 16)
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.FillNone(), canvas.OutlineWidth(3), canvas.Tags("I11i"))
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.FillColor(bg), canvas.OutlineWidth(3))

	// Left side
	g.c.CreateLine([]float64{23, 264, 23, 591},
		canvas.OutlineColor(fg), canvas.OutlineWidth(3))

	// Top left curve
	x1, y1, x2, y2 = g.box(1, 266, 23)
	g.c.CreateArc(x1, y1, x2, y2,
		canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(90))

	// The ball
	x1, y1, x2, y2 = g.box(75, 235, 9)
	g.c.CreateOval(x1, y1, x2, y2,
		canvas.FillColor(color2), canvas.OutlineWidth(3), canvas.Tags("I11"))
}

// draw12 draws the hand.
func (g *goldberg) draw12() {
	xy := []float64{20, 637, 20, 617, 20, 610, 20, 590, 40, 590, 40, 590, 60, 590, 60, 610, 60, 610}
	xy = append(xy, 60, 610, 65, 620, 60, 631)
	xy = append(xy, 60, 631, 60, 637, 60, 662, 60, 669, 52, 669, 56, 669, 50, 669, 50, 662, 50, 637)
	for x := 50.0; x > 20; x -= 10 {
		xy = append(xy, x, 637, x-5, 645, x-10, 637)
	}
	g.c.CreatePolygon(xy, canvas.FillColor(colors["12"]), canvas.OutlineColor(colors["fg"]),
		canvas.Smooth(true), canvas.Tags("I12"), canvas.OutlineWidth(3))
}

// draw13 draws the fax machines.
func (g *goldberg) draw13() {
	color := colors["13a"]
	fg := colors["fg"]
	xy := []float64{86, 663, 149, 663, 149, 704, 50, 704, 50, 681, 64, 681, 86, 671}
	xy2 := []float64{784, 663, 721, 663, 721, 704, 820, 704, 820, 681, 806, 681, 784, 671}
	radii := []float64{2, 9, 9, 8, 5, 5, 2}
	g.roundPoly(xy, radii, canvas.OutlineWidth(3), canvas.OutlineColor(fg), canvas.FillColor(color))
	g.roundPoly(xy2, radii, canvas.OutlineWidth(3), canvas.OutlineColor(fg), canvas.FillColor(color))
	x0, y0, x1, y1 := g.box(56, 677, 4)
	g.c.CreateRectangle(x0, y0, x1, y1, canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I13"))
	x0, y0, x1, y1 = g.box(809, 677, 4)
	g.c.CreateRectangle(x0, y0, x1, y1, canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I13R"))
	g.c.CreateText(112, 687, canvas.TextOpt("FAX"), canvas.TextColor(fg), canvas.FontOpt("Times 12 bold"))
	g.c.CreateText(762, 687, canvas.TextOpt("FAX"), canvas.TextColor(fg), canvas.FontOpt("Times 12 bold"))
	g.c.CreateLine([]float64{138, 663, 148, 636, 178, 636}, canvas.Smooth(true), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateLine([]float64{732, 663, 722, 636, 692, 636}, canvas.Smooth(true), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.sine(149, 688, 720, 688, 5, 15, canvas.Tags("I13_s"), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
}

// draw14 draws the paper in fax.
func (g *goldberg) draw14() {
	color := colors["14"]
	g.c.CreateLine([]float64{102, 661, 113, 632, 130, 618}, canvas.Smooth(true), canvas.OutlineColor(color), canvas.OutlineWidth(3), canvas.Tags("I14L_0"))
	g.c.CreateLine([]float64{148, 629, 125, 640, 124, 662}, canvas.Smooth(true), canvas.OutlineColor(color), canvas.OutlineWidth(3), canvas.Tags("I14L_1"))
	g.draw14a("L")
	g.c.CreateLine([]float64{768.0, 662.5, 767.99, 662.43, 767.93, 662.40}, canvas.Smooth(true), canvas.OutlineColor(color), canvas.OutlineWidth(3), canvas.Tags("I14R_0"))
	g.c.Lower("I14R_0")
	g.c.CreateLine([]float64{745.95, 662.43, 746.00, 662.45, 746.0, 662.5}, canvas.Smooth(true), canvas.OutlineColor(color), canvas.OutlineWidth(3), canvas.Tags("I14R_1"))
	g.c.Lower("I14R_1")
}

// draw14a reconstructs the paper polygon from edge coords.
func (g *goldberg) draw14a(side string) {
	color := colors["14"]
	tag0 := fmt.Sprintf("I14%s_0", side)
	tag1 := fmt.Sprintf("I14%s_1", side)
	tagFill := fmt.Sprintf("I14%s", side)
	xy := g.c.ItemCoords(tag0)
	xy2 := g.c.ItemCoords(tag1)
	if len(xy) < 6 || len(xy2) < 6 {
		return
	}
	x0, y0 := xy[0], xy[1]
	x2, y2 := xy[4], xy[5]
	x3, y3 := xy2[0], xy2[1]
	x5, y5 := xy2[4], xy2[5]
	zz := []float64{
		x0, y0, x0, y0,
		xy[0], xy[1], xy[2], xy[3], xy[4], xy[5],
		x2, y2, x2, y2,
		x3, y3, x3, y3,
		xy2[0], xy2[1], xy2[2], xy2[3], xy2[4], xy2[5],
		x5, y5, x5, y5,
	}
	g.c.Delete(tagFill)
	g.c.CreatePolygon(zz, canvas.Tags(tagFill), canvas.Smooth(true), canvas.FillColor(color), canvas.OutlineColor(color), canvas.OutlineWidth(3))
	g.c.Lower(tagFill)
}

// draw15 draws the light beam apparatus.
func (g *goldberg) draw15() {
	color := colors["15a"]
	fg := colors["fg"]
	g.c.CreateLine([]float64{824, 599, 824, 585, 820, 585, 829, 585}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I15a"))
	g.c.CreateRectangle(789, 599, 836, 643, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateRectangle(778, 610, 788, 632, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateRectangle(766, 617, 776, 625, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateRectangle(633, 600, 681, 640, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateRectangle(635, 567, 657, 599, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateRectangle(765, 557, 784, 583, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.sine(658, 580, 765, 580, 3, 15, canvas.Tags("I15_s"), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
}

// draw16 draws the bell.
func (g *goldberg) draw16() {
	color := colors["16"]
	fg := colors["fg"]
	g.c.CreateRectangle(722, 485, 791, 556, canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	x0, y0, x1, y1 := g.box(752, 515, 25)
	g.c.CreateOval(x0, y0, x1, y1, canvas.FillColor(color), canvas.OutlineColor("black"), canvas.Tags("I16b"), canvas.OutlineWidth(2))
	x0, y0, x1, y1 = g.box(752, 515, 5)
	g.c.CreateOval(x0, y0, x1, y1, canvas.FillColor("black"), canvas.OutlineColor("black"), canvas.Tags("I16b"))
	g.c.CreateLine([]float64{784, 523, 764, 549}, canvas.OutlineWidth(3), canvas.Tags("I16c"), canvas.OutlineColor(fg))
	x0, y0, x1, y1 = g.box(784, 523, 4)
	g.c.CreateOval(x0, y0, x1, y1, canvas.FillColor(fg), canvas.OutlineColor(fg), canvas.Tags("I16d"))
}

// draw17 draws the cat.
func (g *goldberg) draw17() {
	color := colors["17"]
	fg := colors["fg"]
	g.c.CreateLine([]float64{584, 556, 722, 556}, canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	g.c.CreateLine([]float64{584, 485, 722, 485}, canvas.OutlineColor(fg), canvas.OutlineWidth(3))
	// Body
	g.c.CreateArc(664, 523, 717, 549, canvas.OutlineColor(fg), canvas.FillColor(color), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleChord), canvas.StartAngle(128), canvas.Extent(-260), canvas.Tags("I17"))
	// Paws
	g.c.CreateOval(709, 554, 690, 543, canvas.OutlineColor(fg), canvas.FillColor(color), canvas.OutlineWidth(3), canvas.Tags("I17"))
	g.c.CreateOval(657, 544, 676, 555, canvas.OutlineColor(fg), canvas.FillColor(color), canvas.OutlineWidth(3), canvas.Tags("I17"))
	// Lower face
	x0, y0, x1, y1 := g.box(660, 535, 15)
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(3),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(150), canvas.Extent(240), canvas.Tags("I17_"))
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(color), canvas.OutlineWidth(1),
		canvas.ArcStyleOpt(canvas.ArcStyleChord), canvas.StartAngle(150), canvas.Extent(240), canvas.Tags("I17_"))
	// Ears
	g.c.CreateLine([]float64{674, 529, 670, 513, 662, 521, 658, 521, 650, 513, 647, 529}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
	g.c.CreatePolygon([]float64{674, 529, 670, 513, 662, 521, 658, 521, 650, 513, 647, 529}, canvas.FillColor(color), canvas.OutlineNone(), canvas.Tags("I17_", "I17_c"))
	// Whiskers left
	g.c.CreateLine([]float64{652, 542, 628, 539}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
	g.c.CreateLine([]float64{652, 543, 632, 545}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
	g.c.CreateLine([]float64{652, 546, 632, 552}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
	// Whiskers right
	g.c.CreateLine([]float64{668, 543, 687, 538}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_", "I17w"))
	g.c.CreateLine([]float64{668, 544, 688, 546}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_", "I17w"))
	g.c.CreateLine([]float64{668, 547, 688, 553}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_", "I17w"))
	// Eyes and mouth
	g.c.CreateLine([]float64{649, 530, 654, 538, 659, 530}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Smooth(true), canvas.Tags("I17"))
	g.c.CreateLine([]float64{671, 530, 666, 538, 661, 530}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Smooth(true), canvas.Tags("I17"))
	g.c.CreateLine([]float64{655, 543, 660, 551, 665, 543}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Smooth(true), canvas.Tags("I17"))
}

// draw18 draws the slingshot.
func (g *goldberg) draw18() {
	fg := colors["fg"]
	g.c.CreateLine([]float64{721, 506, 627, 506}, canvas.OutlineWidth(4), canvas.OutlineColor(fg), canvas.Tags("I18"))
	g.c.CreateOval(607, 500, 628, 513, canvas.FillColor(colors["18"]), canvas.Tags("I18a"))
	g.c.CreateLine([]float64{526, 513, 606, 507, 494, 502}, canvas.OutlineColor(fg), canvas.OutlineWidth(4), canvas.Tags("I18b"))
	g.c.CreateLine([]float64{485, 490, 510, 540, 510, 575, 510, 540, 535, 491}, canvas.OutlineColor(fg), canvas.OutlineWidth(6))
}

// draw19 draws the water pipe.
func (g *goldberg) draw19() {
	color := colors["19"]
	fg := colors["fg"]
	xx := [][2]float64{{249, 181}, {155, 118}, {86, 55}, {22, 0}}
	for _, pair := range xx {
		g.c.CreateRectangle(pair[0], 453, pair[1], 467, canvas.FillColor(color), canvas.Tags("I19"))
		g.c.CreateLine([]float64{pair[0], 453, pair[1], 453}, canvas.OutlineColor(fg), canvas.OutlineWidth(1))
		g.c.CreateLine([]float64{pair[0], 467, pair[1], 467}, canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	}
	g.c.Raise("I11i")
	// Bulge
	x0, y0, x1, y1 := g.box(168, 460, 16)
	g.c.CreateOval(x0, y0, x1, y1, canvas.FillColor(color))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(21), canvas.Extent(136))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(-21), canvas.Extent(-130))
	// First joint
	g.c.CreateRectangle(249, 447, 255, 473, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	// Bend up
	x0, y0, x1, y1 = g.box(257, 433, 34)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(color), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(0), canvas.Extent(-91))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(-90))
	x0, y0, x1, y1 = g.box(257, 433, 20)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(colors["bg"]), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(0), canvas.Extent(-92))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(-90))
	// Bend left
	x0, y0, x1, y1 = g.box(257, 421, 34)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(color), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(1), canvas.Extent(91))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(90))
	x0, y0, x1, y1 = g.box(257, 421, 20)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(colors["bg"]), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(0), canvas.Extent(90))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(0), canvas.Extent(90))
	// Bend down
	x0, y0, x1, y1 = g.box(243, 421, 34)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(color), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(90), canvas.Extent(90))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(90), canvas.Extent(90))
	x0, y0, x1, y1 = g.box(243, 421, 20)
	g.c.CreateArc(x0, y0, x1, y1, canvas.FillColor(colors["bg"]), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStylePieslice), canvas.StartAngle(90), canvas.Extent(90))
	g.c.CreateArc(x0, y0, x1, y1, canvas.OutlineColor(fg), canvas.OutlineWidth(1), canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(90), canvas.Extent(90))
	// Joints
	g.c.CreateRectangle(270, 427, 296, 433, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	g.c.CreateRectangle(270, 421, 296, 427, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	g.c.CreateRectangle(249, 382, 255, 408, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	g.c.CreateRectangle(243, 382, 249, 408, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	g.c.CreateRectangle(203, 420, 229, 426, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(1))
	// Handle
	x0, y0, x1, y1 = g.box(168, 460, 6)
	g.c.CreateOval(x0, y0, x1, y1, canvas.FillColor(fg), canvas.Tags("I19a"))
	g.c.CreateLine([]float64{168, 460, 168, 512}, canvas.OutlineColor(fg), canvas.OutlineWidth(5), canvas.Tags("I19b"))
}

// draw21 draws the bucket.
func (g *goldberg) draw21() {
	color := colors["21"]
	fg := colors["fg"]
	g.c.CreateLine([]float64{217, 451, 244, 490}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21_a"))
	g.c.CreateLine([]float64{201, 467, 182, 490}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21_a"))
	xy := []float64{245, 490, 237, 535, 189, 535, 181, 490}
	g.c.CreatePolygon(xy, canvas.FillColor(color), canvas.OutlineNone(), canvas.Tags("I21", "I21f"))
	g.c.CreateLine([]float64{245, 490, 237, 535}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21"))
	g.c.CreateLine([]float64{189, 535, 181, 490}, canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21"))
	g.c.CreateOval(182, 486, 244, 498, canvas.FillColor(color), canvas.Tags("I21", "I21f"))
	g.c.CreateOval(182, 486, 244, 498, canvas.FillNone(), canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21", "I21t"))
	g.c.CreateOval(189, 532, 237, 540, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I21", "I21b"))
}

// draw23 draws the blow dart.
func (g *goldberg) draw23() {
	fg := colors["fg"]
	g.c.CreateRectangle(185, 623, 253, 650, canvas.FillColor("black"), canvas.OutlineColor(fg), canvas.OutlineWidth(2), canvas.Tags("I23a"))
	g.c.CreateOval(187, 592, 241, 623, canvas.FillColor(colors["23a"]), canvas.OutlineNone(), canvas.Tags("I23b"))
	g.c.CreateArc(187, 592, 241, 623, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23b"),
		canvas.ArcStyleOpt(canvas.ArcStyleArc), canvas.StartAngle(12), canvas.Extent(336))
	g.c.CreatePolygon([]float64{239, 604, 258, 589, 258, 625, 239, 610}, canvas.FillColor(colors["23a"]), canvas.Tags("I23b"))
	g.c.CreateLine([]float64{239, 604, 258, 589, 258, 625, 239, 610}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23b"))
	g.c.CreateOval(285, 611, 250, 603, canvas.FillColor(colors["23b"]), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23d"))
	g.c.CreatePolygon([]float64{249, 596, 249, 618, 264, 607, 249, 596}, canvas.FillColor(colors["23c"]), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23d"))
	g.c.CreateLine([]float64{249, 607, 268, 607}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23d"))
	g.c.CreateLine([]float64{285, 607, 305, 607}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I23d"))
}

// draw24 draws the balloon.
func (g *goldberg) draw24() {
	color := colors["24a"]
	fg := colors["fg"]
	g.c.CreateOval(366, 518, 462, 665, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I24"))
	g.c.CreateLine([]float64{414, 666, 414, 729}, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I24"))
	g.c.CreatePolygon([]float64{410, 666, 404, 673, 422, 673, 418, 666}, canvas.FillColor(color), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I24"))
	g.c.CreateLine([]float64{387, 567, 390, 549, 404, 542}, canvas.OutlineColor(fg), canvas.Smooth(true), canvas.OutlineWidth(2), canvas.Tags("I24"))
	g.c.CreateLine([]float64{395, 568, 399, 554, 413, 547}, canvas.OutlineColor(fg), canvas.Smooth(true), canvas.OutlineWidth(2), canvas.Tags("I24"))
	g.c.CreateLine([]float64{403, 570, 396, 555, 381, 553}, canvas.OutlineColor(fg), canvas.Smooth(true), canvas.OutlineWidth(2), canvas.Tags("I24"))
	g.c.CreateLine([]float64{408, 564, 402, 547, 386, 545}, canvas.OutlineColor(fg), canvas.Smooth(true), canvas.OutlineWidth(2), canvas.Tags("I24"))
}

// drawAll clears the canvas and redraws all elements, then applies global scale.
func (g *goldberg) drawAll() {
	g.resetStep()
	g.c.Delete("all")
	drawFuncs := []func(){
		g.draw0, g.draw1, g.draw2, g.draw3, g.draw4, g.draw5, g.draw6,
		g.draw7, g.draw8, g.draw9, g.draw10, g.draw11, g.draw12, g.draw13,
		g.draw14, g.draw15, g.draw16, g.draw17, g.draw18, g.draw19,
		nil, g.draw21, nil, g.draw23, g.draw24,
	}
	for _, fn := range drawFuncs {
		if fn != nil {
			fn()
		}
	}
	g.c.Scale("all", 0, 0, g.sf, g.sf)
}

// reset redraws everything and returns to start mode.
func (g *goldberg) reset() {
	g.drawAll()
	g.mode = mStart
	g.active = []int{0}
}

// nextStep drives one animation tick of all active state machines.
func (g *goldberg) nextStep() int {
	rval := 0
	if g.mode != mStart && g.mode != mDone {
		g.cnt++
	}
	var alive []int
	for _, who := range g.active {
		n := g.callMove(who)
		if n&1 != 0 {
			alive = append(alive, who)
		}
		if n&2 != 0 {
			alive = append(alive, who+1)
			rval = 1
		}
		if n&4 != 0 {
			g.mode = mDone
			g.active = nil
			return 1
		}
	}
	g.active = alive
	return rval
}

// callMove dispatches to the appropriate Move function.
func (g *goldberg) callMove(who int) int {
	switch who {
	case 0:
		return g.move0()
	case 1:
		return g.move1()
	case 2:
		return g.move2()
	case 3:
		return g.move3()
	case 4:
		return g.move4()
	case 5:
		return g.move5()
	case 6:
		return g.move6()
	case 7:
		return g.move7()
	case 8:
		return g.move8()
	case 9:
		return g.move9()
	case 10:
		return g.move10()
	case 11:
		return g.move11()
	case 12:
		return g.move12()
	case 13:
		return g.move13()
	case 14:
		return g.move14()
	case 15:
		return g.move15()
	case 16:
		return g.move16()
	case 17:
		return g.move17()
	case 18:
		return g.move18()
	case 19:
		return g.move19()
	case 20:
		return g.move20()
	case 21:
		return g.move21()
	case 22:
		return g.move22()
	case 23:
		return g.move23()
	case 24:
		return g.move24()
	case 25:
		return g.move25()
	case 26:
		return g.move26()
	}
	return 0
}

// go_ is the main animation loop callback.
func (g *goldberg) go_() {
	if g.mode == mDone || g.mode == -1 {
		return
	}
	n := 0
	if g.mode != mPause {
		n = g.nextStep()
	}
	if g.mode == mSStep {
		g.mode = mPause
	}
	if g.mode == mBStep && n != 0 {
		g.mode = mSStep
	}
	delay := time.Duration(delays[g.speed]) * time.Millisecond
	g.app.After(delay, g.go_)
}

// Move functions

func (g *goldberg) move0() int {
	step := g.getStep(0)
	if g.mode > mStart {
		g.moveAbs("I0", -100, -100)
		return 2
	}
	positions := g.sclPos([]pos{
		{719, 119, 0, ""}, {724, 119, 0, ""}, {729, 119, 0, ""}, {734, 119, 0, ""},
		{739, 119, 0, ""}, {734, 119, 0, ""}, {729, 119, 0, ""}, {724, 119, 0, ""},
	})
	idx := step % len(positions)
	x, y := positions[idx].x, positions[idx].y
	g.c.SetItemCoords("I0_0", []float64{x - math.Round(20*g.sf), y})
	g.c.SetItemCoords("I0_1", []float64{x, y, x + math.Round(44*g.sf), y})
	return 1
}

func (g *goldberg) move1() int {
	step := g.getStep(1)
	positions := g.sclPos([]pos{
		{807, 122, 0, ""}, {802, 122, 0, ""}, {797, 123, 0, ""}, {793, 124, 0, ""}, {789, 129, 0, ""}, {785, 153, 0, ""},
		{785, 203, 0, ""}, {785, 278, 0, "x"}, {785, 367, 0, ""}, {810, 392, 0, ""}, {816, 438, 0, ""}, {821, 503, 0, ""},
		{824, 585, 0, "y"}, {838, 587, 0, ""}, {848, 593, 0, ""}, {857, 601, 0, ""}, {-100, -100, 0, ""},
	})
	if step >= len(positions) {
		return 0
	}
	w := positions[step]
	g.moveAbs("I1", w.x, w.y)
	if w.flag == "y" {
		g.move15a()
	}
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move2() int {
	step := g.getStep(2)
	stages := []int{0, 0, 1, 2, 0, 2, 1, 0, 1, 2, 0, 2, 1}
	flameXY := map[int][]float64{
		0: g.scl([]float64{686, 333, 692, 323, 682, 316, 674, 309, 671, 295, 668, 307, 662, 318, 662, 328, 671, 336}),
		1: g.scl([]float64{687, 331, 698, 322, 703, 295, 680, 320, 668, 297, 663, 311, 661, 327, 671, 335}),
		2: g.scl([]float64{686, 331, 704, 322, 688, 300, 678, 283, 678, 283, 674, 298, 666, 309, 660, 324, 672, 336}),
	}
	if step >= len(stages) {
		g.c.Delete("I2")
		return 0
	}
	if step == 0 {
		ox, oy := g.anchor("I2_0", "s")
		for i := 0; ; i++ {
			tag := fmt.Sprintf("I2_%d", i)
			if len(g.c.FindWithTag(tag)) == 0 {
				break
			}
			g.rotateItem(tag, ox, oy, 20)
		}
		g.c.CreatePolygon([]float64{0, 0, 1, 0, 0, 1}, canvas.Tags("I2"), canvas.Smooth(true), canvas.FillColor(colors["2"]))
		return 1
	}
	g.c.SetItemCoords("I2", flameXY[stages[step]])
	if step == 7 {
		return 3
	}
	return 1
}

func (g *goldberg) move3() int {
	step := g.getStep(3)
	positions := g.sclPos([]pos{
		{505, 247, 0, ""}, {505, 297, 0, ""}, {505, 386.5, 0, ""}, {505, 386.5, 0, ""},
	})
	rope := map[int][]float64{
		0: g.scl([]float64{750, 309, 729, 301, 711, 324, 690, 300}),
		1: g.scl([]float64{750, 309, 737, 292, 736, 335, 717, 315, 712, 320}),
		2: g.scl([]float64{750, 309, 737, 309, 740, 343, 736, 351, 725, 340}),
		3: g.scl([]float64{750, 309, 738, 321, 746, 345, 742, 356}),
	}
	if step >= len(positions) {
		return 0
	}
	g.c.Delete(fmt.Sprintf("I3_%d", step))
	g.moveAbs("I3_", positions[step].x, positions[step].y)
	g.c.SetItemCoords("I3_s", rope[step])
	wire := g.scl([]float64{505, 174})
	g.c.SetItemCoords("I3_w", []float64{wire[0], wire[1], positions[step].x, positions[step].y})
	if step == 2 {
		g.c.Move("I3__", 0, 30)
		return 2
	}
	return 1
}

func (g *goldberg) move4() int {
	step := g.getStep(4)
	angles := []float64{-10, -20, -30, -30}
	if step >= len(angles) {
		return 0
	}
	g.rotateItem("I4", math.Round(518*g.sf), math.Round(464*g.sf), angles[step])
	g.c.Raise("I4")
	if step == 3 {
		return 3
	}
	return 1
}

func (g *goldberg) move5() int {
	step := g.getStep(5)
	positions := g.sclPos([]pos{
		{553, 452, 0, ""}, {533, 452, 0, ""}, {513, 452, 0, ""}, {493, 452, 0, ""}, {473, 452, 0, ""},
		{463, 442, 30, ""}, {445.5, 441.5, 30, ""}, {425.5, 434.5, 30, ""}, {422, 414, 0, ""}, {422, 394, 0, ""},
		{422, 374, 0, ""}, {422, 354, 0, ""}, {422, 334, 0, ""}, {422, 314, 0, ""}, {422, 294, 0, ""},
		{422, 274, -30, ""}, {422, 260.5, -30, "x"}, {422.5, 248.5, -28, ""}, {425, 237, 0, ""},
	})
	if step >= len(positions) {
		return 0
	}
	p := positions[step]
	g.moveAbs("I5", p.x, p.y)
	if p.beta != 0 {
		ox, oy := g.centroid("I5_0")
		for i := 0; i <= 6; i++ {
			g.rotateItem(fmt.Sprintf("I5_%d", i), ox, oy, p.beta)
		}
	}
	if p.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move6() int {
	step := g.getStep(6)
	if step > 62 {
		return 0
	}
	if step < 2 {
		g.c.Move("I6_", -7, 0)
		if step == 1 {
			xy := g.scl([]float64{348, 226, 365, 240})
			g.c.CreateRectangle(xy[0], xy[1], xy[2], xy[3], canvas.FillColor(colors["6"]))
		}
		return 1
	}
	s := step - 1
	for i := 0; i <= (s-1)/3; i++ {
		tag := fmt.Sprintf("I6_b%d", i)
		if len(g.c.FindWithTag(tag)) == 0 {
			break
		}
		loc := s - 3*i
		key := fmt.Sprintf("%d,%d", loc, i)
		if coords, ok := xy6[key]; ok {
			sc := g.scl([]float64{coords[0], coords[1]})
			g.moveAbs(tag, sc[0], sc[1])
		} else {
			skey := fmt.Sprintf("%d", loc)
			if coords, ok := xy6[skey]; ok {
				sc := g.scl([]float64{coords[0], coords[1]})
				g.moveAbs(tag, sc[0], sc[1])
			}
		}
	}
	if s%3 == 1 {
		first := (s + 2) / 3
		for i := first; ; i++ {
			tag := fmt.Sprintf("I6_b%d", i)
			if len(g.c.FindWithTag(tag)) == 0 {
				break
			}
			loc := first - i
			skey := fmt.Sprintf("%d", loc)
			if coords, ok := xy6[skey]; ok {
				sc := g.scl([]float64{coords[0], coords[1]})
				g.moveAbs(tag, sc[0], sc[1])
			}
		}
	}
	if s >= 3 {
		g.draw6a(float64(12+s*15), true)
	}
	if s == 3 {
		return 3
	}
	return 1
}

func (g *goldberg) move7() int {
	step := g.getStep(7)
	if step > 30 {
		return 0
	}
	g.rotateItem("I7", math.Round(275*g.sf), math.Round(343*g.sf), 1.0)
	if step == 30 {
		return 3
	}
	return 1
}

func (g *goldberg) move8() int {
	step := g.getStep(8)
	if step > 3 {
		return 0
	}
	if step == 0 {
		sx, sy := g.anchor("I8_s", "s")
		g.sparkle(sx, sy, "I8")
		return 1
	} else if step == 1 {
		cx, cy := g.anchor("I8_s", "c")
		g.moveAbs("I8", cx, cy)
	} else if step == 2 {
		nx, ny := g.anchor("I8_s", "n")
		g.moveAbs("I8", nx, ny)
	} else {
		g.c.Delete("I8")
	}
	if step == 2 {
		return 3
	}
	return 1
}

func (g *goldberg) move9() int {
	step := g.getStep(9)
	if step&1 != 0 {
		g.c.ItemConfigure("I9_0", canvas.OutlineWidth(4))
		g.c.ItemConfigure("I9_1", canvas.OutlineWidth(1))
		g.c.Lower("I9_1")
	} else {
		g.c.ItemConfigure("I9_0", canvas.OutlineWidth(1))
		g.c.ItemConfigure("I9_1", canvas.OutlineWidth(4))
		g.c.Lower("I9_0")
	}
	if step == 0 {
		return 3
	}
	return 1
}

func (g *goldberg) move10() int {
	step := g.getStep(10)
	positions := g.sclPos([]pos{
		{195, 212, 0, ""}, {193, 212, 0, ""}, {190, 212, 0, ""}, {186, 212, 0, ""}, {181, 212, 0, ""}, {176, 212, 0, ""},
		{171, 212, 0, ""}, {166, 212, 0, ""}, {161, 212, 0, ""}, {156, 212, 0, ""}, {151, 212, 0, ""}, {147, 212, 0, ""}, {142, 212, 0, ""},
		{137, 212, 0, ""}, {132, 212, 0, "x"}, {127, 212, 0, ""}, {121, 212, 0, ""}, {116, 212, 0, ""}, {111, 212, 0, ""},
	})
	if step >= len(positions) {
		return 0
	}
	w := positions[step]
	g.moveAbs("I10", w.x, w.y)
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move11() int {
	step := g.getStep(11)
	positions := g.sclPos([]pos{
		{75, 235, 0, ""}, {70, 235, 0, ""}, {65, 237, 0, ""}, {56, 240, 0, ""}, {46, 247, 0, ""}, {38, 266, 0, ""}, {38, 296, 0, ""},
		{38, 333, 0, ""}, {38, 399, 0, ""}, {38, 475, 0, ""}, {74, 496, 0, ""}, {105, 472, 0, ""}, {100, 437, 0, ""}, {65, 423, 0, ""},
		{-100, -100, 0, ""}, {38, 505, 0, ""}, {38, 527, 0, "x"}, {38, 591, 0, ""},
	})
	if step >= len(positions) {
		return 0
	}
	w := positions[step]
	g.moveAbs("I11", w.x, w.y)
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move12() int {
	step := g.getStep(12)
	if step >= 1 {
		return 0
	}
	p := g.sclPos([]pos{{42, 641, 0, "x"}})[0]
	g.moveAbs("I12", p.x, p.y)
	return 3
}

func (g *goldberg) move13() int {
	step := g.getStep(13)
	if step == 9 {
		g.moveAbs("I13_star", -100, -100)
		g.c.ItemConfigure("I13R", canvas.FillColor(colors["13b"]), canvas.OutlineWidth(2))
		return 2
	}
	if step == 0 {
		g.c.Delete("I13")
		g.sparkle(-100, -100, "I13_star")
		return 1
	}
	x0, y0 := g.anchor("I13_s", "w")
	x1, _ := g.anchor("I13_s", "e")
	x := x0 + (x1-x0)*float64(step-1)/7.0
	g.moveAbs("I13_star", x, y0)
	return 1
}

func (g *goldberg) move14() int {
	step := g.getStep(14)
	sc := 0.9 - 0.05*float64(step)
	if sc < 0.3 {
		g.c.Delete("I14L")
		return 0
	}
	coords0 := g.c.ItemCoords("I14L_0")
	if len(coords0) >= 2 {
		g.c.Scale("I14L_0", coords0[0], coords0[1], sc, sc)
	}
	coords1 := g.c.ItemCoords("I14L_1")
	if n := len(coords1); n >= 2 {
		g.c.Scale("I14L_1", coords1[n-2], coords1[n-1], sc, sc)
	}
	g.draw14a("L")
	scR := 1.0 / (0.35 + 0.05*float64(step))
	coords0 = g.c.ItemCoords("I14R_0")
	if len(coords0) >= 2 {
		g.c.Scale("I14R_0", coords0[0], coords0[1], scR, scR)
	}
	coords1 = g.c.ItemCoords("I14R_1")
	if n := len(coords1); n >= 2 {
		g.c.Scale("I14R_1", coords1[n-2], coords1[n-1], scR, scR)
	}
	g.draw14a("R")
	if step == 10 {
		return 3
	}
	return 1
}

func (g *goldberg) move15a() {
	g.c.Scale("I15a", math.Round(824*g.sf), math.Round(599*g.sf), 1, 0.3)
	xy := g.scl([]float64{765, 621, 681, 621})
	g.c.CreateLine(xy, canvas.Dash(6, 4), canvas.OutlineWidth(3), canvas.OutlineColor(colors["15b"]), canvas.Tags("I15"))
}

func (g *goldberg) move15() int {
	step := g.getStep(15)
	if step == 8 {
		g.moveAbs("I15_star", -100, -100)
		return 2
	}
	if step == 0 {
		g.sparkle(-100, -100, "I15_star")
		g.c.SetItemCoords("I15", g.scl([]float64{765, 621, 745, 621}))
		return 1
	}
	x0, y0 := g.anchor("I15_s", "w")
	x1, _ := g.anchor("I15_s", "e")
	x := x0 + (x1-x0)*float64(step-1)/6.0
	g.moveAbs("I15_star", x, y0)
	return 1
}

func (g *goldberg) move16() int {
	step := g.getStep(16)
	ox := math.Round(760 * g.sf)
	oy := math.Round(553 * g.sf)
	if step&1 != 0 {
		g.c.Move("I16b", 3, 0)
		g.rotateItem("I16c", ox, oy, 12)
		g.rotateItem("I16d", ox, oy, 12)
	} else {
		g.c.Move("I16b", -3, 0)
		g.rotateItem("I16c", ox, oy, -12)
		g.rotateItem("I16d", ox, oy, -12)
	}
	if step == 1 {
		return 3
	}
	return 1
}

func (g *goldberg) move17() int {
	step := g.getStep(17)
	if step == 0 {
		g.c.Delete("I17")
		fg := colors["fg"]
		// Mouth
		g.c.CreateLine(g.scl([]float64{655, 543, 660, 535, 665, 543}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Smooth(true), canvas.Tags("I17_"))
		// Eyes
		x0, y0, x1, y1 := g.box(654, 530, 4)
		g.c.CreateOval(x0*g.sf, y0*g.sf, x1*g.sf, y1*g.sf, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.FillNone(), canvas.Tags("I17_"))
		x0, y0, x1, y1 = g.box(666, 530, 4)
		g.c.CreateOval(x0*g.sf, y0*g.sf, x1*g.sf, y1*g.sf, canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.FillNone(), canvas.Tags("I17_"))
		g.c.Move("I17_", 0, -20)
		// Front legs
		g.c.CreateLine(g.scl([]float64{652, 528, 652, 554}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
		g.c.CreateLine(g.scl([]float64{670, 528, 670, 554}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
		// Body
		g.c.CreatePolygon(g.scl([]float64{675, 506, 694, 489, 715, 513, 715, 513, 715, 513, 716, 525, 716, 525, 716, 525,
			706, 530, 695, 530, 679, 535, 668, 527, 668, 527, 668, 527, 675, 522, 676, 517, 677, 512}),
			canvas.FillColor(colors["17"]), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Smooth(true), canvas.Tags("I17_"))
		// Back legs
		g.c.CreateLine(g.scl([]float64{716, 514, 716, 554}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
		g.c.CreateLine(g.scl([]float64{694, 532, 694, 554}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Tags("I17_"))
		// Tail
		g.c.CreateLine(g.scl([]float64{715, 514, 718, 506, 719, 495, 716, 488}), canvas.OutlineColor(fg), canvas.OutlineWidth(3), canvas.Smooth(true), canvas.Tags("I17_"))
		g.c.Raise("I17w")
		g.c.Move("I17_", -5, 0)
		return 2
	}
	return 0
}

func (g *goldberg) move18() int {
	step := g.getStep(18)
	positions := g.sclPos([]pos{
		{587, 506, 0, ""}, {537, 506, 0, ""}, {466, 506, 0, ""}, {376, 506, 0, ""}, {266, 506, 0, "x"}, {136, 506, 0, ""},
		{16, 506, 0, ""}, {-100, -100, 0, ""},
	})
	band := map[int][]float64{
		0: g.scl([]float64{490, 502, 719, 507, 524, 512}),
		1: g.scl([]float64{491, 503, 524, 557, 563, 505, 559, 496, 546, 506, 551, 525, 553, 536, 538, 534, 532, 519, 529, 499}),
		2: g.scl([]float64{491, 503, 508, 563, 542, 533, 551, 526, 561, 539, 549, 550, 530, 500}),
		3: g.scl([]float64{491, 503, 508, 563, 530, 554, 541, 562, 525, 568, 519, 544, 530, 501}),
	}
	if step >= len(positions) {
		return 0
	}
	if step == 0 {
		g.c.Delete("I18")
		g.c.ItemConfigure("I18b", canvas.Smooth(true))
	}
	if b, ok := band[step]; ok {
		g.c.SetItemCoords("I18b", b)
	}
	w := positions[step]
	g.moveAbs("I18a", w.x, w.y)
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move19() int {
	step := g.getStep(19)
	if step >= 3 {
		return 2
	}
	ox, oy := g.centroid("I19a")
	g.rotateItem("I19b", ox, oy, 30)
	return 1
}

func (g *goldberg) move20() int {
	step := g.getStep(20)
	type waterStep struct {
		y, freq float64
		flag    string
	}
	ws := []waterStep{
		{451, 20, ""}, {462, 40, ""}, {473, 40, ""}, {484, 40, ""}, {496, 40, ""},
		{504, 40, ""}, {513, 40, ""}, {523, 40, ""}, {532, 40, "x"},
	}
	if step >= len(ws) {
		return 0
	}
	g.c.Delete("I20")
	w := ws[step]
	g.h2o(math.Round(w.y*g.sf), w.freq)
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move21() int {
	step := g.getStep(21)
	if step >= 30 {
		return 0
	}
	coords := g.c.ItemCoords("I21b")
	if len(coords) < 4 {
		return 0
	}
	x1b, y1b, x2b, y2b := coords[0], coords[1], coords[2], coords[3]
	target := g.scl([]float64{183, 492, 243, 504})
	tX1, tY1, tX2, _ := target[0], target[1], target[2], target[3]
	f := float64(step) / 30.0
	y2b -= math.Round(3 * g.sf)
	xx1 := x1b + (tX1-x1b)*f
	yy1 := y1b + (tY1-y1b)*f
	xx2 := x2b + (tX2-x2b)*f
	g.c.ItemConfigure("I21b", canvas.FillColor(colors["20"]))
	g.c.Delete("I21w")
	g.c.CreatePolygon([]float64{x2b, y2b, x1b, y1b, xx1, yy1, xx2, yy1}, canvas.Tags("I21", "I21w"), canvas.OutlineNone(), canvas.FillColor(colors["20"]))
	g.c.Lower("I21w")
	g.c.Raise("I21b")
	g.c.Lower("I21f")
	if step == 29 {
		return 3
	}
	return 1
}

func (g *goldberg) move22() int {
	step := g.getStep(22)
	positions := g.sclPos([]pos{
		{213, 513, 0, ""}, {213, 523, 0, ""}, {213, 543, 0, "x"}, {213, 583, 0, ""}, {213, 593, 0, ""},
	})
	if step == 0 {
		g.c.ItemConfigure("I21f", canvas.FillColor(colors["22"]))
	}
	if step >= len(positions) {
		return 0
	}
	w := positions[step]
	g.moveAbs("I21", w.x, w.y)
	g.h2o(w.y, 40)
	g.c.Delete("I21_a")
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move23() int {
	step := g.getStep(23)
	positions := g.sclPos([]pos{
		{277, 607, 0, ""}, {287, 607, 0, ""}, {307, 607, 0, "x"}, {347, 607, 0, ""}, {407, 607, 0, ""}, {487, 607, 0, ""},
		{587, 607, 0, ""}, {687, 607, 0, ""}, {787, 607, 0, ""}, {-100, -100, 0, ""},
	})
	if step >= len(positions) {
		return 0
	}
	if step <= 1 {
		ax, ay := g.anchor("I23a", "n")
		g.c.Scale("I23b", ax, ay, 0.9, 0.5)
	}
	w := positions[step]
	g.moveAbs("I23d", w.x, w.y)
	if w.flag == "x" {
		return 3
	}
	return 1
}

func (g *goldberg) move24() int {
	step := g.getStep(24)
	if step > 4 {
		return 0
	}
	if step == 4 {
		return 2
	}
	if step == 0 {
		g.c.Delete("I24")
		xy := g.scl([]float64{347, 465, 361, 557, 271, 503, 272, 503, 342, 574, 259, 594, 259, 593, 362, 626,
			320, 737, 320, 740, 398, 691, 436, 738, 436, 739, 476, 679, 528, 701, 527, 702,
			494, 627, 548, 613, 548, 613, 480, 574, 577, 473, 577, 473, 474, 538, 445, 508,
			431, 441, 431, 440, 400, 502, 347, 465, 347, 465})
		g.c.CreatePolygon(xy, canvas.Tags("I24"), canvas.FillColor(colors["24b"]), canvas.OutlineColor(colors["24a"]), canvas.OutlineWidth(10), canvas.Smooth(true))
		cx, cy := g.centroid("I24")
		g.c.CreateText(cx, cy, canvas.TextOpt(g.message), canvas.Tags("I24", "I24t"),
			canvas.TextColor(colors["24c"]), canvas.FontOpt("Times 18 bold"))
		return 1
	}
	g.c.ItemConfigure("I24t", canvas.FontOpt(fmt.Sprintf("Times %d bold", 18+6*step)))
	g.c.Move("I24", 7.5, -33.75)
	cx, cy := g.centroid("I24")
	g.c.Scale("I24", cx, cy, 1.25, 1.25)
	return 1
}

func (g *goldberg) move25() int {
	step := g.getStep(25)
	if step == 0 {
		g.move25Start = time.Now()
		return 1
	}
	if time.Since(g.move25Start) < 5*time.Second {
		return 1
	}
	return 2
}

func (g *goldberg) move26() int {
	step := g.getStep(26)
	if step >= 3 {
		g.c.Delete("I24")
		g.c.Delete("I26")
		msgX := math.Round(338 * g.sf)
		msgY := math.Round(573 * g.sf)
		g.c.CreateText(msgX, msgY, canvas.AnchorOpt(option.AnchorS), canvas.Tags("I26"),
			canvas.TextColor(colors["26"]), canvas.TextOpt("click to continue"), canvas.FontOpt("Times 24 bold"))
		g.c.BindItem("I26", event.ButtonPressMask, func(_ *event.Event) {
			g.reset()
		})
		return 4
	}
	cx, cy := g.centroid("I24")
	g.c.Scale("I24", cx, cy, 0.8, 0.8)
	g.c.Move("I24", 0, 60)
	g.c.ItemConfigure("I24t", canvas.FontOpt(fmt.Sprintf("Times %d bold", 30-6*step)))
	return 1
}

// doButton handles control button clicks.
func (g *goldberg) doButton(what int) {
	switch what {
	case 0: // Start
		if g.mode == mDone {
			g.reset()
		}
		g.mode = mGo
	case 1: // Pause
		if g.pause {
			g.mode = mPause
		} else {
			g.mode = mGo
		}
	case 2: // Single step
		g.mode = mSStep
	case 3: // Reset
		g.reset()
	case 4: // Big step
		g.mode = mBStep
	}
}

func main() {
	app, err := takigo.NewApp(
		takigo.Title("Tk Goldberg (demonstration)"),
		takigo.IconName("goldberg"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	app.WmInfo().SetResizable(false, false)

	// BaseDimensions (CanX 675, CanY 540, ScrX/ScrY 750) times
	// overallFactor 0.75, in points.
	dim := func(base float64) int { return screenunit.Px(fmt.Sprintf("%gp", base*0.75)) }

	g := &goldberg{
		app:     app,
		sf:      float64(screenunit.ScalingPct()) / 100 * 0.75,
		mode:    mStart,
		speed:   5,
		message: "\nWelcome\nto\nTcl/Tk!",
		active:  []int{0},
		step:    make(map[int]int),
	}

	// DoDisplay.
	ctrl := ttk.NewFrame(app, "ctrl", ttk.FrameRelief(option.ReliefRidge),
		ttk.FrameBorderWidth(1), ttk.FramePadding(ttk.UniformPadding(screenunit.Px("3p"))))
	screen := frame.New(app, "screen", frame.BorderWidth(1), frame.Relief(option.ReliefRaised))
	pack.Pack(screen, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	c := canvas.New(screen, "c",
		canvas.Width(dim(675)), canvas.Height(dim(540)),
		canvas.Background(colors["bg"]),
		canvas.HighlightWidthOpt(0),
		canvas.ScrollRegion(0, 0, dim(750), dim(750)),
	)
	g.c = c
	pack.Pack(c, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	c.Win.ConfigureCallback = func() { c.YViewMoveTo(0.06) }

	// DoCtrlFrame: the widgets are gridded -in ctrl in Tk; here they are
	// its children.
	pause := widget.NewVariable(false)
	start := ttk.NewButton(ctrl, "start", ttk.ButtonText("Start"), ttk.ButtonCommand(func() { g.doButton(0) }))
	pauseCb := ttk.NewCheckbutton(ctrl, "pause", ttk.CheckbuttonText("Pause"), ttk.CheckbuttonVar(pause),
		ttk.CheckbuttonCommand(func() { g.pause = pause.Get(); g.doButton(1) }))
	step := ttk.NewButton(ctrl, "step", ttk.ButtonText("Single Step"), ttk.ButtonCommand(func() { g.doButton(2) }))
	bstep := ttk.NewButton(ctrl, "bstep", ttk.ButtonText("Big Step"), ttk.ButtonCommand(func() { g.doButton(4) }))
	reset := ttk.NewButton(ctrl, "reset", ttk.ButtonText("Reset"), ttk.ButtonCommand(func() { g.doButton(3) }))
	details := ttk.NewLabelframe(ctrl, "details", ttk.LabelframeText("Details"))
	message := ttk.NewLabelframe(ctrl, "message", ttk.LabelframeText("Message"))
	msgEntry := ttk.NewEntry(message, "e", ttk.EntryText(g.message), ttk.EntryJustify(option.JustifyCenter))
	speedLf := ttk.NewLabelframe(ctrl, "speed", ttk.LabelframeText("Speed: 0"))
	speedVar := widget.NewVariable(5.0)
	speedScale := ttk.NewScale(speedLf, "scale", ttk.ScaleOrient(ttk.Horizontal), ttk.ScaleFrom(1),
		ttk.ScaleTo(10), ttk.ScaleVariable(speedVar), ttk.ScaleCommand(func(v float64) { g.speed = int(v) }))
	about := ttk.NewButton(ctrl, "about", ttk.ButtonText("About"))
	row := 0
	gridRow := func(w geometry.Elementer, opts ...grid.GridOption) {
		grid.Grid(w, append([]grid.GridOption{grid.Row(row), grid.Column(0), grid.Sticky(grid.EW)}, opts...)...)
		row++
	}
	gridRow(start)
	grid.RowConfigure(ctrl, 1, grid.MinSize(screenunit.Px("3p")))
	row = 2
	gridRow(pauseCb)
	gridRow(step, grid.PadY("1.5p"))
	gridRow(bstep)
	gridRow(reset, grid.PadY("1.5p"))
	grid.RowConfigure(ctrl, 10, grid.MinSize(screenunit.Px("3p")))
	row = 11
	gridRow(details)
	grid.RowConfigure(ctrl, 11, grid.MinSize(screenunit.Px("3p")))
	grid.RowConfigure(ctrl, 50, grid.Weight(1))
	row = 98
	gridRow(message, grid.PadYPair(0, "3p"))
	grid.Grid(msgEntry, grid.Sticky(grid.NSEW))
	gridRow(speedLf, grid.PadYPair(0, "3p"))
	pack.Pack(speedScale, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	gridRow(about)
	gridRow(ttk.NewSeparator(ctrl, "sep"), grid.PadYPair("3p", "1.5p"))
	// "See Code / Dismiss buttons hack!": copies of the two buttons, stacked.
	gridRow(ttk.NewButton(ctrl, "b1", ttk.ButtonText("See Code"), ttk.ButtonImage(demohelper.Image("view")),
		ttk.ButtonCompound(widget.CompoundLeft), ttk.ButtonCommand(func() { demohelper.ShowCode(app) })),
		grid.PadYPair("1.5p", 0))
	gridRow(ttk.NewButton(ctrl, "b2", ttk.ButtonText("Dismiss"), ttk.ButtonImage(demohelper.Image("delete")),
		ttk.ButtonCompound(widget.CompoundLeft), ttk.ButtonCommand(app.Quit)),
		grid.PadYPair("1.5p", 0))

	show := ttk.NewButton(c, "show", ttk.ButtonText("▶"), ttk.ButtonWidth(2))
	show.Command = func() {
		if ctrl.Win.Flags&window.FlagMapped != 0 {
			pack.Forget(ctrl)
			show.Text = "▶"
		} else {
			pack.Pack(ctrl, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillBoth), pack.IPadY(5))
			show.Text = "◀"
		}
		show.Display()
	}
	place.Place(show, place.RelX(1), place.RelY(0), place.Anchor(option.AnchorNE))

	g.drawAll()
	c.YViewMoveTo(0.06)
	g.go_()

	// StartMessage / PlacedDialog.
	placedDialog := func(msg, fnt string) {
		mf := frame.New(c, "messframe", frame.Relief(option.ReliefRaised), frame.BorderWidth(screenunit.Px("3p")))
		lab := label.New(mf, "lab", label.FontOpt(fnt), label.WrapLength("3i"),
			label.JustifyOpt(option.JustifyLeft), label.Text(msg))
		but := ttk.NewButton(mf, "but", ttk.ButtonText("OK"), ttk.ButtonUnderline(0))
		but.Command = func() { mf.Destroy() }
		pack.Pack(lab, pack.PadX("10p"), pack.PadYPair("10p", "5p"))
		pack.Pack(but, pack.PadX("10p"), pack.PadYPair(0, "10p"))
		place.Place(mf, place.Anchor(option.AnchorCenter), place.RelX(0.5), place.RelY(0.5))
		app.After(0, func() { app.FocusManager().SetFocus(but.Window()) }) // focus $w.but
	}
	about.Command = func() {
		placedDialog("Tk Goldberg\nby Keith Vetter, March 2003\n(Reproduced by kind permission of the author)\n\n"+
			"\"Man will always find a difficult means to perform a simple task.\"\n - Rube Goldberg", "Helvetica 12 bold")
	}
	placedDialog("This is a demonstration of just how complex you can make your animations become. "+
		"Close this dialog and click the ball to start things moving!\n\n"+
		"\"Man will always find a difficult means to perform a simple task\"\n - Rube Goldberg", "Helvetica 12")

	app.Run()
}
