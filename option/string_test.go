package option

import "testing"

func TestStrings(t *testing.T) {
	tests := []struct {
		got, want string
	}{
		{AnchorNE.String(), "ne"},
		{AnchorCenter.String(), "center"},
		{Anchor(99).String(), "center"},
		{JustifyRight.String(), "right"},
		{Horizontal.String(), "horizontal"},
		{Vertical.String(), "vertical"},
		{SideLeft.String(), "left"},
		{DirAbove.String(), "above"},
		{StickNSEW.String(), "nesw"},
		{(StickE | StickW).String(), "ew"},
		{Sticky(0).String(), ""},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("String() = %q, want %q", tt.got, tt.want)
		}
	}
}
