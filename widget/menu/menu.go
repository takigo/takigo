// Package menu implements a popup menu widget.
// It ports the simplified core of tk/generic/tkMenu.c.
package menu

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// EntryType is the type of a menu entry.
type EntryType int

const (
	Command EntryType = iota
	Separator
	Cascade
	Checkbutton
	Radiobutton
)

// MenuEntry represents a single entry in a menu.
type MenuEntry struct {
	Type     EntryType
	Label    string
	Command  func()
	SubMenu  *Menu
	Checked  bool
	State    widget.State
	AccelStr string             // accelerator text for display
	Image    widget.WidgetImage // optional image
	Compound widget.Compound    // how to combine image and text
	Fg       *color.ColorRef   // per-entry foreground (nil = use menu default)
	Bg       *color.ColorRef   // per-entry background (nil = use menu default)
}

// Menu is a popup menu with a list of entries.
type Menu struct {
	widget.Base

	entries       []MenuEntry
	activeIndex   int // -1 = none; -2 = tearoff region
	postedCascade *Menu

	// Layout.
	entryHeight   int
	sepHeight     int
	menuWidth     int
	tearoffHeight int // height of tearoff grip area (0 if TearOff=false)

	// Colors.
	ActiveBg *color.ColorRef
	ActiveFg *color.ColorRef

	// State.
	posted  bool
	grabbed bool

	// TearOff enables a tearoff grip at the top of the menu.
	TearOff bool

	app widget.AppContext
}

// MenuOption configures a Menu.
type MenuOption func(*Menu)

// TearOffOpt enables or disables the tearoff grip at the top of the menu.
func TearOffOpt(on bool) MenuOption {
	return func(m *Menu) { m.TearOff = on }
}

func Background(name string) MenuOption {
	return func(m *Menu) {
		col, err := m.App.ColorCache().Get(name)
		if err == nil {
			m.Background = col
			m.UpdateBorder()
		}
	}
}

// New creates a new Menu. The menu is an override-redirect window,
// initially unmapped, created as a child of the root X window.
func New(parent widget.Caregiver, name string, opts ...MenuOption) *Menu {
	app := parent.AppContext()
	d := parent.Window().Display

	// Create override-redirect toplevel window.
	attrs := &platform.WindowAttrs{
		BackgroundPixel:  d.WhitePixel,
		BorderPixel:      d.BlackPixel,
		OverrideRedirect: true,
		EventMask: int64(
			platform.KeyPressMask |
				platform.KeyReleaseMask |
				platform.ButtonPressMask |
				platform.ButtonReleaseMask |
				platform.PointerMotionMask |
				platform.EnterWindowMask |
				platform.LeaveWindowMask |
				platform.ExposureMask |
				platform.StructureNotifyMask),
	}

	xwin := d.Server.CreateWindow(
		d.RootWindow,
		0, 0, 1, 1, 1,
		d.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWOverrideRedirect|platform.CWEventMask,
		attrs,
	)

	w := &window.Window{
		PlatformID:      xwin,
		Display:         d,
		Parent:          parent.Window(),
		Name:            name,
		PathName:        window.BuildPathName(parent.Window(), name),
		Width:           1,
		Height:          1,
		ReqWidth:        1,
		ReqHeight:       1,
		Depth:           d.Depth,
		BackgroundPixel: d.WhitePixel,
	}

	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})

	d.RegisterWindow(xwin, w)
	parent.Window().AddChild(w)

	m := &Menu{
		activeIndex: -1,
		entryHeight: 24,
		sepHeight:   6,
		app:         app,
	}
	widget.InitBase(&m.Base, w, app)
	m.BorderWidth = 1
	m.Relief = option.ReliefRaised

	// Default active colors.
	if abg, err := app.ColorCache().Get("#3399ff"); err == nil {
		m.ActiveBg = abg.Ref()
	}
	if afg, err := app.ColorCache().Get("#ffffff"); err == nil {
		m.ActiveFg = afg.Ref()
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.Background != nil {
		w.BackgroundPixel = m.Background.Pixel
	}

	bindMenu(m, app)
	return m
}

// AddCommand adds a command entry to the menu.
func (m *Menu) AddCommand(label string, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:    Command,
		Label:   label,
		Command: command,
	})
}

// AddSeparator adds a separator entry.
func (m *Menu) AddSeparator() {
	m.entries = append(m.entries, MenuEntry{Type: Separator})
}

// AddCascade adds a cascade entry with a submenu.
func (m *Menu) AddCascade(label string, subMenu *Menu) {
	m.entries = append(m.entries, MenuEntry{
		Type:    Cascade,
		Label:   label,
		SubMenu: subMenu,
	})
}

