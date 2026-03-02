package text

import (
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	// Should have 1 empty line.
	if doc.LineCount() != 1 {
		t.Errorf("LineCount = %d, want 1", doc.LineCount())
	}
	if len(doc.Lines[0].Text) != 0 {
		t.Errorf("first line not empty: %q", string(doc.Lines[0].Text))
	}

	// Should have built-in marks.
	if pos := doc.MarkPos("insert"); pos == nil {
		t.Error("insert mark missing")
	} else if *pos != (Index{1, 0}) {
		t.Errorf("insert mark = %v, want {1,0}", *pos)
	}
	if pos := doc.MarkPos("current"); pos == nil {
		t.Error("current mark missing")
	}
}

func TestDocumentEndIndex(t *testing.T) {
	doc := NewDocument()
	if got := doc.EndIndex(); got != (Index{1, 0}) {
		t.Errorf("empty EndIndex = %v, want {1,0}", got)
	}

	doc.Lines[0].Text = []rune("hello")
	if got := doc.EndIndex(); got != (Index{1, 5}) {
		t.Errorf("EndIndex = %v, want {1,5}", got)
	}
}

func TestDocumentInsertSingleChar(t *testing.T) {
	doc := NewDocument()
	end := doc.Insert(Index{1, 0}, "a")
	if end != (Index{1, 1}) {
		t.Errorf("Insert end = %v, want {1,1}", end)
	}
	if s := doc.Get(Index{1, 0}, Index{1, 1}); s != "a" {
		t.Errorf("Get = %q, want \"a\"", s)
	}
}

func TestDocumentInsertMultiChar(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	if s := doc.Get(Index{1, 0}, Index{1, 5}); s != "hello" {
		t.Errorf("Get = %q, want \"hello\"", s)
	}
}

func TestDocumentInsertWithNewlines(t *testing.T) {
	doc := NewDocument()
	end := doc.Insert(Index{1, 0}, "abc\ndef\nghi")
	if doc.LineCount() != 3 {
		t.Fatalf("LineCount = %d, want 3", doc.LineCount())
	}
	if end != (Index{3, 3}) {
		t.Errorf("Insert end = %v, want {3,3}", end)
	}
	if s := string(doc.Lines[0].Text); s != "abc" {
		t.Errorf("line 1 = %q, want \"abc\"", s)
	}
	if s := string(doc.Lines[1].Text); s != "def" {
		t.Errorf("line 2 = %q, want \"def\"", s)
	}
	if s := string(doc.Lines[2].Text); s != "ghi" {
		t.Errorf("line 3 = %q, want \"ghi\"", s)
	}
}

func TestDocumentInsertAtMiddle(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	doc.Insert(Index{1, 2}, "XY")
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "heXYllo" {
		t.Errorf("Get = %q, want \"heXYllo\"", s)
	}
}

func TestDocumentInsertAtEnd(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "abc")
	doc.Insert(Index{1, 3}, "def")
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "abcdef" {
		t.Errorf("Get = %q, want \"abcdef\"", s)
	}
}

func TestDocumentDeleteSameLine(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	doc.Delete(Index{1, 1}, Index{1, 3}) // delete "el"
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "hlo" {
		t.Errorf("after delete = %q, want \"hlo\"", s)
	}
}

func TestDocumentDeleteCrossLine(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "abc\ndef\nghi")
	doc.Delete(Index{1, 2}, Index{3, 1}) // delete "c\ndef\ng"
	if doc.LineCount() != 1 {
		t.Fatalf("LineCount = %d, want 1", doc.LineCount())
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "abhi" {
		t.Errorf("after cross-line delete = %q, want \"abhi\"", s)
	}
}

func TestDocumentDeleteFullLine(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "abc\ndef\nghi")
	doc.Delete(Index{1, 3}, Index{2, 3}) // delete "\ndef"
	if doc.LineCount() != 2 {
		t.Fatalf("LineCount = %d, want 2", doc.LineCount())
	}
	if s := doc.Get(Index{1, 0}, Index{1, 3}); s != "abc" {
		t.Errorf("line 1 = %q, want \"abc\"", s)
	}
}

func TestDocumentGetSameLine(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello world")
	got := doc.Get(Index{1, 0}, Index{1, 5})
	if got != "hello" {
		t.Errorf("Get same line = %q, want \"hello\"", got)
	}
}

func TestDocumentGetCrossLine(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "abc\ndef")
	got := doc.Get(Index{1, 0}, Index{2, 3})
	if got != "abc\ndef" {
		t.Errorf("Get cross line = %q, want \"abc\\ndef\"", got)
	}
}

func TestDocumentGetEmpty(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	got := doc.Get(Index{1, 3}, Index{1, 3}) // empty range
	if got != "" {
		t.Errorf("Get empty = %q, want \"\"", got)
	}
}

func TestDocumentGetReversed(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	got := doc.Get(Index{1, 5}, Index{1, 0}) // reversed range
	if got != "" {
		t.Errorf("Get reversed = %q, want \"\"", got)
	}
}

func TestMarkAdjustOnInsert(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	doc.MarkSet("mymark", Index{1, 3})

	// Insert before the mark.
	doc.Insert(Index{1, 1}, "XX")
	pos := doc.MarkPos("mymark")
	if pos == nil {
		t.Fatal("mark missing")
	}
	// Mark should shift right by 2.
	if *pos != (Index{1, 5}) {
		t.Errorf("mark after insert = %v, want {1,5}", *pos)
	}
}

func TestMarkAdjustOnDelete(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	doc.MarkSet("mymark", Index{1, 4})

	// Delete range before mark.
	doc.Delete(Index{1, 1}, Index{1, 3}) // delete "el"
	pos := doc.MarkPos("mymark")
	if pos == nil {
		t.Fatal("mark missing")
	}
	// Mark should shift left by 2.
	if *pos != (Index{1, 2}) {
		t.Errorf("mark after delete = %v, want {1,2}", *pos)
	}
}

func TestMarkDeleteWithinRange(t *testing.T) {
	doc := NewDocument()
	doc.Insert(Index{1, 0}, "hello")
	doc.MarkSet("mymark", Index{1, 2})

	// Delete range that includes the mark.
	doc.Delete(Index{1, 1}, Index{1, 4})
	pos := doc.MarkPos("mymark")
	if pos == nil {
		t.Fatal("mark missing")
	}
	// Mark should move to delete start.
	if *pos != (Index{1, 1}) {
		t.Errorf("mark after encompassing delete = %v, want {1,1}", *pos)
	}
}

func TestMarkUnsetBuiltin(t *testing.T) {
	doc := NewDocument()
	doc.MarkUnset("insert")
	doc.MarkUnset("current")
	if doc.MarkPos("insert") == nil {
		t.Error("insert mark should not be removable")
	}
	if doc.MarkPos("current") == nil {
		t.Error("current mark should not be removable")
	}
}

func TestMarkUnsetCustom(t *testing.T) {
	doc := NewDocument()
	doc.MarkSet("mymark", Index{1, 0})
	doc.MarkUnset("mymark")
	if doc.MarkPos("mymark") != nil {
		t.Error("custom mark should be removable")
	}
}
