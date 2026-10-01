package ttk

// State is a bitmask of widget states, ported from ttkTheme.h.
type State uint32

const (
	StateActive     State = 1 << iota // mouse cursor is over the widget
	StateDisabled                     // widget is disabled
	StateFocus                        // widget has keyboard focus
	StatePressed                      // widget is being pressed
	StateSelected                     // "on", "true", or "current" for check/radiobuttons
	StateBackground                   // widget is in a background window
	StateAlternate                    // widget-specific alternate display
	StateInvalid                      // widget value is invalid
	StateReadonly                     // widget is read-only
	StateHover                        // mouse is hovering (synonym of Active in many themes)
)

// StateSpec matches a state: OnBits must all be set, OffBits must all be clear.
type StateSpec struct {
	OnBits  State
	OffBits State
}

// Matches returns true if all OnBits are set and all OffBits are clear in state.
func (s StateSpec) Matches(state State) bool {
	return (state&s.OnBits == s.OnBits) && (state&s.OffBits == 0)
}

// Modify returns state with OnBits set and OffBits cleared.
func (s StateSpec) Modify(state State) State {
	return (state & ^s.OffBits) | s.OnBits
}

// StateMapEntry pairs a StateSpec with a value of type T.
type StateMapEntry[T any] struct {
	Spec  StateSpec
	Value T
}

// StateMap is an ordered list of state-value pairs. First match wins.
type StateMap[T any] []StateMapEntry[T]

// Lookup finds the first entry whose spec matches state.
func (m StateMap[T]) Lookup(state State) (T, bool) {
	for _, e := range m {
		if e.Spec.Matches(state) {
			return e.Value, true
		}
	}
	var zero T
	return zero, false
}
