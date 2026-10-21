package widget

import (
	"sync"
	"time"
)

// The times the insert cursor of an entry-like widget stays on and off,
// Tk's insertontime and insertofftime defaults.
const (
	InsertOnTime  = 600 * time.Millisecond
	InsertOffTime = 300 * time.Millisecond
)

// CursorBlinker blinks the insert cursor of an entry-like widget: it
// calls toggle after InsertOnTime and again after InsertOffTime, as Tk's
// entry does. The cursor is on when the widget takes the focus and stops
// blinking when it loses it.
type CursorBlinker struct {
	sched  Scheduler
	toggle func()

	mu      sync.Mutex
	cancel  func() bool
	running bool
	gen     int  // bumped by Start, so a timer of an older run cannot re-arm
	on      bool // the cursor state after the last toggle
}

// NewCursorBlinker returns a blinker that calls toggle on every change of
// the cursor state. The
// scheduler is the App's, so the toggle runs on the event loop.
func NewCursorBlinker(sched Scheduler, toggle func()) *CursorBlinker {
	return &CursorBlinker{sched: sched, toggle: toggle}
}

// Start begins the blinking; it does nothing when the blinker already
// runs.
func (b *CursorBlinker) Start() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.running || b.sched == nil {
		return
	}
	b.running = true
	b.gen++
	b.on = true
	b.arm(b.gen)
}

// Stop ends the blinking.
func (b *CursorBlinker) Stop() {
	if b == nil {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.running {
		return
	}
	b.running = false
	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
}

// arm schedules the next toggle of run gen. The caller holds the lock.
func (b *CursorBlinker) arm(gen int) {
	d := InsertOffTime
	if b.on {
		d = InsertOnTime
	}
	b.cancel = b.sched.After(d, func() {
		b.mu.Lock()
		if !b.running || b.gen != gen {
			b.mu.Unlock()
			return
		}
		b.on = !b.on
		b.arm(gen)
		b.mu.Unlock()
		b.toggle()
	})
}
