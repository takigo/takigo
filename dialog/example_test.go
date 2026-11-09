package dialog_test

import (
	"log"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/dialog"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/widget/button"
	"github.com/takigo/takigo/widget/label"
)

// The dialogs are modal: OpenFile returns once the user picked a file or
// cancelled, while the event loop keeps running.
func ExampleOpenFile() {
	app, err := takigo.NewApp(takigo.Title("Open"))
	if err != nil {
		log.Fatal(err)
	}
	chosen := label.New(app, "chosen", label.Text("no file"))
	open := button.New(app, "open", button.Text("Open..."), button.Command(func() {
		path, ok := dialog.OpenFile(app,
			dialog.FileTitle("Open a Go file"),
			dialog.FileTypes(
				dialog.FileType{Name: "Go files", Pattern: "*.go"},
				dialog.FileType{Name: "All files", Pattern: "*"},
			),
		)
		if ok {
			chosen.Configure(label.Text(path))
		}
	}))
	pack.Pack(open, pack.PadY(4))
	pack.Pack(chosen, pack.PadX(10), pack.PadY(4))
	app.Run()
}

func ExampleShowMessage() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	save := button.New(app, "save", button.Text("Close"), button.Command(func() {
		switch dialog.ShowMessage(app,
			dialog.MsgTitle("Unsaved changes"),
			dialog.MsgMessage("Save the document before closing?"),
			dialog.MsgDetail("Your changes are lost if you don't save them."),
			dialog.MsgType(dialog.MsgWarning),
			dialog.MsgButtons(dialog.BtnYesNoCancel),
		) {
		case dialog.ResultYes:
			log.Println("saving")
			app.Quit()
		case dialog.ResultNo:
			app.Quit()
		default: // Cancel or the window closed: keep editing
		}
	}))
	pack.Pack(save, pack.PadX(20), pack.PadY(20))
	app.Run()
}

func ExampleAskString() {
	app, err := takigo.NewApp()
	if err != nil {
		log.Fatal(err)
	}
	greeting := label.New(app, "greeting", label.Text("Hello"))
	ask := button.New(app, "ask", button.Text("Name..."), button.Command(func() {
		if name, ok := dialog.AskString(app, "Name", "What is your name?", ""); ok {
			greeting.Configure(label.Text("Hello, " + name))
		}
	}))
	pack.Pack(greeting, pack.PadX(20), pack.PadY(10))
	pack.Pack(ask, pack.PadY(10))
	app.Run()
}
