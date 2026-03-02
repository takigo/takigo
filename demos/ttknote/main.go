// Demo: TTK Notebook with multiple tabbed pages.
// Ported from Tk's ttknote.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Notebook Demonstration"), takigo.Size(500, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	ttk.SetCurrentTheme("clam")

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	// Description.
	msg := label.New(root, "msg", app,
		label.Text("A notebook widget with three tabs. Click each tab\nto switch between pages."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Notebook.
	nb := ttk.NewNotebook(root, "nb", app)
	pack.Pack(nb.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(15), pack.PadY(10))

	// Tab 1: Description.
	page1 := frame.New(nb.Window(), "page1", app)
	descLabel := label.New(page1.Window(), "desc", app,
		label.Text("This is the first tab.\n\nNotebook widgets allow you to\norganize content into tabs.\nClick on a tab to view its content."),
		label.Anchor(option.AnchorNW),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(descLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))
	nb.Add(page1.Window(), "Description")

	// Tab 2: Buttons.
	page2 := frame.New(nb.Window(), "page2", app)

	statusLabel := label.New(page2.Window(), "status2", app,
		label.Text("Click a button:"),
		label.Anchor(option.AnchorW),
		label.PadX(10),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	for i, text := range []string{"Button A", "Button B", "Button C"} {
		btnText := text
		btn := ttk.NewButton(page2.Window(), fmt.Sprintf("btn%d", i), app,
			ttk.ButtonText(btnText),
			ttk.ButtonCommand(func() {
				statusLabel.Text = "Clicked: " + btnText
				statusLabel.Display()
			}),
		)
		pack.Pack(btn.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(3))
		_ = btn
	}
	nb.Add(page2.Window(), "Buttons")

	// Tab 3: Labels.
	page3 := frame.New(nb.Window(), "page3", app)
	for _, text := range []string{"Label One", "Label Two", "Label Three"} {
		l := ttk.NewLabel(page3.Window(), "l_"+text, app,
			ttk.LabelText(text),
		)
		pack.Pack(l.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
		_ = l
	}
	nb.Add(page3.Window(), "Labels")

	// Root events.
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
		d.SetForeground(root.GC, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), root.GC, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})
	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	_ = descLabel
	_ = statusLabel
	app.MainLoop()
}
