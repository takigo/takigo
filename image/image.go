// Package image provides the image system for takigo, supporting photo images
// that can be displayed in widgets. It ports the core concepts from
// tk/generic/tkImage.c simplified for modern 24/32-bit displays.
package image

import (
	"sync"

	"github.com/takigo/takigo/platform"
)

// Image is the interface implemented by all takigo image types.
type Image interface {
	// Name returns the image's registered name.
	Name() string

	// Width returns the image width in pixels.
	Width() int

	// Height returns the image height in pixels.
	Height() int

	// Draw renders a region of the image onto a drawable.
	// imgX, imgY, w, h define the source region within the image.
	// dstX, dstY define the destination position on the drawable.
	// bgPixel is the background color for alpha compositing.
	Draw(d platform.DisplayServer, drawable platform.DrawableID, gc platform.GCID,
		depth int,
		imgX, imgY, w, h, dstX, dstY int,
		bgPixel uint64)

	// Destroy releases resources associated with the image.
	Destroy()
}

// Registry manages a collection of named images.
type Registry struct {
	mu     sync.RWMutex
	images map[string]Image
}

// NewRegistry creates an empty image registry.
func NewRegistry() *Registry {
	return &Registry{
		images: make(map[string]Image),
	}
}

// Register adds an image to the registry under the given name.
// If an image with the same name already exists, it is destroyed first.
func (r *Registry) Register(img Image) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if old, ok := r.images[img.Name()]; ok {
		old.Destroy()
	}
	r.images[img.Name()] = img
}

// Get returns the image with the given name, or nil if not found.
func (r *Registry) Get(name string) Image {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.images[name]
}

// Unregister removes and destroys the image with the given name.
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if img, ok := r.images[name]; ok {
		img.Destroy()
		delete(r.images, name)
	}
}

// DestroyAll destroys and removes all registered images.
func (r *Registry) DestroyAll() {
	r.mu.Lock()
	defer r.mu.Unlock()
	for name, img := range r.images {
		img.Destroy()
		delete(r.images, name)
	}
}
