package text

import "slices"

// ActionType is the type of undo action.
type ActionType int

const (
	ActionInsert ActionType = iota
	ActionDelete
)

// UndoAction records a single insert or delete.
type UndoAction struct {
	Type  ActionType
	Start Index
	End   Index
	Text  string
}

// UndoStack manages undo/redo history with auto-grouping.
type UndoStack struct {
	undoStack [][]UndoAction
	redoStack [][]UndoAction
	current   []UndoAction
	maxDepth  int
}

// NewUndoStack creates a new undo stack.
func NewUndoStack(maxDepth int) *UndoStack {
	if maxDepth <= 0 {
		maxDepth = 100
	}
	return &UndoStack{maxDepth: maxDepth}
}

// RecordInsert records an insert action. endIdx is the index after insertion.
func (u *UndoStack) RecordInsert(start, end Index, text string) {
	action := UndoAction{Type: ActionInsert, Start: start, End: end, Text: text}

	// Auto-group: if current group is a single-char insert at adjacent position, merge.
	if len(u.current) == 1 && u.current[0].Type == ActionInsert &&
		len(text) == 1 && text[0] != '\n' &&
		Compare(u.current[0].End, start) == 0 {
		u.current[0].End = end
		u.current[0].Text += text
		u.redoStack = nil
		return
	}

	if len(u.current) > 0 && !u.canMerge(action) {
		u.Separator()
	}
	u.current = append(u.current, action)
	u.redoStack = nil
}

// RecordDelete records a delete action.
func (u *UndoStack) RecordDelete(start, end Index, text string) {
	action := UndoAction{Type: ActionDelete, Start: start, End: end, Text: text}

	// Auto-group: if current group is a single-char backspace at adjacent position, merge.
	if len(u.current) == 1 && u.current[0].Type == ActionDelete &&
		len(text) == 1 && text[0] != '\n' &&
		Compare(start, u.current[0].Start) == 0 {
		// Backspace from same position (forward delete).
		u.current[0].End = end
		u.current[0].Text += text
		u.redoStack = nil
		return
	}
	if len(u.current) == 1 && u.current[0].Type == ActionDelete &&
		len(text) == 1 && text[0] != '\n' &&
		Compare(end, u.current[0].Start) == 0 {
		// Backspace (cursor moving left).
		u.current[0].Start = start
		u.current[0].Text = text + u.current[0].Text
		u.redoStack = nil
		return
	}

	if len(u.current) > 0 && !u.canMerge(action) {
		u.Separator()
	}
	u.current = append(u.current, action)
	u.redoStack = nil
}

// canMerge returns whether the action can be merged into the current group.
func (u *UndoStack) canMerge(action UndoAction) bool {
	if len(u.current) == 0 {
		return true
	}
	// Different action types don't merge.
	return u.current[len(u.current)-1].Type == action.Type
}

// Separator closes the current group, pushing it to the undo stack.
func (u *UndoStack) Separator() {
	if len(u.current) == 0 {
		return
	}
	u.undoStack = append(u.undoStack, u.current)
	u.current = nil
	// Trim to maxDepth.
	if len(u.undoStack) > u.maxDepth {
		u.undoStack = u.undoStack[len(u.undoStack)-u.maxDepth:]
	}
}

// Undo undoes the last group. Returns true if anything was undone.
func (u *UndoStack) Undo(doc *Document) bool {
	// Close any open group first.
	u.Separator()
	if len(u.undoStack) == 0 {
		return false
	}

	group := u.undoStack[len(u.undoStack)-1]
	u.undoStack = u.undoStack[:len(u.undoStack)-1]

	// Apply inverse actions in reverse order.
	var redoGroup []UndoAction
	for _, action := range slices.Backward(group) {
		switch action.Type {
		case ActionInsert:
			// Undo insert = delete
			doc.Delete(action.Start, action.End)
			redoGroup = append(redoGroup, action)
		case ActionDelete:
			// Undo delete = insert; redoing it deletes the text again.
			endIdx := doc.Insert(action.Start, action.Text)
			redoGroup = append(redoGroup, UndoAction{
				Type: ActionDelete, Start: action.Start, End: endIdx, Text: action.Text,
			})
		}
	}
	u.redoStack = append(u.redoStack, redoGroup)
	return true
}

// Redo redoes the last undone group. Returns true if anything was redone.
func (u *UndoStack) Redo(doc *Document) bool {
	if len(u.redoStack) == 0 {
		return false
	}

	group := u.redoStack[len(u.redoStack)-1]
	u.redoStack = u.redoStack[:len(u.redoStack)-1]

	var undoGroup []UndoAction
	// Redo group was stored in reverse order; apply in reverse to get original order.
	for _, action := range slices.Backward(group) {
		switch action.Type {
		case ActionInsert:
			endIdx := doc.Insert(action.Start, action.Text)
			undoGroup = append(undoGroup, UndoAction{
				Type: ActionInsert, Start: action.Start, End: endIdx, Text: action.Text,
			})
		case ActionDelete:
			text := doc.Get(action.Start, action.End)
			doc.Delete(action.Start, action.End)
			undoGroup = append(undoGroup, UndoAction{
				Type: ActionDelete, Start: action.Start, End: action.End, Text: text,
			})
		}
	}
	u.undoStack = append(u.undoStack, undoGroup)
	return true
}

// Reset clears the undo and redo stacks.
func (u *UndoStack) Reset() {
	u.undoStack = nil
	u.redoStack = nil
	u.current = nil
}
