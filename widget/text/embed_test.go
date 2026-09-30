package text

import (
	"testing"

	"github.com/msorc/takigo/platform"
	"github.com/msorc/takigo/widget"
	"github.com/msorc/takigo/window"
)

// testImage is a display-free embedded image.
type testImage struct {
	name string
	w, h int
}

func (i testImage) Name() string { return i.name }
func (i testImage) Width() int   { return i.w }
func (i testImage) Height() int  { return i.h }
func (testImage) Draw(platform.DisplayServer, platform.DrawableID, platform.GCID, int, int, int, int, int, int, int, uint64) {
}

func imagePos(t *testing.T, tw *TextWidget, name string) (Index, bool) {
	t.Helper()
	return tw.doc.imageIndex(name)
}

// TestEmbeddedImageTracksEdits checks that an embedded image takes one
// index position and moves with the text, as in Tk: inserting before it or
// at its position shifts it, inserting after it doesn't, and deleting its
// position removes it.
func TestEmbeddedImageTracksEdits(t *testing.T) {
	tw := newBenchWidget(docWithText("abcd"), 400, 100)
	tw.ImageCreate("1.2", testImage{"pic", 30, 40})

	if got := tw.doc.Get(Index{1, 0}, tw.doc.EndIndex()); got != "ab"+string(runeEmbeddedWindow)+"cd" {
		t.Fatalf("document text = %q, want the image's placeholder at 1.2", got)
	}
	// As in Tk, Get and the selection return only the characters.
	if got := tw.Get("1.0", "end"); got != "abcd" {
		t.Errorf(`Get("1.0", "end") = %q, want "abcd"`, got)
	}
	tw.SelectAll()
	if got := tw.GetSelection(); got != "abcd" {
		t.Errorf("GetSelection() = %q, want %q", got, "abcd")
	}
	tw.clearSelection()
	steps := []struct {
		name string
		do   func()
		want Index
	}{
		{"created", func() {}, Index{1, 2}},
		{"insert before", func() { tw.Insert("1.0", "xx") }, Index{1, 4}},
		{"insert at its position", func() { tw.Insert("1.4", "Z") }, Index{1, 5}},
		{"insert after", func() { tw.Insert("1.6", "Y") }, Index{1, 5}},
		{"line break before", func() { tw.Insert("1.1", "\n") }, Index{2, 4}},
	}
	for _, s := range steps {
		s.do()
		got, ok := imagePos(t, tw, "pic")
		if !ok || got != s.want {
			t.Fatalf("%s: image at %v (%v), want %v", s.name, got, ok, s.want)
		}
		if img := tw.doc.imageAt(s.want); img == nil {
			t.Fatalf("%s: no image at %v", s.name, s.want)
		}
	}
	if got, ok := tw.index("pic +1c"); !ok || got != (Index{2, 5}) {
		t.Errorf(`index("pic +1c") = %v, %v; want 2.5`, got, ok)
	}

	tw.Delete("2.4", "2.5")
	if _, ok := imagePos(t, tw, "pic"); ok {
		t.Error("image still indexed after deleting its position")
	}
	if n := countImages(tw.doc); n != 0 {
		t.Errorf("%d images still laid out after deleting the placeholder", n)
	}
}

func countImages(d *Document) int {
	n := 0
	d.eachImage(func(Index, widget.WidgetImage) { n++ })
	return n
}

// TestEmbeddedImageLayout checks that an image occupies its own width in the
// line, makes the display line tall enough, and is laid out by peers too.
func TestEmbeddedImageLayout(t *testing.T) {
	doc := docWithText("ab cd")
	tw := newBenchWidget(doc, 400, 100)
	peer := newBenchWidget(doc, 400, 100)
	tw.ImageCreate("1.3", testImage{"pic", 30, 40})

	for name, w := range map[string]*TextWidget{"widget": tw, "peer": peer} {
		// 5 characters at 7px plus the 30px image.
		if got := w.measureRange(1, 0, 6); got != 5*7+30 {
			t.Errorf("%s: line width = %d, want %d", name, got, 5*7+30)
		}
		segs := w.segmentsForRange(1, 0, 6)
		var img *textSegment
		for i := range segs {
			if segs[i].img != nil {
				img = &segs[i]
			}
		}
		if img == nil {
			t.Fatalf("%s: no image segment in %+v", name, segs)
		}
		if img.x != 3*7 || img.width != 30 || img.runes != 1 {
			t.Errorf("%s: image segment x=%d width=%d runes=%d, want x=21 width=30 runes=1",
				name, img.x, img.width, img.runes)
		}
		if h := w.layout.line(1).dls[0].height; h < 40 {
			t.Errorf("%s: display line height %d, want at least the image's 40", name, h)
		}
	}

	// Pixels past the image map to the text after it: x=53 is in the left
	// half of "c" (51-58), so the index before it.
	if got := tw.indexFromPixel(53, 5); got != (Index{1, 4}) {
		t.Errorf("indexFromPixel(53, 5) = %v, want 1.4", got)
	}
}

// TestDeletingEmbeddedObjects checks what deleting an embedded window's or
// image's position does: as in Tk the window is destroyed, and an adjacent
// object is unaffected.
func TestDeletingEmbeddedObjects(t *testing.T) {
	root := &window.Window{PathName: ".", Display: &window.Display{}}
	tw := newBenchWidget(docWithText("abc"), 400, 100)
	win := window.NewChildWindow(root, "b", 0, 0, 10, 10)
	tw.WindowCreate("1.1", win)
	tw.ImageCreate("1.2", testImage{"one", 10, 10})
	tw.ImageCreate("1.3", testImage{"two", 10, 10})
	// a W one two b c

	tw.Delete("1.2", "1.3") // image "one"
	if _, ok := tw.doc.imageIndex("one"); ok {
		t.Error(`image "one" still there after deleting it`)
	}
	if got, ok := tw.doc.imageIndex("two"); !ok || got != (Index{1, 2}) {
		t.Errorf(`image "two" at %v, %v; want 1.2`, got, ok)
	}
	if win.IsDestroyed() || len(tw.embeddedWindows) != 1 {
		t.Fatal("deleting an image affected the window")
	}

	tw.Delete("1.0", "1.2") // "a" and the window
	if !win.IsDestroyed() {
		t.Error("window not destroyed with its position")
	}
	if len(tw.embeddedWindows) != 0 {
		t.Errorf("%d embedded windows left", len(tw.embeddedWindows))
	}
	if got, ok := tw.doc.imageIndex("two"); !ok || got != (Index{1, 0}) {
		t.Errorf(`image "two" at %v, %v; want 1.0`, got, ok)
	}
	if len(tw.doc.objects) != 1 {
		t.Errorf("%d embedded objects tracked, want 1", len(tw.doc.objects))
	}
}

// TestUndoDeleteOmitsEmbeddedObjects checks Tk's undo of a deletion across
// embedded objects: the text comes back, the objects don't, and no bare
// placeholder is left behind.
func TestUndoDeleteOmitsEmbeddedObjects(t *testing.T) {
	tw := newBenchWidget(docWithText("ab"), 400, 100)
	tw.ImageCreate("1.1", testImage{"pic", 10, 10})
	tw.undoEnabled, tw.undoStack = true, NewUndoStack(0)

	tw.Delete("1.0", "end")
	tw.Edit("undo")
	if got := tw.doc.Get(Index{1, 0}, tw.doc.EndIndex()); got != "ab" {
		t.Errorf("after undo the document holds %q, want %q", got, "ab")
	}
	if _, ok := tw.doc.imageIndex("pic"); ok {
		t.Error("undo brought the image back")
	}
}
