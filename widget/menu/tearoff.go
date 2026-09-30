package menu

import (
	"github.com/msorc/takigo/color"
	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// TearoffWindow is a persistent toplevel window created when a menu is torn off.
// It displays the menu entries permanently and handles click-to-invoke without
// closing the window.
type TearoffWindow struct {
	win         *window.Window
	entries     []MenuEntry
	font        interface{ platform.DrawableFont }
	fontI       font.Font
	app         widget.AppContext
	entryHeight int
	sepHeight   int
	menuWidth   int
	activeIndex int
	bg          *color.ColorRef
	fg          *color.ColorRef
	activeBg    *color.ColorRef
	activeFg    *color.ColorRef
	border      *draw.Border
	borderWidth int
	depth       int
	destroyed   bool
}

// Detach creates and shows a TearoffWindow containing copies of m's entries.
// The tearoff window is positioned near the menu's current location.
func (m *Menu) Detach() {
	if m.Font == nil {
		return
	}

	d := m.Win.Display
	app := m.app

	// Compute geometry using the menu's metrics.
	fm := m.Font.Metrics()
	entryH := fm.Linespace() + 6
	sepH := 6
	maxW := 0
	totalH := 2 * m.BorderWidth

	for _, e := range m.entries {
		if e.Type == Separator {
			totalH += sepH
		} else {
			totalH += entryH
			w := m.Font.MeasureString(e.Label) + 40
			if e.AccelStr != "" {
				w += m.Font.MeasureString(e.AccelStr) + 20
			}
			if w > maxW {
				maxW = w
			}
		}
	}
	menuW := max(maxW+2*m.BorderWidth, 60)

	// Position near the menu's current on-screen location.
	posX := max(m.Win.X, 0)
	posY := max(m.Win.Y, 0)

	attrs := &platform.WindowAttrs{
		BackgroundPixel: d.WhitePixel,
		BorderPixel:     d.BlackPixel,
		EventMask: int64(
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
		posX, posY, uint(menuW), uint(totalH), 1,
		d.Depth, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWEventMask,
		attrs,
	)

	w := &window.Window{
		PlatformID:      xwin,
		Display:         d,
		Parent:          m.Win.Parent,
		Name:            m.Win.Name + "_tearoff",
		PathName:        m.Win.PathName + "_tearoff",
		Width:           menuW,
		Height:          totalH,
		ReqWidth:        menuW,
		ReqHeight:       totalH,
		Depth:           d.Depth,
		BackgroundPixel: d.WhitePixel,
		Flags:           window.FlagTopLevel,
	}
	w.GC = d.Server.CreateGC(w.Drawable(), platform.GCForeground|platform.GCBackground, &platform.GCValues{
		Foreground: d.BlackPixel,
		Background: d.WhitePixel,
	})
	d.RegisterWindow(xwin, w)
	m.Win.Parent.AddChild(w)

	// Set window title.
	d.Server.StoreName(xwin, "Menu")

	tw := &TearoffWindow{
		win:         w,
		entries:     append([]MenuEntry(nil), m.entries...),
		app:         app,
		entryHeight: entryH,
		sepHeight:   sepH,
		menuWidth:   menuW,
		activeIndex: -1,
		bg:          nil,
		activeBg:    m.ActiveBg,
		activeFg:    m.ActiveFg,
		border:      m.Border,
		borderWidth: m.BorderWidth,
		depth:       d.Depth,
	}
	if m.Background != nil {
		tw.bg = m.Background.Ref()
		w.SetBackgroundPixel(m.Background.Pixel)
	}
	if m.Foreground != nil {
		tw.fg = m.Foreground.Ref()
	}
	if df, ok := m.Font.(platform.DrawableFont); ok {
		tw.font = df
	}
	tw.fontI = m.Font

	disp := app.Dispatcher()

	disp.Bind(xwin, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		tw.display()
	})

	disp.Bind(xwin, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			w.Width = ev.ConfigWidth
			w.Height = ev.ConfigHeight
		} else if ev.Type == event.DestroyType {
			tw.destroyed = true
		}
	})

	disp.Bind(xwin, event.MotionMask, func(ev *event.Event) {
		idx := tw.entryAtY(ev.Y)
		if idx != tw.activeIndex {
			tw.activeIndex = idx
			tw.display()
		}
	})

	disp.Bind(xwin, event.EnterMask|event.LeaveMask, func(ev *event.Event) {
		if ev.Type == event.LeaveType {
			tw.activeIndex = -1
			tw.display()
		}
	})

	disp.Bind(xwin, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		idx := tw.entryAtY(ev.Y)
		if idx < 0 || idx >= len(tw.entries) {
			return
		}
		e := &tw.entries[idx]
		if e.State == widget.StateDisabled {
			return
		}
		switch e.Type {
		case Command:
			if e.Command != nil {
				e.Command()
			}
		case Checkbutton:
			e.Checked = !e.Checked
			tw.display()
			if e.Command != nil {
				e.Command()
			}
		case Radiobutton:
			for j := range tw.entries {
				if tw.entries[j].Type == Radiobutton && j != idx {
					tw.entries[j].Checked = false
				}
			}
			e.Checked = true
			tw.display()
			if e.Command != nil {
				e.Command()
			}
		}
	})

	d.Server.MapWindow(xwin)
	m.Unpost()
}

