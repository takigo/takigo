// Demo: Tag bindings in a text widget for hypertext-like effects.
// Ported from Tk's bind.tcl demo.
package main

import (
	"fmt"
	"github.com/msorc/takigo/font"
	"os"
	"os/exec"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Demonstration - Tag Bindings"),
		takigo.Geometry("+300+300"),
		takigo.IconName("bind"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	tw := text.New(f, "text",
		text.FontOpt(font.TkDefaultFont),
		text.Width(60),
		text.Height(24),
		text.SetGridOpt(true),
		text.WrapModeOpt(text.WrapWord),
	)

	yscroll := ttk.NewScrollbar(f, "scroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(widget.ScrollY(tw)),
	)
	tw.YScrollCmd = func(first, last float64) { yscroll.Set(first, last) }

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.Expand(true), pack.FillOpt(pack.FillBoth))

	// Insert intro paragraph.
	tw.Insert("0.0", " The same tag mechanism that controls display styles in text "+
		"widgets can also be used to associate Tcl commands with regions of "+
		"text, so that mouse or keyboard actions on the text cause "+
		"particular Tcl commands to be invoked.  For example, in the text "+
		"below the descriptions of the canvas demonstrations have been "+
		"tagged.  When you move the mouse over a demo description the "+
		"description lights up, and when you press button 1 over a "+
		"description then that particular demonstration is invoked.\n\n")

	// Demo entries: tag name, description text, demo directory.
	type demoLink struct {
		tag  string
		desc string
		dir  string
	}
	links := []demoLink{
		{"d1", "1. Samples of all the different types of items that can be created in canvas widgets.", "items"},
		{"d2", "2. A simple two-dimensional plot that allows you to adjust the positions of the data points.", "plot"},
		{"d3", "3. Anchoring and justification modes for text items.", "ctext"},
		{"d4", "4. An editor for arrow-head shapes for line items.", "arrow"},
		{"d5", "5. A ruler with facilities for editing tab stops.", "ruler"},
		{"d6", "6. A grid that demonstrates how canvases can be scrolled.", "cscroll"},
	}

	for _, link := range links {
		startIdx := tw.Doc().EndIndex()
		tw.Insert("end", link.desc)
		endIdx := tw.Doc().EndIndex()
		tw.TagAdd(link.tag, idxStr(startIdx), idxStr(endIdx))
		tw.Insert("end", "\n\n")
	}

	// Bind Enter/Leave/Button-1 for each tag.
	demosRoot := demohelper.DemoDir("")
	for _, link := range links {
		tag := link.tag
		dir := link.dir
		tw.TagBind(tag, "<Enter>", func() {
			tw.TagConfigure(tag,
				text.TagBackground("#43ce80"),
				text.TagRelief(option.ReliefRaised),
				text.TagBorderWidth(1),
			)
			tw.Display()
		})
		tw.TagBind(tag, "<Leave>", func() {
			tw.TagConfigure(tag,
				text.TagBackground(""),
				text.TagRelief(option.ReliefFlat),
			)
			tw.Display()
		})
		tw.TagBind(tag, "<Button-1>", func() {
			cmd := exec.Command("go", "run", ".")
			cmd.Dir = demosRoot + "/" + dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			// Fire-and-forget: the child demo owns its own event loop.
			// Surface Start errors so a missing demo or "go" not on PATH
			// doesn't fail silently.
			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Demo %s: failed to start: %v\n", dir, err)
			}
		})
	}

	tw.SetInsertPos("1.0")
	// Match Tcl's `-state disabled` (read-only after setup).
	tw.Configure(text.ReadOnly(true))
	app.Run()
}

// idxStr converts a text.Index to a "line.char" string for TagAdd.
func idxStr(idx text.Index) string {
	return fmt.Sprintf("%d.%d", idx.Line, idx.Char)
}
