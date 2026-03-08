// Demo: Labels with various styles and options.
// Ported from Tk's label.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Label Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("label"),
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
		label.Text("Five labels are displayed below: three textual ones on the left, "+
			"and an image label and a text label on the right.  Labels are "+
			"pretty boring because you can't do anything with them."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	left := frame.New(f, "left")
	right := frame.New(f, "right")
	pack.Pack(left, pack.SideOpt(pack.Left), pack.Expand(true),
		pack.PadX("7.5p"), pack.PadY("7.5p"), pack.FillOpt(pack.FillBoth))
	pack.Pack(right, pack.SideOpt(pack.Left), pack.Expand(true),
		pack.PadX("7.5p"), pack.PadY("7.5p"), pack.FillOpt(pack.FillBoth))

	l1 := label.New(left, "l1", label.Text("First label"))
	l2 := label.New(left, "l2", label.Text("Second label, raised"),
		label.Relief(option.ReliefRaised))
	l3 := label.New(left, "l3", label.Text("Third label, sunken"),
		label.Relief(option.ReliefSunken))
	// pack $w.left.l1 $w.left.l2 $w.left.l3 -side top -expand yes -pady 1.5p -anchor w
	pack.Pack(l1, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY("1.5p"), pack.Anchor(option.AnchorW))
	pack.Pack(l2, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY("1.5p"), pack.Anchor(option.AnchorW))
	pack.Pack(l3, pack.SideOpt(pack.Top), pack.Expand(true),
		pack.PadY("1.5p"), pack.Anchor(option.AnchorW))

	ousterhout, err := image.NewPhotoFromFile("ousterhout", findImage("ouster.png"))
	if err == nil {
		// Zoom by integer scaling factor, matching Tk's: -zoom [expr {$tk::scalingPct / 100}]
		// At 144 DPI scalingPct=150, 150/100=1 (integer division), so no zoom.
		// At 192 DPI scalingPct=200, 200/100=2, so 2x zoom.
		ousterhout2 := image.NewPhotoFromPhoto(ousterhout, "ousterhout2", image.Zoom(float64(screenunit.ScalingPct()/100)))

		picture := label.New(right, "picture",
			label.ImageOpt(ousterhout2),
			label.BorderWidth(2),
			label.Relief(option.ReliefSunken),
		)
		// pack $w.right.picture $w.right.caption -side top
		pack.Pack(picture, pack.SideOpt(pack.Top))
	} else {
		fmt.Fprintf(os.Stderr, "Warning: could not load image: %v\n", err)
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
