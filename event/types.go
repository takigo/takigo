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
	VirtualType
)

// Virtual events a backend sends while an input method composes text, as
// tkMacOSXKeyEvent.c does: the composition is inserted as ordinary key
// presses between IMEStart and IMEEnd, and IMEClear deletes it again before
// the next composition or the committed text. AccentBackspace erases the
// character an accent menu replaces.
const (
	IMEStart        = "TkStartIMEMarkedText"
	IMEEnd          = "TkEndIMEMarkedText"
	IMEClear        = "TkClearIMEMarkedText"
	AccentBackspace = "TkAccentBackspace"
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
	VirtualMask

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
	case VirtualType:
		return VirtualMask
	default:
		return 0
	}
}
