package canvas_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/canvas"
	"github.com/takigo/takigo/geometry/pack"
)

func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Canvas"))
	if err != nil {
		log.Fatal(err)
	}
	c := canvas.New(app, "c", canvas.Width(300), canvas.Height(200), canvas.Background("white"))
	pack.Pack(c, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	c.CreateRectangle(20, 20, 120, 90, canvas.FillColor("lightblue"), canvas.OutlineColor("navy"))
	oval := c.CreateOval(150, 30, 270, 150, canvas.FillColor("gold"), canvas.OutlineWidth(2))
	c.CreateLine([]float64{20, 180, 280, 180}, canvas.OutlineWidth(3), canvas.Arrow(canvas.ArrowLast))
	c.Move(oval, -10, 10)
	app.Run()
}
