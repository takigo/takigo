// Demo: Text display styles using tags.
// Ported from Tk's style.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Demonstration - Display Styles", 600, 500,
		"This window shows text tags that control display styles. Tags are textual names applied to ranges of characters in a text widget, configuring those characters with various display styles.")

	// Scrollbar packed right, text fills rest (matches Tcl: pack $w.scroll -side right; pack $w.text).
	yscroll := scrollbar.New(app, "scroll",
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			// forward to text widget — set after tw is created
		}),
	)

	tw := text.New(app, "text",
		text.Width(70),
		text.Height(32),
		text.WrapModeOpt(text.WrapWord),
		text.FontOpt("Courier 12"),
	)

	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	// Wire scrollbar command to text widget.
	yscroll.Command = func(args ...any) {
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

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Configure display style tags matching Tcl's style.tcl.
	tw.TagConfigure("bold", text.TagFont("Courier 12 bold italic"))
	tw.TagConfigure("big", text.TagFont("Courier 14 bold"))
	tw.TagConfigure("verybig", text.TagFont("Helvetica 24 bold"))
	tw.TagConfigure("tiny", text.TagFont("Times 8 bold"))
	tw.TagConfigure("color1", text.TagBackground("#a0b7ce"))
	tw.TagConfigure("color2", text.TagForeground("red"))
	tw.TagConfigure("underline", text.TagUnderline(true))
	tw.TagConfigure("overstrike", text.TagOverstrike(true))

	// Insert text matching Tcl's style.tcl content.
	// We insert line by line and tag via "line.start line.end" notation.
	type line struct {
		text string
		tag  string
	}
	lines := []line{
		{"Text widgets like this one allow you to display information in a", ""},
		{"variety of styles.  Display styles are controlled using a mechanism", ""},
		{"called tags.  Tags are just textual names that you can apply to one", "bold"},
		{"or more ranges of characters within a text widget.  You can configure", ""},
		{"tags with various display styles.  If you do this, then the tagged", ""},
		{"characters will be displayed with the styles you chose.  The", ""},
		{"available display styles are:", ""},
		{"", ""},
		{"1. Font.", "big"},
		{"  You can choose any system font, large or small.", ""},
		{"", ""},
		{"2. Color.", "big"},
		{"  You can change the foreground or background color of text.", ""},
		{"", ""},
		{"3. Underline.", "big"},
		{"  You can underline characters in a text widget.", ""},
		{"", ""},
		{"4. Overstrike.", "big"},
		{"  You can draw lines through characters.", ""},
	}

	for i, l := range lines {
		tw.Insert("end", l.text+"\n")
		if l.tag != "" {
			lineNum := i + 1
			// Tag just the header word(s) on that line using "line.char" index format.
			startIdx := fmt.Sprintf("%d.0", lineNum)
			endIdx := fmt.Sprintf("%d.%d", lineNum, len(l.text))
			tw.TagAdd(l.tag, startIdx, endIdx)
		}
	}

	// Add additional highlighting examples.
	// "large" on line 10 chars 39-44, "small" chars 48-53.
	tw.TagConfigure("large", text.TagFont("Helvetica 24 bold"))
	tw.TagConfigure("small", text.TagFont("Times 8 bold"))
	// Colorize line 13 ("foreground" and "background" words).
	tw.TagConfigure("fgexample", text.TagForeground("red"))
	tw.TagConfigure("bgexample", text.TagBackground("#a0b7ce"))
	tw.TagAdd("fgexample", "13.25", "13.35")
	tw.TagAdd("bgexample", "13.40", "13.55")
	// Underline line 16.
	tw.TagAdd("underline", "16.8", "16.18")
	// Overstrike line 19.
	tw.TagAdd("overstrike", "19.8", "19.36")

	app.Run()
}
