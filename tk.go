// Package takigo is a pure Go port of the Tk GUI toolkit.
// It provides a functional-options API for building cross-platform
// graphical applications, initially targeting X11/Linux.
package takigo

import (
	"fmt"
	"time"

	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// App is the top-level application, managing the display connection,
// event loop, and root window. It implements widget.AppContext.
type App struct {
	display    *window.Display
	root       *window.Window
	dispatcher *event.Dispatcher
	loop       *event.Loop
	colorCache *color.Cache
	fontReg    *font.Registry
	imageReg   *image.Registry
	bindEng    *bind.Engine
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

	d, err := window.NewDisplay(cfg.displayName)
	if err != nil {
		return nil, fmt.Errorf("takigo: %w", err)
	}

	root := window.CreateMainWindow(d, 0, 0, cfg.width, cfg.height)
	d.XDisplay.StoreName(root.XWindow, cfg.title)

	// Initialize X Input Method for proper non-Latin keyboard handling.
	d.XDisplay.InitIM(root.XWindow)

	dispatcher := event.NewDispatcher()
	loop := event.NewLoop(d.XDisplay, dispatcher)

	colors := color.NewCache(d.XDisplay, d.Screen, d.Colormap)
	fontReg := font.NewRegistry(d.XDisplay, d.Screen, d.Visual, d.Colormap)

	bindEng := bind.NewEngine(d)

	app := &App{
		display:    d,
		root:       root,
		dispatcher: dispatcher,
		loop:       loop,
		colorCache: colors,
		fontReg:    fontReg,
		imageReg:   image.NewRegistry(),
		bindEng:    bindEng,
	}

	// Install bind engine as a global handler (fires after per-window handlers).
	bindEng.Install(dispatcher)

	// Handle WM_DELETE_WINDOW (window close button).
	dispatcher.BindGlobal(event.AllEventsMask, func(ev *event.Event) {
		if ev.Type == event.ClientMessageType {
			if xlib.Atom(ev.MessageData[0]) == d.WMDeleteWindow {
				app.Quit()
			}
		}
	})

	return app, nil
}

// Root returns the root window of the application.
func (a *App) Root() *window.Window {
	return a.root
}

// Display returns the display.
func (a *App) Display() *window.Display {
	return a.display
}

// Dispatcher returns the event dispatcher.
func (a *App) Dispatcher() *event.Dispatcher {
	return a.dispatcher
}

// MainLoop maps the root window and runs the event loop.
// It blocks until Quit is called.
func (a *App) MainLoop() {
	a.display.XDisplay.MapWindow(a.root.XWindow)
	a.root.Flags |= window.FlagMapped
	a.display.XDisplay.Flush()
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

// DisplayPtr returns the underlying xlib.Display pointer.
func (a *App) DisplayPtr() *xlib.Display {
	return a.display.XDisplay
}

// BindEngine returns the application's binding engine.
func (a *App) BindEngine() widget.BindEngine {
	return a.bindEng
}

// BindEng returns the full bind.Engine for direct access.
func (a *App) BindEng() *bind.Engine {
	return a.bindEng
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
}

// DisplayName sets the X11 display name (e.g., ":0").
func DisplayName(name string) AppOption {
	return func(c *appConfig) { c.displayName = name }
}

// Title sets the window title.
func Title(title string) AppOption {
	return func(c *appConfig) { c.title = title }
}

// Size sets the initial window size.
func Size(width, height int) AppOption {
	return func(c *appConfig) {
		c.width = width
		c.height = height
	}
}
