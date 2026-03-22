package focus

import (
	"testing"

	"github.com/msorc/takigo/window"
)

func TestFlattenTree(t *testing.T) {
	root := &window.Window{Name: "root"}
	a := &window.Window{Name: "a", Parent: root}
	b := &window.Window{Name: "b", Parent: root}
	a1 := &window.Window{Name: "a1", Parent: a}
	a2 := &window.Window{Name: "a2", Parent: a}
	root.Children = []*window.Window{a, b}
	a.Children = []*window.Window{a1, a2}

	got := flattenTree(root)
	want := []string{"root", "a", "a1", "a2", "b"}
	if len(got) != len(want) {
		t.Fatalf("flattenTree returned %d nodes, want %d", len(got), len(want))
	}
	for i, w := range got {
		if w.Name != want[i] {
			t.Errorf("flattenTree[%d] = %q, want %q", i, w.Name, want[i])
		}
	}
}

func TestFlattenTreeSkipsToplevel(t *testing.T) {
	root := &window.Window{Name: "root"}
	child := &window.Window{Name: "child", Parent: root}
	embedded := &window.Window{Name: "embedded", Parent: root, Flags: window.FlagTopLevel}
	embeddedChild := &window.Window{Name: "ec", Parent: embedded}
	embedded.Children = []*window.Window{embeddedChild}
	root.Children = []*window.Window{child, embedded}

	got := flattenTree(root)
	for _, w := range got {
		if w.Name == "embedded" || w.Name == "ec" {
			t.Errorf("flattenTree included toplevel subtree node %q", w.Name)
		}
	}
	if len(got) != 2 { // root, child
		t.Errorf("flattenTree returned %d nodes, want 2", len(got))
	}
}

func TestFlattenTreeLeaf(t *testing.T) {
	w := &window.Window{Name: "leaf"}
	got := flattenTree(w)
	if len(got) != 1 || got[0] != w {
		t.Errorf("flattenTree(leaf) = %v, want [leaf]", got)
	}
}

func TestFindToplevel(t *testing.T) {
	tl := &window.Window{Name: "top", Flags: window.FlagTopLevel}
	child := &window.Window{Name: "child", Parent: tl}
	gc := &window.Window{Name: "gc", Parent: child}

	if got := findToplevel(gc); got != tl {
		t.Errorf("findToplevel(gc) = %v, want toplevel", got)
	}
}

func TestFindToplevelNone(t *testing.T) {
	w := &window.Window{Name: "orphan"}
	if got := findToplevel(w); got != nil {
		t.Errorf("findToplevel(orphan) = %v, want nil", got)
	}
}

func TestNextFocusableForward(t *testing.T) {
	root := &window.Window{Name: "root"}
	a := &window.Window{Name: "a", Parent: root, Flags: window.FlagFocusable, PlatformID: 1}
	b := &window.Window{Name: "b", Parent: root}
	c := &window.Window{Name: "c", Parent: root, Flags: window.FlagFocusable, PlatformID: 2}
	root.Children = []*window.Window{a, b, c}

	m := &Manager{
		IsFocusable: func(w *window.Window) bool {
			return w.Flags&window.FlagFocusable != 0 && w.PlatformID != 0
		},
	}

	got := m.nextFocusable(root, a, true)
	if got != c {
		t.Errorf("nextFocusable(a, forward) = %v, want c", got)
	}
}

func TestNextFocusableBackward(t *testing.T) {
	root := &window.Window{Name: "root"}
	a := &window.Window{Name: "a", Parent: root, Flags: window.FlagFocusable, PlatformID: 1}
	b := &window.Window{Name: "b", Parent: root}
	c := &window.Window{Name: "c", Parent: root, Flags: window.FlagFocusable, PlatformID: 2}
	root.Children = []*window.Window{a, b, c}

	m := &Manager{
		IsFocusable: func(w *window.Window) bool {
			return w.Flags&window.FlagFocusable != 0 && w.PlatformID != 0
		},
	}

	got := m.nextFocusable(root, c, false)
	if got != a {
		t.Errorf("nextFocusable(c, backward) = %v, want a", got)
	}
}

func TestNextFocusableWraps(t *testing.T) {
	root := &window.Window{Name: "root"}
	a := &window.Window{Name: "a", Parent: root, Flags: window.FlagFocusable, PlatformID: 1}
	b := &window.Window{Name: "b", Parent: root, Flags: window.FlagFocusable, PlatformID: 2}
	root.Children = []*window.Window{a, b}

	m := &Manager{
		IsFocusable: func(w *window.Window) bool {
			return w.Flags&window.FlagFocusable != 0 && w.PlatformID != 0
		},
	}

	// From b forward should wrap to a.
	got := m.nextFocusable(root, b, true)
	if got != a {
		t.Errorf("nextFocusable(b, forward) = %v, want a (wrap)", got)
	}
}

func TestNextFocusableNone(t *testing.T) {
	root := &window.Window{Name: "root"}
	a := &window.Window{Name: "a", Parent: root}
	root.Children = []*window.Window{a}

	m := &Manager{
		IsFocusable: func(w *window.Window) bool { return false },
	}

	got := m.nextFocusable(root, root, true)
	if got != nil {
		t.Errorf("nextFocusable with no focusable = %v, want nil", got)
	}
}
