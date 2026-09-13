// Package widget defines the Widget interface and Base struct shared by
// all takigo widgets. It ports the common widget lifecycle from
// tk/generic/tkButton.c and tk/generic/tkFrame.c.
package widget

import (
	"context"
	"time"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// WidgetImage is the interface for images that can be displayed in widgets.
// It avoids widgets importing the image package directly.
type WidgetImage interface {
	Width() int
	Height() int
	Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
		depth int,
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

// Caregiver is implemented by anything that can serve as a parent for
// widget construction — the App (for root-level widgets) or any widget
// (for nested widgets). It provides both a parent window and the
// application services needed during widget initialization.
type Caregiver interface {
	window.Windower
	AppContext() AppContext
}

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
	Background          *color.Color
	Foreground          *color.Color
	HighlightBackground *color.Color
	HighlightColor      *color.Color
	Font                font.Font
	Relief              option.Relief
	BorderWidth         int
	HighlightWidth      int
	PadX, PadY          int

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

// ClipboardManager provides clipboard read/write for widgets.
type ClipboardManager interface {
	// Set stores text as the CLIPBOARD owner.
	Set(owner platform.WindowID, text string, time platform.Timestamp)
	// Get retrieves clipboard text. If this process owns the clipboard the
	// callback is called immediately; otherwise an async X11 request is sent
	// and the callback fires when the SelectionNotify arrives.
	Get(requestor platform.WindowID, time platform.Timestamp, callback func(string))
}

// AppContext provides the application services widgets need.
// This avoids importing the top-level takigo package.
type AppContext interface {
	AppContext() AppContext
	Window() *window.Window
	Dispatcher() *event.Dispatcher
	DoWhenIdle(fn func())
	ColorCache() *color.Cache
	FontRegistry() *font.Registry
	Server() platform.DisplayServer
	ImageRegistry() *image.Registry
	BindEngine() BindEngine
	// RunNestedLoop processes events until done is closed.
	// Used by modal dialogs to keep the event loop alive while blocking.
	RunNestedLoop(done <-chan struct{})
	// RunNestedLoopContext processes events until the context is cancelled or done is closed.
	// Context-aware version for cancellation support.
	RunNestedLoopContext(ctx context.Context, done <-chan struct{})
	// RegisterCloseHandler registers a WM_DELETE_WINDOW handler for a toplevel window.
	RegisterCloseHandler(w platform.WindowID, fn func())
	// UnregisterCloseHandler removes a WM_DELETE_WINDOW handler.
	UnregisterCloseHandler(w platform.WindowID)
	// After schedules a function to run after a delay.
	After(d time.Duration, fn func())
	Quit()
	// Clipboard returns the application clipboard manager.
	Clipboard() ClipboardManager
}

// Window returns the widget's underlying window.
func (b *Base) Window() *window.Window {
	return b.Win
}

// To be used by geometry managers as a geometry.Elementer
func (b *Base) GeometryElements() []window.Windower {
	return []window.Windower{b}
}

// AppContext returns the application context, satisfying the Caregiver interface.
func (b *Base) AppContext() AppContext {
	return b.App
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
	if w.PlatformID == 0 {
		return
	}

	d := w.Display.Server
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

// DrawHighlightBorder draws the focus highlight ring around the widget.
// When the widget has focus, HighlightColor is used; otherwise HighlightBackground.
// The padding parameter specifies pixels of padding outside the highlight ring
// (used by buttons with default rings).
func (b *Base) DrawHighlightBorder(focused bool, padding int) {
	w := b.Win
	if w.PlatformID == 0 || b.HighlightWidth <= 0 {
		return
	}

	d := w.Display.Server
	gc := w.GC

	var pixel uint64
	if focused && b.HighlightColor != nil {
		pixel = b.HighlightColor.Pixel
	} else if b.HighlightBackground != nil {
		pixel = b.HighlightBackground.Pixel
	} else {
		return
	}
	d.SetForeground(gc, pixel)

	hlw := b.HighlightWidth
	ww := w.Width
	wh := w.Height
	drawable := w.Drawable()

	// Top
	d.FillRectangle(drawable, gc, padding, padding,
		uint(ww-2*padding), uint(hlw))
	// Bottom
	d.FillRectangle(drawable, gc, padding, wh-hlw-padding,
		uint(ww-2*padding), uint(hlw))
	// Left
	d.FillRectangle(drawable, gc, padding, padding+hlw,
		uint(hlw), uint(wh-2*hlw-2*padding))
	// Right
	d.FillRectangle(drawable, gc, ww-hlw-padding, padding+hlw,
		uint(hlw), uint(wh-2*hlw-2*padding))
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
	b.HighlightBackground, _ = cache.Get(DefHighlightBg)
	b.HighlightColor, _ = cache.Get(DefHighlightColor)
	b.UpdateBorder()

	// Get default font.
	reg := app.FontRegistry()
	b.Font, _ = reg.Get(font.TkDefaultFont)

	// Register background hook so ApplyBackgroundRecursive can update this widget.
	win.BackgroundHook = func(colorName string) {
		if c, err := app.ColorCache().Get(colorName); err == nil {
			b.Background = c
			b.Win.BackgroundPixel = c.Pixel
			b.UpdateBorder()
		}
	}
}

// AnchorText computes the x,y position for content of size (textW x textH)
// within a frame at (frameX, frameY) of size (frameW x frameH) according
// to the given anchor.
func AnchorText(a option.Anchor, frameX, frameY, frameW, frameH, textW, textH int) (int, int) {
	var x, y int
	switch a {
	case option.AnchorNW:
		x, y = frameX, frameY
	case option.AnchorN:
		x, y = frameX+(frameW-textW)/2, frameY
	case option.AnchorNE:
		x, y = frameX+frameW-textW, frameY
	case option.AnchorW:
		x, y = frameX, frameY+(frameH-textH)/2
	case option.AnchorCenter:
		x, y = frameX+(frameW-textW)/2, frameY+(frameH-textH)/2
	case option.AnchorE:
		x, y = frameX+frameW-textW, frameY+(frameH-textH)/2
	case option.AnchorSW:
		x, y = frameX, frameY+frameH-textH
	case option.AnchorS:
		x, y = frameX+(frameW-textW)/2, frameY+frameH-textH
	case option.AnchorSE:
		x, y = frameX+frameW-textW, frameY+frameH-textH
	}
	return x, y
}

// CompoundSize returns the bounding-box width and height of a compound
// image+text layout, given the layout mode, an optional image, and the
// measured text size. A nil image degrades to the text size; empty text
// degrades to the image size; the four-pixel gap is the conventional
// Tk separator between image and text.
//
// Matches the geometry expected by -compound on label, button, and ttk
// widgets, so callers in widget/button and ttk/elements share it.
func CompoundSize(c Compound, img WidgetImage, textW, textH int) (int, int) {
	if img == nil {
		return textW, textH
	}
	imgW := img.Width()
	imgH := img.Height()

	if textW == 0 && textH == 0 {
		return imgW, imgH
	}

	switch c {
	case CompoundLeft, CompoundRight:
		return imgW + 4 + textW, max(imgH, textH)
	case CompoundTop, CompoundBottom:
		return max(imgW, textW), imgH + 4 + textH
	case CompoundCenter:
		return max(imgW, textW), max(imgH, textH)
	default:
		return imgW, imgH
	}
}
