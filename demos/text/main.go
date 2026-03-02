// Demo: Basic text editing with scrollbar and undo/redo.
// Ported from Tk's text.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Widget Demo", 550, 450,
		"A text widget with scrollbar. Click to position\ncursor. Select by dragging. Ctrl+Z to undo.")

	// Text widget with scrollbar.
	txtFrame := frame.New(app, "txtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tw := text.New(txtFrame, "text1",
		text.Width(60),
		text.Height(20),
		text.WrapModeOpt(text.WrapWord),
		text.UndoOpt(true),
	)

	yscroll := scrollbar.New(txtFrame, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Insert sample text.
	sampleText := `This window is a text widget. It displays one or more lines of text and allows you to edit the text.

You can scroll through the text using the scrollbar on the right side, or by dragging with the middle mouse button.

Text editing features:
  - Click to position the insertion cursor
  - Drag to select text
  - Ctrl+Z to undo, Ctrl+Y to redo
  - Backspace and Delete to remove characters
  - Arrow keys to move the cursor
  - Home/End for line start/end

The text widget supports word wrapping, which is enabled in this demo. Long lines are automatically wrapped at word boundaries to fit within the visible width.

Try typing some text, selecting it, and using undo/redo to see the editing capabilities in action.
`
	tw.Insert("1.0", sampleText)

	app.Run()
}
