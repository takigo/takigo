// Package square implements the square demo widget, which displays a
// movable/resizable colored square. It ports tk/generic/tkSquare.c.
package square

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// Square displays a movable/resizable colored square within a widget window.
// Used primarily as an example of custom widget implementation.
type Square struct {
	widget.Base

	// Position of square's upper-left corner within the widget.
	PosX, PosY int
	// Width and height of the square in pixels.
	Size int
	// Foreground border color (drawn as the square itself).
	FgBorder *draw.Border
	// Whether to use double-buffered rendering.
	DoubleBuffer bool
}

// SquareOption configures a Square widget.
type SquareOption func(*Square)

// Background sets the background color.
func Background(name string) SquareOption {
	return func(s *Square) { s.SetBackgroundName(name) }
}

// Foreground sets the square's color.
func Foreground(name string) SquareOption {
	return func(s *Square) {
		if s.SetForegroundName(name) {
			col := s.Foreground
			s.FgBorder = draw.NewBorder(col.Red, col.Green, col.Blue)
		}
	}
}

// BorderWidthOpt sets the 3D border width.
func BorderWidthOpt(w any) SquareOption {
	return func(s *Square) { s.BorderWidth = screenunit.PxOr(w, s.BorderWidth) }
}

// Relief sets the border relief.
func Relief(r option.Relief) SquareOption {
	return func(s *Square) { s.Relief = r }
}

// PosXOpt sets the X position of the square.
func PosXOpt(x int) SquareOption {
	return func(s *Square) { s.PosX = x }
}

// PosYOpt sets the Y position of the square.
func PosYOpt(y int) SquareOption {
	return func(s *Square) { s.PosY = y }
}

// SizeOpt sets the width and height of the square.
func SizeOpt(sz int) SquareOption {
	return func(s *Square) { s.Size = sz }
}

// DoubleBufferOpt enables or disables double buffering.
func DoubleBufferOpt(on bool) SquareOption {
	return func(s *Square) { s.DoubleBuffer = on }
}

// New creates a new Square widget as a child of parent.
func New(parent widget.Caregiver, name string, opts ...SquareOption) *Square {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	s := &Square{
		PosX:         0,
		PosY:         0,
		Size:         20,
		DoubleBuffer: true,
	}
	widget.InitBase(&s.Base, w, app)
	w.Class = "Square"

	// Square-specific defaults from tkSquare.c.
	s.BorderWidth = 2
	s.Relief = option.ReliefRaised

	// Default foreground (#b03060 = maroon-ish).
	if col, err := app.ColorCache().Get("#b03060"); err == nil {
		s.Foreground = col
		s.FgBorder = draw.NewBorder(col.Red, col.Green, col.Blue)
	}

	for _, opt := range opts {
		opt(s)
	}

	// Default requested size: 200x150 (from tkSquare.c).
	w.ReqWidth = 200
	w.ReqHeight = 150

	s.keepInWindow()

	if s.Background != nil {
		w.SetBackgroundPixel(s.Background.Pixel)
	}

	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		s.Display()
	})

	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			s.keepInWindow()
			s.Display()
		}
	})

	return s
}

// keepInWindow constrains the square's position so it stays within the widget.
func (s *Square) keepInWindow() {
	w := s.Win
	bd := s.BorderWidth

	maxX := w.Width - bd - s.Size
	maxY := w.Height - bd - s.Size

	if s.PosX > maxX {
		s.PosX = maxX
	}
	if s.PosX < bd {
		s.PosX = bd
	}
	if s.PosY > maxY {
		s.PosY = maxY
	}
	if s.PosY < bd {
		s.PosY = bd
	}
}

// SetPosition sets the square's position and redraws.
func (s *Square) SetPosition(x, y int) {
	s.PosX = x
	s.PosY = y
	s.keepInWindow()
	s.Display()
}

// Display draws the square widget.
func (s *Square) Display() {
	if s.Destroyed {
		return
	}
	w := s.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	var drawable platform.DrawableID
	var pixmap platform.PixmapID

	if s.DoubleBuffer {
		pixmap = d.CreatePixmap(w.Drawable(), uint(w.Width), uint(w.Height), uint(w.Depth))
		if pixmap == 0 {
			return
		}
		drawable = platform.DrawableID(pixmap)
	} else {
		drawable = w.Drawable()
	}

	// Fill background.
	if s.Background != nil {
		d.SetForeground(gc, s.Background.Pixel)
	}
	d.FillRectangle(drawable, gc, 0, 0, uint(w.Width), uint(w.Height))

	// Draw background 3D border.
	if s.Border != nil && s.BorderWidth > 0 && s.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, drawable, gc, s.Border,
			0, 0, w.Width, w.Height, s.BorderWidth, s.Relief)
	}

	// Draw the foreground square with raised relief.
	if s.FgBorder != nil && s.Size > 0 {
		if s.Foreground != nil {
			d.SetForeground(gc, s.Foreground.Pixel)
		}
		d.FillRectangle(drawable, gc, s.PosX, s.PosY, uint(s.Size), uint(s.Size))
		draw.Draw3DRectangle(d, drawable, gc, s.FgBorder,
			s.PosX, s.PosY, s.Size, s.Size, 1, option.ReliefRaised)
	}

	if s.DoubleBuffer {
		d.CopyArea(drawable, w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height), 0, 0)
		d.FreePixmap(pixmap)
	}

	d.Flush()
}

// Configure applies options to the square.
func (s *Square) Configure(opts ...SquareOption) {
	widget.Configure(s, opts, s.keepInWindow)
}

// Destroy cleans up the square widget.
func (s *Square) Destroy() {
	if s.Destroyed {
		return
	}
	s.Destroyed = true
	window.DestroyWindow(s.Win)
}
