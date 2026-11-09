package bitmap_test

import (
	"fmt"
	gocolor "image/color"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/bitmap"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
)

// Data returns a built-in bitmap's size and XBM bits.
func ExampleData() {
	for _, name := range []string{"info", "questhead", "gray50", "nosuch"} {
		w, h, _, ok := bitmap.Data(name)
		fmt.Println(name, w, h, ok)
	}
	// Output:
	// info 8 21 true
	// questhead 20 22 true
	// gray50 16 16 true
	// nosuch 0 0 false
}

// Get renders a built-in bitmap as a Photo in the given colours, for any
// widget's image option.
func ExampleGet() {
	app, err := takigo.NewApp(takigo.Title("Bitmap"))
	if err != nil {
		log.Fatal(err)
	}
	transparent := gocolor.RGBA{}
	for _, name := range []string{"info", "warning", "error", "question"} {
		img := bitmap.Get(name, gocolor.RGBA{A: 255}, transparent)
		app.ImageRegistry().Register(img)
		pack.Pack(label.New(app, name, label.ImageOpt(img)), pack.SideOpt(pack.Left), pack.PadX(8), pack.PadY(8))
	}
	app.Run()
}
