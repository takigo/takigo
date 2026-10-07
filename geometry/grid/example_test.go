package grid_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget/entry"
	"github.com/takigo/takigo/widget/label"
)

// Grid places children in rows and columns; a weighted column grows with
// the window.
func ExampleGrid() {
	app, err := takigo.NewApp(takigo.Title("grid"))
	if err != nil {
		log.Fatal(err)
	}
	for row, field := range []string{"Name", "Email"} {
		l := label.New(app, field, label.Text(field+":"))
		e := entry.New(app, field+"entry", entry.Width(30))
		grid.Grid(l, grid.Row(row), grid.Column(0), grid.Sticky(option.StickE), grid.PadX(4), grid.PadY(2))
		grid.Grid(e, grid.Row(row), grid.Column(1), grid.Sticky(option.StickEW), grid.PadX(4), grid.PadY(2))
	}
	grid.ColumnConfigure(app, 1, grid.Weight(1))
	app.Run()
}
