// Demo: TTK buttons, labels, and separator.
// Ported from Tk's ttkbut.tcl demo.
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
	app, err := takigo.NewApp(takigo.Title("TTK Widgets"), takigo.Size(400, 350))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	ttk.SetCurrentTheme("clam")

	root := app.Root()
	bgColor, _ := app.ColorCache().Get("#d9d9d9")
	root.BackgroundPixel = bgColor.Pixel

	msg := label.New(root, "msg", app,
		label.Text("TTK themed widgets: labels, buttons, separators.\nThey use the clam theme for modern appearance."),
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

	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// TTK Frame.
	ttkFrame := ttk.NewFrame(root, "ttkframe", app,
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	pack.Pack(ttkFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX),
		pack.PadX(15), pack.PadY(10))

	// TTK Label.
	ttkLabel := ttk.NewLabel(ttkFrame.Window(), "ttklabel", app,
		ttk.LabelText("This is a TTK Label"),
	)
	pack.Pack(ttkLabel.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(8))

	// Separator.
	sep := ttk.NewSeparator(ttkFrame.Window(), "sep", app)
	pack.Pack(sep.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(5))

	// TTK Buttons.
	for i, text := range []string{"TTK Button 1", "TTK Button 2", "TTK Button 3"} {
		idx := i + 1
		btnText := text
		btn := ttk.NewButton(ttkFrame.Window(), fmt.Sprintf("ttkbtn%d", idx), app,
			ttk.ButtonText(btnText),
			ttk.ButtonCommand(func() {
				setStatus(fmt.Sprintf("Clicked: %s", btnText))
			}),
		)
		pack.Pack(btn.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
		_ = btn
	}

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
	_ = statusLabel
	_ = ttkLabel
	_ = sep
	app.MainLoop()
}
