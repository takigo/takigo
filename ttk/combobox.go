package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// ComboboxState controls the editability of a combobox.
type ComboboxState int

const (
	// ComboNormal allows editing.
	ComboNormal ComboboxState = iota
	// ComboReadonly only allows selection from the dropdown.
	ComboReadonly
	// ComboDisabled is fully disabled.
	ComboDisabled
)

// Combobox is a themed entry with a dropdown list.
type Combobox struct {
	TtkWidget
	Values    []string
	text      []rune
	insertPos int
	CbState   ComboboxState
	Font      font.Font
	Command   func(value string) // called when value changes

	// Dropdown state.
	dropWin  *window.Window
	dropSel  int
	dropOpen bool
	grabbed  bool

	// Layout.
	arrowWidth int
	insetX     int
	insetY     int
}

// ComboboxOption configures a Combobox.
type ComboboxOption func(*Combobox)

// ComboboxValues sets the dropdown values.
func ComboboxValues(vals []string) ComboboxOption {
	return func(c *Combobox) { c.Values = vals }
}

// ComboboxText sets the initial text.
func ComboboxText(s string) ComboboxOption {
	return func(c *Combobox) { c.text = []rune(s) }
}

// ComboboxCbState sets the combobox state.
func ComboboxCbState(s ComboboxState) ComboboxOption {
	return func(c *Combobox) { c.CbState = s }
}

// ComboboxCommand sets the value change callback.
func ComboboxCommand(fn func(string)) ComboboxOption {
	return func(c *Combobox) { c.Command = fn }
}

// NewCombobox creates a themed combobox widget.
func NewCombobox(parent widget.Caregiver, name string, opts ...ComboboxOption) *Combobox {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	c := &Combobox{
		arrowWidth: 20,
		insetX:     4,
		insetY:     2,
		dropSel:    -1,
	}

	c.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&c.TtkWidget, win, app, "TCombobox")

	// Set reasonable size.
	if c.Font != nil {
		m := c.Font.Metrics()
		win.ReqWidth = 200
		win.ReqHeight = m.Linespace() + 2*c.insetY + 4
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.CbState == ComboDisabled {
		c.ChangeState(StateDisabled, 0)
	} else if c.CbState == ComboReadonly {
		c.ChangeState(StateReadonly, 0)
	}

	bindCombobox(c, app)

	return c
}

// Get returns the current text value.
func (c *Combobox) Get() string {
	return string(c.text)
}

// Set sets the text value.
func (c *Combobox) Set(s string) {
	c.text = []rune(s)
	c.insertPos = len(c.text)
	c.Display()
	if c.Command != nil {
		c.Command(s)
	}
}

