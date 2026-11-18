// Ports the cascade arrow of DrawMenuEntryArrow (tk/unix/tkUnixMenu.c) and
// the CHECK_MENU/RADIO_MENU marks of TkpDrawCheckIndicator
// (tk/unix/tkUnixButton.c).

package menu

import (
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
)

const (
	cascadeArrowWidth  = 8
	cascadeArrowHeight = 10
	// indicatorMargin is the left margin reserved for check/radio marks.
	indicatorMargin = 20
)

// decorationScale stands in for TkScalingLevel: the marks grow with the
// menu font, as they grow with the screen scaling in Tk.
func decorationScale(fm font.Metrics) float64 {
	return max(1.0, float64(fm.Linespace())/17)
}

// drawCascadeArrow draws Tk's raised 3D triangle, its right edge at right,
// centered vertically in the entry at y with the given height.
func drawCascadeArrow(d platform.DisplayServer, dr platform.DrawableID, gc platform.GCID,
	bgPixel uint64, fm font.Metrics, right, y, height int, sunken bool) {
	s := decorationScale(fm)
	aw := int(cascadeArrowWidth * s)
	ah := int(cascadeArrowHeight * s)
	x0 := right - aw
	y0 := y + (height-ah)/2
	pts := []platform.Point{
		{X: int32(x0), Y: int32(y0)},
		{X: int32(x0), Y: int32(y0 + ah)},
		{X: int32(x0 + aw), Y: int32(y0 + ah/2)},
	}
	relief := option.ReliefRaised
	if sunken {
		relief = option.ReliefSunken
	}
	draw.Fill3DPolygon(d, dr, gc, draw.NewBorderFromPixel(bgPixel), pts, 2, relief)
}

// drawIndicator draws a check mark or a radio dot centered at (cx, cy).
// Callers put cx left of the margin centre so the mark clears the label.
func drawIndicator(d platform.DisplayServer, dr platform.DrawableID, gc platform.GCID,
	pixel uint64, fm font.Metrics, cx, cy int, radio bool) {
	s := decorationScale(fm)
	d.SetForeground(gc, pixel)
	if radio {
		r := int(3 * s)
		d.FillArc(dr, gc, cx-r, cy-r, uint(2*r), uint(2*r), 0, 360*64)
		return
	}
	p := func(dx, dy float64) platform.Point {
		return platform.Point{X: int32(cx + int(dx*s)), Y: int32(cy + int(dy*s))}
	}
	lw := max(2, int(2*s))
	d.SetLineAttributes(gc, uint(lw), platform.LineSolid, platform.CapButt, platform.JoinMiter)
	d.DrawLines(dr, gc, []platform.Point{p(-4, 0), p(-1, 3), p(3, -3)}, platform.CoordModeOrigin)
	d.SetLineAttributes(gc, 0, platform.LineSolid, platform.CapButt, platform.JoinMiter)
}
