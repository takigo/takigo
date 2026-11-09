package systray

import (
	"fmt"
	"time"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
)

const (
	systemTrayRequestDock = 0
	trayIconSize          = 24
)

// dock creates the icon window and asks the X11 tray manager to dock it
// (DockToManager). Like tktray without an ARGB visual, the window takes
// its parent's background so the image blends with the tray.
func (t *TrayIcon) dock() error {
	d := t.display.Server
	trayAtom := d.InternAtom(fmt.Sprintf("_NET_SYSTEM_TRAY_S%d", t.display.Screen), false)
	manager := d.GetSelectionOwner(trayAtom)
	if manager == 0 {
		return fmt.Errorf("systray: no system tray manager found")
	}

	size := uint(trayIconSize)
	if t.image != nil {
		size = uint(max(t.image.Width(), t.image.Height(), 1))
	}
	t.win = d.CreateWindow(t.display.RootWindow, 0, 0, size, size, 0,
		platform.CopyFromParent, platform.InputOutput,
		platform.CWBackPixmap|platform.CWEventMask,
		&platform.WindowAttrs{
			ParentRelative: true,
			EventMask: int64(platform.ButtonPressMask | platform.ButtonReleaseMask |
				platform.ExposureMask | platform.StructureNotifyMask |
				platform.EnterWindowMask | platform.LeaveWindowMask),
		})

	// SYSTEM_TRAY_REQUEST_DOCK goes to the manager itself, with the icon
	// in data.l[2].
	opcodeAtom := d.InternAtom("_NET_SYSTEM_TRAY_OPCODE", false)
	d.SendClientMessage(manager, manager, opcodeAtom, int64(platform.CurrentTime),
		systemTrayRequestDock, int64(t.win), 0, 0)
	d.Flush()

	disp := t.app.Dispatcher()
	disp.Bind(t.win, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount == 0 {
			t.draw()
		}
	})
	disp.Bind(t.win, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			t.width, t.height = ev.ConfigWidth, ev.ConfigHeight
			t.draw()
		}
	})
	// systray.tcl's _balloon: show the -text 100 ms after the pointer
	// comes in, take it down when it leaves.
	disp.Bind(t.win, event.EnterMask|event.LeaveMask, func(ev *event.Event) {
		switch ev.Type {
		case event.EnterType:
			x, y := ev.RootX, ev.RootY
			t.balloon.close()
			t.balloon = &balloon{}
			b := t.balloon
			b.cancel = t.app.After(100*time.Millisecond, func() { b.show(t, x, y) })
		case event.LeaveType:
			t.balloon.close()
			t.balloon = nil
		}
	})
	t.width, t.height = int(size), int(size)
	return nil
}

// draw ports DisplayIcon's non-ARGB path: clear to the parent's background
// and draw the image centred, cropped to the window.
func (t *TrayIcon) draw() {
	if t.win == 0 || t.native {
		return
	}
	d := t.display.Server
	d.ClearWindow(t.win)
	if t.image == nil {
		d.Flush()
		return
	}
	w, h := t.image.Width(), t.image.Height()
	imgx, outx, outw := 0, (t.width-w)/2, w
	if t.width < w {
		imgx, outx, outw = (w-t.width)/2, 0, t.width
	}
	imgy, outy, outh := 0, (t.height-h)/2, h
	if t.height < h {
		imgy, outy, outh = (h-t.height)/2, 0, t.height
	}
	drawable := platform.WindowDrawable(t.win)
	gc := d.CreateGC(drawable, 0, &platform.GCValues{})
	defer d.FreeGC(gc)
	t.image.Draw(d, drawable, gc, t.display.Depth, imgx, imgy, outw, outh, outx, outy, 0)
	d.Flush()
}

// balloon is the tooltip window of systray.tcl's _balloon_show: black
// border, a message in TkDefaultFont, below the pointer.
type balloon struct {
	cancel func() bool
	win    platform.WindowID
	t      *TrayIcon
}

func (b *balloon) close() {
	if b == nil {
		return
	}
	if b.cancel != nil {
		b.cancel()
	}
	if b.win != 0 {
		b.t.app.Dispatcher().Unbind(b.win)
		b.t.display.Server.DestroyWindow(b.win)
		b.t.display.Server.Flush()
		b.win = 0
	}
}

func (b *balloon) show(t *TrayIcon, px, py int) {
	if t.tooltip == "" || t.win == 0 {
		return
	}
	f, err := t.app.FontRegistry().Get(font.TkDefaultFont)
	df, ok := f.(platform.DrawableFont)
	if err != nil || !ok {
		return
	}
	d := t.display.Server
	m := f.Metrics()
	pad := max(f.MeasureString("0")/2, 1) // message's -padx/-pady -1: half the average width
	w, h := f.MeasureString(t.tooltip)+2*pad, m.Linespace()+2*pad
	sw, sh := d.ScreenWidth(t.display.Screen), d.ScreenHeight(t.display.Screen)
	x, y := px, py+20
	if y < sh && y+h+2 > sh {
		y = py - h - 2 - 5
	}
	if x < 0 && x+w+2 > 0 {
		x = 0
	} else if x < sw && x+w+2 > sw {
		x = sw - w - 2
	}
	b.t = t
	b.win = d.CreateWindow(t.display.RootWindow, x, y, uint(w), uint(h), 1,
		platform.CopyFromParent, platform.InputOutput,
		platform.CWBackPixel|platform.CWBorderPixel|platform.CWOverrideRedirect|platform.CWEventMask,
		&platform.WindowAttrs{BackgroundPixel: balloonBg, BorderPixel: 0, OverrideRedirect: true,
			EventMask: int64(platform.ExposureMask)})
	text := t.tooltip
	t.app.Dispatcher().Bind(b.win, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		drawable := platform.WindowDrawable(b.win)
		gc := d.CreateGC(drawable, 0, &platform.GCValues{})
		defer d.FreeGC(gc)
		d.SetForeground(gc, balloonBg)
		d.FillRectangle(drawable, gc, 0, 0, uint(w), uint(h))
		df.DrawString(drawable, pad, pad+m.Ascent, text, 0, 0, 0, 0)
		d.Flush()
	})
	d.MapRaised(b.win)
	d.Flush()
}

// balloonBg is the message widget's default background.
const balloonBg = 0xd9d9d9
