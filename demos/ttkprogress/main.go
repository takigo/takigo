// Demo: This demonstration script creates several progress bar widgets.
// Ported from Tk's ttkprogress.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/screenunit"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Progress Bar Demonstration"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttkprogress"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("Below are two progress bars. The top one is a \u201cdeterminate\u201d progress bar, which is used for showing how far through a defined task the program has got. The bottom one is an \u201cindeterminate\u201d progress bar, which is used to show that the program is busy but does not know how long for. Both are run here in self-animated mode, which can be turned on and off using the buttons underneath."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	ttk.SetCurrentTheme("clam")

	padX := screenunit.Px("7.5p")
	padY := screenunit.Px("3p")

	// Container frame with grid layout.
	body := ttk.NewFrame(f, "body")
	pack.Pack(body, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	detPbar := ttk.NewProgressbar(body, "p1",
		ttk.ProgressbarMode(ttk.ProgressDeterminate),
	)
	indPbar := ttk.NewProgressbar(body, "p2",
		ttk.ProgressbarMode(ttk.ProgressIndeterminate),
	)

	startBtn := ttk.NewButton(body, "start",
		ttk.ButtonText("Start Progress"),
	)
	stopBtn := ttk.NewButton(body, "stop",
		ttk.ButtonText("Stop Progress"),
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
	grid.ColumnConfigure(body, 0, grid.Weight(1))
	grid.ColumnConfigure(body, 1, grid.Weight(1))

	startBtn.Command = func() {
		detPbar.Start(50 * time.Millisecond)
		indPbar.Start(30 * time.Millisecond)
	}
	stopBtn.Command = func() {
		detPbar.Stop()
		indPbar.Stop()
	}

	app.Run()
}
