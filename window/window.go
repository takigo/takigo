package window

import (
	"github.com/msorc/takigo/internal/xlib"
)

// Flags for Window.Flags field.
const (
	FlagTopLevel   = 1 << iota // this is a top-level window
	FlagMapped                 // window is currently mapped
	FlagAlreadyDead            // destruction in progress
)

// GeomManager is the interface for geometry managers, defined here
// to avoid circular imports between window and geometry packages.
type GeomManager interface {
	Name() string
	RequestProc(content *Window)
	LostContentProc(content *Window)
}

// Window represents a single window in the takigo hierarchy.
// Ports TkWindow from tk/generic/tkInt.h.
type Window struct {
	// X11 identity.
	XWindow xlib.Window // X11 window ID (0 = not yet created)
	Display *Display

	// Hierarchy.
	Parent   *Window
	Children []*Window
	PathName string // full path like ".frame1.button1"
	Name     string // local name within parent

	// Geometry.
	X, Y          int
	Width, Height int
	BorderWidth   int
	ReqWidth      int // requested width
	ReqHeight     int // requested height

	// Internal borders (area where children cannot be placed).
	InternalBorderLeft   int
	InternalBorderRight  int
	InternalBorderTop    int
	InternalBorderBottom int

	// Geometry manager currently managing this window.
	GeomManager GeomManager
	GeomData    any // manager-specific data for this window

	// Visual.
	Depth    int
	Visual   *xlib.Visual
	Colormap xlib.Colormap

	// State flags.
	Flags int

	// Background pixel for the window.
	BackgroundPixel uint64

	// Graphics context for basic drawing.
	GC xlib.GC
}

// IsTopLevel returns true if this is a top-level window.
func (w *Window) IsTopLevel() bool {
	return w.Flags&FlagTopLevel != 0
}

// IsMapped returns true if the window is mapped.
func (w *Window) IsMapped() bool {
	return w.Flags&FlagMapped != 0
}

// Drawable returns the window as an xlib.Drawable for drawing operations.
func (w *Window) Drawable() xlib.Drawable {
	return xlib.Drawable(w.XWindow)
}
