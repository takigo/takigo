package takigo

import (
	"bytes"
	"context"
	"fmt"
	goimage "image"
	"image/png"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/msorc/takigo/appearance"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/grab"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/internal/treedump"
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
	grabMgr    *grab.Manager
	logger     *slog.Logger
}

// NewApp creates a new takigo application. It opens the X11 display,
// creates the root window, and initializes the event system.
func NewApp(opts ...AppOption) (*App, error) {
	cfg := appConfig{
		displayName: "",
		title:       "takigo",
		logger:      slog.Default(),
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
	widget.SetClassic(root, cfg.classic || os.Getenv("TAKIGO_CLASSIC") == "1")

	// Initialize input method for proper non-Latin keyboard handling.
	d.Server.InitIM(root.PlatformID)

	dispatcher := event.NewDispatcher()
	loop := event.NewLoop(server, parser, dispatcher)
	d.DoWhenIdle = loop.DoWhenIdle

	colors := color.NewCache(d.Screen)
	if cfg.followSystem {
		cfg.appearance = appearance.System()
	}
	if cfg.appearance == appearance.Dark {
		widget.SetPalette(root, widget.DarkPalette)
		if bg, err := colors.Get(widget.DarkPalette.Background); err == nil {
			root.SetBackgroundPixel(bg.Pixel)
		}
	}
	fontReg := font.NewRegistry(fontOpener)

	bindEng := bind.NewEngine(d)

	selMgr := selection.NewManager(server, dispatcher)
	selMgr.SetTimer(func(d time.Duration, fn func()) { loop.After(d, fn) })

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
		grabMgr:    grab.NewManager(server, dispatcher),
		logger:     cfg.logger,
	}

	// Tk never reads a child window's size back from X: the geometry
	// managers own it. A queued ConfigureNotify can describe a size that has
	// since been replaced, so report the current one to every handler.
	loop.SetEventFilter(func(ev *event.Event) {
		switch ev.Type {
		case event.ConfigureType:
			w := d.LookupWindow(ev.Window)
			switch {
			case w == nil:
			case !w.IsTopLevel():
				ev.ConfigWidth, ev.ConfigHeight = w.Width, w.Height
			case w.WmData != nil:
				// A size the user dragged to becomes the toplevel's
				// geometry (ConfigureEvent in tkUnixWm.c).
				w.WmData.ConfigureNotify(ev.ConfigWidth, ev.ConfigHeight)
			}
		case event.FocusInType, event.FocusOutType:
			// Tk's focus model: the manager turns toplevel focus changes
			// into FocusIn/FocusOut on its focus windows and drops the
			// rest (TkFocusFilterEvent).
			if app.focusMgr != nil && !app.focusMgr.FilterEvent(ev) {
				ev.Type = 0
			}
		case event.KeyPressType, event.KeyReleaseType, event.ButtonPressType, event.ButtonReleaseType,
			event.MotionType, event.EnterType, event.LeaveType, event.MouseWheelType:
			// Keys go to the focus window (TkFocusKeyEvent), and leaving
			// a toplevel can end an implicit focus.
			if app.focusMgr != nil && (ev.Type == event.KeyPressType || ev.Type == event.KeyReleaseType || ev.Type == event.LeaveType) {
				app.focusMgr.FilterEvent(ev)
			}
			// A local grab (tkGrab.c's TkPointerEvent) discards input
			// for this application's windows outside the grab tree.
			if app.grabMgr.Current() == nil {
				return
			}
			if w := d.LookupWindow(ev.Window); w != nil && app.grabMgr.ShouldRedirect(w) {
				ev.Type = 0
			}
		}
	})

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
		case platform.PropertyNotifyEvent:
			prop := parser.ParsePropertyEvent(raw)
			selMgr.HandlePropertyNotify(prop.EventWindow, prop.Atom, prop.Deleted)
		}
	})

	// Initialize WM state for root window, same as Tk does for ".".
	// This sets WM_CLASS, WM_HINTS, size hints, and WM_PROTOCOLS.
	app.wmInfo = wm.Init(root)
	app.wmInfo.SetTitle(cfg.title)

	// Apply geometry string if provided (overrides Size).
	if cfg.geometry != "" {
		if err := app.wmInfo.SetGeometry(cfg.geometry); err != nil {
			cfg.logger.Warn("bad geometry", "geometry", cfg.geometry, "err", err)
		}
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
			root.NotifyConfigure()
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

	// Tk_DestroyWindow delivers <Destroy> and then forgets the window:
	// its event handlers, bind tags and focus state go with it, so their
	// closures do not leak and a reused window ID starts clean.
	d.OnWindowDestroy(func(w *window.Window) {
		if w.PlatformID == 0 {
			return
		}
		dispatcher.Dispatch(&event.Event{Type: event.DestroyType, Window: w.PlatformID})
		dispatcher.Unbind(w.PlatformID)
		bindEng.UnregisterWindow(w)
		focusMgr.HandleDestroyWindow(w)
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
		if info := w.WmData; info != nil {
			info.HandleClientMessage(ev.MessageType, ev.MessageData)
		}
	})

	return app, nil
}

