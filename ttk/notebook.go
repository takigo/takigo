package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
	"math"
)

// notebookTab holds information about a single notebook tab.
type notebookTab struct {
	Text      string
	Window    *window.Window
	State     State   // tab-specific state (disabled, etc.)
	Padding   Padding // tab edge to label: tab border, -padding, focus ring
	PanePad   Padding // the tab's -padding option: space around the pane
	Underline int     // index of character to underline for keyboard shortcut; -1 = none
}

// Notebook is a themed tabbed container widget.
type Notebook struct {
	TtkWidget
	tabs     []notebookTab
	selected int
	hoverTab int
	Font     font.Font

	// Tab bar geometry.
	tabHeight int
	tabWidths []int
}

// NotebookOption configures a Notebook.
type NotebookOption func(*Notebook)

// NewNotebook creates a themed notebook widget.
func NewNotebook(parent widget.Caregiver, name string, opts ...NotebookOption) *Notebook {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 300, 200)
	window.MakeWindowExist(win)

	nb := &Notebook{
		selected: -1,
		hoverTab: -1,
	}

	nb.Font, _ = app.FontRegistry().Get(font.TkDefaultFont)

	InitTtkWidget(&nb.TtkWidget, win, app, "TNotebook")
	nb.DisplayFunc = nb.Display

	for _, opt := range opts {
		opt(nb)
	}

	// Bind notebook-specific events.
	bindNotebook(nb, app)

	return nb
}

// Add appends a tab to the notebook.
func (nb *Notebook) Add(pane *window.Window, text string) {
	tab := notebookTab{
		Text:   text,
		Window: pane,
		// Default theme: TabElement border {1 1 1 0}, TNotebook.Tab
		// -padding {3p 1.5p}, and a 1px focus ring around the label.
		Padding:   Padding{Left: 1 + 4 + 1, Top: 1 + 2 + 1, Right: 1 + 4 + 1, Bottom: 0 + 2 + 1},
		Underline: -1,
	}
	nb.tabs = append(nb.tabs, tab)
	nb.computeTabGeometry()
	nb.updateReqSize()

	// If this is the first tab, select it.
	if nb.selected < 0 {
		nb.Select(0)
	} else {
		// Unmap the new pane (not selected).
		pane.Display.Server.UnmapWindow(pane.PlatformID)
		pane.Flags &^= window.FlagMapped
		nb.Display()
	}
}

// updateReqSize computes the notebook's requested size from the maximum
// pane content size plus the tab bar height and content border.
// This mirrors Tk's NotebookSize in ttkNotebook.c.
func (nb *Notebook) updateReqSize() {
	bw := nbClientBorder
	maxW, maxH := 0, 0
	for _, tab := range nb.tabs {
		pw := tab.Window.ReqWidth + tab.PanePad.Width()
		ph := tab.Window.ReqHeight + tab.PanePad.Height()
		if pw > maxW {
			maxW = pw
		}
		if ph > maxH {
			maxH = ph
		}
	}
	reqW := maxW + 2*bw
	reqH := maxH + nb.tabHeight + 2*bw
	if reqW != nb.Win.ReqWidth || reqH != nb.Win.ReqHeight {
		nb.Win.ReqWidth = reqW
		nb.Win.ReqHeight = reqH
		// Notify the parent geometry manager so it can re-layout.
		if nb.Win.GeomManager != nil {
			nb.Win.GeomManager.RequestProc(nb.Win)
		}
	}
}

// Select makes the tab at index visible.
func (nb *Notebook) Select(index int) {
	if index < 0 || index >= len(nb.tabs) {
		return
	}
	old := nb.selected
	nb.selected = index

	// Map selected pane, unmap others.
	for i, tab := range nb.tabs {
		if i == index {
			nb.layoutPane(tab)
			tab.Window.Display.Server.MapWindow(tab.Window.PlatformID)
			window.MarkMapped(tab.Window)
		} else {
			tab.Window.Display.Server.UnmapWindow(tab.Window.PlatformID)
			tab.Window.Flags &^= window.FlagMapped
		}
		_ = old
	}

	nb.Display()
}