// AddCommandAccel adds a command entry with accelerator display text.
func (m *Menu) AddCommandAccel(label string, accel string, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:     Command,
		Label:    label,
		AccelStr: accel,
		Command:  command,
	})
}

// AddCommandImage adds a command entry with an image (and optional label).
// compound controls how the image and label are combined (CompoundLeft = image left of text).
func (m *Menu) AddCommandImage(label string, img widget.WidgetImage, compound widget.Compound, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:     Command,
		Label:    label,
		Image:    img,
		Compound: compound,
		Command:  command,
	})
}

// AddCommandBg adds a command entry with per-entry foreground and background colors.
// Pass empty string for fg/bg to use menu defaults.
func (m *Menu) AddCommandBg(label, fg, bg string, command func()) {
	e := MenuEntry{
		Type:    Command,
		Label:   label,
		Command: command,
	}
	if fg != "" {
		if col, err := m.App.ColorCache().Get(fg); err == nil {
			e.Fg = col.Ref()
		}
	}
	if bg != "" {
		if col, err := m.App.ColorCache().Get(bg); err == nil {
			e.Bg = col.Ref()
		}
	}
	m.entries = append(m.entries, e)
}

// AddCheckbutton adds a checkbutton entry.
func (m *Menu) AddCheckbutton(label string, checked bool, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:    Checkbutton,
		Label:   label,
		Checked: checked,
		Command: command,
	})
}

// AddRadiobutton adds a radiobutton entry. It behaves like a command
// but is displayed with a radio-style indicator when checked.
func (m *Menu) AddRadiobutton(label string, checked bool, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:    Radiobutton,
		Label:   label,
		Checked: checked,
		Command: command,
	})
}

// Entries returns the menu entries.
func (m *Menu) Entries() []MenuEntry {
	return m.entries
}

// Post maps the menu at screen coordinates (x, y).
func (m *Menu) Post(x, y int) {
	m.computeGeometry()
	w := m.Win
	d := w.Display.Server

	d.MoveResizeWindow(w.PlatformID, x, y, uint(w.Width), uint(w.Height))
	d.MapRaised(w.PlatformID)
	m.posted = true
	m.activeIndex = -1

	// Grab pointer and keyboard with owner_events=false so all pointer
	// events go to the menu window. This ensures clicks outside the menu
	// (including inside other app windows) are caught and close the menu.
	d.GrabPointer(w.PlatformID, false,
		uint(platform.ButtonPressMask|platform.ButtonReleaseMask|platform.PointerMotionMask|platform.EnterWindowMask|platform.LeaveWindowMask),
		platform.GrabModeAsync, platform.GrabModeAsync,
		platform.WindowID(0), platform.CursorID(0), platform.CurrentTime)
	d.GrabKeyboard(w.PlatformID, false, platform.GrabModeAsync, platform.GrabModeAsync, platform.CurrentTime)
	m.grabbed = true

	m.Display()
}

// Unpost unmaps the menu and releases grabs.
func (m *Menu) Unpost() {
	if !m.posted {
		return
	}

	// Unpost any cascade submenu.
	if m.postedCascade != nil {
		m.postedCascade.Unpost()
		m.postedCascade = nil
	}

	w := m.Win
	d := w.Display.Server

	if m.grabbed {
		d.UngrabPointer(platform.CurrentTime)
		d.UngrabKeyboard(platform.CurrentTime)
		m.grabbed = false
	}

	d.UnmapWindow(w.PlatformID)
	m.posted = false
	m.activeIndex = -1
}

// IsPosted returns whether the menu is currently posted.
func (m *Menu) IsPosted() bool {
	return m.posted
}

func (m *Menu) computeGeometry() {
	if m.Font == nil {
		return
	}

	fm := m.Font.Metrics()
	m.entryHeight = fm.Linespace() + 6
	m.sepHeight = 6

	// Tearoff grip area.
	if m.TearOff {
		m.tearoffHeight = 10
	} else {
		m.tearoffHeight = 0
	}

	// Expand entryHeight if any entry has a tall image.
	for _, e := range m.entries {
		if e.Image != nil && e.Image.Height()+6 > m.entryHeight {
			m.entryHeight = e.Image.Height() + 6
		}
	}

	maxWidth := 0
	totalHeight := 2*m.BorderWidth + m.tearoffHeight

	for _, e := range m.entries {
		if e.Type == Separator {
			totalHeight += m.sepHeight
		} else {
			totalHeight += m.entryHeight
			// Width: image + text (or image alone when CompoundNone and image set).
			var w int
			if e.Image != nil && e.Compound == widget.CompoundNone && e.Label == "" {
				w = e.Image.Width()
			} else {
				w = m.Font.MeasureString(e.Label)
				if e.Image != nil {
					w += e.Image.Width() + 4
				}
			}
			// Add space for check indicator and cascade arrow.
			w += 40
			if e.AccelStr != "" {
				w += m.Font.MeasureString(e.AccelStr) + 20
			}
			if w > maxWidth {
				maxWidth = w
			}
		}
	}

	m.menuWidth = maxWidth + 2*m.BorderWidth
	if m.menuWidth < 60 {
		m.menuWidth = 60
	}

	w := m.Win
	w.Width = m.menuWidth
	w.Height = totalHeight
	w.ReqWidth = m.menuWidth
	w.ReqHeight = totalHeight
}

