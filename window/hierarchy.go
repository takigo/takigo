package window

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

// BuildPathName constructs the full path name for a child window.
func BuildPathName(parent *Window, name string) string {
	if parent == nil || parent.PathName == "." {
		return "." + name
	}
	return parent.PathName + "." + name
}
