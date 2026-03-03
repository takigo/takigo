// Package grid implements the grid geometry manager, which arranges
// children in a two-dimensional table. It ports tk/generic/tkGrid.c.
package grid

import (
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/window"
)

// Sticky flags for positioning within a cell.
const (
	StickN = 1 << iota
	StickE
	StickS
	StickW

	NSEW = StickN | StickS | StickE | StickW
	NS   = StickN | StickS
	EW   = StickE | StickW
)

// GridOption configures a Grid call.
type GridOption func(*gridConfig)

type gridConfig struct {
	row        int
	column     int
	rowSpan    int
	columnSpan int
	sticky     int
	padX       int
	padY       int
	iPadX      int
	iPadY      int
}

// Row sets the row.
func Row(r int) GridOption { return func(c *gridConfig) { c.row = r } }

// Column sets the column.
func Column(col int) GridOption { return func(c *gridConfig) { c.column = col } }

// RowSpan sets the number of rows to span.
func RowSpan(n int) GridOption { return func(c *gridConfig) { c.rowSpan = n } }

// ColumnSpan sets the number of columns to span.
func ColumnSpan(n int) GridOption { return func(c *gridConfig) { c.columnSpan = n } }

// Sticky sets the sticky flags.
func Sticky(s int) GridOption { return func(c *gridConfig) { c.sticky = s } }

// PadX sets the exterior horizontal padding.
func PadX(p int) GridOption { return func(c *gridConfig) { c.padX = p } }

// PadY sets the exterior vertical padding.
func PadY(p int) GridOption { return func(c *gridConfig) { c.padY = p } }

// IPadX sets the interior horizontal padding.
func IPadX(p int) GridOption { return func(c *gridConfig) { c.iPadX = p } }

// IPadY sets the interior vertical padding.
func IPadY(p int) GridOption { return func(c *gridConfig) { c.iPadY = p } }

// SlotConfig holds configuration for a row or column.
type SlotConfig struct {
	MinSize int
	Weight  int
	Pad     int
}

// gridEntry holds grid configuration for a single child.
type gridEntry struct {
	window *window.Window
	config gridConfig
}

// gridder manages grid state for a container.
type gridder struct {
	container *window.Window
	entries   []*gridEntry
	rowConf   map[int]*SlotConfig
	colConf   map[int]*SlotConfig
}

// singleton manager instance.
var mgr = &gridManager{}

// gridders tracks per-container state.
var gridders = map[*window.Window]*gridder{}

type gridManager struct{}

func (m *gridManager) Name() string { return "grid" }

func (m *gridManager) RequestProc(content *window.Window) {
	parent := content.Parent
	if parent == nil {
		return
	}
	if g, ok := gridders[parent]; ok {
		g.arrange()
	}
}

func (m *gridManager) LostContentProc(content *window.Window) {
	parent := content.Parent
	if parent == nil {
		return
	}
	if g, ok := gridders[parent]; ok {
		g.remove(content)
	}
}

