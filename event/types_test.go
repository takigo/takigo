package event

import "testing"

func TestEventTypesDistinct(t *testing.T) {
	types := []Type{
		KeyPressType, KeyReleaseType, ButtonPressType, ButtonReleaseType,
		MotionType, EnterType, LeaveType, FocusInType, FocusOutType,
		ExposeType, DestroyType, UnmapType, MapType, ConfigureType,
		ClientMessageType, PropertyType,
	}
	seen := make(map[Type]bool)
	for _, typ := range types {
		if typ == 0 {
			t.Errorf("event type should be non-zero")
		}
		if seen[typ] {
			t.Errorf("duplicate event type value %d", typ)
		}
		seen[typ] = true
	}
}

func TestEventMasksDistinct(t *testing.T) {
	masks := []Mask{
		KeyPressMask, KeyReleaseMask, ButtonPressMask, ButtonReleaseMask,
		MotionMask, EnterMask, LeaveMask, FocusChangeMask,
		ExposureMask, StructureNotifyMask, PropertyChangeMask, ClientMessageMask,
	}
	for i, m := range masks {
		if m == 0 {
			t.Errorf("mask %d should be non-zero", i)
		}
		// Each mask should be a single bit.
		if m&(m-1) != 0 {
			t.Errorf("mask %d = %b is not a single bit", i, m)
		}
	}
}

func TestTypeToMask(t *testing.T) {
	if TypeToMask(KeyPressType) != KeyPressMask {
		t.Error("KeyPressType mask mismatch")
	}
	if TypeToMask(ButtonPressType) != ButtonPressMask {
		t.Error("ButtonPressType mask mismatch")
	}
	if TypeToMask(ExposeType) != ExposureMask {
		t.Error("ExposeType mask mismatch")
	}
	// DestroyType maps to StructureNotifyMask.
	if TypeToMask(DestroyType) != StructureNotifyMask {
		t.Error("DestroyType mask mismatch")
	}
	// FocusIn and FocusOut map to same mask.
	if TypeToMask(FocusInType) != TypeToMask(FocusOutType) {
		t.Error("FocusIn/FocusOut should share mask")
	}
	// Unknown type returns 0.
	if TypeToMask(Type(999)) != 0 {
		t.Error("unknown type should return 0")
	}
}

func TestAllEventsMask(t *testing.T) {
	masks := []Mask{
		KeyPressMask, KeyReleaseMask, ButtonPressMask, ButtonReleaseMask,
		MotionMask, EnterMask, LeaveMask, FocusChangeMask,
		ExposureMask, StructureNotifyMask, PropertyChangeMask, ClientMessageMask,
	}
	var combined Mask
	for _, m := range masks {
		combined |= m
	}
	if AllEventsMask&combined != combined {
		t.Errorf("AllEventsMask (%b) doesn't cover all masks (%b)", AllEventsMask, combined)
	}
}
