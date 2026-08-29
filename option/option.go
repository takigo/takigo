// Package option defines the functional option types used throughout takigo.
// Options are used both at widget creation time and for runtime configuration.
package option

// Option is a functional option that configures a widget or resource.
// It is applied to an Configurable target during creation or reconfiguration.
type Option func(target any)

// Apply applies a slice of options to a target.
func Apply(target any, opts []Option) {
	for _, opt := range opts {
		opt(target)
	}
}

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