// Grid adds a child to its parent's grid layout.
func Grid(child window.Windower, opts ...GridOption) {
	w := child.Window()
	parent := w.Parent
	if parent == nil {
		return
	}

	cfg := gridConfig{
		rowSpan:    1,
		columnSpan: 1,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.rowSpan < 1 {
		cfg.rowSpan = 1
	}
	if cfg.columnSpan < 1 {
		cfg.columnSpan = 1
	}

	geometry.ManageGeometry(w, mgr)

	g, ok := gridders[parent]
	if !ok {
		g = &gridder{
			container: parent,
			rowConf:   make(map[int]*SlotConfig),
			colConf:   make(map[int]*SlotConfig),
		}
		gridders[parent] = g
	}

	// Auto-assign row if not specified.
	if cfg.row == 0 && cfg.column == 0 {
		cfg.row = g.nextRow()
	}

	// Update or add entry.
	for _, e := range g.entries {
		if e.window == w {
			e.config = cfg
			g.arrange()
			return
		}
	}

	g.entries = append(g.entries, &gridEntry{window: w, config: cfg})
	g.arrange()
}

// Forget removes a child from grid management.
func Forget(child window.Windower) {
	w := child.Window()
	parent := w.Parent
	if parent == nil {
		return
	}
	if g, ok := gridders[parent]; ok {
		g.remove(w)
	}
	w.GeomManager = nil
}

// RowConfigure sets configuration for a row.
func RowConfigure(container *window.Window, row int, conf SlotConfig) {
	g, ok := gridders[container]
	if !ok {
		g = &gridder{
			container: container,
			rowConf:   make(map[int]*SlotConfig),
			colConf:   make(map[int]*SlotConfig),
		}
		gridders[container] = g
	}
	g.rowConf[row] = &conf
	g.arrange()
}

// ColumnConfigure sets configuration for a column.
func ColumnConfigure(container *window.Window, col int, conf SlotConfig) {
	g, ok := gridders[container]
	if !ok {
		g = &gridder{
			container: container,
			rowConf:   make(map[int]*SlotConfig),
			colConf:   make(map[int]*SlotConfig),
		}
		gridders[container] = g
	}
	g.colConf[col] = &conf
	g.arrange()
}

func (g *gridder) remove(child *window.Window) {
	for i, e := range g.entries {
		if e.window == child {
			g.entries = append(g.entries[:i], g.entries[i+1:]...)
			break
		}
	}
	if len(g.entries) == 0 {
		delete(gridders, g.container)
	}
}

func (g *gridder) nextRow() int {
	maxRow := 0
	for _, e := range g.entries {
		end := e.config.row + e.config.rowSpan
		if end > maxRow {
			maxRow = end
		}
	}
	return maxRow
}

// arrange performs the grid layout.
func (g *gridder) arrange() {
	container := g.container
	if container.PlatformID == platform.WindowID(0) || len(g.entries) == 0 {
		return
	}

	// Find grid dimensions.
	maxCol, maxRow := 0, 0
	for _, e := range g.entries {
		endCol := e.config.column + e.config.columnSpan
		endRow := e.config.row + e.config.rowSpan
		if endCol > maxCol {
			maxCol = endCol
		}
		if endRow > maxRow {
			maxRow = endRow
		}
	}

	if maxCol == 0 || maxRow == 0 {
		return
	}

	// Compute minimum column widths and row heights.
	colWidths := make([]int, maxCol)
	rowHeights := make([]int, maxRow)

	for _, e := range g.entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth
		childW := child.ReqWidth + bw2 + cfg.iPadX*2 + cfg.padX*2
		childH := child.ReqHeight + bw2 + cfg.iPadY*2 + cfg.padY*2

		// For span=1 items, contribute directly.
		if cfg.columnSpan == 1 {
			if childW > colWidths[cfg.column] {
				colWidths[cfg.column] = childW
			}
		}
		if cfg.rowSpan == 1 {
			if childH > rowHeights[cfg.row] {
				rowHeights[cfg.row] = childH
			}
		}
	}

	// Apply slot configs (minimum sizes).
	for col, conf := range g.colConf {
		if col < maxCol && conf.MinSize > colWidths[col] {
			colWidths[col] = conf.MinSize
		}
	}
	for row, conf := range g.rowConf {
		if row < maxRow && conf.MinSize > rowHeights[row] {
			rowHeights[row] = conf.MinSize
		}
	}

	// Handle spanning widgets (distribute across spanned slots).
	for _, e := range g.entries {
		cfg := &e.config
		if cfg.columnSpan > 1 {
			child := e.window
			bw2 := 2 * child.BorderWidth
			needed := child.ReqWidth + bw2 + cfg.iPadX*2 + cfg.padX*2
			current := 0
			for c := cfg.column; c < cfg.column+cfg.columnSpan && c < maxCol; c++ {
				current += colWidths[c]
			}
			if needed > current {
				extra := needed - current
				perCol := extra / cfg.columnSpan
				for c := cfg.column; c < cfg.column+cfg.columnSpan && c < maxCol; c++ {
					colWidths[c] += perCol
				}
			}
		}
		if cfg.rowSpan > 1 {
			child := e.window
			bw2 := 2 * child.BorderWidth
			needed := child.ReqHeight + bw2 + cfg.iPadY*2 + cfg.padY*2
			current := 0
			for r := cfg.row; r < cfg.row+cfg.rowSpan && r < maxRow; r++ {
				current += rowHeights[r]
			}
			if needed > current {
				extra := needed - current
				perRow := extra / cfg.rowSpan
				for r := cfg.row; r < cfg.row+cfg.rowSpan && r < maxRow; r++ {
					rowHeights[r] += perRow
				}
			}
		}
	}

	// Compute total natural size.
	totalW, totalH := 0, 0
	for _, w := range colWidths {
		totalW += w
	}
	for _, h := range rowHeights {
		totalH += h
	}

	reqW := totalW + container.InternalBorderLeft + container.InternalBorderRight
	reqH := totalH + container.InternalBorderTop + container.InternalBorderBottom

	if !container.IsTopLevel() {
		geometry.GeometryRequest(container, reqW, reqH)
	}

	// Distribute extra space via weights.
	availW := container.Width - container.InternalBorderLeft - container.InternalBorderRight
	availH := container.Height - container.InternalBorderTop - container.InternalBorderBottom
	distributeExtra(colWidths, g.colConf, availW-totalW)
	distributeExtra(rowHeights, g.rowConf, availH-totalH)

	// Compute offsets.
	colOffsets := make([]int, maxCol+1)
	colOffsets[0] = container.InternalBorderLeft
	for c := range maxCol {
		colOffsets[c+1] = colOffsets[c] + colWidths[c]
	}

	rowOffsets := make([]int, maxRow+1)
	rowOffsets[0] = container.InternalBorderTop
	for r := range maxRow {
		rowOffsets[r+1] = rowOffsets[r] + rowHeights[r]
	}

	// Position children.
	for _, e := range g.entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth

		cavX := colOffsets[cfg.column] + cfg.padX
		cavY := rowOffsets[cfg.row] + cfg.padY
		cavW := colOffsets[min(cfg.column+cfg.columnSpan, maxCol)] - colOffsets[cfg.column] - cfg.padX*2
		cavH := rowOffsets[min(cfg.row+cfg.rowSpan, maxRow)] - rowOffsets[cfg.row] - cfg.padY*2

		childW := child.ReqWidth + bw2 + cfg.iPadX*2
		childH := child.ReqHeight + bw2 + cfg.iPadY*2

		// Apply sticky.
		x, y, w, h := applySticky(cfg.sticky, cavX, cavY, cavW, cavH, childW, childH)

		child.X = x
		child.Y = y
		child.Width = w - bw2
		child.Height = h - bw2
		if child.Width < 1 {
			child.Width = 1
		}
		if child.Height < 1 {
			child.Height = 1
		}

		if child.PlatformID != platform.WindowID(0) {
			container.Display.Server.MoveResizeWindow(child.PlatformID,
				child.X, child.Y, uint(child.Width), uint(child.Height))
			if child.Flags&window.FlagMapped == 0 {
				container.Display.Server.MapWindow(child.PlatformID)
				child.Flags |= window.FlagMapped
			}
		}
	}
}

