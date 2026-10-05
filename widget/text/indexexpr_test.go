package text

import (
	"testing"

	"github.com/takigo/takigo/window"
)

// TestParseIndexExpressions checks the Tk index grammar against the results
// Tk gives for the same text. Lines: 1 "hello world", 2 "foo_bar baz",
// 3 "", 4 "last"; the end of the text is 4.4 (Tk's "end -1c").
func TestParseIndexExpressions(t *testing.T) {
	doc := docWithText("hello world\nfoo_bar baz\n\nlast")
	doc.MarkSet("insert", Index{1, 2})
	doc.MarkSet("my mark+1", Index{2, 1})
	doc.TagAdd("hi", Index{1, 6}, Index{1, 11})
	doc.TagAdd("hi", Index{2, 0}, Index{2, 3})
	doc.TagAdd("a.b", Index{3, 0}, Index{4, 2})
	doc.TagConfigure("empty", nil, nil)
	end := Index{4, 4}

	tests := []struct {
		spec string
		want Index
		ok   bool
	}{
		// Bases.
		{"1.0", Index{1, 0}, true},
		{"2.end", Index{2, 11}, true},
		{"1.99", Index{1, 11}, true},
		{"0.3", Index{1, 0}, true},
		{"9.2", end, true},
		{"end", end, true},
		{"insert", Index{1, 2}, true},
		{"my mark+1", Index{2, 1}, true},
		{"hi.first", Index{1, 6}, true},
		{"hi.last", Index{2, 3}, true},
		{"a.b.first", Index{3, 0}, true},
		{"a.b.last", Index{4, 2}, true},

		// Character offsets, with Tk's spellings.
		{"1.0 +1c", Index{1, 1}, true},
		{"1.0+3c", Index{1, 3}, true},
		{"1.0 + 3 chars", Index{1, 3}, true},
		{"1.0 +3 indices", Index{1, 3}, true},
		{"1.0 +3 any chars", Index{1, 3}, true},
		{"1.0 +3 display chars", Index{1, 3}, true},
		{"1.0 +3displaychars", Index{1, 3}, true},
		{"1.10 +2c", Index{2, 0}, true},
		{"2.0 -1c", Index{1, 11}, true},
		{"1.2 + -2c", Index{1, 0}, true},
		{"1.0 -5c", Index{1, 0}, true},
		{"insert +2c", Index{1, 4}, true},
		{"hi.first +1c", Index{1, 7}, true},
		{"end -1c", end, true},
		{"end - 2 chars", Index{4, 3}, true},
		{"end -5c", Index{4, 0}, true},
		{"end -6c", Index{3, 0}, true},
		{"4.3 +1c", end, true},
		{"4.3 +2c", end, true},
		{"4.4 +1c -1c", end, true},
		{"4.3 +5c -2c", Index{4, 3}, true},

		// Lines.
		{"2.3 -1 lines", Index{1, 3}, true},
		{"2.3 +1 lines", Index{3, 0}, true},
		{"1.5 +10 lines", end, true},
		{"1.5 -3 lines", Index{1, 5}, true},
		{"end -1 lines", Index{4, 0}, true},
		{"1.5 +1l", Index{2, 5}, true},

		// Line and word boundaries.
		{"2.4 linestart", Index{2, 0}, true},
		{"2.4 lineend", Index{2, 11}, true},
		{"2.4 lines", Index{2, 0}, true}, // Tk: an abbreviation of linestart
		{"insert lineend -1c", Index{1, 10}, true},
		{"2.2 wordstart", Index{2, 0}, true}, // "_" is a word character
		{"2.2 wordend", Index{2, 7}, true},
		{"2.9 wordstart", Index{2, 8}, true},
		{"1.5 wordend", Index{1, 6}, true},   // not in a word: one character on
		{"1.5 wordstart", Index{1, 5}, true}, // not in a word: stays
		{"1.11 wordend", Index{2, 0}, true},
		{"end wordend", end, true},
		{"end linestart", end, true},
		{"2.4 linestart +2c lineend", Index{2, 11}, true},

		// Errors.
		{"", Index{}, false},
		{"bogus", Index{}, false},
		{"1.x", Index{}, false},
		{"1.0 +", Index{}, false},
		{"1.0 +1q", Index{}, false},
		{"1.0 +1 dlines", Index{}, false},
		{"1.0 bogus", Index{}, false},
		{"1.0 line", Index{}, false}, // shorter than five letters
		{"missing.first", Index{}, false},
		{"empty.first", Index{}, false},
		{"sel.first", Index{}, false},
		{"@1,2", Index{}, false}, // needs a widget
		{"1.0 +1 display lines", Index{}, false},
	}
	for _, tt := range tests {
		got, ok := ParseIndex(doc, tt.spec)
		if ok != tt.ok {
			t.Errorf("ParseIndex(%q) ok = %v, want %v (got %v)", tt.spec, ok, tt.ok, got)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("ParseIndex(%q) = %v, want %v", tt.spec, got, tt.want)
		}
	}
}

