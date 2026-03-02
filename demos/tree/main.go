// Demo: Directory browser using TTK Treeview.
// Ported from Tk's tree.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

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
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app, err := takigo.NewApp(takigo.Title("Directory Browser"), takigo.Size(500, 400))
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
		label.Text("A directory browser. Click the arrows\nto expand directories."),
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

	// Treeview frame with scrollbar.
	tvFrame := frame.New(root, "tvframe", app)
	pack.Pack(tvFrame.Window(), pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true), pack.PadX(10), pack.PadY(5))

	tv := ttk.NewTreeview(tvFrame.Window(), "tree", app,
		ttk.TreeviewColumns("size"),
		ttk.TreeviewShow("tree", "headings"),
	)

	tv.ColumnConfigure("size", ttk.ColWidth(100), ttk.ColAnchor(option.AnchorE))
	tv.HeadingConfigure("#0", ttk.HeadText("Name"))
	tv.HeadingConfigure("size", ttk.HeadText("Size"))

	// Populate a directory's children into the treeview.
	populateDir := func(parentID, dirPath string) {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return
		}

		// Sort: directories first, then files, both alphabetical.
		sort.Slice(entries, func(i, j int) bool {
			di, dj := entries[i].IsDir(), entries[j].IsDir()
			if di != dj {
				return di
			}
			return entries[i].Name() < entries[j].Name()
		})

		for _, entry := range entries {
			name := entry.Name()
			// Skip hidden files.
			if len(name) > 0 && name[0] == '.' {
				continue
			}
			fullPath := filepath.Join(dirPath, name)
			sizeStr := ""

			if !entry.IsDir() {
				if info, err := entry.Info(); err == nil {
					sizeStr = formatSize(info.Size())
				}
			}

			childID := tv.Insert(parentID, -1,
				ttk.ItemText(name),
				ttk.ItemValues(sizeStr),
				ttk.ItemID(fullPath),
			)

			// If directory, add a dummy child so the expand indicator shows.
			if entry.IsDir() {
				tv.Insert(childID, -1, ttk.ItemText(""), ttk.ItemID(fullPath+"/__dummy__"))
			}
		}
	}

	// On open: replace dummy child with real directory contents.
	tv.OnOpen = func(id string) {
		children := tv.Children(id)
		// Check if it's a dummy placeholder.
		if len(children) == 1 {
			child := tv.Item(children[0])
			if child != nil && child.Text == "" {
				tv.Delete(children[0])
				populateDir(id, id)
			}
		}
	}

	// Start with home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/"
	}
	populateDir("", homeDir)

	// Scrollbar.
	yscroll := scrollbar.New(tvFrame.Window(), "yscroll", app,
		scrollbar.OrientOpt(scrollbar.Vertical),
		scrollbar.CommandOpt(func(args ...any) {
			if len(args) < 1 {
				return
			}
			switch args[0] {
			case "moveto":
				if len(args) >= 2 {
					if f, ok := args[1].(float64); ok {
						tv.YViewMoveTo(f)
					}
				}
			case "scroll":
				if len(args) >= 3 {
					n, _ := args[1].(int)
					unit, _ := args[2].(string)
					tv.YViewScroll(n, unit == "pages")
				}
			}
		}),
	)

	tv.YScrollCmd = func(first, last float64) {
		yscroll.Set(first, last)
	}

	pack.Pack(yscroll.Window(), pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tv.Window(), pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := tv.YVisibleRange()
	yscroll.Set(first, last)

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
	app.MainLoop()
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}