// Display draws the combobox.
func (c *Combobox) Display() {
	if c.Destroyed {
		return
	}
	win := c.Win
	if win.XWindow == xlib.Window(0) {
		return
	}

	d := win.Display.XDisplay
	gc := win.GC
	width := win.Width
	height := win.Height

	if width <= 0 || height <= 0 {
		return
	}

	// Double buffer.
	if c.pixmap == xlib.Pixmap(0) || c.pixmapW != width || c.pixmapH != height {
		if c.pixmap != xlib.Pixmap(0) {
			d.FreePixmap(c.pixmap)
		}
		c.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		c.pixmapW = width
		c.pixmapH = height
	}

	pixDrawable := xlib.PixmapDrawable(c.pixmap)

	bg := LookupColor(c.Context.Style, "-background", c.State, 0xd9d9d9)
	fg := LookupColor(c.Context.Style, "-foreground", c.State, 0x000000)

	// Entry field background (white for editable).
	fieldBg := uint64(0xffffff)
	if c.CbState == ComboDisabled {
		fieldBg = bg
	}

	// Fill entry area.
	d.SetForeground(gc, fieldBg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Draw border.
	border := draw.NewBorderFromPixel(bg)
	d.SetForeground(gc, uint64(0x9e9a91))
	d.DrawLine(pixDrawable, gc, 0, 0, width-1, 0)
	d.DrawLine(pixDrawable, gc, 0, 0, 0, height-1)
	d.SetForeground(gc, border.LightPixel)
	d.DrawLine(pixDrawable, gc, 0, height-1, width-1, height-1)
	d.DrawLine(pixDrawable, gc, width-1, 0, width-1, height-1)

	// Arrow button area.
	arrowX := width - c.arrowWidth
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, arrowX, 1, uint(c.arrowWidth-1), uint(height-2))

	// Draw separator line.
	d.SetForeground(gc, uint64(0x9e9a91))
	d.DrawLine(pixDrawable, gc, arrowX, 1, arrowX, height-2)

	// Draw arrow.
	arrowCX := arrowX + c.arrowWidth/2
	arrowCY := height / 2
	d.SetForeground(gc, fg)
	for row := 0; row < 4; row++ {
		d.DrawLine(pixDrawable, gc, arrowCX-3+row, arrowCY-2+row, arrowCX+3-row, arrowCY-2+row)
	}

	// Draw text.
	if c.Font != nil && len(c.text) > 0 {
		textStr := string(c.text)
		m := c.Font.Metrics()
		textX := c.insetX + 1
		textY := (height-m.Linespace())/2 + m.Ascent

		if xftFont, ok := c.Font.(*font.XftFont); ok {
			r := uint16((fg >> 16) & 0xFF) << 8
			g := uint16((fg >> 8) & 0xFF) << 8
			b := uint16((fg) & 0xFF) << 8
			xftFont.DrawString(pixDrawable, textX, textY, textStr, fg, r, g, b)
		}
	}

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

// openDropdown shows the dropdown list.
func (c *Combobox) openDropdown() {
	if c.dropOpen || len(c.Values) == 0 {
		return
	}

	win := c.Win
	disp := win.Display
	d := disp.XDisplay

	// Calculate dropdown position (below the combobox in screen coordinates).
	screenX, screenY := d.TranslateCoordinates(win.XWindow, disp.RootXWindow, 0, win.Height)

	lineH := 0
	if c.Font != nil {
		lineH = c.Font.Metrics().Linespace() + 4
	} else {
		lineH = 20
	}
	dropH := lineH * len(c.Values)
	if dropH > 200 {
		dropH = 200
	}
	dropW := win.Width

	// Create override-redirect popup window (same pattern as menu.go).
	attrs := &xlib.WindowAttributes{
		BackgroundPixel:  disp.WhitePixel,
		BorderPixel:      disp.BlackPixel,
		OverrideRedirect: true,
		EventMask: int64(
			xlib.ButtonPressMask |
				xlib.ButtonReleaseMask |
				xlib.PointerMotionMask |
				xlib.EnterWindowMask |
				xlib.LeaveWindowMask |
				xlib.ExposureMask |
				xlib.StructureNotifyMask),
	}

	xwin := d.CreateWindow(
		disp.RootXWindow,
		screenX, screenY, uint(dropW), uint(dropH), 1,
		disp.Depth, xlib.InputOutput, disp.Visual,
		xlib.CWBackPixel|xlib.CWBorderPixel|xlib.CWOverrideRedirect|xlib.CWEventMask,
		attrs,
	)

	dw := &window.Window{
		XWindow:         xwin,
		Display:         disp,
		Parent:          win,
		Name:            "dropdown",
		PathName:        window.BuildPathName(win, "dropdown"),
		Width:           dropW,
		Height:          dropH,
		ReqWidth:        dropW,
		ReqHeight:       dropH,
		Depth:           disp.Depth,
		Visual:          disp.Visual,
		Colormap:        disp.Colormap,
		BackgroundPixel: disp.WhitePixel,
	}

	dw.GC = d.CreateGC(dw.Drawable(), xlib.GCForeground|xlib.GCBackground, &xlib.GCValues{
		Foreground: disp.BlackPixel,
		Background: disp.WhitePixel,
	})

	disp.RegisterWindow(xwin, dw)
	win.AddChild(dw)

	c.dropWin = dw
	c.dropSel = -1

	// Find current value in dropdown.
	curText := string(c.text)
	for i, v := range c.Values {
		if v == curText {
			c.dropSel = i
			break
		}
	}

	c.dropOpen = true
	d.MapRaised(xwin)

	// Bind dropdown events.
	c.App.Dispatcher().Bind(xwin, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.displayDropdown()
	})

	c.App.Dispatcher().Bind(xwin, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 && lineH > 0 {
			idx := ev.Y / lineH
			if idx >= 0 && idx < len(c.Values) {
				c.Set(c.Values[idx])
			}
			c.closeDropdown()
		}
	})

	c.App.Dispatcher().Bind(xwin, event.MotionMask, func(ev *event.Event) {
		if lineH > 0 {
			idx := ev.Y / lineH
			if idx != c.dropSel && idx >= 0 && idx < len(c.Values) {
				c.dropSel = idx
				c.displayDropdown()
			}
		}
	})

	c.App.Dispatcher().Bind(xwin, event.LeaveMask, func(ev *event.Event) {
		c.dropSel = -1
		c.displayDropdown()
	})

	// Grab pointer.
	d.GrabPointer(xwin, true,
		uint(xlib.ButtonPressMask|xlib.ButtonReleaseMask|xlib.PointerMotionMask|xlib.EnterWindowMask|xlib.LeaveWindowMask),
		xlib.GrabModeAsync, xlib.GrabModeAsync,
		xlib.Window(0), xlib.Cursor(0), xlib.CurrentTime)
	c.grabbed = true

	c.displayDropdown()
}

