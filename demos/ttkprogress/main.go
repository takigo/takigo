// Demo: TTK Progressbar with determinate and indeterminate modes.
// Ported from Tk's ttkprogress.tcl demo.
package main

import (
	"time"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/button"
)

func main() {
	app := demohelper.Setup("Progressbar Demonstration", 400, 350, "Below are two progress bars. The top one is a \"determinate\" progress bar, which is used for showing how far through a defined task the program has got. The bottom one is an \"indeterminate\" progress bar, which is used to show that the program is busy but does not know how long for. Both are run here in self-animated mode, which can be turned on and off using the buttons underneath.")

	ttk.SetCurrentTheme("clam")

	padX := screenunit.Px("7.5p")
	padY := screenunit.Px("3p")

	// Container frame with grid layout.
	f := ttk.NewFrame(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	detPbar := ttk.NewProgressbar(f, "p1",
		ttk.ProgressbarMode(ttk.ProgressDeterminate),
	)
	indPbar := ttk.NewProgressbar(f, "p2",
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
	)

	startBtn := button.New(f, "start",
		button.Text("Start Progress"),
		button.PadX(padX), button.PadY(padY),
	)
	stopBtn := button.New(f, "stop",
		button.Text("Stop Progress"),
		button.PadX(padX), button.PadY(padY),
	)

	// Grid layout: bars span 2 columns; start sticky-e, stop sticky-w.
	grid.Grid(detPbar, grid.Row(0), grid.Column(0), grid.ColumnSpan(2),
		grid.PadX(padX), grid.PadY(padY))
	grid.Grid(indPbar, grid.Row(1), grid.Column(0), grid.ColumnSpan(2),
		grid.PadX(padX), grid.PadY(padY))
	grid.Grid(startBtn, grid.Row(2), grid.Column(0),
		grid.Sticky(grid.StickE), grid.PadX(padX), grid.PadY(padY))
	grid.Grid(stopBtn, grid.Row(2), grid.Column(1),
		grid.Sticky(grid.StickW), grid.PadX(padX), grid.PadY(padY))
	grid.ColumnConfigure(f, 0, grid.Weight(1))
	grid.ColumnConfigure(f, 1, grid.Weight(1))

	// Determinate animation state.
	var running bool
	progress := 0.0
	var step func()
	step = func() {
		if !running {
			return
		}
		progress += 2
		if progress > 100 {
			progress = 0
		}
		detPbar.SetValue(progress)
		app.After(50*time.Millisecond, step)
	}

	startBtn.Command = func() {
		if running {
			return
		}
		running = true
		progress = 0
		step()
		indPbar.Start(30 * time.Millisecond)
	}
	stopBtn.Command = func() {
		running = false
		indPbar.Stop()
	}

	_ = startBtn
	_ = stopBtn
	app.Run()
}
