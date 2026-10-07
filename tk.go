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

	"github.com/takigo/takigo/appearance"
	"github.com/takigo/takigo/bind"
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/focus"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/grab"
	"github.com/takigo/takigo/image"
	"github.com/takigo/takigo/internal/selection"
	"github.com/takigo/takigo/internal/treedump"
	"github.com/takigo/takigo/internal/xdnd"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
	"github.com/takigo/takigo/wm"
)

// App is the top-level application, managing the display connection,
// event loop, and root window. It implements widget.AppContext.
type App struct {
	display    *window.Display
	root       *window.Window
	wmInfo     *wm.WmInfo
	dispatcher *event.Dispatcher
	parser     platform.EventParser
	loop       *event.Loop
	colorCache *color.Cache
	fontReg    *font.Registry
	imageReg   *image.Registry
	bindEng    *bind.Engine
	focusMgr   *focus.Manager
	selMgr     *selection.Manager
	grabMgr    *grab.Manager
	logger     *slog.Logger
	dnd        *xdnd.Manager
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
		parser:     parser,
		loop:       loop,
		colorCache: colors,
		fontReg:    fontReg,
		imageReg:   image.NewRegistry(),
		bindEng:    bindEng,
		selMgr:     selMgr,
		grabMgr:    grab.NewManager(server, dispatcher),
		logger:     cfg.logger,
	}

	loop.SetEventFilter(app.filterEvent)
	loop.SetRawEventHandler(app.handleRawEvent)
	app.initWM(cfg)

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
	app.focusMgr = focus.NewManager(dispatcher, server, d)
	app.focusMgr.BindTraversal(root)

	d.OnWindowDestroy(app.forgetWindow)
	dispatcher.BindGlobal(event.ClientMessageMask, app.routeClientMessage)

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

// Drop is what another application dropped on a window registered with
// OnDrop; DragData what StartDrag offers.
type (
	Drop     = xdnd.Drop
	DragData = xdnd.DragData
)

// OnDrop makes w accept text and files dragged onto it from other
// applications and calls handler with each drop; a nil handler stops that.
// With several registered windows under the pointer, the innermost gets
// the drop. It works on X11 desktops (XDND).
func (a *App) OnDrop(w window.Windower, handler func(Drop)) {
	a.dragAndDrop().OnDrop(w.Window(), handler)
}

// StartDrag starts dragging data out of w to wherever the user releases the
// mouse button: another application or one of this App's own OnDrop
// windows. Call it from a handler while a button is held on w, typically
// on the first Motion after a ButtonPress. done, which may be nil, is
// called when the drag ends, with whether a target took the data. It works
// on X11 desktops (XDND); elsewhere done is called with false at once.
func (a *App) StartDrag(w window.Windower, data DragData, done func(dropped bool)) {
	a.dragAndDrop().StartDrag(w.Window(), data, done)
}

func (a *App) dragAndDrop() *xdnd.Manager {
	if a.dnd == nil {
		a.dnd = xdnd.New(a.display, a.dispatcher, a.selMgr, func(d time.Duration, fn func()) { a.loop.After(d, fn) })
	}
	return a.dnd
}
