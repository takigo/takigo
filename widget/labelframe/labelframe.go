// Package labelframe implements a frame with a text label in the border gap.
// It ports the labelframe-specific parts of tk/generic/tkFrame.c.
package labelframe

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Labelframe is a container widget with a labeled border.
type Labelframe struct {
	widget.Base
	Text        string
	LabelAnchor option.Anchor // where the label sits on the border (default NW)

	textWidth  int
	textHeight int
}

// LabelframeOption configures a Labelframe.
type LabelframeOption func(*Labelframe)

// Text sets the label text.
func Text(s string) LabelframeOption {
	return func(lf *Labelframe) { lf.Text = s }
}

// Background sets the background color.
func Background(name string) LabelframeOption {
	return func(lf *Labelframe) {
		col, err := lf.App.ColorCache().Get(name)
		if err == nil {
			lf.Background = col
			lf.UpdateBorder()
		}
	}
}

// BorderWidth sets the border width.
func BorderWidth(w int) LabelframeOption {
	return func(lf *Labelframe) { lf.BorderWidth = w }
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
func FontOpt(name string) LabelframeOption {
	return func(lf *Labelframe) {
		f, err := lf.App.FontRegistry().Get(name)
		if err == nil {
			lf.Font = f
		}
	}
}

// LabelAnchor sets where the label sits on the border.
func LabelAnchor(a option.Anchor) LabelframeOption {
	return func(lf *Labelframe) { lf.LabelAnchor = a }
}

// PadX sets internal horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("1.5p", "2m", etc.).
func PadX(p any) LabelframeOption {
	return func(lf *Labelframe) { lf.PadX = screenunit.Px(p) }
}

// PadY sets internal vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("1.5p", "2m", etc.).
func PadY(p any) LabelframeOption {
	return func(lf *Labelframe) { lf.PadY = screenunit.Px(p) }
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

	// Labelframe defaults.
	lf.BorderWidth = 2
	lf.Relief = option.ReliefGroove

	for _, opt := range opts {
		opt(lf)
	}

	lf.computeTextSize()
	lf.updateInternalBorder()

	if lf.Background != nil {
		w.BackgroundPixel = lf.Background.Pixel
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
			if w.ConfigureCallback != nil {
				w.ConfigureCallback()
			}
		}
	})

	return lf
}

func (lf *Labelframe) computeTextSize() {
	if lf.Font != nil && lf.Text != "" {
		lf.textWidth = lf.Font.MeasureString(lf.Text)
		m := lf.Font.Metrics()
		lf.textHeight = m.Linespace()
	} else {
		lf.textWidth = 0
		lf.textHeight = 0
	}
}

func (lf *Labelframe) updateInternalBorder() {
	w := lf.Win
	bw := lf.BorderWidth
	// Tk C: bWidthTop = borderWidth + (labelReqHeight - borderWidth) = labelReqHeight
	// The full label height replaces the top border width, since the label
	// sits centered on the border and the content area starts below it.
	topBorder := bw
	if lf.Text != "" && lf.textHeight > 0 {
		topBorder = lf.textHeight
	}
	// Tk C: padX/padY are added to all four internal borders.
	w.InternalBorderLeft = bw + lf.PadX
	w.InternalBorderRight = bw + lf.PadX
	w.InternalBorderTop = topBorder + lf.PadY
	w.InternalBorderBottom = bw + lf.PadY
}

// Display draws the labelframe.
func (lf *Labelframe) Display() {
	if lf.Destroyed {
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

	bw := lf.BorderWidth
	hasLabel := lf.Font != nil && lf.Text != "" && lf.textHeight > 0

	if bw > 0 && lf.Relief != option.ReliefFlat {
		border := lf.Border
		if border == nil {
			border = draw.NewBorderFromPixel(bgPixel)
		}

		if hasLabel {
			lf.drawBorderWithGap(d, gc, border, bw)
		} else {
			draw.Draw3DRectangle(d, w.Drawable(), gc, border,
				0, 0, w.Width, w.Height, bw, lf.Relief)
		}
	}

	// Draw label text.
	if hasLabel && lf.Foreground != nil {
		labelX := lf.labelX()
		labelY := 0 // label top is at y=0
		m := lf.Font.Metrics()
		baseline := labelY + m.Ascent
		if df, ok := lf.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), labelX, baseline, lf.Text,
				lf.Foreground.Pixel, lf.Foreground.Red, lf.Foreground.Green, lf.Foreground.Blue)
		}
	}

	d.Flush()
}

