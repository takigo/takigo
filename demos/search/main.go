// Demo: Text search and highlight.
// Ported from Tk's search.tcl demo (simplified — manual search loop).
package main

import (
	"fmt"
	"strings"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/focus"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Search Demo", 600, 500,
		"Type a search string below and click Highlight\nto find and highlight all matches in the text.")
	root := app.Window()

	focusMgr := focus.NewManager(app.Dispatcher(), app.DisplayPtr())
	focusMgr.BindTraversal(root)

	// Search bar.
	searchFrame := frame.New(app, "searchframe")
	pack.Pack(searchFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(10), pack.PadY(5))

	searchLabel := label.New(searchFrame, "slabel",
		label.Text("Search:"),
	)
	pack.Pack(searchLabel, pack.SideOpt(pack.Left), pack.PadX(5))

	searchEntry := entry.New(searchFrame, "sentry",
		entry.Width(20),
	)
	pack.Pack(searchEntry, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillX),
		pack.Expand(true), pack.PadX(5))

	// Status label.
	statusLabel := label.New(app, "status",
		label.Text("0 matches"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5),
	)
	pack.Pack(statusLabel, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Text widget with scrollbar.
	txtFrame := frame.New(app, "txtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tw := text.New(txtFrame, "searchtext",
		text.Width(60),
		text.Height(20),
		text.WrapModeOpt(text.WrapWord),
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

	// Configure search highlight tag.
	tw.TagConfigure("search", text.TagForeground("white"), text.TagBackground("#cc0000"))

	// Insert sample text.
	sampleText := `The quick brown fox jumps over the lazy dog.
Pack my box with five dozen liquor jugs.
How vexingly quick daft zebras jump.
The five boxing wizards jump quickly.

Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.
Duis aute irure dolor in reprehenderit in voluptate velit esse.
Excepteur sint occaecat cupidatat non proident, sunt in culpa.

Go is an open-source programming language that makes it easy
to build simple, reliable, and efficient software.
The Go programming language was designed at Google.
Go has built-in concurrency and a robust standard library.
`
	tw.Insert("1.0", sampleText)

	// Search and highlight function.
	doSearch := func() {
		query := searchEntry.GetText()
		tw.TagRemove("search", "1.0", "end")

		if query == "" {
			statusLabel.Text = "0 matches"
			statusLabel.Display()
			return
		}

		// Get all text and find matches manually.
		allText := tw.Get("1.0", "end")
		lines := strings.Split(allText, "\n")
		count := 0
		qLower := strings.ToLower(query)

		for lineNum, line := range lines {
			lineLower := strings.ToLower(line)
			pos := 0
			for {
				idx := strings.Index(lineLower[pos:], qLower)
				if idx < 0 {
					break
				}
				charStart := pos + idx
				charEnd := charStart + len(query)
				tw.TagAdd("search",
					fmt.Sprintf("%d.%d", lineNum+1, charStart),
					fmt.Sprintf("%d.%d", lineNum+1, charEnd))
				count++
				pos = charEnd
			}
		}

		statusLabel.Text = fmt.Sprintf("%d match(es)", count)
		statusLabel.Display()
	}

	// Highlight button.
	highlightBtn := button.New(searchFrame, "highlight",
		button.Text("Highlight"),
		button.Command(doSearch),
		button.PadX(8),
		button.PadY(2),
	)
	pack.Pack(highlightBtn, pack.SideOpt(pack.Left), pack.PadX(5))

	_ = searchLabel
	_ = statusLabel
	_ = highlightBtn
	_ = focusMgr
	app.Run()
}
