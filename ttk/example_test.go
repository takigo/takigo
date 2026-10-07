package ttk_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
)

func ExampleNewButton() {
	app, err := takigo.NewApp(takigo.Title("ttk"))
	if err != nil {
		log.Fatal(err)
	}
	b := ttk.NewButton(app, "ok",
		ttk.ButtonText("OK"),
		ttk.ButtonStyle("Accent.TButton"),
		ttk.ButtonCommand(app.Quit),
	)
	pack.Pack(b, pack.PadX(10), pack.PadY(10))
	app.Run()
}

// A notebook shows one of its panes at a time; the panes are its children.
func ExampleNewNotebook() {
	app, err := takigo.NewApp(takigo.Title("Notebook"))
	if err != nil {
		log.Fatal(err)
	}
	nb := ttk.NewNotebook(app, "nb")
	pack.Pack(nb, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	for _, name := range []string{"General", "Advanced"} {
		page := ttk.NewFrame(nb, name)
		nb.Add(page.Win, name)
	}
	nb.TabConfigure(1, ttk.TabState(ttk.StateDisabled))
	app.Run()
}

// A treeview with columns is a table; each item is a row.
func ExampleNewTreeview() {
	app, err := takigo.NewApp(takigo.Title("Treeview"))
	if err != nil {
		log.Fatal(err)
	}
	tv := ttk.NewTreeview(app, "tv", ttk.TreeviewColumns("size", "modified"))
	pack.Pack(tv, pack.FillOpt(pack.FillBoth), pack.Expand(true))
	docs := tv.Insert("", 0, ttk.ItemText("Documents"), ttk.ItemValues("", "today"))
	tv.Insert(docs, 0, ttk.ItemText("notes.txt"), ttk.ItemValues("2 KB", "today"))
	app.Run()
}
