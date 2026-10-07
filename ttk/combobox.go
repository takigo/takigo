package ttk

import (
	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk/internal/entrytext"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Combobox is a themed entry with a dropdown list.
type Combobox struct {
	TtkWidget
	Values  []string
	edit    entrytext.Helper
	CbState FieldState
	Font    font.Font
	Command func(value string) // called when value changes
	// Placeholder is -placeholder: shown while the text is empty.
	Placeholder string
	selBg       uint64 // selection highlight background
	selFg       uint64 // selection text foreground

	// Dropdown state.
	dropWin      *window.Window
	dropSel      int
	dropOpen     bool
	grabbed      bool
	arrowPressed bool

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
	return func(c *Combobox) { c.edit.Text = []rune(s) }
}

// ComboboxState sets -state: FieldReadonly only allows choosing from the
// dropdown.
func ComboboxState(s FieldState) ComboboxOption {
	return func(c *Combobox) { c.CbState = s }
}

// ComboboxPlaceholder sets -placeholder.
func ComboboxPlaceholder(s string) ComboboxOption {
	return func(c *Combobox) { c.Placeholder = s }
}

// ComboboxCommand sets the value change callback.
func ComboboxCommand(fn func(string)) ComboboxOption {
	return func(c *Combobox) { c.Command = fn }
}

// ComboboxStyle sets -style.
func ComboboxStyle(name string) ComboboxOption {
	return func(c *Combobox) { c.StyleName = name }
}

// NewCombobox creates a themed combobox widget.
func NewCombobox(parent widget.Caregiver, name string, opts ...ComboboxOption) *Combobox {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)
	win.Flags |= window.FlagFocusable // ttk::takefocus accepts it

	c := &Combobox{
		arrowWidth: 20,
		insetX:     4,
		insetY:     2,
		dropSel:    -1,
		selBg:      0x3399ff,
		selFg:      0xffffff,
	}

	c.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	c.edit = entrytext.Helper{
		Font:     c.Font,
		TextX:    c.insetX + 1,
		App:      app,
		Win:      win,
		Redraw:   c.Display,
		Editable: func() bool { return c.CbState == FieldNormal },
	}

	InitTtkWidget(&c.TtkWidget, win, app, "TCombobox")
	c.reconfigure = func() { _ = c.Configure() }
	win.OnDestroy(c.Destroy)
	c.DisplayFunc = c.Display

	for _, opt := range opts {
		opt(c)
	}
	if c.StyleName != "TCombobox" {
		c.RefreshTheme()
	}
	c.requestSize()
	c.syncState()
	bindCombobox(c, app)

	return c
}

// syncState applies -state to the widget state and cursor.
func (c *Combobox) syncState() {
	switch c.CbState {
	case FieldDisabled:
		c.ChangeState(StateDisabled, StateReadonly)
	case FieldReadonly:
		c.ChangeState(StateReadonly, StateDisabled)
	default:
		c.ChangeState(0, StateDisabled|StateReadonly)
	}
	if c.CbState == FieldNormal {
		c.Win.SetCursor(cursor.XTerm)
	} else {
		c.Win.SetCursor(cursor.LeftPtr)
	}
}

// Configure sets options after creation.
func (c *Combobox) Configure(opts ...ComboboxOption) error {
	return configure(&c.TtkWidget, c, opts, c.syncState, c.requestSize)
}

// fieldPad ports FieldElementSize: -borderwidth widened to -focuswidth.
func (c *Combobox) fieldPad() int {
	bw := LookupInt(c.Context.Style, "-borderwidth", c.State, 2)
	fw := LookupInt(c.Context.Style, "-focuswidth", c.State, 2)
	if fw > 0 && bw < 2 {
		bw = fw
	}
	return bw
}

// arrowSize ports BoxArrowElementSize for the down arrow.
func (c *Combobox) arrowSize() (int, int) {
	pad := 3 * screenunit.ScalingPct() / 100
	size := LookupInt(c.Context.Style, "-arrowsize", c.State, 14) - 2*pad +
		2*((screenunit.ScalingPct()+50)/100)
	return 2*(size/2) + 1 + 2*pad, size/2 + 1 + 2*pad
}

