// Demo: Text widget with embedded tags, styles, and undo/redo.
// Ported from Tk's twind.tcl demo (embedded windows/images not available).
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Demonstration - Embedded Windows and Other Features", 650, 550,
		"This window demonstrates a range of text widget features. Note: embedded windows inside the text widget are not yet supported in this port.")

	// Side control buttons.
	ctrlFrame := frame.New(app, "ctrl")
	pack.Pack(ctrlFrame, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY), pack.PadX(5), pack.PadY(5))

	// Text widget with scrollbar.
	txtFrame := frame.New(app, "txtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	tw := text.New(txtFrame, "tw",
		text.Width(70), text.Height(35),
		text.WrapModeOpt(text.WrapWord),
		text.UndoOpt(true),
		text.BorderWidthOpt(0),
	)
	tw.HighlightWidth = 0

	yscroll := ttk.NewScrollbar(txtFrame, "yscroll")
	tw.YScrollCmd = func(first, last float64) { yscroll.Set(first, last) }
	yscroll.Command = func(args ...interface{}) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				tw.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				tw.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Configure tags matching twind.tcl.
	// center: justified + spacing (used for title/header sections).
	tw.TagConfigure("center",
		text.TagJustify(option.JustifyCenter),
		text.TagSpacing1Str("5m"),
		text.TagSpacing3Str("5m"),
	)
	// buttons: margins for embedded-widget regions.
	tw.TagConfigure("buttons",
		text.TagLMargin1Str("1c"),
		text.TagLMargin2Str("1c"),
		text.TagRMarginStr("1c"),
		text.TagSpacing1Str("3m"),
		text.TagSpacing2(0),
		text.TagSpacing3(0),
	)
	tw.TagConfigure("bold", text.TagFont("Sans Bold 11"))
	tw.TagConfigure("italic", text.TagFont("Sans Italic 11"))
	tw.TagConfigure("big", text.TagFont("Sans Bold 14"))
	tw.TagConfigure("red", text.TagForeground("#cc0000"))
	tw.TagConfigure("blue", text.TagForeground("#0044cc"))
	tw.TagConfigure("underline", text.TagUnderline(true))

	// Helper: insert text and tag the inserted range.
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

	// Content matching Tcl's twind.tcl text (without embedded window markers).
	ins("Text Widget Features\n", "big", "center")
	ins("\n")

	ins("A text widget can contain many different kinds of items, ")
	ins("both active and passive.  It can lay these out in various ")
	ins("ways, with wrapping, tabs, centering, etc.  In addition, ")
	ins("when the contents are too big for the window, smooth ")
	ins("scrolling in all directions is provided.\n\n")

	ins("A text widget can contain other widgets embedded in ")
	ins("it.  These are called \"embedded windows\", ")
	ins("and they can consist of arbitrary widgets.  ")
	ins("For example, here are two embedded button widgets ")
	ins("(not available in this port): ")
	ins("[Turn On]", "bold")
	ins(" to enable horizontal scrolling and ")
	ins("[Turn Off]", "bold")
	ins(" to disable it and restore word wrapping.\n\n", "buttons")

	ins("Or, here is another example. If you ")
	ins("[Click Here]", "bold")
	ins(" a canvas displaying an x-y plot would appear right here. ")
	ins("You could drag the data points around with the mouse, ")
	ins("or click ")
	ins("[Delete]", "bold")
	ins(" to remove the plot again.\n\n")

	ins("You can also create multiple text widgets each of which ")
	ins("display the same underlying text. Click ")
	ins("[Make A Peer]", "bold")
	ins(" to make a peer widget. Notice how peer widgets can have ")
	ins("different font settings, and by default contain all the images ")
	ins("of the 'parent'.\n\n")

	ins("Users of previous versions of Tk will also be interested ")
	ins("to note that cursor movement is now by visual line by ")
	ins("default, and that all scrolling of this widget is by pixel.\n\n")

	ins("You may also find it useful to put embedded windows in ")
	ins("a text without any actual text.  In this case the ")
	ins("text widget acts like a geometry manager.  For ")
	ins("example, here is a collection of buttons laid out ")
	ins("neatly into rows by the text widget.  These buttons ")
	ins("can be used to change the background color of the ")
	ins("text widget.  The button area below has left/right margins ")
	ins("applied via the 'buttons' tag:\n")
	ins("\n[Default] [Short] [AntiqueWhite3] [Bisque1] [Bisque2]\n", "buttons")
	ins("[SlateBlue3] [RoyalBlue1] [SteelBlue2] [DeepSkyBlue3]\n", "buttons")
	ins("[DarkSlateGray1] [Aquamarine2] [DarkSeaGreen2] [SeaGreen1]\n", "buttons")
	ins("[Yellow1] [IndianRed1] [IndianRed2] [Tan1] [Tan4]\n\n", "buttons")

	ins("You can also change the usual border width and ")
	ins("highlightthickness and padding.\n")
	ins("[Big borders] [Small borders] [Big highlight] [Small highlight] [Big pad] [Small pad]\n\n", "buttons")

	ins("Finally, images fit comfortably in text widgets too:\n")
	ins("[ouster.png would appear here]\n\n", "italic")

	ins("Centered text with spacing:\n")
	ins("This line uses the 'center' tag — centered with 5m spacing above and below.\n", "center")
	ins("Back to normal left-justified text.\n")

	text.ReadOnly(true)(tw)

	// Control buttons.
	ctrlLabel := label.New(ctrlFrame, "ctrllabel",
		label.Text("Actions:"), label.Anchor(option.AnchorW))
	pack.Pack(ctrlLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	undoBtn := button.New(ctrlFrame, "undo",
		button.Text("Undo"),
		button.Command(func() {
			text.ReadOnly(false)(tw)
			tw.Edit("undo")
			text.ReadOnly(true)(tw)
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(undoBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	redoBtn := button.New(ctrlFrame, "redo",
		button.Text("Redo"),
		button.Command(func() {
			text.ReadOnly(false)(tw)
			tw.Edit("redo")
			text.ReadOnly(true)(tw)
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(redoBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	_ = ctrlLabel
	_ = undoBtn
	_ = redoBtn
	app.Run()
}
