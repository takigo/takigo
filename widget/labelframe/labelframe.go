// Package labelframe implements a frame with a text label in the border gap.
// It ports the labelframe-specific parts of tk/generic/tkFrame.c.
package labelframe

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/geometry"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Labelframe is a container widget with a labeled border.
type Labelframe struct {
	widget.Base
	Text        string
	LabelAnchor option.Anchor // where the label sits on the border (default NW)
	LabelWidget widget.Widget // optional widget to use as the label instead of text

	// Disabled state — when true, title text is drawn in gray.
	Disabled bool

	textWidth   int
	textHeight  int
	labelWidth  int // effective label width (text or widget)
	labelHeight int // effective label height (text or widget)
}

// LabelframeOption configures a Labelframe.
type LabelframeOption func(*Labelframe)

// Text sets the label text.
func Text(s string) LabelframeOption {
	return func(lf *Labelframe) { lf.Text = s }
}

// Background sets the background color.
func Background[C color.Spec](name C) LabelframeOption {
	return func(lf *Labelframe) { lf.SetBackgroundColor(name) }
}

// Foreground sets the label text color.
func Foreground[C color.Spec](name C) LabelframeOption {
	return func(lf *Labelframe) { lf.SetForegroundColor(name) }
}

// BorderWidth sets the border width.
func BorderWidth[L screenunit.Length](w L) LabelframeOption {
	return func(lf *Labelframe) { lf.BorderWidth = screenunit.ToPixels(w) }
}

// Relief sets the border relief.
func Relief(r option.Relief) LabelframeOption {
	return func(lf *Labelframe) { lf.Relief = r }
}

// Width sets the requested width.
func Width(w int) LabelframeOption {
	return func(lf *Labelframe) { lf.Win.ReqWidth = w }
}

// Height sets the requested height.
func Height(h int) LabelframeOption {
	return func(lf *Labelframe) { lf.Win.ReqHeight = h }
}

// FontOpt sets the font.
func FontOpt[F font.Spec](name F) LabelframeOption {
	return func(lf *Labelframe) { lf.SetFont(name) }
}

// LabelAnchor sets where the label sits on the border.
func LabelAnchor(a option.Anchor) LabelframeOption {
	return func(lf *Labelframe) { lf.LabelAnchor = a }
}

// LabelWidgetOpt sets a widget to use as the label instead of text.
// The widget should be a child of the labelframe. When set, the widget
// is positioned on the border where the text label would normally go.
func LabelWidgetOpt(w widget.Widget) LabelframeOption {
	return func(lf *Labelframe) { lf.manageLabel(w) }
}

// PadX sets internal horizontal padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadX[L screenunit.Length](p L) LabelframeOption {
	return func(lf *Labelframe) { lf.PadX = screenunit.ToPixels(p) }
}

// PadY sets internal vertical padding.
// Accepts a number of pixels or a screenunit.Distance such as screenunit.Pt(3).
func PadY[L screenunit.Length](p L) LabelframeOption {
	return func(lf *Labelframe) { lf.PadY = screenunit.ToPixels(p) }
}

// New creates a new Labelframe widget.
func New(parent widget.Caregiver, name string, opts ...LabelframeOption) *Labelframe {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 200, 200)
	window.MakeWindowExist(w)

	lf := &Labelframe{
		LabelAnchor: option.AnchorNW,
	}
	widget.InitBase(&lf.Base, w, app)
	lf.SetDisplayProc(lf.display)
	w.Class = "Labelframe"

	// Labelframe defaults.
	lf.BorderWidth = 2
	lf.Relief = option.ReliefGroove

	for _, opt := range opts {
		opt(lf)
	}

	lf.computeTextSize()
	lf.updateInternalBorder()

	if lf.Background != nil {
		w.SetBackgroundPixel(lf.Background.Pixel)
	}

	// Bind events.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		lf.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			lf.Display()
			w.NotifyConfigure()
		}
	})

	return lf
}

// Tk's labelframe constants (tk/generic/tkFrame.c).
const (
	labelSpacing = 1 // LABELSPACING: space around the label text
	labelMargin  = 4 // LABELMARGIN: distance of the label from the corner
)

// hasLabel reports whether a text or widget label is shown.
func (lf *Labelframe) hasLabel() bool {
	return lf.LabelWidget != nil || (lf.Font != nil && lf.Text != "")
}

// labelSide classifies LabelAnchor into Tk's label sides.
func (lf *Labelframe) labelSide() (top, bottom, left, right bool) {
	switch lf.LabelAnchor {
	case option.AnchorS, option.AnchorSW, option.AnchorSE:
		return false, true, false, false
	case option.AnchorW:
		return false, false, true, false
	case option.AnchorE:
		return false, false, false, true
	default:
		return true, false, false, false
	}
}

