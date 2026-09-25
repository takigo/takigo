// Demo: Directory browser using TTK Treeview.
// Ported from Tk's tree.tcl demo.
package main

import (
	"fmt"
	goimage "image"
	"image/color"
	"os"
	"path/filepath"
	"sort"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	tkimage "github.com/msorc/takigo/image"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/widget/frame"
)

// makeFolderIcon creates a 16x16 yellow folder icon.
func makeFolderIcon() *tkimage.Photo {
	img := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
	tab := color.RGBA{R: 0xd4, G: 0xaa, B: 0x00, A: 0xff}     // folder tab
	body := color.RGBA{R: 0xff, G: 0xcc, B: 0x00, A: 0xff}    // folder body
	outline := color.RGBA{R: 0x99, G: 0x77, B: 0x00, A: 0xff} // border
	// Tab: top-left 7 wide, 3 tall (rows 2-4, cols 1-7)
	for x := 1; x <= 7; x++ {
		for y := 2; y <= 4; y++ {
			img.SetRGBA(x, y, tab)
		}
	}
	// Body: rows 4-13, cols 1-14
	for x := 1; x <= 14; x++ {
		for y := 4; y <= 13; y++ {
			img.SetRGBA(x, y, body)
		}
	}
	// Outline
	for x := 1; x <= 14; x++ {
		img.SetRGBA(x, 4, outline)
		img.SetRGBA(x, 13, outline)
	}
	for y := 4; y <= 13; y++ {
		img.SetRGBA(1, y, outline)
		img.SetRGBA(14, y, outline)
	}
	for x := 1; x <= 7; x++ {
		img.SetRGBA(x, 2, outline)
	}
	img.SetRGBA(7, 3, outline)
	img.SetRGBA(8, 3, outline)
	return tkimage.NewPhoto("folder-icon", img)
}

// makeFileIcon creates a 16x16 white file icon with a folded corner.
func makeFileIcon() *tkimage.Photo {
	img := goimage.NewRGBA(goimage.Rect(0, 0, 16, 16))
	paper := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	fold := color.RGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff}
	outline := color.RGBA{R: 0x88, G: 0x88, B: 0x88, A: 0xff}
	// Body: rows 1-14, cols 2-12 (with folded corner at top-right)
	for x := 2; x <= 12; x++ {
		for y := 1; y <= 14; y++ {
			if x >= 9 && y <= 4 && (x-9)+(4-y) < 4 {
				continue // cut out corner
			}
			img.SetRGBA(x, y, paper)
		}
	}
	// Fold triangle
	for d := 0; d < 4; d++ {
		img.SetRGBA(9+d, 1+d, fold)
		img.SetRGBA(9+d, 4, fold)
		img.SetRGBA(12, 1+d, fold)
	}
	// Outline
	for y := 1; y <= 14; y++ {
		img.SetRGBA(2, y, outline)
	}
	for x := 2; x <= 12; x++ {
		img.SetRGBA(x, 14, outline)
	}
	for y := 4; y <= 14; y++ {
		img.SetRGBA(12, y, outline)
	}
	for x := 2; x <= 8; x++ {
		img.SetRGBA(x, 1, outline)
	}
	for d := 0; d <= 3; d++ {
		img.SetRGBA(9+d, d+1, outline)
	}
	return tkimage.NewPhoto("file-icon", img)
}

func main() {
	app, err := takigo.NewApp(takigo.Title("Directory Browser"),
		takigo.Geometry("+300+300"),
		takigo.IconName("tree"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	f := frame.New(app, "f")
	pack.Pack(f, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth), pack.Expand(true))

	msg := ttk.NewLabel(f, "msg",
		ttk.LabelWrapLength("4i"),
		ttk.LabelJustify(option.JustifyLeft),
		ttk.LabelAnchor(option.AnchorN),
		ttk.LabelPadding("10 2 10 6"),
		ttk.LabelText("Ttk is the new Tk themed widget set. One of the widgets it includes is a tree widget, which allows the user to browse a hierarchical data-set such as a filesystem. The tree widget not only allows for the tree part itself, but it also supports an arbitrary number of additional columns which can show additional data (in this case, the size of the files found in your filesystem). You can also change the width of the columns by dragging the boundary between them."),
	)
	pack.Pack(msg, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillX))

	btns := demohelper.AddSeeDismiss(f)
	pack.Pack(btns, pack.SideOpt(pack.Bottom), pack.FillOpt(pack.FillX))


	// Dummy frame for grid layout of treeview + scrollbars.
	tvFrame := frame.New(f, "dummy")
	pack.Pack(tvFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	tv := ttk.NewTreeview(tvFrame, "tree",
		ttk.TreeviewColumns("size"),
		ttk.TreeviewShow("tree", "headings"),
	)

	tv.ColumnConfigure("size", ttk.ColWidth(70), ttk.ColAnchor(option.AnchorE))
	tv.HeadingConfigure("#0", ttk.HeadText("Directory Structure"))
	tv.HeadingConfigure("size", ttk.HeadText("File Size"))

	// Create simple folder and file icons.
	folderIcon := makeFolderIcon()
	fileIcon := makeFileIcon()
	app.ImageRegistry().Register(folderIcon)
	app.ImageRegistry().Register(fileIcon)

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

			var icon widget.WidgetImage
			if entry.IsDir() {
				icon = folderIcon
			} else {
				icon = fileIcon
				if info, err := entry.Info(); err == nil {
					sizeStr = formatSize(info.Size())
				}
			}

			displayName := name
			if entry.IsDir() {
				displayName = name + "/"
			}

			childID := tv.Insert(parentID, -1,
				ttk.ItemText(displayName),
				ttk.ItemValues(sizeStr),
				ttk.ItemID(fullPath),
				ttk.ItemImage(icon),
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

	// Start with root directory (matches Tcl's [file volumes] on Unix).
	populateDir("", "/")

	// Vertical scrollbar.
	yscroll := ttk.NewScrollbar(tvFrame, "vsb",
		ttk.ScrollbarOrientOpt(ttk.Vertical),
		ttk.ScrollbarCommandOpt(func(args ...any) {
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
	xscroll := ttk.NewScrollbar(tvFrame, "hsb",
		ttk.ScrollbarOrientOpt(ttk.Horizontal),
	)
	_ = xscroll

	// Grid layout: treeview row 0 col 0, yscroll row 0 col 1, xscroll row 1 col 0.
	grid.Grid(tv, grid.Row(0), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.Grid(yscroll, grid.Row(0), grid.Column(1), grid.Sticky(grid.NSEW))
	grid.Grid(xscroll, grid.Row(1), grid.Column(0), grid.Sticky(grid.NSEW))
	grid.ColumnConfigure(tvFrame, 0, grid.Weight(1))
	grid.RowConfigure(tvFrame, 0, grid.Weight(1))

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
