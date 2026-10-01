package window

import (
	"slices"
	"testing"
)

func TestChildNamesAreUnique(t *testing.T) {
	d := &Display{}
	root := &Window{Display: d, PathName: "."}

	a := NewChildWindow(root, "f", 0, 0, 1, 1)
	b := NewChildWindow(root, "f", 0, 0, 1, 1)
	c := NewChildWindow(root, "f", 0, 0, 1, 1)
	if a.PathName != ".f" || b.PathName != ".f#2" || c.PathName != ".f#3" {
		t.Errorf("paths = %q, %q, %q; want .f, .f#2, .f#3", a.PathName, b.PathName, c.PathName)
	}

	// A name is free again once its window is gone.
	root.RemoveChild(b)
	if w := NewChildWindow(root, "f#2", 0, 0, 1, 1); w.PathName != ".f#2" {
		t.Errorf("reused name gave %q", w.PathName)
	}
}

func TestEmptyNameIsGenerated(t *testing.T) {
	root := &Window{Display: &Display{}, PathName: "."}
	NewChildWindow(root, "w2", 0, 0, 1, 1)
	first := NewChildWindow(root, "", 0, 0, 1, 1)
	second := NewChildWindow(root, "", 0, 0, 1, 1)
	if first.Name != "w1" || second.Name != "w3" {
		t.Errorf("generated names = %q, %q; want w1, w3 (w2 is taken)", first.Name, second.Name)
	}
}

func TestLookup(t *testing.T) {
	root := &Window{Display: &Display{}, PathName: "."}
	f := NewChildWindow(root, "f", 0, 0, 1, 1)
	ok := NewChildWindow(f, "ok", 0, 0, 1, 1)
	NewChildWindow(root, "fx", 0, 0, 1, 1)

	tests := []struct {
		from *Window
		path string
		want *Window
	}{
		{root, ".", root},
		{root, ".f", f},
		{root, ".f.ok", ok},
		{f, ".f.ok", ok},
		{f, ".f", f},
		{root, ".f.missing", nil},
		{root, ".missing.ok", nil},
		{f, ".fx", nil},
		{root, "", nil},
		{root, "f", nil},
	}
	for _, tt := range tests {
		if got := tt.from.Lookup(tt.path); got != tt.want {
			t.Errorf("%s.Lookup(%q) = %v, want %v", tt.from.PathName, tt.path, got, tt.want)
		}
	}
}

func TestDescendants(t *testing.T) {
	root := &Window{Display: &Display{}, PathName: "."}
	a := NewChildWindow(root, "a", 0, 0, 1, 1)
	NewChildWindow(a, "x", 0, 0, 1, 1)
	NewChildWindow(a, "y", 0, 0, 1, 1)
	NewChildWindow(root, "b", 0, 0, 1, 1)

	var got []string
	for w := range root.Descendants() {
		got = append(got, w.PathName)
	}
	want := []string{".a", ".a.x", ".a.y", ".b"}
	if !slices.Equal(got, want) {
		t.Errorf("Descendants = %v, want %v", got, want)
	}

	got = got[:0]
	for w := range root.Descendants() {
		got = append(got, w.PathName)
		if w.PathName == ".a.x" {
			break
		}
	}
	if !slices.Equal(got, []string{".a", ".a.x"}) {
		t.Errorf("after break: %v", got)
	}
}