func (lf *Labelframe) computeTextSize() {
	lf.textWidth, lf.textHeight = 0, 0
	lf.labelWidth, lf.labelHeight = 0, 0
	if lf.Font != nil && lf.Text != "" {
		lf.textWidth = lf.Font.MeasureString(lf.Text)
		lf.textHeight = lf.Font.Metrics().Linespace()
	}
	switch {
	case lf.LabelWidget != nil:
		lw := lf.LabelWidget.Window()
		lf.labelWidth, lf.labelHeight = lw.ReqWidth, lw.ReqHeight
	case lf.hasLabel():
		lf.labelWidth = lf.textWidth + 2*labelSpacing
		lf.labelHeight = lf.textHeight + 2*labelSpacing
	default:
		return
	}
	// The label is at least as big as the border.
	top, bottom, _, _ := lf.labelSide()
	if top || bottom {
		lf.labelHeight = max(lf.labelHeight, lf.BorderWidth)
	} else {
		lf.labelWidth = max(lf.labelWidth, lf.BorderWidth)
	}
}

// updateInternalBorder ports the labelframe part of FrameWorldChanged and
// ComputeFrameGeometry: children go inside border+highlight+pad, the label
// side grows by the label size, and the frame requests room for the label.
func (lf *Labelframe) updateInternalBorder() {
	w := lf.Win
	bw, hl := lf.BorderWidth, lf.HighlightWidth
	left, right := bw+hl+lf.PadX, bw+hl+lf.PadX
	top, bottom := bw+hl+lf.PadY, bw+hl+lf.PadY
	w.MinReqWidth, w.MinReqHeight = 0, 0
	if lf.hasLabel() {
		onTop, onBottom, onLeft, _ := lf.labelSide()
		switch {
		case onTop:
			top += lf.labelHeight - bw
		case onBottom:
			bottom += lf.labelHeight - bw
		case onLeft:
			left += lf.labelWidth - bw
		default:
			right += lf.labelWidth - bw
		}
		padding := hl
		if bw > 0 {
			padding += bw + labelMargin
		}
		padding *= 2
		if onTop || onBottom {
			w.MinReqWidth = lf.labelWidth + padding
			w.MinReqHeight = lf.labelHeight + bw + hl
		} else {
			w.MinReqHeight = lf.labelHeight + padding
			w.MinReqWidth = lf.labelWidth + bw + hl
		}
	}
	w.InternalBorderLeft, w.InternalBorderRight = left, right
	w.InternalBorderTop, w.InternalBorderBottom = top, bottom
}

