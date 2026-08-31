// Package takigo is a pure Go port of the Tk GUI toolkit.
// It provides a functional-options API for building cross-platform
// graphical applications, initially targeting X11/Linux.
package takigo

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/selection"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
	"github.com/msorc/takigo/wm"
)

// App is the top-level application, managing the display connection,
// event loop, and root window. It implements widget.AppContext.
type App struct {
	display    *window.Display
	root       *window.Window
	wmInfo     *wm.WmInfo
	dispatcher *event.Dispatcher
	loop       *event.Loop
	colorCache *color.Cache
	fontReg    *font.Registry
	imageReg   *image.Registry
	bindEng    *bind.Engine
	focusMgr   *focus.Manager
	selMgr     *selection.Manager
}

// NewApp creates a new takigo application. It opens the X11 display,
// creates the root window, and initializes the event system.
func NewApp(opts ...AppOption) (*App, error) {
	cfg := appConfig{
		displayName: "",
		title:       "takigo",
		width:       400,
		height:      300,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Create platform-specific display server, event parser, and font opener.
	server, parser, fontOpener, err := platformInit(cfg.displayName)
	if err != nil {
		return nil, fmt.Errorf("takigo: %w", err)
	}

	// Configure screen unit conversion from actual screen metrics.
	defScreen := server.DefaultScreen()
	xftDPI, hasXftDPI := parseXftDPI(server.ResourceManagerString())
	var dpiArg float64
	if hasXftDPI {
		dpiArg = xftDPI
	}
	screenunit.SetScreenDPI(server.ScreenWidth(defScreen), server.ScreenWidthMM(defScreen), dpiArg)

	d, err := window.NewDisplay(server)
	if err != nil {
		server.Close()
		return nil, fmt.Errorf("takigo: %w", err)
	}

	root := window.CreateMainWindow(d, 0, 0, cfg.width, cfg.height)

	// Initialize input method for proper non-Latin keyboard handling.
	d.Server.InitIM(root.PlatformID)

	dispatcher := event.NewDispatcher()
	loop := event.NewLoop(server, parser, dispatcher)

	colors := color.NewCache(d.Screen)
	fontReg := font.NewRegistry(fontOpener)

	bindEng := bind.NewEngine(d)

	selMgr := selection.NewManager(server, dispatcher)

	app := &App{
		display:    d,
		root:       root,
		dispatcher: dispatcher,
		loop:       loop,
		colorCache: colors,
		fontReg:    fontReg,
		imageReg:   image.NewRegistry(),
		bindEng:    bindEng,
		selMgr:     selMgr,
	}

	// Handle selection events (clipboard serve + async paste response).
	loop.SetRawEventHandler(func(raw *platform.RawEvent) {
		switch raw.EventType {
		case platform.SelectionRequestEvent:
			req := parser.ParseSelectionRequestEvent(raw)
			selMgr.HandleSelectionRequest(req.Requestor, req.Selection, req.Target, req.Property, req.Time)
		case platform.SelectionClearEvent:
			clr := parser.ParseSelectionClearEvent(raw)
			selMgr.HandleSelectionClear(clr.Selection)
		case platform.SelectionNotifyEvent:
			ntf := parser.ParseSelectionNotifyEvent(raw)
			selMgr.HandleSelectionNotify(ntf.Requestor, ntf.Property)
		}
	})

	// Initialize WM state for root window, same as Tk does for ".".
	// This sets WM_CLASS, WM_HINTS, size hints, and WM_PROTOCOLS.
	app.wmInfo = wm.Init(root)
	app.wmInfo.SetTitle(cfg.title)

	// Apply geometry string if provided (overrides Size).
	if cfg.geometry != "" {
		app.wmInfo.SetGeometry(cfg.geometry)
	}

	// Apply icon name if provided.
	if cfg.iconName != "" {
		app.wmInfo.SetIconName(cfg.iconName)
	}

	// Default close action: quit the application.
	app.wmInfo.OnDeleteWindow(func() {
		app.Quit()
	})

	// Handle ConfigureNotify on root window so geometry managers
	// (pack/grid) re-layout when the window is resized.
	dispatcher.Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			if root.ConfigureCallback != nil {
				root.ConfigureCallback()
			}
		}
	})

	// Install bind engine as a global handler (fires after per-window handlers).
	bindEng.Install(dispatcher)

	// Set up global focus manager with Tab/Shift-Tab traversal,
	// matching Tk's "bind all <<NextWindow>>" / "bind all <<PrevWindow>>"
	// from tk.tcl.
	focusMgr := focus.NewManager(dispatcher, server, d)
	focusMgr.BindTraversal(root)
	app.focusMgr = focusMgr

	// Route real X FocusIn events on toplevels to the focus manager.
	// This marks the toplevel as viewable (WM has confirmed it), which
	// allows SetInputFocus to be called safely on child widgets.
	// Only process FocusIn (not FocusOut) to avoid interfering with
	// synthetic FocusOut events dispatched internally by SetFocus.
	dispatcher.BindGlobal(event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type != event.FocusInType {
			return
		}
		w := d.LookupWindow(ev.Window)
		if w == nil || !w.IsTopLevel() {
			return
		}
		focusMgr.HandleFocusIn(w)
	})

	// Route WM protocol messages (WM_DELETE_WINDOW, _NET_WM_PING, etc.)
	// to the appropriate toplevel's WmInfo handler.
	// This mirrors TkWmProtocolEventProc in tk/unix/tkUnixWm.c.
	dispatcher.BindGlobal(event.ClientMessageMask, func(ev *event.Event) {
		if ev.Type != event.ClientMessageType {
			return
		}
		// Look up the window and dispatch via its WmInfo.
		w := d.LookupWindow(ev.Window)
		if w == nil {
			return
		}
		if info, ok := w.WmData.(*wm.WmInfo); ok {
			info.HandleClientMessage(ev.MessageType, ev.MessageData)
		}
	})

	return app, nil
}

