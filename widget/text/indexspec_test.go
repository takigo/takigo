package text

import (
	"errors"
	"testing"
)

func TestIndexSpecs(t *testing.T) {
	tw := newBenchWidget(docWithText("hello\nworld"), 200, 100)

	if err := tw.Insert(Index{Line: 1, Char: 5}, ","); err != nil {
		t.Fatalf("Insert(Index) = %v", err)
	}
	if err := tw.Insert("2.end", "!"); err != nil {
		t.Fatalf("Insert(string) = %v", err)
	}
	if got := tw.Get("1.0", tw.EndIndex()); got != "hello,\nworld!" {
		t.Errorf("content = %q", got)
	}
	if got := tw.Get(Index{Line: 2}, "2.5"); got != "world" {
		t.Errorf("Get(Index, string) = %q, want world", got)
	}

	if err := tw.Insert("no such mark", "x"); !errors.Is(err, ErrBadIndex) {
		t.Errorf("Insert(bad index) = %v, want ErrBadIndex", err)
	}
	if err := tw.Delete("1.0", "bogus+"); !errors.Is(err, ErrBadIndex) {
		t.Errorf("Delete(bad index) = %v, want ErrBadIndex", err)
	}
	if err := tw.TagAdd("t", "bogus", "end"); !errors.Is(err, ErrBadIndex) {
		t.Errorf("TagAdd(bad index) = %v, want ErrBadIndex", err)
	}
	if got := tw.Get("1.0", "end -1c"); got != "hello,\nworld!" {
		t.Errorf("a failed edit changed the content: %q", got)
	}

	idx, err := tw.Index("1.0 lineend")
	if err != nil || idx != (Index{Line: 1, Char: 6}) {
		t.Errorf(`Index("1.0 lineend") = %v, %v; want 1.6`, idx, err)
	}
	if _, err := tw.Index("@"); !errors.Is(err, ErrBadIndex) {
		t.Errorf("Index(bad) = %v, want ErrBadIndex", err)
	}
	if got := idx.String(); got != "1.6" {
		t.Errorf("String() = %q, want 1.6", got)
	}
	// An Index beyond the content is clamped, as the document does.
	if idx, _ := tw.Index(Index{Line: 99, Char: 99}); idx != tw.EndIndex() {
		t.Errorf("Index(out of range) = %v, want %v", idx, tw.EndIndex())
	}
}
