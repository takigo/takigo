// Package toplevel implements additional toplevel windows.
// It creates new X11 top-level windows managed by the window manager.
package toplevel

import (
	"log"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
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
func Background(name string) ToplevelOption {
	return func(t *Toplevel) {
		col, err := t.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("toplevel: failed to get color %q: %v", name, err)
			return
		}
		t.Background = col
		t.UpdateBorder()
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
func TransientFor(parent window.Windower) ToplevelOption {
	return func(t *Toplevel) {
		t.WmInfo.SetTransientFor(parent.Window())
	}
}

// --- Ttk-compatible aliases (prefix with Toplevel) for consistent naming ---
// These aliases match the naming convention used by ttk widgets
// allowing consistent option naming when both classic and ttk widgets are used.

// ToplevelTitle is an alias for Title.
var ToplevelTitle = Title

// ToplevelIconName is an alias for IconName.
var ToplevelIconName = IconName

// ToplevelBackground is an alias for Background.
var ToplevelBackground = Background

// ToplevelGeometry is an alias for Geometry.
var ToplevelGeometry = Geometry

// ToplevelResizable is an alias for Resizable.
var ToplevelResizable = Resizable

// ToplevelMinSize is an alias for MinSize.
var ToplevelMinSize = MinSize

// ToplevelTransientFor is an alias for TransientFor.
var ToplevelTransientFor = TransientFor

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
	w.Class = "Toplevel"

	// Initialize WM state.
	t.WmInfo = wm.Init(w)

	for _, opt := range opts {
		opt(t)
	}

	// Update window background.
	if t.Background != nil {
		w.BackgroundPixel = t.Background.Pixel
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

// Display draws the toplevel background.
func (t *Toplevel) Display() {
	if t.Destroyed {
		return
	}
	t.DrawBackground()
	t.Win.Display.Server.Flush()
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