func (tw *TearoffWindow) entryAtY(y int) int {
	offset := tw.borderWidth
	for i, e := range tw.entries {
		var h int
		if e.Type == Separator {
			h = tw.sepHeight
		} else {
			h = tw.entryHeight
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

func (tw *TearoffWindow) display() {
	if tw.destroyed {
		return
	}
	w := tw.win
	if w.PlatformID == 0 {
		return
	}
	d := w.Display.Server
	gc := w.GC

	// Background.
	if tw.bg != nil {
		d.SetForeground(gc, tw.bg.Pixel)
	} else {
		d.SetForeground(gc, w.BackgroundPixel)
	}
	d.FillRectangle(w.Drawable(), gc, 0, 0, uint(w.Width), uint(w.Height))

	// Border.
	if tw.border != nil && tw.borderWidth > 0 {
		draw.Draw3DRectangle(d, w.Drawable(), gc, tw.border,
			0, 0, w.Width, w.Height, tw.borderWidth, option.ReliefRaised)
	}

	df := tw.font
	if df == nil || tw.fontI == nil {
		return
	}
	fm := tw.fontI.Metrics()
	yPos := tw.borderWidth

	for i, e := range tw.entries {
		if e.Type == Separator {
			sepY := yPos + tw.sepHeight/2
			if tw.border != nil {
				draw.Draw3DRectangle(d, w.Drawable(), gc, tw.border,
					tw.borderWidth+2, sepY-1, w.Width-2*tw.borderWidth-4, 2,
					1, option.ReliefSunken)
			}
			yPos += tw.sepHeight
			continue
		}

		isActive := i == tw.activeIndex && e.State != widget.StateDisabled
		if isActive && tw.activeBg != nil {
			d.SetForeground(gc, tw.activeBg.Pixel)
			d.FillRectangle(w.Drawable(), gc, tw.borderWidth, yPos,
				uint(w.Width-2*tw.borderWidth), uint(tw.entryHeight))
		}

		textX := tw.borderWidth + 20
		textY := yPos + (tw.entryHeight-fm.Linespace())/2 + fm.Ascent

		var fgCol *color.ColorRef
		if e.State == widget.StateDisabled {
			if dfg, err := tw.app.ColorCache().Get(widget.DefDisabledForeground); err == nil {
				fgCol = dfg.Ref()
			}
		} else if isActive && tw.activeFg != nil {
			fgCol = tw.activeFg
		} else if tw.fg != nil {
			fgCol = tw.fg
		} else {
			if col, err := tw.app.ColorCache().Get("#000000"); err == nil {
				fgCol = col.Ref()
			}
		}

		if fgCol != nil {
			if e.Type == Checkbutton && e.Checked {
				df.DrawString(w.Drawable(), tw.borderWidth+4, textY, "\u2713",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			} else if e.Type == Radiobutton && e.Checked {
				df.DrawString(w.Drawable(), tw.borderWidth+4, textY, "\u25cf",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}
			if e.Label != "" {
				df.DrawString(w.Drawable(), textX, textY, e.Label,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)

				// Underline for keyboard mnemonic.
				runes := []rune(e.Label)
				if e.Underline >= 0 && e.Underline < len(runes) {
					prefix := string(runes[:e.Underline])
					ch := string(runes[e.Underline])
					ulX := textX + tw.fontI.MeasureString(prefix)
					ulW := tw.fontI.MeasureString(ch)
					ulY := textY + 2
					d.SetForeground(gc, fgCol.Pixel)
					d.DrawLine(w.Drawable(), gc, ulX, ulY, ulX+ulW, ulY)
				}
			}
			if e.AccelStr != "" {
				accelW := tw.fontI.MeasureString(e.AccelStr)
				accelX := w.Width - tw.borderWidth - accelW - 8
				df.DrawString(w.Drawable(), accelX, textY, e.AccelStr,
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}
			if e.Type == Cascade {
				arrowX := w.Width - tw.borderWidth - 14
				df.DrawString(w.Drawable(), arrowX, textY, "\u25b6",
					fgCol.Pixel, fgCol.Red, fgCol.Green, fgCol.Blue)
			}
		}
		yPos += tw.entryHeight
	}

}
