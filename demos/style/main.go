// Demo: Text display styles using tags.
// Ported from Tk's style.tcl demo.
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
	"github.com/msorc/takigo/widget/scrollbar"
	"github.com/msorc/takigo/widget/text"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Text Display Styles"), takigo.Size(550, 500))
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
		label.Text("This demo shows text tags that control display styles.\nDifferent fonts, colors, underline, and overstrike."),
		label.Anchor(option.AnchorW),
		label.PadX(10),
		label.PadY(5),
	)
	pack.Pack(msg.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	// Dismiss button.
	btnFrame := frame.New(root, "btnframe", app)
	pack.Pack(btnFrame.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX), pack.PadY(5))

	dismissBtn := button.New(btnFrame.Window(), "dismiss", app,
		button.Text("Dismiss"),
		button.Command(func() { app.Quit() }),
		button.PadX(10),
		button.PadY(4),
	)
	pack.Pack(dismissBtn.Window(), pack.SideOpt(pack.Left), pack.PadX(10))

	// Text widget with scrollbar.
	txtFrame := frame.New(root, "txtframe", app)
	pack.Pack(txtFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tw := text.New(txtFrame.Window(), "styled", app,
		text.Width(60),
		text.Height(24),
		text.WrapModeOpt(text.WrapWord),
	)

	yscroll := scrollbar.New(txtFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tw.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tw.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)
	tw.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tw.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	// Configure tags for different styles.
	tw.TagConfigure("bold", text.TagFont("Sans Bold 12"))
	tw.TagConfigure("italic", text.TagFont("Sans Italic 12"))
	tw.TagConfigure("big", text.TagFont("Sans Bold 18"))
	tw.TagConfigure("small", text.TagFont("Sans 8"))
	tw.TagConfigure("red", text.TagForeground("red"))
	tw.TagConfigure("blue", text.TagForeground("blue"))
	tw.TagConfigure("green", text.TagForeground("darkgreen"))
	tw.TagConfigure("highlight", text.TagBackground("yellow"))
	tw.TagConfigure("underline", text.TagUnderline(true))
	tw.TagConfigure("overstrike", text.TagOverstrike(true))
	tw.TagConfigure("redbg", text.TagForeground("white"), text.TagBackground("#cc0000"))

	// Helper: insert text with a tag.
	insertTagged := func(s, tag string) {
		start := tw.Get("end", "end") // Not reliable, use line counting approach
		_ = start
		tw.Insert("end", s)
		// We need to figure out the range. Use mark-based approach.
	}
	_ = insertTagged

	// Instead, insert each section and tag it by line positions.
	tw.Insert("1.0", "1. Font Styles\n")
	tw.TagAdd("big", "1.0", "1.15")

	tw.Insert("end", "\nThis is normal text.\n")
	tw.Insert("end", "This is bold text.\n")
	tw.TagAdd("bold", "4.0", "4.18")

	tw.Insert("end", "This is italic text.\n")
	tw.TagAdd("italic", "5.0", "5.20")

	tw.Insert("end", "This is small text.\n")
	tw.TagAdd("small", "6.0", "6.19")

	tw.Insert("end", "\n2. Colors\n")
	tw.TagAdd("big", "8.0", "8.9")

	tw.Insert("end", "\nRed text.\n")
	tw.TagAdd("red", "10.0", "10.9")

	tw.Insert("end", "Blue text.\n")
	tw.TagAdd("blue", "11.0", "11.10")

	tw.Insert("end", "Green text.\n")
	tw.TagAdd("green", "12.0", "12.11")

	tw.Insert("end", "Highlighted text.\n")
	tw.TagAdd("highlight", "13.0", "13.17")

	tw.Insert("end", "White on red background.\n")
	tw.TagAdd("redbg", "14.0", "14.24")

	tw.Insert("end", "\n3. Decorations\n")
	tw.TagAdd("big", "16.0", "16.14")

	tw.Insert("end", "\nUnderlined text.\n")
	tw.TagAdd("underline", "18.0", "18.16")

	tw.Insert("end", "Overstrike text.\n")
	tw.TagAdd("overstrike", "19.0", "19.16")

	tw.Insert("end", "\n4. Combined Styles\n")
	tw.TagAdd("big", "21.0", "21.18")

	tw.Insert("end", "\nBold and red.\n")
	tw.TagAdd("bold", "23.0", "23.13")
	tw.TagAdd("red", "23.0", "23.13")

	tw.Insert("end", "Italic, blue, and underlined.\n")
	tw.TagAdd("italic", "24.0", "24.29")
	tw.TagAdd("blue", "24.0", "24.29")
	tw.TagAdd("underline", "24.0", "24.29")

	// Root event handlers.
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
		gc := root.GC
		d.SetForeground(gc, bgColor.Pixel)
		d.FillRectangle(root.Drawable(), gc, 0, 0, uint(root.Width), uint(root.Height))
		d.Flush()
	})

	app.Dispatcher().BindGlobal(event.KeyPressMask, func(ev *event.Event) {
		if ev.KeySym == xlib.XK_Escape {
			app.Quit()
		}
	})

	_ = msg
	_ = dismissBtn
	app.MainLoop()
}
