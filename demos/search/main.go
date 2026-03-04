// Demo: Text search and highlight.
// Ported from Tk's search.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Demonstration - Search and Highlight", 600, 500,
		"This window demonstrates how to use the tagging facilities in text "+
			"widgets to implement a search/highlight mechanism.")

	// File name row.
	fileFrame := frame.New(app, "file")
	pack.Pack(fileFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	fileLabel := label.New(fileFrame, "label", label.Text("File name:"))
	fileEntry := entry.New(fileFrame, "entry", entry.Width(40))

	// Search string row.
	searchFrame := frame.New(app, "string")
	pack.Pack(searchFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	searchLabel := label.New(searchFrame, "label", label.Text("Search string:"))
	searchEntry := entry.New(searchFrame, "entry", entry.Width(40))

	// Text widget + scrollbar.
	scroll := ttk.NewScrollbar(app, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
	)

	tw := text.New(app, "text",
		text.WrapModeOpt(text.WrapWord),
	)

	scroll.Command = func(args ...any) {
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
	}
	tw.YScrollCmd = func(first, last float64) {
		scroll.Set(first, last)
	}

	// textLoadFile loads a file into the text widget.
	textLoadFile := func(filename string) {
		data, err := os.ReadFile(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
			return
		}
		tw.Delete("1.0", "end")
		tw.Insert("end", string(data))
	}

	// textSearch finds all instances of query and tags them.
	textSearch := func(query string) {
		tw.TagRemove("search", "1.0", "end")
		if query == "" {
			return
		}
		allText := tw.Get("1.0", "end")
		lines := strings.Split(allText, "\n")
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
				pos = charEnd
			}
		}
	}

	// Wire buttons.
	loadBtn := button.New(fileFrame, "button",
		button.Text("Load File"),
		button.Command(func() {
			textLoadFile(fileEntry.GetText())
		}),
	)
	highlightBtn := button.New(searchFrame, "button",
		button.Text("Highlight"),
		button.Command(func() {
			textSearch(searchEntry.GetText())
		}),
	)

	// Pack file row.
	pack.Pack(fileLabel, pack.SideOpt(pack.Left))
	pack.Pack(fileEntry, pack.SideOpt(pack.Left))
	pack.Pack(loadBtn, pack.SideOpt(pack.Left), pack.PadY("3p"), pack.PadX("7.5p"))

	// Pack search row.
	pack.Pack(searchLabel, pack.SideOpt(pack.Left))
	pack.Pack(searchEntry, pack.SideOpt(pack.Left))
	pack.Pack(highlightBtn, pack.SideOpt(pack.Left), pack.PadY("3p"), pack.PadX("7.5p"))

	// Pack scrollbar then text.
	pack.Pack(scroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Configure search highlight tag.
	tw.TagConfigure("search", text.TagForeground("white"), text.TagBackground("#ce5555"))

	// Initial text matches Tcl's description.
	tw.Insert("1.0", "This window demonstrates how to use the tagging facilities in text\nwidgets to implement a searching mechanism.  First, type a file name\nin the top entry, then type <Return> or click on \"Load File\".  Then\ntype a string in the lower entry and type <Return> or click on\n\"Load File\".  This will cause all of the instances of the string to\nbe tagged with the tag \"search\", and it will arrange for the tag's\ndisplay attributes to change to make all of the strings blink.")

	_ = fileLabel
	_ = searchLabel
	_ = loadBtn
	_ = highlightBtn
	app.Run()
}
