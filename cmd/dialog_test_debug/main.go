package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/window"
)

func dumpWindow(w *window.Window, depth int) {
	indent := strings.Repeat("  ", depth)
	flags := ""
	if w.Flags&window.FlagMapped != 0 {
		flags += " MAPPED"
	}
	if w.Flags&window.FlagTopLevel != 0 {
		flags += " TOPLEVEL"
	}
	fmt.Printf("%s%s: pos=(%d,%d) size=%dx%d req=%dx%d xwin=%d%s\n",
		indent, w.PathName, w.X, w.Y, w.Width, w.Height, w.ReqWidth, w.ReqHeight, w.XWindow, flags)
	for _, c := range w.Children {
		dumpWindow(c, depth+1)
	}
}

func main() {
	app, err := takigo.NewApp(
		takigo.Title("Dialog Debug"),
		takigo.Size(400, 300),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	// Open dialogs after a short delay.
	app.After(500*time.Millisecond, func() {
		fmt.Println("=== Opening message dialog ===")
		app.After(500*time.Millisecond, func() {
			fmt.Println("\n=== Window tree ===")
			dumpWindow(app.Root(), 0)
		})
		result := dialog.ShowMessage(app,
			dialog.MsgParent(app.Root()),
			dialog.MsgTitle("Test"),
			dialog.MsgMessage("Hello World"),
			dialog.MsgDetail("Some detail text"),
			dialog.MsgType(dialog.MsgInfo),
			dialog.MsgButtons(dialog.BtnOKCancel),
		)
		fmt.Printf("Result: %d\n", result)

		fmt.Println("\n=== Opening color chooser ===")
		app.After(500*time.Millisecond, func() {
			fmt.Println("\n=== Window tree ===")
			dumpWindow(app.Root(), 0)
		})
		color, ok := dialog.ChooseColor(app,
			dialog.ColorParent(app.Root()),
			dialog.ColorInitial("#3399ff"),
		)
		fmt.Printf("Color result: %q ok=%v\n", color, ok)
	})

	app.MainLoop()
}
