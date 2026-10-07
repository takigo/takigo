package grid

import (
	"testing"

	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/window"
)

// FuzzGridOption fuzzes GridOption functions.
func FuzzGridOption(f *testing.F) {
	f.Fuzz(func(t *testing.T, row, col, rowSpan, colSpan, sticky, padX, padY, iPadX, iPadY int) {
		cfg := gridConfig{}
		Row(row)(&cfg)
		Column(col)(&cfg)
		RowSpan(rowSpan)(&cfg)
		ColumnSpan(colSpan)(&cfg)
		Sticky(option.Sticky(sticky))(&cfg)
		PadX(padX)(&cfg)
		PadY(padY)(&cfg)
		IPadX(iPadX)(&cfg)
		IPadY(iPadY)(&cfg)
		_ = cfg
	})
}

// FuzzSlotOption fuzzes SlotOption functions.
func FuzzSlotOption(f *testing.F) {
	f.Fuzz(func(t *testing.T, minSize, weight, pad int, uniform string) {
		conf := SlotConfig{}
		MinSize(minSize)(&conf)
		Weight(weight)(&conf)
		Pad(pad)(&conf)
		Uniform(uniform)(&conf)
		_ = conf
	})
}

// FuzzResolveConstraints fuzzes the resolveConstraints function.
func FuzzResolveConstraints(f *testing.F) {
	// Create a minimal gridder with some entries
	g := newGridder(&window.Window{})
	g.container = &window.Window{
		InternalBorderLeft:   0,
		InternalBorderRight:  0,
		InternalBorderTop:    0,
		InternalBorderBottom: 0,
	}
	g.anchor = option.AnchorNW
	g.propagate = true

	// Seed with some entries
	f.Add(0, 0, 1, 1, 100, 100, 1)
	f.Add(1, 1, 1, 1, 100, 100, 2)
	f.Add(2, 0, 1, 2, 0, 0, 0)

	f.Fuzz(func(t *testing.T, row, col, rowSpan, colSpan, reqW, reqH, weight int) {
		if row < 0 || col < 0 || rowSpan < 1 || colSpan < 1 {
			return
		}
		if row+rowSpan > maxElement || col+colSpan > maxElement {
			return
		}

		w := &window.Window{
			ReqWidth:    reqW,
			ReqHeight:   reqH,
			BorderWidth: 0,
		}
		g.entries = append(g.entries, &gridEntry{
			window: w,
			config: gridConfig{
				row:        row,
				column:     col,
				rowSpan:    rowSpan,
				columnSpan: colSpan,
				padX:       0,
				padY:       0,
				iPadX:      0,
				iPadY:      0,
			},
		})
		g.colConf[col] = &SlotConfig{Weight: weight}
		g.rowConf[row] = &SlotConfig{Weight: weight}

		// Test resolveConstraints for columns
		_, _ = resolveConstraints(g.entries, g.colConf, col+colSpan+1, true)
		// Test resolveConstraints for rows
		_, _ = resolveConstraints(g.entries, g.rowConf, row+rowSpan+1, false)

		// Clean up for next iteration
		g.entries = g.entries[:len(g.entries)-1]
		delete(g.colConf, col)
		delete(g.rowConf, row)
	})
}

// FuzzAdjustOffsets fuzzes the adjustOffsets function.
func FuzzAdjustOffsets(f *testing.F) {
	offsets := []int{0, 100, 200, 300}
	conf := map[int]*SlotConfig{
		0: {Weight: 1, MinSize: 50},
		1: {Weight: 2, MinSize: 50},
		2: {Weight: 1, MinSize: 50},
	}

	f.Add(400)
	f.Add(500)
	f.Add(100)
	f.Add(0)
	f.Add(-100)

	f.Fuzz(func(t *testing.T, size int) {
		// Make copies to avoid mutation issues
		off := make([]int, len(offsets))
		copy(off, offsets)
		_ = adjustOffsets(size, off, conf)
	})
}

// FuzzApplySticky fuzzes the applySticky function.
func FuzzApplySticky(f *testing.F) {
	f.Fuzz(func(t *testing.T, sticky, cavX, cavY, cavW, cavH, childW, childH int) {
		if cavW < 0 || cavH < 0 || childW < 0 || childH < 0 {
			return
		}
		_, _, _, _ = applySticky(option.Sticky(sticky), cavX, cavY, cavW, cavH, childW, childH)
	})
}

// FuzzComputeAnchor fuzzes the computeAnchor function.
func FuzzComputeAnchor(f *testing.F) {
	container := &window.Window{
		Width:                800,
		Height:               600,
		InternalBorderLeft:   10,
		InternalBorderRight:  10,
		InternalBorderTop:    10,
		InternalBorderBottom: 10,
	}

	f.Add(int(option.AnchorNW), 400, 300)
	f.Add(int(option.AnchorCenter), 400, 300)
	f.Add(int(option.AnchorSE), 400, 300)

	f.Fuzz(func(t *testing.T, anchor int, usedW, usedH int) {
		a := option.Anchor(anchor)
		_, _ = computeAnchor(a, container, usedW, usedH)
	})
}
