package cursor_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/cursor"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/label"
)

// A window shows its cursor while the pointer is over it; ResetCursor goes
// back to the parent's.
func ExampleShape() {
	app, err := takigo.NewApp(takigo.Title("Cursor"))
	if err != nil {
		log.Fatal(err)
	}
	link := label.New(app, "link", label.Text("https://example.com"), label.Foreground("blue"))
	pack.Pack(link, pack.PadX(20), pack.PadY(20))
	link.Window().SetCursor(cursor.Hand2)
	app.Run()
}
