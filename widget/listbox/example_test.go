package listbox_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/listbox"
	"github.com/takigo/takigo/widget/scrollbar"
)

// A listbox and the scrollbar that scrolls it are wired to each other.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Listbox"))
	if err != nil {
		log.Fatal(err)
	}
	lb := listbox.New(app, "list", listbox.Height(5), listbox.SelectModeOpt(listbox.SelectExtended))
	for _, name := range []string{"Alpha", "Beta", "Gamma", "Delta", "Epsilon", "Zeta", "Eta"} {
		lb.Insert(lb.ItemCount(), name)
	}
	sb := scrollbar.New(app, "sb", scrollbar.CommandOpt(widget.ScrollY(lb)))
	lb.Configure(listbox.YScrollCommand(sb.Set))
	pack.Pack(sb, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(lb, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	app.Run()
}
