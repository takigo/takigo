// Package menubutton implements a menubutton widget that posts
// an associated menu when clicked.
// It ports tk/generic/tkMenubutton.c.
package menubutton

import (
	"log"
	"unicode"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
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
	IndicatorOn bool // -indicatoron: draw the raised option-menu indicator

	indicatorWidth, indicatorHeight int
	OptionMenu                      bool // draw a horizontal rectangle indicator instead of triangle

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

// Relief sets -relief.
func Relief(r option.Relief) MenubuttonOption { return func(mb *Menubutton) { mb.Relief = r } }

func Background(name string) MenubuttonOption {
	return func(mb *Menubutton) {
		col, err := mb.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("menubutton: failed to get color %q: %v", name, err)
			return
		}
		mb.Background = col
		mb.UpdateBorder()
	}
}

func Foreground(name string) MenubuttonOption {
	return func(mb *Menubutton) {
		col, err := mb.App.ColorCache().Get(name)
		if err != nil {
			log.Printf("menubutton: failed to get color %q: %v", name, err)
			return
		}
		mb.Foreground = col
	}
}

// --- Ttk-compatible aliases (prefix with Menubutton) for consistent naming ---
// These aliases match the naming convention used by ttk widgets (ttk.MenubuttonText, etc.)
// allowing consistent option naming when both classic and ttk widgets are used.

// MenubuttonText is an alias for Text.
var MenubuttonText = Text

// MenubuttonMenuOpt is an alias for MenuOpt.
var MenubuttonMenuOpt = MenuOpt

// MenubuttonDirectionOpt is an alias for DirectionOpt.
var MenubuttonDirectionOpt = DirectionOpt

// MenubuttonPadX is an alias for PadX.
var MenubuttonPadX = PadX

// MenubuttonPadY is an alias for PadY.
var MenubuttonPadY = PadY

// MenubuttonUnderlineOpt is an alias for UnderlineOpt.
var MenubuttonUnderlineOpt = UnderlineOpt

// MenubuttonRelief is an alias for Relief.
var MenubuttonRelief = Relief

// MenubuttonIndicatorOnOpt is an alias for IndicatorOnOpt.
var MenubuttonIndicatorOnOpt = IndicatorOnOpt

// MenubuttonOptionMenuOpt is an alias for OptionMenuOpt.
var MenubuttonOptionMenuOpt = OptionMenuOpt

// MenubuttonBackground is an alias for Background.
var MenubuttonBackground = Background

// MenubuttonForeground is an alias for Foreground.
var MenubuttonForeground = Foreground

// New creates a new Menubutton widget.
func New(parent widget.Caregiver, name string, opts ...MenubuttonOption) *Menubutton {
	app := parent.AppContext()
	w := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)

	mb := &Menubutton{
		Direction: Below,
		Anchor:    option.AnchorCenter,
		Underline: -1,
	}
	widget.InitBase(&mb.Base, w, app)
	mb.SetDisplayProc(mb.display)
	w.Class = "Menubutton"
	// tkUnixDefault.h DEF_MENUBUTTON_*.
	mb.BorderWidth = 1
	mb.Relief = option.ReliefFlat
	mb.PadX = screenunit.Px("4p")
	mb.PadY = screenunit.Px("3p")

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
	if mb.OptionMenu {
		// tk_optionMenu (library/optMenu.tcl).
		mb.IndicatorOn = true
		mb.Relief = option.ReliefRaised
		mb.HighlightWidth = 1
		mb.Anchor = option.AnchorCenter
	}

	mb.computeGeometry()

	if mb.Background != nil {
		w.BackgroundPixel = mb.Background.Pixel
	}

	bindMenubutton(mb, app)
	return mb
}

// computeGeometry ports TkpComputeMenuButtonGeometry (tkUnixMenubu.c) for
// text menubuttons; the indicator's size follows the screen density.
func (mb *Menubutton) computeGeometry() {
	if mb.Font == nil {
		return
	}
	mb.textWidth = mb.Font.MeasureString(mb.Text)
	mb.textHeight = mb.Font.Metrics().Linespace()
	width := mb.textWidth + 2*mb.PadX
	height := mb.textHeight + 2*mb.PadY
	mb.indicatorWidth, mb.indicatorHeight = 0, 0
	if mb.IndicatorOn {
		dpi := screenunit.DPI()
		mb.indicatorHeight = int(17 * dpi / 254)
		mb.indicatorWidth = int(40*dpi/254) + 2*mb.indicatorHeight
		width += mb.indicatorWidth
	}
	inset := mb.BorderWidth + mb.HighlightWidth
	w := mb.Win
	w.ReqWidth = width + 2*inset
	w.ReqHeight = height + 2*inset
	w.InternalBorderLeft, w.InternalBorderRight = inset, inset
	w.InternalBorderTop, w.InternalBorderBottom = inset, inset
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

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (mb *Menubutton) Display() {
	mb.EventuallyRedraw()
}

// display draws the menubutton.
func (mb *Menubutton) display() {
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

	// TkpDisplayMenuButton: background, text, indicator, then the border
	// and highlight on top.
	d.SetForeground(gc, bgPixel)
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))
	border := draw.NewBorderFromPixel(bgPixel)
	inset := mb.BorderWidth + mb.HighlightWidth

	if mb.Font != nil && fgCol != nil {
		m := mb.Font.Metrics()
		x, y := widget.ComputeAnchor(mb.Anchor, w.Width, w.Height, inset, mb.PadX, mb.PadY,
			mb.textWidth+mb.indicatorWidth, mb.textHeight)
		if df, ok := mb.Font.(platform.DrawableFont); ok && mb.Text != "" {
			df.DrawString(w.Drawable(), x, y+m.Ascent, mb.Text,
				fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			if mb.Underline >= 0 && mb.Underline < len([]rune(mb.Text)) {
				runes := []rune(mb.Text)
				ulX := x + mb.Font.MeasureString(string(runes[:mb.Underline]))
				ulW := mb.Font.MeasureString(string(runes[mb.Underline]))
				d.SetForeground(gc, fgCol.Pixel)
				ulPos, ulH := font.Underline(mb.Font)
				d.FillRectangle(w.Drawable(), gc, ulX, y+m.Ascent+ulPos, uint(ulW), uint(ulH))
			}
		}
	}

	if mb.IndicatorOn {
		ibw := max(1, (mb.indicatorHeight+1)/3)
		draw.Fill3DRectangle(d, w.Drawable(), gc, border,
			w.Width-inset-mb.indicatorWidth+mb.indicatorHeight,
			(w.Height-mb.indicatorHeight)/2,
			mb.indicatorWidth-2*mb.indicatorHeight, mb.indicatorHeight, ibw, option.ReliefRaised)
	}

	hl := mb.HighlightWidth
	if mb.Relief != option.ReliefFlat {
		draw.Draw3DRectangle(d, w.Drawable(), gc, border, hl, hl,
			w.Width-2*hl, w.Height-2*hl, mb.BorderWidth, mb.Relief)
	}
	if hl > 0 {
		pixel := bgPixel
		if mb.HighlightBackground != nil {
			pixel = mb.HighlightBackground.Pixel
		}
		d.SetForeground(gc, pixel)
		for i := range hl {
			d.DrawRectangle(w.Drawable(), gc, i, i, uint(w.Width-1-2*i), uint(w.Height-1-2*i))
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
		app.Dispatcher().BindGlobalFor(w.PlatformID, event.KeyPressMask, func(ev *event.Event) {
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
