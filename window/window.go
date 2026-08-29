package window

import (
	"github.com/msorc/takigo/platform"
)

// Flags for Window.Flags field.
const (
	FlagTopLevel    = 1 << iota // this is a top-level window
	FlagMapped                  // window is currently mapped
	FlagAlreadyDead             // destruction in progress
	FlagFocusable               // widget can receive keyboard focus
)

// GeomManager is the interface for geometry managers, defined here
// to avoid circular imports between window and geometry packages.
type GeomManager interface {
	Name() string
	RequestProc(content *Window)
	LostContentProc(content *Window)
}

// Windower is implemented by anything that wraps a Window (widgets, bare windows).
type Windower interface {
	Window() *Window
}

// Window returns the Window itself, satisfying the Windower interface.
func (w *Window) Window() *Window { return w }

// Window represents a single window in the takigo hierarchy.
// Ports TkWindow from tk/generic/tkInt.h.
type Window struct {
	// Platform identity.
	PlatformID platform.WindowID // platform window handle (0 = not yet created)
	Display    *Display

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
	MinReqWidth   int // minimum requested width (Tk_SetMinimumRequestSize)
	MinReqHeight  int // minimum requested height

	// Internal borders (area where children cannot be placed).
	InternalBorderLeft   int
	InternalBorderRight  int
	InternalBorderTop    int
	InternalBorderBottom int

	// Geometry manager currently managing this window.
	GeomManager GeomManager
	GeomData    any // manager-specific data for this window

	// Visual depth (used for pixmap creation).
	Depth int

	// State flags.
	Flags int

	// Background pixel for the window.
	BackgroundPixel uint64

	// Graphics context for basic drawing.
	GC platform.GCID

	// ConfigureCallback is called when the window is resized.
	// Set by geometry managers (e.g. pack) to re-layout children.
	ConfigureCallback func()

	// WmData stores per-toplevel WM state (*wm.WmInfo) for toplevel windows.
	// Typed as any to avoid circular imports between window and wm packages.
	// Mirrors TkWindow.wmInfoPtr in Tk's C code.
	WmData any

	// BackgroundHook is called when a recursive background change is applied.
	// Widgets register this to update their own Background field and pixel.
	BackgroundHook func(colorName string)
}

// ApplyBackgroundRecursive propagates a background color change through the
// window hierarchy. It calls BackgroundHook(colorName) on every window (and
// descendant) that has registered one, then triggers an expose event so the
// widget redraws with the new color.
func ApplyBackgroundRecursive(w *Window, colorName string) {
	if w.BackgroundHook != nil {
		w.BackgroundHook(colorName)
	}
	// Update X11 window background attribute and trigger a redraw.
	if w.PlatformID != 0 && w.Flags&FlagMapped != 0 && w.Width > 0 && w.Height > 0 {
		w.Display.Server.SetWindowBackground(w.PlatformID, w.BackgroundPixel)
		w.Display.Server.ClearArea(w.PlatformID, 0, 0, uint(w.Width), uint(w.Height), true)
	}
	for _, child := range w.Children {
		ApplyBackgroundRecursive(child, colorName)
	}
}

// IsTopLevel returns true if this is a top-level window.
func (w *Window) IsTopLevel() bool {
	return w.Flags&FlagTopLevel != 0
}

// Toplevel walks up the window hierarchy and returns the nearest toplevel
// ancestor (or w itself if w is a toplevel). Returns nil if no toplevel found.
func Toplevel(w *Window) *Window {
	for w != nil {
		if w.Flags&FlagTopLevel != 0 {
			return w
		}
		w = w.Parent
	}
	return nil
}

// IsMapped returns true if the window is mapped.
func (w *Window) IsMapped() bool {
	return w.Flags&FlagMapped != 0
}

// SetCursor sets the cursor for this window to the given font cursor shape.
// SetCursor sets the cursor shape for this window.
// Use cursor.Shape constants from the cursor package.
func (w *Window) SetCursor(shape uint) {
	if w.PlatformID == 0 {
		return
	}
	w.Display.Server.SetCursorShape(w.PlatformID, shape)
}

// ResetCursor reverts this window to its parent's cursor.
func (w *Window) ResetCursor() {
	if w.PlatformID == 0 {
		return
	}
	w.Display.Server.UndefineCursor(w.PlatformID)
}

// Drawable returns the window as a DrawableID for drawing operations.
func (w *Window) Drawable() platform.DrawableID {
	return platform.WindowDrawable(w.PlatformID)
}
