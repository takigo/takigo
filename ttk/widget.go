package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// TtkWidget is the base struct for all TTK themed widgets.
type TtkWidget struct {
	Win       *window.Window
	App       widget.AppContext
	State     State
	Layout    *Layout
	Theme     *Theme
	StyleName string
	Context   *DrawContext

	// Double-buffering pixmap.
	pixmap  platform.PixmapID
	pixmapW int
	pixmapH int

	NeedRedraw bool
	Destroyed  bool
}

// InitTtkWidget sets up a TTK widget: resolves style, creates layout, binds events.
func InitTtkWidget(w *TtkWidget, win *window.Window, app widget.AppContext, styleName string) {
	w.Win = win
	w.App = app
	w.StyleName = styleName
	w.Theme = CurrentTheme()

	if w.Theme == nil {
		return
	}

	style := w.Theme.ResolveStyle(styleName)

	w.Context = &DrawContext{
		Display: app.Server(),
		Depth:   win.Depth,
		Style:   style,
	}

	tmpl := w.Theme.GetLayout(styleName)
	if tmpl != nil {
		w.Layout = NewLayout(tmpl, w.Theme, w.Context, style)
	}

	// Set window background from style.
	bg := LookupColor(style, "-background", 0, 0xd9d9d9)
	win.BackgroundPixel = bg

	// Compute initial requested size.
	if w.Layout != nil {
		rw, rh := w.Layout.Size(w.State)
		if rw > 0 {
			win.ReqWidth = rw
		}
		if rh > 0 {
			win.ReqHeight = rh
		}
	}

	bindTtkCommon(w, app)
}

// Display renders the widget using double-buffered drawing.
func (w *TtkWidget) Display() {
	if w.Destroyed || w.Layout == nil {
		return
	}
	win := w.Win
	if win.PlatformID == 0 {
		return
	}

	d := win.Display.Server
	gc := win.GC
	width := win.Width
	height := win.Height

	if width <= 0 || height <= 0 {
		return
	}

	// Allocate or resize pixmap.
	if w.pixmap == 0 || w.pixmapW != width || w.pixmapH != height {
		if w.pixmap != 0 {
			d.FreePixmap(w.pixmap)
		}
		w.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		w.pixmapW = width
		w.pixmapH = height
	}

	pixDrawable := platform.PixmapDrawable(w.pixmap)

	// Clear pixmap with background.
	bg := LookupColor(w.Context.Style, "-background", w.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Layout and draw.
	bounds := Box{0, 0, width, height}
	w.Layout.Place(w.State, bounds)
	w.Layout.Draw(w.State, DrawArgs{Display: d, Drawable: pixDrawable, GC: gc})

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// ChangeState updates the widget state, redisplaying if changed.
func (w *TtkWidget) ChangeState(set, clear State) {
	old := w.State
	w.State = (w.State & ^clear) | set
	if w.State != old {
		w.Display()
	}
}

// Destroy frees resources and destroys the window.
func (w *TtkWidget) Destroy() {
	if w.Destroyed {
		return
	}
	w.Destroyed = true
	if w.pixmap != 0 {
		w.Win.Display.Server.FreePixmap(w.pixmap)
		w.pixmap = 0
	}
	window.DestroyWindow(w.Win)
}

// Window returns the underlying window.
func (w *TtkWidget) Window() *window.Window {
	return w.Win
}

// AppContext returns the application context, satisfying the widget.Caregiver interface.
func (w *TtkWidget) AppContext() widget.AppContext {
	return w.App
}

// bindTtkCommon binds common TTK events: Expose, Configure, Enter, Leave, Focus.
func bindTtkCommon(w *TtkWidget, app widget.AppContext) {
	win := w.Win

	// Expose.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		w.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			w.Display()
		}
	})

	// Enter → +StateHover +StateActive.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		w.ChangeState(StateHover|StateActive, 0)
	})

	// Leave → -StateHover -StateActive -StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		w.ChangeState(0, StateHover|StateActive|StatePressed)
	})

	// FocusIn → +StateFocus.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			w.ChangeState(StateFocus, 0)
		} else if ev.Type == event.FocusOutType {
			w.ChangeState(0, StateFocus)
		}
	})
}
