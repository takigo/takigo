// Demo: Text widget with embedded tags, styles, and undo/redo.
// Ported from Tk's twind.tcl demo (embedded windows/images not available).
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app := demohelper.Setup("Text Widget Features", 650, 550,
		"This demo shows text tags, colors, fonts, and undo/redo.\nThe original Tk demo also embeds windows and images.")

	// Side control buttons.
	ctrlFrame := frame.New(app, "ctrl")
	pack.Pack(ctrlFrame, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY), pack.PadX(5), pack.PadY(5))

	// Text widget with scrollbar.
	txtFrame := frame.New(app, "txtframe")
	pack.Pack(txtFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	tw := text.New(txtFrame, "tw",
		text.Width(55), text.Height(28),
		text.WrapModeOpt(text.WrapWord),
		text.UndoOpt(true),
	)

	yscroll := scrollbar.New(txtFrame, "yscroll")
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

	// Configure tags.
	tw.TagConfigure("bold", text.TagFont("Sans Bold 11"))
	tw.TagConfigure("italic", text.TagFont("Sans Italic 11"))
	tw.TagConfigure("title", text.TagFont("Sans Bold 14"))
	tw.TagConfigure("red", text.TagForeground("#cc0000"))
	tw.TagConfigure("blue", text.TagForeground("#0044cc"))
	tw.TagConfigure("green", text.TagForeground("#007700"))
	tw.TagConfigure("highlight", text.TagBackground("#ffffaa"))
	tw.TagConfigure("underline", text.TagUnderline(true))
	tw.TagConfigure("redbg", text.TagForeground("white"), text.TagBackground("#cc0000"))

	// Insert rich content.
	tw.Insert("1.0", "Text Widget Features\n")
	tw.TagAdd("title", "1.0", "1.20")

	tw.Insert("end", "\nThis text widget demonstrates various tag-based styling features available in the Tk text widget port.\n")

	tw.Insert("end", "\n1. Colored Text\n")
	tw.TagAdd("bold", "4.0", "4.16")
	tw.Insert("end", "This is red text. ")
	tw.TagAdd("red", "5.0", "5.18")
	tw.Insert("end", "This is blue text. ")
	tw.TagAdd("blue", "5.18", "5.37")
	tw.Insert("end", "This is green text.\n")
	tw.TagAdd("green", "5.37", "5.57")

	tw.Insert("end", "\n2. Font Styles\n")
	tw.TagAdd("bold", "7.0", "7.15")
	tw.Insert("end", "Bold text here. ")
	tw.TagAdd("bold", "8.0", "8.15")
	tw.Insert("end", "Italic text here. ")
	tw.TagAdd("italic", "8.15", "8.33")
	tw.Insert("end", "Underlined text here.\n")
	tw.TagAdd("underline", "8.33", "8.54")

	tw.Insert("end", "\n3. Background Colors\n")
	tw.TagAdd("bold", "10.0", "10.21")
	tw.Insert("end", "Highlighted text. ")
	tw.TagAdd("highlight", "11.0", "11.17")
	tw.Insert("end", "White on red background.\n")
	tw.TagAdd("redbg", "11.17", "11.41")

	tw.Insert("end", "\n4. Editing\n")
	tw.TagAdd("bold", "13.0", "13.10")
	tw.Insert("end", "You can edit this text freely. The undo/redo buttons on the right let you reverse changes.\n")

	tw.Insert("end", "\n5. About Embedded Content\n")
	tw.TagAdd("bold", "16.0", "16.26")
	tw.Insert("end", "The original Tk twind.tcl demo embeds buttons, other widgets, and images directly inside the text widget. Since embedded windows and images in text are not yet implemented in the Go port, this demo focuses on the tag-based rich text features that are available.\n")

	// Control buttons.
	ctrlLabel := label.New(ctrlFrame, "ctrllabel",
		label.Text("Actions:"), label.Anchor(option.AnchorW))
	pack.Pack(ctrlLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	undoBtn := button.New(ctrlFrame, "undo",
		button.Text("Undo"),
		button.Command(func() { tw.Edit("undo") }),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(undoBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	redoBtn := button.New(ctrlFrame, "redo",
		button.Text("Redo"),
		button.Command(func() { tw.Edit("redo") }),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(redoBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	boldBtn := button.New(ctrlFrame, "boldbtn",
		button.Text("Bold Sel"),
		button.Command(func() {
			sel := tw.GetSelection()
			if sel == "" {
				return
			}
			tw.TagAdd("bold", "sel.first", "sel.last")
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(boldBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	underBtn := button.New(ctrlFrame, "underbtn",
		button.Text("Underline Sel"),
		button.Command(func() {
			sel := tw.GetSelection()
			if sel == "" {
				return
			}
			tw.TagAdd("underline", "sel.first", "sel.last")
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(underBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	redBtn := button.New(ctrlFrame, "redbtn",
		button.Text("Red Sel"),
		button.Command(func() {
			sel := tw.GetSelection()
			if sel == "" {
				return
			}
			tw.TagAdd("red", "sel.first", "sel.last")
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(redBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	blueBtn := button.New(ctrlFrame, "bluebtn",
		button.Text("Blue Sel"),
		button.Command(func() {
			sel := tw.GetSelection()
			if sel == "" {
				return
			}
			tw.TagAdd("blue", "sel.first", "sel.last")
		}),
		button.PadX(8), button.PadY(3),
	)
	pack.Pack(blueBtn, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(3))

	_ = ctrlLabel
	_ = undoBtn
	_ = redoBtn
	_ = boldBtn
	_ = underBtn
	_ = redBtn
	_ = blueBtn
	app.Run()
}
