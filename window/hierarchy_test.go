package window

import "testing"

func TestBuildPathNameFromRoot(t *testing.T) {
	root := &Window{PathName: "."}
	got := BuildPathName(root, "frame1")
	if got != ".frame1" {
		t.Errorf("BuildPathName(root, frame1) = %q, want .frame1", got)
	}
}

func TestBuildPathNameNested(t *testing.T) {
	parent := &Window{PathName: ".frame1"}
	got := BuildPathName(parent, "button1")
	if got != ".frame1.button1" {
		t.Errorf("BuildPathName(.frame1, button1) = %q, want .frame1.button1", got)
	}
}

func TestBuildPathNameNilParent(t *testing.T) {
	got := BuildPathName(nil, "child")
	if got != ".child" {
		t.Errorf("BuildPathName(nil, child) = %q, want .child", got)
	}
}

func TestAddChild(t *testing.T) {
	parent := &Window{PathName: "."}
	child := &Window{Name: "child1"}

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if parent.Children[0] != child {
		t.Error("child not found in parent's children")
	}
	if child.Parent != parent {
		t.Error("child.Parent not set")
	}
}

func TestAddMultipleChildren(t *testing.T) {
	parent := &Window{PathName: "."}
	c1 := &Window{Name: "c1"}
	c2 := &Window{Name: "c2"}

	parent.AddChild(c1)
	parent.AddChild(c2)

	if len(parent.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(parent.Children))
	}
}

func TestRemoveChild(t *testing.T) {
	parent := &Window{PathName: "."}
	c1 := &Window{Name: "c1"}
	c2 := &Window{Name: "c2"}

	parent.AddChild(c1)
	parent.AddChild(c2)
	parent.RemoveChild(c1)

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child after remove, got %d", len(parent.Children))
	}
	if parent.Children[0] != c2 {
		t.Error("wrong child remaining")
	}
	if c1.Parent != nil {
		t.Error("removed child's Parent should be nil")
	}
}

func TestRemoveNonexistentChild(t *testing.T) {
	parent := &Window{PathName: "."}
	c1 := &Window{Name: "c1"}
	c2 := &Window{Name: "c2"}

	parent.AddChild(c1)
	parent.RemoveChild(c2) // should not panic

	if len(parent.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(parent.Children))
	}
	if c2.Parent != nil {
		t.Error("non-child's Parent should be nil")
	}
}

func TestToplevelSelf(t *testing.T) {
	w := &Window{Flags: FlagTopLevel}
	if got := Toplevel(w); got != w {
		t.Error("Toplevel of a toplevel should return itself")
	}
}

func TestToplevelWalkUp(t *testing.T) {
	top := &Window{Flags: FlagTopLevel}
	mid := &Window{Parent: top}
	child := &Window{Parent: mid}

	got := Toplevel(child)
	if got != top {
		t.Error("Toplevel should walk up to nearest toplevel ancestor")
	}
}

func TestToplevelNil(t *testing.T) {
	w := &Window{} // no FlagTopLevel, no Parent
	if got := Toplevel(w); got != nil {
		t.Error("Toplevel should return nil when no toplevel in chain")
	}
}

func TestIsTopLevel(t *testing.T) {
	w := &Window{Flags: FlagTopLevel}
	if !w.IsTopLevel() {
		t.Error("IsTopLevel should return true")
	}
	w2 := &Window{}
	if w2.IsTopLevel() {
		t.Error("IsTopLevel should return false for non-toplevel")
	}
}

func TestIsMapped(t *testing.T) {
	w := &Window{Flags: FlagMapped}
	if !w.IsMapped() {
		t.Error("IsMapped should return true")
	}
	w2 := &Window{}
	if w2.IsMapped() {
		t.Error("IsMapped should return false for unmapped")
	}
}
