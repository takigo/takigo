package window

import (
	"github.com/msorc/takigo/internal/xlib"
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
		Visual:          d.Visual,
		Colormap:        d.Colormap,
		BackgroundPixel: d.WhitePixel,
		Flags:           FlagTopLevel,
	}

	// Create the actual X11 window.
	attrs := &xlib.WindowAttributes{
		BackgroundPixel: w.BackgroundPixel,
		BorderPixel:     d.BlackPixel,
		EventMask: int64(
			xlib.KeyPressMask |
				xlib.KeyReleaseMask |
				xlib.ButtonPressMask |
				xlib.ButtonReleaseMask |
				xlib.PointerMotionMask |
				xlib.EnterWindowMask |
				xlib.LeaveWindowMask |
				xlib.ExposureMask |
				xlib.StructureNotifyMask |
				xlib.FocusChangeMask),
		Colormap: d.Colormap,
	}

	w.XWindow = d.XDisplay.CreateWindow(
		d.RootXWindow,
		x, y, uint(width), uint(height), 0,
		d.Depth, xlib.InputOutput, d.Visual,
		xlib.CWBackPixel|xlib.CWBorderPixel|xlib.CWEventMask|xlib.CWColormap,
		attrs,
	)

	// Register in display's window table.
	d.RegisterWindow(w.XWindow, w)

	// Set WM_DELETE_WINDOW protocol.
	protocols := []xlib.Atom{d.WMDeleteWindow}
	d.XDisplay.SetWMProtocols(w.XWindow, protocols)

	// Create a default GC for drawing.
	w.GC = d.XDisplay.CreateGC(w.Drawable(), xlib.GCForeground|xlib.GCBackground, &xlib.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})

	return w
}

// MakeWindowExist ensures the X11 window exists for a non-top-level window.
// Top-level windows are created eagerly; child windows may be created lazily.
func MakeWindowExist(w *Window) {
	if w.XWindow != xlib.Window(0) {
		return
	}

	d := w.Display
	parent := w.Parent
	if parent == nil || parent.XWindow == xlib.Window(0) {
		return
	}

	attrs := &xlib.WindowAttributes{
		BackgroundPixel: w.BackgroundPixel,
		BorderPixel:     d.BlackPixel,
		EventMask: int64(
			xlib.KeyPressMask |
				xlib.KeyReleaseMask |
				xlib.ButtonPressMask |
				xlib.ButtonReleaseMask |
				xlib.PointerMotionMask |
				xlib.EnterWindowMask |
				xlib.LeaveWindowMask |
				xlib.ExposureMask |
				xlib.StructureNotifyMask |
				xlib.FocusChangeMask),
	}

	w.XWindow = d.XDisplay.CreateWindow(
		parent.XWindow,
		w.X, w.Y, uint(w.Width), uint(w.Height), uint(w.BorderWidth),
		w.Depth, xlib.InputOutput, w.Visual,
		xlib.CWBackPixel|xlib.CWBorderPixel|xlib.CWEventMask,
		attrs,
	)

	d.RegisterWindow(w.XWindow, w)

	w.GC = d.XDisplay.CreateGC(w.Drawable(), xlib.GCForeground|xlib.GCBackground, &xlib.GCValues{
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

	// Destroy X resources.
	d := w.Display
	if !xlib.IsZeroGC(w.GC) {
		d.XDisplay.FreeGC(w.GC)
		w.GC = xlib.ZeroGC()
	}
	if w.XWindow != xlib.Window(0) {
		d.UnregisterWindow(w.XWindow)
		d.XDisplay.DestroyWindow(w.XWindow)
		w.XWindow = xlib.Window(0)
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
		Visual:          parent.Visual,
		Colormap:        parent.Colormap,
		BackgroundPixel: parent.Display.WhitePixel,
	}

	parent.AddChild(w)
	return w
}
