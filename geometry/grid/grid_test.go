package grid

import (
	"testing"

	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/window"
)

func TestApplySticky(t *testing.T) {
	tests := []struct {
		name                   string
		sticky                 option.Sticky
		cavX, cavY, cavW, cavH int
		childW, childH         int
		wantX, wantY           int
		wantW, wantH           int
	}{
		{
			name: "no_sticky_centers",
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 45, wantW: 40, wantH: 30,
		},
		{
			name:   "nsew_fills_cavity",
			sticky: NSEW,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 20, wantW: 100, wantH: 80,
		},
		{
			name:   "ew_stretches_horizontal",
			sticky: EW,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 45, wantW: 100, wantH: 30,
		},
		{
			name:   "ns_stretches_vertical",
			sticky: NS,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 20, wantW: 40, wantH: 80,
		},
		{
			name:   "n_only",
			sticky: StickN,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 20, wantW: 40, wantH: 30,
		},
		{
			name:   "s_only",
			sticky: StickS,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 70, wantW: 40, wantH: 30,
		},
		{
			name:   "w_only",
			sticky: StickW,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 45, wantW: 40, wantH: 30,
		},
		{
			name:   "e_only",
			sticky: StickE,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 70, wantY: 45, wantW: 40, wantH: 30,
		},
		{
			name:   "nw_corner",
			sticky: StickN | StickW,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 20, wantW: 40, wantH: 30,
		},
		{
			name:   "se_corner",
			sticky: StickS | StickE,
			cavX:   10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 70, wantY: 70, wantW: 40, wantH: 30,
		},
		{
			name: "exact_fit",
			cavX: 0, cavY: 0, cavW: 50, cavH: 50,
			childW: 50, childH: 50,
			wantX: 0, wantY: 0, wantW: 50, wantH: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, w, h := applySticky(tt.sticky, tt.cavX, tt.cavY, tt.cavW, tt.cavH, tt.childW, tt.childH)
			if x != tt.wantX || y != tt.wantY || w != tt.wantW || h != tt.wantH {
				t.Errorf("applySticky(%d) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					tt.sticky, x, y, w, h, tt.wantX, tt.wantY, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestAdjustOffsets(t *testing.T) {
	tests := []struct {
		name    string
		offsets []int
		conf    map[int]*SlotConfig
		size    int
		want    []int
		wantRet int
	}{
		{
			name:    "exact_fit",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 1}},
			size:    100,
			want:    []int{50, 100},
			wantRet: 100,
		},
		{
			name:    "grow_equal_weight",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 1}},
			size:    200,
			want:    []int{100, 200},
			wantRet: 200,
		},
		{
			name:    "grow_unequal_weight",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 2}, 1: {Weight: 1}},
			size:    190,
			want:    []int{110, 190},
			wantRet: 190,
		},
		{
			name:    "grow_single_weighted",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{1: {Weight: 1}},
			size:    200,
			want:    []int{50, 200},
			wantRet: 200,
		},
		{
			name:    "no_weights_no_change",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{},
			size:    200,
			want:    []int{50, 100},
			wantRet: 100,
		},
		{
			name:    "shrink_equal_weight",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 1}},
			size:    60,
			want:    []int{30, 60},
			wantRet: 60,
		},
		{
			name:    "shrink_respects_minsize",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 1, MinSize: 40}, 1: {Weight: 1}},
			size:    60,
			want:    []int{40, 60},
			wantRet: 60,
		},
		{
			name:    "shrink_below_minimum",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{0: {Weight: 1, MinSize: 40}, 1: {Weight: 1, MinSize: 40}},
			size:    50,
			want:    []int{40, 80},
			wantRet: 80,
		},
		{
			name:    "negative_no_weight",
			offsets: []int{50, 100},
			conf:    map[int]*SlotConfig{},
			size:    50,
			want:    []int{50, 100},
			wantRet: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offsets := make([]int, len(tt.offsets))
			copy(offsets, tt.offsets)
			ret := adjustOffsets(tt.size, offsets, tt.conf)
			if ret != tt.wantRet {
				t.Errorf("adjustOffsets returned %d, want %d", ret, tt.wantRet)
			}
			for i, got := range offsets {
				if got != tt.want[i] {
					t.Errorf("offsets[%d] = %d, want %d (full: %v, want: %v)", i, got, tt.want[i], offsets, tt.want)
					break
				}
			}
		})
	}
}

