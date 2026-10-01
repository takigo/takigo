// Demo: Text search and highlight.
// Ported from Tk's search.tcl demo.
package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/bind"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/entry"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Search and Highlight"),
		takigo.Geometry("+300+300"),
		takigo.IconName("search"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// File name row.
	fileFrame := frame.New(f, "file")
	pack.Pack(fileFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	fileLabel := label.New(fileFrame, "label",
		label.Text("File name:"),
		label.Width(13),
		label.Anchor(option.AnchorW),
	)
	fileEntry := entry.New(fileFrame, "entry", entry.Width(40))

	// Search string row.
	searchFrame := frame.New(f, "string")
	pack.Pack(searchFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	searchLabel := label.New(searchFrame, "label",
		label.Text("Search string:"),
		label.Width(13),
		label.Anchor(option.AnchorW),
	)
	searchEntry := entry.New(searchFrame, "entry", entry.Width(40))

	// Text widget + scrollbar.
	scroll := ttk.NewScrollbar(f, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
	)

	tw := text.New(f, "text",
		text.SetGridOpt(true),
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

	// Bind Return key on entries.
	eng := app.Bind()
	eng.Bind(fileEntry.Window().PathName, "<Return>", func(ev *bind.EventData) bool {
		textLoadFile(fileEntry.GetText())
		app.FocusManager().SetFocus(searchEntry.Window())
		return false
	})
	eng.Bind(searchEntry.Window().PathName, "<Return>", func(ev *bind.EventData) bool {
		textSearch(searchEntry.GetText())
		return false
	})

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

	// Set up display styles for text highlighting (blinking toggle).
	var textToggle func()
	textToggle = func() {
		tw.TagConfigure("search", text.TagForeground("white"), text.TagBackground("#ce5555"))
		tw.Display()
		app.After(800*time.Millisecond, func() {
			tw.TagConfigure("search", text.TagForeground(""), text.TagBackground(""))
			tw.Display()
			app.After(200*time.Millisecond, textToggle)
		})
	}
	textToggle()

	// Initial text matches Tcl's description.
	tw.Insert("1.0", "This window demonstrates how to use the tagging facilities in text\nwidgets to implement a searching mechanism.  First, type a file name\nin the top entry, then type <Return> or click on \"Load File\".  Then\ntype a string in the lower entry and type <Return> or click on\n\"Load File\".  This will cause all of the instances of the string to\nbe tagged with the tag \"search\", and it will arrange for the tag's\ndisplay attributes to change to make all of the strings blink.")
	tw.MarkSet("insert", "0.0")

	// Initial focus on file entry.
	app.After(0, func() {
		app.FocusManager().SetFocus(fileEntry.Window())
	})

	_ = fileLabel
	_ = searchLabel
	_ = loadBtn
	_ = highlightBtn
	app.Run()
}