// labelBox returns where a label of size lw×lh goes (LabelframeLayout in
// tkFrame.c); it is used both for the clipped label box and the text.
func (lf *Labelframe) labelBox(lw, lh int) (x, y int) {
	W, H := lf.Win.Width, lf.Win.Height
	bw, hl := lf.BorderWidth, lf.HighlightWidth
	top, bottom, _, right := lf.labelSide()
	otherW, otherH := W-lw, H-lh
	padding := hl
	switch {
	case right:
		x = otherW - padding
	case top:
		y = padding
	case bottom:
		y = otherH - padding
	default:
		x = padding
	}
	if bw > 0 {
		padding += bw + labelMargin
	}
	switch lf.LabelAnchor {
	case option.AnchorSW, option.AnchorNW:
		x = padding
	case option.AnchorN, option.AnchorS:
		x = otherW / 2
	case option.AnchorNE, option.AnchorSE:
		x = otherW - padding
	case option.AnchorE, option.AnchorW:
		y = otherH / 2
	default:
		x = padding
	}
	return x, y
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (lf *Labelframe) Display() {
	lf.EventuallyRedraw()
}

// display draws the labelframe.
func (lf *Labelframe) display() {
	if lf.Destroyed() {
		return
	}
	w := lf.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	bgPixel := uint64(0)
	if lf.Background != nil {
		bgPixel = lf.Background.Pixel
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	border := lf.Border
	if border == nil {
		border = draw.NewBorderFromPixel(bgPixel)
	}
	bw, hl := lf.BorderWidth, lf.HighlightWidth

	if !lf.hasLabel() {
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, bw, lf.Relief)
		return
	}

	// Label box, clamped to the room left by the border (LabelframeLayout).
	padding := hl
	if bw > 0 {
		padding += bw + labelMargin
	}
	padding *= 2
	top, bottom, left, _ := lf.labelSide()
	boxW, boxH := lf.labelWidth, lf.labelHeight
	if top || bottom {
		boxW = min(boxW, max(1, w.Width-padding))
	} else {
		boxH = min(boxH, max(1, w.Height-padding))
	}
	boxX, boxY := lf.labelBox(boxW, boxH)

	// The border runs through the middle of the label (DisplayFrame).
	bdX1, bdY1 := hl, hl
	bdX2, bdY2 := w.Width-hl, w.Height-hl
	switch {
	case top:
		bdY1 += (boxH - bw + 1) / 2
	case bottom:
		bdY2 -= (boxH - bw) / 2
	case left:
		bdX1 += (boxW - bw) / 2
	default:
		bdX2 -= (boxW - bw) / 2
	}
	draw.Draw3DRectangle(d, w.Drawable(), gc, border,
		bdX1, bdY1, bdX2-bdX1, bdY2-bdY1, bw, lf.Relief)

	if lf.LabelWidget == nil {
		d.SetForeground(gc, bgPixel)
		d.FillRectangle(w.Drawable(), gc, boxX, boxY, uint(boxW), uint(boxH))
		if lf.Foreground != nil {
			textX, textY := lf.labelBox(lf.labelWidth, lf.labelHeight)
			baseline := textY + labelSpacing + lf.Font.Metrics().Ascent
			if df, ok := lf.Font.(platform.DrawableFont); ok {
				fgPixel := lf.Foreground.Pixel
				fgR, fgG, fgB := lf.Foreground.Red, lf.Foreground.Green, lf.Foreground.Blue
				if lf.Disabled {
					fgPixel, fgR, fgG, fgB = widget.DisabledColor(lf.App)
				}
				df.DrawString(w.Drawable(), textX+labelSpacing, baseline, lf.Text,
					fgPixel, fgR, fgG, fgB)
			}
		}
	} else {
		lf.placeLabelWidget(boxX, boxY, boxW, boxH)
	}

}

// placeLabelWidget puts the -labelwidget window on the label box, like
// Tk_MaintainGeometry does for a label window that need not be a child.
func (lf *Labelframe) placeLabelWidget(x, y, width, height int) {
	lw := lf.LabelWidget.Window()
	for anc := lf.Win; anc != nil && anc != lw.Parent; anc = anc.Parent {
		if anc.Parent == nil {
			return
		}
		x += anc.X
		y += anc.Y
	}
	lw.X, lw.Y, lw.Width, lw.Height = x, y, width, height
	if lw.PlatformID == platform.WindowID(0) {
		return
	}
	d := lf.Win.Display.Server
	d.MoveResizeWindow(lw.PlatformID, x, y, uint(width), uint(height))
	d.RaiseWindow(lw.PlatformID)
	if lw.Flags&window.FlagMapped == 0 {
		d.MapWindow(lw.PlatformID)
		window.MarkMapped(lw)
	}
}

// Window returns the underlying window.
func (lf *Labelframe) Window() *window.Window {
	return lf.Win
}

// Configure applies options.
func (lf *Labelframe) Configure(opts ...LabelframeOption) error {
	return widget.Configure(lf, opts, func() {
		lf.computeTextSize()
		lf.updateInternalBorder()
	})
}

// SetLabelWidget sets a widget as the label after construction.
// This is needed when the label widget is a child of the labelframe itself.
func (lf *Labelframe) SetLabelWidget(w widget.Widget) {
	lf.manageLabel(w)
	lf.labelChanged()
}

// manageLabel makes w the label widget and the labelframe its geometry
// manager, as tkFrame.c does with Tk_ManageGeometry, so the label's size
// requests re-lay the labelframe out and its destruction drops it.
func (lf *Labelframe) manageLabel(w widget.Widget) {
	lf.LabelWidget = w
	if w != nil {
		geometry.ManageGeometry(w.Window(), &labelGeomMgr{lf})
	}
}

// labelChanged recomputes the label space after the label changed.
func (lf *Labelframe) labelChanged() {
	lf.computeTextSize()
	lf.updateInternalBorder()
	lf.Display()
	// Notify geometry manager that internal borders changed.
	lf.Win.NotifyConfigure()
}

// labelGeomMgr manages a labelframe's -labelwidget (FrameRequestProc,
// FrameLostContentProc).
type labelGeomMgr struct{ lf *Labelframe }

func (m *labelGeomMgr) Name() string { return "labelframe" }

func (m *labelGeomMgr) RequestProc(content *window.Window) {
	if lw := m.lf.LabelWidget; lw != nil && lw.Window() == content && !m.lf.Destroyed() {
		m.lf.labelChanged()
	}
}

func (m *labelGeomMgr) LostContentProc(content *window.Window) {
	if lw := m.lf.LabelWidget; lw != nil && lw.Window() == content {
		m.lf.LabelWidget = nil
		if !m.lf.Destroyed() {
			m.lf.labelChanged()
		}
	}
}

// Destroy cleans up the labelframe.
func (lf *Labelframe) Destroy() {
	if lf.Destroyed() {
		return
	}
	lf.MarkDestroyed()
	window.DestroyWindow(lf.Win)
}
