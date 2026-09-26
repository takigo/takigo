package event

import (
	"context"
	"os"
	"time"

	"github.com/msorc/takigo/platform"
)

// EventPumper is an optional interface that platform backends can implement
// when they need to be periodically polled from the main goroutine.
// macOS/Cocoa requires this because NSEvents must be pumped on the main thread.
type EventPumper interface {
	// PumpEvents processes pending platform events. Called from the main
	// goroutine on each iteration of the event loop. Implementations should
	// process all available events and return quickly.
	PumpEvents()
}

// Loop is the main event loop, integrating platform events with idle callbacks,
// timers, and cross-goroutine dispatching.
//
// The loop runs on a single goroutine (the main goroutine). All handlers,
// idle callbacks, and timer callbacks execute on this goroutine. The only
// concurrency is the readEvents goroutine that posts raw events to eventCh.
// Fields like rawHandler and idleQueue are not synchronized and must only
// be accessed from the loop goroutine or before calling Run().
type Loop struct {
	server     platform.DisplayServer
	parser     platform.EventParser
	hasIM      bool
	dispatcher *Dispatcher
	pumper     EventPumper // non-nil on platforms needing main-thread event pumping

	// Channels for the select-based event loop.
	// Buffer sizes provide natural backpressure: senders block when full,
	// which is acceptable since the X server queues events internally and
	// DoWhenIdle/After callers are on the event loop goroutine.
	eventCh chan *platform.RawEvent // raw events from reader goroutine (cap 64)
	idleCh  chan func()             // idle callbacks (cap 256)
	timerCh chan func()             // timer-fired callbacks (cap 64)
	mainCh  chan func()             // cross-goroutine calls via RunOnMain (cap 64)
	done    chan struct{}           // signal to stop the loop

	// Pending idle callbacks (coalesced).
	idleQueue []func()

	// rawHandler is called for every raw event before type conversion.
	// Used to handle event types (e.g. selection) not routed through Dispatcher.
	rawHandler func(*platform.RawEvent)

	// filter is called for every converted event before dispatch.
	filter func(*Event)
}

// NewLoop creates a new event loop for the given display server.
func NewLoop(server platform.DisplayServer, parser platform.EventParser, dispatcher *Dispatcher) *Loop {
	l := &Loop{
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
	// Check if the server needs main-thread event pumping (e.g. macOS/Cocoa).
	if p, ok := server.(EventPumper); ok {
		l.pumper = p
	}
	return l
}

// Run starts the event loop. It blocks until Quit is called.
func (l *Loop) Run() {
	l.run(nil)
}

// run is the unified event-loop body shared by Run and RunNested.
// extraDone may be nil for the top-level loop (Quit is the only way out)
// or the caller's done channel for a nested loop. Pump mode is enabled
// automatically when the display server implements EventPumper (macOS/Cocoa).
func (l *Loop) run(extraDone <-chan struct{}) {
	var pumpTick <-chan time.Time
	if l.pumper != nil {
		// Adaptive pump interval: start at 10ms (idle), drop to 1ms when
		// events are flowing. This avoids 500Hz spin when idle while
		// keeping responsiveness during interaction.
		const (
			minInterval = 1 * time.Millisecond
			maxInterval = 10 * time.Millisecond
		)
		interval := maxInterval
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		pumpTick = ticker.C

		// Track event activity to adjust interval dynamically.
		eventsSinceTick := 0

		// FD-based mode (X11) blocks on NextEvent in a goroutine; on pump
		// mode the reader polls the C ring buffer that PumpEvents feeds.
		go l.readEvents()

		for {
			l.processIdleQueue()

			select {
			case <-l.done:
				return

			case <-extraDone:
				return

			case raw := <-l.eventCh:
				l.handleRaw(raw)
				eventsSinceTick++

			case fn := <-l.idleCh:
				l.idleQueue = append(l.idleQueue, fn)

			case fn := <-l.timerCh:
				fn()
				l.server.Flush()

			case fn := <-l.mainCh:
				fn()
				l.server.Flush()

			case <-pumpTick:
				// Pump platform events from the main thread.
				// This processes NSEvents on macOS, which triggers
				// callbacks that post to eventCh.
				l.pumper.PumpEvents()

				// Drain any events that were posted during pumping.
				drainedCount := 0
				for drained := false; !drained; {
					select {
					case raw := <-l.eventCh:
						l.handleRaw(raw)
						drainedCount++
					default:
						drained = true
					}
				}
				eventsSinceTick += drainedCount

				// Adjust ticker interval based on recent event activity.
				// If events were processed this tick, speed up; otherwise slow down.
				if eventsSinceTick > 0 {
					if interval > minInterval {
						interval /= 2
						if interval < minInterval {
							interval = minInterval
						}
						ticker.Reset(interval)
					}
				} else if interval < maxInterval {
					interval *= 2
					if interval > maxInterval {
						interval = maxInterval
					}
					ticker.Reset(interval)
				}
				eventsSinceTick = 0
			}
		}
	}

	// FD-based mode (X11) blocks on NextEvent in a goroutine.
	go l.readEvents()

	for {
		l.processIdleQueue()

		select {
		case <-l.done:
			return

		case <-extraDone:
			return

		case raw := <-l.eventCh:
			l.handleRaw(raw)

		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)

		case fn := <-l.timerCh:
			fn()
			l.server.Flush()

		case fn := <-l.mainCh:
			fn()
			l.server.Flush()

		case <-pumpTick:
			// Pump platform events from the main thread.
			// This processes NSEvents on macOS, which triggers
			// callbacks that post to eventCh.
			l.pumper.PumpEvents()

			// Drain any events that were posted during pumping.
			for drained := false; !drained; {
				select {
				case raw := <-l.eventCh:
					l.handleRaw(raw)
				default:
					drained = true
				}
			}
		}
	}
}

