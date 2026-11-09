package message_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/message"
)

// A message breaks its text into lines to keep the -aspect ratio
// (100*width/height) unless -width is given.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Message"))
	if err != nil {
		log.Fatal(err)
	}
	m := message.New(app, "msg",
		message.Text("A message widget displays a multi-line string, breaking it at word boundaries."),
		message.Aspect(300),
		message.JustifyOpt(option.JustifyCenter),
		message.Relief(option.ReliefSunken),
		message.BorderWidth(1),
	)
	pack.Pack(m, pack.PadX(10), pack.PadY(10))
	app.Run()
}
