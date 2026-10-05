// Package bind implements the Tk binding system.
package bind

import (
	"testing"

	"github.com/takigo/takigo/event"
	"github.com/takigo/takigo/platform"
)

// FuzzParse fuzzes the Parse function with arbitrary input.
func FuzzParse(f *testing.F) {
	// Seed with known valid patterns
	seeds := []string{
		"<Button-1>",
		"<ButtonRelease-1>",
		"<Key-a>",
		"<Key-Return>",
		"<KeyRelease-Escape>",
		"<Control-a>",
		"<Control-Shift-x>",
		"<Double-Button-1>",
		"<Triple-Button-1>",
		"<Motion>",
		"<Enter>",
		"<Leave>",
		"<FocusIn>",
		"<FocusOut>",
		"<Configure>",
		"<Expose>",
		"<Destroy>",
		"<Map>",
		"<Unmap>",
		"<<VirtualName>>",
		"a",
		"<>",
		"<invalid",
		"<Control-Button-1>",
		"<Alt-Key-Tab>",
		"<Meta-Motion>",
		"<Double-Key-a>",
		"<Triple-Button-2>",
		"<<Copy>>",
		"<<Paste>>",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		_, _ = Parse(input)
	})
}

// FuzzParseSequenceString fuzzes the Sequence.String method.
func FuzzParseSequenceString(f *testing.F) {
	seeds := []string{
		"<Button-1>",
		"<Control-a>",
		"<<Virtual>>",
		"<Double-Button-1>",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, input string) {
		seq, err := Parse(input)
		if err == nil {
			_ = seq.String()
		}
	})
}

// FuzzPatternMatches fuzzes the Pattern.matches method.
func FuzzPatternMatches(f *testing.F) {
	// This requires constructing events, so we use a simpler approach
	f.Fuzz(func(t *testing.T, pattern string, eventType int, state int, button int, keysym int) {
		seq, err := Parse(pattern)
		if err != nil {
			return
		}
		if len(seq.Patterns) == 0 {
			return
		}
		// Create a mock event
		ev := &event.Event{
			Type:   event.Type(eventType % 20),
			State:  uint(state),
			Button: uint(button % 6),
			KeySym: platform.KeySym(keysym),
		}
		// Just test that matches doesn't panic
		for _, p := range seq.Patterns {
			_ = p.matches(ev, 0)
		}
	})
}

// FuzzPatternSpecificity fuzzes the Pattern.specificity method.
func FuzzPatternSpecificity(f *testing.F) {
	seeds := []string{
		"<Button-1>",
		"<Control-a>",
		"<Control-Shift-x>",
		"<<Virtual>>",
		"<Double-Button-1>",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, pattern string) {
		seq, err := Parse(pattern)
		if err != nil {
			return
		}
		for _, p := range seq.Patterns {
			_ = p.specificity()
		}
	})
}
