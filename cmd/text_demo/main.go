// Phase 11 demo: Multi-line text widget with scrollbar, tags, and editing.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Takigo Phase 11 — Text Widget"), takigo.Size(800, 600))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.SetBackgroundPixel(bgColor.Pixel)

	// Status label at bottom.
	statusLabel := label.New(app, "status",
		label.Text("Phase 11: Text Widget. Edit text, Ctrl+A select all, Ctrl+Z undo, Ctrl+Y redo. Esc to quit."),
		label.Background("#e8e8e8"),
		label.Anchor(option.AnchorW),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Create text widget.
	txt := text.New(app, "text",
		text.Width(80),
		text.Height(24),
		text.Background("white"),
		text.BorderWidthOpt(2),
		text.UndoOpt(true),
	)

	// Vertical scrollbar.
	yscroll := scrollbar.New(app, "yscroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.WidthOpt(14),
		scrollbar.CommandOpt(widget.ScrollY(txt)),
	)
	txt.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	// Pack scrollbar and text widget.
	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(txt, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Insert sample text.
	sampleText := `Welcome to the Takigo Text Widget!
This is a multi-line text editor.

Features:
  - Arrow keys navigate (with Shift for selection)
  - Ctrl+Left/Right for word movement
  - Home/End for line start/end
  - Ctrl+Home/End for document start/end
  - PageUp/PageDown for page scrolling
  - Mouse click to position cursor
  - Mouse drag to select text
  - Ctrl+A to select all
  - Ctrl+Z to undo, Ctrl+Y to redo
  - Backspace/Delete to remove text
  - Mouse wheel to scroll
  - Scrollbar integration

Tags Demo:
The words below have colored tags applied:
  - This line has a keyword highlighted.
  - This line has a comment style.
  - This line has an error style.

Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
Nisi ut aliquip ex ea commodo consequat.

Type here to test editing...`

	txt.Insert("1.0", sampleText)

	// Configure tags with colors.
	txt.TagConfigure("keyword", text.TagForeground("#0000cc"), text.TagBackground("#e8e8ff"))
	txt.TagConfigure("comment", text.TagForeground("#008000"))
	txt.TagConfigure("error", text.TagForeground("#cc0000"), text.TagUnderline(true))

	// Apply tags to specific ranges.
	txt.TagAdd("keyword", "20.34", "20.41") // "keyword"
	txt.TagAdd("comment", "21.34", "21.41") // "comment"
	txt.TagAdd("error", "22.30", "22.35")   // "error"

	// Move cursor to beginning.
	txt.SetInsertPos("1.0")
	txt.See("1.0")

	// Root event handlers.
	app.Dispatcher().Bind(root.PlatformID, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.PlatformID, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.Server
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == platform.XK_Escape {
			app.Quit()
		}
	})

	fmt.Println("Takigo Phase 11 Demo — Text Widget")
	fmt.Println("Edit text, use arrow keys, mouse, Ctrl+A/Z/Y. Esc to quit.")
	app.Run()
	fmt.Println("Goodbye!")

	// Keep references alive.
	_ = statusLabel
	_ = txt
	_ = yscroll
	_ = widget.CompoundNone
}
