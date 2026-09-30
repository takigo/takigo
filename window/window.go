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

// WmInfo defines the interface for WM-specific window data.
// Implemented by wm.WmInfo to avoid circular imports.
type WmInfo interface {
	HandleClientMessage(messageType platform.AtomID, data [5]int64) bool
	OnDeleteWindow(fn func())
	OffDeleteWindow()
	SetGeometry(geom string) error
	// GeometryRequest returns the size the toplevel takes for a content
	// request of reqW x reqH: the request, unless the user set a size.
	GeometryRequest(reqW, reqH int) (w, h int)
	// ExpectSize records that the application asked for this size, so
	// the ConfigureNotify answering it is not taken for a user resize.
	ExpectSize(w, h int)
	// ConfigureNotify handles the toplevel's reported size before the
	// window takes it, recording a user resize as its geometry.
	ConfigureNotify(w, h int)
}

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
	Class    string // widget class like "Button" or "TButton" (TkWindow.classUid)

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

	// Visual depth (used for pixmap creation).
	Depth int

	// State flags.
	Flags int

	// Background pixel for the window.
	BackgroundPixel uint64
	// serverBackground is the background the platform window last got.
	serverBackground uint64

	// Graphics context for basic drawing.
	GC platform.GCID

	// Menubar is a toplevel's -menu window. Tk draws it in the wrapper
	// above the toplevel; takigo draws it at the top of the toplevel with
	// the content pushed below (InternalBorderTop).
	Menubar *Window

	// ConfigureCallback is called when the window is resized.
	//
	// Deprecated: it holds a single callback that any other user
	// overwrites; use OnConfigure.
	ConfigureCallback func()

	// configureHooks run on resize, after ConfigureCallback; see OnConfigure.
	configureHooks []func()

	// WmData stores per-toplevel WM state for toplevel windows.
	// Uses WmInfo interface to avoid circular imports between window and wm packages.
	// Mirrors TkWindow.wmInfoPtr in Tk's C code.
	WmData WmInfo

	// BackgroundHook is called when a recursive background change is applied.
	// Widgets register this to update their own Background field and pixel.
	BackgroundHook func(colorName string)

	// destroyHooks run when the window is destroyed; see OnDestroy.
	destroyHooks []func()

	// drawTarget, when set, is returned by Drawable in place of the
	// window so a widget's display procedure renders off-screen.
	drawTarget platform.DrawableID

	// values holds other packages' per-window state; see Value.
	values []valueEntry
}

// ValueKey identifies one package's entry in a window's values; use a
// *ValueKey of its own, e.g. var key = new(window.ValueKey). Keys compare
// by address, so the struct must not be empty.
type ValueKey struct{ _ byte }

type valueEntry struct {
	key *ValueKey
	v   any
}

// Value returns what SetValue stored on w under key, or nil. Packages keep
// their per-window state here, under a key of their own, instead of in a
// package-level map, which every App in the process would share: Tk keeps
// such tables per display (dispPtr->packerHashTable and the like). A window
// has only a few entries, so they are a slice searched by key address.
// A nil window has no values.
func (w *Window) Value(key *ValueKey) any {
	if w == nil {
		return nil
	}
	for i := range w.values {
		if w.values[i].key == key {
			return w.values[i].v
		}
	}
	return nil
}

// SetValue stores v on w under key; a nil v removes the entry.
func (w *Window) SetValue(key *ValueKey, v any) {
	for i := range w.values {
		if w.values[i].key != key {
			continue
		}
		if v != nil {
			w.values[i].v = v
			return
		}
		last := len(w.values) - 1
		w.values[i] = w.values[last]
		w.values[last] = valueEntry{}
		w.values = w.values[:last]
		return
	}
	if v == nil {
		return
	}
	if w.values == nil {
		w.values = make([]valueEntry, 0, 2)
	}
	w.values = append(w.values, valueEntry{key, v})
}

// OnDestroy registers fn to run when w is destroyed, whether directly or
// because an ancestor was. Hooks run last-registered first, like defer,
// so a widget's own cleanup runs before that of the base it embeds.
func (w *Window) OnDestroy(fn func()) {
	w.destroyHooks = append(w.destroyHooks, fn)
}

// OnConfigure registers fn to run whenever w is resized (NotifyConfigure).
// Geometry managers use it to re-arrange content; any number of hooks
// can coexist.
func (w *Window) OnConfigure(fn func()) {
	w.configureHooks = append(w.configureHooks, fn)
}

// NotifyConfigure runs w's resize callbacks. Call it after changing
// w.Width or w.Height.
func (w *Window) NotifyConfigure() {
	if w.ConfigureCallback != nil {
		w.ConfigureCallback()
	}
	for _, fn := range w.configureHooks {
		fn()
	}
}

// IsDestroyed reports whether w has been (or is being) destroyed.
func (w *Window) IsDestroyed() bool {
	return w.Flags&FlagAlreadyDead != 0
}

// ApplyBackgroundRecursive propagates a background color change through the
// window hierarchy. It calls BackgroundHook(colorName) on every window (and
// descendant) that has registered one, then triggers an expose event so the
// widget redraws with the new color.
func ApplyBackgroundRecursive(w *Window, colorName string) {
	applyBackgroundRecursiveDepth(w, colorName, 0)
}