// requestSize ports the ComboboxLayout size: field, arrow packed right,
// -padding, and a textarea of -width 20 average ("0") characters.
func (c *Combobox) requestSize() {
	if c.Context == nil || c.Font == nil {
		return
	}
	fp := c.fieldPad()
	aw, ah := c.arrowSize()
	p := LookupPadding(c.Context.Style, "-padding", c.State, Padding{})
	m := c.Font.Metrics()
	c.Win.ReqWidth = 2*fp + aw + p.Left + p.Right + 20*c.Font.MeasureString("0")
	c.Win.ReqHeight = 2*fp + max(ah, m.Linespace()+p.Top+p.Bottom)
	c.arrowWidth = aw
	c.insetX = fp + p.Left
	c.edit.TextX = c.insetX
}

// Get returns the current text value.
func (c *Combobox) Get() string {
	return c.edit.Get()
}

// Set sets the text value.
func (c *Combobox) Set(s string) {
	c.edit.Set(s)
	if c.Command != nil {
		c.Command(c.edit.Get())
	}
}

// Display draws the combobox.
func (c *Combobox) Display() {
	if c.Destroyed() {
		return
	}
	win := c.Win
	if win.PlatformID == 0 {
		return
	}

	d := win.Display.Server
	gc := win.GC
	width := win.Width
	height := win.Height

	if width <= 0 || height <= 0 {
		return
	}

	pixDrawable := c.backBuffer(width, height)
	if pixDrawable == 0 {
		return
	}

	st := c.Context.Style
	bg := LookupColor(st, "-background", c.State, 0xd9d9d9)
	fg := LookupColor(st, "-foreground", c.State, 0x000000)
	c.selBg = LookupColor(st, "-selectbackground", c.State, 0x4a6984)
	c.selFg = LookupColor(st, "-selectforeground", c.State, 0xffffff)

	// Combobox.field (FieldElementDraw, ttkElements.c).
	fieldBg := LookupColor(st, "-fieldbackground", c.State, 0xffffff)
	bw := LookupInt(st, "-borderwidth", c.State, 2)
	fw := LookupInt(st, "-focuswidth", c.State, 2)
	draw.Fill3DRectangle(d, pixDrawable, gc, draw.NewBorderFromPixel(fieldBg),
		0, 0, width, height, bw, option.ReliefSunken)
	if fw > 0 && c.State&StateFocus != 0 {
		d.SetForeground(gc, LookupColor(st, "-focuscolor", c.State, 0x4a6984))
		d.DrawRectangle(pixDrawable, gc, 0, 0, uint(width-1), uint(height-1))
	}

	// Combobox.downarrow (BoxArrowElementDraw), packed right, filling Y.
	fp := c.fieldPad()
	ab := Box{width - fp - c.arrowWidth, fp, c.arrowWidth, height - 2*fp}
	arrowX := ab.X
	border := draw.NewBorderFromPixel(bg)
	relief := LookupRelief(st, "-relief", c.State, option.ReliefRaised)
	if c.arrowPressed {
		relief = option.ReliefSunken
	}
	draw.Fill3DRectangle(d, pixDrawable, gc, border, ab.X, ab.Y, ab.Width, ab.Height,
		LookupInt(st, "-borderwidth", c.State, 1), relief)
	d.SetForeground(gc, border.DarkPixel)
	d.DrawLine(pixDrawable, gc, ab.X, ab.Y+1, ab.X, ab.Y+ab.Height-1)
	pad := 3 * screenunit.ScalingPct() / 100
	ib := Box{ab.X + pad, ab.Y + pad, ab.Width - 2*pad, ab.Height - 2*pad}
	cx, cy := 2*(ib.Width/2)+1, ib.Width/2+1
	if (ib.Height-cy)%2 == 1 {
		cy++
	}
	arrowBox := Box{ib.X + (ib.Width-cx)/2, ib.Y + (ib.Height-cy)/2, cx, cy}
	d.SetForeground(gc, LookupColor(st, "-arrowcolor", c.State, 0x000000))
	pts := arrowDownPoints(arrowBox)
	d.FillPolygon(pixDrawable, gc, pts, 2, 0)
	d.DrawLines(pixDrawable, gc, append(pts, pts[0]), 0)
	d.DrawLine(pixDrawable, gc, int(pts[2].X), int(pts[2].Y), int(pts[2].X), int(pts[2].Y))

	// The text with its selection and cursor, or the placeholder.
	ft := fieldText{
		font: c.Font, text: c.edit.Text, x: c.insetX,
		left: c.insetX, right: arrowX - 1, top: c.insetY, height: height - 2*c.insetY,
		selFirst: -1, selLast: -1, cursor: -1,
		fg: fg, selBg: c.selBg, selFg: c.selFg,
		insertColor: LookupColor(st, "-insertcolor", c.State, fg),
		insertWidth: LookupInt(st, "-insertwidth", c.State, 1),
	}
	if len(c.edit.Text) == 0 && c.Placeholder != "" {
		ft.text = []rune(c.Placeholder)
		ft.fg = LookupColor(st, "-placeholderforeground", c.State, 0xb3b3b3)
	} else if c.State&StateFocus != 0 && c.edit.HasSelection() {
		ft.selFirst, ft.selLast = c.edit.SelFirst, c.edit.SelLast
	}
	if c.State&StateFocus != 0 && c.CbState == FieldNormal {
		ft.cursor = c.edit.InsertPos
	}
	drawFieldText(d, pixDrawable, gc, ft)

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
	d := disp.Server

	// Calculate dropdown position (below the combobox in screen coordinates).
	screenX, screenY := d.TranslateCoordinates(win.PlatformID, disp.RootWindow, 0, win.Height)

	lineH := 0
	if c.Font != nil {
		lineH = c.Font.Metrics().Linespace() + 4
	} else {
		lineH = 20
	}
	dropH := min(lineH*len(c.Values), 200)
	dropW := win.Width

	// Create override-redirect popup window.
	attrs := &platform.WindowAttrs{
		BackgroundPixel:  disp.WhitePixel,
		BorderPixel:      disp.BlackPixel,
		OverrideRedirect: true,
		EventMask: int64(
			platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.PointerMotionMask |
				platform.EnterWindowMask |
				platform.LeaveWindowMask |
				platform.ExposureMask |
				platform.StructureNotifyMask),
	}

	xwin := d.CreateWindow(
		disp.RootWindow,
		screenX, screenY, uint(dropW), uint(dropH), 1,
		disp.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWOverrideRedirect|platform.CWEventMask,
		attrs,
	)

	dw := &window.Window{
		PlatformID:      xwin,
		Display:         disp,
		Parent:          win,
		Name:            "dropdown",
		PathName:        window.BuildPathName(win, "dropdown"),
		Width:           dropW,
		Height:          dropH,
		ReqWidth:        dropW,
		ReqHeight:       dropH,
		Depth:           disp.Depth,
		BackgroundPixel: disp.WhitePixel,
	}

	dw.GC = d.CreateGC(dw.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: disp.BlackPixel,
		Background: disp.WhitePixel,
	})

	disp.RegisterWindow(xwin, dw)
	win.AddChild(dw)

	c.dropWin = dw
	c.dropSel = -1

	// Find current value in dropdown.
	curText := c.edit.Get()
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
		uint(platform.ButtonPressMask|platform.ButtonReleaseMask|platform.PointerMotionMask|platform.EnterWindowMask|platform.LeaveWindowMask),
		platform.GrabModeAsync, platform.GrabModeAsync,
		platform.WindowID(0), platform.CursorID(0), platform.CurrentTime)
	c.grabbed = true

	c.displayDropdown()
}