// distributeExtra distributes extra space among weighted slots.
func distributeExtra(sizes []int, conf map[int]*SlotConfig, extra int) {
	if extra <= 0 {
		return
	}

	totalWeight := 0
	for i := range sizes {
		if c, ok := conf[i]; ok && c.Weight > 0 {
			totalWeight += c.Weight
		}
	}

	if totalWeight == 0 {
		return
	}

	cumWeight := 0
	distributed := 0
	for i := range sizes {
		if c, ok := conf[i]; ok && c.Weight > 0 {
			cumWeight += c.Weight
			newDist := extra * cumWeight / totalWeight
			sizes[i] += newDist - distributed
			distributed = newDist
		}
	}
}

// applySticky positions a child within its cavity based on sticky flags.
func applySticky(sticky, cavX, cavY, cavW, cavH, childW, childH int) (x, y, w, h int) {
	w = childW
	h = childH

	if sticky&StickE != 0 && sticky&StickW != 0 {
		w = cavW
	}
	if sticky&StickN != 0 && sticky&StickS != 0 {
		h = cavH
	}

	// Horizontal positioning.
	if sticky&StickW != 0 {
		x = cavX
	} else if sticky&StickE != 0 {
		x = cavX + cavW - w
	} else {
		x = cavX + (cavW-w)/2
	}

	// Vertical positioning.
	if sticky&StickN != 0 {
		y = cavY
	} else if sticky&StickS != 0 {
		y = cavY + cavH - h
	} else {
		y = cavY + (cavH-h)/2
	}

	return x, y, w, h
}

// ArrangeAll triggers layout for all grid-managed containers.
func ArrangeAll() {
	for _, g := range gridders {
		g.arrange()
	}
}

// ArrangeContainer triggers layout for a specific container.
func ArrangeContainer(container *window.Window) {
	if g, ok := gridders[container]; ok {
		g.arrange()
	}
}
