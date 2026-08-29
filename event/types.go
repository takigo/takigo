// Package event provides the event system for takigo, including event types,
// handler dispatch, and the main event loop with idle/timer support.
package event

// Type identifies the kind of event.
type Type int

const (
	KeyPressType Type = iota + 1
	KeyReleaseType
	ButtonPressType
	ButtonReleaseType
	MotionType
	EnterType
	LeaveType
	FocusInType
	FocusOutType
	ExposeType
	DestroyType
	UnmapType
	MapType
	ConfigureType
	ClientMessageType
	PropertyType
)

// Mask is a bitmask for selecting event types.
type Mask uint64

const (
	KeyPressMask Mask = 1 << iota
	KeyReleaseMask
	ButtonPressMask
	ButtonReleaseMask
	MotionMask
	EnterMask
	LeaveMask
	FocusChangeMask
	ExposureMask
	StructureNotifyMask
	PropertyChangeMask
	ClientMessageMask

	AllEventsMask Mask = (1 << iota) - 1
)

// TypeToMask maps an event type to its corresponding mask.
func TypeToMask(t Type) Mask {
	switch t {
	case KeyPressType:
		return KeyPressMask
	case KeyReleaseType:
		return KeyReleaseMask
	case ButtonPressType:
		return ButtonPressMask
	case ButtonReleaseType:
		return ButtonReleaseMask
	case MotionType:
		return MotionMask
	case EnterType:
		return EnterMask
	case LeaveType:
		return LeaveMask
	case FocusInType, FocusOutType:
		return FocusChangeMask
	case ExposeType:
		return ExposureMask
	case DestroyType, UnmapType, MapType, ConfigureType:
		return StructureNotifyMask
	case PropertyType:
		return PropertyChangeMask
	case ClientMessageType:
		return ClientMessageMask
	default:
		return 0
	}
}
