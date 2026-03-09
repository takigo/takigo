// Demo: Text display styles using tags.
// Ported from Tk's style.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Display Styles"),
		takigo.Geometry("+300+300"),
		takigo.IconName("style"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Scrollbar packed right, text fills rest (matches Tcl: pack $w.scroll -side right; pack $w.text).
	yscroll := ttk.NewScrollbar(f, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {}),
	)

	tw := text.New(f, "text",
		text.Width(70),
		text.Height(32),
		text.SetGridOpt(true),
		text.WrapModeOpt(text.WrapWord),
		text.FontOpt("Courier 12"),
	)

	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}
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
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Configure display style tags matching Tcl's style.tcl.
	tw.TagConfigure("bold", text.TagFont("Courier 12 bold italic"))
	tw.TagConfigure("big", text.TagFont("Courier 14 bold"))
	tw.TagConfigure("verybig", text.TagFont("Helvetica 24 bold"))
	tw.TagConfigure("tiny", text.TagFont("Times 8 bold"))
	tw.TagConfigure("color1", text.TagBackground("#a0b7ce"))
	tw.TagConfigure("color2", text.TagForeground("red"))
	tw.TagConfigure("underline", text.TagUnderline(true))
	tw.TagConfigure("overstrike", text.TagOverstrike(true))
	tw.TagConfigure("right", text.TagJustify(option.JustifyRight))
	tw.TagConfigure("center", text.TagJustify(option.JustifyCenter))
	tw.TagConfigure("super", text.TagOffsetStr("4p"), text.TagFont("Courier 10"))
	tw.TagConfigure("sub", text.TagOffsetStr("-2p"), text.TagFont("Courier 10"))
	tw.TagConfigure("margins",
		text.TagLMargin1Str("12m"),
		text.TagLMargin2Str("6m"),
		text.TagRMarginStr("10m"),
	)
	tw.TagConfigure("spacing",
		text.TagSpacing1Str("10p"),
		text.TagSpacing2Str("2p"),
		text.TagLMargin1Str("12m"),
		text.TagLMargin2Str("6m"),
		text.TagRMarginStr("10m"),
	)
	tw.TagConfigure("raised", text.TagRelief(option.ReliefRaised), text.TagBorderWidth(1))
	tw.TagConfigure("sunken", text.TagRelief(option.ReliefSunken), text.TagBorderWidth(1))

	// Insert content matching Tcl's style.tcl (inline tag segments).
	// ins inserts text at end and optionally applies tags to the inserted range.
	ins := func(s string, tags ...string) {
		start := tw.EndIndex()
		tw.Insert("end", s)
		if len(tags) > 0 {
			end := tw.EndIndex()
			for _, tag := range tags {
				tw.TagAdd(tag, start, end)
			}
		}
	}

	ins("Text widgets like this one allow you to display information in a\nvariety of styles.  Display styles are controlled using a mechanism\ncalled ")
	ins("tags", "bold")
	ins(".  Tags are just textual names that you can apply to one\nor more ranges of characters within a text widget.  You can configure\ntags with various display styles.  If you do this, then the tagged\ncharacters will be displayed with the styles you chose.  The\navailable display styles are:\n")
	ins("\n1. Font.", "big")
	ins("  You can choose any system font, ")
	ins("large", "verybig")
	ins(" or ")
	ins("small", "tiny")
	ins(".\n")
	ins("\n2. Color.", "big")
	ins("  You can change either the ")
	ins("background", "color1")
	ins(" or ")
	ins("foreground", "color2")
	ins("\ncolor, or ")
	ins("both", "color1", "color2")
	ins(".\n")
	ins("\n3. Underlining.", "big")
	ins("  You can ")
	ins("underline", "underline")
	ins(" ranges of text.\n")
	ins("\n4. Overstrikes.", "big")
	ins("  You can ")
	ins("draw lines through", "overstrike")
	ins(" ranges of text.\n")
	ins("\n5. 3-D effects.", "big")
	ins("  You can arrange for the background to be drawn\nwith a border that makes characters appear either ")
	ins("raised", "raised")
	ins(" or ")
	ins("sunken", "sunken")
	ins(".\n")
	ins("\n6. Justification.", "big")
	ins(" You can arrange for lines to be displayed\n")
	ins("left-justified,\n")
	ins("right-justified, or\n", "right")
	ins("centered.\n", "center")
	ins("\n7. Superscripts and subscripts.", "big")
	ins(" You can control the vertical\n")
	ins("position of text to generate superscript effects like 10")
	ins("n", "super")
	ins(" or\nsubscript effects like X")
	ins("i", "sub")
	ins(".\n")
	ins("\n8. Margins.", "big")
	ins(" You can control the amount of extra space left")
	ins(" on\neach side of the text:\n")
	ins("This paragraph is an example of the use of ", "margins")
	ins("margins.  It consists of a single line of text ", "margins")
	ins("that wraps around on the screen.  There are two ", "margins")
	ins("separate left margin values, one for the first ", "margins")
	ins("display line associated with the text line, ", "margins")
	ins("and one for the subsequent display lines, which ", "margins")
	ins("occur because of wrapping.  There is also a ", "margins")
	ins("separate specification for the right margin, ", "margins")
	ins("which is used to choose wrap points for lines.\n", "margins")
	ins("\n9. Spacing.", "big")
	ins(" You can control the spacing of lines with three\n")
	ins("separate parameters.  \"Spacing1\" tells how much ")
	ins("extra space to leave\nabove a line, \"spacing3\" ")
	ins("tells how much space to leave below a line,\nand ")
	ins("if a text line wraps, \"spacing2\" tells how much ")
	ins("space to leave\nbetween the display lines that ")
	ins("make up the text line.\n")
	ins("These indented paragraphs illustrate how spacing ", "spacing")
	ins("can be used.  Each paragraph is actually a ", "spacing")
	ins("single line in the text widget, which is ", "spacing")
	ins("word-wrapped by the widget.\n", "spacing")
	ins("Spacing1 is set to 10 points for this text, ", "spacing")
	ins("which results in relatively large gaps between ", "spacing")
	ins("the paragraphs.  Spacing2 is set to 2 points, ", "spacing")
	ins("which results in just a bit of extra space ", "spacing")
	ins("within a pararaph.  Spacing3 isn't used ", "spacing")
	ins("in this example.\n", "spacing")
	ins("To see where the space is, select ranges of ", "spacing")
	ins("text within these paragraphs.  The selection ", "spacing")
	ins("highlight will cover the extra space.", "spacing")

	text.ReadOnly(true)(tw)

	app.Run()
}
