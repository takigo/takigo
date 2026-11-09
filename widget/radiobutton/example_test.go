package radiobutton_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/label"
	"github.com/takigo/takigo/widget/radiobutton"
)

// Radiobuttons sharing a Variable[string] form a group: selecting one sets
// the variable to its Value.
func ExampleNew() {
	app, err := takigo.NewApp(takigo.Title("Radiobutton"))
	if err != nil {
		log.Fatal(err)
	}
	size := widget.NewVariable("medium")
	shown := label.New(app, "shown", label.Text("Size: medium"))
	size.OnChange(func(_, cur string) { shown.Configure(label.Text("Size: " + cur)) })
	for _, v := range []string{"small", "medium", "large"} {
		r := radiobutton.New(app, v,
			radiobutton.Text(v),
			radiobutton.Value(v),
			radiobutton.Var(size),
		)
		pack.Pack(r, pack.Anchor(option.AnchorW))
	}
	pack.Pack(shown, pack.PadY(4))
	app.Run()
}
