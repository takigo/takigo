// Demo: Photo images in labels.
// Ported from Tk's image1.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Image Demonstration #1"),
		takigo.Geometry("+300+300"),
		takigo.IconName("Image1"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This demonstration displays two images, each in a separate label widget."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Load the two earth images from demos/images/.
	for _, info := range []struct {
		name, file string
	}{
		{"image1a", "earth.gif"},
		{"image1b", "earthris.gif"},
	} {
		path := findImage(info.file)
		if path == "" {
			fmt.Fprintf(os.Stderr, "Warning: could not find image %s\n", info.file)
			continue
		}
		photo, err := tkimage.NewPhotoFromFile(info.name, path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not load image %s: %v\n", info.file, err)
			continue
		}
		app.ImageRegistry().Register(photo)

		l := label.New(f, "l_"+info.name,
			label.ImageOpt(photo),
			label.BorderWidth(1),
			label.Relief(option.ReliefSunken),
		)
		pack.Pack(l, pack.SideOpt(pack.Top), pack.PadX(".5m"), pack.PadY(".5m"))
		_ = l
	}

	app.Run()
}

// findImage locates an image in the demos/images/ directory.
func findImage(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	demosRoot := filepath.Dir(filepath.Dir(file))
	path := filepath.Join(demosRoot, "images", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
