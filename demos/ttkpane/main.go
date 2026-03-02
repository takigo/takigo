// Demo: TTK frame with paned windows.
// Ported from Tk's ttkpane.tcl demo.
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
	"github.com/msorc/takigo/widget/panedwindow"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("TTK + Paned Windows"), takigo.Size(500, 350))
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
		label.Text("TTK frames inside a paned window.\nDrag the sash to resize."),
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

	pw := panedwindow.New(root, "pw", app,
		panedwindow.OrientOpt(panedwindow.Horizontal),
	)
	pack.Pack(pw.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	// Left pane: TTK frame with label.
	leftFrame := ttk.NewFrame(pw.Window(), "left", app,
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	leftLabel := ttk.NewLabel(leftFrame.Window(), "leftlbl", app,
		ttk.LabelText("TTK Frame (Left Pane)"),
	)
	pack.Pack(leftLabel.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))
	pw.Add(leftFrame.Window(), 150)

	// Right pane: TTK frame with button.
	rightFrame := ttk.NewFrame(pw.Window(), "right", app,
		ttk.FrameBorderWidth(2),
		ttk.FrameRelief(option.ReliefGroove),
	)
	rightLabel := ttk.NewLabel(rightFrame.Window(), "rightlbl", app,
		ttk.LabelText("TTK Frame (Right Pane)"),
	)
	pack.Pack(rightLabel.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(10))

	ttkBtn := ttk.NewButton(rightFrame.Window(), "ttkbtn", app,
		ttk.ButtonText("TTK Button"),
		ttk.ButtonCommand(func() { fmt.Println("TTK button clicked!") }),
	)
	pack.Pack(ttkBtn.Window(), pack.SideOpt(pack.Top), pack.PadX(10), pack.PadY(5))
	pw.Add(rightFrame.Window(), 150)

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
	_ = leftLabel
	_ = rightLabel
	_ = ttkBtn
	app.MainLoop()
}
