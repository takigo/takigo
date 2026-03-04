// Demo: Photo images in labels.
// Ported from Tk's image1.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("Image Demonstration #1", 500, 400,
		"This demonstration displays two images, each in a separate label widget.")

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

		l := label.New(app, "l_"+info.name,
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
