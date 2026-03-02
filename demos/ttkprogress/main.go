// Demo: TTK Progressbar with determinate and indeterminate modes.
// Ported from Tk's ttkprogress.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

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
	app, err := takigo.NewApp(takigo.Title("Progressbar Demonstration"), takigo.Size(400, 350))
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
		label.Text("Below are two progressbars: one determinate (showing\nprogress percentage) and one indeterminate (bouncing)."),
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

	// Determinate progressbar.
	detLabel := label.New(root, "detlabel", app,
		label.Text("Determinate:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(detLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	detPbar := ttk.NewProgressbar(root, "detpbar", app,
		ttk.ProgressbarMode(ttk.ProgressDeterminate),
		ttk.ProgressbarLength(300),
	)
	pack.Pack(detPbar.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Progress simulation button.
	var simulating bool
	simBtn := button.New(root, "simbtn", app,
		button.Text("Start Progress"),
		button.PadX(10), button.PadY(4),
	)

	var step func()
	progress := 0.0
	step = func() {
		if !simulating {
			return
		}
		progress += 2
		if progress > 100 {
			progress = 0
		}
		detPbar.SetValue(progress)
		app.After(50*time.Millisecond, step)
	}
	simBtn.Command = func() {
		if simulating {
			simulating = false
			simBtn.Text = "Start Progress"
			simBtn.Display()
		} else {
			simulating = true
			progress = 0
			simBtn.Text = "Stop Progress"
			simBtn.Display()
			step()
		}
	}
	pack.Pack(simBtn.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

	// Separator.
	sep := ttk.NewSeparator(root, "sep", app)
	pack.Pack(sep.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(10))

	// Indeterminate progressbar.
	indLabel := label.New(root, "indlabel", app,
		label.Text("Indeterminate:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(indLabel.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	indPbar := ttk.NewProgressbar(root, "indpbar", app,
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
		ttk.ProgressbarLength(300),
	)
	pack.Pack(indPbar.Window(), pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

	// Start/Stop button for indeterminate.
	indBtn := button.New(root, "indbtn", app,
		button.Text("Start Bouncing"),
		button.PadX(10), button.PadY(4),
	)
	var bouncing bool
	indBtn.Command = func() {
		if bouncing {
			bouncing = false
			indPbar.Stop()
			indBtn.Text = "Start Bouncing"
			indBtn.Display()
		} else {
			bouncing = true
			indPbar.Start(30 * time.Millisecond)
			indBtn.Text = "Stop Bouncing"
			indBtn.Display()
		}
	}
	pack.Pack(indBtn.Window(), pack.SideOpt(pack.Top), pack.PadY(5))

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
	_ = detLabel
	_ = simBtn
	_ = indLabel
	_ = indBtn
	_ = sep
	app.MainLoop()
}
