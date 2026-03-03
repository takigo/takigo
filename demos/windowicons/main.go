// Demo: Setting a window icon via the _NET_WM_ICON X11 property.
// Ported from Tk's windowicons.tcl demo (simplified to EWMH icon).
package main

import (
	"encoding/binary"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/window"
)

func main() {
	app := demohelper.Setup("Window Icon Demonstration", 400, 300,
		"This demo sets the window icon using the _NET_WM_ICON\nX11 property. The icon should be visible in the window\nmanager's title bar and taskbar.")
	root := app.Window()

	// Set window icon via _NET_WM_ICON.
	setWindowIcon(root)

	// Badge buttons (badge is not supported on X11, show info dialog).
	badgeMsg := func() {
		dialog.ShowMessage(app,
			dialog.MsgTitle("Badge"),
			dialog.MsgMessage("Icon badge is not supported on this platform."),
			dialog.MsgType(dialog.MsgInfo),
		)
	}

	badge3Btn := button.New(app, "badge3",
		button.Text("Set Badge to 3"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge3Btn, pack.FillOpt(pack.FillX), pack.PadX(3), pack.PadY(2))

	badge11Btn := button.New(app, "badge11",
		button.Text("Set Badge to 11"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge11Btn, pack.FillOpt(pack.FillX), pack.PadX(3), pack.PadY(2))

	resetBadgeBtn := button.New(app, "resetbadge",
		button.Text("Reset Badge"),
		button.Command(badgeMsg),
	)
	pack.Pack(resetBadgeBtn, pack.FillOpt(pack.FillX), pack.PadX(3), pack.PadY(2))

	app.Run()
}

// setWindowIcon sets a 16x16 icon on the window via _NET_WM_ICON.
// The format is: [width, height, ARGB pixels...] as 32-bit values.
func setWindowIcon(win *window.Window) {
	d := win.Display.Server
	const size = 16

	// Generate a simple icon: blue square with white "T" letter.
	data := make([]byte, (2+size*size)*4)

	binary.LittleEndian.PutUint32(data[0:4], size)
	binary.LittleEndian.PutUint32(data[4:8], size)

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			offset := (2 + y*size + x) * 4
			a, r, g, b := uint8(0xFF), uint8(0x4a), uint8(0x69), uint8(0x84)

			// Draw "T" in white.
			topBar := y >= 3 && y <= 4 && x >= 3 && x <= 12
			stem := y >= 4 && y <= 12 && x >= 7 && x <= 8
			if topBar || stem {
				r, g, b = 0xFF, 0xFF, 0xFF
			}

			binary.LittleEndian.PutUint32(data[offset:offset+4],
				uint32(a)<<24|uint32(r)<<16|uint32(g)<<8|uint32(b))
		}
	}

	netWmIcon := d.InternAtom("_NET_WM_ICON", false)
	d.ChangeProperty(win.PlatformID, netWmIcon, platform.XA_CARDINAL, 32, platform.PropModeReplace, data, 2+size*size)
}
