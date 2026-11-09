//go:build linux || freebsd || openbsd || netbsd

package systray_test

import (
	goimage "image"
	gocolor "image/color"
	"image/draw"
	"testing"
	"time"

	"github.com/takigo/takigo/event"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/systray"
)

// TestDockDrawsImage plays the tray manager: it owns _NET_SYSTEM_TRAY_S<n>,
// takes the dock request, maps the icon and reads back its pixels.
func TestDockDrawsImage(t *testing.T) {
	app := testutil.NewTestApp(t)
	disp := app.Display()
	d := disp.Server

	manager := d.CreateWindow(disp.RootWindow, 0, 0, 1, 1, 0, platform.CopyFromParent, platform.InputOutput,
		platform.CWEventMask, &platform.WindowAttrs{EventMask: 0})
	defer d.DestroyWindow(manager)
	trayAtom := d.InternAtom("_NET_SYSTEM_TRAY_S"+string(rune('0'+disp.Screen)), false)
	d.SetSelectionOwner(trayAtom, manager, platform.CurrentTime)
	opcode := d.InternAtom("_NET_SYSTEM_TRAY_OPCODE", false)
	var docked platform.WindowID
	app.Dispatcher().Bind(manager, event.ClientMessageMask, func(ev *event.Event) {
		if ev.Type == event.ClientMessageType && ev.MessageType == opcode && ev.MessageData[1] == 0 {
			docked = platform.WindowID(ev.MessageData[2])
		}
	})

	rgba := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
	draw.Draw(rgba, rgba.Bounds(), &goimage.Uniform{gocolor.RGBA{0, 0, 255, 255}}, goimage.Point{}, draw.Src)
	icon, err := systray.New(app, disp, systray.TrayImage(tkimage.NewPhoto("blue", rgba)),
		systray.TrayTooltip("tip"))
	if err != nil {
		t.Fatal(err)
	}
	defer icon.Destroy()

	var pix []byte
	var balloonShown, balloonGone bool
	app.After(100*time.Millisecond, func() {
		if docked == 0 {
			app.Quit()
			return
		}
		// What a tray manager does: give the icon a size and show it.
		d.MoveResizeWindow(docked, 10, 10, 24, 24)
		d.MapWindow(docked)
		d.Flush()
	})
	app.After(300*time.Millisecond, func() {
		if docked != 0 {
			pix = d.GetImageRGBA(platform.WindowDrawable(docked), 0, 0, 24, 24)
		}
		// The pointer comes in: the balloon shows 20 pixels below it.
		app.Dispatcher().Dispatch(&event.Event{Type: event.EnterType, Window: docked, RootX: 600, RootY: 600})
	})
	app.After(500*time.Millisecond, func() {
		balloonShown = d.ChildAt(disp.RootWindow, 605, 625) != 0
		app.Dispatcher().Dispatch(&event.Event{Type: event.LeaveType, Window: docked})
		d.Sync(false)
	})
	app.After(600*time.Millisecond, func() {
		defer app.Quit()
		balloonGone = d.ChildAt(disp.RootWindow, 605, 625) == 0
	})
	app.MainLoop()

	if docked == 0 {
		t.Fatal("no SYSTEM_TRAY_REQUEST_DOCK reached the manager")
	}
	if pix == nil {
		t.Fatal("could not read the icon back")
	}
	blue := func(x, y int) bool {
		i := 4 * (y*24 + x)
		return pix[i] < 40 && pix[i+1] < 40 && pix[i+2] > 200
	}
	// The 16x16 image is centred in the 24x24 window: 4..19.
	if !blue(4, 4) || !blue(19, 19) || !blue(12, 12) {
		t.Error("the image is not drawn centred in the icon")
	}
	if blue(2, 2) || blue(21, 21) {
		t.Error("the image covers the icon's margin")
	}

	if !balloonShown || !balloonGone {
		t.Errorf("balloon: shown %v, gone after Leave %v", balloonShown, balloonGone)
	}

	icon.Configure(systray.TrayTooltip("changed"))
	if !icon.Exists() {
		t.Error("Configure destroyed the icon")
	}
}
