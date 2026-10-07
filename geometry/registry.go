package geometry

import (
	"github.com/takigo/takigo/platform"
	"github.com/takigo/takigo/window"
)

// Container is a geometry manager's per-container state (pack's packer,
// grid's gridder, place's placer).
type Container interface {
	// Window returns the container window.
	Window() *window.Window
	// Contents returns the windows managed in the container.
	Contents() []*window.Window
	// Arrange lays the contents out now; ScheduleArrange at idle time.
	Arrange()
	ScheduleArrange()
	// Remove forgets one content window.
	Remove(content *window.Window)
}

// Registry holds a geometry manager's containers and the window each
// content is managed in (Tk's containerPtr), with what the three managers
// share of Tk's structure procs and Tk_MaintainGeometry: creating a
// container's state with its destroy and resize hooks, dropping it with
// its window, and keeping content managed -in a container mapped, placed
// and unmapped with it. It is the window.GeomManager the manager registers.
type Registry[C Container] struct {
	name string
	new  func(*window.Window) C

	// Containers maps a container window to its state; ContainerOf maps
	// a content window to the -in container it was given (nil: its
	// parent). Both live on the windows (Table).
	Containers  Table[C]
	ContainerOf Table[*window.Window]
	hooked      Table[bool]
}

// NewRegistry makes a registry for the manager called name whose
// per-container state newContainer creates, and hooks it into the
// windows' map, unmap and move notifications.
func NewRegistry[C Container](name string, newContainer func(*window.Window) C) *Registry[C] {
	r := &Registry[C]{name: name, new: newContainer}
	// Arrange (and so map) the content once its container is mapped, as
	// the managers' structure procs do on MapNotify; keep -in content with
	// its container when it moves or is unmapped (Tk_MaintainGeometry).
	window.AddMappedHook(r.ArrangeContainer)
	window.AddMovedHook(func(w *window.Window) {
		if c, ok := r.Containers.Get(w); ok && HasForeign(c) {
			c.ScheduleArrange()
		}
	})
	window.AddUnmappedHook(func(w *window.Window) {
		if c, ok := r.Containers.Get(w); ok {
			unmapForeign(c)
		}
	})
	return r
}

// Name returns the manager's name.
func (r *Registry[C]) Name() string { return r.name }

// RequestProc re-arranges the container of content, whose size request
// changed.
func (r *Registry[C]) RequestProc(content *window.Window) {
	if c, ok := r.Containers.Get(r.ContainerFor(content)); ok {
		c.ScheduleArrange()
	}
}

// LostContentProc drops content from its container, e.g. when content is
// destroyed or taken over by another geometry manager.
func (r *Registry[C]) LostContentProc(content *window.Window) {
	container := r.ContainerFor(content)
	r.ContainerOf.Delete(content)
	if c, ok := r.Containers.Get(container); ok {
		c.Remove(content)
		c.ScheduleArrange()
	}
}

// ContainerFor returns the window content is managed in: its -in
// container if one was given, else its parent.
func (r *Registry[C]) ContainerFor(content *window.Window) *window.Window {
	if c := r.ContainerOf.Of(content); c != nil {
		return c
	}
	return content.Parent
}

// For returns container's state, creating it and, once per window, the
// hooks that drop it on destroy and re-arrange on resize (the container
// may be resized by anything, e.g. a paned window).
func (r *Registry[C]) For(container *window.Window) C {
	c, ok := r.Containers.Get(container)
	if !ok {
		c = r.new(container)
		r.Containers.Set(container, c)
	}
	if !r.hooked.Of(container) {
		// The state may be dropped when its last content goes but the
		// hooks stay with the window, so register them once.
		r.hooked.Set(container, true)
		container.OnDestroy(func() {
			r.Forget(container)
			r.hooked.Delete(container)
		})
		container.OnConfigure(func() {
			if c, ok := r.Containers.Get(container); ok {
				c.ScheduleArrange()
			}
		})
	}
	return c
}

// Forget drops a destroyed container's state; content managed in it from
// outside its subtree (via -in) becomes unmanaged, as in Tk's
// DestroyNotify handling in the geometry managers.
func (r *Registry[C]) Forget(container *window.Window) {
	c, ok := r.Containers.Get(container)
	if !ok {
		return
	}
	r.Containers.Delete(container)
	for _, w := range c.Contents() {
		if r.ContainerOf.Of(w) != container {
			continue
		}
		r.ContainerOf.Delete(w)
		w.GeomManager = nil
		if w.IsMapped() && !w.IsDestroyed() && w.PlatformID != 0 {
			w.Display.Server.UnmapWindow(w.PlatformID)
			window.MarkUnmapped(w)
		}
	}
}

// ArrangeContainer lays out container's contents now, if it has any.
func (r *Registry[C]) ArrangeContainer(container *window.Window) {
	if c, ok := r.Containers.Get(container); ok {
		c.Arrange()
	}
}

// ArrangeAll schedules a layout of every managed container in root's
// subtree, root included.
func (r *Registry[C]) ArrangeAll(root *window.Window) {
	if root == nil {
		return
	}
	if c, ok := r.Containers.Get(root); ok {
		c.ScheduleArrange()
	}
	for _, child := range root.Children {
		r.ArrangeAll(child)
	}
}

// HasForeign reports whether some content is managed -in c without being
// its child.
func HasForeign(c Container) bool {
	container := c.Window()
	for _, w := range c.Contents() {
		if w.Parent != container {
			return true
		}
	}
	return false
}

// unmapForeign unmaps -in content whose container was unmapped; X does
// this for real children.
func unmapForeign(c Container) {
	container := c.Window()
	for _, w := range c.Contents() {
		if w.Parent != container && w.IsMapped() && w.PlatformID != platform.WindowID(0) {
			w.Display.Server.UnmapWindow(w.PlatformID)
			window.MarkUnmapped(w)
		}
	}
}

// PlaceContent gives content its place in container, (x, y) and
// width x height in the container's coordinates, maps it once the
// container is viewable and tells the managers when it moved: the tail
// the managers' arrange procedures share.
func PlaceContent(container, content *window.Window, x, y, width, height int) {
	// Content managed -in another container is offset by its position.
	dx, dy := window.ContentOffset(container, content)
	moved := content.X != x+dx || content.Y != y+dy
	content.X, content.Y = x+dx, y+dy
	content.Width, content.Height = max(width, 1), max(height, 1)
	if content.PlatformID != platform.WindowID(0) {
		container.Display.Server.MoveResizeWindow(content.PlatformID,
			content.X, content.Y, uint(content.Width), uint(content.Height))
		// Tk maps content only once its container is mapped; the
		// container's MarkMapped re-arranges and maps it then.
		if !content.IsMapped() && window.ContainerViewable(container, content) {
			window.SyncBackground(content)
			container.Display.Server.MapWindow(content.PlatformID)
			window.MarkMapped(content)
		}
	}
	if moved {
		window.NotifyMoved(content)
	}
}
