// Package option defines the option value types shared by takigo widgets:
// Relief, Anchor and Justify.
package option

// Relief specifies how a widget border appears.
type Relief int

const (
	ReliefFlat Relief = iota
	ReliefRaised
	ReliefSunken
	ReliefGroove
	ReliefRidge
	ReliefSolid
)

// String returns the relief name.
func (r Relief) String() string {
	switch r {
	case ReliefFlat:
		return "flat"
	case ReliefRaised:
		return "raised"
	case ReliefSunken:
		return "sunken"
	case ReliefGroove:
		return "groove"
	case ReliefRidge:
		return "ridge"
	case ReliefSolid:
		return "solid"
	default:
		return "flat"
	}
}

// ParseRelief parses a relief string.
func ParseRelief(s string) Relief {
	switch s {
	case "flat":
		return ReliefFlat
	case "raised":
		return ReliefRaised
	case "sunken":
		return ReliefSunken
	case "groove":
		return ReliefGroove
	case "ridge":
		return ReliefRidge
	case "solid":
		return ReliefSolid
	default:
		return ReliefFlat
	}
}

// Anchor specifies the position of content within a widget.
type Anchor int

const (
	AnchorCenter Anchor = iota
	AnchorN
	AnchorNE
	AnchorE
	AnchorSE
	AnchorS
	AnchorSW
	AnchorW
	AnchorNW
)

// Justify specifies text justification.
type Justify int

const (
	JustifyLeft Justify = iota
	JustifyCenter
	JustifyRight
)

// String returns the Tk name of the anchor ("center", "n", "ne", ...).
func (a Anchor) String() string {
	names := [...]string{"center", "n", "ne", "e", "se", "s", "sw", "w", "nw"}
	if a < 0 || int(a) >= len(names) {
		return "center"
	}
	return names[a]
}

// String returns the Tk name of the justification.
func (j Justify) String() string {
	switch j {
	case JustifyCenter:
		return "center"
	case JustifyRight:
		return "right"
	}
	return "left"
}

// Orient is the orientation of a scrollbar, scale, paned window, separator
// or progress bar.
type Orient int

const (
	Horizontal Orient = iota
	Vertical
)

// String returns the Tk name of the orientation.
func (o Orient) String() string {
	if o == Vertical {
		return "vertical"
	}
	return "horizontal"
}

// Side is an edge of a container, as in pack's -side.
type Side int

const (
	SideTop Side = iota
	SideBottom
	SideLeft
	SideRight
)

// String returns the Tk name of the side.
func (s Side) String() string {
	names := [...]string{"top", "bottom", "left", "right"}
	if s < 0 || int(s) >= len(names) {
		return "top"
	}
	return names[s]
}

// Direction is where a menubutton posts its menu (-direction).
type Direction int

const (
	DirBelow Direction = iota
	DirAbove
	DirLeft
	DirRight
)

// String returns the Tk name of the direction.
func (d Direction) String() string {
	names := [...]string{"below", "above", "left", "right"}
	if d < 0 || int(d) >= len(names) {
		return "below"
	}
	return names[d]
}

// Sticky is the set of cell or parcel edges a widget or element sticks to
// (-sticky).
type Sticky uint

const (
	StickN Sticky = 1 << iota
	StickE
	StickS
	StickW

	StickNS   = StickN | StickS
	StickEW   = StickE | StickW
	StickNSEW = StickNS | StickEW
)

// String returns the Tk form of the set, e.g. "nsew" or "ew".
func (s Sticky) String() string {
	var b []byte
	for i, c := range "nesw" {
		if s&(1<<i) != 0 {
			b = append(b, byte(c))
		}
	}
	return string(b)
}
