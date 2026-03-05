// Demo: Setting a window icon via the _NET_WM_ICON X11 property.
// Ported from Tk's windowicons.tcl demo (simplified to EWMH icon).
package main

import (
	"encoding/binary"
	"image"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/window"
)

func main() {
	app := demohelper.Setup("Window Icon Demonstration", 400, 300,
		"This demo sets the window icon using the _NET_WM_ICON\nX11 property. The icon should be visible in the window\nmanager's title bar and taskbar.")
	root := app.Window()

	// Set window icon via _NET_WM_ICON using the Tk feather PNG.
	setWindowIcon(root)

	// Badge buttons (badge is not supported on X11, show info dialog).
	badgeMsg := func() {
		dialog.ShowMessage(app,
			dialog.MsgTitle("Badge"),
			dialog.MsgMessage("Icon badge is not supported on this platform."),
			dialog.MsgType(dialog.MsgInfo),
		)
	}

	// Set icon button (matches Tcl's "Set Window Icon to Globe").
	iconBtn := button.New(app, "seticon",
		button.Text("Set Window Icon to Feather"),
		button.Command(func() { setWindowIcon(root) }),
	)
	pack.Pack(iconBtn, pack.FillOpt(pack.FillX), pack.PadX("3p"))

	badge3Btn := button.New(app, "badge3",
		button.Text("Set Badge to 3"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge3Btn, pack.FillOpt(pack.FillX), pack.PadX("3p"))

	badge11Btn := button.New(app, "badge11",
		button.Text("Set Badge to 11"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge11Btn, pack.FillOpt(pack.FillX), pack.PadX("3p"))

	resetBadgeBtn := button.New(app, "resetbadge",
		button.Text("Reset Badge"),
		button.Command(badgeMsg),
	)
	pack.Pack(resetBadgeBtn, pack.FillOpt(pack.FillX), pack.PadX("3p"))

	app.Run()
}

// demoImagesDir returns the path to demos/images/.
func demoImagesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "images")
}

// setWindowIcon sets the Tk feather icon via _NET_WM_ICON.
// Falls back to a procedurally generated icon if the PNG cannot be loaded.
func setWindowIcon(win *window.Window) {
	imgPath := filepath.Join(demoImagesDir(), "Tk_feather.png")
	photo, err := tkimage.NewPhotoFromFile("feather", imgPath)
	if err != nil {
		setFallbackIcon(win)
		return
	}
	rgba := photo.RGBA()
	setIconFromRGBA(win, rgba)
}

// setIconFromRGBA encodes an RGBA image as _NET_WM_ICON ARGB data.
func setIconFromRGBA(win *window.Window, rgba *image.RGBA) {
	w := rgba.Bounds().Dx()
	h := rgba.Bounds().Dy()
	data := make([]byte, (2+w*h)*4)
	binary.LittleEndian.PutUint32(data[0:4], uint32(w))
	binary.LittleEndian.PutUint32(data[4:8], uint32(h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			offset := (2 + y*w + x) * 4
			r, g, b, a := rgba.At(x+rgba.Bounds().Min.X, y+rgba.Bounds().Min.Y).RGBA()
			// Convert 16-bit → 8-bit and pack as ARGB.
			binary.LittleEndian.PutUint32(data[offset:offset+4],
				uint32(a>>8)<<24|uint32(r>>8)<<16|uint32(g>>8)<<8|uint32(b>>8))
		}
	}
	d := win.Display.Server
	netWmIcon := d.InternAtom("_NET_WM_ICON", false)
	d.ChangeProperty(win.PlatformID, netWmIcon, platform.XA_CARDINAL, 32, platform.PropModeReplace, data, 2+w*h)
}

// setFallbackIcon generates a procedural 16x16 blue "T" icon.
func setFallbackIcon(win *window.Window) {
	d := win.Display.Server
	const size = 16
	data := make([]byte, (2+size*size)*4)
	binary.LittleEndian.PutUint32(data[0:4], size)
	binary.LittleEndian.PutUint32(data[4:8], size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			offset := (2 + y*size + x) * 4
			a, r, g, b := uint8(0xFF), uint8(0x4a), uint8(0x69), uint8(0x84)
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
