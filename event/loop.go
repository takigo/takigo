package event

import (
	"time"

	"github.com/msorc/takigo/platform"
)

// Loop is the main event loop, integrating platform events with idle callbacks,
// timers, and cross-goroutine dispatching.
type Loop struct {
	server     platform.DisplayServer
	parser     platform.EventParser
	hasIM      bool
	dispatcher *Dispatcher

	// Channels for the select-based event loop.
	eventCh chan *platform.RawEvent // raw events from reader goroutine
	idleCh  chan func()             // idle callbacks (replaces Tcl_DoWhenIdle)
	timerCh chan func()             // timer-fired callbacks
	mainCh  chan func()             // cross-goroutine calls via RunOnMain
	done    chan struct{}            // signal to stop the loop

	// Pending idle callbacks (coalesced).
	idleQueue []func()

	// rawHandler is called for every raw event before type conversion.
	// Used to handle event types (e.g. selection) not routed through Dispatcher.
	rawHandler func(*platform.RawEvent)
}

// NewLoop creates a new event loop for the given display server.
func NewLoop(server platform.DisplayServer, parser platform.EventParser, dispatcher *Dispatcher) *Loop {
	return &Loop{
		server:     server,
		parser:     parser,
		hasIM:      server.HasIM(),
		dispatcher: dispatcher,
		eventCh:    make(chan *platform.RawEvent, 64),
		idleCh:     make(chan func(), 256),
		timerCh:    make(chan func(), 64),
		mainCh:     make(chan func(), 64),
		done:       make(chan struct{}),
	}
}

// Run starts the event loop. It blocks until Quit is called.
func (l *Loop) Run() {
	// Start event reader goroutine.
	go l.readEvents()

	for {
		// Drain idle queue first (before blocking).
		l.processIdleQueue()

		select {
		case <-l.done:
			return

		case raw := <-l.eventCh:
			if l.server.FilterEvent(raw) {
				continue
			}
			if l.rawHandler != nil {
				l.rawHandler(raw)
			}
			ev := FromRawEventIM(raw, l.parser, l.hasIM)
			if ev.Type != 0 {
				l.dispatcher.Dispatch(&ev)
			}
			l.server.Flush()

		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)

		case fn := <-l.timerCh:
			fn()
			l.server.Flush()

		case fn := <-l.mainCh:
			fn()
			l.server.Flush()
		}
	}
}

// SetRawEventHandler installs a handler called for every raw event before
// type conversion. Use it for event types not routed through the Dispatcher
// (e.g. X11 selection events).
func (l *Loop) SetRawEventHandler(h func(*platform.RawEvent)) {
	l.rawHandler = h
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

// RunNested processes events until the done channel is closed.
// This implements nested event loops needed for modal dialogs (like Tcl's vwait).
// It must be called from within a handler running on the main goroutine.
func (l *Loop) RunNested(done <-chan struct{}) {
	for {
		l.processIdleQueue()

		select {
		case <-done:
			return

		case <-l.done:
			return

		case raw := <-l.eventCh:
			if l.server.FilterEvent(raw) {
				continue
			}
			if l.rawHandler != nil {
				l.rawHandler(raw)
			}
			ev := FromRawEventIM(raw, l.parser, l.hasIM)
			if ev.Type != 0 {
				l.dispatcher.Dispatch(&ev)
			}
			l.server.Flush()

		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)

		case fn := <-l.timerCh:
			fn()
			l.server.Flush()

		case fn := <-l.mainCh:
			fn()
			l.server.Flush()
		}
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
	l.server.Flush()
}

// readEvents runs in a separate goroutine, blocking on NextEvent
// and posting raw events to the eventCh channel.
func (l *Loop) readEvents() {
	for {
		raw := l.server.NextEvent()
		select {
		case l.eventCh <- raw:
		case <-l.done:
			return
		}
	}
}
