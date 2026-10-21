package window

import (
	"slices"
	"sync"
	"sync/atomic"
)

// hookEntry is one registered hook. A dead entry has been unregistered
// and is skipped when the hooks run.
type hookEntry struct {
	fn   func(*Window)
	dead atomic.Bool
}

// hookList is a list of hooks, guarded by its own lock.
type hookList struct {
	mu      sync.Mutex
	entries []*hookEntry
}

// add registers fn and returns the function that unregisters it.
func (l *hookList) add(fn func(*Window)) func() {
	entry := &hookEntry{fn: fn}
	l.mu.Lock()
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
	return func() {
		l.mu.Lock()
		defer l.mu.Unlock()
		if entry.dead.Swap(true) {
			return
		}
		// A new slice, not an in-place delete: the list a reader
		// holds must not change under it.
		kept := make([]*hookEntry, 0, len(l.entries))
		for _, e := range l.entries {
			if e != entry {
				kept = append(kept, e)
			}
		}
		l.entries = kept
	}
}

// snapshot returns the hooks of the list.
func (l *hookList) snapshot() []*hookEntry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.entries
}

// set replaces the list with a copy of entries.
func (l *hookList) set(entries []*hookEntry) {
	l.mu.Lock()
	l.entries = slices.Clone(entries)
	l.mu.Unlock()
}

// run calls the hooks that are still registered.
func (l *hookList) run(w *Window) {
	l.mu.Lock()
	entries := l.entries
	l.mu.Unlock()
	for _, e := range entries {
		if !e.dead.Load() {
			e.fn(w)
		}
	}
}

// The hooks the geometry managers register for the process. Their
// managers are process-wide singletons that register in init(), before
// any display exists, so their hooks are shared; a display copies them
// into its own lists (see refreshHooks), so the hooks a display runs
// live with it.
var (
	sharedMapped   hookList
	sharedUnmapped hookList
	sharedMoved    hookList

	// sharedVersion counts the changes to the shared lists, so that a
	// display knows when to copy them again.
	sharedVersion atomic.Int64
)

// AddMappedHook registers fn to run whenever MarkMapped maps a window,
// on every display. It returns a function that unregisters the hook.
func AddMappedHook(fn func(*Window)) func() {
	return addSharedHook(&sharedMapped, fn)
}

// AddUnmappedHook registers fn to run whenever MarkUnmapped unmaps a
// window, on every display. It returns a function that unregisters the
// hook.
func AddUnmappedHook(fn func(*Window)) func() {
	return addSharedHook(&sharedUnmapped, fn)
}

// AddMovedHook registers fn to run whenever NotifyMoved reports a move,
// on every display. It returns a function that unregisters the hook.
func AddMovedHook(fn func(*Window)) func() {
	return addSharedHook(&sharedMoved, fn)
}

// addSharedHook registers fn in a shared list and returns the function
// that unregisters it.
func addSharedHook(list *hookList, fn func(*Window)) func() {
	remove := list.add(fn)
	sharedVersion.Add(1)
	return func() {
		remove()
		sharedVersion.Add(1)
	}
}

// displayHooks holds the hooks of one display: its own and the copies
// of the shared ones.
type displayHooks struct {
	version      atomic.Int64
	ownMapped    hookList
	ownUnmapped  hookList
	ownMoved     hookList
	sharedMapped hookList
	sharedUnmap  hookList
	sharedMoved  hookList
}

// OnMapped registers fn to run whenever MarkMapped maps a window on this
// display. It returns a function that unregisters the hook.
func (d *Display) OnMapped(fn func(*Window)) func() {
	return d.hooks.ownMapped.add(fn)
}

// OnUnmapped registers fn to run whenever MarkUnmapped unmaps a window
// on this display. It returns a function that unregisters the hook.
func (d *Display) OnUnmapped(fn func(*Window)) func() {
	return d.hooks.ownUnmapped.add(fn)
}

// OnMoved registers fn to run whenever NotifyMoved reports a move on
// this display. It returns a function that unregisters the hook.
func (d *Display) OnMoved(fn func(*Window)) func() {
	return d.hooks.ownMoved.add(fn)
}

// refreshHooks copies the shared hooks into the display's lists when
// they changed since the last copy.
func (d *Display) refreshHooks() {
	version := sharedVersion.Load()
	if d.hooks.version.Load() == version {
		return
	}
	d.hooks.sharedMapped.set(sharedMapped.snapshot())
	d.hooks.sharedUnmap.set(sharedUnmapped.snapshot())
	d.hooks.sharedMoved.set(sharedMoved.snapshot())
	// Published last, so that a reader which sees the version also sees
	// the lists.
	d.hooks.version.Store(version)
}

// runMapped runs the mapped hooks of the display.
func (d *Display) runMapped(w *Window) {
	if d == nil {
		return
	}
	d.refreshHooks()
	d.hooks.sharedMapped.run(w)
	d.hooks.ownMapped.run(w)
}

// runUnmapped runs the unmapped hooks of the display.
func (d *Display) runUnmapped(w *Window) {
	if d == nil {
		return
	}
	d.refreshHooks()
	d.hooks.sharedUnmap.run(w)
	d.hooks.ownUnmapped.run(w)
}

// runMoved runs the moved hooks of the display.
func (d *Display) runMoved(w *Window) {
	if d == nil {
		return
	}
	d.refreshHooks()
	d.hooks.sharedMoved.run(w)
	d.hooks.ownMoved.run(w)
}
