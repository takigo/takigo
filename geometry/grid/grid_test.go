package grid

import (
	"testing"
)

func TestApplySticky(t *testing.T) {
	tests := []struct {
		name                   string
		sticky                 int
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
			name: "nsew_fills_cavity",
			sticky: NSEW,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 20, wantW: 100, wantH: 80,
		},
		{
			name: "ew_stretches_horizontal",
			sticky: EW,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 45, wantW: 100, wantH: 30,
		},
		{
			name: "ns_stretches_vertical",
			sticky: NS,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 20, wantW: 40, wantH: 80,
		},
		{
			name: "n_only",
			sticky: StickN,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 20, wantW: 40, wantH: 30,
		},
		{
			name: "s_only",
			sticky: StickS,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 40, wantY: 70, wantW: 40, wantH: 30,
		},
		{
			name: "w_only",
			sticky: StickW,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 45, wantW: 40, wantH: 30,
		},
		{
			name: "e_only",
			sticky: StickE,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 70, wantY: 45, wantW: 40, wantH: 30,
		},
		{
			name: "nw_corner",
			sticky: StickN | StickW,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
			childW: 40, childH: 30,
			wantX: 10, wantY: 20, wantW: 40, wantH: 30,
		},
		{
			name: "se_corner",
			sticky: StickS | StickE,
			cavX: 10, cavY: 20, cavW: 100, cavH: 80,
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

func TestDistributeExtra(t *testing.T) {
	tests := []struct {
		name  string
		sizes []int
		conf  map[int]*SlotConfig
		extra int
		want  []int
	}{
		{
			name:  "no_extra",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 1}},
			extra: 0,
			want:  []int{50, 50},
		},
		{
			name:  "negative_extra",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{0: {Weight: 1}},
			extra: -10,
			want:  []int{50, 50},
		},
		{
			name:  "equal_weight",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 1}},
			extra: 100,
			want:  []int{100, 100},
		},
		{
			name:  "unequal_weight_2_1",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{0: {Weight: 2}, 1: {Weight: 1}},
			extra: 90,
			want:  []int{110, 80},
		},
		{
			name:  "single_weighted",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{1: {Weight: 1}},
			extra: 100,
			want:  []int{50, 150},
		},
		{
			name:  "no_weights",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{},
			extra: 100,
			want:  []int{50, 50},
		},
		{
			name:  "zero_weight_ignored",
			sizes: []int{50, 50},
			conf:  map[int]*SlotConfig{0: {Weight: 0}, 1: {Weight: 1}},
			extra: 100,
			want:  []int{50, 150},
		},
		{
			name:  "three_cols_weight_1_2_1",
			sizes: []int{30, 30, 30},
			conf:  map[int]*SlotConfig{0: {Weight: 1}, 1: {Weight: 2}, 2: {Weight: 1}},
			extra: 40,
			want:  []int{40, 50, 40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sizes := make([]int, len(tt.sizes))
			copy(sizes, tt.sizes)
			distributeExtra(sizes, tt.conf, tt.extra)
			for i, got := range sizes {
				if got != tt.want[i] {
					t.Errorf("sizes[%d] = %d, want %d (full: %v, want: %v)", i, got, tt.want[i], sizes, tt.want)
					break
				}
			}
		})
	}
}