// closeDropdown hides the dropdown.
func (c *Combobox) closeDropdown() {
	if !c.dropOpen {
		return
	}
	c.dropOpen = false
	d := c.Win.Display.Server

	if c.grabbed {
		d.UngrabPointer(platform.CurrentTime)
		c.grabbed = false
	}

	if c.dropWin != nil {
		// The dropdown is a child window: destroying the combobox already
		// destroyed it before this runs from the destroy hook.
		if !c.dropWin.IsDestroyed() {
			c.App.Dispatcher().Unbind(c.dropWin.PlatformID)
			d.UnmapWindow(c.dropWin.PlatformID)
			window.DestroyWindow(c.dropWin)
		}
		c.dropWin = nil
	}
}

// Destroy cleans up the combobox, closing any open dropdown.
func (c *Combobox) Destroy() {
	c.closeDropdown()
	c.TtkWidget.Destroy()
}

// displayDropdown draws the dropdown list contents.
func (c *Combobox) displayDropdown() {
	if c.dropWin == nil || !c.dropOpen {
		return
	}

	dw := c.dropWin
	d := dw.Display.Server
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

			if df, ok := c.Font.(platform.DrawableFont); ok {
				drawString(df, dw.Drawable(), 4, textY, val, fg)
			}
		}
	}

	d.Flush()
}

// updateCursor sets the cursor shape based on mouse x position and combobox state.
func (c *Combobox) updateCursor(x int) {
	arrowX := c.Win.Width - c.arrowWidth
	if x >= arrowX || c.CbState != FieldNormal {
		c.Win.SetCursor(cursor.LeftPtr)
	} else {
		c.Win.SetCursor(cursor.XTerm)
	}
}

