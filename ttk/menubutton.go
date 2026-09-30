package ttk

import (
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/menu"
	"github.com/msorc/takigo/window"
)

// Direction specifies where the menu pops up relative to the menubutton.
type Direction int

const (
	DirBelow Direction = iota
	DirAbove
	DirLeft
	DirRight
)

// Menubutton is a themed button that posts a menu when clicked.
type Menubutton struct {
	TtkWidget
	Text      string
	Menu      *menu.Menu
	Direction Direction
	Font      font.Font
	Img       widget.WidgetImage
	Compound  widget.Compound
}

// GetText implements TextProvider.
func (mb *Menubutton) GetText() string { return mb.Text }

// GetFont implements TextProvider.
func (mb *Menubutton) GetFont() font.Font { return mb.Font }

// GetImage implements TextProvider.
func (mb *Menubutton) GetImage() widget.WidgetImage { return mb.Img }

// GetCompound implements TextProvider.
func (mb *Menubutton) GetCompound() widget.Compound { return mb.Compound }

// MenubuttonOption configures a Menubutton.
type MenubuttonOption func(*Menubutton)

// MenubuttonText sets the button text.
func MenubuttonText(s string) MenubuttonOption {
	return func(mb *Menubutton) { mb.Text = s }
}

// MenubuttonMenu sets the menu to post.
func MenubuttonMenu(m *menu.Menu) MenubuttonOption {
	return func(mb *Menubutton) { mb.Menu = m }
}

// MenubuttonDirection sets the menu pop direction.
func MenubuttonDirection(d Direction) MenubuttonOption {
	return func(mb *Menubutton) { mb.Direction = d }
}

// MenubuttonImage sets the image.
func MenubuttonImage(img widget.WidgetImage) MenubuttonOption {
	return func(mb *Menubutton) { mb.Img = img }
}

// MenubuttonCompound sets how text and image are combined.
func MenubuttonCompound(c widget.Compound) MenubuttonOption {
	return func(mb *Menubutton) { mb.Compound = c }
}

// MenubuttonStyleOpt overrides the TTK style name (e.g. "TMenubutton.Toolbutton").
func MenubuttonStyleOpt(name string) MenubuttonOption {
	return func(mb *Menubutton) { mb.StyleName = name }
}

// NewMenubutton creates a themed menubutton widget.
func NewMenubutton(parent widget.Caregiver, name string, opts ...MenubuttonOption) *Menubutton {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)
	win.Flags |= window.FlagFocusable // ttk::takefocus accepts it

	mb := &Menubutton{
		Direction: DirBelow,
	}
	mb.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&mb.TtkWidget, win, app, "TMenubutton")
	mb.DisplayFunc = mb.Display

	// Bind label element to this widget.
	if mb.Theme != nil {
		labelFactory := NewLabelElementFactory(mb)
		mb.LabelFactory = labelFactory
		tmpl := mb.Theme.GetLayout("TMenubutton")
		if tmpl != nil {
			ctx := &DrawContext{
				Display: mb.Context.Display,
				Depth:   mb.Context.Depth,
				Style:   mb.Context.Style,
			}
			mb.Layout = newLayoutWithLabel(tmpl, mb.Theme, ctx, mb.Context.Style, labelFactory)
			mb.Context = ctx
		}
	}

	for _, opt := range opts {
		opt(mb)
	}

	// If style was overridden, rebuild layout with the new style.
	if mb.StyleName != "TMenubutton" && mb.Theme != nil {
		style := mb.Theme.ResolveStyle(mb.StyleName)
		mb.setStyle(style)
		style = mb.Context.Style
		labelFactory := NewLabelElementFactory(mb)
		tmpl := mb.Theme.GetLayout(mb.StyleName)
		if tmpl == nil {
			tmpl = mb.Theme.GetLayout("TMenubutton")
		}
		if tmpl != nil {
			ctx := &DrawContext{
				Display: mb.Context.Display,
				Depth:   mb.Context.Depth,
				Style:   style,
			}
			mb.Layout = newLayoutWithLabel(tmpl, mb.Theme, ctx, style, labelFactory)
			mb.Context = ctx
		}
	}

	if mb.Layout != nil {
		rw, rh := mb.Layout.Size(mb.State)
		if rw > 0 {
			win.ReqWidth = rw
		}
		if rh > 0 {
			win.ReqHeight = rh
		}
	}

	bindTtkHover(&mb.TtkWidget, app)
	bindMenubutton(mb, app)

	return mb
}

// Configure sets options after creation.
func (mb *Menubutton) Configure(opts ...MenubuttonOption) {
	configure(&mb.TtkWidget, mb, opts, nil, nil)
}

// Display renders the menubutton with an indicator arrow.
func (mb *Menubutton) Display() {
	if mb.Destroyed {
		return
	}
	win := mb.Win
	if win.PlatformID == 0 {
		return
	}

	mb.TtkWidget.Display()
}

func bindMenubutton(mb *Menubutton, app widget.AppContext) {
	win := mb.Win

	// Button1 press → post menu.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if mb.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			mb.ChangeState(StatePressed, 0)
			mb.postMenu()
		}
	})

	// Button1 release → unpress.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		mb.ChangeState(0, StatePressed)
	})
}

func (mb *Menubutton) postMenu() {
	if mb.Menu == nil {
		return
	}

	win := mb.Win
	d := win.Display.Server

	// Compute menu position.
	var x, y int
	switch mb.Direction {
	case DirBelow:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, win.Height)
	case DirAbove:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, 0)
		y -= mb.Menu.Window().Height
	case DirLeft:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, 0, 0)
		x -= mb.Menu.Window().Width
	case DirRight:
		x, y = d.TranslateCoordinates(win.PlatformID, win.Display.RootWindow, win.Width, 0)
	}

	mb.Menu.Post(x, y)
}
