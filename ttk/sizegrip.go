package ttk

import (
	"fmt"

	"github.com/msorc/takigo/draw"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
	"github.com/msorc/takigo/wm"
)

// Sizegrip is a themed grip handle for resizing a toplevel window.
type Sizegrip struct {
	TtkWidget

	// Drag state.
	pressed bool
	pressX  int
	pressY  int
	startW  int
	startH  int
}

// SizegripOption configures a Sizegrip.
type SizegripOption func(*Sizegrip)

// NewSizegrip creates a themed sizegrip widget.
func NewSizegrip(parent widget.Caregiver, name string, opts ...SizegripOption) *Sizegrip {
	app := parent.AppContext()
	win := window.NewChildWindow(parent.Window(), name, 0, 0, 1, 1)
	window.MakeWindowExist(win)

	sg := &Sizegrip{}

	for _, opt := range opts {
		opt(sg)
	}

	InitTtkWidget(&sg.TtkWidget, win, app, "TSizegrip")

	// Set size to match grip element — Tk default is "11.25p" (11.25 points).
	gripSize := screenunit.Px("11.25p")
	win.ReqWidth = gripSize
	win.ReqHeight = gripSize

	win.SetCursor(14) // XC_bottom_right_corner (resize cursor)
	bindSizegrip(sg, app)

	return sg
}

// Display renders the sizegrip with diagonal grip lines.
func (sg *Sizegrip) Display() {
	if sg.Destroyed {
		return
	}
	win := sg.Win
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

	// Double buffer.
	if sg.pixmap == 0 || sg.pixmapW != width || sg.pixmapH != height {
		if sg.pixmap != 0 {
			d.FreePixmap(sg.pixmap)
		}
		sg.pixmap = d.CreatePixmap(win.Drawable(), uint(width), uint(height), uint(win.Depth))
		sg.pixmapW = width
		sg.pixmapH = height
	}

	pixDrawable := platform.PixmapDrawable(sg.pixmap)

	bg := LookupColor(sg.Context.Style, "-background", sg.State, 0xd9d9d9)

	// Fill background.
	d.SetForeground(gc, bg)
	d.FillRectangle(pixDrawable, gc, 0, 0, uint(width), uint(height))

	// Draw grip lines using Tk's algorithm.
	gripSize := height
	if width < gripSize {
		gripSize = width
	}

	gripCount := 3
	gripThickness := gripSize * 3 / (gripCount * 5)
	gripSpace := gripSize/3 - gripThickness

	border := draw.NewBorderFromPixel(bg)

	x1 := width - 1
	y1 := height - 1
	x2 := x1
	y2 := y1

	for g := 0; g < gripCount; g++ {
		x1 -= gripSpace
		y2 -= gripSpace
		for i := 1; i < gripThickness; i++ {
			d.SetForeground(gc, border.DarkPixel)
			d.DrawLine(pixDrawable, gc, x1, y1, x2, y2)
			x1--
			y2--
		}
		d.SetForeground(gc, border.LightPixel)
		d.DrawLine(pixDrawable, gc, x1, y1, x2, y2)
		x1--
		y2--
	}

	// Copy to window.
	d.CopyArea(pixDrawable, win.Drawable(), gc, 0, 0, uint(width), uint(height), 0, 0)
	d.Flush()
}

func bindSizegrip(sg *Sizegrip, app widget.AppContext) {
	win := sg.Win

	app.Dispatcher().Bind(win.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		sg.Display()
	})

	app.Dispatcher().Bind(win.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			win.Width = ev.ConfigWidth
			win.Height = ev.ConfigHeight
			sg.Display()
		}
	})

	app.Dispatcher().Bind(win.PlatformID, event.EnterMask, func(ev *event.Event) {
		sg.ChangeState(StateHover|StateActive, 0)
	})

	app.Dispatcher().Bind(win.PlatformID, event.LeaveMask, func(ev *event.Event) {
		sg.ChangeState(0, StateHover|StateActive)
	})

	// Button-1: start resize.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonPressMask, func(ev *event.Event) {
		if ev.Button != 1 {
			return
		}
		top := window.Toplevel(win)
		if top == nil {
			return
		}

		sg.pressed = true
		// Use root coordinates for drag tracking.
		sg.pressX = ev.RootX
		sg.pressY = ev.RootY
		sg.startW = top.Width
		sg.startH = top.Height
	})

	// B1-Motion: resize toplevel.
	app.Dispatcher().Bind(win.PlatformID, event.MotionMask, func(ev *event.Event) {
		if !sg.pressed || ev.State&platform.Button1Mask == 0 {
			return
		}
		top := window.Toplevel(win)
		if top == nil {
			return
		}

		dx := ev.RootX - sg.pressX
		dy := ev.RootY - sg.pressY
		newW := sg.startW + dx
		newH := sg.startH + dy
		if newW < 50 {
			newW = 50
		}
		if newH < 50 {
			newH = 50
		}

		if info, ok := top.WmData.(*wm.WmInfo); ok {
			info.SetGeometry(formatGeom(newW, newH, top.X, top.Y))
		}
	})

	// ButtonRelease-1: stop resize.
	app.Dispatcher().Bind(win.PlatformID, event.ButtonReleaseMask, func(ev *event.Event) {
		if ev.Button == 1 {
			sg.pressed = false
		}
	})
}

func formatGeom(w, h, x, y int) string {
	return fmt.Sprintf("%dx%d+%d+%d", w, h, x, y)
}
