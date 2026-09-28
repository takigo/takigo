package ttk

import (
	"testing"

	"github.com/msorc/takigo/option"
)

func TestStyleLookupDefaults(t *testing.T) {
	s := &Style{
		Name:     "TButton",
		Defaults: map[string]any{"-background": uint64(0xd9d9d9)},
	}
	v, ok := s.Lookup("-background", 0)
	if !ok {
		t.Fatal("expected to find -background")
	}
	if v.(uint64) != 0xd9d9d9 {
		t.Errorf("got %v, want 0xd9d9d9", v)
	}
}

func TestStyleLookupAs(t *testing.T) {
	s := &Style{
		Name:     "TCheckbutton",
		Defaults: map[string]any{"-indicatorrelief": option.ReliefRaised, "-upperbordercolor": uint64(0xffffff)},
	}
	if r, ok := s.LookupAs[option.Relief]("-indicatorrelief", 0); !ok || r != option.ReliefRaised {
		t.Errorf("relief: got %v, %v", r, ok)
	}
	if c, ok := s.LookupAs[uint64]("-upperbordercolor", 0); !ok || c != 0xffffff {
		t.Errorf("color: got %#x, %v", c, ok)
	}
	if _, ok := s.LookupAs[option.Relief]("-upperbordercolor", 0); ok {
		t.Error("wrong type reported found")
	}
	if _, ok := s.LookupAs[uint64]("-missing", 0); ok {
		t.Error("missing option reported found")
	}
	var nilStyle *Style
	if _, ok := nilStyle.LookupAs[uint64]("-upperbordercolor", 0); ok {
		t.Error("nil style reported found")
	}
}

func TestStyleLookupStateMap(t *testing.T) {
	s := &Style{
		Name:     "TButton",
		Defaults: map[string]any{"-background": uint64(0xd9d9d9)},
		Maps: map[string]StateMap[any]{
			"-background": {
				{Spec: StateSpec{OnBits: StateActive}, Value: uint64(0xececec)},
				{Spec: StateSpec{OnBits: StatePressed}, Value: uint64(0xc0c0c0)},
			},
		},
	}

	// Active state uses map value
	v, ok := s.Lookup("-background", StateActive)
	if !ok || v.(uint64) != 0xececec {
		t.Errorf("active: got %v, want 0xececec", v)
	}

	// Pressed state uses map value
	v, ok = s.Lookup("-background", StatePressed)
	if !ok || v.(uint64) != 0xc0c0c0 {
		t.Errorf("pressed: got %v, want 0xc0c0c0", v)
	}

	// Normal state (no map match) falls through to default
	v, ok = s.Lookup("-background", 0)
	if !ok || v.(uint64) != 0xd9d9d9 {
		t.Errorf("normal: got %v, want 0xd9d9d9", v)
	}
}

func TestStyleLookupParentChain(t *testing.T) {
	root := &Style{
		Name:     ".",
		Defaults: map[string]any{"-foreground": uint64(0x000000)},
	}
	child := &Style{
		Name:     "TButton",
		Parent:   root,
		Defaults: map[string]any{"-background": uint64(0xd9d9d9)},
	}

	// Child has its own default
	v, ok := child.Lookup("-background", 0)
	if !ok || v.(uint64) != 0xd9d9d9 {
		t.Errorf("child -background: got %v", v)
	}

	// Child inherits from parent
	v, ok = child.Lookup("-foreground", 0)
	if !ok || v.(uint64) != 0x000000 {
		t.Errorf("inherited -foreground: got %v", v)
	}

	// Missing option
	_, ok = child.Lookup("-nonexistent", 0)
	if ok {
		t.Error("should not find -nonexistent")
	}
}

func TestLookupColorFallback(t *testing.T) {
	s := &Style{Name: "TButton", Defaults: map[string]any{}}
	got := LookupColor(s, "-background", 0, 0xFF0000)
	if got != 0xFF0000 {
		t.Errorf("fallback: got 0x%x, want 0xFF0000", got)
	}
}

func TestLookupColorWrongType(t *testing.T) {
	s := &Style{
		Name:     "TButton",
		Defaults: map[string]any{"-background": "not-a-color"},
	}
	got := LookupColor(s, "-background", 0, 0xFF0000)
	if got != 0xFF0000 {
		t.Errorf("wrong type should fall back: got 0x%x", got)
	}
}

func TestLookupInt(t *testing.T) {
	s := &Style{
		Name:     "TButton",
		Defaults: map[string]any{"-borderwidth": 2},
	}
	got := LookupInt(s, "-borderwidth", 0, 0)
	if got != 2 {
		t.Errorf("got %d, want 2", got)
	}
	got = LookupInt(s, "-missing", 0, 99)
	if got != 99 {
		t.Errorf("missing should fallback: got %d, want 99", got)
	}
}
