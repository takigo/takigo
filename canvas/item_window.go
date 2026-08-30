package canvas

import (
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// WindowItem embeds a child window at a fixed canvas position.
// The window is positioned as an overlay; canvas drawing is not affected.
type WindowItem struct {
	ItemBase
	x, y   float64
	win    *window.Window
	anchor string // "nw" (default) or other anchor strings (currently only nw supported)
}

func newWindowItem(x, y float64, win *window.Window, c *Canvas) *WindowItem {
	item := &WindowItem{x: x, y: y, win: win}
	item.ItemBase.canvas = c
	item.updateBBox()
	return item
}

func (wi *WindowItem) base() *ItemBase { return &wi.ItemBase }
func (wi *WindowItem) Type() string    { return "window" }

func (wi *WindowItem) BBox() (x1, y1, x2, y2 int) {
	return wi.X1, wi.Y1, wi.X2, wi.Y2
}

func (wi *WindowItem) Coords() []float64 {
	return []float64{wi.x, wi.y}
}

func (wi *WindowItem) SetCoords(coords []float64) error {
	if len(coords) >= 2 {
		wi.x = coords[0]
		wi.y = coords[1]
		wi.updateBBox()
	}
	return nil
}

func (wi *WindowItem) Configure(opts []ItemOption) error {
	c := wi.canvas
	for _, opt := range opts {
		if err := opt(c, wi); err != nil {
			return err
		}
	}
	wi.updateBBox()
	return nil
}

func (wi *WindowItem) updateBBox() {
	wi.X1 = int(wi.x)
	wi.Y1 = int(wi.y)
	wi.X2 = int(wi.x)
	wi.Y2 = int(wi.y)
	if wi.win != nil {
		w := wi.win.ReqWidth
		h := wi.win.ReqHeight
		if w == 0 {
			w = wi.win.Width
		}
		if h == 0 {
			h = wi.win.Height
		}
		wi.X2 = wi.X1 + w
		wi.Y2 = wi.Y1 + h
	}
}

// Display is a no-op: window positioning is done in Canvas.positionWindowItems.
func (wi *WindowItem) Display(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
	clipX, clipY, clipW, clipH, originX, originY int) {
}

func (wi *WindowItem) PointDistance(x, y float64) float64 {
	return rectPointDistance(x, y, float64(wi.X1), float64(wi.Y1), float64(wi.X2), float64(wi.Y2))
}

func (wi *WindowItem) AreaOverlap(ax1, ay1, ax2, ay2 float64) int {
	x1, y1, x2, y2 := float64(wi.X1), float64(wi.Y1), float64(wi.X2), float64(wi.Y2)
	if ax2 < x1 || ax1 > x2 || ay2 < y1 || ay1 > y2 {
		return -1
	}
	if ax1 <= x1 && ax2 >= x2 && ay1 <= y1 && ay2 >= y2 {
		return 1
	}
	return 0
}

func (wi *WindowItem) Scale(ox, oy, sx, sy float64) {
	wi.x = ox + (wi.x-ox)*sx
	wi.y = oy + (wi.y-oy)*sy
	wi.updateBBox()
}

func (wi *WindowItem) Translate(dx, dy float64) {
	wi.x += dx
	wi.y += dy
	wi.updateBBox()
}

func (wi *WindowItem) Delete(d platform.DisplayServer) {
	if wi.win != nil {
		d.UnmapWindow(wi.win.PlatformID)
	}
}

// Postscript emits a placeholder rectangle outline at the window item's
// bounding box. The embedded widget's contents are not captured — this
// matches the simplified scope agreed for the first port.
//
// Tk's full WinItemToPostscript tries `$win postscript` first and otherwise
// emits a pixmap dump; we keep takigo self-contained with the rectangle
// stub.
func (wi *WindowItem) Postscript(ps *PSContext) error {
	if wi.State() == ItemStateHidden {
		return nil
	}
	if ps.Prepass {
		return nil
	}
	if wi.X2 <= wi.X1 || wi.Y2 <= wi.Y1 {
		return nil
	}
	ps.Path([]float64{
		float64(wi.X1), float64(wi.Y1),
		float64(wi.X2), float64(wi.Y1),
		float64(wi.X2), float64(wi.Y2),
		float64(wi.X1), float64(wi.Y2),
	})
	// Light grey 1px outline so the placeholder is visible but unobtrusive.
	ps.writef("0.75 setgray %.15g setlinewidth\n", 1.0)
	ps.write("[] 0 setdash\nstroke newpath\n")
	ps.write("setgray\n")
	return nil
}