// Root returns the root window of the application.
func (a *App) Root() *window.Window {
	return a.root
}

// Window returns the root window, satisfying the widget.Caregiver interface.
func (a *App) Window() *window.Window {
	return a.root
}

// AppContext returns the App itself as a widget.AppContext, satisfying
// the widget.Caregiver interface.
func (a *App) AppContext() widget.AppContext {
	return a
}

// Display returns the display.
func (a *App) Display() *window.Display {
	return a.display
}

// Dispatcher returns the event dispatcher.
func (a *App) Dispatcher() *event.Dispatcher {
	return a.dispatcher
}

// Run maps the root window, runs the event loop, and cleans up resources
// when done. It blocks until Quit is called.
func (a *App) Run() {
	a.MainLoop()
	a.Destroy()
}

// MainLoop maps the root window and runs the event loop.
// It blocks until Quit is called.
func (a *App) MainLoop() {
	a.display.Server.MapWindow(a.root.PlatformID)
	a.root.Flags |= window.FlagMapped
	a.display.Server.Flush()
	a.loop.Run()
}

// Quit stops the event loop and cleans up resources.
func (a *App) Quit() {
	a.loop.Quit()
}

// Destroy cleans up all resources. Call after MainLoop returns.
func (a *App) Destroy() {
	if a.imageReg != nil {
		a.imageReg.DestroyAll()
	}
	if a.fontReg != nil {
		a.fontReg.Close()
	}
	window.DestroyWindow(a.root)
	a.display.Close()
}

// ColorCache returns the application's color cache.
func (a *App) ColorCache() *color.Cache {
	return a.colorCache
}

// FontRegistry returns the application's font registry.
func (a *App) FontRegistry() *font.Registry {
	return a.fontReg
}

// ImageRegistry returns the application's image registry.
func (a *App) ImageRegistry() *image.Registry {
	return a.imageReg
}

// Server returns the platform display server.
func (a *App) Server() platform.DisplayServer {
	return a.display.Server
}

// WmInfo returns the root window's WM state, providing access to
// title, geometry, size constraints, resizable, iconify, etc.
func (a *App) WmInfo() *wm.WmInfo {
	return a.wmInfo
}

