// Demo: Message boxes of various type.
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
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Message Box Demonstration"),
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

	// Icon and type variables.
	iconVar := widget.NewVariable("info")
	typeVar := widget.NewVariable("ok")

	btns := demohelper.AddBottomButtons(f, func(bf *ttk.Frame) *ttk.Button {
		return ttk.NewButton(bf, "vars",
			ttk.ButtonText("Message Box"),
			ttk.ButtonCommand(func() { showMessageBox(app, iconVar, typeVar) }),
		)
	})
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// Left: Icon radios.
	leftFrame := frame.New(f, "left")
	pack.Pack(leftFrame, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY),
		pack.Expand(true), pack.PadX(".5c"), pack.PadY(".5c"))

	iconLabel := label.New(leftFrame, "label",
		label.Text("Icon"),
	)
	pack.Pack(iconLabel, pack.SideOpt(pack.Top))

	// Separator.
	sepLeft := frame.New(leftFrame, "sep",
		frame.Relief(option.ReliefRidge), frame.BorderWidth(1), frame.Height(2),
	)
	pack.Pack(sepLeft, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, icon := range []string{"error", "info", "question", "warning"} {
		rb := radiobutton.New(leftFrame, "b"+icon,
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

	typeLabel := label.New(rightFrame, "label",
		label.Text("Type"),
	)
	pack.Pack(typeLabel, pack.SideOpt(pack.Top))

	// Separator.
	sepRight := frame.New(rightFrame, "sep",
		frame.Relief(option.ReliefRidge), frame.BorderWidth(1), frame.Height(2),
	)
	pack.Pack(sepRight, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	for _, t := range []string{"abortretryignore", "ok", "okcancel", "retrycancel", "yesno", "yesnocancel"} {
		rb := radiobutton.New(rightFrame, t,
			radiobutton.Text(t),
			radiobutton.Value(t),
			radiobutton.Var(typeVar),
			radiobutton.Anchor(option.AnchorW),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
			pack.PadY("1.5p"), pack.Anchor(option.AnchorW))
		_ = rb
	}

	_ = iconLabel
	_ = typeLabel
	_ = sepLeft
	_ = sepRight
	app.Run()
}

func showMessageBox(app *takigo.App, iconVar, typeVar *widget.Variable[string]) {
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

	msgText := fmt.Sprintf("This is a \"%s\" type messagebox with the \"%s\" icon", typ, icon)
	result := dialog.ShowMessage(app,
		dialog.MsgTitle("Message"),
		dialog.MsgMessage(msgText),
		dialog.MsgType(iconMap[icon]),
		dialog.MsgButtons(typeMap[typ]),
	)

	// Map result to button name.
	resultName := map[dialog.DialogResult]string{
		dialog.ResultOK:     "ok",
		dialog.ResultCancel: "cancel",
		dialog.ResultYes:    "yes",
		dialog.ResultNo:     "no",
		dialog.ResultAbort:  "abort",
		dialog.ResultRetry:  "retry",
		dialog.ResultIgnore: "ignore",
	}
	name := resultName[result]
	if name == "" {
		name = "unknown"
	}

	// Show result in a follow-up info dialog.
	dialog.ShowMessage(app,
		dialog.MsgMessage(fmt.Sprintf("You have selected \"%s\"", name)),
		dialog.MsgType(dialog.MsgInfo),
		dialog.MsgButtons(dialog.BtnOK),
	)
}