// Root returns the root window of the application.
func (a *App) Root() *window.Window {
	return a.root
}

// Appearance reports whether the desktop currently asks for a light or a
// dark look; see package appearance.
func (a *App) Appearance() appearance.Mode {
	return appearance.System()
}

// Lookup returns the window with the given Tk path name (".frame.ok"), or
// nil if there is none.
func (a *App) Lookup(path string) *window.Window {
	return a.root.Lookup(path)
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

// RunContext is Run that also quits when ctx is done. It returns nil when
// the application quit by itself and the context's cause otherwise.
func (a *App) RunContext(ctx context.Context) error {
	stop := context.AfterFunc(ctx, a.Quit)
	defer stop()
	a.Run()
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	return nil
}

// MainLoop maps the root window and runs the event loop.
// It blocks until Quit is called.
func (a *App) MainLoop() {
	// Lay out before mapping, as Tk maps "." at idle time after geometry
	// propagation, so the window appears at its final size.
	a.loop.UpdateIdleTasks()
	window.SyncBackground(a.root)
	a.display.Server.MapWindow(a.root.PlatformID)
	window.MarkMapped(a.root)
	a.display.Server.Flush()
	if path := os.Getenv("TAKIGO_DUMP_TREE"); path != "" {
		a.startTreeDump(path)
	}
	a.loop.Run()
}

// startTreeDump rewrites the widget-tree JSON at path whenever it changes,
// polling from a plain ticker so TAKIGO_FREEZE_TIMERS does not stop it.
// Used by scripts/demo_screenshot.sh for structural comparison with Tk.
func (a *App) startTreeDump(path string) {
	var last []byte
	dump := func() {
		b, err := treedump.Marshal(treedump.Collect(a.display, a.fontReg))
		if err != nil || bytes.Equal(b, last) {
			return
		}
		if err := treedump.WriteFileAtomic(path, b); err != nil {
			a.logger.Warn("tree dump failed", "path", path, "err", err)
			return
		}
		last = b
	}
	go func() {
		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				a.loop.RunOnMain(dump)
			case <-a.loop.Done():
				return
			}
		}
	}()
}

// Quit stops the event loop and cleans up resources.
func (a *App) Quit() {
	a.loop.Quit()
}

// Done returns a channel that is closed once Quit has been called. A
// goroutine waiting for the result of a RunOnMain callback selects on it
// too: callbacks still queued when the loop quits never run.
func (a *App) Done() <-chan struct{} {
	return a.loop.Done()
}

// Destroy cleans up all resources. Call after MainLoop returns.
func (a *App) Destroy() {
	// The reader goroutine may be blocked reading this display; closing it
	// underneath the read is a use-after-free in Xlib.
	a.loop.Stop(time.Second)
	// Destroy handlers may still draw or measure text, so the windows go
	// before the images and fonts they use (as in Tk's DeleteWindowsExitProc).
	window.DestroyWindow(a.root)
	if a.imageReg != nil {
		a.imageReg.DestroyAll()
	}
	if a.fontReg != nil {
		a.fontReg.Close()
	}
	a.display.Close()
}

// Logger returns the application's logger; see WithLogger.
func (a *App) Logger() *slog.Logger {
	return a.logger
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

// GrabManager returns the application's grab manager; a local grab set on
// it confines input to the grab window's subtree, as Tk's grab does.
func (a *App) GrabManager() *grab.Manager {
	return a.grabMgr
}

// Bind returns the application's binding engine (Tk's "bind" and
// "bindtags").
func (a *App) Bind() *bind.Engine {
	return a.bindEng
}

// Clipboard returns the application's clipboard manager.
func (a *App) Clipboard() widget.ClipboardManager {
	return appClipboard{mgr: a.selMgr}
}

// appClipboard adapts selection.Manager to widget.ClipboardManager. It is
// pointer-shaped, so returning it as an interface does not allocate.
type appClipboard struct {
	mgr *selection.Manager
}

func (c appClipboard) Set(owner platform.WindowID, text string, time platform.Timestamp) {
	c.mgr.OwnClipboard(owner, text, time)
}

func (c appClipboard) Get(requestor platform.WindowID, time platform.Timestamp, callback func(string)) {
	c.mgr.RequestWithCallback(requestor, time, callback)
}

// SetClipboardImage puts img on the clipboard, as PNG. Other applications
// can paste it on X11; on Windows and macOS only this process sees it,
// since their native clipboards are used for text alone.
func (a *App) SetClipboardImage(img goimage.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return fmt.Errorf("takigo: clipboard image: %w", err)
	}
	a.selMgr.OwnFormats(a.selMgr.ClipboardAtom(), a.root.PlatformID, map[platform.AtomID][]byte{
		a.display.Server.InternAtom("image/png", false): buf.Bytes(),
	}, platform.CurrentTime)
	return nil
}