// labelX returns the x position for the label text.
func (lf *Labelframe) labelX() int {
	bw := lf.BorderWidth
	gap := 8 // gap from border edge to label
	switch lf.LabelAnchor {
	case option.AnchorNW, option.AnchorW, option.AnchorSW:
		return bw + gap
	case option.AnchorNE, option.AnchorE, option.AnchorSE:
		return lf.Win.Width - bw - gap - lf.textWidth
	default: // center
		return (lf.Win.Width - lf.textWidth) / 2
	}
}

// drawBorderWithGap draws the 3D border with a gap in the top for the label.
func (lf *Labelframe) drawBorderWithGap(d platform.DisplayServer, gc platform.GCID, border *draw.Border, bw int) {
	w := lf.Win
	labelX := lf.labelX()
	gapLeft := labelX - 4
	gapRight := labelX + lf.textWidth + 4

	// The border frame is offset down by half the label height.
	frameY := lf.textHeight / 2
	frameH := w.Height - frameY

	// Draw left, right, bottom borders normally.
	// Left border.
	for i := 0; i < bw; i++ {
		lp, dp := borderPixels(border, lf.Relief, i, bw)
		d.SetForeground(gc, lp)
		d.DrawLine(w.Drawable(), gc, i, frameY+i, i, frameY+frameH-1-i)
		_ = dp
	}
	// Right border.
	for i := 0; i < bw; i++ {
		_, dp := borderPixels(border, lf.Relief, i, bw)
		d.SetForeground(gc, dp)
		d.DrawLine(w.Drawable(), gc, w.Width-1-i, frameY+i, w.Width-1-i, frameY+frameH-1-i)
	}
	// Bottom border.
	for i := 0; i < bw; i++ {
		_, dp := borderPixels(border, lf.Relief, i, bw)
		d.SetForeground(gc, dp)
		d.DrawLine(w.Drawable(), gc, i, frameY+frameH-1-i, w.Width-1-i, frameY+frameH-1-i)
	}
	// Top border — split around gap.
	for i := 0; i < bw; i++ {
		lp, _ := borderPixels(border, lf.Relief, i, bw)
		d.SetForeground(gc, lp)
		// Left part of top.
		if gapLeft > i {
			d.DrawLine(w.Drawable(), gc, i, frameY+i, gapLeft, frameY+i)
		}
		// Right part of top.
		if gapRight < w.Width-1-i {
			d.DrawLine(w.Drawable(), gc, gapRight, frameY+i, w.Width-1-i, frameY+i)
		}
	}
}

// borderPixels returns the light and dark pixels for the given border layer.
func borderPixels(border *draw.Border, relief option.Relief, layer, bw int) (light, dark uint64) {
	switch relief {
	case option.ReliefRaised:
		return border.LightPixel, border.DarkPixel
	case option.ReliefSunken:
		return border.DarkPixel, border.LightPixel
	case option.ReliefGroove:
		if layer < bw/2 {
			return border.DarkPixel, border.LightPixel
		}
		return border.LightPixel, border.DarkPixel
	case option.ReliefRidge:
		if layer < bw/2 {
			return border.LightPixel, border.DarkPixel
		}
		return border.DarkPixel, border.LightPixel
	default:
		return border.BgPixel, border.BgPixel
	}
}

// Window returns the underlying window.
func (lf *Labelframe) Window() *window.Window {
	return lf.Win
}

// Configure applies options.
func (lf *Labelframe) Configure(opts ...option.Option) {
	option.Apply(lf, opts)
	lf.UpdateBorder()
	lf.computeTextSize()
	lf.updateInternalBorder()
	if lf.Background != nil {
		lf.Win.BackgroundPixel = lf.Background.Pixel
	}
	lf.Display()
}

// Destroy cleans up the labelframe.
func (lf *Labelframe) Destroy() {
	if lf.Destroyed {
		return
	}
	lf.Destroyed = true
	window.DestroyWindow(lf.Win)
}
