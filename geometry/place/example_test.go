package place_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/place"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/label"
)

// Place puts a child at absolute or relative coordinates in its parent;
// relative ones follow the parent when it is resized.
func ExamplePlace() {
	app, err := takigo.NewApp(takigo.Title("place"), takigo.Size(300, 200))
	if err != nil {
		log.Fatal(err)
	}
	banner := label.New(app, "banner", label.Text("top strip"), label.Background("light blue"))
	place.Place(banner, place.X(0), place.Y(0), place.RelWidth(1), place.Height(30))

	ok := button.New(app, "ok", button.Text("OK"), button.Command(app.Quit))
	place.Place(ok, place.RelX(0.5), place.RelY(0.5), place.Anchor(option.AnchorCenter))

	corner := label.New(app, "corner", label.Text("corner"))
	place.Place(corner, place.RelX(1), place.RelY(1), place.X(-4), place.Y(-4), place.Anchor(option.AnchorSE))
	app.Run()
}
