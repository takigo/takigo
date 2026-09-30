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
