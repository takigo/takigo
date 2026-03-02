package event

import (
	"time"

	"github.com/msorc/takigo/internal/xlib"
)

// Loop is the main event loop, integrating X11 events with idle callbacks,
// timers, and cross-goroutine dispatching.
type Loop struct {
	display    *xlib.Display
	dispatcher *Dispatcher

	// Channels for the select-based event loop.
	eventCh chan *xlib.RawEvent // raw X events from reader goroutine
	idleCh  chan func()         // idle callbacks (replaces Tcl_DoWhenIdle)
	timerCh chan func()         // timer-fired callbacks
	mainCh  chan func()         // cross-goroutine calls via RunOnMain
	done    chan struct{}        // signal to stop the loop

	// Pending idle callbacks (coalesced).
	idleQueue []func()
}

// NewLoop creates a new event loop for the given display.
func NewLoop(display *xlib.Display, dispatcher *Dispatcher) *Loop {
	return &Loop{
		display:    display,
		dispatcher: dispatcher,
		eventCh:    make(chan *xlib.RawEvent, 64),
		idleCh:     make(chan func(), 256),
		timerCh:    make(chan func(), 64),
		mainCh:     make(chan func(), 64),
		done:       make(chan struct{}),
	}
}

// Run starts the event loop. It blocks until Quit is called.
func (l *Loop) Run() {
	// Start X11 event reader goroutine.
	go l.readEvents()

	for {
		// Drain idle queue first (before blocking).
		l.processIdleQueue()

		select {
		case <-l.done:
			return

		case raw := <-l.eventCh:
			if raw.FilterEvent() {
				continue
			}
			ev := FromRawEventIM(raw, l.display)
			if ev.Type != 0 {
				l.dispatcher.Dispatch(&ev)
			}
			// Flush after dispatching so any X requests issued by
			// event handlers (e.g. MapWindow, IconifyWindow) are
			// sent to the server immediately, not deferred until
			// the next XNextEvent call in the reader goroutine.
			l.display.Flush()

		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)

		case fn := <-l.timerCh:
			fn()
			l.display.Flush()

		case fn := <-l.mainCh:
			fn()
			l.display.Flush()
		}
	}
}

// Quit stops the event loop.
func (l *Loop) Quit() {
	select {
	case <-l.done:
		// Already closed.
	default:
		close(l.done)
	}
}

// DoWhenIdle schedules a function to run during the next idle phase.
// Multiple calls coalesce — the function runs once before the next event.
func (l *Loop) DoWhenIdle(fn func()) {
	select {
	case l.idleCh <- fn:
	case <-l.done:
	}
}

// After schedules a function to run after the given duration.
func (l *Loop) After(d time.Duration, fn func()) {
	time.AfterFunc(d, func() {
		select {
		case l.timerCh <- fn:
		case <-l.done:
		}
	})
}

// RunOnMain schedules a function to run on the main (event loop) goroutine.
// This is safe to call from any goroutine.
func (l *Loop) RunOnMain(fn func()) {
	select {
	case l.mainCh <- fn:
	case <-l.done:
	}
}

// processIdleQueue runs all pending idle callbacks.
func (l *Loop) processIdleQueue() {
	// Drain any pending idle callbacks from the channel.
	for {
		select {
		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)
		default:
			goto drain
		}
	}
drain:
	if len(l.idleQueue) == 0 {
		return
	}
	queue := l.idleQueue
	l.idleQueue = nil
	for _, fn := range queue {
		fn()
	}
	l.display.Flush()
}

// readEvents runs in a separate goroutine, blocking on XNextEvent
// and posting raw events to the eventCh channel.
func (l *Loop) readEvents() {
	for {
		raw := l.display.NextEvent()
		select {
		case l.eventCh <- raw:
		case <-l.done:
			return
		}
	}
}
