// Demo: Iconic buttons that display bitmaps instead of text.
// Ported from Tk's icon.tcl demo.
package main

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
	"github.com/msorc/takigo/geometry"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Iconic Button Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("icon"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("5i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("This window shows three ways of using bitmaps or images in "+
			"radiobuttons and checkbuttons.  On the left are two "+
			"radiobuttons, each of which displays a bitmap and an "+
			"indicator.  In the middle is a checkbutton that displays a "+
			"different image depending on whether it is selected or not.  "+
			"On the right is a checkbutton that displays a single bitmap "+
			"but changes its background color to indicate whether or not "+
			"it is selected."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Load XBM images.
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 0}

	flagup, _ := tkimage.NewPhotoFromXBMFile("flagup", findImage("flagup.xbm"), black, white)
	flagdown, _ := tkimage.NewPhotoFromXBMFile("flagdown", findImage("flagdown.xbm"), black, white)
	lettersImg, _ := tkimage.NewPhotoFromXBMFile("letters", findImage("letters.xbm"), black, white)
	noletterImg, _ := tkimage.NewPhotoFromXBMFile("noletter", findImage("noletter.xbm"), black, white)

	// frame $w.frame -borderwidth 7.5p
	fr := frame.New(f, "frame",
		frame.BorderWidth(screenunit.Px("7.5p")),
	)
	pack.Pack(fr, pack.SideOpt(pack.Top))

	// checkbutton $w.frame.b1 -image flagdown -selectimage flagup -indicatoron 0
	b1 := checkbutton.New(fr, "b1",
		checkbutton.ImageOpt(flagdown),
		checkbutton.SelectImageOpt(flagup),
		checkbutton.IndicatorOnOpt(false),
	)
	// $w.frame.b1 configure -selectcolor [$w.frame.b1 cget -background]
	// (set selectcolor to background so it's invisible when selected)
	b1.SelectColor = b1.Background.Ref()

	// checkbutton $w.frame.b2 -bitmap letters -indicatoron 0 -selectcolor SeaGreen1
	b2 := checkbutton.New(fr, "b2",
		checkbutton.ImageOpt(lettersImg),
		checkbutton.IndicatorOnOpt(false),
	)

	if sc, err := app.ColorCache().Get("SeaGreen1"); err == nil {
		b2.SelectColor = sc.Ref()
	}

	// frame $w.frame.left
	left := frame.New(fr, "left")

	// pack $w.frame.left $w.frame.b1 $w.frame.b2 -side left -expand yes -padx 5m
	pack.Pack(geometry.Group{left, b1, b2}, pack.SideOpt(pack.Left), pack.Expand(true), pack.PadX("5m"))

	// radiobutton $w.frame.left.b3 -bitmap letters -variable letters -value full
	lettersVar := widget.NewVariable("letters")
	b3 := radiobutton.New(left, "b3",
		radiobutton.ImageOpt(lettersImg),
		radiobutton.Var(lettersVar),
		radiobutton.Value("full"),
	)
	// radiobutton $w.frame.left.b4 -bitmap noletter -variable letters -value empty
	b4 := radiobutton.New(left, "b4",
		radiobutton.ImageOpt(noletterImg),
		radiobutton.Var(lettersVar),
		radiobutton.Value("empty"),
	)
	pack.Pack(geometry.Group{b3, b4}, pack.SideOpt(pack.Top), pack.Expand(true))

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