// TestWidgetIndexExpressions checks the forms that need a widget: pixel
// positions, embedded window names and display lines. With the 7-pixel
// test font and a 70-pixel widget, "aaaa bbbb cccc dddd" wraps into the
// display lines "aaaa bbbb " (1.0-1.10) and "cccc dddd" (1.10-1.19); line 3
// holds an embedded window at 3.1.
func TestWidgetIndexExpressions(t *testing.T) {
	tw := newBenchWidget(docWithText("aaaa bbbb cccc dddd\nxyz\nw\npqrs"), 70, 200)
	tw.WindowCreate("3.1", &window.Window{PathName: ".t.b"})
	if n := len(tw.layout.line(1).dls); n != 2 {
		t.Fatalf("line 1 has %d display lines, want 2", n)
	}

	tests := []struct {
		spec string
		want Index
		ok   bool
	}{
		{"1.12 display linestart", Index{1, 10}, true},
		{"1.12 display lineend", Index{1, 19}, true},
		{"1.2 display lineend", Index{1, 9}, true},
		{"1.2 display linestart", Index{1, 0}, true},
		{"1.12 linestart", Index{1, 0}, true},
		{"1.2 +1 display lines", Index{1, 12}, true},
		{"1.2 +1displaylines", Index{1, 12}, true},
		{"1.12 -1 display lines", Index{1, 2}, true},
		{"1.12 +1 display lines", Index{2, 2}, true},
		{"1.2 -1 display lines", Index{1, 0}, true}, // Tk: the first display line's start
		{"1.2 +2 display lines", Index{2, 2}, true},
		{"1.2 +9 display lines", Index{4, 2}, true},
		{".t.b", Index{3, 1}, true},
		{".t.b +1c", Index{3, 2}, true},
		{".t.missing", Index{}, false},
	}
	for _, tt := range tests {
		got, ok := tw.index(tt.spec)
		if ok != tt.ok {
			t.Errorf("index(%q) ok = %v, want %v (got %v)", tt.spec, ok, tt.ok, got)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("index(%q) = %v, want %v", tt.spec, got, tt.want)
		}
	}

	// @x,y is the character at that pixel: x=16 is in the third column
	// (14-21), y=0 on the first line.
	if got, ok := tw.index("@16,0"); !ok || got != (Index{1, 2}) {
		t.Errorf("index(@16,0) = %v, %v; want 1.2", got, ok)
	}
	if got, ok := tw.index("@16,0 lineend"); !ok || got != (Index{1, 19}) {
		t.Errorf("index(@16,0 lineend) = %v, %v; want 1.19", got, ok)
	}
	if _, ok := tw.index("@16"); ok {
		t.Error("index(@16) parsed")
	}
}

// FuzzParseIndex checks that any index string either fails or resolves to
// a position inside the document, with or without a widget.
func FuzzParseIndex(f *testing.F) {
	for _, s := range []string{"1.0", "end -1c", "insert +3 display lines", "sel.first wordend",
		"@3,4 lineend", "2.end - 1 lines +2c", "a.b.last", ".t.b +1c", "1.0 + -9999999999c"} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, spec string) {
		tw := newBenchWidget(docWithText("aaaa bbbb cccc dddd\nx_y\n\nz"), 70, 200)
		tw.doc.TagAdd("a.b", Index{1, 2}, Index{2, 1})
		for _, resolve := range []func(string) (Index, bool){
			tw.index,
			func(s string) (Index, bool) { return ParseIndex(tw.doc, s) },
		} {
			idx, ok := resolve(spec)
			if ok && idx != Clamp(idx, tw.doc) {
				t.Fatalf("%q resolved to %v, outside the document", spec, idx)
			}
		}
	})
}

// TestEmbeddedWindowFollowsInsertBefore checks that text inserted at an
// embedded window's position goes before the window and the window keeps
// its place, as with Tcl's "$t insert plot \n" after "window create plot".
func TestEmbeddedWindowFollowsInsertBefore(t *testing.T) {
	tw := newBenchWidget(docWithText("ab"), 200, 100)
	tw.MarkSet("plot", "1.1")
	tw.MarkGravity("plot", GravityLeft)
	w := &window.Window{PathName: ".t.c"}
	tw.WindowCreate("plot", w)
	tw.Insert("plot", "\n")
	if got, ok := tw.index(".t.c"); !ok || got != (Index{2, 0}) {
		t.Fatalf("window at %v, %v after inserting before it; want 2.0", got, ok)
	}
	if got := tw.doc.Get(Index{2, 0}, Index{2, 1}); got != string(runeEmbeddedWindow) {
		t.Errorf("char at the window's index = %q, want the placeholder", got)
	}
	if got, _ := tw.index("plot"); got != (Index{1, 1}) {
		t.Errorf("plot mark at %v, want 1.1", got)
	}
}
