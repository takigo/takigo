// Package menu implements a popup menu widget.
// It ports the simplified core of tk/generic/tkMenu.c.
package menu

import (
	"log"

	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/font"
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
	Type        EntryType
	Label       string
	Command     func()
	SubMenu     *Menu
	Checked     bool
	State       widget.State
	AccelStr    string             // accelerator text for display
	Underline   int                // index of char to underline for keyboard nav; -1 = none
	Image       widget.WidgetImage // optional image
	SelectImage widget.WidgetImage // image shown when entry is active/highlighted
	Compound    widget.Compound    // how to combine image and text
	Fg          *color.ColorRef    // per-entry foreground (nil = use menu default)
	Bg          *color.ColorRef    // per-entry background (nil = use menu default)
	ColumnBreak bool               // start a new column before this entry
	HideMargin  bool               // suppress check/radio indicator margin
}

// colData holds layout data for one column in a multi-column menu.
type colData struct {
	entries []int // indices into Menu.entries
	x       int   // left x of this column
	width   int   // pixel width
}

// Menu is a popup menu with a list of entries.
type Menu struct {
	widget.Base

	entries       []MenuEntry
	activeIndex   int // -1 = none; -2 = tearoff region
	postedCascade *Menu
	parent        *Menu // non-nil when this menu is posted as a cascade of parent

	// Layout.
	entryHeight   int
	sepHeight     int
	menuWidth     int
	tearoffHeight int // height of tearoff grip area (0 if TearOff=false)
	cols          []colData

	// Colors.
	ActiveBg *color.ColorRef
	ActiveFg *color.ColorRef

	// State.
	posted                bool
	grabbed               bool
	motionSincePost       bool // true once pointer moves after Post(); gates first ButtonRelease
	suppressFocusOut      bool // set briefly when we ourselves call SetInputFocus for a cascade
	skipGlobalButtonPress bool // skip the first BindGlobal ButtonPress (the click that opened us)
	screenX               int  // absolute screen X set by Post()
	screenY               int  // absolute screen Y set by Post()

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
		if err != nil {
			log.Printf("menu: failed to get color %q: %v", name, err)
			return
		}
		m.Background = col
		m.UpdateBorder()
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
				platform.StructureNotifyMask |
				platform.FocusChangeMask),
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
	m.SetDisplayProc(m.display)
	w.OnDestroy(m.Destroy)
	w.Class = "Menu"
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
		Type:      Command,
		Label:     label,
		Command:   command,
		Underline: -1,
	})
}

// AddCommandUL adds a command entry with a specific underline index for keyboard navigation.
func (m *Menu) AddCommandUL(label string, underline int, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Command,
		Label:     label,
		Command:   command,
		Underline: underline,
	})
}

// AddSeparator adds a separator entry.
func (m *Menu) AddSeparator() {
	m.entries = append(m.entries, MenuEntry{Type: Separator, Underline: -1})
}

// AddCascade adds a cascade entry with a submenu.
func (m *Menu) AddCascade(label string, subMenu *Menu) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Cascade,
		Label:     label,
		SubMenu:   subMenu,
		Underline: -1,
	})
}

// AddCascadeUL adds a cascade entry with a specific underline index.
func (m *Menu) AddCascadeUL(label string, underline int, subMenu *Menu) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Cascade,
		Label:     label,
		SubMenu:   subMenu,
		Underline: underline,
	})
}

// AddCommandAccel adds a command entry with accelerator display text.
func (m *Menu) AddCommandAccel(label string, accel string, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Command,
		Label:     label,
		AccelStr:  accel,
		Command:   command,
		Underline: -1,
	})
}

// AddCommandAccelUL adds a command entry with accelerator text and underline index.
func (m *Menu) AddCommandAccelUL(label string, accel string, underline int, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Command,
		Label:     label,
		AccelStr:  accel,
		Command:   command,
		Underline: underline,
	})
}

