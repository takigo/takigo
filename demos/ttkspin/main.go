// Demo: Spinboxes with integer, float, and string values.
// Ported from Tk's ttkspin.tcl demo (adapted for classic spinbox).
package main

import (
	"fmt"
	"os"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/internal/xlib"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/spinbox"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Spinbox Demonstration"), takigo.Size(400, 350))
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
		label.Text("Three spinboxes are shown below. The first uses\nan integer range, the second uses float values,\nand the third uses a list of string values."),
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

	// Integer spinbox (0-100).
	intLabel := label.New(root, "intlabel", app,
		label.Text("Integer (0 to 100):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(intLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	intSpin := spinbox.New(root, "intspin", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(100),
		spinbox.IncrementOpt(1),
	)
	intSpin.SetText("0")
	pack.Pack(intSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Float spinbox (0-10 step 0.5).
	floatLabel := label.New(root, "floatlabel", app,
		label.Text("Float (0.0 to 10.0, step 0.5):"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(floatLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	floatSpin := spinbox.New(root, "floatspin", app,
		spinbox.FromOpt(0),
		spinbox.ToOpt(10),
		spinbox.IncrementOpt(0.5),
		spinbox.FormatOpt("%.1f"),
	)
	floatSpin.SetText("0.0")
	pack.Pack(floatSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Values spinbox (days of week).
	valLabel := label.New(root, "vallabel", app,
		label.Text("Day of week:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(valLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	valSpin := spinbox.New(root, "valspin", app,
		spinbox.ValuesOpt([]string{
			"Sunday", "Monday", "Tuesday", "Wednesday",
			"Thursday", "Friday", "Saturday",
		}),
	)
	valSpin.SetText("Sunday")
	pack.Pack(valSpin.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

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
	_ = intLabel
	_ = intSpin
	_ = floatLabel
	_ = floatSpin
	_ = valLabel
	_ = valSpin
	app.MainLoop()
}
