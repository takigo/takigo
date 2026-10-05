package image

import (
	"testing"

	"github.com/takigo/takigo/platform"
)

// stubImage implements Image for testing without X11.
type stubImage struct {
	name      string
	w, h      int
	destroyed bool
}

func (s *stubImage) Name() string { return s.name }
func (s *stubImage) Width() int   { return s.w }
func (s *stubImage) Height() int  { return s.h }
func (s *stubImage) Destroy()     { s.destroyed = true }
func (s *stubImage) Draw(_ platform.DisplayServer, _ platform.DrawableID, _ platform.GCID, _ int, _, _, _, _, _, _ int, _ uint64) {
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	img := &stubImage{name: "test", w: 10, h: 20}
	r.Register(img)

	got := r.Get("test")
	if got == nil || got.Name() != "test" {
		t.Fatalf("Get returned %v, want image 'test'", got)
	}
}

func TestRegistryGetMissing(t *testing.T) {
	r := NewRegistry()
	if got := r.Get("nonexistent"); got != nil {
		t.Fatalf("Get returned %v for missing image", got)
	}
}

func TestRegistryRegisterOverwrite(t *testing.T) {
	r := NewRegistry()
	old := &stubImage{name: "img", w: 10, h: 10}
	new_ := &stubImage{name: "img", w: 20, h: 20}

	r.Register(old)
	r.Register(new_)

	if !old.destroyed {
		t.Error("old image was not destroyed on overwrite")
	}
	if got := r.Get("img"); got == nil || got.Width() != 20 {
		t.Error("Get did not return the new image")
	}
}

func TestRegistryUnregister(t *testing.T) {
	r := NewRegistry()
	img := &stubImage{name: "test"}
	r.Register(img)
	r.Unregister("test")

	if !img.destroyed {
		t.Error("image was not destroyed on unregister")
	}
	if got := r.Get("test"); got != nil {
		t.Errorf("Get returned %v after unregister", got)
	}
}

func TestRegistryUnregisterMissing(t *testing.T) {
	r := NewRegistry()
	r.Unregister("nonexistent") // should not panic
}

func TestRegistryDestroyAll(t *testing.T) {
	r := NewRegistry()
	imgs := []*stubImage{
		{name: "a"}, {name: "b"}, {name: "c"},
	}
	for _, img := range imgs {
		r.Register(img)
	}

	r.DestroyAll()

	for _, img := range imgs {
		if !img.destroyed {
			t.Errorf("image %q was not destroyed", img.name)
		}
		if got := r.Get(img.name); got != nil {
			t.Errorf("Get(%q) returned %v after DestroyAll", img.name, got)
		}
	}
}
