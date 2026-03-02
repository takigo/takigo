// Demo: Two text widgets with content synchronization.
// Ported from Tk's textpeer.tcl demo (peering simulated with copy buttons).
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
	d := demohelper.Setup("Text Peer Demonstration", 700, 500,
		"Two text widgets are shown side by side. Use the\nbuttons to copy content between them.")
	defer d.App.Destroy()
	root, app := d.Root, d.App

	noteLabel := label.New(root, "note", app,
		label.Text("Note: Tk text peering (shared document) is not implemented. Using copy buttons instead."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.Foreground("#666666"),
	)
	pack.Pack(noteLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(3))

	// Main content area.
	contentFrame := frame.New(root, "content", app)
	pack.Pack(contentFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(5), pack.PadY(5))

	// Left text + scrollbar.
	leftFrame := frame.New(contentFrame.Window(), "left", app)
	pack.Pack(leftFrame.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	leftLabel := label.New(leftFrame.Window(), "llabel", app,
		label.Text("Text A"), label.Anchor(option.AnchorW), label.PadX(5))
	pack.Pack(leftLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	leftSb := scrollbar.New(leftFrame.Window(), "lsb", app)
	leftText := text.New(leftFrame.Window(), "lefttxt", app,
		text.Width(30), text.Height(20), text.WrapModeOpt(text.WrapWord),
	)
	leftText.YScrollCmd = func(first, last float64) { leftSb.Set(first, last) }
	leftSb.Command = func(args ...interface{}) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				leftText.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				leftText.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(leftSb.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(leftText.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Center buttons.
	centerFrame := frame.New(contentFrame.Window(), "center", app)
	pack.Pack(centerFrame.Window(), pack.SideOpt(pack.Left), pack.PadX(5), pack.PadY(20))

	// Right text + scrollbar.
	rightFrame := frame.New(contentFrame.Window(), "right", app)
	pack.Pack(rightFrame.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	rightLabel := label.New(rightFrame.Window(), "rlabel", app,
		label.Text("Text B"), label.Anchor(option.AnchorW), label.PadX(5))
	pack.Pack(rightLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	rightSb := scrollbar.New(rightFrame.Window(), "rsb", app)
	rightText := text.New(rightFrame.Window(), "righttxt", app,
		text.Width(30), text.Height(20), text.WrapModeOpt(text.WrapWord),
	)
	rightText.YScrollCmd = func(first, last float64) { rightSb.Set(first, last) }
	rightSb.Command = func(args ...interface{}) {
		if len(args) >= 2 {
			action, _ := args[0].(string)
			number, _ := args[1].(float64)
			switch action {
			case "moveto":
				rightText.YViewMoveTo(number)
			case "scroll":
				unit := "units"
				if len(args) >= 3 {
					if u, ok := args[2].(string); ok {
						unit = u
					}
				}
				rightText.YViewScroll(int(number), unit == "pages")
			}
		}
	}
	pack.Pack(rightSb.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(rightText.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Copy buttons.
	copyRight := button.New(centerFrame.Window(), "copyright", app,
		button.Text("Copy -->"),
		button.Command(func() {
			content := leftText.Get("1.0", "end")
			rightText.Delete("1.0", "end")
			rightText.Insert("1.0", content)
		}),
		button.PadX(8), button.PadY(4),
	)
	pack.Pack(copyRight.Window(), pack.SideOpt(pack.Top), pack.PadY(10))

	copyLeft := button.New(centerFrame.Window(), "copyleft", app,
		button.Text("<-- Copy"),
		button.Command(func() {
			content := rightText.Get("1.0", "end")
			leftText.Delete("1.0", "end")
			leftText.Insert("1.0", content)
		}),
		button.PadX(8), button.PadY(4),
	)
	pack.Pack(copyLeft.Window(), pack.SideOpt(pack.Top), pack.PadY(10))

	// Initial content.
	leftText.Insert("1.0", `This is Text A.

In Tk, text peers share the same underlying document, so edits in one widget appear instantly in the other.

Since peering is not implemented in the Go port, you can use the copy buttons to transfer content between the two text widgets.

Try editing this text and then clicking "Copy -->" to send it to Text B.`)

	rightText.Insert("1.0", `This is Text B.

It starts with different content from Text A.

Click "<-- Copy" to replace this with the content from Text A, or edit freely and copy back.`)

	_ = noteLabel
	_ = leftLabel
	_ = rightLabel
	_ = copyRight
	_ = copyLeft
	d.Run()
}
