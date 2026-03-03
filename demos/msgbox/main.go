// Demo: Message box with selectable icon and type.
// Ported from Tk's msgbox.tcl demo.
package main

import (
	"fmt"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app := demohelper.Setup("Message Boxes", 500, 400,
		"Choose the icon and type option of the message box. Then press the "+
			"\"Message Box\" button to see the message box.")

	// Icon selection (left column).
	iconVar := widget.NewVariable("info")
	typeVar := widget.NewVariable("ok")

	columns := frame.New(app, "columns")
	pack.Pack(columns, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Left: Icon radios.
	leftFrame := frame.New(columns, "left")
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.Expand(true), pack.PadX(10))

	iconLabel := label.New(leftFrame, "iconlabel",
		label.Text("Icon"),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(iconLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, icon := range []string{"error", "info", "question", "warning"} {
		rb := radiobutton.New(leftFrame, "icon_"+icon,
			radiobutton.Text(icon),
			radiobutton.Value(icon),
			radiobutton.Var(iconVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))
		_ = rb
	}

	// Right: Type radios.
	rightFrame := frame.New(columns, "right")
	pack.Pack(rightFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.Expand(true), pack.PadX(10))

	typeLabel := label.New(rightFrame, "typelabel",
		label.Text("Type"),
		label.Anchor(option.AnchorCenter),
	)
	pack.Pack(typeLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, t := range []string{"ok", "okcancel", "yesno", "yesnocancel"} {
		rb := radiobutton.New(rightFrame, "type_"+t,
			radiobutton.Text(t),
			radiobutton.Value(t),
			radiobutton.Var(typeVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(2))
		_ = rb
	}

	// Message Box button.
	msgBtn := button.New(app, "msgbtn",
		button.Text("Message Box"),
		button.Command(func() {
			iconMap := map[string]dialog.MessageType{
				"error":    dialog.MsgError,
				"info":     dialog.MsgInfo,
				"question": dialog.MsgQuestion,
				"warning":  dialog.MsgWarning,
			}
			typeMap := map[string]dialog.ButtonSet{
				"ok":             dialog.BtnOK,
				"okcancel":       dialog.BtnOKCancel,
				"yesno":          dialog.BtnYesNo,
				"yesnocancel":    dialog.BtnYesNoCancel,
			}

			icon := iconVar.Get()
			typ := typeVar.Get()

			msg := fmt.Sprintf("This is a %q type messagebox with the %q icon.", typ, icon)
			result := dialog.ShowMessage(app,
				dialog.MsgTitle("Message"),
				dialog.MsgMessage(msg),
				dialog.MsgType(iconMap[icon]),
				dialog.MsgButtons(typeMap[typ]),
			)

			// Show result in a follow-up info dialog.
			dialog.ShowMessage(app,
				dialog.MsgTitle("Result"),
				dialog.MsgMessage(fmt.Sprintf("You pressed button %d.", result)),
				dialog.MsgType(dialog.MsgInfo),
				dialog.MsgButtons(dialog.BtnOK),
			)
		}),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(msgBtn, pack.SideOpt(pack.Top), pack.PadY(10))

	_ = iconLabel
	_ = typeLabel
	_ = columns
	_ = msgBtn
	app.Run()
}