// nbClientBorder is the default theme's Notebook.client border width.
const nbClientBorder = 1

// SetPanePadding sets a tab's -padding (a Tk padding spec such as "1.5p"):
// extra space between the notebook's client area and the pane.
func (nb *Notebook) SetPanePadding(index int, spec string) {
	if index < 0 || index >= len(nb.tabs) {
		return
	}
	nb.tabs[index].PanePad = ParsePadding(spec)
	nb.updateReqSize()
	if index == nb.selected {
		nb.layoutPane(nb.tabs[index])
	}
	nb.Display()
}

// TabCount returns the number of tabs.
func (nb *Notebook) TabCount() int {
	return len(nb.tabs)
}

// SetTabUnderline sets the character index to underline in the tab label at the
// given notebook tab index. Pass -1 to clear the underline. The underlined
// character acts as an Alt+letter keyboard shortcut to select the tab.
func (nb *Notebook) SetTabUnderline(tabIndex, charIndex int) {
	if tabIndex < 0 || tabIndex >= len(nb.tabs) {
		return
	}
	nb.tabs[tabIndex].Underline = charIndex
	nb.Display()
}

// SetTabState sets the state flags on the tab at the given index.
// Use StateDisabled to prevent tab selection.
func (nb *Notebook) SetTabState(index int, state State) {
	if index < 0 || index >= len(nb.tabs) {
		return
	}
	nb.tabs[index].State = state
	nb.Display()
}

// Selected returns the index of the selected tab.
func (nb *Notebook) Selected() int {
	return nb.selected
}

func (nb *Notebook) computeTabGeometry() {
	nb.tabWidths = make([]int, len(nb.tabs))
	maxH := 0

	for i, tab := range nb.tabs {
		tw := 0
		th := 0
		if nb.Font != nil && tab.Text != "" {
			tw = nb.Font.MeasureString(tab.Text)
			th = nb.Font.Metrics().Linespace()
		}
		nb.tabWidths[i] = tw + tab.Padding.Width()
		h := th + tab.Padding.Height()
		if h > maxH {
			maxH = h
		}
	}
	nb.tabHeight = maxH
}

// layoutPane positions the pane window in the content area.
func (nb *Notebook) layoutPane(tab notebookTab) {
	pane := tab.Window
	win := nb.Win
	bw := nbClientBorder
	x := bw + tab.PanePad.Left
	y := nb.tabHeight + bw + tab.PanePad.Top
	w := win.Width - 2*bw - tab.PanePad.Width()
	h := win.Height - nb.tabHeight - 2*bw - tab.PanePad.Height()
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	pane.Display.Server.MoveResizeWindow(pane.PlatformID, x, y, uint(w), uint(h))
	pane.X, pane.Y = x, y
	pane.Width = w
	pane.Height = h
}

