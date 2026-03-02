package ttk

import "testing"

func TestBoxContains(t *testing.T) {
	box := Box{X: 10, Y: 20, Width: 100, Height: 50}
	tests := []struct {
		px, py int
		want   bool
	}{
		{10, 20, true},   // top-left corner (inclusive)
		{50, 40, true},   // interior
		{109, 69, true},  // bottom-right just inside
		{110, 70, false}, // just outside right-bottom edge
		{9, 20, false},   // just left
		{10, 19, false},  // just above
	}
	for _, tt := range tests {
		if got := box.Contains(tt.px, tt.py); got != tt.want {
			t.Errorf("Contains(%d,%d) = %v, want %v", tt.px, tt.py, got, tt.want)
		}
	}
}

func TestBoxContainsZeroSize(t *testing.T) {
	box := Box{X: 5, Y: 5, Width: 0, Height: 0}
	if box.Contains(5, 5) {
		t.Error("zero-size box should not contain any point")
	}
}

func TestPadBox(t *testing.T) {
	box := Box{X: 10, Y: 20, Width: 100, Height: 80}
	p := Padding{Left: 5, Top: 10, Right: 15, Bottom: 20}
	got := PadBox(box, p)
	want := Box{X: 15, Y: 30, Width: 80, Height: 50}
	if got != want {
		t.Errorf("PadBox = %+v, want %+v", got, want)
	}
}

func TestPadBoxOverflow(t *testing.T) {
	box := Box{X: 0, Y: 0, Width: 10, Height: 10}
	p := Padding{Left: 20, Top: 20, Right: 20, Bottom: 20}
	got := PadBox(box, p)
	if got.Width != 0 || got.Height != 0 {
		t.Errorf("PadBox overflow: Width=%d Height=%d, want 0,0", got.Width, got.Height)
	}
}

func TestExpandBox(t *testing.T) {
	box := Box{X: 20, Y: 30, Width: 60, Height: 40}
	p := Padding{Left: 5, Top: 10, Right: 5, Bottom: 10}
	got := ExpandBox(box, p)
	want := Box{X: 15, Y: 20, Width: 70, Height: 60}
	if got != want {
		t.Errorf("ExpandBox = %+v, want %+v", got, want)
	}
}

func TestUniformPadding(t *testing.T) {
	p := UniformPadding(8)
	if p.Left != 8 || p.Top != 8 || p.Right != 8 || p.Bottom != 8 {
		t.Errorf("UniformPadding(8) = %+v", p)
	}
}

func TestPaddingWidthHeight(t *testing.T) {
	p := Padding{Left: 3, Top: 5, Right: 7, Bottom: 11}
	if w := p.Width(); w != 10 {
		t.Errorf("Width() = %d, want 10", w)
	}
	if h := p.Height(); h != 16 {
		t.Errorf("Height() = %d, want 16", h)
	}
}

func TestPaddingAdd(t *testing.T) {
	a := Padding{1, 2, 3, 4}
	b := Padding{10, 20, 30, 40}
	got := a.Add(b)
	want := Padding{11, 22, 33, 44}
	if got != want {
		t.Errorf("Add = %+v, want %+v", got, want)
	}
}

func TestPackBoxTop(t *testing.T) {
	cavity := Box{X: 0, Y: 0, Width: 100, Height: 100}
	parcel := PackBox(&cavity, 100, 30, SideTop)
	if parcel.Y != 0 || parcel.Height != 30 {
		t.Errorf("PackBox top parcel = %+v", parcel)
	}
	if cavity.Y != 30 || cavity.Height != 70 {
		t.Errorf("cavity after top = %+v", cavity)
	}
}

func TestPackBoxBottom(t *testing.T) {
	cavity := Box{X: 0, Y: 0, Width: 100, Height: 100}
	parcel := PackBox(&cavity, 100, 25, SideBottom)
	if parcel.Y != 75 || parcel.Height != 25 {
		t.Errorf("PackBox bottom parcel = %+v", parcel)
	}
	if cavity.Height != 75 {
		t.Errorf("cavity after bottom = %+v", cavity)
	}
}

func TestPackBoxLeft(t *testing.T) {
	cavity := Box{X: 0, Y: 0, Width: 100, Height: 100}
	parcel := PackBox(&cavity, 40, 100, SideLeft)
	if parcel.X != 0 || parcel.Width != 40 {
		t.Errorf("PackBox left parcel = %+v", parcel)
	}
	if cavity.X != 40 || cavity.Width != 60 {
		t.Errorf("cavity after left = %+v", cavity)
	}
}

func TestPackBoxRight(t *testing.T) {
	cavity := Box{X: 0, Y: 0, Width: 100, Height: 100}
	parcel := PackBox(&cavity, 35, 100, SideRight)
	if parcel.X != 65 || parcel.Width != 35 {
		t.Errorf("PackBox right parcel = %+v", parcel)
	}
	if cavity.Width != 65 {
		t.Errorf("cavity after right = %+v", cavity)
	}
}

func TestStickBoxFillBoth(t *testing.T) {
	parcel := Box{X: 10, Y: 20, Width: 100, Height: 80}
	got := StickBox(parcel, 30, 20, FillBoth)
	if got != parcel {
		t.Errorf("StickBox FillBoth = %+v, want %+v", got, parcel)
	}
}

func TestStickBoxCenter(t *testing.T) {
	parcel := Box{X: 0, Y: 0, Width: 100, Height: 100}
	got := StickBox(parcel, 20, 10, 0) // no sticky = center
	want := Box{X: 40, Y: 45, Width: 20, Height: 10}
	if got != want {
		t.Errorf("StickBox center = %+v, want %+v", got, want)
	}
}

func TestStickBoxWest(t *testing.T) {
	parcel := Box{X: 0, Y: 0, Width: 100, Height: 100}
	got := StickBox(parcel, 20, 10, StickW)
	if got.X != 0 {
		t.Errorf("StickBox W: X = %d, want 0", got.X)
	}
}

func TestStickBoxEast(t *testing.T) {
	parcel := Box{X: 0, Y: 0, Width: 100, Height: 100}
	got := StickBox(parcel, 20, 10, StickE)
	if got.X != 80 {
		t.Errorf("StickBox E: X = %d, want 80", got.X)
	}
}

func TestStickBoxSouth(t *testing.T) {
	parcel := Box{X: 0, Y: 0, Width: 100, Height: 100}
	got := StickBox(parcel, 20, 10, StickS)
	if got.Y != 90 {
		t.Errorf("StickBox S: Y = %d, want 90", got.Y)
	}
}
