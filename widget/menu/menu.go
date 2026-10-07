// Package menu implements a popup menu widget.
// It ports the simplified core of tk/generic/tkMenu.c.
package menu

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
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

	// widths caches text widths in widthFont (labels, accelerators and
	// underline prefixes), measured once rather than on every redraw.
	widths    map[string]int
	widthFont font.Font

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
	motionSincePost       bool           // true once pointer moves after Post(); gates first ButtonRelease
	suppressFocusOut      bool           // set briefly when we ourselves call SetInputFocus for a cascade
	skipGlobalButtonPress bool           // skip the first BindGlobal ButtonPress (the click that opened us)
	onUnpost              func()         // run once when the menu is next unposted (menubar deactivation)
	prevFocus             *window.Window // focus to restore on unpost
	screenX               int            // absolute screen X set by Post()
	screenY               int            // absolute screen Y set by Post()

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

// Background sets the background colour.
func Background[C color.Spec](name C) MenuOption {
	return func(m *Menu) { m.SetBackgroundColor(name) }
}

// FontOpt sets the font of the menu entries.
func FontOpt[F font.Spec](name F) MenuOption {
	return func(m *Menu) { m.SetFont(name) }
}

// New creates a new Menu. The menu is an override-redirect window,
// initially unmapped, created as a child of the root X window.
func New(parent widget.Caregiver, name string, opts ...MenuOption) *Menu {
	app := parent.AppContext()

	// An override-redirect window under the root.
	w := window.NewTopLevelWindow(parent.Window(), name, window.TopLevelSpec{
		Width:            1,
		Height:           1,
		BorderWidth:      1,
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
	})

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
	if abg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectBackground); err == nil {
		m.ActiveBg = abg.Ref()
	}
	if afg, err := app.ColorCache().Get(widget.PaletteFor(app).SelectForeground); err == nil {
		m.ActiveFg = afg.Ref()
	}

	for _, opt := range opts {
		opt(m)
	}

	if m.Background != nil {
		w.SetBackgroundPixel(m.Background.Pixel)
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

// textWidth returns the width of s in the menu's font, cached.
func (m *Menu) textWidth(s string) int {
	if m.Font == nil {
		return 0
	}
	if m.widthFont != m.Font {
		m.widths, m.widthFont = map[string]int{}, m.Font
	}
	w, ok := m.widths[s]
	if !ok {
		w = m.Font.MeasureString(s)
		m.widths[s] = w
	}
	return w
}

// postedKey is the Window.Value key under which a posted menu's window
// holds its *Menu.
var postedKey = new(window.ValueKey)

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
	// Keyboard events reach the menu as the focus window (no keyboard
	// grab); like tk::MenuUnpost, the old focus is restored on unpost. A
	// menu posted while another menu has the focus inherits the focus
	// that menu will restore.
	// skipGlobalButtonPress may have been set by PostFromButton() to skip the
	// ButtonPress event that caused this Post() call; leave it as-is here.
	// Focus first: the FocusOut it sends can unpost another menu, which
	// releases the pointer grab this menu then takes.
	prev := widget.FocusWindow(m.App)
	if pm, _ := prev.Value(postedKey).(*Menu); pm != nil {
		prev = pm.prevFocus
	}
	if prev != w {
		m.prevFocus = prev
	}
	w.SetValue(postedKey, m)
	widget.Focus(m.App, w)
	m.grab()

	m.Display()
}

// grab takes the pointer grab for the posted menu.
func (m *Menu) grab() {
	w := m.Win
	d := w.Display.Server
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
	m.skipGlobalButtonPress = false
	w.SetValue(postedKey, nil)
	prev := m.prevFocus
	m.prevFocus = nil
	// A cascade handed its parent's grab over when it was posted; give it
	// back so the still-posted parent keeps tracking clicks elsewhere.
	if p := m.parent; p != nil {
		if p.postedCascade == m {
			p.postedCascade = nil
		}
		if p.posted && !p.grabbed {
			p.grab()
			widget.Focus(m.App, p.Win)
		}
	}
	// Give the focus back if this menu still has it.
	if widget.FocusWindow(m.App) == w && prev != nil && !prev.IsDestroyed() {
		widget.Focus(m.App, prev)
	}
	m.parent = nil
	if fn := m.onUnpost; fn != nil {
		m.onUnpost = nil
		fn()
	}
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
		m.menuWidth = max(maxWidth+2*m.BorderWidth, 60)
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

	totalWidth := max(x+m.BorderWidth, 60)
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
		w = m.textWidth(e.Label)
		if e.Image != nil {
			w += e.Image.Width() + 4
		}
	}
	w += 40
	if e.AccelStr != "" {
		w += m.textWidth(e.AccelStr) + 20
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
	for j := range i {
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
	if m.Destroyed() || !m.posted {
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
		dashColor := d.BlackPixel(0)
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
			if dfg, err := m.App.ColorCache().Get(widget.PaletteFor(m.App).DisabledForeground); err == nil {
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
					ulX := textX + m.textWidth(prefix)
					ulW := m.textWidth(ch)
					ulY := textY + 2
					d.SetForeground(gc, fgCol.Pixel)
					d.DrawLine(w.Drawable(), gc, ulX, ulY, ulX+ulW, ulY)
				}
			}

			if e.AccelStr != "" {
				accelW := m.textWidth(e.AccelStr)
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
	default:
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
func (m *Menu) Configure(opts ...MenuOption) error {
	return widget.Configure(m, opts, m.computeGeometry)
}

// Destroy cleans up.
func (m *Menu) Destroy() {
	if m.Destroyed() {
		return
	}
	m.Unpost()
	m.MarkDestroyed()
	window.DestroyWindow(m.Win)
}
