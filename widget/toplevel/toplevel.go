// Package toplevel implements additional toplevel windows.
// It creates new X11 top-level windows managed by the window manager.
package toplevel

import (
	"fmt"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
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

// IconName sets the icon name (WM_ICON_NAME / _NET_WM_ICON_NAME).
func IconName(s string) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetIconName(s)
	}
}

// Background sets the background color.
func Background[C color.Spec](name C) ToplevelOption {
	return func(t *Toplevel) { t.SetBackgroundColor(name) }
}

// Geometry sets the geometry string (e.g. "400x300+100+100").
func Geometry(geom string) ToplevelOption {
	return func(t *Toplevel) {
		if err := t.WmInfo.SetGeometry(geom); err != nil {
			t.OptionFailed(fmt.Errorf("toplevel: geometry %q: %w", geom, err))
		}
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
func TransientFor(parent window.Windower) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetTransientFor(parent.Window())
	}
}

// New creates a new Toplevel window.
func New(parent widget.Caregiver, name string, opts ...ToplevelOption) *Toplevel {
	app := parent.AppContext()
	d := parent.Window().Display

	// Create a new top-level window.
	w := &window.Window{
		Display:         d,
		Parent:          parent.Window(),
		Name:            name,
		PathName:        window.BuildPathName(parent.Window(), name),
		Width:           200,
		Height:          200,
		ReqWidth:        200,
		ReqHeight:       200,
		Depth:           d.Depth,
		BackgroundPixel: d.WhitePixel,
		Flags:           window.FlagTopLevel,
	}

	parent.Window().AddChild(w)

	// Create window as a child of the root (not the parent widget).
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
		0, 0, uint(w.Width), uint(w.Height), 0,
		d.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWEventMask,
		attrs,
	)

	d.RegisterWindow(w.PlatformID, w)

	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})

	t := &Toplevel{}
	widget.InitBase(&t.Base, w, app)
	t.SetDisplayProc(t.display)
	w.Class = "Toplevel"

	// Initialize WM state.
	t.WmInfo = wm.Init(w)

	for _, opt := range opts {
		opt(t)
	}

	// Update window background.
	if t.Background != nil {
		w.SetBackgroundPixel(t.Background.Pixel)
	}

	// Bind events.
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		t.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			t.Display()
			w.NotifyConfigure()
		}
	})

	return t
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (t *Toplevel) Display() {
	t.EventuallyRedraw()
}

// display draws the toplevel background.
func (t *Toplevel) display() {
	if t.Destroyed {
		return
	}
	t.DrawBackground()
	t.Win.Display.Server.Flush()
}

// Configure applies options.
func (t *Toplevel) Configure(opts ...ToplevelOption) error {
	return widget.Configure(t, opts, nil)
}

// Show maps the toplevel window.
func (t *Toplevel) Show() {
	if t.Win.PlatformID != platform.WindowID(0) {
		t.Win.Display.Server.MapRaised(t.Win.PlatformID)
		window.MarkMapped(t.Win)
	}
}

// Hide unmaps the toplevel window.
func (t *Toplevel) Hide() {
	if t.Win.PlatformID != platform.WindowID(0) {
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
