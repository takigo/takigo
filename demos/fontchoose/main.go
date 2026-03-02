// Demo: Font chooser dialog.
// Ported from Tk's fontchoose.tcl demo.
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/dialog"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Font Chooser"), takigo.Size(450, 250))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("Click the button to open the font chooser.\nThe selected font description is shown below."),
		label.Anchor(option.AnchorW),
		label.PadX(10), label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))
	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"), button.Command(func() { app.Quit() }),
		button.PadX(10), button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Font display label.
	fontLabel := label.New(root, "fontlabel", app,
		label.Text("Selected: (none)"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(fontLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(10))

	// Preview label.
	previewLabel := label.New(root, "preview", app,
		label.Text("The quick brown fox jumps over the lazy dog."),
		label.PadX(10), label.PadY(10),
	)
	pack.Pack(previewLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX),
		pack.PadX(20), pack.PadY(5))

	chooseBtn := button.New(root, "choose", app,
		button.Text("Choose Font..."),
		button.Command(func() {
			fontDesc, ok := dialog.ChooseFont(app,
				dialog.FontParent(root),
			)
			if ok {
				fontLabel.Text = fmt.Sprintf("Selected: %s", fontDesc)
				fontLabel.Display()
			}
		}),
		button.PadX(15), button.PadY(8),
	)
	pack.Pack(chooseBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(30), pack.PadY(20))

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
	_ = fontLabel
	_ = previewLabel
	_ = chooseBtn
	app.MainLoop()
}
