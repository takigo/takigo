package color_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/color"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
)

// Parse accepts Tk colour names and "#rgb" strings and returns 16-bit
// components.
func ExampleParse() {
	for _, s := range []string{"red", "light blue", "#3a7", "#ff8000"} {
		r, g, b, err := color.Parse(s)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%-10s %04x %04x %04x\n", s, r, g, b)
	}
	// Output:
	// red        ffff 0000 0000
	// light blue adad d8d8 e6e6
	// #3a7       3333 aaaa 7777
	// #ff8000    ffff 8080 0000
}

// A colour option takes a name or a color.Value.
func ExampleRGB() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	l := label.New(app, "l",
		label.Text("orange on navy"),
		label.Foreground(color.RGB(255, 165, 0)),
		label.Background("navy"),
	)
	pack.Pack(l, pack.PadX(20), pack.PadY(20))
	app.Run()
}
