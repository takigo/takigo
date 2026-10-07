// Package frame implements the frame widget, a simple container for
// grouping other widgets. It ports tk/generic/tkFrame.c.
package frame

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Frame is a container widget that provides a background and optional
// border for grouping child widgets.
type Frame struct {
	widget.Base
}

// FrameOption configures a Frame.
type FrameOption func(*Frame)

// Background sets the background color by name.
func Background[C color.Spec](name C) FrameOption {
	return func(f *Frame) { f.SetBackgroundColor(name) }
}

// BorderWidth sets the border width in pixels.
func BorderWidth[L screenunit.Length](w L) FrameOption {
	return func(f *Frame) { f.BorderWidth = screenunit.ToPixels(w) }
}

// Relief sets the border relief.
func Relief(r option.Relief) FrameOption {
	return func(f *Frame) { f.Relief = r }
}

// Width sets the requested width.
// HighlightThickness sets -highlightthickness.
func HighlightThickness[L screenunit.Length](n L) FrameOption {
	return func(f *Frame) { f.HighlightWidth = screenunit.ToPixels(n) }
}

func Width(w int) FrameOption {
	return func(f *Frame) { f.Win.ReqWidth = w }
}

// Height sets the requested height.
func Height(h int) FrameOption {
	return func(f *Frame) { f.Win.ReqHeight = h }
}

// New creates a new Frame widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...FrameOption) *Frame {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 200, 200)
	window.MakeWindowExist(w)

	f := &Frame{}
	widget.InitBase(&f.Base, w, app)
	f.SetDisplayProc(f.display)
	w.Class = "Frame"

	// Frame-specific defaults. Like a Tk frame without -width/-height, it
	// requests no size of its own (Tk windows start at 1x1).
	w.ReqWidth, w.ReqHeight = 1, 1
	f.BorderWidth = 0
	f.Relief = option.ReliefFlat

	for _, opt := range opts {
		opt(f)
	}
	f.updateInternalBorder()

	// Update X window background.
	if f.Background != nil {
		w.SetBackgroundPixel(f.Background.Pixel)
	}

	// Bind events.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		f.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			f.Display()
			w.NotifyConfigure()
		}
	})

	return f
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (f *Frame) Display() {
	f.EventuallyRedraw()
}

// display draws the frame.
func (f *Frame) display() {
	if f.Destroyed() {
		return
	}
	f.DrawBackground()
}

// Configure applies options to the frame.
func (f *Frame) Configure(opts ...FrameOption) error {
	return widget.Configure(f, opts, f.updateInternalBorder)
}

// Destroy cleans up the frame.
func (f *Frame) Destroy() {
	if f.Destroyed() {
		return
	}
	f.MarkDestroyed()
	window.DestroyWindow(f.Win)
}

// Window returns the underlying window.
func (f *Frame) Window() *window.Window {
	return f.Win
}

// updateInternalBorder keeps children inside the border and highlight ring,
// as FrameWorldChanged (tk/generic/tkFrame.c) does via Tk_SetInternalBorderEx.
func (f *Frame) updateInternalBorder() {
	b := f.BorderWidth + f.HighlightWidth
	f.SetInternalBorder(b, b, b, b)
}

// SetInternalBorder sets the internal border for child layout.
func (f *Frame) SetInternalBorder(left, right, top, bottom int) {
	w := f.Win
	w.InternalBorderLeft = left
	w.InternalBorderRight = right
	w.InternalBorderTop = top
	w.InternalBorderBottom = bottom
}

// DrawBackground fills the frame background and draws the border.
func (f *Frame) DrawBackground() {
	w := f.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}
	// DisplayFrame (tkFrame.c): border inside the highlight ring, ring last.
	d := w.Display.Server
	if f.Background != nil {
		d.SetForeground(w.GC, f.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), w.GC, 0, 0, uint(w.Width), uint(w.Height))
	hl := f.HighlightWidth
	if f.Border != nil && f.BorderWidth > 0 && f.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, w.Drawable(), w.GC, f.Border,
			hl, hl, w.Width-2*hl, w.Height-2*hl, f.BorderWidth, f.Relief)
	}
	f.DrawHighlightBorder(false, 0)
	d.Flush()
}
