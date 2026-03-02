package ttk

import (
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Frame is a themed container widget.
type Frame struct {
	TtkWidget
}

// FrameOption configures a Frame.
type FrameOption func(*Frame)

// FramePadding sets the internal padding.
func FramePadding(p Padding) FrameOption {
	return func(f *Frame) {
		s := f.Context.Style
		s.Defaults["-padding"] = p
	}
}

// FrameBorderWidth sets the border width.
func FrameBorderWidth(w int) FrameOption {
	return func(f *Frame) {
		s := f.Context.Style
		s.Defaults["-borderwidth"] = w
	}
}

// FrameRelief sets the border relief.
func FrameRelief(r option.Relief) FrameOption {
	return func(f *Frame) {
		s := f.Context.Style
		s.Defaults["-relief"] = r
	}
}

// FrameBackground sets the background color pixel.
func FrameBackground(pixel uint64) FrameOption {
	return func(f *Frame) {
		s := f.Context.Style
		s.Defaults["-background"] = pixel
	}
}

// NewFrame creates a themed frame widget.
func NewFrame(parent *window.Window, name string, app widget.AppContext, opts ...FrameOption) *Frame {
	win := window.NewChildWindow(parent, name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	f := &Frame{}
	InitTtkWidget(&f.TtkWidget, win, app, "TFrame")

	for _, opt := range opts {
		opt(f)
	}

	// Re-resolve size after options.
	if f.Layout != nil {
		rw, rh := f.Layout.Size(f.State)
		if rw > win.ReqWidth {
			win.ReqWidth = rw
		}
		if rh > win.ReqHeight {
			win.ReqHeight = rh
		}
	}

	return f
}