// closeDropdown hides the dropdown.
func (c *Combobox) closeDropdown() {
	if !c.dropOpen {
		return
	}
	c.dropOpen = false
	d := c.Win.Display.XDisplay

	if c.grabbed {
		d.UngrabPointer(xlib.CurrentTime)
		c.grabbed = false
	}

	if c.dropWin != nil {
		d.UnmapWindow(c.dropWin.XWindow)
		d.DestroyWindow(c.dropWin.XWindow)
		c.dropWin = nil
	}
}

// displayDropdown draws the dropdown list contents.
func (c *Combobox) displayDropdown() {
	if c.dropWin == nil || !c.dropOpen {
		return
	}

	dw := c.dropWin
	d := dw.Display.XDisplay
	gc := dw.GC
	width := dw.Width
	height := dw.Height

	// White background.
	d.SetForeground(gc, uint64(0xffffff))
	d.FillRectangle(dw.Drawable(), gc, 0, 0, uint(width), uint(height))

	// Border.
	d.SetForeground(gc, uint64(0x9e9a91))
	d.DrawRectangle(dw.Drawable(), gc, 0, 0, uint(width-1), uint(height-1))

	lineH := 0
	if c.Font != nil {
		lineH = c.Font.Metrics().Linespace() + 4
	} else {
		lineH = 20
	}

	for i, val := range c.Values {
		y := i * lineH

		// Highlight selected.
		if i == c.dropSel {
			d.SetForeground(gc, uint64(0x4a6984))
			d.FillRectangle(dw.Drawable(), gc, 1, y, uint(width-2), uint(lineH))
		}

		// Draw text.
		if c.Font != nil {
			m := c.Font.Metrics()
			textY := y + 2 + m.Ascent

			fg := uint64(0x000000)
			if i == c.dropSel {
				fg = uint64(0xffffff)
			}

			if xftFont, ok := c.Font.(*font.XftFont); ok {
				r := uint16((fg >> 16) & 0xFF) << 8
				g := uint16((fg >> 8) & 0xFF) << 8
				b := uint16((fg) & 0xFF) << 8
				xftFont.DrawString(dw.Drawable(), 4, textY, val, fg, r, g, b)
			}
		}
	}

	d.Flush()
}

func bindCombobox(c *Combobox, app widget.AppContext) {
	win := c.Win

	app.Dispatcher().Bind(win.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		c.Display()
	})

	app.Dispatcher().Bind(win.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			c.Display()
		}
	})

	app.Dispatcher().Bind(win.XWindow, event.EnterMask, func(ev *event.Event) {
		c.ChangeState(StateHover|StateActive, 0)
	})

	app.Dispatcher().Bind(win.XWindow, event.LeaveMask, func(ev *event.Event) {
		c.ChangeState(0, StateHover|StateActive)
	})

	// Button1 → open/close dropdown.
	app.Dispatcher().Bind(win.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if c.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			arrowX := win.Width - c.arrowWidth
			if ev.X >= arrowX || c.CbState == ComboReadonly {
				if c.dropOpen {
					c.closeDropdown()
				} else {
					c.openDropdown()
				}
			}
		}
	})

	// Key events for editable combobox.
	app.Dispatcher().Bind(win.XWindow, event.KeyPressMask, func(ev *event.Event) {
		if c.State&StateDisabled != 0 || c.CbState != ComboNormal {
			return
		}
		switch ev.KeySym {
		case xlib.XK_BackSpace:
			if c.insertPos > 0 {
				c.text = append(c.text[:c.insertPos-1], c.text[c.insertPos:]...)
				c.insertPos--
				c.Display()
			}
		case xlib.XK_Delete:
			if c.insertPos < len(c.text) {
				c.text = append(c.text[:c.insertPos], c.text[c.insertPos+1:]...)
				c.Display()
			}
		case xlib.XK_Left:
			if c.insertPos > 0 {
				c.insertPos--
				c.Display()
			}
		case xlib.XK_Right:
			if c.insertPos < len(c.text) {
				c.insertPos++
				c.Display()
			}
		case xlib.XK_Home:
			c.insertPos = 0
			c.Display()
		case xlib.XK_End:
			c.insertPos = len(c.text)
			c.Display()
		default:
			if ev.Str != "" {
				r := []rune(ev.Str)
				if len(r) == 1 && r[0] >= 32 {
					newText := make([]rune, 0, len(c.text)+1)
					newText = append(newText, c.text[:c.insertPos]...)
					newText = append(newText, r[0])
					newText = append(newText, c.text[c.insertPos:]...)
					c.text = newText
					c.insertPos++
					c.Display()
				}
			}
		}
	})

	// Focus events.
	app.Dispatcher().Bind(win.XWindow, event.FocusChangeMask, func(ev *event.Event) {
		if ev.Type == event.FocusInType {
			c.ChangeState(StateFocus, 0)
		} else if ev.Type == event.FocusOutType {
			c.ChangeState(0, StateFocus)
			c.closeDropdown()
		}
	})
}
