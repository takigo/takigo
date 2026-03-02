// Demo: Hypertext-like tag bindings in text widget.
// Ported from Tk's bind.tcl demo (simplified).
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
	app, err := takigo.NewApp(takigo.Title("Text Tag Bindings"), takigo.Size(550, 450))
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
		label.Text("This demo shows hypertext-like tag bindings.\nColored text acts as links — visual feedback on hover."),
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

	// Status label.
	statusLabel := label.New(root, "status", app,
		label.Text("Status: Ready"),
		label.Anchor(option.AnchorW),
		label.Background("#e8e8e8"),
		label.PadX(5),
		label.PadY(2),
	)
	pack.Pack(statusLabel.Window(), pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))

	setStatus := func(s string) {
		statusLabel.Text = s
		statusLabel.Display()
	}

	// Text widget.
	txtFrame := frame.New(root, "txtframe", app)
	pack.Pack(txtFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tw := text.New(txtFrame.Window(), "hypertext", app,
		text.Width(60),
		text.Height(20),
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

	// Configure link-like tags.
	tw.TagConfigure("link", text.TagForeground("blue"), text.TagUnderline(true))
	tw.TagConfigure("heading", text.TagFont("Sans Bold 14"))

	// Insert content with tagged links.
	tw.Insert("1.0", "Tag Bindings Demo\n")
	tw.TagAdd("heading", "1.0", "1.18")

	tw.Insert("end", "\nThis demo shows text with hyperlink-style tags.\n")
	tw.Insert("end", "The following are clickable links:\n\n")

	links := []struct {
		name string
		desc string
	}{
		{"Canvas Items", "View various canvas item types"},
		{"Text Styles", "View text display styles"},
		{"Labels", "View label widget styles"},
		{"Buttons", "View button widgets"},
		{"Listboxes", "View listbox widgets"},
		{"Entries", "View entry widgets"},
	}

	for i, link := range links {
		lineNum := 5 + i*2
		tw.Insert("end", fmt.Sprintf("  %s\n", link.name))
		tw.TagAdd("link", fmt.Sprintf("%d.2", lineNum), fmt.Sprintf("%d.%d", lineNum, 2+len(link.name)))
		tw.Insert("end", fmt.Sprintf("    %s\n", link.desc))
	}

	tw.Insert("end", "\nHover over links to see visual feedback.\nClick a link to see its name in the status bar.\n")

	_ = setStatus

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
	_ = statusLabel
	app.MainLoop()
}
