// Demo: Directory browser using TTK Treeview.
// Ported from Tk's tree.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
	"github.com/msorc/takigo/widget/scrollbar"
)

func main() {
	app := demohelper.Setup("Directory Browser", 500, 400,
		"Ttk is the new Tk themed widget set. One of the widgets it includes is a tree widget, which allows the user to browse a hierarchical data-set such as a filesystem. The tree widget not only allows for the tree part itself, but it also supports an arbitrary number of additional columns which can show additional data (in this case, the size of the files found in your filesystem). You can also change the width of the columns by dragging the boundary between them.")

	ttk.SetCurrentTheme("clam")

	// Dummy frame for grid layout of treeview + scrollbars.
	tvFrame := frame.New(app, "tvframe")
	pack.Pack(tvFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	tv := ttk.NewTreeview(tvFrame, "tree",
		ttk.TreeviewColumns("size"),
		ttk.TreeviewShow("tree", "headings"),
	)

	tv.ColumnConfigure("size", ttk.ColWidth(70), ttk.ColAnchor(option.AnchorE))
	tv.HeadingConfigure("#0", ttk.HeadText("Directory Structure"))
	tv.HeadingConfigure("size", ttk.HeadText("File Size"))

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

	// Vertical scrollbar.
	yscroll := scrollbar.New(tvFrame, "vsb",
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

	// Horizontal scrollbar (display only; X scrolling not yet implemented).
	xscroll := scrollbar.New(tvFrame, "hsb",
		scrollbar.OrientOpt(scrollbar.Horizontal),
	)
	_ = xscroll

	// Grid layout: treeview row 0 col 0, yscroll row 0 col 1, xscroll row 1 col 0.
	grid.Grid(tv, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NS))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.EW))
	grid.ColumnConfigure(tvFrame.Window(), 0, grid.SlotConfig{Weight: 1})
	grid.RowConfigure(tvFrame.Window(), 0, grid.SlotConfig{Weight: 1})

	first, last := tv.YVisibleRange()
	yscroll.Set(first, last)

	app.Run()
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d bytes", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f kB", float64(bytes)/1024)
	}
	if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
}
