// Demo: Text widget with embedded windows, tags, and styles.
// Ported from Tk's twind.tcl demo.
package main

import (
	"fmt"

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
		"This window demonstrates a range of text widget features including embedded windows, word wrapping, and tag-based styling.")

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
	tw.TagConfigure("center",
		text.TagJustify(option.JustifyCenter),
		text.TagSpacing1Str("5m"),
		text.TagSpacing3Str("5m"),
	)
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
	ins("For example, here are two embedded buttons — click ")
	ins("Turn On", "bold")
	ins(" to enable horizontal scrolling and ")
	ins("Turn Off", "bold")
	ins(" to disable it and restore word wrapping:\n")

	// Embedded "Turn On" button.
	turnOnLine := tw.EndIndex()
	ins("\n", "buttons")
	turnOnBtn := button.New(tw, "turnon",
		button.Text("Turn On"),
		button.Command(func() { tw.SetWrapMode(text.WrapNone) }),
		button.PadX(6), button.PadY(2),
	)
	tw.WindowCreate(turnOnLine, turnOnBtn.Window())

	// Embedded "Turn Off" button.
	turnOffLine := tw.EndIndex()
	ins("\n", "buttons")
	turnOffBtn := button.New(tw, "turnoff",
		button.Text("Turn Off"),
		button.Command(func() { tw.SetWrapMode(text.WrapWord) }),
		button.PadX(6), button.PadY(2),
	)
	tw.WindowCreate(turnOffLine, turnOffBtn.Window())

	ins("\n")

	ins("You may find it useful to put embedded windows in ")
	ins("a text without any actual text.  In this case the ")
	ins("text widget acts like a geometry manager.  For ")
	ins("example, here are buttons to change the background ")
	ins("color of the text widget:\n")

	// Color buttons embedded in the text widget.
	type colorEntry struct {
		name string
		hex  string
	}
	colors := []colorEntry{
		{"Default", "#ffffff"},
		{"AntiqueWhite3", "#cdc0b0"},
		{"Bisque1", "#ffe4c4"},
		{"SlateBlue3", "#6959cd"},
		{"RoyalBlue1", "#4169e1"},
		{"Aquamarine2", "#76eec6"},
		{"Yellow1", "#ffff00"},
		{"IndianRed1", "#ff6a6a"},
		{"SeaGreen1", "#54ff9f"},
	}
	for i, ce := range colors {
		name := fmt.Sprintf("clr%d", i)
		hex := ce.hex
		btnLine := tw.EndIndex()
		ins("\n", "buttons")
		clrBtn := button.New(tw, name,
			button.Text(ce.name),
			button.Command(func() {
				text.Background(hex)(tw)
				tw.Display()
			}),
			button.PadX(4), button.PadY(1),
		)
		tw.WindowCreate(btnLine, clrBtn.Window())
	}
	ins("\n")

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
