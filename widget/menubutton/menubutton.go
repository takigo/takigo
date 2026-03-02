// Package menubutton implements a menubutton widget that posts
// an associated menu when clicked.
// It ports tk/generic/tkMenubutton.c.
package menubutton

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/window"
)

// Direction specifies where the menu is posted relative to the button.
type Direction int

const (
	Below Direction = iota
	Above
	Left
	Right
)

// Menubutton is a button that posts an associated menu when clicked.
type Menubutton struct {
	widget.Base

	Text      string
	Menu      *menu.Menu
	Direction Direction
	Anchor    option.Anchor

	// Active colors.
	ActiveBg *colorRef
	ActiveFg *colorRef

	// State.
	State widget.State

	textWidth  int
	textHeight int
}

type colorRef struct {
	Pixel uint64
	Red   uint16
	Green uint16
	Blue  uint16
}

// MenubuttonOption configures a Menubutton.
type MenubuttonOption func(*Menubutton)

func Text(s string) MenubuttonOption      { return func(mb *Menubutton) { mb.Text = s } }
func MenuOpt(m *menu.Menu) MenubuttonOption { return func(mb *Menubutton) { mb.Menu = m } }
func DirectionOpt(d Direction) MenubuttonOption { return func(mb *Menubutton) { mb.Direction = d } }
func PadX(p int) MenubuttonOption         { return func(mb *Menubutton) { mb.PadX = p } }
func PadY(p int) MenubuttonOption         { return func(mb *Menubutton) { mb.PadY = p } }

func Background(name string) MenubuttonOption {
	return func(mb *Menubutton) {
		col, err := mb.App.ColorCache().Get(name)
		if err == nil {
			mb.Background = col
			mb.UpdateBorder()
		}
	}
}

func Foreground(name string) MenubuttonOption {
	return func(mb *Menubutton) {
		col, err := mb.App.ColorCache().Get(name)
		if err == nil {
			mb.Foreground = col
		}
	}
}

// New creates a new Menubutton widget.
func New(parent widget.Caregiver, name string, opts ...MenubuttonOption) *Menubutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	mb := &Menubutton{
		Direction: Below,
		Anchor:    option.AnchorCenter,
	}
	widget.InitBase(&mb.Base, w, app)
	mb.BorderWidth = widget.DefBorderWidth
	mb.Relief = option.ReliefRaised
	mb.PadX = 4
	mb.PadY = 2

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		mb.ActiveBg = &colorRef{ac.Pixel, ac.Red, ac.Green, ac.Blue}
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		mb.ActiveFg = &colorRef{af.Pixel, af.Red, af.Green, af.Blue}
	}

	for _, opt := range opts {
		opt(mb)
	}

	mb.computeGeometry()

	if mb.Background != nil {
		w.BackgroundPixel = mb.Background.Pixel
	}

	bindMenubutton(mb, app)
	return mb
}

func (mb *Menubutton) computeGeometry() {
	if mb.Font == nil {
		return
	}
	// Add space for indicator triangle.
	mb.textWidth = mb.Font.MeasureString(mb.Text) + 12
	m := mb.Font.Metrics()
	mb.textHeight = m.Linespace()

	inset := mb.BorderWidth + mb.HighlightWidth
	w := mb.Win
	w.ReqWidth = mb.textWidth + 2*mb.PadX + 2*inset
	w.ReqHeight = mb.textHeight + 2*mb.PadY + 2*inset
}

// PostMenu posts the associated menu.
func (mb *Menubutton) PostMenu() {
	if mb.Menu == nil || mb.Menu.IsPosted() {
		return
	}

	w := mb.Win
	// Translate widget coordinates to screen coordinates.
	// Since we're using child windows, we need the absolute position.
	screenX, screenY := mb.screenPos()

	switch mb.Direction {
	case Below:
		mb.Menu.Post(screenX, screenY+w.Height)
	case Above:
		mb.Menu.Post(screenX, screenY-mb.Menu.Win.ReqHeight)
	case Right:
		mb.Menu.Post(screenX+w.Width, screenY)
	case Left:
		mb.Menu.Post(screenX-mb.Menu.Win.ReqWidth, screenY)
	}
}

// screenPos calculates the screen position of the widget.
func (mb *Menubutton) screenPos() (int, int) {
	x, y := 0, 0
	for w := mb.Win; w != nil; w = w.Parent {
		x += w.X
		y += w.Y
	}
	return x, y
}

// Display draws the menubutton.
func (mb *Menubutton) Display() {
	if mb.Destroyed {
		return
	}
	w := mb.Win
	if w.XWindow == xlib.Window(0) {
		return
	}

	d := w.Display.XDisplay
	gc := w.GC

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *colorRef
	if mb.Background != nil {
		bgPixel = mb.Background.Pixel
	}
	if mb.Foreground != nil {
		fgCol = &colorRef{mb.Foreground.Pixel, mb.Foreground.Red, mb.Foreground.Green, mb.Foreground.Blue}
	}
	if mb.State == widget.StateActive && mb.ActiveBg != nil {
		bgPixel = mb.ActiveBg.Pixel
	}
	if mb.State == widget.StateActive && mb.ActiveFg != nil {
		fgCol = mb.ActiveFg
	}

	// Fill background.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Border.
	border := mb.Border
	if mb.State == widget.StateActive && mb.ActiveBg != nil {
		border = draw.NewBorder(mb.ActiveBg.Red, mb.ActiveBg.Green, mb.ActiveBg.Blue)
	}
	if border != nil && mb.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, border,
			0, 0, w.Width, w.Height, mb.BorderWidth, mb.Relief)
	}

	// Text.
	if mb.Font != nil && mb.Text != "" && fgCol != nil {
		inset := mb.BorderWidth + mb.HighlightWidth
		m := mb.Font.Metrics()
		textX := inset + mb.PadX
		textY := inset + mb.PadY + m.Ascent

		if xftFont, ok := mb.Font.(*font.XftFont); ok {
			xftFont.DrawString(w.Drawable(), textX, textY, mb.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)

			// Draw dropdown indicator triangle.
			triX := w.Width - inset - mb.PadX - 10
			triY := textY - m.Ascent/2
			xftFont.DrawString(w.Drawable(), triX, triY+m.Ascent, "\u25bc",
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
		}
	}

	d.Flush()
}

// Configure applies options.
func (mb *Menubutton) Configure(opts ...option.Option) {
	option.Apply(mb, opts)
	mb.UpdateBorder()
	mb.computeGeometry()
	if mb.Background != nil {
		mb.Win.BackgroundPixel = mb.Background.Pixel
	}
	mb.Display()
}

// Destroy cleans up.
func (mb *Menubutton) Destroy() {
	if mb.Destroyed {
		return
	}
	mb.Destroyed = true
	window.DestroyWindow(mb.Win)
}

func bindMenubutton(mb *Menubutton, app widget.AppContext) {
	w := mb.Win

	// Expose.
	app.Dispatcher().Bind(w.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		mb.Display()
	})

	// Configure.
	app.Dispatcher().Bind(w.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			mb.Display()
		}
	})

	// Enter → active.
	app.Dispatcher().Bind(w.XWindow, event.EnterMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		mb.State = widget.StateActive
		mb.Display()
	})

	// Leave → normal.
	app.Dispatcher().Bind(w.XWindow, event.LeaveMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		mb.State = widget.StateNormal
		mb.Display()
	})

	// Button press → post menu.
	app.Dispatcher().Bind(w.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			mb.PostMenu()
		}
	})
}