// entryAtY returns the entry index at pixel y, -1 for separator/nothing,
// or -2 for the tearoff grip region.
func (m *Menu) entryAtY(y int) int {
	offset := m.BorderWidth
	if m.TearOff {
		if y >= offset && y < offset+m.tearoffHeight {
			return -2 // tearoff region
		}
		offset += m.tearoffHeight
	}
	for i, e := range m.entries {
		var h int
		if e.Type == Separator {
			h = m.sepHeight
		} else {
			h = m.entryHeight
		}
		if y >= offset && y < offset+h {
			if e.Type == Separator {
				return -1
			}
			return i
		}
		offset += h
	}
	return -1
}

// entryY returns the top y coordinate of entry i.
func (m *Menu) entryY(i int) int {
	y := m.BorderWidth + m.tearoffHeight
	for j := 0; j < i; j++ {
		if m.entries[j].Type == Separator {
			y += m.sepHeight
		} else {
			y += m.entryHeight
		}
	}
	return y
}

// Display draws the menu.
func (m *Menu) Display() {
	if m.Destroyed || !m.posted {
		return
	}
	w := m.Win
	if w.PlatformID == 0 {
		return
	}

	d := w.Display.Server
	gc := w.GC

	// Background.
	if m.Background != nil {
		d.SetForeground(gc, m.Background.Pixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Outer border.
	if m.Border != nil && m.BorderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, m.Border,
			0, 0, w.Width, w.Height, m.BorderWidth, option.ReliefRaised)
	}

	df, isDF := m.Font.(platform.DrawableFont)
	if !isDF {
		d.Flush()
		return
	}

	fm := m.Font.Metrics()
	yPos := m.BorderWidth

	// Draw tearoff grip (dashed line).
	if m.TearOff {
		isActive := m.activeIndex == -2
		if isActive && m.ActiveBg != nil {
			d.SetForeground(gc, m.ActiveBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yPos,
				uint(w.Width-2*m.BorderWidth), uint(m.tearoffHeight))
		}
		// Dashed line: alternating segments.
		var dashColor uint64 = d.BlackPixel(0)
		if m.Foreground != nil {
			dashColor = m.Foreground.Pixel
		}
		if isActive && m.ActiveFg != nil {
			dashColor = m.ActiveFg.Pixel
		}
		d.SetForeground(gc, dashColor)
		cx := m.BorderWidth + 4
		cy := yPos + m.tearoffHeight/2
		for cx+8 < w.Width-m.BorderWidth {
			d.FillRectangle(w.Drawable(), gc, cx, cy-1, 6, 2)
			cx += 10
		}
		yPos += m.tearoffHeight
	}

	for i, e := range m.entries {
		if e.Type == Separator {
			// Draw separator line.
			sepY := yPos + m.sepHeight/2
			if m.Border != nil {
				draw.Draw3DRectangle(d, w.Drawable(), gc, m.Border,
					m.BorderWidth+2, sepY-1, w.Width-2*m.BorderWidth-4, 2,
					1, option.ReliefSunken)
			}
			yPos += m.sepHeight
			continue
		}

		isActive := i == m.activeIndex && e.State != widget.StateDisabled

		// Per-entry or active background fill.
		if isActive && m.ActiveBg != nil {
			d.SetForeground(gc, m.ActiveBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yPos,
				uint(w.Width-2*m.BorderWidth), uint(m.entryHeight))
		} else if !isActive && e.Bg != nil {
			d.SetForeground(gc, e.Bg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yPos,
				uint(w.Width-2*m.BorderWidth), uint(m.entryHeight))
		}

		// Text and image.
		textX := m.BorderWidth + 20
		textY := yPos + (m.entryHeight-fm.Linespace())/2 + fm.Ascent

		// Draw image if present.
		if e.Image != nil {
			imgW := e.Image.Width()
			imgH := e.Image.Height()
			imgY := yPos + (m.entryHeight-imgH)/2
			var imgX int
			if e.Compound == widget.CompoundNone && e.Label == "" {
				// Image only — center where text would be.
				imgX = textX
			} else {
				// Image left of text.
				imgX = textX
				textX += imgW + 4
			}
			bgPx := uint64(0xD9D9D9)
			if e.Bg != nil {
				bgPx = e.Bg.Pixel
			} else if m.Background != nil {
				bgPx = m.Background.Pixel
			}
			e.Image.Draw(d, w.Drawable(), w.GC, w.Depth,
				0, 0, imgW, imgH, imgX, imgY, bgPx)
		}

		var fgCol *color.ColorRef
		if e.State == widget.StateDisabled {
			if dfg, err := m.App.ColorCache().Get(widget.DefDisabledForeground); err == nil {
				fgCol = dfg.Ref()
			}
		} else if isActive && m.ActiveFg != nil {
			fgCol = m.ActiveFg
		} else if e.Fg != nil {
			fgCol = e.Fg
		} else if m.Foreground != nil {
			fgCol = m.Foreground.Ref()
		}

		if fgCol != nil {
			// Check/Radio indicator.
			if e.Type == Checkbutton && e.Checked {
				df.DrawString(w.Drawable(), m.BorderWidth+4, textY, "\u2713",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			} else if e.Type == Radiobutton && e.Checked {
				df.DrawString(w.Drawable(), m.BorderWidth+4, textY, "\u25cf",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}

			// Label (skip if image-only).
			if e.Label != "" {
				df.DrawString(w.Drawable(), textX, textY, e.Label,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}

			// Accelerator text.
			if e.AccelStr != "" {
				accelW := m.Font.MeasureString(e.AccelStr)
				accelX := w.Width - m.BorderWidth - accelW - 8
				df.DrawString(w.Drawable(), accelX, textY, e.AccelStr,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}

			// Cascade arrow.
			if e.Type == Cascade {
				arrowX := w.Width - m.BorderWidth - 14
				df.DrawString(w.Drawable(), arrowX, textY, "\u25b6",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}
		}

		yPos += m.entryHeight
	}

	d.Flush()
}

// activate sets the active entry and redraws.
func (m *Menu) activate(index int) {
	if index == m.activeIndex {
		return
	}
	m.activeIndex = index
	m.Display()
}

// invoke invokes the active entry.
func (m *Menu) invoke(index int) {
	if index < 0 || index >= len(m.entries) {
		return
	}
	e := &m.entries[index]
	if e.State == widget.StateDisabled {
		return
	}

	switch e.Type {
	case Command:
		m.Unpost()
		if e.Command != nil {
			e.Command()
		}
	case Checkbutton:
		e.Checked = !e.Checked
		m.Unpost()
		if e.Command != nil {
			e.Command()
		}
	case Radiobutton:
		// Uncheck all other radiobuttons in this menu, check this one.
		for j := range m.entries {
			if m.entries[j].Type == Radiobutton && j != index {
				m.entries[j].Checked = false
			}
		}
		e.Checked = true
		m.Unpost()
		if e.Command != nil {
			e.Command()
		}
	case Cascade:
		if e.SubMenu != nil {
			m.postCascade(index)
		}
	}
}

func (m *Menu) postCascade(index int) {
	e := &m.entries[index]
	if e.SubMenu == nil {
		return
	}

	// Unpost old cascade.
	if m.postedCascade != nil && m.postedCascade != e.SubMenu {
		m.postedCascade.Unpost()
	}

	// Position submenu to the right of this entry.
	w := m.Win
	entryY := m.entryY(index)
	subX := w.X + w.Width
	subY := w.Y + entryY

	// Transfer grab temporarily.
	d := w.Display.Server
	if m.grabbed {
		d.UngrabPointer(platform.CurrentTime)
		d.UngrabKeyboard(platform.CurrentTime)
		m.grabbed = false
	}

	e.SubMenu.Post(subX, subY)
	m.postedCascade = e.SubMenu
}

// Configure applies options.
func (m *Menu) Configure(opts ...option.Option) {
	option.Apply(m, opts)
	m.UpdateBorder()
	if m.Background != nil {
		m.Win.BackgroundPixel = m.Background.Pixel
	}
	m.Display()
}

// Destroy cleans up.
func (m *Menu) Destroy() {
	if m.Destroyed {
		return
	}
	m.Unpost()
	m.Destroyed = true
	window.DestroyWindow(m.Win)
}
