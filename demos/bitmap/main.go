// Demo: Displays all of Tk's built-in bitmaps.
// Ported from Tk's bitmap.tcl demo.
package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget/frame"
	"github.com/takigo/takigo/widget/label"
)

func findBitmap(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	path := filepath.Join(projectRoot, "tk", "bitmaps", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Bitmap Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("bitmap"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength(screenunit.In(4)),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window displays all of Tk's built-in bitmaps, along with the names you can use for them in Tcl scripts."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	fg := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	transparent := color.RGBA{R: 0, G: 0, B: 0, A: 0}

	bitmapNames := []string{
		"error", "gray12", "gray25", "gray50", "gray75",
		"hourglass", "info", "question", "questhead", "warning",
	}

	photos := make([]*tkimage.Photo, len(bitmapNames))
	for i, name := range bitmapNames {
		path := findBitmap(name + ".xbm")
		photo, loadErr := tkimage.NewPhotoFromXBMFile("bm_"+name, path, fg, transparent)
		if loadErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load %s: %v\n", name, loadErr)
			continue
		}
		photos[i] = photo
		app.ImageRegistry().Register(photo)
	}

	container := frame.New(f, "frame")

	// bitmapRow 0: first 5 bitmaps.
	row0 := frame.New(container, "0")
	pack.Pack(row0, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth))

	for i := range 5 {
		col := frame.New(row0, fmt.Sprintf("%d", i))
		pack.Pack(col, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.PadY(screenunit.Cm(0.25)), pack.PadX(screenunit.Cm(0.25)))
		nl := label.New(col, "label", label.Text(bitmapNames[i]), label.Width(9))
		var il *label.Label
		if photos[i] != nil {
			il = label.New(col, "bitmap", label.ImageOpt(photos[i]))
		} else {
			il = label.New(col, "bitmap", label.Text("?"))
		}
		pack.Pack(nl, pack.SideOpt(pack.Bottom))
		pack.Pack(il, pack.SideOpt(pack.Bottom))
	}

	// bitmapRow 1: last 5 bitmaps.
	row1 := frame.New(container, "1")
	pack.Pack(row1, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth))

	for i := range 5 {
		idx := i + 5
		col := frame.New(row1, fmt.Sprintf("%d", i))
		pack.Pack(col, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.PadY(screenunit.Cm(0.25)), pack.PadX(screenunit.Cm(0.25)))
		nl := label.New(col, "label", label.Text(bitmapNames[idx]), label.Width(9))
		var il *label.Label
		if photos[idx] != nil {
			il = label.New(col, "bitmap", label.ImageOpt(photos[idx]))
		} else {
			il = label.New(col, "bitmap", label.Text("?"))
		}
		pack.Pack(nl, pack.SideOpt(pack.Bottom))
		pack.Pack(il, pack.SideOpt(pack.Bottom))
	}

	pack.Pack(container, pack.SideOpt(pack.Top), pack.Expand(true), pack.FillOpt(pack.FillBoth))

	app.Run()
}
