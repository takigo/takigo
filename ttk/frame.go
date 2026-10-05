package ttk

import (
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Frame is a themed container widget.
type Frame struct {
	TtkWidget
	padding     Padding
	borderWidth int
	width       int // -width; 0 = none
	height      int // -height; 0 = none
}

// FrameOption configures a Frame.
type FrameOption func(*Frame)

// FrameWidth sets -width (a Tk distance).
func FrameWidth[L screenunit.Length](v L) FrameOption {
	return func(f *Frame) { f.width = screenunit.ToPixels(v) }
}

// FrameHeight sets -height (a Tk distance).
func FrameHeight[L screenunit.Length](v L) FrameOption {
	return func(f *Frame) { f.height = screenunit.ToPixels(v) }
}

// FramePadding sets the internal padding.
func FramePadding(p Padding) FrameOption {
	return func(f *Frame) {
		f.padding = p
		f.SetWidgetOption("-padding", p)
	}
}

// FrameBorderWidth sets the border width (a Tk distance).
func FrameBorderWidth[L screenunit.Length](w L) FrameOption {
	return func(f *Frame) {
		f.borderWidth = screenunit.ToPixels(w)
		f.SetWidgetOption("-borderwidth", f.borderWidth)
	}
}

// FrameRelief sets the border relief.
func FrameRelief(r option.Relief) FrameOption {
	return func(f *Frame) {
		f.SetWidgetOption("-relief", r)
	}
}

// FrameBackground sets the background color pixel.
func FrameBackground(pixel uint64) FrameOption {
	return func(f *Frame) {
		f.SetWidgetOption("-background", pixel)
	}
}

// FrameStyleOpt sets the -style, e.g. "TEntry" to draw a frame like an
// entry field (as the ttkpane demo does around a classic text widget).
func FrameStyleOpt(name string) FrameOption {
	return func(f *Frame) { f.StyleName = name }
}

// NewFrame creates a themed frame widget.
func NewFrame(parent widget.Caregiver, name string, opts ...FrameOption) *Frame {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	f := &Frame{}
	InitTtkWidget(&f.TtkWidget, win, app, "TFrame")
	f.reconfigure = func() { _ = f.Configure() }

	for _, opt := range opts {
		opt(f)
	}
	if f.StyleName != "TFrame" {
		f.RefreshTheme()
	}
	f.updateMargins()

	// FrameSize requests only -width/-height (0 when unset); the margins
	// are added on top.
	win.ReqWidth, win.ReqHeight = max(f.width, 1), max(f.height, 1)

	return f
}

// updateMargins ports FrameMargins: the widget's own -padding plus
// -borderwidth (not the style's) become the content margins.
func (f *Frame) updateMargins() {
	win := f.Win
	win.InternalBorderLeft = f.padding.Left + f.borderWidth
	win.InternalBorderRight = f.padding.Right + f.borderWidth
	win.InternalBorderTop = f.padding.Top + f.borderWidth
	win.InternalBorderBottom = f.padding.Bottom + f.borderWidth
}

// Configure sets options after creation. Only a changed -width/-height
// resets the request: otherwise it belongs to the content's geometry
// manager, which the changed margins make recompute it.
func (f *Frame) Configure(opts ...FrameOption) error {
	width, height := f.width, f.height
	return configure(&f.TtkWidget, f, opts, nil, func() {
		f.updateMargins()
		if f.width != width || f.height != height {
			f.Win.ReqWidth, f.Win.ReqHeight = max(f.width, 1), max(f.height, 1)
		}
	})
}
