package text

import (
	"testing"

	"github.com/takigo/takigo/window"
)

// newTestPeer builds a text widget without a display, sharing the given
// document the way New and NewPeer do.
func newTestPeer(doc *Document, w, h int) (*TextWidget, func()) {
	t := &TextWidget{
		doc:      doc,
		topLine:  1,
		wrapMode: WrapWord,
	}
	t.Win = &window.Window{Width: w, Height: h}
	t.Font = fixedFont{}
	t.doc.putTag(&Tag{Name: "sel", Priority: selPriority})
	return t, t.shareDocument(doc)
}

// TestPeerEditInvalidatesPrimaryLayout checks that an edit in a peer
// invalidates the primary's layout cache: the two share the document, and
// Tk redraws every peer of a shared text when the text changes.
func TestPeerEditInvalidatesPrimaryLayout(t *testing.T) {
	doc := NewDocument()
	primary, _ := newTestPeer(doc, 400, 100)
	peer, _ := newTestPeer(doc, 400, 100)

	// Lay the primary out, so its cache is valid.
	primary.layout.ensureAll()
	if !primary.layout.lines[0].valid {
		t.Fatal("the primary's layout is not valid")
	}

	peer.Insert("1.0", "hello")

	if primary.layout.lines[0].valid {
		t.Error("the peer's edit did not invalidate the primary's layout")
	}
}

// TestPeerEditRequestsPrimaryRepaint checks that an edit in a peer asks the
// primary for a repaint; without it the primary shows stale content.
func TestPeerEditRequestsPrimaryRepaint(t *testing.T) {
	doc := NewDocument()
	primary, _ := newTestPeer(doc, 400, 100)
	peer, _ := newTestPeer(doc, 400, 100)

	primary.YScrollCmd = func(float64, float64) {}
	peer.Insert("1.0", "hello")

	if !primary.yScrollPending {
		t.Error("the peer's edit did not notify the primary")
	}
}

// TestUnsubscribedWidgetIsIgnored checks that a widget that unsubscribed
// from the document is no longer notified: the document must not keep a
// destroyed widget alive and run its work on every later edit.
func TestUnsubscribedWidgetIsIgnored(t *testing.T) {
	doc := NewDocument()
	primary, _ := newTestPeer(doc, 400, 100)
	peer, unsubPeer := newTestPeer(doc, 400, 100)

	primary.layout.ensureAll()
	peer.layout.ensureAll()
	if !peer.layout.lines[0].valid {
		t.Fatal("the peer's layout is not valid")
	}

	unsubPeer()
	primary.Insert("1.0", "x")

	if !peer.layout.lines[0].valid {
		t.Error("an unsubscribed widget was still notified")
	}
}

// TestPeersShareUndoHistory checks that the peers of a document share one
// undo history, as Tk's peers share the undo stack of the shared text
// (TkSharedText.undoStack): an undo in the primary undoes what a peer
// did, and an undo in a peer undoes the shared history.
func TestPeersShareUndoHistory(t *testing.T) {
	doc := NewDocument()
	primary, _ := newTestPeer(doc, 400, 100)
	peer, _ := newTestPeer(doc, 400, 100)

	if primary.undoStack != peer.undoStack {
		t.Fatal("the peers do not share the undo history")
	}

	peer.Insert("1.0", "hello")
	if got := primary.Get("1.0", "end"); got != "hello" {
		t.Fatalf("the primary shows %q, want %q", got, "hello")
	}

	// The primary undoes the peer's edit.
	primary.Edit("undo")
	if got := primary.Get("1.0", "end"); got != "" {
		t.Errorf("after the primary's undo the text is %q, want empty", got)
	}
	if got := peer.Get("1.0", "end"); got != "" {
		t.Errorf("after the primary's undo the peer shows %q, want empty", got)
	}

	// The peer undoes the primary's edit.
	primary.Insert("1.0", "abc")
	peer.Edit("undo")
	if got := peer.Get("1.0", "end"); got != "" {
		t.Errorf("after the peer's undo the text is %q, want empty", got)
	}
	if got := primary.Get("1.0", "end"); got != "" {
		t.Errorf("after the peer's undo the primary shows %q, want empty", got)
	}
}

// TestDocumentUndoStackIsShared checks that a document has one undo
// history, whether it is reached through the document or through a widget.
func TestDocumentUndoStackIsShared(t *testing.T) {
	doc := NewDocument()
	if doc.UndoStack() != doc.UndoStack() {
		t.Error("the document returned two different undo stacks")
	}
	widget, _ := newTestPeer(doc, 400, 100)
	if widget.undoStack != doc.UndoStack() {
		t.Error("the widget does not use the document's undo stack")
	}
}

// TestDisabledUndoStack checks that a stack with a negative depth records
// nothing and undoes nothing.
func TestDisabledUndoStack(t *testing.T) {
	doc := NewDocument()
	u := NewUndoStack(-1)

	u.RecordInsert(Index{1, 0}, Index{1, 1}, "a")
	u.Separator()
	if u.Undo(doc) {
		t.Error("a disabled stack undid something")
	}
	if u.Redo(doc) {
		t.Error("a disabled stack redid something")
	}
	if len(u.undoStack) != 0 {
		t.Errorf("a disabled stack kept %d groups", len(u.undoStack))
	}
}
