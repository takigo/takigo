package image_test

import (
	"fmt"
	goimage "image"
	gocolor "image/color"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	tkimage "github.com/takigo/takigo/image"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/label"
)

// A Photo wraps Go RGBA pixels; NewPhotoFromPhoto copies one, scaled like
// Tk's "copy -zoom".
func ExampleNewPhoto() {
	rgba := goimage.NewRGBA(goimage.Rect(0, 0, 16, 8))
	for y := range 8 {
		for x := range 16 {
			rgba.Set(x, y, gocolor.RGBA{R: uint8(x * 16), G: 0, B: 255, A: 255})
		}
	}
	p := tkimage.NewPhoto("gradient", rgba)
	big := tkimage.NewPhotoFromPhoto(p, "gradient2", tkimage.Zoom(2))
	fmt.Println(p.Name(), p.Width(), p.Height())
	fmt.Println(big.Name(), big.Width(), big.Height())
	// Output:
	// gradient 16 8
	// gradient2 32 16
}

// Widgets show a photo through their image option; the App's registry
// keeps it by name and destroys it with the App.
func ExampleNewPhotoFromFile() {
	app, err := takigo.NewApp(takigo.Title("Image"))
	if err != nil {
		log.Fatal(err)
	}
	logo, err := tkimage.NewPhotoFromFile("logo", "logo.png")
	if err != nil {
		log.Fatal(err)
	}
	app.ImageRegistry().Register(logo)
	pack.Pack(label.New(app, "logo", label.ImageOpt(logo)))
	pack.Pack(button.New(app, "b", button.ImageOpt(logo), button.Command(app.Quit)))
	app.Run()
}
