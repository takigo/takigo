package canvas_test

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"strconv"
	"testing"
	"time"

	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
)

// TestDamageRedrawMatchesFullRedraw applies random item changes, lets the
// idle redraw repaint only the damaged area, and checks the window is
// pixel-identical to a full repaint.
func TestDamageRedrawMatchesFullRedraw(t *testing.T) {
	app := testutil.NewTestApp(t)
	c := canvas.New(app, "c", canvas.Width(300), canvas.Height(220),
		canvas.Background("white"), canvas.BorderWidthOpt(2), canvas.HighlightWidthOpt(1))
	pack.Pack(geometry.Group{c})

	rng := rand.New(rand.NewPCG(1, 2))
	colors := []string{"red", "blue", "green", "black", "orange", "purple"}
	pick := func() string { return colors[rng.IntN(len(colors))] }
	coord := func() float64 { return float64(rng.IntN(320) - 10) }
	var ids []canvas.ItemID
	create := func() {
		var id canvas.ItemID
		switch rng.IntN(6) {
		case 0:
			id = c.CreateRectangle(coord(), coord(), coord(), coord(),
				canvas.FillColor(pick()), canvas.OutlineColor(pick()), canvas.OutlineWidth(1+rng.IntN(6)))
		case 1:
			id = c.CreateOval(coord(), coord(), coord(), coord(),
				canvas.FillColor(pick()), canvas.OutlineWidth(1+rng.IntN(4)))
		case 2:
			id = c.CreateLine([]float64{coord(), coord(), coord(), coord(), coord(), coord()},
				canvas.FillColor(pick()), canvas.WidthOpt(1+rng.IntN(8)), canvas.Arrow(canvas.ArrowMode(rng.IntN(4))))
		case 3:
			id = c.CreatePolygon([]float64{coord(), coord(), coord(), coord(), coord(), coord()},
				canvas.FillColor(pick()), canvas.OutlineColor(pick()), canvas.OutlineWidth(1+rng.IntN(4)))
		case 4:
			id = c.CreateArc(coord(), coord(), coord(), coord(),
				canvas.StartAngle(float64(rng.IntN(360))), canvas.Extent(float64(rng.IntN(300)+30)),
				canvas.FillColor(pick()), canvas.OutlineWidth(1+rng.IntN(4)))
		case 5:
			id = c.CreateText(coord(), coord(), canvas.TextOpt("damage "+strconv.Itoa(rng.IntN(1000))),
				canvas.TextColor(pick()), canvas.TextAngle(float64(rng.IntN(4)*30)))
		}
		ids = append(ids, id)
	}
	for range 25 {
		create()
	}
	randomID := func() canvas.ItemID { return ids[rng.IntN(len(ids))] }

	mutations := []struct {
		name string
		do   func()
	}{
		{"move", func() { c.Move(randomID(), float64(rng.IntN(81)-40), float64(rng.IntN(81)-40)) }},
		{"scale", func() { c.Scale(randomID(), 150, 110, 0.5+rng.Float64(), 0.5+rng.Float64()) }},
		{"fill", func() { _ = c.ItemConfigure(randomID(), canvas.FillColor(pick())) }},
		{"width", func() { _ = c.ItemConfigure(randomID(), canvas.OutlineWidth(1+rng.IntN(8))) }},
		{"hide", func() { _ = c.ItemConfigure(randomID(), canvas.StateOpt(canvas.ItemStateHidden)) }},
		{"show", func() { _ = c.ItemConfigure(randomID(), canvas.StateOpt(canvas.ItemStateNormal)) }},
		{"raise", func() { c.Raise(randomID()) }},
		{"lower", func() { c.Lower(randomID()) }},
		{"coords", func() {
			id := randomID()
			if n := len(c.ItemCoords(id)); n > 0 {
				cs := make([]float64, n)
				for i := range cs {
					cs[i] = coord()
				}
				_ = c.SetItemCoords(id, cs)
			}
		}},
		{"delete", func() {
			i := rng.IntN(len(ids))
			c.Delete(ids[i])
			ids = append(ids[:i], ids[i+1:]...)
		}},
		{"create", create},
		{"text", func() {
			id := randomID()
			c.Focus(id)
			switch rng.IntN(3) {
			case 0:
				c.Insert(id, "end", "xyz")
			case 1:
				c.Dchars(id, "0", "2")
			case 2:
				c.ICursor(id, strconv.Itoa(rng.IntN(8)))
			}
		}},
		{"unfocus", func() { c.Focus("") }},
	}

	d := app.Server()
	win := c.Window()
	grab := func() []byte {
		d.Sync(false)
		return d.GetImageRGBA(platform.WindowDrawable(win.PlatformID), 0, 0, win.Width, win.Height)
	}

	var failure string
	app.After(100*time.Millisecond, func() {
		defer app.Quit()
		c.Display()
		if grab() == nil {
			failure = "skip: window contents cannot be read back"
			return
		}
		for step := range 300 {
			m := mutations[rng.IntN(len(mutations))]
			m.do()
			app.UpdateIdleTasks()
			partial := grab()
			c.Display()
			full := grab()
			if !bytes.Equal(partial, full) {
				failure = fmt.Sprintf("step %d (%s): damage repaint differs from a full repaint at pixel %d",
					step, m.name, firstDiff(partial, full)/4)
				return
			}
		}
	})
	app.MainLoop()
	if len(failure) > 5 && failure[:5] == "skip:" {
		t.Skip(failure)
	}
	if failure != "" {
		t.Fatal(failure)
	}
}

func firstDiff(a, b []byte) int {
	for i := range min(len(a), len(b)) {
		if a[i] != b[i] {
			return i
		}
	}
	return min(len(a), len(b))
}
