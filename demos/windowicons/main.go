// Demo: Showcases the wm iconphoto and wm iconbadge commands.
// Ported from Tk's windowicons.tcl demo.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/frame"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Window Icon Demonstration"),
		takigo.Geometry("+300+300"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Load the globe icon (windowicons.tcl's embedded PNG) for both the
	// button image and the window icon.
	iconPhoto := loadIconPhoto()
	setWindowIcon(app)

	// Badge buttons (badge is not supported on X11, show info dialog).
	badgeMsg := func() {
		dialog.ShowMessage(app,
			dialog.MsgTitle("Badge"),
			dialog.MsgMessage("Icon badge is not supported on this platform."),
			dialog.MsgType(dialog.MsgInfo),
		)
	}

	// Set icon button with image (matches Tcl's $w.i).
	var iconBtnOpts []button.ButtonOption
	iconBtnOpts = append(iconBtnOpts,
		button.Text("Set Window Icon to Globe"),
		button.Command(func() { setWindowIcon(app) }),
	)
	if iconPhoto != nil {
		iconBtnOpts = append(iconBtnOpts,
			button.ImageOpt(iconPhoto),
			button.CompoundOpt(widget.CompoundTop),
		)
	}
	iconBtn := button.New(f, "i", iconBtnOpts...)
	pack.Pack(iconBtn, pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(3)))

	badge3Btn := button.New(f, "b",
		button.Text("Set Badge to 3"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge3Btn, pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(3)))

	badge11Btn := button.New(f, "e",
		button.Text("Set Badge to 11"),
		button.Command(badgeMsg),
	)
	pack.Pack(badge11Btn, pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(3)))

	resetBadgeBtn := button.New(f, "f",
		button.Text("Reset Badge"),
		button.Command(badgeMsg),
	)
	pack.Pack(resetBadgeBtn, pack.FillOpt(pack.FillX), pack.PadX(screenunit.Pt(3)))

	app.Run()
}

// demoImagesDir returns the path to demos/images/.
func demoImagesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "images")
}

// loadIconPhoto loads the globe PNG as a Photo for use as a button image.
func loadIconPhoto() *tkimage.Photo {
	imgPath := filepath.Join(demoImagesDir(), "windowicons_globe.png")
	photo, err := tkimage.NewPhotoFromFile("icon", imgPath)
	if err != nil {
		return nil
	}
	return photo
}

// setWindowIcon is "wm iconphoto . icon": the globe, decoded straight to
// non-premultiplied pixels, or a drawn "T" if the PNG cannot be read.
func setWindowIcon(app *takigo.App) {
	icon, err := decodeGlobe()
	if err != nil {
		icon = fallbackIcon()
	}
	app.WmInfo().SetIconPhoto(icon)
}

func decodeGlobe() (image.Image, error) {
	f, err := os.Open(filepath.Join(demoImagesDir(), "windowicons_globe.png"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

// fallbackIcon is a 16x16 white "T" on blue.
func fallbackIcon() image.Image {
	const size = 16
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			c := color.NRGBA{0x4a, 0x69, 0x84, 0xff}
			topBar := y >= 3 && y <= 4 && x >= 3 && x <= 12
			stem := y >= 4 && y <= 12 && x >= 7 && x <= 8
			if topBar || stem {
				c = color.NRGBA{0xff, 0xff, 0xff, 0xff}
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}
