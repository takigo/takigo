package entry_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/entry"
)

func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Entry"))
	if err != nil {
		log.Fatal(err)
	}
	e := entry.New(app, "name", entry.Width(30), entry.Placeholder("Your name"))
	pack.Pack(e, pack.PadX(10), pack.PadY(10))
	pack.Pack(button.New(app, "ok", button.Text("OK"), button.Command(func() {
		log.Println("name:", e.GetText())
		app.Quit()
	})))
	app.Run()
}

// Validation sees the text an edit would produce and may refuse it.
func ExampleValidateCmdOpt() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	digits := entry.New(app, "digits",
		entry.ValidateOpt("key"),
		entry.ValidateCmdOpt(func(prospective string) bool {
			for _, r := range prospective {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}),
	)
	pack.Pack(digits)
	app.Run()
}
