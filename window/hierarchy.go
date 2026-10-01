package window

import "strings"

// AddChild adds a child window to this window's children list.
func (w *Window) AddChild(child *Window) {
	child.Parent = w
	w.Children = append(w.Children, child)
}

// RemoveChild removes a child window from this window's children list.
func (w *Window) RemoveChild(child *Window) {
	for i, c := range w.Children {
		if c == child {
			w.Children = append(w.Children[:i], w.Children[i+1:]...)
			break
		}
	}
	child.Parent = nil
}

// Lookup returns the descendant of w with the given Tk path name
// (".frame.ok"), w itself for its own path, or nil. Call it on the root
// window to find any window of the application.
func (w *Window) Lookup(path string) *Window {
	if path == w.PathName {
		return w
	}
	rest, ok := strings.CutPrefix(path, strings.TrimSuffix(w.PathName, ".")+".")
	if !ok {
		return nil
	}
	cur := w
	for name := range strings.SplitSeq(rest, ".") {
		var next *Window
		for _, c := range cur.Children {
			if c.Name == name {
				next = c
				break
			}
		}
		if next == nil {
			return nil
		}
		cur = next
	}
	return cur
}

// BuildPathName constructs the full path name for a child window.
func BuildPathName(parent *Window, name string) string {
	if parent == nil || parent.PathName == "." {
		return "." + name
	}
	return parent.PathName + "." + name
}
