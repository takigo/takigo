// Demo: TTK Progressbar with determinate and indeterminate modes.
// Ported from Tk's ttkprogress.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/button"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	d := demohelper.Setup("Progressbar Demonstration", 400, 350, "Below are two progressbars: one determinate (showing\nprogress percentage) and one indeterminate (bouncing).")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	// Determinate progressbar.
	detLabel := label.New(root, "detlabel", app,
		label.Text("Determinate:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(detLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	detPbar := ttk.NewProgressbar(root, "detpbar", app,
		ttk.ProgressbarMode(ttk.ProgressDeterminate),
		ttk.ProgressbarLength(300),
	)
	pack.Pack(detPbar, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

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
	pack.Pack(simBtn, pack.SideOpt(pack.Top), pack.PadY(5))

	// Separator.
	sep := ttk.NewSeparator(root, "sep", app)
	pack.Pack(sep, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadX(10), pack.PadY(10))

	// Indeterminate progressbar.
	indLabel := label.New(root, "indlabel", app,
		label.Text("Indeterminate:"),
		label.Anchor(option.AnchorW),
		label.PadX(20),
	)
	pack.Pack(indLabel, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX), pack.PadY(5))

	indPbar := ttk.NewProgressbar(root, "indpbar", app,
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
		ttk.ProgressbarLength(300),
	)
	pack.Pack(indPbar, pack.SideOpt(pack.Top), pack.PadX(20), pack.PadY(5))

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
	pack.Pack(indBtn, pack.SideOpt(pack.Top), pack.PadY(5))

	_ = detLabel
	_ = simBtn
	_ = indLabel
	_ = indBtn
	_ = sep
	d.Run()
}
