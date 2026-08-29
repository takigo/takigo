// Package menubutton implements a menubutton widget that posts
// an associated menu when clicked.
// It ports tk/generic/tkMenubutton.c.
package menubutton

import (
	"unicode"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
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

	Text        string
	Menu        *menu.Menu
	Direction   Direction
	Anchor      option.Anchor
	Underline   int  // index of underlined character for Alt+letter, -1=none
	IndicatorOn bool // whether to draw the dropdown arrow indicator
	OptionMenu  bool // draw a horizontal rectangle indicator instead of triangle

	// Active colors.
	ActiveBg *color.ColorRef
	ActiveFg *color.ColorRef

	// State.
	State widget.State

	textWidth  int
	textHeight int
}

// MenubuttonOption configures a Menubutton.
type MenubuttonOption func(*Menubutton)

func Text(s string) MenubuttonOption            { return func(mb *Menubutton) { mb.Text = s } }
func MenuOpt(m *menu.Menu) MenubuttonOption     { return func(mb *Menubutton) { mb.Menu = m } }
func DirectionOpt(d Direction) MenubuttonOption { return func(mb *Menubutton) { mb.Direction = d } }
func PadX(p int) MenubuttonOption               { return func(mb *Menubutton) { mb.PadX = p } }
func PadY(p int) MenubuttonOption               { return func(mb *Menubutton) { mb.PadY = p } }
func UnderlineOpt(i int) MenubuttonOption       { return func(mb *Menubutton) { mb.Underline = i } }
func IndicatorOnOpt(on bool) MenubuttonOption   { return func(mb *Menubutton) { mb.IndicatorOn = on } }
func OptionMenuOpt(on bool) MenubuttonOption    { return func(mb *Menubutton) { mb.OptionMenu = on } }

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
		Direction:   Below,
		Anchor:      option.AnchorCenter,
		Underline:   -1,
		IndicatorOn: true,
	}
	widget.InitBase(&mb.Base, w, app)
	mb.BorderWidth = widget.DefBorderWidth
	mb.Relief = option.ReliefRaised
	mb.PadX = 4
	mb.PadY = 2

	// Active colors.
	if ac, err := app.ColorCache().Get(widget.DefActiveBackground); err == nil {
		mb.ActiveBg = ac.Ref()
	}
	if af, err := app.ColorCache().Get(widget.DefActiveForeground); err == nil {
		mb.ActiveFg = af.Ref()
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
	mb.textWidth = mb.Font.MeasureString(mb.Text)
	if mb.IndicatorOn {
		if mb.OptionMenu {
			mb.textWidth += 18 // space for option-menu rectangle indicator
		} else {
			mb.textWidth += 12 // space for dropdown arrow
		}
	}
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

	// Ensure menu size is current before we use it to compute position.
	mb.Menu.PrepareGeometry()

	win := mb.Win
	d := win.Display.Server
	var x, y int

	switch mb.Direction {
	case Below:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, win.Height)
	case Above:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, 0)
		y -= mb.Menu.Win.ReqHeight
	case Right:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, win.Width, 0)
	case Left:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, 0)
		x -= mb.Menu.Win.ReqWidth
	}
	mb.Menu.PostFromButton(x, y)
}

// Display draws the menubutton.
func (mb *Menubutton) Display() {
	if mb.Destroyed {
		return
	}
	w := mb.Win
	if w.PlatformID == platform.WindowID(0) {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Choose colors based on state.
	bgPixel := uint64(0)
	var fgCol *color.ColorRef
	if mb.Background != nil {
		bgPixel = mb.Background.Pixel
	}
	if mb.Foreground != nil {
		fgCol = mb.Foreground.Ref()
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

		if df, ok := mb.Font.(platform.DrawableFont); ok {
			df.DrawString(w.Drawable(), textX, textY, mb.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)

			// Draw underline for Alt+letter mnemonic.
			if mb.Underline >= 0 && mb.Underline < len(mb.Text) {
				prefix := mb.Text[:mb.Underline]
				ch := string([]rune(mb.Text)[mb.Underline])
				ulX := textX + mb.Font.MeasureString(prefix)
				ulW := mb.Font.MeasureString(ch)
				ulY := textY + 2
				d.SetForeground(gc, fgCol.Pixel)
				d.DrawLine(w.Drawable(), gc, ulX, ulY, ulX+ulW, ulY)
			}

			// Draw indicator.
			if mb.IndicatorOn {
				if mb.OptionMenu {
					// Horizontal rectangle indicator (option-menu style).
					rectW := 10
					rectH := 3
					rectX := w.Width - inset - mb.PadX - rectW - 4
					rectY := textY - m.Ascent/2 - rectH/2
					d.SetForeground(gc, fgCol.Pixel)
					d.FillRectangle(w.Drawable(), gc, rectX, rectY, uint(rectW), uint(rectH))
				} else {
					// Dropdown triangle.
					triX := w.Width - inset - mb.PadX - 10
					triY := textY - m.Ascent/2
					df.DrawString(w.Drawable(), triX, triY+m.Ascent, "\u25bc",
						fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
				}
			}
		}
	}

	d.Flush()
}

// SetText changes the button label and requests a re-layout if the size changed.
func (mb *Menubutton) SetText(text string) {
	mb.Text = text
	mb.computeGeometry()
	if mb.Win.GeomManager != nil {
		mb.Win.GeomManager.RequestProc(mb.Win)
	}
	mb.Display()
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
	app.Dispatcher().Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		mb.Display()
	})

	// Configure.
	app.Dispatcher().Bind(w.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
			mb.Display()
		}
	})

	// Enter -> active.
	app.Dispatcher().Bind(w.PlatformID, event.EnterMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		mb.State = widget.StateActive
		mb.Display()
	})

	// Leave -> normal.
	app.Dispatcher().Bind(w.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		mb.State = widget.StateNormal
		mb.Display()
	})

	// Button press -> post menu.
	app.Dispatcher().Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if mb.State == widget.StateDisabled {
			return
		}
		if ev.Button == 1 {
			mb.PostMenu()
		}
	})

	// Alt+letter global binding for mnemonic navigation.
	if mb.Underline >= 0 && mb.Underline < len([]rune(mb.Text)) {
		mnemonicRune := unicode.ToLower([]rune(mb.Text)[mb.Underline])
		app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
			if ev.State&platform.Mod1Mask == 0 {
				return
			}
			if mb.Destroyed || mb.State == widget.StateDisabled {
				return
			}
			r := platform.KeySymToRune(ev.KeySym)
			if unicode.ToLower(r) == mnemonicRune {
				mb.PostMenu()
			}
		})
	}
}
