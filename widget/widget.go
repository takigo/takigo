// Package widget defines the Widget interface and Base struct shared by
// all takigo widgets. It ports the common widget lifecycle from
// tk/generic/tkButton.c and tk/generic/tkFrame.c.
package widget

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

// WidgetImage is the interface for images that can be displayed in widgets.
// It avoids widgets importing the image package directly.
type WidgetImage interface {
	Width() int
	Height() int
	Draw(d *xlib.Display, drawable xlib.Drawable, gc xlib.GC,
		visual *xlib.Visual, depth int,
		imgX, imgY, w, h, dstX, dstY int,
		bgPixel uint64)
}

// Compound specifies how text and image are combined in a widget.
type Compound int

const (
	CompoundNone   Compound = iota // show image if set, else text
	CompoundLeft                   // image left of text
	CompoundRight                  // image right of text
	CompoundTop                    // image above text
	CompoundBottom                 // image below text
	CompoundCenter                 // image behind text
)

// Widget is the interface implemented by all takigo widgets.
type Widget interface {
	// Window returns the widget's underlying window.
	Window() *window.Window

	// Display draws the widget.
	Display()

	// Configure sets options on the widget.
	Configure(opts ...option.Option)

	// Destroy cleans up the widget.
	Destroy()
}

// Base provides common fields and behavior shared by all widgets.
type Base struct {
	Win    *window.Window
	App    AppContext
	Border *draw.Border

	// Common visual options.
	Background      *color.Color
	Foreground      *color.Color
	Font            font.Font
	Relief          option.Relief
	BorderWidth     int
	HighlightWidth  int
	PadX, PadY      int

	// State.
	NeedRedraw bool
	Destroyed  bool
}

// BindEngine is the interface for the binding engine, defined here to
// avoid circular imports between widget and bind packages.
type BindEngine interface {
	RegisterWindow(w *window.Window, className string)
	UnregisterWindow(w *window.Window)
}

// AppContext provides the application services widgets need.
// This avoids importing the top-level takigo package.
type AppContext interface {
	Dispatcher() *event.Dispatcher
	DoWhenIdle(fn func())
	ColorCache() *color.Cache
	FontRegistry() *font.Registry
	DisplayPtr() *xlib.Display
	BindEngine() BindEngine
	// RunNestedLoop processes events until done is closed.
	// Used by modal dialogs to keep the event loop alive while blocking.
	RunNestedLoop(done <-chan struct{})
	// RegisterCloseHandler registers a WM_DELETE_WINDOW handler for a toplevel window.
	RegisterCloseHandler(w xlib.Window, fn func())
	// UnregisterCloseHandler removes a WM_DELETE_WINDOW handler.
	UnregisterCloseHandler(w xlib.Window)
}

// Window returns the widget's underlying window.
func (b *Base) Window() *window.Window {
	return b.Win
}

// UpdateBorder recomputes the 3D border from the background color.
func (b *Base) UpdateBorder() {
	if b.Background != nil {
		b.Border = draw.NewBorder(b.Background.Red, b.Background.Green, b.Background.Blue)
	}
}

// DrawBackground fills the widget's window background and draws the 3D border.
func (b *Base) DrawBackground() {
	w := b.Win
	if w.XWindow == xlib.Window(0) {
		return
	}

	d := w.Display.XDisplay
	gc := w.GC

	if b.Background != nil {
		d.SetForeground(gc, b.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	if b.Border != nil && b.BorderWidth > 0 && b.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, w.Drawable(), gc, b.Border,
			0, 0, w.Width, w.Height, b.BorderWidth, b.Relief)
	}
}

// InitBase initializes common widget fields with defaults.
func InitBase(b *Base, win *window.Window, app AppContext) {
	b.Win = win
	b.App = app
	b.Relief = option.ReliefFlat
	b.BorderWidth = 0
	b.HighlightWidth = 0

	// Get default colors.
	cache := app.ColorCache()
	b.Background, _ = cache.Get(DefBackground)
	b.Foreground, _ = cache.Get(DefForeground)
	b.UpdateBorder()

	// Get default font.
	reg := app.FontRegistry()
	b.Font, _ = reg.Get(font.TkDefaultFont)
}