func applyBackgroundRecursiveDepth(w *Window, colorName string, depth int) {
	if w == nil || depth > 1000 {
		return
	}
	if w.BackgroundHook != nil {
		w.BackgroundHook(colorName)
	}
	// Update X11 window background attribute and trigger a redraw.
	if w.PlatformID != 0 && w.Flags&FlagMapped != 0 && w.Width > 0 && w.Height > 0 {
		SyncBackground(w)
		w.Display.Server.ClearArea(w.PlatformID, 0, 0, uint(w.Width), uint(w.Height), true)
	}
	for _, child := range w.Children {
		applyBackgroundRecursiveDepth(child, colorName, depth+1)
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

// mappedHooks run when a window becomes mapped; geometry managers use them
// to arrange (and so map) the window's content, as Tk's managers do on
// MapNotify of a container.
var mappedHooks []func(*Window)

// AddMappedHook registers fn to run whenever MarkMapped maps a window.
func AddMappedHook(fn func(*Window)) { mappedHooks = append(mappedHooks, fn) }

// ResizeToplevel gives toplevel w the size for a content request of
// reqW x reqH, as Tk_GeometryRequest does through the window manager
// code: a size set with wm geometry is kept.
func ResizeToplevel(w *Window, reqW, reqH int) {
	width, height := reqW, reqH
	if w.WmData != nil {
		width, height = w.WmData.GeometryRequest(reqW, reqH)
	}
	if width == w.Width && height == w.Height {
		return
	}
	w.Width, w.Height = width, height
	if w.PlatformID != 0 {
		if w.WmData != nil {
			w.WmData.ExpectSize(width, height)
		}
		w.Display.Server.ResizeWindow(w.PlatformID, uint(width), uint(height))
	}
}

// SyncBackground gives w's platform window its current BackgroundPixel
// if that changed since the window was created: widgets pick their colour
// after the window exists, and Tk_SetWindowBackground keeps the server's
// copy in step, so the server clears to it on map and resize instead of
// the creation-time white. Call it before mapping w.
func SyncBackground(w *Window) {
	if w.PlatformID == 0 || w.serverBackground == w.BackgroundPixel {
		return
	}
	w.Display.Server.SetWindowBackground(w.PlatformID, w.BackgroundPixel)
	w.serverBackground = w.BackgroundPixel
}

// ContentInsets returns the internal borders and minimum request, which
// together decide how w's content is arranged. Tk_SetInternalBorderEx and
// Tk_SetMinimumRequestSize make the geometry managers recompute when they
// change; callers compare two snapshots and call NotifyConfigure.
func (w *Window) ContentInsets() [6]int {
	return [6]int{
		w.InternalBorderLeft, w.InternalBorderRight,
		w.InternalBorderTop, w.InternalBorderBottom,
		w.MinReqWidth, w.MinReqHeight,
	}
}

// SetBackgroundPixel records pixel as w's background and gives it to the
// platform window if that exists (Tk_SetWindowBackground).
func (w *Window) SetBackgroundPixel(pixel uint64) {
	w.BackgroundPixel = pixel
	SyncBackground(w)
}

// MarkMapped records that w has been mapped (after MapWindow) and, if it was
// not mapped before, lets the geometry managers map its content.
func MarkMapped(w *Window) {
	if w.Flags&FlagMapped != 0 {
		return
	}
	w.Flags |= FlagMapped
	for _, fn := range mappedHooks {
		fn(w)
	}
}

var unmappedHooks, movedHooks []func(*Window)

// AddUnmappedHook registers fn to run whenever MarkUnmapped unmaps a window.
func AddUnmappedHook(fn func(*Window)) { unmappedHooks = append(unmappedHooks, fn) }

// AddMovedHook registers fn to run whenever NotifyMoved reports a move.
func AddMovedHook(fn func(*Window)) { movedHooks = append(movedHooks, fn) }

// MarkUnmapped records that w has been unmapped (after UnmapWindow) and lets
// geometry managers unmap content they keep inside w from elsewhere
// (Tk_MaintainGeometry's UnmapNotify handling).
func MarkUnmapped(w *Window) {
	if w.Flags&FlagMapped == 0 {
		return
	}
	w.Flags &^= FlagMapped
	for _, fn := range unmappedHooks {
		fn(w)
	}
}

// NotifyMoved tells geometry managers that w changed position, so content
// they maintain inside it from elsewhere follows (Tk_MaintainGeometry's
// ConfigureNotify handling).
func NotifyMoved(w *Window) {
	for _, fn := range movedHooks {
		fn(w)
	}
}

// ContentOffset returns the position of container relative to content's
// parent, for content managed -in a container that is not its parent (the
// container must be the parent or one of its descendants).
func ContentOffset(container, content *Window) (dx, dy int) {
	for w := container; w != nil && w != content.Parent; w = w.Parent {
		dx += w.X
		dy += w.Y
	}
	return dx, dy
}

// ContainerViewable reports whether content in container may be mapped:
// the container and every window between it and content's parent must be
// mapped, as MaintainContentProc requires.
func ContainerViewable(container, content *Window) bool {
	for w := container; w != nil && w != content.Parent; w = w.Parent {
		if w.Flags&FlagMapped == 0 {
			return false
		}
	}
	return container.Flags&FlagMapped != 0
}

// IsViewable reports whether w and every ancestor up to its toplevel are
// mapped. Geometry managers map content before its toplevel is shown, so
// this stands in for Tk_IsMapped where Tk defers work until mapping.
func (w *Window) IsViewable() bool {
	for ; w != nil; w = w.Parent {
		if w.Flags&FlagMapped == 0 {
			return false
		}
		if w.Flags&FlagTopLevel != 0 {
			return true
		}
	}
	return false
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

// Drawable returns the DrawableID drawing operations should target: the
// window itself, or the pixmap set by SetDrawTarget.
func (w *Window) Drawable() platform.DrawableID {
	if w.drawTarget != 0 {
		return w.drawTarget
	}
	return platform.WindowDrawable(w.PlatformID)
}

// SetDrawTarget redirects Drawable to d until it is called again with 0.
func (w *Window) SetDrawTarget(d platform.DrawableID) {
	w.drawTarget = d
}
