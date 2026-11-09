package scrollbar_test

import (
	"fmt"
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/scrollbar"
	"github.com/takigo/takigo/widget/text"
)

// A scrollbar scrolls a widget through its command, and the widget moves the
// slider through its scroll command: the two are wired to each other.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Scrollbar"))
	if err != nil {
		log.Fatal(err)
	}
	tw := text.New(app, "t", text.Width(40), text.Height(10))
	sb := scrollbar.New(app, "sb",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(widget.ScrollY(tw)),
	)
	tw.Configure(text.YScrollCommand(sb.Set))
	for i := range 100 {
		tw.Insert("end", fmt.Sprintf("line %d\n", i+1))
	}
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	app.Run()
}
