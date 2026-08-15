package window

import (
	"os"

	"github.com/msorc/takigo/platform"
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

	// Debug aid: when TAKIGO_DEBUG_NAME_WIDGETS=1 is set, set each widget's
	// Go name as its X11 window name so external tools (xdotool, scripts/
	// demo_interact.sh) can resolve widgets by identity rather than by
	// pixel coordinates. Opt-in: no behaviour change for normal runs.
	if os.Getenv("TAKIGO_DEBUG_NAME_WIDGETS") == "1" && w.Name != "" {
		d.Server.StoreName(w.PlatformID, w.Name)
	}

	// Set WM_DELETE_WINDOW protocol.
	protocols := []platform.AtomID{d.WMDeleteWindow}
	d.Server.SetWMProtocols(w.PlatformID, protocols)

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
	w.Flags |= FlagAlreadyDead

	// Destroy children first (copy slice since it mutates).
	children := make([]*Window, len(w.Children))
	copy(children, w.Children)
	for _, child := range children {
		DestroyWindow(child)
	}

	// Remove from parent.
	if w.Parent != nil {
		w.Parent.RemoveChild(w)
	}

	// Destroy platform resources.
	d := w.Display
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

// NewChildWindow creates a new child window.
func NewChildWindow(parent *Window, name string, x, y, width, height int) *Window {
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
