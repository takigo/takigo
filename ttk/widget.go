// Package ttk implements themed (Tile) widgets with a theme engine, layout
// engine, state system, and element interface. It ports tk/generic/ttk/*.c.
package ttk

import (
	"strings"

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

	// LabelFactory is set by widgets that use a bound label element (Button, Label, Menubutton).
	// Used by RefreshTheme to rebuild the layout with the correct label provider.
	LabelFactory ElementFactory

	// Double-buffering pixmap.
	pixmap  platform.PixmapID
	pixmapW int
	pixmapH int

	// NeedRedraw is set while a redisplay is queued for idle time.
	NeedRedraw bool
	Destroyed  bool

	// widgetOpts holds widget-level element options (e.g. a label's
	// -padding or -wraplength). Tk resolves element options on the widget
	// before its style, so this is a child style layered over the real one.
	widgetOpts *Style

	// DisplayFunc is the concrete widget's Display method.
	// Set by widgets with custom Display (notebook, scrollbar, etc.)
	// so ChangeState calls the right method.
	DisplayFunc func()
}

// InitTtkWidget sets up a TTK widget: resolves style, creates layout, binds events.
func InitTtkWidget(w *TtkWidget, win *window.Window, app widget.AppContext, styleName string) {
	initTtkBase(w, win, app, styleName)
	if w.Theme == nil {
		return
	}
	if tmpl := w.Theme.GetLayout(styleName); tmpl != nil {
		w.Layout = NewLayout(tmpl, w.Theme, w.Context, w.Context.Style)
	}
	w.updateReqFromLayout()
}

// initTtkBase is InitTtkWidget without the layout, for widgets that build
// their own (with a bound label element) once their options are applied.
func initTtkBase(w *TtkWidget, win *window.Window, app widget.AppContext, styleName string) {
	win.OnDestroy(w.Destroy)
	w.Win = win
	w.App = app
	w.StyleName = styleName
	w.Theme = CurrentTheme()
	if win.Class == "" {
		win.Class = classForStyle(styleName)
	}
	bindTtkCommon(w, app)

	if w.Theme == nil {
		return
	}

	style := w.Theme.ResolveStyle(styleName)

	w.Context = &DrawContext{
		Display: app.Server(),
		Depth:   win.Depth,
		Style:   style,
	}

	// Set window background from style.
	bg := LookupColor(style, "-background", 0, 0xd9d9d9)
	win.BackgroundPixel = bg
}

// updateReqFromLayout sets the window's request from the layout's size.
func (w *TtkWidget) updateReqFromLayout() {
	if w.Layout == nil {
		return
	}
	rw, rh := w.Layout.Size(w.State)
	if rw > 0 {
		w.Win.ReqWidth = rw
	}
	if rh > 0 {
		w.Win.ReqHeight = rh
	}
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
	if w.pixmap == 0 {
		return
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
		w.redisplay()
	}
}

// RefreshTheme updates the widget to use the current global theme.
// It re-resolves the style, rebuilds the layout, and updates the window background.
// Callers should call Display() afterwards (the concrete widget's Display, not TtkWidget's).
func (w *TtkWidget) RefreshTheme() {
	theme := CurrentTheme()
	if theme == nil {
		return
	}
	w.Theme = theme

	style := theme.ResolveStyle(w.StyleName)
	w.setStyle(style)
	style = w.Context.Style

	tmpl := theme.GetLayout(w.StyleName)
	if tmpl != nil {
		if w.LabelFactory != nil {
			w.Layout = newLayoutWithLabel(tmpl, theme, w.Context, style, w.LabelFactory)
		} else {
			w.Layout = NewLayout(tmpl, theme, w.Context, style)
		}
	} else {
		w.Layout = nil
	}

	// Update window background from new style.
	bg := LookupColor(style, "-background", 0, 0xd9d9d9)
	w.Win.BackgroundPixel = bg
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

// To be used by geometry managers as a geometry.Elementer
func (w *TtkWidget) GeometryElements() []window.Windower {
	return []window.Windower{w}
}

// AppContext returns the application context, satisfying the widget.Caregiver interface.
func (w *TtkWidget) AppContext() widget.AppContext {
	return w.App
}

// redisplay draws with the concrete widget's DisplayFunc when it has one
// (widgets without a layout, e.g. progressbar), else the layout display.
// redisplay schedules one idle-time redraw, as TtkRedisplayWidget does, so
// the Expose, Configure and state changes of one event burst draw once.
func (w *TtkWidget) redisplay() {
	if w.NeedRedraw || w.Destroyed {
		return
	}
	if w.App == nil {
		w.displayNow()
		return
	}
	w.NeedRedraw = true
	w.App.DoWhenIdle(func() {
		w.NeedRedraw = false
		if !w.Destroyed {
			w.displayNow()
		}
	})
}

// displayNow draws with the concrete widget's Display.
func (w *TtkWidget) displayNow() {
	if w.DisplayFunc != nil {
		w.DisplayFunc()
		return
	}
	w.Display()
}

// bindTtkCommon binds common TTK events: Expose, Configure, Enter, Leave, Focus.
func bindTtkCommon(w *TtkWidget, app widget.AppContext) {
	win := w.Win

	// Expose.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		w.redisplay()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			w.redisplay()
			win.NotifyConfigure()
		}
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

// bindTtkHover binds Enter/Leave events to set hover state.
// Only interactive widgets (buttons, scrollbars, etc.) should call this.
// Container widgets (frames, notebooks, etc.) should NOT have hover effects.
func bindTtkHover(w *TtkWidget, app widget.AppContext) {
	win := w.Win

	// Enter → +StateHover +StateActive.
	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		w.ChangeState(StateHover|StateActive, 0)
	})

	// Leave → -StateHover -StateActive -StatePressed.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		w.ChangeState(0, StateHover|StateActive|StatePressed)
	})
}

// setStyle installs style for the widget, keeping widget-level options
// layered on top of it.
func (w *TtkWidget) setStyle(style *Style) {
	if w.widgetOpts != nil {
		w.widgetOpts.Parent = style
		w.Context.Style = w.widgetOpts
		return
	}
	w.Context.Style = style
}

// SetWidgetOption sets a widget-level element option such as "-padding",
// "-wraplength", "-justify" or "-anchor"; it takes precedence over the
// style, as Tk's widget options do. Call RequestSize afterwards.
func (w *TtkWidget) SetWidgetOption(name string, value any) {
	if w.Context == nil || w.Context.Style == nil {
		return
	}
	if w.widgetOpts == nil {
		w.widgetOpts = &Style{
			Name:     w.Context.Style.Name,
			Parent:   w.Context.Style,
			Defaults: map[string]any{},
			Maps:     map[string]StateMap[any]{},
		}
		w.Context.Style = w.widgetOpts
	}
	w.widgetOpts.Defaults[name] = value
}

// classForStyle derives the Tk widget class from a style name, e.g.
// "Vertical.TScrollbar" -> "TScrollbar". ttk::treeview's class is "Treeview".
func classForStyle(styleName string) string {
	parts := strings.Split(styleName, ".")
	class := parts[len(parts)-1]
	for _, p := range parts {
		if len(p) > 1 && p[0] == 'T' && p[1] >= 'A' && p[1] <= 'Z' {
			class = p
			break
		}
	}
	if class == "TTreeview" {
		return "Treeview"
	}
	return class
}
