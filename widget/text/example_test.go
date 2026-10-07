package text_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/text"
)

// A text widget edits lines; indexes are "line.char" or a text.Index.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Text"))
	if err != nil {
		log.Fatal(err)
	}
	tw := text.New(app, "t", text.Width(40), text.Height(10), text.WrapModeOpt(text.WrapWord))
	pack.Pack(tw, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	tw.Insert("end", "Hello, world.\nA second line.")
	tw.Insert("1.0", "> ")
	app.Run()
}
