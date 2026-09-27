// Package grid implements the grid geometry manager, which arranges
// children in a two-dimensional table. It ports tk/generic/tkGrid.c.
package grid

import (
	"github.com/msorc/takigo/geometry"
	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/screenunit"
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

// maxElement limits grid dimensions to prevent denial of service.
// Matches Tk's MAX_ELEMENT.
const maxElement = 10000

// GridOption configures a Grid call.
type GridOption func(*gridConfig)

type gridConfig struct {
	row        int
	column     int
	rowSpan    int
	columnSpan int
	sticky     int
	padX       int            // total horizontal padding (left + right)
	padY       int            // total vertical padding (top + bottom)
	padLeft    int            // left portion of padX
	padTop     int            // top portion of padY
	iPadX      int            // total internal horizontal padding (2× user value)
	iPadY      int            // total internal vertical padding (2× user value)
	in         *window.Window // -in: a container other than the parent
}

// In is grid's -in: manage the content inside container, which must be the
// content's parent or a descendant of it (Tk_MaintainGeometry keeps it there).
func In(container window.Windower) GridOption {
	return func(c *gridConfig) { c.in = container.Window() }
}

// containerOf records each content window's container (Tk's containerPtr).
var containerOf = map[*window.Window]*window.Window{}

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

// PadX sets the exterior horizontal padding (symmetric).
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadX(p any) GridOption {
	return func(c *gridConfig) {
		v := screenunit.Px(p)
		c.padX = v * 2
		c.padLeft = v
	}
}

// PadY sets the exterior vertical padding (symmetric).
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func PadY(p any) GridOption {
	return func(c *gridConfig) {
		v := screenunit.Px(p)
		c.padY = v * 2
		c.padTop = v
	}
}

// PadXPair sets asymmetric exterior horizontal padding.
func PadXPair(left, right any) GridOption {
	return func(c *gridConfig) {
		l := screenunit.Px(left)
		r := screenunit.Px(right)
		c.padLeft = l
		c.padX = l + r
	}
}

// PadYPair sets asymmetric exterior vertical padding.
func PadYPair(top, bottom any) GridOption {
	return func(c *gridConfig) {
		t := screenunit.Px(top)
		b := screenunit.Px(bottom)
		c.padTop = t
		c.padY = t + b
	}
}

// IPadX sets the interior horizontal padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func IPadX(p any) GridOption {
	return func(c *gridConfig) { c.iPadX = screenunit.Px(p) * 2 }
}

// IPadY sets the interior vertical padding.
// Accepts int (pixels), float64 (rounded pixels), or string with unit suffix ("3p", "2m", "1c", "0.5i").
func IPadY(p any) GridOption {
	return func(c *gridConfig) { c.iPadY = screenunit.Px(p) * 2 }
}

// SlotConfig holds configuration for a row or column.
type SlotConfig struct {
	MinSize int
	Weight  int
	Pad     int
	// Uniform groups slots so they all get the same minimum size (the max
	// of their individual minimums). Matches Tk's -uniform option.
	Uniform string
}

// SlotOption configures a SlotConfig.
type SlotOption func(*SlotConfig)

// MinSize sets the minimum size for a row or column.
func MinSize(n int) SlotOption { return func(c *SlotConfig) { c.MinSize = n } }

// Weight sets the weight for distributing extra space.
func Weight(n int) SlotOption { return func(c *SlotConfig) { c.Weight = n } }

// Pad sets the padding for a row or column.
func Pad(n int) SlotOption { return func(c *SlotConfig) { c.Pad = screenunit.Px(n) } }

// Uniform sets the uniform group name.
func Uniform(name string) SlotOption { return func(c *SlotConfig) { c.Uniform = name } }

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
	anchor    option.Anchor
	propagate bool
}

func newGridder(container *window.Window) *gridder {
	return &gridder{
		container: container,
		rowConf:   make(map[int]*SlotConfig),
		colConf:   make(map[int]*SlotConfig),
		anchor:    option.AnchorNW,
		propagate: true,
	}
}

