// Package gc provides a value-based GC (graphics context) caching pool.
// It ports tk/generic/tkGC.c — GCs with identical values are shared
// and reference-counted.
package gc

import (
	"sync"

	"github.com/msorc/takigo/platform"
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
	gc       platform.GCID
	key      key
	refCount int
}

// Pool manages a cache of shared GCs per display.
type Pool struct {
	mu      sync.Mutex
	server  platform.DisplayServer
	screen  int
	depth   int

	byValue map[key]*entry
	byGC    map[platform.GCID]*entry
}

// NewPool creates a new GC pool for the given display server.
func NewPool(server platform.DisplayServer, screen, depth int) *Pool {
	return &Pool{
		server:  server,
		screen:  screen,
		depth:   depth,
		byValue: make(map[key]*entry),
		byGC:    make(map[platform.GCID]*entry),
	}
}

// Get returns a GC with the specified values, creating one if necessary.
// The returned GC is shared — do not modify it. Call Free when done.
func (p *Pool) Get(drawable platform.DrawableID, mask uint64, values *Values) platform.GCID {
	k := key{Values: *values, depth: p.depth}

	p.mu.Lock()
	defer p.mu.Unlock()

	if e, ok := p.byValue[k]; ok {
		e.refCount++
		return e.gc
	}

	// Create new GC.
	pv := &platform.GCValues{
		Foreground: values.Foreground,
		Background: values.Background,
		LineWidth:  values.LineWidth,
		Function:   values.Function,
	}
	gc := p.server.CreateGC(drawable, mask, pv)

	e := &entry{gc: gc, key: k, refCount: 1}
	p.byValue[k] = e
	p.byGC[gc] = e

	return gc
}

// Free decrements the reference count on a GC, freeing it when unused.
func (p *Pool) Free(gc platform.GCID) {
	p.mu.Lock()
	defer p.mu.Unlock()

	e, ok := p.byGC[gc]
	if !ok {
		return
	}

	e.refCount--
	if e.refCount <= 0 {
		p.server.FreeGC(gc)
		delete(p.byGC, gc)
		delete(p.byValue, e.key)
	}
}

// Cleanup frees all GCs in the pool.
func (p *Pool) Cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for gc := range p.byGC {
		p.server.FreeGC(gc)
	}
	p.byValue = make(map[key]*entry)
	p.byGC = make(map[platform.GCID]*entry)
}
