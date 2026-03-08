// Demo: Hypertext-like tag bindings in text widget.
// Ported from Tk's bind.tcl demo.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/ttk"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/label"
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

	msg := label.New(f, "msg",
		label.WrapLength("4i"),
		label.JustifyOpt(option.JustifyLeft),
		label.Text("The same tag mechanism that controls display styles in text widgets can also be\n"+
			"used to associate commands with regions of text, so that mouse or keyboard actions\n"+
			"on the text cause particular actions to be invoked. In the text below the\n"+
			"descriptions of the canvas demonstrations have been tagged. When you move the\n"+
			"mouse over a demo description the description lights up, and when you press\n"+
			"button 1 over a description then that particular demonstration is invoked."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	tw := text.New(f, "hypertext",
		text.Width(60),
		text.Height(24),
		text.WrapModeOpt(text.WrapWord),
	)

	yscroll := ttk.NewScrollbar(f, "yscroll",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if f, ok := args[1].(float64); ok {
					tw.YViewMoveTo(f)
				}
			case "scroll":
				n, _ := args[1].(int)
				unit, _ := args[2].(string)
				tw.YViewScroll(n, unit == "pages")
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) { yscroll.Set(first, last) }

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

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
			tw.TagConfigure(tag, text.TagBackground("#43ce80"))
			tw.Display()
		})
		tw.TagBind(tag, "<Leave>", func() {
			tw.TagConfigure(tag, text.TagBackground(""))
			tw.Display()
		})
		tw.TagBind(tag, "<Button-1>", func() {
			go func() {
				cmd := exec.Command("go", "run", ".")
				cmd.Dir = demosRoot + "/" + dir
				_ = cmd.Start()
			}()
		})
	}

	tw.SetInsertPos("1.0")
	// Match Tcl's `-state disabled` (read-only after setup).
	text.ReadOnly(true)(tw)
	// Set initial focus to text widget.
	app.After(0, func() {
		app.Server().SetInputFocus(tw.Window().PlatformID, platform.RevertToParent, platform.CurrentTime)
	})
	app.Run()
}

// idxStr converts a text.Index to a "line.char" string for TagAdd.
func idxStr(idx text.Index) string {
	return fmt.Sprintf("%d.%d", idx.Line, idx.Char)
}

// DemoDir returns the absolute path to a demo directory by name,
// relative to the demos/ root found via the caller's source file location.
func DemoDir(name string) string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return name
	}
	// Walk up from callers' source file to find the demos/ root.
	// demos/demohelper/demohelper.go → demos/ is one level up.
	demosRoot := filepath.Dir(filepath.Dir(file))
	return filepath.Join(demosRoot, name)
}
