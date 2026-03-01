// Package gc provides a value-based GC (graphics context) caching pool.
// It ports tk/generic/tkGC.c — GCs with identical values are shared
// and reference-counted.
package gc

import (
	"sync"

	"github.com/msorc/takigo/internal/xlib"
)

// Values describes the desired attributes for a graphics context.
type Values struct {
	Foreground uint64
	Background uint64
	LineWidth  int
	Function   int
}

// key is the cache lookup key combining GC values with display identity.
type key struct {
	Values
	depth int
}

// entry is a cached GC with a reference count.
type entry struct {
	gc       xlib.GC
	refCount int
}

// Pool manages a cache of shared GCs per display.
type Pool struct {
	mu      sync.Mutex
	display *xlib.Display
	screen  int
	depth   int

	byValue map[key]*entry
	byGC    map[xlib.GC]*entry
}

// NewPool creates a new GC pool for the given display.
func NewPool(display *xlib.Display, screen, depth int) *Pool {
	return &Pool{
		display: display,
		screen:  screen,
		depth:   depth,
		byValue: make(map[key]*entry),
		byGC:    make(map[xlib.GC]*entry),
	}
}

// Get returns a GC with the specified values, creating one if necessary.
// The returned GC is shared — do not modify it. Call Free when done.
func (p *Pool) Get(drawable xlib.Drawable, mask uint64, values *Values) xlib.GC {
	k := key{Values: *values, depth: p.depth}

	p.mu.Lock()
	defer p.mu.Unlock()

	if e, ok := p.byValue[k]; ok {
		e.refCount++
		return e.gc
	}

	// Create new GC.
	xv := &xlib.GCValues{
		Foreground: values.Foreground,
		Background: values.Background,
		LineWidth:  values.LineWidth,
		Function:   values.Function,
	}
	gc := p.display.CreateGC(drawable, mask, xv)

	e := &entry{gc: gc, refCount: 1}
	p.byValue[k] = e
	p.byGC[gc] = e

	return gc
}

// Free decrements the reference count on a GC, freeing it when unused.
func (p *Pool) Free(gc xlib.GC) {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.byGC[gc]
	if !ok {
		return
	}

	e.refCount--
	if e.refCount <= 0 {
		p.display.FreeGC(gc)
		delete(p.byGC, gc)
		// Remove from byValue map.
		for k, v := range p.byValue {
			if v == e {
				delete(p.byValue, k)
				break
			}
		}
	}
}

// Cleanup frees all GCs in the pool.
func (p *Pool) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for gc := range p.byGC {
		p.display.FreeGC(gc)
	}
	p.byValue = make(map[key]*entry)
	p.byGC = make(map[xlib.GC]*entry)
}
