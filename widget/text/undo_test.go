package text

import "testing"

func TestUndoRecordInsert(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	end := doc.Insert(Index{1, 0}, "hello")
	u.RecordInsert(Index{1, 0}, end, "hello")
	u.Separator()

	// Undo should remove "hello".
	if !u.Undo(doc) {
		t.Fatal("Undo returned false")
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "" {
		t.Errorf("after undo = %q, want empty", s)
	}
}

func TestUndoRecordDelete(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	doc.Insert(Index{1, 0}, "hello")
	u.Separator() // don't group with deletes

	text := doc.Get(Index{1, 1}, Index{1, 4})
	doc.Delete(Index{1, 1}, Index{1, 4})
	u.RecordDelete(Index{1, 1}, Index{1, 4}, text)
	u.Separator()

	// Undo should restore "ell".
	if !u.Undo(doc) {
		t.Fatal("Undo returned false")
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "hello" {
		t.Errorf("after undo = %q, want \"hello\"", s)
	}
}

func TestUndoAutoGroupInsert(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	// Type characters one at a time — should auto-group.
	for i, ch := range "abc" {
		start := Index{1, i}
		end := doc.Insert(start, string(ch))
		u.RecordInsert(start, end, string(ch))
	}
	u.Separator()

	// Single undo should remove all three characters.
	if !u.Undo(doc) {
		t.Fatal("Undo returned false")
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "" {
		t.Errorf("after undo = %q, want empty", s)
	}
}

func TestUndoAutoGroupBackspace(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	doc.Insert(Index{1, 0}, "abcd")

	// Backspace 'd', 'c', 'b' — deleting one char at a time from the right.
	for i := 3; i >= 1; i-- {
		text := doc.Get(Index{1, i}, Index{1, i + 1})
		doc.Delete(Index{1, i}, Index{1, i + 1})
		u.RecordDelete(Index{1, i}, Index{1, i + 1}, text)
	}
	u.Separator()

	// Single undo should restore "bcd".
	if !u.Undo(doc) {
		t.Fatal("Undo returned false")
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "abcd" {
		t.Errorf("after undo = %q, want \"abcd\"", s)
	}
}

func TestUndoSeparatorForcesNewGroup(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	end := doc.Insert(Index{1, 0}, "a")
	u.RecordInsert(Index{1, 0}, end, "a")
	u.Separator()

	end = doc.Insert(Index{1, 1}, "b")
	u.RecordInsert(Index{1, 1}, end, "b")
	u.Separator()

	// First undo removes "b".
	u.Undo(doc)
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "a" {
		t.Errorf("after first undo = %q, want \"a\"", s)
	}

	// Second undo removes "a".
	u.Undo(doc)
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "" {
		t.Errorf("after second undo = %q, want empty", s)
	}
}

func TestUndoRedoRoundTrip(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	end := doc.Insert(Index{1, 0}, "hello")
	u.RecordInsert(Index{1, 0}, end, "hello")
	u.Separator()

	u.Undo(doc)
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "" {
		t.Fatalf("after undo = %q", s)
	}

	if !u.Redo(doc) {
		t.Fatal("Redo returned false")
	}
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "hello" {
		t.Errorf("after redo = %q, want \"hello\"", s)
	}
}

func TestUndoMaxDepth(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(3)

	// Push 5 groups.
	for i := range 5 {
		ch := string(rune('a' + i))
		end := doc.Insert(doc.EndIndex(), ch)
		u.RecordInsert(doc.EndIndex(), end, ch)
		u.Separator()
	}

	// Only 3 undos should succeed.
	count := 0
	for u.Undo(doc) {
		count++
	}
	if count != 3 {
		t.Errorf("undo count = %d, want 3", count)
	}
}

func TestUndoEmptyStack(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)
	if u.Undo(doc) {
		t.Error("Undo on empty stack should return false")
	}
}

func TestRedoEmptyStack(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)
	if u.Redo(doc) {
		t.Error("Redo on empty stack should return false")
	}
}

func TestRedoClearedByNewAction(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	end := doc.Insert(Index{1, 0}, "a")
	u.RecordInsert(Index{1, 0}, end, "a")
	u.Separator()

	u.Undo(doc)

	// New action should clear redo stack.
	end = doc.Insert(Index{1, 0}, "b")
	u.RecordInsert(Index{1, 0}, end, "b")
	u.Separator()

	if u.Redo(doc) {
		t.Error("Redo should return false after new action")
	}
}

func TestNewlineBreaksAutoGroup(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(100)

	end := doc.Insert(Index{1, 0}, "a")
	u.RecordInsert(Index{1, 0}, end, "a")

	// Newline should not merge with previous.
	end2 := doc.Insert(end, "\n")
	u.RecordInsert(end, end2, "\n")
	u.Separator()

	// Undo should undo the whole group (both actions are in same group since canMerge allows same type).
	u.Undo(doc)
	if s := doc.Get(Index{1, 0}, doc.EndIndex()); s != "" {
		t.Errorf("after undo = %q, want empty", s)
	}
}

// TestRedoUndoneDelete checks that redoing an undone deletion deletes again.
func TestRedoUndoneDelete(t *testing.T) {
	doc := docWithText("abc")
	u := NewUndoStack(0)
	u.RecordDelete(Index{1, 1}, Index{1, 2}, "b")
	doc.Delete(Index{1, 1}, Index{1, 2})
	text := func() string { return doc.Get(Index{1, 0}, doc.EndIndex()) }

	steps := []struct {
		do   func(*Document) bool
		want string
	}{
		{u.Undo, "abc"},
		{u.Redo, "ac"},
		{u.Undo, "abc"},
		{u.Redo, "ac"},
	}
	for i, s := range steps {
		if !s.do(doc) {
			t.Fatalf("step %d did nothing", i)
		}
		if got := text(); got != s.want {
			t.Fatalf("step %d: text = %q, want %q", i, got, s.want)
		}
	}
}