// AddCommandImage adds a command entry with an image (and optional label).
// compound controls how the image and label are combined (CompoundLeft = image left of text).
func (m *Menu) AddCommandImage(label string, img widget.WidgetImage, compound widget.Compound, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Command,
		Label:     label,
		Image:     img,
		Compound:  compound,
		Command:   command,
		Underline: -1,
	})
}

// AddImageSwatch adds an image-only entry for a palette/grid menu.
// img is drawn normally; selectImg is drawn when the entry is active.
// columnBreak starts a new column before this entry.
func (m *Menu) AddImageSwatch(img, selectImg widget.WidgetImage, columnBreak bool, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:        Command,
		Image:       img,
		SelectImage: selectImg,
		ColumnBreak: columnBreak,
		HideMargin:  true,
		Compound:    widget.CompoundNone,
		Command:     command,
		Underline:   -1,
	})
}

// AddCommandBg adds a command entry with per-entry foreground and background colors.
// Pass empty string for fg/bg to use menu defaults.
func (m *Menu) AddCommandBg(label, fg, bg string, command func()) {
	e := MenuEntry{
		Type:      Command,
		Label:     label,
		Command:   command,
		Underline: -1,
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
		Type:      Checkbutton,
		Label:     label,
		Checked:   checked,
		Command:   command,
		Underline: -1,
	})
}

// AddRadiobutton adds a radiobutton entry. It behaves like a command
// but is displayed with a radio-style indicator when checked.
func (m *Menu) AddRadiobutton(label string, checked bool, command func()) {
	m.entries = append(m.entries, MenuEntry{
		Type:      Radiobutton,
		Label:     label,
		Checked:   checked,
		Command:   command,
		Underline: -1,
	})
}

// Entries returns the menu entries.
func (m *Menu) Entries() []MenuEntry {
	return m.entries
}

// PostFromButton maps the menu at screen coordinates (x, y) and arranges
// for the BindGlobal handler to ignore the ButtonPress that triggered this
// call (so the click that opens the menu doesn't also immediately close it).
// Call this instead of Post() when the caller is in a ButtonPress handler.
func (m *Menu) PostFromButton(x, y int) {
	m.skipGlobalButtonPress = true
	m.Post(x, y)
}

// Post maps the menu at screen coordinates (x, y).
func (m *Menu) Post(x, y int) {
	m.computeGeometry()
	w := m.Win
	d := w.Display.Server

	m.screenX = x
	m.screenY = y
	m.motionSincePost = false
	d.MoveResizeWindow(w.PlatformID, x, y, uint(w.Width), uint(w.Height))
	d.MapRaised(w.PlatformID)
	m.posted = true
	m.activeIndex = -1

	// Sync so the X server has processed MapRaised before GrabPointer:
	// XGrabPointer requires the grab window to be viewable.
	d.Sync(false)

	// Grab pointer with owner_events=true so that events over our own client
	// windows are still delivered normally (hover effects, cursor shapes).
	// Only clicks outside all client windows are redirected to the grab window.
	// Keyboard events reach the menu via SetInputFocus (no keyboard grab).
	// skipGlobalButtonPress may have been set by PostFromButton() to skip the
	// ButtonPress event that caused this Post() call; leave it as-is here.
	const grabMask = uint(platform.ButtonPressMask | platform.ButtonReleaseMask)
	ret := d.GrabPointer(w.PlatformID, true, grabMask,
		platform.GrabModeAsync, platform.GrabModeAsync,
		platform.WindowID(0), platform.CursorID(0), platform.CurrentTime)
	if ret != platform.GrabSuccess {
		// Grab failed (another client holds the grab); retry once after a flush.
		d.Sync(false)
		ret = d.GrabPointer(w.PlatformID, true, grabMask,
			platform.GrabModeAsync, platform.GrabModeAsync,
			platform.WindowID(0), platform.CursorID(0), platform.CurrentTime)
	}
	m.grabbed = ret == platform.GrabSuccess
	d.SetInputFocus(w.PlatformID, platform.RevertToParent, platform.CurrentTime)

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
		m.grabbed = false
	}

	d.UnmapWindow(w.PlatformID)
	m.posted = false
	m.activeIndex = -1
	m.parent = nil
}

