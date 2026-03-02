// Demo: Directory browser using TTK Treeview.
// Ported from Tk's tree.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	d := demohelper.Setup("Directory Browser", 500, 400,
		"A directory browser. Click the arrows\nto expand directories.")
	root, app := d.Root, d.App

	ttk.SetCurrentTheme("clam")

	// Treeview frame with scrollbar.
	tvFrame := frame.New(root, "tvframe", app)
	pack.Pack(tvFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
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

	pack.Pack(yscroll, pack.SideOpt(pack.Right), pack.FillOpt(pack.FillY))
	pack.Pack(tv, pack.SideOpt(pack.Left), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	first, last := tv.YVisibleRange()
	yscroll.Set(first, last)

	d.Run()
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
