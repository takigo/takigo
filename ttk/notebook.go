package ttk

import (
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// notebookTab holds information about a single notebook tab.
type notebookTab struct {
	Text    string
	Window  *window.Window
	State   State // tab-specific state (disabled, etc.)
	Padding Padding
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
		Text:    text,
		Window:  pane,
		Padding: Padding{Left: 8, Top: 4, Right: 8, Bottom: 4},
	}
	nb.tabs = append(nb.tabs, tab)
	nb.computeTabGeometry()

	// If this is the first tab, select it.
	if nb.selected < 0 {
		nb.Select(0)
	} else {
		// Unmap the new pane (not selected).
		pane.Display.XDisplay.UnmapWindow(pane.XWindow)
		nb.Display()
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
			nb.layoutPane(tab.Window)
			tab.Window.Display.XDisplay.MapWindow(tab.Window.XWindow)
		} else {
			tab.Window.Display.XDisplay.UnmapWindow(tab.Window.XWindow)
		}
		_ = old
	}

	nb.Display()
}

// TabCount returns the number of tabs.
func (nb *Notebook) TabCount() int {
	return len(nb.tabs)
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
func (nb *Notebook) layoutPane(pane *window.Window) {
	win := nb.Win
	bw := 2 // content border width
	x := bw
	y := nb.tabHeight + bw
	w := win.Width - 2*bw
	h := win.Height - nb.tabHeight - 2*bw
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	pane.Display.XDisplay.MoveResizeWindow(pane.XWindow, x, y, uint(w), uint(h))
	pane.Width = w
	pane.Height = h
}

// Display draws the notebook.
func (nb *Notebook) Display() {
	if nb.Destroyed {
		return
	}
	win := nb.Win
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

	// Allocate or resize pixmap.
	if nb.pixmap == xlib.Pixmap(0) || nb.pixmapW != width || nb.pixmapH != height {
		if nb.pixmap != xlib.Pixmap(0) {
			d.FreePixmap(nb.pixmap)
		}
		nb.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		nb.pixmapW = width
		nb.pixmapH = height
	}

	pixDrawable := xlib.PixmapDrawable(nb.pixmap)

	bg := LookupColor(nb.Context.Style, "-background", nb.State, 0xd9d9d9)
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	border := draw.NewBorderFromPixel(bg)

	// Draw tabs.
	tabX := 2
	for i, tab := range nb.tabs {
		tw := nb.tabWidths[i]
		th := nb.tabHeight
		isSelected := i == nb.selected
		isHover := i == nb.hoverTab && !isSelected

		// Tab background.
		tabBg := bg
		if isSelected {
			tabBg = bg
		} else if isHover {
			tabBg = LookupColor(nb.Context.Style, "-background", StateHover, 0xececec)
		} else {
			tabBg = LookupColor(nb.Context.Style, "-background", StateBackground, 0xd0d0d0)
		}

		tabY := 0
		tabH := th
		if !isSelected {
			tabY = 2 // non-selected tabs are shorter
			tabH -= 2
		}

		// Fill tab background.
		d.SetForeground(gc, tabBg)
		d.FillRectangle(pixDrawable, gc, tabX, tabY, uint(tw), uint(tabH))

		// Tab border: top, left, right.
		d.SetForeground(gc, border.LightPixel)
		d.DrawLine(pixDrawable, gc, tabX, tabY, tabX+tw-1, tabY)         // top
		d.DrawLine(pixDrawable, gc, tabX, tabY, tabX, tabY+tabH-1)       // left
		d.SetForeground(gc, border.DarkPixel)
		d.DrawLine(pixDrawable, gc, tabX+tw-1, tabY, tabX+tw-1, tabY+tabH-1) // right

		// Tab text.
		if nb.Font != nil && tab.Text != "" {
			fgPixel := LookupColor(nb.Context.Style, "-foreground", nb.State, 0x000000)
			textW := nb.Font.MeasureString(tab.Text)
			textX := tabX + (tw-textW)/2
			m := nb.Font.Metrics()
			textY := tabY + (tabH-m.Linespace())/2 + m.Ascent

			if xftFont, ok := nb.Font.(*font.XftFont); ok {
				r := uint16((fgPixel >> 16) & 0xFF) << 8
				g := uint16((fgPixel >> 8) & 0xFF) << 8
				b := uint16((fgPixel) & 0xFF) << 8
				xftFont.DrawString(pixDrawable, textX, textY, tab.Text, fgPixel, r, g, b)
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
		nb.layoutPane(nb.tabs[nb.selected].Window)
	}
}

// hitTestTab returns the tab index at the given position, or -1.
func (nb *Notebook) hitTestTab(x, y int) int {
	if y >= nb.tabHeight {
		return -1
	}
	tabX := 2
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
	app.Dispatcher().Bind(win.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		nb.Display()
	})

	// Configure (resize).
	app.Dispatcher().Bind(win.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			nb.Display()
		}
	})

	// Button1 on tab → select.
	app.Dispatcher().Bind(win.XWindow, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button == 1 {
			idx := nb.hitTestTab(ev.X, ev.Y)
			if idx >= 0 && nb.tabs[idx].State&StateDisabled == 0 {
				nb.Select(idx)
			}
		}
	})

	// Motion for tab hover.
	app.Dispatcher().Bind(win.XWindow, event.MotionMask, func(ev *event.Event) {
		idx := nb.hitTestTab(ev.X, ev.Y)
		if idx != nb.hoverTab {
			nb.hoverTab = idx
			nb.Display()
		}
	})

	// Leave → clear hover.
	app.Dispatcher().Bind(win.XWindow, event.LeaveMask, func(ev *event.Event) {
		if nb.hoverTab >= 0 {
			nb.hoverTab = -1
			nb.Display()
		}
	})
}