// Display draws the notebook.
func (nb *Notebook) Display() {
	if nb.Destroyed {
		return
	}
	win := nb.Win
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

	// Allocate or resize pixmap.
	if nb.pixmap == 0 || nb.pixmapW != width || nb.pixmapH != height {
		if nb.pixmap != 0 {
			d.FreePixmap(nb.pixmap)
		}
		nb.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		nb.pixmapW = width
		nb.pixmapH = height
	}
	if nb.pixmap == 0 {
		return
	}

	pixDrawable := platform.PixmapDrawable(nb.pixmap)

	bg := LookupColor(nb.Context.Style, "-background", 0, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	border := draw.NewBorderFromPixel(bg)

	// Draw tabs.
	tabX := 0 // default theme: no -tabmargins, no -expand (ttkNotebook.c)
	for i, tab := range nb.tabs {
		tw := nb.tabWidths[i]
		th := nb.tabHeight
		isSelected := i == nb.selected
		isHover := i == nb.hoverTab && !isSelected

		// TNotebook.Tab -background: -darker, the frame colour when selected.
		tabBg := uint64(0xc3c3c3)
		if isSelected {
			tabBg = bg
		}
		_ = isHover

		tabY, tabH := 0, th

		// TabElementDraw (ttkElements.c), tabs on the north side: a polygon
		// with cut corners, a selected tab reaching 1px over the client
		// border, 3D edges, and the -highlightcolor stripe when selected.
		b := Box{tabX, tabY, tw, tabH}
		if isSelected {
			b.Height++
		}
		cut := int(math.Round(2 * float64(screenunit.ScalingPct()) / 100))
		pts := []platform.Point{
			{X: int16(b.X), Y: int16(b.Y + b.Height - 1)},
			{X: int16(b.X), Y: int16(b.Y + cut)},
			{X: int16(b.X + cut), Y: int16(b.Y)},
			{X: int16(b.X + b.Width - 1 - cut), Y: int16(b.Y)},
			{X: int16(b.X + b.Width - 1), Y: int16(b.Y + cut)},
			{X: int16(b.X + b.Width - 1), Y: int16(b.Y + b.Height)},
		}
		tb := draw.NewBorderFromPixel(tabBg)
		d.SetForeground(gc, tabBg)
		d.FillPolygon(pixDrawable, gc, pts, 2, 0)
		pts[5].Y--
		d.SetForeground(gc, tb.LightPixel)
		d.DrawLines(pixDrawable, gc, pts[:4], 0)
		d.SetForeground(gc, tb.DarkPixel)
		d.DrawLines(pixDrawable, gc, pts[3:], 0)
		if isSelected && b.Width >= 2*cut {
			d.SetForeground(gc, LookupColor(nb.Context.Style, "-selectbackground", 0, 0x4a6984))
			d.FillRectangle(pixDrawable, gc, b.X+cut, b.Y, uint(b.Width-2*cut), uint(cut))
		}

		// Tab text.
		if nb.Font != nil && tab.Text != "" {
			tabState := nb.State | tab.State
			fgPixel := LookupColor(nb.Context.Style, "-foreground", tabState, 0x000000)
			textW := nb.Font.MeasureString(tab.Text)
			textX := tabX + (tw-textW)/2
			m := nb.Font.Metrics()
			// Notebook.label is packed top inside the tab's border, padding
			// and focus, at the same place whether or not the tab is selected.
			textY := tab.Padding.Top + m.Ascent

			if df, ok := nb.Font.(platform.DrawableFont); ok {
				r := uint16((fgPixel>>16)&0xFF) * 257
				g := uint16((fgPixel>>8)&0xFF) * 257
				b := uint16((fgPixel)&0xFF) * 257
				df.DrawString(pixDrawable, textX, textY, tab.Text, fgPixel, r, g, b)
			}

			// Underline a specific character for Alt+letter keyboard shortcut.
			runes := []rune(tab.Text)
			if tab.Underline >= 0 && tab.Underline < len(runes) {
				underX := textX + nb.Font.MeasureString(string(runes[:tab.Underline]))
				underW := nb.Font.MeasureString(string(runes[tab.Underline : tab.Underline+1]))
				pos, h := font.Underline(nb.Font) // Tk_UnderlineTextLayout
				d.SetForeground(gc, fgPixel)
				d.FillRectangle(pixDrawable, gc, underX, textY+pos, uint(underW), uint(h))
			}
		}

		tabX += tw
	}

	// Content area border.
	contentY := nb.tabHeight
	contentH := height - contentY
	if contentH > 0 {
		// Draw raised border around content area.
		d.SetForeground(gc, border.LightPixel)
		// Left.
		d.DrawLine(pixDrawable, gc, 0, contentY, 0, height-1)
		// Top — but skip the selected tab area.
		if nb.selected >= 0 {
			selLeft := 2
			for i := 0; i < nb.selected; i++ {
				selLeft += nb.tabWidths[i]
			}
			selRight := selLeft + nb.tabWidths[nb.selected]
			// Left part of top.
			if selLeft > 0 {
				d.DrawLine(pixDrawable, gc, 0, contentY, selLeft, contentY)
			}
			// Right part of top.
			if selRight < width {
				d.DrawLine(pixDrawable, gc, selRight-1, contentY, width-1, contentY)
			}
		} else {
			d.DrawLine(pixDrawable, gc, 0, contentY, width-1, contentY)
		}

		d.SetForeground(gc, border.DarkPixel)
		// Bottom.
		d.DrawLine(pixDrawable, gc, 0, height-1, width-1, height-1)
		// Right.
		d.DrawLine(pixDrawable, gc, width-1, contentY, width-1, height-1)
	}

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()

	// Re-layout selected pane.
	if nb.selected >= 0 && nb.selected < len(nb.tabs) {
		nb.layoutPane(nb.tabs[nb.selected])
	}
}

// hitTestTab returns the tab index at the given position, or -1.
func (nb *Notebook) hitTestTab(x, y int) int {
	if y >= nb.tabHeight {
		return -1
	}
	tabX := 0
	for i := range nb.tabs {
		tw := nb.tabWidths[i]
		if x >= tabX && x < tabX+tw {
			return i
		}
		tabX += tw
	}
	return -1
}

func bindNotebook(nb *Notebook, app widget.AppContext) {
	win := nb.Win

	// Expose.
	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		nb.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			nb.Display()
		}
	})

	// Button1 on tab → select + take focus (enables Ctrl+Tab traversal).
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			idx := nb.hitTestTab(ev.X, ev.Y)
			if idx >= 0 && nb.tabs[idx].State&StateDisabled == 0 {
				nb.Select(idx)
				win.Display.Server.SetInputFocus(win.PlatformID, platform.RevertToParent, ev.Time)
			}
		}
	})

	// Motion for tab hover (skip disabled tabs).
	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		idx := nb.hitTestTab(ev.X, ev.Y)
		if idx >= 0 && nb.tabs[idx].State&StateDisabled != 0 {
			idx = -1
		}
		if idx != nb.hoverTab {
			nb.hoverTab = idx
			nb.Display()
		}
	})

	// Leave → clear hover.
	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		if nb.hoverTab >= 0 {
			nb.hoverTab = -1
			nb.Display()
		}
	})

	// Ctrl+Tab → next tab; Ctrl+Shift+Tab → previous tab.
	// Alt+letter → select tab with matching underline character.
	// Matches ttk::notebook::enableTraversal behavior.
	app.Dispatcher().Bind(win.PlatformID, event.KeyPressMask, func(ev *event.Event) {
		// Alt+letter shortcut: switch to tab whose underline character matches.
		if ev.State&platform.Mod1Mask != 0 && ev.KeySym >= 'a' && ev.KeySym <= 'z' {
			pressedRune := rune(ev.KeySym)
			for i, tab := range nb.tabs {
				if tab.Underline < 0 || tab.State&StateDisabled != 0 {
					continue
				}
				runes := []rune(tab.Text)
				if tab.Underline < len(runes) {
					tabRune := rune(runes[tab.Underline])
					if tabRune >= 'A' && tabRune <= 'Z' {
						tabRune += 32 // toLower
					}
					if tabRune == pressedRune {
						nb.Select(i)
						return
					}
				}
			}
		}

		if ev.KeySym != platform.XK_Tab {
			return
		}
		if ev.State&platform.ControlMask == 0 {
			return
		}
		n := len(nb.tabs)
		if n <= 1 {
			return
		}
		cur := nb.Selected()
		var next int
		if ev.State&platform.ShiftMask != 0 {
			// Ctrl+Shift+Tab → previous enabled tab.
			next = cur
			for range n {
				next = (next - 1 + n) % n
				if nb.tabs[next].State&StateDisabled == 0 {
					break
				}
			}
		} else {
			// Ctrl+Tab → next enabled tab.
			next = cur
			for range n {
				next = (next + 1) % n
				if nb.tabs[next].State&StateDisabled == 0 {
					break
				}
			}
		}
		if next != cur {
			nb.Select(next)
		}
	})
}
