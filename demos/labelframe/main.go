// Demo: Labelframes with text labels containing check/radio buttons.
// Ported from Tk's labelframe.tcl demo.
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	d := demohelper.Setup("Labelframe Demonstration", 500, 450, "Labelframes are used to group related widgets together.\nThe label can be positioned at different locations.")
	root, app := d.Root, d.App

	// Gender labelframe with radiobuttons.
	genderFrame := labelframe.New(root, "gender", app,
		labelframe.Text("Gender"),
		labelframe.Width(200),
		labelframe.Height(130),
	)
	pack.Pack(genderFrame, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10),
		pack.FillOpt(pack.FillX))

	genderVar := widget.NewVariable("male")
	genders := []struct{ text, value string }{
		{"Male", "male"},
		{"Female", "female"},
		{"Other", "other"},
	}
	for _, g := range genders {
		rb := radiobutton.New(genderFrame.Window(), "gender_"+g.value, app,
			radiobutton.Text(g.text),
			radiobutton.Value(g.value),
			radiobutton.Var(genderVar),
		)
		pack.Pack(rb, pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	// Options labelframe with checkbuttons.
	optFrame := labelframe.New(root, "options", app,
		labelframe.Text("Options"),
		labelframe.Width(200),
		labelframe.Height(130),
	)
	pack.Pack(optFrame, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10),
		pack.FillOpt(pack.FillX))

	bold := widget.NewVariable(false)
	italic := widget.NewVariable(false)
	underline := widget.NewVariable(false)

	cb1 := checkbutton.New(optFrame.Window(), "bold", app,
		checkbutton.Text("Bold"),
		checkbutton.Var(bold),
	)
	pack.Pack(cb1, pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	cb2 := checkbutton.New(optFrame.Window(), "italic", app,
		checkbutton.Text("Italic"),
		checkbutton.Var(italic),
	)
	pack.Pack(cb2, pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	cb3 := checkbutton.New(optFrame.Window(), "underline", app,
		checkbutton.Text("Underline"),
		checkbutton.Var(underline),
	)
	pack.Pack(cb3, pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	_ = cb1
	_ = cb2
	_ = cb3
	d.Run()
}
