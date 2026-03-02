// Demo: Labelframes with text labels containing check/radio buttons.
// Ported from Tk's labelframe.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/checkbutton"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/labelframe"
	"github.com/msorc/takigo/widget/radiobutton"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Labelframe Demonstration"), takigo.Size(500, 450))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("Labelframes are used to group related widgets together.\nThe label can be positioned at different locations."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button at bottom.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Gender labelframe with radiobuttons.
	genderFrame := labelframe.New(root, "gender", app,
		labelframe.Text("Gender"),
		labelframe.Width(200),
		labelframe.Height(130),
	)
	pack.Pack(genderFrame.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10),
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
		pack.Pack(rb.Window(), pack.SideOpt(pack.Top), pack.PadY(2),
			pack.Anchor(option.AnchorW), pack.PadX(10))
		_ = rb
	}

	// Options labelframe with checkbuttons.
	optFrame := labelframe.New(root, "options", app,
		labelframe.Text("Options"),
		labelframe.Width(200),
		labelframe.Height(130),
	)
	pack.Pack(optFrame.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(10),
		pack.FillOpt(pack.FillX))

	bold := widget.NewVariable(false)
	italic := widget.NewVariable(false)
	underline := widget.NewVariable(false)

	cb1 := checkbutton.New(optFrame.Window(), "bold", app,
		checkbutton.Text("Bold"),
		checkbutton.Var(bold),
	)
	pack.Pack(cb1.Window(), pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	cb2 := checkbutton.New(optFrame.Window(), "italic", app,
		checkbutton.Text("Italic"),
		checkbutton.Var(italic),
	)
	pack.Pack(cb2.Window(), pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	cb3 := checkbutton.New(optFrame.Window(), "underline", app,
		checkbutton.Text("Underline"),
		checkbutton.Var(underline),
	)
	pack.Pack(cb3.Window(), pack.SideOpt(pack.Top), pack.PadY(2),
		pack.Anchor(option.AnchorW), pack.PadX(10))

	// Root event handlers.
	app.Dispatcher().Bind(root.XWindow, event.StructureNotifyMask, func(ev *event.Event) {
		if ev.Type == event.ConfigureType {
			root.Width = ev.ConfigWidth
			root.Height = ev.ConfigHeight
			pack.ArrangeContainer(root)
		}
	})

	app.Dispatcher().Bind(root.XWindow, event.ExposureMask, func(ev *event.Event) {
		if ev.ExposeCount > 0 {
			return
		}
		d := root.Display.XDisplay
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = cb1
	_ = cb2
	_ = cb3
	app.MainLoop()
}
