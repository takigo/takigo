package bind

import (
	"reflect"
	"testing"

	"github.com/msorc/takigo/event"
	"github.com/msorc/takigo/platform"
)

// The builders must produce exactly what Parse does for the same pattern,
// or a built sequence would not match (or unbind) a parsed one.
func TestBuildersMatchParse(t *testing.T) {
	tests := []struct {
		pattern string
		built   Sequence
	}{
		{"<Control-s>", Key(ModControl, platform.KeySym('s'))},
		{"<Control-Shift-a>", Key(ModControl|ModShift, platform.KeySym('a'))},
		{"<Button-1>", Button(0, 1)},
		{"<Double-Button-1>", Button(ModDouble, 1)},
		{"<Shift-Button-3>", Button(ModShift, 3)},
		{"<Enter>", On(event.EnterType)},
		{"<<Copy>>", Virtual("Copy")},
		{"<Control-x><Control-s>", Key(ModControl, platform.KeySym('x')).Then(Key(ModControl, platform.KeySym('s')))},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			parsed, err := Parse(tt.pattern)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(parsed, tt.built) {
				t.Errorf("built %+v, parsed %+v", tt.built, parsed)
			}
		})
	}
}

func TestBindTakesSequenceOrString(t *testing.T) {
	e := &Engine{table: NewBindingTable()}
	calls := 0
	h := func(*EventData) bool { calls++; return false }
	if err := e.Bind(".w", Button(ModDouble, 1), h); err != nil {
		t.Fatal(err)
	}
	if err := e.Bind(".w", "<Bogus-Thing", h); err == nil {
		t.Error("Bind of a malformed pattern returned no error")
	}
	if n := len(e.table.Lookup(".w")); n != 1 {
		t.Fatalf("%d bindings, want 1", n)
	}
	// A parsed pattern removes the built binding: they are the same sequence.
	if err := e.Unbind(".w", "<Double-Button-1>"); err != nil {
		t.Fatal(err)
	}
	if n := len(e.table.Lookup(".w")); n != 0 {
		t.Errorf("%d bindings after Unbind, want 0", n)
	}
}