// BindEngine returns the application's binding engine.
func (a *App) BindEngine() widget.BindEngine {
	return a.bindEng
}

// BindEng returns the full bind.Engine for direct access.
func (a *App) BindEng() *bind.Engine {
	return a.bindEng
}

// Clipboard returns the application's clipboard manager.
func (a *App) Clipboard() widget.ClipboardManager {
	return &appClipboard{mgr: a.selMgr}
}

// appClipboard adapts selection.Manager to widget.ClipboardManager.
type appClipboard struct {
	mgr *selection.Manager
}

func (c *appClipboard) Set(owner platform.WindowID, text string, time platform.Timestamp) {
	c.mgr.OwnClipboard(owner, text, time)
}

func (c *appClipboard) Get(requestor platform.WindowID, time platform.Timestamp, callback func(string)) {
	c.mgr.RequestWithCallback(requestor, time, callback)
}

// FocusManager returns the application's focus manager.
func (a *App) FocusManager() *focus.Manager {
	return a.focusMgr
}

// RunNestedLoop processes events until done is closed.
// Used by modal dialogs to keep the event loop alive while blocking.
func (a *App) RunNestedLoop(done <-chan struct{}) {
	a.loop.RunNested(done)
}

// RegisterCloseHandler registers a WM_DELETE_WINDOW handler for a toplevel window.
// It routes through the window's WmInfo if available.
func (a *App) RegisterCloseHandler(w platform.WindowID, fn func()) {
	win := a.display.LookupWindow(w)
	if win == nil {
		return
	}
	if info, ok := win.WmData.(*wm.WmInfo); ok {
		info.OnDeleteWindow(fn)
	}
}

// UnregisterCloseHandler removes a WM_DELETE_WINDOW handler.
func (a *App) UnregisterCloseHandler(w platform.WindowID) {
	// No-op: WmInfo always has WM_DELETE_WINDOW in its protocol set.
	// The handler can be overwritten via RegisterCloseHandler.
}

// DoWhenIdle schedules a function to run during the next idle phase.
func (a *App) DoWhenIdle(fn func()) {
	a.loop.DoWhenIdle(fn)
}

// After schedules a function to run after the given duration on the main goroutine.
func (a *App) After(d time.Duration, fn func()) {
	a.loop.After(d, fn)
}

// RunOnMain schedules a function to run on the main (event loop) goroutine.
func (a *App) RunOnMain(fn func()) {
	a.loop.RunOnMain(fn)
}

// AppOption configures a NewApp call.
type AppOption func(*appConfig)

type appConfig struct {
	displayName string
	title       string
	width       int
	height      int
	geometry    string
	iconName    string
}

// DisplayName sets the X11 display name (e.g., ":0").
func DisplayName(name string) AppOption {
	return func(c *appConfig) { c.displayName = name }
}

// Title sets the window title.
func Title(title string) AppOption {
	return func(c *appConfig) { c.title = title }
}

// Geometry sets the root window geometry string (e.g. "800x600+100+50").
func Geometry(geom string) AppOption {
	return func(c *appConfig) { c.geometry = geom }
}

// IconName sets the icon name (WM_ICON_NAME / _NET_WM_ICON_NAME).
func IconName(name string) AppOption {
	return func(c *appConfig) { c.iconName = name }
}

// Size sets the initial window size.
func Size(width, height int) AppOption {
	return func(c *appConfig) {
		c.width = width
		c.height = height
	}
}

// parseXftDPI extracts the Xft.dpi value from an X RESOURCE_MANAGER string.
// Returns (dpi, true) when a positive value is found, (0, false) otherwise.
// The string is newline-separated "key:\tvalue" pairs.
func parseXftDPI(resources string) (float64, bool) {
	for _, line := range strings.Split(resources, "\n") {
		line = strings.TrimSpace(line)
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		if !strings.EqualFold(key, "Xft.dpi") {
			continue
		}
		val := strings.TrimSpace(line[idx+1:])
		dpi, err := strconv.ParseFloat(val, 64)
		if err != nil || dpi <= 0 {
			continue
		}
		return dpi, true
	}
	return 0, false
}
