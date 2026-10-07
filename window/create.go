package window

import (
	"log/slog"
	"os"
	"slices"
	"strconv"

	"github.com/takigo/takigo/platform"
)

// CreateMainWindow creates the root window of a takigo application.
// This is analogous to Tk_CreateMainWindow in tk/generic/tkWindow.c.
func CreateMainWindow(d *Display, x, y, width, height int) *Window {
	w := &Window{
		Display:         d,
		PathName:        ".",
		Name:            "",
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		ReqWidth:        width,
		ReqHeight:       height,
		Depth:           d.Depth,
		BackgroundPixel: d.WhitePixel,
		Flags:           FlagTopLevel,
	}

	// Create the actual platform window.
	attrs := &platform.WindowAttrs{
		BackgroundPixel: w.BackgroundPixel,
		BorderPixel:     d.BlackPixel,
		EventMask: int64(
			platform.KeyPressMask |
				platform.KeyReleaseMask |
				platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.PointerMotionMask |
				platform.EnterWindowMask |
				platform.LeaveWindowMask |
				platform.ExposureMask |
				platform.StructureNotifyMask |
				platform.PropertyChangeMask |
				platform.FocusChangeMask),
	}

	w.PlatformID = d.Server.CreateWindow(
		d.RootWindow,
		x, y, uint(width), uint(height), 0,
		d.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWEventMask|platform.CWOverrideRedirect,
		attrs,
	)

	// Register in display's window table.
	d.RegisterWindow(w.PlatformID, w)
	w.serverBackground = w.BackgroundPixel

	// Debug aid: when TAKIGO_DEBUG_NAME_WIDGETS=1 is set, set each widget's
	// Go name as its X11 window name so external tools (xdotool, scripts/
	// demo_interact.sh) can resolve widgets by identity rather than by
	// pixel coordinates. Opt-in: no behaviour change for normal runs.
	if os.Getenv("TAKIGO_DEBUG_NAME_WIDGETS") == "1" && w.Name != "" {
		d.Server.StoreName(w.PlatformID, w.Name)
	}

	// Create a default GC for drawing.
	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})

	return w
}

// MakeWindowExist ensures the platform window exists for a non-top-level window.
// Top-level windows are created eagerly; child windows may be created lazily.
func MakeWindowExist(w *Window) {
	if w.PlatformID != 0 {
		return
	}

	d := w.Display
	parent := w.Parent
	if parent == nil || parent.PlatformID == 0 {
		return
	}

	attrs := &platform.WindowAttrs{
		BackgroundPixel: w.BackgroundPixel,
		BorderPixel:     d.BlackPixel,
		EventMask: int64(
			platform.KeyPressMask |
				platform.KeyReleaseMask |
				platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.PointerMotionMask |
				platform.EnterWindowMask |
				platform.LeaveWindowMask |
				platform.ExposureMask |
				platform.StructureNotifyMask |
				platform.PropertyChangeMask |
				platform.FocusChangeMask),
	}

	w.PlatformID = d.Server.CreateWindow(
		parent.PlatformID,
		w.X, w.Y, uint(w.Width), uint(w.Height), uint(w.BorderWidth),
		w.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWEventMask,
		attrs,
	)

	d.RegisterWindow(w.PlatformID, w)
	w.serverBackground = w.BackgroundPixel

	// See CreateMainWindow for the TAKIGO_DEBUG_NAME_WIDGETS rationale.
	if os.Getenv("TAKIGO_DEBUG_NAME_WIDGETS") == "1" && w.Name != "" {
		d.Server.StoreName(w.PlatformID, w.Name)
	}

	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: w.BackgroundPixel,
	})
}

// DestroyWindow recursively destroys a window and its children.
func DestroyWindow(w *Window) {
	destroyWindowDepth(w, 0)
}

