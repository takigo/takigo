// Demo: Basic text editing with scrollbar.
// Ported from Tk's text.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Basic Facilities"),
		takigo.Geometry("+300+300"),
		takigo.IconName("text"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("A text widget with scrollbar. Click to position cursor. Select by dragging. Ctrl+Z to undo."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	tw := text.New(f, "text",
		text.Height(30),
		text.WrapModeOpt(text.WrapWord),
		text.UndoOpt(true),
		text.SetGridOpt(true),
	)

	yscroll := ttk.NewScrollbar(f, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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

	// Font Chooser button (matches Tcl's fontchooser toggle in button bar).
	fontBtn := ttk.NewButton(f, "fontchooser",
		ttk.ButtonText("Font Chooser..."),
		ttk.ButtonCommand(func() {
			if f, ok := dialog.ChooseFont(app); ok {
				if parsed, err := app.FontRegistry().Get(f); err == nil {
					tw.Font = parsed
					tw.Display()
				}
			}
		}),
	)
	pack.Pack(fontBtn, pack.SideOpt(pack.Bottom), pack.PadX(4), pack.PadY(4))

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Set initial focus and insert position.
	app.After(0, func() {
		app.Server().SetInputFocus(tw.Window().PlatformID, platform.RevertToParent, platform.CurrentTime)
	})

	// Insert sample text matching Tk's text.tcl (7-item list).
	tw.Insert("1.0", `This window is a text widget.  It displays one or more lines of text
and allows you to edit the text.  Here is a summary of the things you
can do to a text widget:

1. Scrolling. Use the scrollbar to adjust the view in the text window.

2. Scanning. Press the middle mouse button in the text window and drag up
or down. This will drag the text at high speed to allow you to scan its
contents.

3. Insert text. Press mouse button 1 to set the insertion cursor, then
type text.  What you type will be added to the widget.

4. Select. Press mouse button 1 and drag to select a range of characters.
Once you've released the button, you can adjust the selection by pressing
button 1 with the shift key down.  This will reset the end of the
selection nearest the mouse cursor and you can drag that end of the
selection by dragging the mouse before releasing the mouse button.
You can double-click to select whole words or triple-click to select
whole lines.

5. Delete and replace. To delete text, select the characters you'd like
to delete and type Backspace or Delete.  Alternatively, you can type new
text, in which case it will replace the selected text.

6. Copy the selection. To copy the selection into this window, select
what you want to copy (either here or in another application), then
click the middle mouse button to copy the selection to the point of the
mouse cursor.

7. Edit.  Text widgets support the standard Motif editing characters
plus many Emacs editing characters.  Backspace and Control-h erase the
character to the left of the insertion cursor.  Delete and Control-d
erase the character to the right of the insertion cursor.  Meta-backspace
deletes the word to the left of the insertion cursor, and Meta-d deletes
the word to the right of the insertion cursor.  Control-k deletes from
the insertion cursor to the end of the line, or it deletes the newline
character if that is the only thing left on the line.  Control-o opens
a new line by inserting a newline character to the right of the insertion
cursor.  Control-t transposes the two characters on either side of the
insertion cursor.  Control-z undoes the last editing action performed,
and Control-Shift-z redoes undone edits.

7. Resize the window.  Also, if you make the window narrow you can see
that long lines automatically wrap around onto additional lines so that
all the information is always visible.`)

	app.Run()
}
