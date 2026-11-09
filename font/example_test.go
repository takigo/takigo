package font_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
)

// ParseDescriptor reads Tk's font descriptions; a negative size is in
// pixels, a positive one in points.
func ExampleParseDescriptor() {
	for _, s := range []string{"Helvetica 12 bold", "{Courier New} 10 italic underline", "-family Times -size -14"} {
		a, err := font.ParseDescriptor(s)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("%s, %v: bold %v, italic %v, underline %v\n",
			a.Family, a.Size, a.Weight == font.WeightBold, a.Slant == font.SlantItalic, a.Underline)
	}
	// Output:
	// Helvetica, 12: bold true, italic false, underline false
	// Courier New, 10: bold false, italic true, underline true
	// Times, -14: bold false, italic false, underline false
}

// A font option takes a descriptor, a named font defined in the App's
// registry or the Attributes themselves.
func ExampleRegistry_Define() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	app.FontRegistry().Define("Heading", font.Attributes{Family: "Helvetica", Size: 16, Weight: font.WeightBold})
	pack.Pack(label.New(app, "h", label.Text("Heading"), label.FontOpt("Heading")))
	pack.Pack(label.New(app, "body", label.Text("Body text"), label.FontOpt("Times 12")))
	pack.Pack(label.New(app, "code", label.Text("x := 1"), label.FontOpt(font.Attributes{Family: "Courier", Size: 11})))
	app.Run()
}
