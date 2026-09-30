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
func NewSeparator(parent widget.Caregiver, name string, opts ...SeparatorOption) *Separator {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	s := &Separator{Orient: Horizontal}

	for _, opt := range opts {
		opt(s)
	}

	InitTtkWidget(&s.TtkWidget, win, app, s.orientStyle())
	s.requestSize()

	return s
}

func (s *Separator) orientStyle() string {
	if s.Orient == Vertical {
		return "TSeparator.Vertical"
	}
	return "TSeparator.Horizontal"
}

// requestSize ports SeparatorElementSize (tk/generic/ttk/ttkElements.c):
// 2x2 for both orientations; the geometry manager stretches it.
func (s *Separator) requestSize() {
	s.Win.ReqWidth, s.Win.ReqHeight = 2, 2
}

// Configure sets options after creation.
func (s *Separator) Configure(opts ...SeparatorOption) {
	configure(&s.TtkWidget, s, opts, func() { s.StyleName = s.orientStyle() }, s.requestSize)
}