func TestResolveConstraints(t *testing.T) {
	makeEntry := func(col, row, colSpan, rowSpan, reqW, reqH int) *gridEntry {
		w := &window.Window{ReqWidth: reqW, ReqHeight: reqH}
		return &gridEntry{
			window: w,
			config: gridConfig{
				column:     col,
				row:        row,
				columnSpan: colSpan,
				rowSpan:    rowSpan,
			},
		}
	}

	t.Run("single_widget", func(t *testing.T) {
		entries := []*gridEntry{makeEntry(0, 0, 1, 1, 100, 50)}
		reqW, offsets := resolveConstraints(entries, map[int]*SlotConfig{}, 1, true)
		if reqW != 100 || len(offsets) != 1 || offsets[0] != 100 {
			t.Errorf("got reqW=%d offsets=%v, want reqW=100 offsets=[100]", reqW, offsets)
		}
	})

	t.Run("two_widgets_same_row", func(t *testing.T) {
		entries := []*gridEntry{
			makeEntry(0, 0, 1, 1, 60, 50),
			makeEntry(1, 0, 1, 1, 40, 50),
		}
		reqW, offsets := resolveConstraints(entries, map[int]*SlotConfig{}, 2, true)
		if reqW != 100 || offsets[0] != 60 || offsets[1] != 100 {
			t.Errorf("got reqW=%d offsets=%v, want reqW=100 offsets=[60,100]", reqW, offsets)
		}
	})

	t.Run("spanning_widget", func(t *testing.T) {
		entries := []*gridEntry{
			makeEntry(0, 0, 1, 1, 40, 50),
			makeEntry(1, 0, 1, 1, 40, 50),
			makeEntry(0, 1, 2, 1, 120, 30), // spans both columns, needs 120
		}
		reqW, offsets := resolveConstraints(entries, map[int]*SlotConfig{}, 2, true)
		if reqW < 120 {
			t.Errorf("got reqW=%d, want >= 120", reqW)
		}
		if offsets[1] < 120 {
			t.Errorf("got total offset=%d, want >= 120", offsets[1])
		}
	})

	t.Run("slot_pad", func(t *testing.T) {
		entries := []*gridEntry{makeEntry(0, 0, 1, 1, 100, 50)}
		conf := map[int]*SlotConfig{0: {Pad: 20}}
		reqW, offsets := resolveConstraints(entries, conf, 1, true)
		if reqW != 120 || offsets[0] != 120 {
			t.Errorf("got reqW=%d offsets=%v, want reqW=120 offsets=[120]", reqW, offsets)
		}
	})

	t.Run("uniform_with_weight", func(t *testing.T) {
		entries := []*gridEntry{
			makeEntry(0, 0, 1, 1, 60, 50),
			makeEntry(1, 0, 1, 1, 40, 50),
		}
		conf := map[int]*SlotConfig{
			0: {Weight: 1, Uniform: "a"},
			1: {Weight: 2, Uniform: "a"},
		}
		_, offsets := resolveConstraints(entries, conf, 2, true)
		// Col 0: minSize=60, weight=1 → normalized=60
		// Col 1: minSize=40, weight=2 → normalized=20
		// Group max normalized = 60
		// Col 0 final = 60*1 = 60
		// Col 1 final = 60*2 = 120
		if offsets[0] != 60 || offsets[1] != 180 {
			t.Errorf("uniform with weight: offsets=%v, want [60,180]", offsets)
		}
	})
}
