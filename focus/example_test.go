package focus_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/entry"
)

// The App's focus manager gives the keyboard focus to a widget; Tab and
// Shift-Tab move it along the stacking order as in Tk.
func Example() {
	app, err := takigo.NewApp(takigo.Title("Focus"))
	if err != nil {
		log.Fatal(err)
	}
	user := entry.New(app, "user", entry.Width(20))
	password := entry.New(app, "password", entry.Width(20), entry.Show('*'))
	pack.Pack(user, pack.PadX(10), pack.PadY(4))
	pack.Pack(password, pack.PadX(10), pack.PadY(4))
	app.FocusManager().SetFocus(user.Window())
	app.Run()
}