// handleRaw processes one raw event: filters, converts, manages IC focus, and dispatches.
func (l *Loop) handleRaw(raw *platform.RawEvent) {
	if l.server.FilterEvent(raw) {
		return
	}
	if l.rawHandler != nil {
		l.rawHandler(raw)
	}
	ev := FromRawEventIM(raw, l.parser, l.hasIM)
	if ev.Type != 0 {
		if l.hasIM {
			switch ev.Type {
			case FocusInType:
				l.server.SetICFocus(ev.Window)
			case FocusOutType:
				l.server.UnsetICFocus()
			}
		}
		if l.filter != nil {
			l.filter(&ev)
		}
		l.dispatcher.Dispatch(&ev)
	}
	l.server.Flush()
}

// SetEventFilter installs a function that may rewrite each converted event
// before it is dispatched. Same threading rules as SetRawEventHandler.
func (l *Loop) SetEventFilter(f func(*Event)) {
	l.filter = f
}

// SetRawEventHandler installs a handler called for every raw event before
// type conversion. Use it for event types not routed through the Dispatcher
// (e.g. X11 selection events).
//
// Must be called before Run() or from a handler running on the event loop
// goroutine. Not safe for concurrent use.
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

// freezeTimers drops every After with a positive delay so demo screenshots
// are deterministic; scripts/demo_wrapper.tcl applies the same rule to Tcl's
// after command.
var freezeTimers = os.Getenv("TAKIGO_FREEZE_TIMERS") == "1"

// After schedules a function to run after the given duration.
func (l *Loop) After(d time.Duration, fn func()) {
	if freezeTimers && d > 0 {
		return
	}
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
	l.run(done)
}

// RunNestedContext processes events until the context is cancelled or the done channel is closed.
// This is the context-aware version of RunNested for cancellation support.
// It must be called from within a handler running on the main goroutine.
func (l *Loop) RunNestedContext(ctx context.Context, done <-chan struct{}) {
	// Create a channel that closes when context is done
	ctxDone := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(ctxDone)
	}()

	// Run with both channels - return when either closes
	l.run(mergeDoneChannels(ctxDone, done))
}

// mergeDoneChannels returns a channel that closes when either input channel closes.
func mergeDoneChannels(ch1, ch2 <-chan struct{}) <-chan struct{} {
	if ch1 == nil {
		return ch2
	}
	if ch2 == nil {
		return ch1
	}
	out := make(chan struct{})
	go func() {
		select {
		case <-ch1:
		case <-ch2:
		}
		close(out)
	}()
	return out
}

// processIdleQueue runs all pending idle callbacks.
func (l *Loop) processIdleQueue() {
	// Drain any pending idle callbacks from the channel.
	for drained := false; !drained; {
		select {
		case fn := <-l.idleCh:
			l.idleQueue = append(l.idleQueue, fn)
		default:
			drained = true
		}
	}
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
