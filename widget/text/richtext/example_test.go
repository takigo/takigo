package richtext_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/text"
	"github.com/takigo/takigo/widget/text/richtext"
)

// Render shows an HTML document in a text widget; Stylesheet adds CSS and
// OnLink receives the links that are clicked.
func ExampleRender() {
	app, err := takigo.NewApp(takigo.Title("Help"))
	if err != nil {
		log.Fatal(err)
	}
	tw := text.New(app, "t", text.Width(60), text.Height(20), text.ReadOnly(true))
	pack.Pack(tw, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	err = richtext.Render(tw, `
		<h1>Welcome</h1>
		<p class="lead">Press <kbd>F1</kbd> for help, or read the
		<a href="https://example.org/manual">manual</a>.</p>
		<ul><li>Fast</li><li>Small</li></ul>
		<pre>go run .</pre>`,
		richtext.Stylesheet(`.lead { font-size: 1.2em; color: #444 }`),
		richtext.OnLink(func(href string) { log.Println("open", href) }),
	)
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
