// Demo: Message box with selectable icon and type.
// Ported from Tk's msgbox.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
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
	app, err := takigo.NewApp(takigo.Title("Message Boxes"),
		takigo.Geometry("+300+300"),
		takigo.IconName("messagebox"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Choose the icon and type option of the message box. Then press the "+
			"\"Message Box\" button to see the message box."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f, nil)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Icon selection (left column).
	iconVar := widget.NewVariable("info")
	typeVar := widget.NewVariable("ok")

	// Left: Icon radios.
	leftFrame := frame.New(f, "left")
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.Expand(true), pack.PadX(".5c"), pack.PadY(".5c"))

	iconLabel := label.New(leftFrame, "iconlabel",
		label.Text("Icon"),
	)
	pack.Pack(iconLabel, pack.SideOpt(pack.Top))

	// Separator.
	sepLeft := frame.New(leftFrame, "sep",
		frame.Relief(option.ReliefRidge), frame.BorderWidth(1), frame.Height(2),
	)
	pack.Pack(sepLeft, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, icon := range []string{"error", "info", "question", "warning"} {
		rb := radiobutton.New(leftFrame, "icon_"+icon,
			radiobutton.Text(icon),
			radiobutton.Value(icon),
			radiobutton.Var(iconVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadY("1.5p"), pack.Anchor(option.AnchorW))
		_ = rb
	}

	// Right: Type radios.
	rightFrame := frame.New(f, "right")
	pack.Pack(rightFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.Expand(true), pack.PadX(".5c"), pack.PadY(".5c"))

	typeLabel := label.New(rightFrame, "typelabel",
		label.Text("Type"),
	)
	pack.Pack(typeLabel, pack.SideOpt(pack.Top))

	// Separator.
	sepRight := frame.New(rightFrame, "sep",
		frame.Relief(option.ReliefRidge), frame.BorderWidth(1), frame.Height(2),
	)
	pack.Pack(sepRight, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, t := range []string{"abortretryignore", "ok", "okcancel", "retrycancel", "yesno", "yesnocancel"} {
		rb := radiobutton.New(rightFrame, "type_"+t,
			radiobutton.Text(t),
			radiobutton.Value(t),
			radiobutton.Var(typeVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadY("1.5p"), pack.Anchor(option.AnchorW))
		_ = rb
	}

	// Message Box button.
	msgBtn := button.New(f, "msgbtn",
		button.Text("Message Box"),
		button.Command(func() {
			iconMap := map[string]dialog.MessageType{
				"error":    dialog.MsgError,
				"info":     dialog.MsgInfo,
				"question": dialog.MsgQuestion,
				"warning":  dialog.MsgWarning,
			}
			typeMap := map[string]dialog.ButtonSet{
				"ok":               dialog.BtnOK,
				"okcancel":         dialog.BtnOKCancel,
				"yesno":            dialog.BtnYesNo,
				"yesnocancel":      dialog.BtnYesNoCancel,
				"abortretryignore": dialog.BtnOKCancel, // closest match
				"retrycancel":      dialog.BtnOKCancel, // closest match
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
	)
	pack.Pack(msgBtn, pack.SideOpt(pack.Top), pack.PadY("1.5p"))

	_ = iconLabel
	_ = typeLabel
	_ = sepLeft
	_ = sepRight
	_ = msgBtn
	app.Run()
}
