// Demo: TTK toolbar with buttons, separator, and menubutton.
// Ported from Tk's toolbar.tcl demo (simplified — no tearoff).
package main

import (
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/label"
	"github.com/msorc/takigo/widget/menu"
)

func main() {
	d := demohelper.Setup("Toolbar Demonstration", 500, 300,
		"A toolbar with TTK buttons, a separator, and\na menubutton. This shows how themed widgets can\ncreate a modern toolbar appearance.")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5), label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Toolbar frame.
	toolbar := ttk.NewFrame(root, "toolbar", app,
		ttk.FrameBorderWidth(1),
		ttk.FrameRelief(option.ReliefRaised),
	)
	pack.Pack(toolbar.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Toolbar buttons.
	newBtn := ttk.NewButton(toolbar.Window(), "new", app,
		ttk.ButtonText("New"),
		ttk.ButtonCommand(func() { setStatus("New document") }),
	)
	pack.Pack(newBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	openBtn := ttk.NewButton(toolbar.Window(), "open", app,
		ttk.ButtonText("Open"),
		ttk.ButtonCommand(func() { setStatus("Open file...") }),
	)
	pack.Pack(openBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	saveBtn := ttk.NewButton(toolbar.Window(), "save", app,
		ttk.ButtonText("Save"),
		ttk.ButtonCommand(func() { setStatus("File saved") }),
	)
	pack.Pack(saveBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	// Vertical separator.
	sep := ttk.NewSeparator(toolbar.Window(), "sep", app, ttk.SeparatorOrient(ttk.Vertical))
	pack.Pack(sep.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(4), pack.PadY(2))

	// Undo/Redo.
	undoBtn := ttk.NewButton(toolbar.Window(), "undo", app,
		ttk.ButtonText("Undo"),
		ttk.ButtonCommand(func() { setStatus("Undo") }),
	)
	pack.Pack(undoBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	redoBtn := ttk.NewButton(toolbar.Window(), "redo", app,
		ttk.ButtonText("Redo"),
		ttk.ButtonCommand(func() { setStatus("Redo") }),
	)
	pack.Pack(redoBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	// Second separator.
	sep2 := ttk.NewSeparator(toolbar.Window(), "sep2", app, ttk.SeparatorOrient(ttk.Vertical))
	pack.Pack(sep2.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillY), pack.PadX(4), pack.PadY(2))

	// Menubutton on the toolbar.
	formatMenu := menu.New(root, "formatmenu", app)
	formatMenu.AddCommand("Bold", func() { setStatus("Format > Bold") })
	formatMenu.AddCommand("Italic", func() { setStatus("Format > Italic") })
	formatMenu.AddCommand("Underline", func() { setStatus("Format > Underline") })

	formatMB := ttk.NewMenubutton(toolbar.Window(), "format", app,
		ttk.MenubuttonText("Format"),
		ttk.MenubuttonMenu(formatMenu),
	)
	pack.Pack(formatMB.Window(), pack.SideOpt(pack.Left), pack.PadX(2), pack.PadY(2))

	_ = statusLabel
	_ = newBtn
	_ = openBtn
	_ = saveBtn
	_ = sep
	_ = undoBtn
	_ = redoBtn
	_ = sep2
	_ = formatMB
	d.Run()
}
