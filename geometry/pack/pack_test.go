package pack

import (
	"testing"

	"github.com/msorc/takigo/option"
	"github.com/msorc/takigo/window"
)

func TestAnchorPosition(t *testing.T) {
	tests := []struct {
		name                           string
		anchor                         option.Anchor
		frameX, frameY, frameW, frameH int
		childW, childH                 int
		wantX, wantY                   int
	}{
		{"center", option.AnchorCenter, 10, 10, 100, 100, 20, 20, 50, 50},
		{"nw", option.AnchorNW, 10, 10, 100, 100, 20, 20, 10, 10},
		{"ne", option.AnchorNE, 10, 10, 100, 100, 20, 20, 90, 10},
		{"sw", option.AnchorSW, 10, 10, 100, 100, 20, 20, 10, 90},
		{"se", option.AnchorSE, 10, 10, 100, 100, 20, 20, 90, 90},
		{"n", option.AnchorN, 10, 10, 100, 100, 20, 20, 50, 10},
		{"s", option.AnchorS, 10, 10, 100, 100, 20, 20, 50, 90},
		{"w", option.AnchorW, 10, 10, 100, 100, 20, 20, 10, 50},
		{"e", option.AnchorE, 10, 10, 100, 100, 20, 20, 90, 50},
		{"center_exact_fit", option.AnchorCenter, 0, 0, 50, 50, 50, 50, 0, 0},
		{"center_odd", option.AnchorCenter, 0, 0, 101, 101, 20, 20, 40, 40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := anchorPosition(tt.anchor, tt.frameX, tt.frameY,
				tt.frameW, tt.frameH, tt.childW, tt.childH)
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("anchorPosition(%v) = (%d, %d), want (%d, %d)",
					tt.anchor, x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestComputeSize(t *testing.T) {
	tests := []struct {
		name    string
		entries []*packEntry
		wantW   int
		wantH   int
	}{
		{
			name:    "empty",
			entries: nil,
			wantW:   0,
			wantH:   0,
		},
		{
			name: "single_top",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top}},
			},
			wantW: 100,
			wantH: 30,
		},
		{
			name: "single_left",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Left}},
			},
			wantW: 100,
			wantH: 30,
		},
		{
			name: "two_top",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top}},
				{window: &window.Window{ReqWidth: 80, ReqHeight: 40}, config: packConfig{side: Top}},
			},
			wantW: 100, // max width
			wantH: 70,  // sum of heights
		},
		{
			name: "two_left",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Left}},
				{window: &window.Window{ReqWidth: 80, ReqHeight: 40}, config: packConfig{side: Left}},
			},
			wantW: 180, // sum of widths
			wantH: 40,  // max height
		},
		{
			name: "mixed_top_left",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top}},
				{window: &window.Window{ReqWidth: 50, ReqHeight: 60}, config: packConfig{side: Left}},
			},
			wantW: 100, // max(top 100, left 50)
			wantH: 90,  // top 30 + left 60
		},
		{
			// Tk order matters: the right-packed scrollbar narrows the cavity
			// for the text packed after it (tkPack.c ArrangePacking).
			name: "style_demo_bottom_right_top",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 222, ReqHeight: 42}, config: packConfig{side: Bottom}},
				{window: &window.Window{ReqWidth: 15, ReqHeight: 38}, config: packConfig{side: Right}},
				{window: &window.Window{ReqWidth: 706, ReqHeight: 614}, config: packConfig{side: Top}},
			},
			wantW: 721,
			wantH: 656,
		},
		{
			name: "left_then_top",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 50, ReqHeight: 60}, config: packConfig{side: Left}},
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top}},
			},
			wantW: 150,
			wantH: 60,
		},
		{
			name: "with_padding",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top, padX: 10, padY: 20, padLeft: 5, padTop: 10}},
			},
			wantW: 110, // 100 + 5*2
			wantH: 50,  // 30 + 10*2
		},
		{
			name: "with_ipadding",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30}, config: packConfig{side: Top, iPadX: 5, iPadY: 10}},
			},
			wantW: 110, // 100 + 5*2
			wantH: 50,  // 30 + 10*2
		},
		{
			name: "with_border_width",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 100, ReqHeight: 30, BorderWidth: 2}, config: packConfig{side: Top}},
			},
			wantW: 104, // 100 + 2*2
			wantH: 34,  // 30 + 2*2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &packer{entries: tt.entries}
			w, h := p.computeSize()
			if w != tt.wantW || h != tt.wantH {
				t.Errorf("computeSize() = (%d, %d), want (%d, %d)", w, h, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestXExpansion(t *testing.T) {
	tests := []struct {
		name      string
		entries   []*packEntry
		targetIdx int
		cavityW   int
		want      int
	}{
		{
			name: "single_expander",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: true}},
			},
			targetIdx: 0,
			cavityW:   200,
			want:      150,
		},
		{
			name: "two_expanders_equal",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: true}},
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: true}},
			},
			targetIdx: 0,
			cavityW:   200,
			want:      50, // (200 - 50 - 50) / 2
		},
		{
			name: "no_extra_space",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 200}, config: packConfig{side: Left, expand: true}},
			},
			targetIdx: 0,
			cavityW:   100,
			want:      0,
		},
		{
			name: "non_expanding_sibling",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: true}},
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: false}},
			},
			targetIdx: 0,
			cavityW:   200,
			want:      100, // (200 - 50 - 50) / 1
		},
		{
			name: "top_side_ignored",
			entries: []*packEntry{
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Left, expand: true}},
				{window: &window.Window{ReqWidth: 50}, config: packConfig{side: Top, expand: true}},
			},
			targetIdx: 0,
			cavityW:   200,
			want:      150, // only Left expander counts
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := xExpansion(tt.entries, tt.entries[tt.targetIdx], tt.cavityW)
			if got != tt.want {
				t.Errorf("xExpansion() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestYExpansion(t *testing.T) {
	tests := []struct {
		name      string
		entries   []*packEntry
		targetIdx int
		cavityH   int
		want      int
	}{
		{
			name: "single_expander",
			entries: []*packEntry{
				{window: &window.Window{ReqHeight: 30}, config: packConfig{side: Top, expand: true}},
			},
			targetIdx: 0,
			cavityH:   200,
			want:      170,
		},
		{
			name: "two_expanders_equal",
			entries: []*packEntry{
				{window: &window.Window{ReqHeight: 30}, config: packConfig{side: Top, expand: true}},
				{window: &window.Window{ReqHeight: 30}, config: packConfig{side: Bottom, expand: true}},
			},
			targetIdx: 0,
			cavityH:   200,
			want:      70, // (200 - 30 - 30) / 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := yExpansion(tt.entries, tt.entries[tt.targetIdx], tt.cavityH)
			if got != tt.want {
				t.Errorf("yExpansion() = %d, want %d", got, tt.want)
			}
		})
	}
}