func destroyWindowDepth(w *Window, depth int) {
	if w == nil || w.Flags&FlagAlreadyDead != 0 {
		return
	}
	if depth > 1000 {
		slog.Warn("window: destroy exceeds depth 1000; subtree leaked", "window", w.PathName)
		return
	}
	w.Flags |= FlagAlreadyDead

	// Destroy children first (copy slice since it mutates).
	children := make([]*Window, len(w.Children))
	copy(children, w.Children)
	for _, child := range children {
		destroyWindowDepth(child, depth+1)
	}

	// As in Tk_DestroyWindow, <Destroy> handlers see the window after its
	// children are gone, then the widget frees its own resources.
	d := w.Display
	if d != nil {
		for _, fn := range d.destroyHooks {
			fn(w)
		}
	}
	hooks := w.destroyHooks
	w.destroyHooks = nil
	for _, hook := range slices.Backward(hooks) {
		hook()
	}
	if w.GeomManager != nil {
		w.GeomManager.LostContentProc(w)
		w.GeomManager = nil
	}
	w.ConfigureCallback = nil
	w.configureHooks = nil
	w.BackgroundHook = nil

	// Remove from parent.
	if w.Parent != nil {
		w.Parent.RemoveChild(w)
	}

	// Destroy platform resources.
	if d == nil || d.Server == nil {
		return
	}
	if !platform.IsZeroGC(w.GC) {
		d.Server.FreeGC(w.GC)
		w.GC = platform.ZeroGC()
	}
	if w.PlatformID != 0 {
		d.UnregisterWindow(w.PlatformID)
		d.Server.DestroyWindow(w.PlatformID)
		w.PlatformID = 0
	}
}

// TopLevelSpec describes a window created directly under the root X window:
// a toplevel, a menu or a torn-off menu.
type TopLevelSpec struct {
	X, Y, Width, Height int
	BorderWidth         uint
	OverrideRedirect    bool // bypass the window manager (menus)
	EventMask           int64
	Flags               int
}

// NewTopLevelWindow creates a window under the root X window, with its GC,
// and adds it to parent's children under name. As in Tk, the widget
// hierarchy is not the X hierarchy for these windows.
func NewTopLevelWindow(parent *Window, name string, spec TopLevelSpec) *Window {
	d := parent.Display
	w := &Window{
		Display:         d,
		Parent:          parent,
		Name:            name,
		PathName:        BuildPathName(parent, name),
		X:               spec.X,
		Y:               spec.Y,
		Width:           spec.Width,
		Height:          spec.Height,
		ReqWidth:        spec.Width,
		ReqHeight:       spec.Height,
		Depth:           d.Depth,
		BackgroundPixel: d.WhitePixel,
		Flags:           spec.Flags,
	}
	mask := uint64(platform.CWBackPixel | platform.CWBorderPixel | platform.CWEventMask)
	if spec.OverrideRedirect {
		mask |= platform.CWOverrideRedirect
	}
	w.PlatformID = d.Server.CreateWindow(d.RootWindow,
		spec.X, spec.Y, uint(spec.Width), uint(spec.Height), spec.BorderWidth,
		d.Depth, platform.InputOutput, mask,
		&platform.WindowAttrs{
			BackgroundPixel:  w.BackgroundPixel,
			BorderPixel:      d.BlackPixel,
			OverrideRedirect: spec.OverrideRedirect,
			EventMask:        spec.EventMask,
		})
	d.RegisterWindow(w.PlatformID, w)
	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})
	parent.AddChild(w)
	return w
}

func uniqueChildName(parent *Window, name string) string {
	taken := func(n string) bool {
		for _, c := range parent.Children {
			if c.Name == n {
				return true
			}
		}
		return false
	}
	if name == "" {
		for {
			parent.autoNames++
			if n := "w" + strconv.Itoa(parent.autoNames); !taken(n) {
				return n
			}
		}
	}
	if !taken(name) {
		return name
	}
	for i := 2; ; i++ {
		if n := name + "#" + strconv.Itoa(i); !taken(n) {
			slog.Warn("window name already exists in parent; renamed",
				"parent", parent.PathName, "name", name, "renamed", n)
			return n
		}
	}
}

// NewChildWindow creates a new child window. An empty name gets a generated
// one ("w1", "w2", ...). Tk rejects a name a sibling already has; here the
// newcomer is renamed ("name#2") and a warning logged, so that the two do
// not share a path, and with it their bindings.
func NewChildWindow(parent *Window, name string, x, y, width, height int) *Window {
	name = uniqueChildName(parent, name)
	w := &Window{
		Display:         parent.Display,
		Parent:          parent,
		Name:            name,
		PathName:        BuildPathName(parent, name),
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		ReqWidth:        width,
		ReqHeight:       height,
		Depth:           parent.Depth,
		BackgroundPixel: parent.Display.WhitePixel,
	}

	parent.AddChild(w)
	return w
}
