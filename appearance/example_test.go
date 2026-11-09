package appearance_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/appearance"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
)

func ExampleMode() {
	fmt.Println(appearance.Light, appearance.Dark)
	// Output: light dark
}

// System reports the desktop's light or dark preference (TAKIGO_APPEARANCE
// overrides it); an App created with UseAppearance in that mode starts its
// widgets with the matching default colours.
func ExampleSystem() {
	app, err := takigo.NewApp(takigo.UseAppearance(appearance.System()))
	if err != nil {
		log.Fatal(err)
	}
	pack.Pack(label.New(app, "l", label.Text("Drawn "+app.Appearance().String())), pack.PadX(20), pack.PadY(20))
	app.Run()
}
