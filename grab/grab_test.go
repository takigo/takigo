package grab

import (
	"testing"

	"github.com/takigo/takigo/window"
)

func makeTree() (root, child1, child2, grandchild *window.Window) {
	root = &window.Window{Name: "root"}
	child1 = &window.Window{Name: "child1", Parent: root}
	child2 = &window.Window{Name: "child2", Parent: root}
	grandchild = &window.Window{Name: "gc", Parent: child1}
	root.Children = []*window.Window{child1, child2}
	child1.Children = []*window.Window{grandchild}
	return
}

func TestStateNoGrab(t *testing.T) {
	m := &Manager{}
	root, _, _, _ := makeTree()
	if s := m.State(root); s != GrabNone {
		t.Errorf("State with no grab = %v, want GrabNone", s)
	}
}

func TestStateGrabWindow(t *testing.T) {
	root, child1, _, _ := makeTree()
	m := &Manager{grabWin: child1}

	if s := m.State(child1); s != GrabInTree {
		t.Errorf("State(grabWin) = %v, want GrabInTree", s)
	}
	if s := m.State(root); s != GrabAncestor {
		t.Errorf("State(ancestor) = %v, want GrabAncestor", s)
	}
}

func TestStateDescendant(t *testing.T) {
	_, child1, _, grandchild := makeTree()
	m := &Manager{grabWin: child1}

	if s := m.State(grandchild); s != GrabInTree {
		t.Errorf("State(descendant) = %v, want GrabInTree", s)
	}
}

func TestStateExcluded(t *testing.T) {
	_, child1, child2, _ := makeTree()
	m := &Manager{grabWin: child1}

	if s := m.State(child2); s != GrabExcluded {
		t.Errorf("State(sibling) = %v, want GrabExcluded", s)
	}
}

func TestShouldRedirect(t *testing.T) {
	_, child1, child2, grandchild := makeTree()
	m := &Manager{grabWin: child1}

	if m.ShouldRedirect(child1) {
		t.Error("ShouldRedirect(grabWin) should be false")
	}
	if m.ShouldRedirect(grandchild) {
		t.Error("ShouldRedirect(descendant) should be false")
	}
	if !m.ShouldRedirect(child2) {
		t.Error("ShouldRedirect(excluded) should be true")
	}
}

func TestShouldRedirectGlobal(t *testing.T) {
	_, child1, child2, _ := makeTree()
	m := &Manager{grabWin: child1, grabGlobal: true}

	if m.ShouldRedirect(child2) {
		t.Error("ShouldRedirect with global grab should be false (X server handles it)")
	}
}

func TestRedirectTarget(t *testing.T) {
	_, child1, child2, _ := makeTree()
	m := &Manager{grabWin: child1}

	if target := m.RedirectTarget(child2); target != child1 {
		t.Errorf("RedirectTarget(excluded) = %v, want grabWin", target)
	}
	if target := m.RedirectTarget(child1); target != nil {
		t.Errorf("RedirectTarget(grabWin) = %v, want nil", target)
	}
}

func TestCurrentAndIsGlobal(t *testing.T) {
	m := &Manager{}
	if m.Current() != nil {
		t.Error("Current() should be nil with no grab")
	}
	if m.IsGlobal() {
		t.Error("IsGlobal() should be false with no grab")
	}
}