// gridderFor returns the container's gridder, creating it with the resize
// hook that re-arranges the grid (whichever of grid, rowconfigure,
// columnconfigure or anchor touches the container first).
func gridderFor(container *window.Window) *gridder {
	if g, ok := gridders[container]; ok {
		return g
	}
	g := newGridder(container)
	gridders[container] = g
	container.ConfigureCallback = func() {
		if gg, ok := gridders[container]; ok {
			gg.arrange()
		}
	}
	return g
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

// RelativePlacement represents Tk's grid relative placement shortcuts.
// See tk/generic/tkGrid.c ConfigureContent and the "RELATIVE PLACEMENT"
// section in the grid manual.
type RelativePlacement int

const (
	// RelEmpty skips a column (Tk's "x").
	RelEmpty RelativePlacement = iota
	// RelLeft extends the previous widget one column to the right (Tk's "-").
	RelLeft
	// RelUp extends the widget above one row downward (Tk's "^").
	RelUp
)

// relativeWidget is a dummy widget used as a sentinel in Grid calls.
// It implements both window.Windower and geometry.Elementer so it can
// appear inside geometry.Group or be passed directly to Grid.
type relativeWidget struct {
	placement RelativePlacement
}

// sentinel Window pointers for each relative placement type.
var relSentinels [3]window.Window

func (r *relativeWidget) Window() *window.Window {
	return &relSentinels[r.placement]
}

func (r *relativeWidget) GeometryElements() []window.Windower {
	return []window.Windower{r}
}

// Relative returns a dummy widget representing a relative placement shortcut.
func Relative(rp RelativePlacement) *relativeWidget {
	return &relativeWidget{placement: rp}
}

func isRelative(w *window.Window) (RelativePlacement, bool) {
	for i := range relSentinels {
		if w == &relSentinels[i] {
			return RelativePlacement(i), true
		}
	}
	return 0, false
}

// Grid adds children to their parent's grid layout.
// Accepts a geometry.Elementer (e.g. geometry.Group) containing real widgets
// and/or relative placement markers created by Relative().
//
// Relative placements (matching Tk's grid shortcuts):
//   - Relative(RelEmpty): skip this column (Tk's "x")
//   - Relative(RelLeft):  extend previous widget's columnSpan (Tk's "-")
//   - Relative(RelUp):    extend widget above's rowSpan (Tk's "^")
//
// See tk/generic/tkGrid.c ConfigureContent.
func Grid(children geometry.Elementer, opts ...GridOption) {
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

	// Bounds check.
	if cfg.row+cfg.rowSpan > maxElement || cfg.column+cfg.columnSpan > maxElement {
		return
	}

	elements := children.GeometryElements()

	// Find the first real widget to determine the parent.
	var parent *window.Window
	for _, elem := range elements {
		w := elem.Window()
		if _, rel := isRelative(w); !rel {
			parent = w.Parent
			break
		}
	}
	if cfg.in != nil {
		parent = cfg.in
	}
	if parent == nil {
		return
	}

	g := gridderFor(parent)

	// Auto-assign row if not specified.
	row := cfg.row
	if row == 0 && cfg.column == 0 {
		row = g.nextRow()
	}

	// First pass: place real widgets, handle RelEmpty (x) and RelLeft (-).
	// RelLeft increases the previous widget's columnSpan.
	// See tk/generic/tkGrid.c lines 3162–3220.
	col := cfg.column
	var lastEntry *gridEntry
	for _, elem := range elements {
		w := elem.Window()
		rp, rel := isRelative(w)
		if rel {
			switch rp {
			case RelEmpty: // 'x' — skip column
				col++
				lastEntry = nil
			case RelLeft: // '-' — extend previous widget
				if lastEntry != nil {
					lastEntry.config.columnSpan++
				}
				col++
			case RelUp: // '^' — handled in second pass
				col++
			}
			continue
		}

		// Count trailing RelLeft to compute default columnSpan.
		ecfg := cfg
		ecfg.row = row
		ecfg.column = col

		if old := containerOf[w]; old != nil && old != parent {
			if og, ok := gridders[old]; ok {
				og.remove(w)
				og.arrange()
			}
		}
		containerOf[w] = parent
		geometry.ManageGeometry(w, mgr)

		// Update or add entry.
		found := false
		for _, e := range g.entries {
			if e.window == w {
				e.config = ecfg
				found = true
				lastEntry = e
				break
			}
		}
		if !found {
			entry := &gridEntry{window: w, config: ecfg}
			g.entries = append(g.entries, entry)
			lastEntry = entry
		}

		col++
	}

	// Second pass: handle RelUp ('^') — extend rowSpan of widget above.
	// See tk/generic/tkGrid.c lines 3480–3565.
	col = cfg.column
	for _, elem := range elements {
		w := elem.Window()
		rp, rel := isRelative(w)
		if !rel {
			col++
			continue
		}
		if rp != RelUp {
			col++
			continue
		}
		// Find the widget in the row above at this column.
		for _, e := range g.entries {
			if e.config.column == col &&
				e.config.row+e.config.rowSpan == row {
				e.config.rowSpan++
				break
			}
		}
		col++
	}

	g.arrange()
}

// Forget removes a child from grid management.
func Forget(child window.Windower) {
	w := child.Window()
	parent := containerOf[w]
	if parent == nil {
		parent = w.Parent
	}
	if parent == nil {
		return
	}
	delete(containerOf, w)
	w.GeomManager = nil
	// grid forget unmaps the content and re-grids the rest.
	if w.IsMapped() && w.PlatformID != platform.WindowID(0) {
		w.Display.Server.UnmapWindow(w.PlatformID)
		window.MarkUnmapped(w)
	}
	if g, ok := gridders[parent]; ok {
		g.remove(w)
		g.arrange()
	}
}

// RowConfigure sets configuration for a row.
func RowConfigure(container window.Windower, row int, opts ...SlotOption) {
	w := container.Window()
	conf := SlotConfig{}
	for _, opt := range opts {
		opt(&conf)
	}
	g := gridderFor(w)
	g.rowConf[row] = &conf
	g.arrange()
}

// ColumnConfigure sets configuration for a column.
func ColumnConfigure(container window.Windower, col int, opts ...SlotOption) {
	w := container.Window()
	conf := SlotConfig{}
	for _, opt := range opts {
		opt(&conf)
	}
	g := gridderFor(w)
	g.colConf[col] = &conf
	g.arrange()
}

// SetAnchor sets the anchor for a grid container. The anchor controls where
// an unweighted grid is placed within its container. Default is NW.
func SetAnchor(container window.Windower, anchor option.Anchor) {
	w := container.Window()
	g := gridderFor(w)
	g.anchor = anchor
	g.arrange()
}

// GetAnchor returns the anchor for a grid container.
func GetAnchor(container window.Windower) option.Anchor {
	w := container.Window()
	if g, ok := gridders[w]; ok {
		return g.anchor
	}
	return option.AnchorNW
}

// SetPropagate controls whether the grid propagates geometry requests to
// its container. Default is true.
func SetPropagate(container window.Windower, propagate bool) {
	w := container.Window()
	g := gridderFor(w)
	g.propagate = propagate
	g.arrange()
}

// GetPropagate returns whether the grid propagates geometry requests.
func GetPropagate(container window.Windower) bool {
	w := container.Window()
	if g, ok := gridders[w]; ok {
		return g.propagate
	}
	return true
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

// binEntry tracks a spanning widget binned by its right/bottom edge.
type binEntry struct {
	size int
	span int
}

// layoutSlot holds per-slot data during constraint resolution.
type layoutSlot struct {
	minSize   int
	pad       int
	weight    int
	uniform   string
	bins      []binEntry
	minOffset int
	maxOffset int
}

// resolveConstraints computes slot offsets for one dimension using Tk's
// 6-step constraint resolution algorithm (tk/generic/tkGrid.c ResolveConstraints).
// Returns the natural (required) size and cumulative slot offsets.
func resolveConstraints(entries []*gridEntry, conf map[int]*SlotConfig, gridCount int, isColumn bool) (int, []int) {
	if gridCount == 0 {
		return 0, nil
	}

	// Layout with dummy slot at index 0.
	layout := make([]layoutSlot, gridCount+1)

	// Step 1: Copy slot constraints into layout.
	step1CopyConstraints(layout, conf, gridCount)

	// Step 2: Process entries, bin spanning widgets.
	step2ProcessEntries(layout, entries, gridCount, isColumn)

	// Step 2b: Uniform groups with weight normalization.
	step2bUniformGroups(layout, gridCount)

	// Step 3: Compute minimum offsets left→right.
	requiredSize := step3ComputeMinOffsets(layout, gridCount)

	// Step 4: Compute maximum offsets right→left.
	step4ComputeMaxOffsets(layout, gridCount, requiredSize)

	// Step 5: Multi-pass weighted distribution within unconstrained spans.
	step5DistributeSpace(layout, gridCount)

	// Step 6: Extract cumulative offsets.
	offsets := make([]int, gridCount)
	for i := range gridCount {
		offsets[i] = layout[i+1].minOffset
	}

	return requiredSize, offsets
}

// step1CopyConstraints initializes layout slots from SlotConfig.
func step1CopyConstraints(layout []layoutSlot, conf map[int]*SlotConfig, gridCount int) {
	for i := range gridCount {
		li := i + 1
		if c, ok := conf[i]; ok {
			layout[li].minSize = c.MinSize
			layout[li].weight = c.Weight
			layout[li].pad = c.Pad
			layout[li].uniform = c.Uniform
		}
	}
}

// step2ProcessEntries processes span=1 entries directly and bins span>1 entries by right edge.
func step2ProcessEntries(layout []layoutSlot, entries []*gridEntry, gridCount int, isColumn bool) {
	for _, e := range entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth

		var size, span, slot int
		if isColumn {
			size = child.ReqWidth + bw2 + cfg.padX + cfg.iPadX
			span = cfg.columnSpan
			slot = cfg.column
		} else {
			size = child.ReqHeight + bw2 + cfg.padY + cfg.iPadY
			span = cfg.rowSpan
			slot = cfg.row
		}

		rightEdge := slot + span - 1
		if rightEdge >= gridCount || rightEdge < 0 {
			continue
		}
		li := rightEdge + 1

		if span > 1 {
			layout[li].bins = append(layout[li].bins, binEntry{size: size, span: span})
		} else {
			slotSize := size + layout[li].pad
			if slotSize > layout[li].minSize {
				layout[li].minSize = slotSize
			}
		}
	}
}

// step2bUniformGroups normalizes uniform groups with weight normalization.
func step2bUniformGroups(layout []layoutSlot, gridCount int) {
	type uniformGroup struct {
		name    string
		minSize int
	}
	var groups []uniformGroup

	for i := range gridCount {
		li := i + 1
		if layout[li].uniform == "" {
			continue
		}
		weight := layout[li].weight
		if weight <= 0 {
			weight = 1
		}
		normalized := (layout[li].minSize + weight - 1) / weight

		found := false
		for g := range groups {
			if groups[g].name == layout[li].uniform {
				if normalized > groups[g].minSize {
					groups[g].minSize = normalized
				}
				found = true
				break
			}
		}
		if !found {
			groups = append(groups, uniformGroup{name: layout[li].uniform, minSize: normalized})
		}
	}

	for _, ug := range groups {
		for i := range gridCount {
			li := i + 1
			if layout[li].uniform != ug.name {
				continue
			}
			weight := layout[li].weight
			if weight <= 0 {
				weight = 1
			}
			layout[li].minSize = ug.minSize * weight
		}
	}
}

// step3ComputeMinOffsets computes minimum offsets left→right. Returns requiredSize.
func step3ComputeMinOffsets(layout []layoutSlot, gridCount int) int {
	offset := 0
	for i := range gridCount {
		li := i + 1
		layout[li].minOffset = layout[li].minSize + offset
		for _, be := range layout[li].bins {
			startLi := li - be.span // may be 0 (dummy)
			required := be.size + layout[startLi].minOffset
			if required > layout[li].minOffset {
				layout[li].minOffset = required
			}
		}
		offset = layout[li].minOffset
	}
	return offset
}

// step4ComputeMaxOffsets computes maximum offsets right→left.
func step4ComputeMaxOffsets(layout []layoutSlot, gridCount int, requiredSize int) {
	for i := 1; i <= gridCount; i++ {
		layout[i].maxOffset = requiredSize
	}

	offset := requiredSize
	for i := gridCount - 1; i > 0; {
		li := i + 1
		for _, be := range layout[li].bins {
			startLi := li - be.span
			require := offset - be.size
			if startLi >= 1 && require < layout[startLi].maxOffset {
				layout[startLi].maxOffset = require
			}
		}
		offset -= layout[li].minSize
		i--
		li = i + 1
		if layout[li].maxOffset < offset {
			offset = layout[li].maxOffset
		} else {
			layout[li].maxOffset = offset
		}
	}
}

// step5DistributeSpace performs multi-pass weighted distribution within unconstrained spans.
func step5DistributeSpace(layout []layoutSlot, gridCount int) {
	for start := 0; start < gridCount; {
		startLi := start + 1
		if layout[startLi].minOffset == layout[startLi].maxOffset {
			start++
			continue
		}

		end := start + 1
		for end < gridCount {
			endLi := end + 1
			if layout[endLi].minOffset == layout[endLi].maxOffset {
				break
			}
			end++
		}
		endLi := end + 1

		totalWeight := 0
		need := 0
		for s := start; s <= end; s++ {
			sli := s + 1
			totalWeight += layout[sli].weight
			need += layout[sli].minSize
		}
		have := layout[endLi].maxOffset - layout[startLi-1].minOffset

		noWeights := false
		if totalWeight == 0 {
			noWeights = true
			totalWeight = end - start + 1
		}

		// Iteratively find the right "have" that fits all maxOffset constraints.
		var slot int
		for {
			prevMinOffset := layout[startLi-1].minOffset
			prevGrow := 0
			accWeight := 0

			for slot = start; slot <= end; slot++ {
				sli := slot + 1
				w := 1
				if !noWeights {
					w = layout[sli].weight
				}
				accWeight += w
				grow := (have-need)*accWeight/totalWeight - prevGrow
				prevGrow += grow

				if w > 0 && (prevMinOffset+layout[sli].minSize+grow) > layout[sli].maxOffset {
					grow = layout[sli].maxOffset - layout[sli].minSize - prevMinOffset
					newHave := grow * totalWeight / w
					if newHave > totalWeight {
						newHave = newHave / totalWeight * totalWeight
					}
					if newHave <= 0 {
						newHave = (have - need) - 1
						if newHave > 3*totalWeight {
							newHave = newHave * 3 / 4
						}
						if newHave > totalWeight {
							newHave = newHave / totalWeight * totalWeight
						}
						if newHave <= 0 {
							newHave = 1
						}
					}
					have = newHave + need
					break
				}
				prevMinOffset += layout[sli].minSize + grow
				if prevMinOffset < layout[sli].minOffset {
					prevMinOffset = layout[sli].minOffset
				}
			}

			if slot > end {
				break
			}
		}

		// Distribute the extra space.
		prevGrow := 0
		accWeight := 0
		for slot := start; slot <= end; slot++ {
			sli := slot + 1
			w := 1
			if !noWeights {
				w = layout[sli].weight
			}
			accWeight += w
			grow := (have-need)*accWeight/totalWeight - prevGrow
			prevGrow += grow
			layout[sli].minSize += grow
			if layout[sli-1].minOffset+layout[sli].minSize > layout[sli].minOffset {
				layout[sli].minOffset = layout[sli-1].minOffset + layout[sli].minSize
			}
		}

		// Propagate maxOffset changes backward.
		for slot := end; slot > start; slot-- {
			sli := slot + 1
			if layout[sli].maxOffset-layout[sli].minSize < layout[sli-1].maxOffset {
				layout[sli-1].maxOffset = layout[sli].maxOffset - layout[sli].minSize
			}
		}

		start = end + 1
	}
}

// adjustOffsets adjusts cumulative slot offsets to fit the available space.
// Handles both growing (adding space via weights) and shrinking (removing
// space down to configured minimums). Ports tk/generic/tkGrid.c AdjustOffsets.
// Returns the actual used size.
func adjustOffsets(size int, offsets []int, conf map[int]*SlotConfig) int {
	slots := len(offsets)
	if slots == 0 {
		return 0
	}

	diff := size - offsets[slots-1]

	if diff == 0 {
		return size
	}

	// Compute total weight.
	totalWeight := 0
	for i := range slots {
		if c, ok := conf[i]; ok && c.Weight > 0 {
			totalWeight += c.Weight
		}
	}

	if totalWeight == 0 {
		return offsets[slots-1]
	}

	// Growing: distribute extra space cumulatively by weight.
	if diff > 0 {
		cumWeight := 0
		for i := range slots {
			if c, ok := conf[i]; ok && c.Weight > 0 {
				cumWeight += c.Weight
			}
			offsets[i] += diff * cumWeight / totalWeight
		}
		return size
	}

	// Shrinking: compute minimum possible size.
	// Weighted slots shrink to their configured minSize.
	// Non-weighted slots keep their current size.
	temp := make([]int, slots)
	minTotal := 0
	for i := range slots {
		w := 0
		ms := 0
		if c, ok := conf[i]; ok {
			w = c.Weight
			ms = c.MinSize
		}
		if w > 0 {
			temp[i] = ms
		} else if i > 0 {
			temp[i] = offsets[i] - offsets[i-1]
		} else {
			temp[i] = offsets[i]
		}
		minTotal += temp[i]
	}

	// If requested size <= minimum, set all to minimum.
	if size <= minTotal {
		off := 0
		for i := range slots {
			off += temp[i]
			offsets[i] = off
		}
		return minTotal
	}

	// Iteratively remove space from weighted slots.
	for diff < 0 {
		// Find total weight for shrinkable slots.
		totalWeight = 0
		for i := range slots {
			current := offsets[i]
			if i > 0 {
				current -= offsets[i-1]
			}
			ms := 0
			if c, ok := conf[i]; ok {
				ms = c.MinSize
			}
			if current > ms {
				w := 0
				if c, ok := conf[i]; ok {
					w = c.Weight
				}
				totalWeight += w
				temp[i] = w
			} else {
				temp[i] = 0
			}
		}
		if totalWeight == 0 {
			break
		}

		// Find maximum shrink this pass.
		newDiff := diff
		for i := range slots {
			if temp[i] == 0 {
				continue
			}
			current := offsets[i]
			if i > 0 {
				current -= offsets[i-1]
			}
			ms := 0
			if c, ok := conf[i]; ok {
				ms = c.MinSize
			}
			maxDiff := totalWeight * (ms - current) / temp[i]
			if maxDiff > newDiff {
				newDiff = maxDiff
			}
		}

		// Distribute the shrink.
		cumWeight := 0
		for i := range slots {
			cumWeight += temp[i]
			offsets[i] += newDiff * cumWeight / totalWeight
		}
		diff -= newDiff
	}

	return size
}

// arrange performs the grid layout.
func (g *gridder) arrange() {
	container := g.container
	if container.PlatformID == platform.WindowID(0) || len(g.entries) == 0 {
		return
	}

	// Find grid dimensions from entries.
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

	// Extend dimensions to include configured constraints.
	for col := range g.colConf {
		if col+1 > maxCol {
			maxCol = col + 1
		}
	}
	for row := range g.rowConf {
		if row+1 > maxRow {
			maxRow = row + 1
		}
	}

	if maxCol == 0 || maxRow == 0 {
		return
	}

	// Resolve constraints for each dimension.
	reqW, colOffsets := resolveConstraints(g.entries, g.colConf, maxCol, true)
	reqH, rowOffsets := resolveConstraints(g.entries, g.rowConf, maxRow, false)

	totalReqW := reqW + container.InternalBorderLeft + container.InternalBorderRight
	totalReqH := reqH + container.InternalBorderTop + container.InternalBorderBottom

	if g.propagate {
		if container.IsTopLevel() {
			// For toplevel windows, resize the X window to fit content.
			// Matches Tk's Tk_GeometryRequest which calls XResizeWindow.
			if container.ReqWidth != totalReqW || container.ReqHeight != totalReqH {
				container.ReqWidth = totalReqW
				container.ReqHeight = totalReqH
				container.Width = totalReqW
				container.Height = totalReqH
				if container.PlatformID != platform.WindowID(0) {
					container.Display.Server.ResizeWindow(container.PlatformID,
						uint(totalReqW), uint(totalReqH))
				}
			}
		} else {
			geometry.GeometryRequest(container, totalReqW, totalReqH)
		}
	}

	// Available space within internal borders.
	availW := container.Width - container.InternalBorderLeft - container.InternalBorderRight
	availH := container.Height - container.InternalBorderTop - container.InternalBorderBottom

	// Adjust offsets for actual available space (grow/shrink).
	usedW := adjustOffsets(availW, colOffsets, g.colConf)
	usedH := adjustOffsets(availH, rowOffsets, g.rowConf)

	// Compute anchor-based start position.
	startX, startY := computeAnchor(g.anchor, container, usedW, usedH)

	// Position children.
	for _, e := range g.entries {
		child := e.window
		cfg := &e.config
		bw2 := 2 * child.BorderWidth

		col := cfg.column
		row := cfg.row
		endCol := min(col+cfg.columnSpan, maxCol) - 1
		endRow := min(row+cfg.rowSpan, maxRow) - 1

		// Slot span boundaries.
		leftEdge := 0
		if col > 0 && col-1 < len(colOffsets) {
			leftEdge = colOffsets[col-1]
		}
		topEdge := 0
		if row > 0 && row-1 < len(rowOffsets) {
			topEdge = rowOffsets[row-1]
		}

		cavX := startX + leftEdge + cfg.padLeft
		cavY := startY + topEdge + cfg.padTop
		cavW := colOffsets[endCol] - leftEdge - cfg.padX
		cavH := rowOffsets[endRow] - topEdge - cfg.padY

		childW := child.ReqWidth + bw2 + cfg.iPadX
		childH := child.ReqHeight + bw2 + cfg.iPadY

		// Apply sticky.
		x, y, w, h := applySticky(cfg.sticky, cavX, cavY, cavW, cavH, childW, childH)

		// Content gridded -in another container is offset by its position.
		dx, dy := window.ContentOffset(container, child)
		moved := child.X != x+dx || child.Y != y+dy
		child.X = x + dx
		child.Y = y + dy
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
			// Tk maps content only once its container is mapped; the
			// container's MarkMapped re-arranges and maps it then.
			if child.Flags&window.FlagMapped == 0 && window.ContainerViewable(container, child) {
				container.Display.Server.MapWindow(child.PlatformID)
				window.MarkMapped(child)
			}
		}
		if moved {
			window.NotifyMoved(child)
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

// computeAnchor computes the start position for the grid within its container,
// matching Tk's TkComputeAnchor behavior.
func computeAnchor(anchor option.Anchor, container *window.Window, usedW, usedH int) (int, int) {
	var x, y int

	switch anchor {
	case option.AnchorNW, option.AnchorW, option.AnchorSW:
		x = container.InternalBorderLeft
	case option.AnchorN, option.AnchorCenter, option.AnchorS:
		x = (container.Width - usedW) / 2
	default: // NE, E, SE
		x = container.Width - container.InternalBorderRight - usedW
	}

	switch anchor {
	case option.AnchorNW, option.AnchorN, option.AnchorNE:
		y = container.InternalBorderTop
	case option.AnchorW, option.AnchorCenter, option.AnchorE:
		y = (container.Height - usedH) / 2
	default: // SW, S, SE
		y = container.Height - container.InternalBorderBottom - usedH
	}

	return x, y
}

// ArrangeAll triggers layout for all grid-managed containers.
func ArrangeAll() {
	for _, g := range gridders {
		g.arrange()
	}
}

// Arrange (and so map) the content once its container is mapped, as the
// managers' structure procs do on MapNotify; keep -in content with its
// container when it moves or is unmapped (Tk_MaintainGeometry).
func init() {
	window.AddMappedHook(ArrangeContainer)
	window.AddMovedHook(func(w *window.Window) {
		if g, ok := gridders[w]; ok && g.hasForeign() {
			g.arrange()
		}
	})
	window.AddUnmappedHook(func(w *window.Window) {
		if g, ok := gridders[w]; ok {
			g.unmapForeign()
		}
	})
}

// hasForeign reports whether some content is gridded -in this container
// without being its child.
func (g *gridder) hasForeign() bool {
	for _, e := range g.entries {
		if e.window.Parent != g.container {
			return true
		}
	}
	return false
}

// unmapForeign unmaps -in content whose container was unmapped; X does this
// for real children.
func (g *gridder) unmapForeign() {
	for _, e := range g.entries {
		w := e.window
		if w.Parent != g.container && w.IsMapped() && w.PlatformID != platform.WindowID(0) {
			w.Display.Server.UnmapWindow(w.PlatformID)
			window.MarkUnmapped(w)
		}
	}
}

// ArrangeContainer triggers layout for a specific container.
func ArrangeContainer(container *window.Window) {
	if g, ok := gridders[container]; ok {
		g.arrange()
	}
}
