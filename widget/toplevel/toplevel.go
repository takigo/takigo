// Package toplevel implements additional toplevel windows.
// It creates new X11 top-level windows managed by the window manager.
package toplevel

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
	"github.com/msorc/takigo/wm"
)

// Toplevel is a top-level window managed by the window manager.
type Toplevel struct {
	widget.Base
	WmInfo *wm.WmInfo
}

// ToplevelOption configures a Toplevel.
type ToplevelOption func(*Toplevel)

// Title sets the window title.
func Title(s string) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetTitle(s)
	}
}

// Background sets the background color.
func Background(name string) ToplevelOption {
	return func(t *Toplevel) {
		col, err := t.App.ColorCache().Get(name)
		if err == nil {
			t.Background = col
			t.UpdateBorder()
		}
	}
}

// Geometry sets the geometry string (e.g. "400x300+100+100").
func Geometry(geom string) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetGeometry(geom)
	}
}

// Resizable sets whether the window can be resized.
func Resizable(w, h bool) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetResizable(w, h)
	}
}

// MinSize sets the minimum window size.
func MinSize(w, h int) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetMinSize(w, h)
	}
}

// TransientFor marks this as a transient window (dialog) for the given parent.
func TransientFor(parent *window.Window) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetTransientFor(parent)
	}
}

// New creates a new Toplevel window.
func New(parent *window.Window, name string, app widget.AppContext, opts ...ToplevelOption) *Toplevel {
	d := parent.Display

	// Create a new X top-level window.
	w := &window.Window{
		Display:         d,
		Parent:          parent,
		Name:            name,
		PathName:        window.BuildPathName(parent, name),
		Width:           200,
		Height:          200,
		ReqWidth:        200,
		ReqHeight:       200,
		Depth:           d.Depth,
		Visual:          d.Visual,
		Colormap:        d.Colormap,
		BackgroundPixel: d.WhitePixel,
		Flags:           window.FlagTopLevel,
	}

	parent.AddChild(w)

	// Create X window as a child of the root (not the parent widget).
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
		0, 0, uint(w.Width), uint(w.Height), 0,
		d.Depth, xlib.InputOutput, d.Visual,
		xlib.CWBackPixel|xlib.CWBorderPixel|xlib.CWEventMask|xlib.CWColormap,
		attrs,
	)

	d.RegisterWindow(w.XWindow, w)

	w.GC = d.XDisplay.CreateGC(w.Drawable(), xlib.GCForeground|xlib.GCBackground, &xlib.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})

	t := &Toplevel{}
	widget.InitBase(&t.Base, w, app)

	// Initialize WM state.
	t.WmInfo = wm.Init(w)

	for _, opt := range opts {
		opt(t)
	}

	// Update X window background.
	if t.Background != nil {
		w.BackgroundPixel = t.Background.Pixel
	}

	// Bind events.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		t.Display()
	})

	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			t.Display()
			if w.ConfigureCallback != nil {
				w.ConfigureCallback()
			}
		}
	})

	return t
}

// Display draws the toplevel background.
func (t *Toplevel) Display() {
	if t.Destroyed {
		return
	}
	t.DrawBackground()
	t.Win.Display.XDisplay.Flush()
}

// Configure applies options.
func (t *Toplevel) Configure(opts ...option.Option) {
	option.Apply(t, opts)
	t.UpdateBorder()
	if t.Background != nil {
		t.Win.BackgroundPixel = t.Background.Pixel
	}
	t.Display()
}

// Show maps the toplevel window.
func (t *Toplevel) Show() {
	if t.Win.XWindow != xlib.Window(0) {
		t.Win.Display.XDisplay.MapRaised(t.Win.XWindow)
		t.Win.Flags |= window.FlagMapped
	}
}

// Hide unmaps the toplevel window.
func (t *Toplevel) Hide() {
	if t.Win.XWindow != xlib.Window(0) {
		t.WmInfo.Withdraw()
		t.Win.Flags &^= window.FlagMapped
	}
}

// Destroy cleans up the toplevel.
func (t *Toplevel) Destroy() {
	if t.Destroyed {
		return
	}
	t.Destroyed = true
	window.DestroyWindow(t.Win)
}

// OnClose registers a callback for when the user clicks the window close button.
func (t *Toplevel) OnClose(fn func()) {
	t.WmInfo.OnDeleteWindow(fn)
}
