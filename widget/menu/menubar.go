package menu

import (
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/draw"
	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/window"
)

// Menubar is a toplevel's "-menu": a menu of type menubar
// (TkpComputeMenubarGeometry in tk/unix/tkUnixMenu.c, DisplayMenu in
// tkMenuDraw.c). Tk puts it in the toplevel's wrapper window; here it is a
// child of the toplevel across its top edge, and the toplevel's content
// starts below it.
type Menubar struct {
	widget.Base

	entries           []barEntry
	active            int
	posted            *Menu
	activeBorderWidth int
	maxWidth          int // totalWidth: the widest row
	ActiveBg          *color.ColorRef
	top               *window.Window
}

type barEntry struct {
	label               string
	underline           int
	ulX, ulW            int // underline offset and width, set by layout
	sub                 *Menu
	x, y, width, height int
}

// NewMenubar creates the menubar for top (Tk: "toplevel -menu").
func NewMenubar(top widget.Caregiver, name string) *Menubar {
	app := top.AppContext()
	w := window.NewChildWindow(top.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(w)
	mb := &Menubar{active: -1, activeBorderWidth: 1, top: top.Window()}
	widget.InitBase(&mb.Base, w, app)
	mb.SetDisplayProc(mb.display)
	w.Class = "Menu"
	mb.BorderWidth = 1 // DEF_MENU_BORDER_WIDTH
	mb.Relief = option.ReliefRaised
	if f, err := app.FontRegistry().Get(font.TkMenuFont); err == nil {
		mb.Font = f
	}
	if c, err := app.ColorCache().Get(widget.PaletteFor(app).ActiveBackground); err == nil {
		mb.ActiveBg = c.Ref()
	}
	top.Window().Menubar = w
	bindMenubar(mb, app)
	mb.layout()
	return mb
}

// AddCascade adds a cascade entry ("add cascade -label -underline -menu").
func (mb *Menubar) AddCascade(label string, underline int, sub *Menu) {
	mb.entries = append(mb.entries, barEntry{label: label, underline: underline, sub: sub})
	mb.layout()
}

// geometry ports TkpComputeMenubarGeometry for text cascades: each entry is
// its label plus 2*activeborderwidth+10 each way, wrapped into rows.
func (mb *Menubar) geometry(maxWidth int) int {
	bw, abw := mb.BorderWidth, mb.activeBorderWidth
	if len(mb.entries) == 0 || mb.Font == nil {
		return 0
	}
	if maxWidth <= 1 {
		maxWidth = 0x7FFFFFF
	}
	ls := mb.Font.Metrics().Linespace()
	x, y, rowH, rowStart := bw, bw, 0, 0
	place := func(from, to int) {
		xx := bw
		for j := from; j < to; j++ {
			e := &mb.entries[j]
			e.y = y + rowH - e.height
			e.x = xx
			xx += e.width
		}
	}
	for i := range mb.entries {
		e := &mb.entries[i]
		// GetMenuIndicatorGeometry: a cascade's indicator space is -borderwidth.
		e.width = mb.Font.MeasureString(e.label) + bw + 2*abw + 10
		if runes := []rune(e.label); e.underline >= 0 && e.underline < len(runes) {
			e.ulX = mb.Font.MeasureString(string(runes[:e.underline]))
			e.ulW = mb.Font.MeasureString(string(runes[e.underline]))
		}
		e.height = ls + 1 + 2*abw + 10 // GetMenuLabelGeometry adds 1
		if x+e.width+bw > maxWidth && i > rowStart {
			place(rowStart, i)
			rowStart = i
			y += rowH
			rowH, x = 0, bw
		}
		x += e.width
		rowH = max(rowH, e.height)
	}
	place(rowStart, len(mb.entries))
	// As in Tk, x already includes the last entry, which is added again.
	last := mb.entries[len(mb.entries)-1]
	mb.maxWidth = max(x, last.x+last.width+last.width+bw)
	return y + rowH + bw
}

// layout sizes the menubar to the toplevel's width and pushes the
// toplevel's content below it.
func (mb *Menubar) layout() {
	top := mb.top
	h := mb.geometry(top.Width)
	w := mb.Win
	w.ReqWidth, w.ReqHeight = max(1, mb.maxWidth), max(1, h)
	if h == 0 {
		return
	}
	d := w.Display.Server
	w.X, w.Y, w.Width, w.Height = 0, 0, max(1, top.Width), h
	d.MoveResizeWindow(w.PlatformID, 0, 0, uint(w.Width), uint(h))
	d.MapWindow(w.PlatformID)
	window.MarkMapped(w)
	if top.InternalBorderTop != h {
		top.InternalBorderTop = h
		for _, c := range top.Children {
			if c != w && c.GeomManager != nil {
				c.GeomManager.RequestProc(c)
				break
			}
		}
	}
	mb.Display()
}

func (mb *Menubar) hit(x, y int) int {
	for i, e := range mb.entries {
		if x >= e.x && x < e.x+e.width && y >= e.y && y < e.y+e.height {
			return i
		}
	}
	return -1
}

// Display schedules a redraw at idle time; see widget.Base.EventuallyRedraw.
func (mb *Menubar) Display() {
	mb.EventuallyRedraw()
}

// display ports DisplayMenu/TkpDrawMenuEntry for a menubar.
func (mb *Menubar) display() {
	w := mb.Win
	if mb.Destroyed() || w.PlatformID == 0 || w.Width <= 0 || w.Height <= 0 {
		return
	}
	d := w.Display.Server
	pix := w.Drawable()
	gc := w.GC
	bg := uint64(0xd9d9d9)
	if mb.Background != nil {
		bg = mb.Background.Pixel
	}
	d.SetForeground(gc, bg)
	d.FillRectangle(pix, gc, 0, 0, uint(w.Width), uint(w.Height))
	fg := uint64(0)
	if mb.Foreground != nil {
		fg = mb.Foreground.Pixel
	}
	df, _ := mb.Font.(platform.DrawableFont)
	m := mb.Font.Metrics()
	const padY = 3 // TkpDrawMenuEntry: menubar entries are drawn 3px in
	for i, e := range mb.entries {
		y, h := e.y+padY, e.height-2*padY
		if i == mb.active && mb.ActiveBg != nil {
			// A menubar entry is raised only while its cascade is posted.
			relief := option.ReliefFlat
			if mb.posted != nil && mb.posted == e.sub {
				relief = option.ReliefRaised
			}
			draw.Fill3DRectangle(d, pix, gc, draw.NewBorderFromPixel(mb.ActiveBg.Pixel),
				e.x, y, e.width, h, mb.activeBorderWidth, relief)
		}
		left := e.x + mb.BorderWidth + mb.activeBorderWidth + 5 // + indicatorSpace
		baseline := y + (h+m.Ascent-m.Descent)/2
		if df != nil {
			r, g, b := uint16(fg>>16&0xff)<<8, uint16(fg>>8&0xff)<<8, uint16(fg&0xff)<<8
			df.DrawString(pix, left, baseline, e.label, fg, r, g, b)
		}
		if runes := []rune(e.label); e.underline >= 0 && e.underline < len(runes) {
			ux, uw := left+e.ulX, e.ulW
			pos, uh := font.Underline(mb.Font)
			d.SetForeground(gc, fg)
			d.FillRectangle(pix, gc, ux, baseline+pos, uint(uw), uint(uh))
		}
	}
	if mb.Relief != option.ReliefFlat && mb.Border != nil {
		draw.Draw3DRectangle(d, pix, gc, mb.Border, 0, 0, w.Width, w.Height, mb.BorderWidth, mb.Relief)
	}
}

// post shows entry i's cascade below it (TkPostSubmenu for a menubar) for
// the click being handled.
func (mb *Menubar) post(i int) { mb.postEntry(i, true) }

// postEntry posts entry i's cascade; fromClick says a ButtonPress is being
// dispatched, which the menu's global click handler must then ignore.
func (mb *Menubar) postEntry(i int, fromClick bool) {
	old := mb.posted
	mb.posted = nil
	mb.active = i
	if i >= 0 && mb.entries[i].sub != nil && mb.entries[i].sub != old {
		e := mb.entries[i]
		w := mb.Win
		x, y := w.Display.Server.TranslateCoordinates(w.PlatformID, w.Display.RootWindow, e.x, e.y+e.height)
		mb.posted = e.sub
		sub := e.sub
		sub.onUnpost = func() {
			// Closed from elsewhere (a click outside, Escape, an invoke):
			// the menubar entry goes back to normal.
			if mb.posted == sub {
				mb.posted, mb.active = nil, -1
				mb.Display()
			}
		}
		if fromClick {
			e.sub.PostFromButton(x, y)
		} else {
			e.sub.Post(x, y)
		}
	} else if i >= 0 && mb.entries[i].sub == old {
		mb.posted, old = old, nil
	}
	// Unpost the previous menu only after the new one holds the pointer
	// grab (a client's new grab replaces its old one) and the focus, so
	// unmapping it neither ungrabs nor reverts the focus, which would
	// otherwise come back through the toplevel and unpost the new menu.
	if old != nil {
		if mb.posted != nil {
			old.grabbed = false
		}
		old.Unpost()
	}
	mb.Display()
}

func bindMenubar(mb *Menubar, app widget.AppContext) {
	w := mb.Win
	disp := app.Dispatcher()
	disp.Bind(w.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount == 0 {
			mb.Display()
		}
	})
	topID := disp.Bind(mb.top.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType && mb.top.Width != w.Width {
			mb.layout()
		}
	})
	w.OnDestroy(func() { disp.UnbindID(topID) })
	disp.Bind(w.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		if i := mb.hit(ev.X, ev.Y); i >= 0 {
			if mb.posted != nil && mb.active == i {
				mb.post(-1)
			} else {
				mb.post(i)
			}
		}
	})
	disp.Bind(w.PlatformID, event.MotionMask|event.LeaveMask, func(ev *event.Event) {
		i := -1
		if ev.Type == event.MotionType {
			i = mb.hit(ev.X, ev.Y)
		}
		switch {
		case mb.posted != nil && mb.posted.IsPosted():
			if i >= 0 && i != mb.active {
				mb.postEntry(i, false)
			}
		case i != mb.active:
			mb.posted = nil
			mb.active = i
			mb.Display()
		}
	})
}
