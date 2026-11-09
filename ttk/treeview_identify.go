// Treeview geometry queries: IdentifyItem, IdentifyDisplayColumn,
// IdentifyRegion, BoundingBox and the identify/bbox commands of
// tk/generic/ttk/ttkTreeview.c. Coordinates are the widget's own.
package ttk

import (
	"math"
	"strconv"

	"github.com/takigo/takigo/screenunit"
	"github.com/takigo/takigo/widget"
)

// TreeRegion is a part of the treeview, as "identify region" names it.
type TreeRegion int

const (
	RegionNothing   TreeRegion = iota // outside the headings and items
	RegionHeading                     // a column heading
	RegionSeparator                   // the edge between two headings
	RegionTree                        // an item's tree column
	RegionCell                        // an item's data cell
)

func (r TreeRegion) String() string {
	return [...]string{"nothing", "heading", "separator", "tree", "cell"}[r]
}

// headingArea is tree.headingArea; zero height when headings are hidden.
func (tv *Treeview) headingArea() Box {
	return Box{treeviewFieldBorder, treeviewFieldBorder, tv.treeAreaWidth(), tv.headerOffset()}
}

// itemArea is tree.treeArea below the headings.
func (tv *Treeview) itemArea() Box {
	fb := treeviewFieldBorder
	return Box{fb, fb + tv.headerOffset(), tv.treeAreaWidth(), tv.Win.Height - 2*fb - tv.headerOffset()}
}

func (b Box) contains(x, y int) bool {
	return x >= b.X && x < b.X+b.Width && y >= b.Y && y < b.Y+b.Height
}

// halo is the heading separator's reach, HALO scaled by TkScalingLevel.
func halo() int { return int(math.Round(4 * float64(screenunit.ScalingPct()) / 100)) }

// columnAt ports IdentifyDisplayColumn: the shown column at x and the x
// of its right edge, counting the separator's halo as the column's.
func (tv *Treeview) columnAt(x int) (*TreeColumn, int) {
	h := halo()
	xpos := treeviewFieldBorder - tv.xFirst
	for _, col := range tv.shownColumns() {
		next := xpos + col.Width
		if xpos <= x && x <= next+h {
			return col, next
		}
		xpos = next
	}
	return nil, 0
}

// rowAt returns the display-list index of the row at y, or -1.
func (tv *Treeview) rowAt(y int) int {
	tv.syncDisplayList()
	area := tv.itemArea()
	if y < area.Y || tv.rowHeight <= 0 {
		return -1
	}
	if i := tv.topIndex + (y-area.Y)/tv.rowHeight; i < len(tv.displayList) {
		return i
	}
	return -1
}

// IdentifyItem returns the item at y, or "" ("identify item").
func (tv *Treeview) IdentifyItem(y int) string {
	if i := tv.rowAt(y); i >= 0 {
		return tv.displayList[i].item.ID
	}
	return ""
}

// IdentifyColumn returns the display column at x as "#n" (#0 is the tree
// column), or "" ("identify column").
func (tv *Treeview) IdentifyColumn(x int) string {
	if col, _ := tv.columnAt(x); col != nil {
		return "#" + strconv.Itoa(tv.displayNumber(col))
	}
	return ""
}

// IdentifyRegion returns the part of the treeview at x, y.
func (tv *Treeview) IdentifyRegion(x, y int) TreeRegion {
	col, x1 := tv.columnAt(x)
	if tv.showHeadings && tv.headingArea().contains(x, y) {
		switch {
		case col == nil:
			return RegionNothing
		case x1-x >= -halo() && x1-x <= halo():
			return RegionSeparator
		default:
			return RegionHeading
		}
	}
	if tv.itemArea().contains(x, y) && tv.IdentifyItem(y) != "" {
		if col != nil && col != tv.column0 {
			return RegionCell
		}
		return RegionTree
	}
	return RegionNothing
}

// indicatorMetrics is the Item style's -indicatorsize 9p (made odd) and
// -indicatormargins {1.5p 1.5p 3p 1.5p} (defaults.tcl).
func indicatorMetrics() (size, left, top, right, bottom int) {
	size = screenunit.Pt(9).Pixels()
	if size%2 == 0 {
		size--
	}
	return size, screenunit.Pt(1.5).Pixels(), screenunit.Pt(1.5).Pixels(),
		screenunit.Pt(3).Pixels(), screenunit.Pt(1.5).Pixels()
}

// cellPadding is the horizontal Treedata.padding of a data cell.
const cellPadding = 4

// itemImage is the image the tree column shows: the item's, else a tag's.
func (tv *Treeview) itemImage(item *TreeItem) widget.WidgetImage {
	if item.Image != nil {
		return item.Image
	}
	if look := tv.itemLook(item, false); look.image != nil {
		return look.image
	}
	return nil
}

// IdentifyElement returns the element of the item layout at x, y:
// Treeitem.indicator, Treeitem.image or Treeitem.text in the tree column,
// Treedata.padding or Treeitem.text in a cell; "" elsewhere.
func (tv *Treeview) IdentifyElement(x, y int) string {
	region := tv.IdentifyRegion(x, y)
	if region != RegionTree && region != RegionCell {
		return ""
	}
	id := tv.IdentifyItem(y)
	col, _ := tv.columnAt(x)
	if col == nil {
		return ""
	}
	bx, _, bw, _, ok := tv.BBox(id, "#"+strconv.Itoa(tv.displayNumber(col)))
	if !ok || x < bx || x >= bx+bw {
		return ""
	}
	if region == RegionCell {
		if x < bx+cellPadding || x >= bx+bw-cellPadding {
			return "Treedata.padding"
		}
		return "Treeitem.text"
	}
	size, ml, _, mr, _ := indicatorMetrics()
	if x < bx+ml+size+mr {
		return "Treeitem.indicator"
	}
	xi := bx + ml + size + mr
	if img := tv.itemImage(tv.items[id]); img != nil && x < xi+img.Width() {
		return "Treeitem.image"
	}
	return "Treeitem.text"
}

// BBox returns the bounding box of the item's row, or with a column ("#n",
// ID or index) of that cell, in widget coordinates; ok is false when the
// item or column is not visible (BoundingBox). The tree column's box
// starts after the item's indentation.
func (tv *Treeview) BBox(id string, column string) (x, y, width, height int, ok bool) {
	item := tv.items[id]
	if item == nil || item == tv.root || tv.detachedOrAncestor(item) {
		return 0, 0, 0, 0, false
	}
	row := tv.displayIndex(id)
	if row < 0 {
		return 0, 0, 0, 0, false
	}
	dispRow := row - tv.topIndex
	if dispRow < 0 || dispRow > tv.visibleRows() {
		return 0, 0, 0, 0, false
	}
	area := tv.itemArea()
	x, y, width, height = area.X-tv.xFirst, area.Y+dispRow*tv.rowHeight, tv.totalWidth(), tv.rowHeight
	if column == "" {
		return x, y, width, height, true
	}
	col := tv.lookupColumn(column)
	if col == nil {
		return 0, 0, 0, 0, false
	}
	found := false
	for _, c := range tv.shownColumns() {
		if c == col {
			found = true
			break
		}
		x += c.Width
	}
	if !found {
		return 0, 0, 0, 0, false
	}
	width = col.Width
	if col == tv.column0 {
		indent := tv.indent * tv.displayList[row].depth
		x += indent
		width -= indent
	}
	return x, y, width, height, true
}