func bindCombobox(c *Combobox, app widget.AppContext) {
	win := c.Win

	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		c.ChangeState(StateHover|StateActive, 0)
		c.updateCursor(ev.X)
	})

	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		c.updateCursor(ev.X)
		if ev.State&platform.Button1Mask != 0 && c.CbState == FieldNormal && c.State&StateFocus != 0 {
			c.edit.MoveCursor(c.edit.ClosestGap(ev.X), c.edit.SelAnchor, true)
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		c.ChangeState(0, StateHover|StateActive)
	})

	// Button1 → open/close dropdown.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if c.State&StateDisabled != 0 {
			return
		}
		if ev.Button == 1 {
			arrowX := win.Width - c.arrowWidth
			if ev.X >= arrowX || c.CbState == FieldReadonly {
				// ttk::combobox::Press: the widget takes the focus so the
				// posted list gets the keys.
				widget.Focus(app, win)
				c.arrowPressed = true
				c.Display()
				if c.dropOpen {
					c.closeDropdown()
				} else {
					c.openDropdown()
				}
			} else {
				// Clicked entry area: take X11 focus, position cursor at click.
				// SetInputFocus alone is not enough: if X11 focus never left (because
				// the previous click was on a non-focusable widget), no FocusIn fires.
				widget.Focus(app, win)
				c.State |= StateFocus
				c.edit.SelAnchor = c.edit.ClosestGap(ev.X)
				c.edit.InsertPos = c.edit.SelAnchor
				c.edit.ClearSelection()
				c.Display()
			}
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 && c.arrowPressed {
			c.arrowPressed = false
			c.Display()
		}
	})

	// Key events for editable combobox (mirrors entry/bindings.go).
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		if c.State&StateDisabled != 0 {
			return
		}
		// The posted listbox's bindings in combobox.tcl: Escape unposts,
		// Up/Down move the highlight and Return takes the value. Escape
		// is consumed so a dialog's own Escape binding leaves it alone.
		if c.dropOpen {
			switch ev.KeySym {
			case platform.XK_Escape:
				ev.Handled = true
				c.closeDropdown()
				return
			case platform.XK_Up, platform.XK_Down:
				delta := 1
				if ev.KeySym == platform.XK_Up {
					delta = -1
				}
				c.dropSel = min(max(c.dropSel+delta, 0), len(c.Values)-1)
				c.displayDropdown()
				return
			case platform.XK_Return, platform.XK_space:
				ev.Handled = true
				if c.dropSel >= 0 && c.dropSel < len(c.Values) {
					c.Set(c.Values[c.dropSel])
				}
				c.closeDropdown()
				return
			}
		} else if ev.KeySym == platform.XK_Down && ev.State&platform.Mod1Mask != 0 {
			// <Alt-Down> posts the list (ttk::combobox::Post).
			c.openDropdown()
			return
		}
		if c.CbState != FieldNormal {
			return
		}

		switch ev.KeySym {
		case platform.XK_Left, platform.XK_Right, platform.XK_Home, platform.XK_End:
			c.edit.HandleNavKey(ev)
		case platform.XK_BackSpace, platform.XK_Delete, platform.XK_Insert:
			c.edit.HandleEditKey(ev)
		default:
			if ev.State&platform.ControlMask != 0 {
				c.edit.HandleCtrlKey(ev)
				return
			}
			c.edit.HandleKey(ev)
		}
	})
	app.Dispatcher().Bind(win.PlatformID, event.VirtualMask, func(ev *event.Event) {
		if c.State&StateDisabled == 0 && c.CbState == FieldNormal {
			c.edit.HandleVirtual(ev)
		}
	})

	// Focus events.
	app.Dispatcher().Bind(win.PlatformID, event.FocusChangeMask, func(ev *event.Event) {
		switch ev.Type {
		case event.FocusInType:
			c.ChangeState(StateFocus, 0)
		case event.FocusOutType:
			c.ChangeState(0, StateFocus)
			c.edit.ClearSelection()
			c.closeDropdown()
			c.Display()
		}
	})

	// Hide cursor when user clicks any other window (non-focusable widgets don't
	// call SetInputFocus, so FocusOut never fires for those clicks).
	app.Dispatcher().BindGlobalFor(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if c.State&StateFocus == 0 || c.CbState != FieldNormal {
			return
		}
		if ev.Window == win.PlatformID {
			return
		}
		if c.dropWin != nil && ev.Window == c.dropWin.PlatformID {
			return
		}
		c.State &^= StateFocus
		c.edit.ClearSelection()
		c.Display()
	})
}
