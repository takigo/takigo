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
		f.SetWidgetOption("-padding", p)
	}
}

// FrameBorderWidth sets the border width.
func FrameBorderWidth(w int) FrameOption {
	return func(f *Frame) {
		f.SetWidgetOption("-borderwidth", w)
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
	return func(f *Frame) {
		f.StyleName = name
		f.RefreshTheme()
	}
}

// NewFrame creates a themed frame widget.
func NewFrame(parent widget.Caregiver, name string, opts ...FrameOption) *Frame {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
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
