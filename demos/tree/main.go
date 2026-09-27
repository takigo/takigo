// Demo: Directory browser using TTK Treeview.
// Ported from Tk's tree.tcl demo.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/msorc/takigo"
	"github.com/msorc/takigo/demos/demohelper"
	"github.com/msorc/takigo/geometry/grid"
	"github.com/msorc/takigo/geometry/pack"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/ttk"
	_ "github.com/msorc/takigo/ttk/clamtheme"
	_ "github.com/msorc/takigo/ttk/defaulttheme"
	"github.com/msorc/takigo/widget/frame"
)

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
	tvFrame := ttk.NewFrame(f, "dummy")
	pack.Pack(tvFrame, pack.SideOpt(pack.Top), pack.FillOpt(pack.FillBoth),
		pack.Expand(true))

	tv := ttk.NewTreeview(tvFrame, "tree",
		ttk.TreeviewColumns("size"),
		ttk.TreeviewShow("tree", "headings"),
	)

	tv.ColumnConfigure("size", ttk.ColWidth(70), ttk.ColAnchor(option.AnchorE))
	tv.HeadingConfigure("#0", ttk.HeadText("Directory Structure"))
	tv.HeadingConfigure("size", ttk.HeadText("File Size"))

	// populateTree: the node's directory contents, sorted like lsort
	// -dictionary; directories get a "dummy" child so they can be opened.
	populateTree := func(node, path string) {
		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}
		var names []string
		for _, e := range entries {
			if !strings.HasPrefix(e.Name(), ".") { // glob * skips dotfiles
				names = append(names, e.Name())
			}
		}
		sort.Slice(names, func(i, j int) bool { return dictLess(names[i], names[j]) })
		for _, name := range names {
			f := filepath.Join(path, name)
			st, err := os.Lstat(f)
			if err != nil {
				continue
			}
			id := tv.Insert(node, -1, ttk.ItemText(name), ttk.ItemID(f),
				ttk.ItemImage(demohelper.FileIcon(f, 16)))
			switch {
			case st.IsDir():
				tv.Insert(id, 0, ttk.ItemText("dummy"), ttk.ItemID(f+"/\x00dummy"))
				tv.SetItemText(id, name+"/")
			case st.Mode().IsRegular():
				tv.SetItemValues(id, formatSize(st.Size()))
			}
		}
	}

	tv.OnOpen = func(id string) {
		if children := tv.Children(id); len(children) == 1 {
			if child := tv.Item(children[0]); child != nil && child.Text == "dummy" {
				tv.Delete(children[0])
				populateTree(id, id)
			}
		}
	}

	// populateRoots: one closed node per [file volumes] entry. Tcl 9 also
	// lists its //zipfs:/ volume, which has no counterpart here.
	root := tv.Insert("", -1, ttk.ItemText("/"), ttk.ItemID("/"),
		ttk.ItemImage(demohelper.FileIcon("/", 16)))
	populateTree(root, "/")

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

// dictLess orders like Tcl's lsort -dictionary: case-insensitive, with
// embedded numbers compared numerically, case breaking ties.
func dictLess(a, b string) bool {
	ra, rb := []rune(a), []rune(b)
	i, j := 0, 0
	for i < len(ra) && j < len(rb) {
		ca, cb := ra[i], rb[j]
		if unicode.IsDigit(ca) && unicode.IsDigit(cb) {
			si, sj := i, j
			for i < len(ra) && unicode.IsDigit(ra[i]) {
				i++
			}
			for j < len(rb) && unicode.IsDigit(rb[j]) {
				j++
			}
			na := strings.TrimLeft(string(ra[si:i]), "0")
			nb := strings.TrimLeft(string(rb[sj:j]), "0")
			if len(na) != len(nb) {
				return len(na) < len(nb)
			}
			if na != nb {
				return na < nb
			}
			continue
		}
		la, lb := unicode.ToLower(ca), unicode.ToLower(cb)
		if la != lb {
			return la < lb
		}
		i++
		j++
	}
	if len(ra)-i != len(rb)-j {
		return len(ra)-i < len(rb)-j
	}
	return a < b
}