// ClipboardImage asks for the image on the clipboard and calls callback
// with it, or with nil when the clipboard holds no PNG image. The callback
// runs on the event loop, at once if this App owns the clipboard.
func (a *App) ClipboardImage(callback func(goimage.Image)) {
	target := a.display.Server.InternAtom("image/png", false)
	a.selMgr.RequestTarget(a.selMgr.ClipboardAtom(), target, a.root.PlatformID, platform.CurrentTime, func(data []byte) {
		if img, err := png.Decode(bytes.NewReader(data)); err == nil {
			callback(img)
			return
		}
		callback(nil)
	})
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

// RunNestedLoopContext processes events until the context is cancelled or done is closed.
// This is the context-aware version of RunNestedLoop for cancellation support.
// Used by modal dialogs to keep the event loop alive while blocking.
func (a *App) RunNestedLoopContext(ctx context.Context, done <-chan struct{}) {
	a.loop.RunNestedContext(ctx, done)
}

// RegisterCloseHandler registers a WM_DELETE_WINDOW handler for a toplevel window.
// It routes through the window's WmInfo if available.
func (a *App) RegisterCloseHandler(w platform.WindowID, fn func()) {
	win := a.display.LookupWindow(w)
	if win == nil {
		return
	}
	if info := win.WmData; info != nil {
		info.OnDeleteWindow(fn)
	}
}

// UnregisterCloseHandler restores the default WM_DELETE_WINDOW
// behavior — destroying the window — for the given toplevel.
// Equivalent to Tk's behavior when no user handler is installed.
func (a *App) UnregisterCloseHandler(w platform.WindowID) {
	win := a.display.LookupWindow(w)
	if win == nil {
		return
	}
	if info := win.WmData; info != nil {
		info.OffDeleteWindow()
	}
}

// UpdateIdleTasks runs pending idle work (layout, redraws) now, like Tk's
// "update idletasks". Call it from the event loop goroutine, or before
// Run, e.g. to read a widget's size right after packing it.
func (a *App) UpdateIdleTasks() {
	a.loop.UpdateIdleTasks()
}

// DoWhenIdle schedules a function to run during the next idle phase.
func (a *App) DoWhenIdle(fn func()) {
	a.loop.DoWhenIdle(fn)
}

// After schedules a function to run after the given duration on the main
// goroutine. The returned function cancels it; see event.Loop.After.
func (a *App) After(d time.Duration, fn func()) (cancel func() bool) {
	return a.loop.After(d, fn)
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
	logger      *slog.Logger
	classic     bool

	appearance   appearance.Mode
	followSystem bool
}

// UseAppearance gives the App a light or a dark look: the classic widgets
// start with the matching default colours (widget.DarkPalette) and themed
// widgets use the "dark" theme when it is registered (import
// ttk/darktheme). The default is light, which is Tk's look.
func UseAppearance(mode appearance.Mode) AppOption {
	return func(c *appConfig) { c.appearance, c.followSystem = mode, false }
}

// FollowSystemAppearance is UseAppearance with the desktop's current
// preference (appearance.System), read once when the App is created.
func FollowSystemAppearance() AppOption {
	return func(c *appConfig) { c.followSystem = true }
}

// Classic makes the App draw and behave exactly as Tk does, giving up the
// departures takigo makes for a better result, such as anti-aliased canvas
// items. The environment variable TAKIGO_CLASSIC=1 does the same for a
// program that does not ask for it.
func Classic() AppOption {
	return func(c *appConfig) { c.classic = true }
}

// WithLogger sets the logger for the App's diagnostics, such as an option
// that could not be applied in a constructor. The default is slog.Default.
func WithLogger(l *slog.Logger) AppOption {
	return func(c *appConfig) {
		if l != nil {
			c.logger = l
		}
	}
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
	for line := range strings.SplitSeq(resources, "\n") {
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
