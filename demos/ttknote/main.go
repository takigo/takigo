// Demo: TTK Notebook with multiple tabbed pages.
// Ported from Tk's ttknote.tcl demo.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/takigo/takigo"
	"github.com/takigo/takigo/demos/demohelper"
	"github.com/takigo/takigo/geometry/grid"
	"github.com/takigo/takigo/geometry/pack"
	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/clamtheme"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
	"github.com/takigo/takigo/widget"
	"github.com/takigo/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Ttk Notebook Widget"),
		takigo.Geometry("+300+300"),
		takigo.IconName("ttknote"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// pack [addSeeDismiss $w.seeDismiss $w] -side bottom -fill x
	btns := demohelper.AddSeeDismiss(app)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	// ttk::frame $w.f; pack $w.f -fill both -expand 1
	f := ttk.NewFrame(app, "f")
	pack.Pack(f, pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// ttk::notebook $w.note; pack ... -padx 1.5p -pady 3p
	nb := ttk.NewNotebook(f, "note")
	pack.Pack(nb, pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(screenunit.Pt(1.5)), pack.PadY(screenunit.Pt(3)))

	// --- Tab 1: Description ---
	// ttk::frame $w.note.msg
	page1 := ttk.NewFrame(nb, "msg")

	// ttk::label ... -wraplength 4i -justify left -anchor n
	descLabel := ttk.NewLabel(page1, "m",
		ttk.LabelText("Ttk is the new Tk themed widget set. One of the widgets it includes is the notebook widget, which provides a set of tabs that allow the selection of a group of panels, each with distinct content. They are a feature of many modern user interfaces. Not only can the tabs be selected with the mouse, but they can also be switched between using Ctrl+Tab when the notebook page heading itself is selected. Note that the second tab is disabled, and cannot be selected."),
		ttk.LabelWrapLength(screenunit.In(4)),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelAnchor(option.AnchorN),
	)
	// grid $w.note.msg.m - -sticky new -pady 1.5p
	grid.Grid(descLabel, grid.Row(0), grid.Column(0), grid.ColumnSpan(2),
		grid.Sticky(grid.StickN+grid.EW), grid.PadY(screenunit.Pt(1.5)))

	neatLabel := ttk.NewLabel(page1, "l",
		ttk.LabelText(""),
	)

	// ttk::button ... "Neat!"
	neatBtn := ttk.NewButton(page1, "b",
		ttk.ButtonText("Neat!"),
		ttk.ButtonUnderline(0),
		ttk.ButtonCommand(func() {
			neatLabel.Configure(ttk.LabelText("Yeah, I know..."))
			app.After(500*time.Millisecond, func() {
				neatLabel.Configure(ttk.LabelText(""))
			})
		}),
	)
	// grid $w.note.msg.b $w.note.msg.l -pady {1.5p 3p}
	grid.Grid(neatBtn, grid.Row(1), grid.Column(0), grid.PadYPair(screenunit.Pt(1.5), screenunit.Pt(3)))
	grid.Grid(neatLabel, grid.Row(1), grid.Column(1), grid.PadYPair(screenunit.Pt(1.5), screenunit.Pt(3)))
	grid.RowConfigure(page1, 1, grid.Weight(1))
	grid.ColumnConfigure(page1, 0, grid.Weight(1), grid.Uniform("1"))
	grid.ColumnConfigure(page1, 1, grid.Weight(1), grid.Uniform("1"))

	// $w.note add $w.note.msg -text "Description" -underline 0
	nb.Add(page1.Window(), "Description")
	nb.SetPanePadding(0, "1.5p")
	nb.SetTabUnderline(0, 0)

	// --- Tab 2: Disabled ---
	// ttk::frame $w.note.disabled
	page2 := ttk.NewFrame(nb, "disabled")
	// $w.note add $w.note.disabled -text "Disabled" -state disabled
	nb.Add(page2.Window(), "Disabled")
	nb.SetTabState(1, ttk.StateDisabled)

	// --- Tab 3: Text Editor ---
	// ttk::frame $w.note.editor
	page3 := ttk.NewFrame(nb, "editor")

	// text ... -width 40 -height 10 -wrap char -yscroll "... set"
	tw := text.New(page3, "t",
		text.Width(40),
		text.Height(10),
		text.WrapModeOpt(text.WrapChar),
	)

	// ttk::scrollbar ... -orient vertical -command "... yview"
	yscroll := ttk.NewScrollbar(page3, "s",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(tw)),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	// pack $w.note.editor.s -side right -fill y -padx {0 1.5p} -pady 1.5p
	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY),
		pack.PadXPair(0, screenunit.Pt(1.5)), pack.PadY(screenunit.Pt(1.5)))
	// pack $w.note.editor.t -fill both -expand 1 -pady 1.5p -padx {1.5p 0}
	pack.Pack(tw, pack.FillOpt(pack.FillBoth), pack.Expand(true),
		pack.PadY(screenunit.Pt(1.5)), pack.PadXPair(screenunit.Pt(1.5), 0))

	// $w.note add $w.note.editor -text "Text Editor" -underline 0
	nb.Add(page3.Window(), "Text Editor")
	nb.SetTabUnderline(2, 0)

	app.Run()
}
