// Demo: Labels with various styles and options.
// Ported from Tk's label.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app := demohelper.Setup("Label Demonstration", 450, 350,
		"Five labels are displayed below: three textual ones on the left, "+
			"and an image label and a text label on the right. Labels are "+
			"pretty boring because you can't do anything with them.")

	// Two side-by-side frames.
	left := frame.New(app, "left")
	right := frame.New(app, "right")
	pack.Pack(left, pack.SideOpt(pack.Left), pack.Expand(true),
		pack.PadX(10), pack.PadY(10), pack.FillOpt(pack.FillBoth))
	pack.Pack(right, pack.SideOpt(pack.Left), pack.Expand(true),
		pack.PadX(10), pack.PadY(10), pack.FillOpt(pack.FillBoth))

	// Left side: three text labels.
	l1 := label.New(left, "l1", label.Text("First label"))
	l2 := label.New(left, "l2", label.Text("Second label, raised"),
		label.Relief(option.ReliefRaised))
	l3 := label.New(left, "l3", label.Text("Third label, sunken"),
		label.Relief(option.ReliefSunken))
	pack.Pack(l1, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY(2), pack.Anchor(option.AnchorW))
	pack.Pack(l2, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY(2), pack.Anchor(option.AnchorW))
	pack.Pack(l3, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY(2), pack.Anchor(option.AnchorW))

	// Right side: image + caption.
	imgPath := findImage("ouster.png")
	if imgPath != "" {
		photo, err := image.NewPhotoFromFile("ouster", imgPath)
		if err == nil {
			app.ImageRegistry().Register(photo)
			pic := label.New(right, "picture",
				label.ImageOpt(photo),
				label.BorderWidth(2),
				label.Relief(option.ReliefSunken),
			)
			pack.Pack(pic, pack.SideOpt(pack.Top))
		} else {
			fmt.Fprintf(os.Stderr, "Warning: could not load image: %v\n", err)
		}
	}
	caption := label.New(right, "caption", label.Text("Tcl/Tk Creator"))
	pack.Pack(caption, pack.SideOpt(pack.Top))

	app.Run()
}

// findImage locates an image in the demos/images/ directory.
func findImage(name string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	// demos/label/main.go → demos/ is one level up
	demosRoot := filepath.Dir(filepath.Dir(file))
	path := filepath.Join(demosRoot, "images", name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}