// IsPosted returns whether the menu is currently posted.
func (m *Menu) IsPosted() bool {
	return m.posted
}

// PrepareGeometry computes the menu's required size without posting it.
// Call before reading Win.ReqWidth / Win.ReqHeight to position the menu.
func (m *Menu) PrepareGeometry() {
	m.computeGeometry()
}

func (m *Menu) computeGeometry() {
	if m.Font == nil {
		return
	}

	fm := m.Font.Metrics()
	m.entryHeight = fm.Linespace() + 6
	m.sepHeight = 6

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
		if e.SelectImage != nil && e.SelectImage.Height()+6 > m.entryHeight {
			m.entryHeight = e.SelectImage.Height() + 6
		}
	}

	// Build columns.
	m.cols = m.buildColumns()

	if len(m.cols) <= 1 {
		// Single-column layout.
		maxWidth := 0
		totalHeight := 2*m.BorderWidth + m.tearoffHeight
		for _, e := range m.entries {
			if e.Type == Separator {
				totalHeight += m.sepHeight
			} else {
				totalHeight += m.entryHeight
				w := m.entryContentWidth(e)
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
		// Store single-column data.
		m.cols[0].x = m.BorderWidth
		m.cols[0].width = m.menuWidth - 2*m.BorderWidth
		return
	}

	// Multi-column layout.
	x := m.BorderWidth
	maxColHeight := 0
	for ci := range m.cols {
		col := &m.cols[ci]
		col.x = x
		maxW := 0
		colH := 0
		for _, idx := range col.entries {
			e := m.entries[idx]
			w := m.entryContentWidth(e)
			if w > maxW {
				maxW = w
			}
			if e.Type == Separator {
				colH += m.sepHeight
			} else {
				colH += m.entryHeight
			}
		}
		col.width = maxW
		if colH > maxColHeight {
			maxColHeight = colH
		}
		x += maxW
	}

	totalWidth := x + m.BorderWidth
	if totalWidth < 60 {
		totalWidth = 60
	}
	totalHeight := 2*m.BorderWidth + m.tearoffHeight + maxColHeight
	m.menuWidth = totalWidth

	w := m.Win
	w.Width = totalWidth
	w.Height = totalHeight
	w.ReqWidth = totalWidth
	w.ReqHeight = totalHeight
}

func (m *Menu) buildColumns() []colData {
	if len(m.entries) == 0 {
		return []colData{{}}
	}
	var cols []colData
	cur := colData{}
	for i, e := range m.entries {
		if e.ColumnBreak && i > 0 && len(cur.entries) > 0 {
			cols = append(cols, cur)
			cur = colData{}
		}
		cur.entries = append(cur.entries, i)
	}
	cols = append(cols, cur)
	return cols
}

func (m *Menu) entryContentWidth(e MenuEntry) int {
	if e.HideMargin {
		if e.Image != nil {
			return e.Image.Width() + 4
		}
		if e.SelectImage != nil {
			return e.SelectImage.Width() + 4
		}
		return 20
	}
	var w int
	if e.Image != nil && e.Compound == widget.CompoundNone && e.Label == "" {
		w = e.Image.Width()
	} else {
		w = m.Font.MeasureString(e.Label)
		if e.Image != nil {
			w += e.Image.Width() + 4
		}
	}
	w += 40
	if e.AccelStr != "" {
		w += m.Font.MeasureString(e.AccelStr) + 20
	}
	return w
}

// entryAt returns the entry index at pixel (x, y). Returns -1 for separator/nothing,
// -2 for tearoff grip region.
func (m *Menu) entryAt(x, y int) int {
	offset := m.BorderWidth
	if m.TearOff {
		if y >= offset && y < offset+m.tearoffHeight {
			return -2
		}
		offset += m.tearoffHeight
	}

	if len(m.cols) <= 1 {
		// Single-column: use y only.
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

	// Multi-column: find column by x.
	var col *colData
	for i := range m.cols {
		c := &m.cols[i]
		if x >= c.x && x < c.x+c.width {
			col = c
			break
		}
	}
	if col == nil {
		return -1
	}
	yPos := offset
	for _, idx := range col.entries {
		e := m.entries[idx]
		var h int
		if e.Type == Separator {
			h = m.sepHeight
		} else {
			h = m.entryHeight
		}
		if y >= yPos && y < yPos+h {
			if e.Type == Separator {
				return -1
			}
			return idx
		}
		yPos += h
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

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (m *Menu) Display() {
	m.EventuallyRedraw()
}

// display draws the menu.
func (m *Menu) display() {
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
	yStart := m.BorderWidth

	// Draw tearoff grip (dashed line).
	if m.TearOff {
		isActive := m.activeIndex == -2
		if isActive && m.ActiveBg != nil {
			d.SetForeground(gc, m.ActiveBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yStart,
				uint(w.Width-2*m.BorderWidth), uint(m.tearoffHeight))
		}
		var dashColor uint64 = d.BlackPixel(0)
		if m.Foreground != nil {
			dashColor = m.Foreground.Pixel
		}
		if isActive && m.ActiveFg != nil {
			dashColor = m.ActiveFg.Pixel
		}
		d.SetForeground(gc, dashColor)
		cx := m.BorderWidth + 4
		cy := yStart + m.tearoffHeight/2
		for cx+8 < w.Width-m.BorderWidth {
			d.FillRectangle(w.Drawable(), gc, cx, cy-1, 6, 2)
			cx += 10
		}
		yStart += m.tearoffHeight
	}

	if len(m.cols) <= 1 {
		m.displaySingleColumn(d, gc, df, fm, yStart)
	} else {
		m.displayMultiColumn(d, gc, yStart)
	}

	d.Flush()
}

func (m *Menu) displaySingleColumn(d platform.DisplayServer, gc platform.GCID, df platform.DrawableFont, fm font.Metrics, yStart int) {
	w := m.Win
	yPos := yStart

	for i, e := range m.entries {
		if e.Type == Separator {
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

		if isActive && m.ActiveBg != nil {
			d.SetForeground(gc, m.ActiveBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yPos,
				uint(w.Width-2*m.BorderWidth), uint(m.entryHeight))
		} else if !isActive && e.Bg != nil {
			d.SetForeground(gc, e.Bg.Pixel)
			d.FillRectangle(w.Drawable(), gc, m.BorderWidth, yPos,
				uint(w.Width-2*m.BorderWidth), uint(m.entryHeight))
		}

		textX := m.BorderWidth + 20
		if e.HideMargin {
			textX = m.BorderWidth + 2
		}
		textY := yPos + (m.entryHeight-fm.Linespace())/2 + fm.Ascent

		img := e.Image
		if isActive && e.SelectImage != nil {
			img = e.SelectImage
		}
		if img != nil {
			imgW := img.Width()
			imgH := img.Height()
			imgY := yPos + (m.entryHeight-imgH)/2
			imgX := textX
			if e.Compound != widget.CompoundNone || e.Label != "" {
				textX += imgW + 4
			}
			bgPx := uint64(0xD9D9D9)
			if e.Bg != nil {
				bgPx = e.Bg.Pixel
			} else if m.Background != nil {
				bgPx = m.Background.Pixel
			}
			img.Draw(d, w.Drawable(), w.GC, w.Depth,
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
			if !e.HideMargin {
				if e.Type == Checkbutton && e.Checked {
					df.DrawString(w.Drawable(), m.BorderWidth+4, textY, "\u2713",
						fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
				} else if e.Type == Radiobutton && e.Checked {
					df.DrawString(w.Drawable(), m.BorderWidth+4, textY, "\u25cf",
						fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
				}
			}

			if e.Label != "" {
				df.DrawString(w.Drawable(), textX, textY, e.Label,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
				runes := []rune(e.Label)
				if e.Underline >= 0 && e.Underline < len(runes) {
					prefix := string(runes[:e.Underline])
					ch := string(runes[e.Underline])
					ulX := textX + m.Font.MeasureString(prefix)
					ulW := m.Font.MeasureString(ch)
					ulY := textY + 2
					d.SetForeground(gc, fgCol.Pixel)
					d.DrawLine(w.Drawable(), gc, ulX, ulY, ulX+ulW, ulY)
				}
			}

			if e.AccelStr != "" {
				accelW := m.Font.MeasureString(e.AccelStr)
				accelX := w.Width - m.BorderWidth - accelW - 8
				df.DrawString(w.Drawable(), accelX, textY, e.AccelStr,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}

			if e.Type == Cascade {
				arrowX := w.Width - m.BorderWidth - 14
				df.DrawString(w.Drawable(), arrowX, textY, "\u25b6",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}
		}

		yPos += m.entryHeight
	}
}

func (m *Menu) displayMultiColumn(d platform.DisplayServer, gc platform.GCID, yStart int) {
	w := m.Win

	for _, col := range m.cols {
		yPos := yStart
		for _, idx := range col.entries {
			e := m.entries[idx]
			if e.Type == Separator {
				yPos += m.sepHeight
				continue
			}

			isActive := idx == m.activeIndex && e.State != widget.StateDisabled

			if isActive && m.ActiveBg != nil {
				d.SetForeground(gc, m.ActiveBg.Pixel)
				d.FillRectangle(w.Drawable(), gc, col.x, yPos, uint(col.width), uint(m.entryHeight))
			} else if !isActive && e.Bg != nil {
				d.SetForeground(gc, e.Bg.Pixel)
				d.FillRectangle(w.Drawable(), gc, col.x, yPos, uint(col.width), uint(m.entryHeight))
			}

			img := e.Image
			if isActive && e.SelectImage != nil {
				img = e.SelectImage
			}
			if img != nil {
				imgW := img.Width()
				imgH := img.Height()
				imgX := col.x + (col.width-imgW)/2
				imgY := yPos + (m.entryHeight-imgH)/2
				bgPx := uint64(0xD9D9D9)
				if m.Background != nil {
					bgPx = m.Background.Pixel
				}
				img.Draw(d, w.Drawable(), w.GC, w.Depth,
					0, 0, imgW, imgH, imgX, imgY, bgPx)
			}

			yPos += m.entryHeight
		}
	}
}

// activate sets the active entry and redraws.
func (m *Menu) activate(index int) {
	if index == m.activeIndex {
		return
	}
	m.activeIndex = index
	m.Display()
}

// unpostChain unposts this menu and all ancestor menus in the cascade chain.
func (m *Menu) unpostChain() {
	root := m
	for root.parent != nil {
		root = root.parent
	}
	root.Unpost()
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
		m.unpostChain()
		if e.Command != nil {
			e.Command()
		}
	case Checkbutton:
		e.Checked = !e.Checked
		m.unpostChain()
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
		m.unpostChain()
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

	// Position submenu to the right of this entry, using stored screen coords.
	w := m.Win
	entryY := m.entryY(index)
	subX := m.screenX + w.Width
	subY := m.screenY + entryY

	// Transfer pointer grab to submenu.
	// Suppress the FocusOut that fires on this menu when SetInputFocus moves
	// focus to the submenu — we don't want to unpost the parent in that case.
	m.suppressFocusOut = true
	d := w.Display.Server
	if m.grabbed {
		d.UngrabPointer(platform.CurrentTime)
		m.grabbed = false
	}

	e.SubMenu.parent = m
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
