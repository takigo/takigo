package ttk

import (
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Separator is a themed separator widget.
type Separator struct {
	TtkWidget
	Orient Orientation
}

// SeparatorOption configures a Separator.
type SeparatorOption func(*Separator)

// SeparatorOrient sets the separator orientation.
func SeparatorOrient(o Orientation) SeparatorOption {
	return func(s *Separator) { s.Orient = o }
}

// NewSeparator creates a themed separator widget.
func NewSeparator(parent *window.Window, name string, app widget.AppContext, opts ...SeparatorOption) *Separator {
	win := window.NewChildWindow(parent, name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	s := &Separator{Orient: Horizontal}

	for _, opt := range opts {
		opt(s)
	}

	// Choose style based on orientation.
	styleName := "TSeparator.Horizontal"
	if s.Orient == Vertical {
		styleName = "TSeparator.Vertical"
	}

	InitTtkWidget(&s.TtkWidget, win, app, styleName)

	// Set sensible minimum size.
	if s.Orient == Horizontal {
		win.ReqWidth = 20
		win.ReqHeight = 2
	} else {
		win.ReqWidth = 2
		win.ReqHeight = 20
	}

	return s
}
