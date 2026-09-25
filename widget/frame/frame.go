// Package frame implements the frame widget, a simple container for
// grouping other widgets. It ports tk/generic/tkFrame.c.
package frame

import (
	"log"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/place"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Frame is a container widget that provides a background and optional
// border for grouping child widgets.
type Frame struct {
	widget.Base
}

// FrameOption configures a Frame.
type FrameOption func(*Frame)

// Background sets the background color by name.
func Background(name string) FrameOption {
	return func(f *Frame) {
		col, err := f.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("frame: failed to get color %q: %v", name, err)
			return
		}
		f.Background = col
		f.UpdateBorder()
	}
}

// BorderWidth sets the border width in pixels.
func BorderWidth(w int) FrameOption {
	return func(f *Frame) { f.BorderWidth = w }
}

// Relief sets the border relief.
func Relief(r option.Relief) FrameOption {
	return func(f *Frame) { f.Relief = r }
}

// Width sets the requested width.
func Width(w int) FrameOption {
	return func(f *Frame) { f.Win.ReqWidth = w }
}

// Height sets the requested height.
func Height(h int) FrameOption {
	return func(f *Frame) { f.Win.ReqHeight = h }
}

// --- Ttk-compatible aliases (prefix with Frame) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.FramePadding, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// FrameBackground is an alias for Background.
var FrameBackground = Background

// FrameBorderWidth is an alias for BorderWidth.
var FrameBorderWidth = BorderWidth

// FrameRelief is an alias for Relief.
var FrameRelief = Relief

// FrameWidth is an alias for Width.
var FrameWidth = Width

// FrameHeight is an alias for Height.
var FrameHeight = Height

// New creates a new Frame widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...FrameOption) *Frame {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 200, 200)
	window.MakeWindowExist(w)

	f := &Frame{}
	widget.InitBase(&f.Base, w, app)
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
		w.BackgroundPixel = f.Background.Pixel
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
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			// Re-run place geometry manager if any children use place.
			place.ArrangeContainer(w)
			f.Display()
			if w.ConfigureCallback != nil {
				w.ConfigureCallback()
			}
		}
	})

	return f
}

// Display draws the frame.
func (f *Frame) Display() {
	if f.Destroyed {
		return
	}
	f.DrawBackground()
}

// Configure applies options to the frame.
func (f *Frame) Configure(opts ...option.Option) {
	option.Apply(f, opts)
	f.UpdateBorder()
	f.updateInternalBorder()
	if f.Background != nil {
		f.Win.BackgroundPixel = f.Background.Pixel
	}
	f.Display()
}

// Destroy cleans up the frame.
func (f *Frame) Destroy() {
	if f.Destroyed {
		return
	}
	f.Destroyed = true
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
	f.Base.DrawBackground()
	w.Display.Server.Flush()
}
