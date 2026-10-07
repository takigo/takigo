package ttk

import (
	"fmt"
	"strings"

	"github.com/takigo/takigo/option"
	"github.com/takigo/takigo/screenunit"
)

// Box represents a rectangular area. Ported from ttkLayout.c.
type Box struct {
	X, Y, Width, Height int
}

// Contains returns true if the point (px, py) is inside the box.
func (b Box) Contains(px, py int) bool {
	return px >= b.X && px < b.X+b.Width && py >= b.Y && py < b.Y+b.Height
}

// Padding represents padding on each side of a box.
type Padding struct {
	Left, Top, Right, Bottom int
}

// UniformPadding returns a Padding with all sides set to n.
func UniformPadding(n int) Padding {
	return Padding{n, n, n, n}
}

// Width returns the total horizontal padding.
func (p Padding) Width() int { return p.Left + p.Right }

// Height returns the total vertical padding.
func (p Padding) Height() int { return p.Top + p.Bottom }

// Add returns the sum of two paddings.
func (p Padding) Add(other Padding) Padding {
	return Padding{
		Left:   p.Left + other.Left,
		Top:    p.Top + other.Top,
		Right:  p.Right + other.Right,
		Bottom: p.Bottom + other.Bottom,
	}
}

// DistancePadding creates a Padding from four screen distances.
func DistancePadding(left, top, right, bottom screenunit.Distance) Padding {
	return Padding{Left: left.Pixels(), Top: top.Pixels(), Right: right.Pixels(), Bottom: bottom.Pixels()}
}

// UniformDistancePadding creates a Padding of d on every side.
func UniformDistancePadding(d screenunit.Distance) Padding {
	return UniformPadding(d.Pixels())
}

// ParsePadding parses a Tk padding spec of 1-4 distances ("2.25p",
// "7.5p 2.25p", "1.5p 0 7.5p 0"), as Ttk_GetPaddingFromObj does: missing
// right/bottom values default to left/top. A bad distance is an error
// wrapping screenunit.ErrBadDistance.
func ParsePadding(spec string) (Padding, error) {
	f := strings.Fields(spec)
	if len(f) > 4 {
		return Padding{}, fmt.Errorf("ttk: padding %q: %w", spec, screenunit.ErrBadDistance)
	}
	var px [4]int
	for i, s := range f {
		d, err := screenunit.Parse(s)
		if err != nil {
			return Padding{}, fmt.Errorf("ttk: padding %q: %w", spec, err)
		}
		px[i] = d.Pixels()
	}
	switch len(f) {
	case 0:
		return Padding{}, nil
	case 1:
		return UniformPadding(px[0]), nil
	case 2:
		return Padding{px[0], px[1], px[0], px[1]}, nil
	case 3:
		return Padding{px[0], px[1], px[2], px[1]}, nil
	default:
		return Padding{px[0], px[1], px[2], px[3]}, nil
	}
}

// RelievePadding adds n pixels of padding according to relief, simulating
// a pressed-in look (Ttk_RelievePadding in tk/generic/ttk/ttkLayout.c).
func RelievePadding(p Padding, relief option.Relief, n int) Padding {
	switch relief {
	case option.ReliefRaised:
		p.Right += n
		p.Bottom += n
	case option.ReliefSunken:
		p.Left += n
		p.Top += n
	default:
		h1 := n / 2
		h2 := h1 + n%2
		p.Left += h1
		p.Top += h1
		p.Right += h2
		p.Bottom += h2
	}
	return p
}

// Side specifies which side to pack from.
type Side = option.Side

// The sides a layout packs an element against.
const (
	SideTop    = option.SideTop
	SideBottom = option.SideBottom
	SideLeft   = option.SideLeft
	SideRight  = option.SideRight
)

// Sticky flags control how content fills its parcel.
type Sticky = option.Sticky

// The edges an element sticks to within its parcel.
const (
	StickW   = option.StickW
	StickE   = option.StickE
	StickN   = option.StickN
	StickS   = option.StickS
	FillX    = StickW | StickE
	FillY    = StickN | StickS
	FillBoth = FillX | FillY
)

// PadBox shrinks a box inward by the given padding.
func PadBox(box Box, p Padding) Box {
	return Box{
		X:      box.X + p.Left,
		Y:      box.Y + p.Top,
		Width:  max(box.Width-p.Width(), 0),
		Height: max(box.Height-p.Height(), 0),
	}
}

// ExpandBox grows a box outward by the given padding.
func ExpandBox(box Box, p Padding) Box {
	return Box{
		X:      box.X - p.Left,
		Y:      box.Y - p.Top,
		Width:  box.Width + p.Width(),
		Height: box.Height + p.Height(),
	}
}

// PackBox carves a parcel of size w x h from the given side of cavity,
// and mutates cavity to reflect the remaining space.
func PackBox(cavity *Box, w, h int, side Side) Box {
	parcel := *cavity
	switch side {
	case SideTop:
		if h > cavity.Height {
			h = cavity.Height
		}
		parcel.Height = h
		cavity.Y += h
		cavity.Height -= h
	case SideBottom:
		if h > cavity.Height {
			h = cavity.Height
		}
		parcel.Y = cavity.Y + cavity.Height - h
		parcel.Height = h
		cavity.Height -= h
	case SideLeft:
		if w > cavity.Width {
			w = cavity.Width
		}
		parcel.Width = w
		cavity.X += w
		cavity.Width -= w
	case SideRight:
		if w > cavity.Width {
			w = cavity.Width
		}
		parcel.X = cavity.X + cavity.Width - w
		parcel.Width = w
		cavity.Width -= w
	}
	return parcel
}

// StickBox positions content of size w x h within parcel according to sticky flags.
func StickBox(parcel Box, w, h int, sticky Sticky) Box {
	if sticky&FillX == FillX {
		w = parcel.Width
	}
	if sticky&FillY == FillY {
		h = parcel.Height
	}
	if w > parcel.Width {
		w = parcel.Width
	}
	if h > parcel.Height {
		h = parcel.Height
	}

	x := parcel.X
	y := parcel.Y

	switch {
	case sticky&StickW != 0 && sticky&StickE != 0:
		// centered & filled (already handled above)
	case sticky&StickW != 0:
		// stick to left (x is already left)
	case sticky&StickE != 0:
		x = parcel.X + parcel.Width - w
	default:
		x = parcel.X + (parcel.Width-w)/2
	}

	switch {
	case sticky&StickN != 0 && sticky&StickS != 0:
		// centered & filled
	case sticky&StickN != 0:
		// stick to top
	case sticky&StickS != 0:
		y = parcel.Y + parcel.Height - h
	default:
		y = parcel.Y + (parcel.Height-h)/2
	}

	return Box{X: x, Y: y, Width: w, Height: h}
}
