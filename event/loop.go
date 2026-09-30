package event

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/msorc/takigo/platform"
)

// EventPumper is an optional interface that platform backends can implement
// when they need to be periodically polled from the main goroutine.
// macOS/Cocoa requires this because NSEvents must be pumped on the main thread.
type EventPumper interface {
	// PumpEvents processes pending platform events. Called from the main
	// goroutine on each iteration of the event loop. Implementations should
	// process all available events and return quickly. It returns true
	// when the platform asked the application to quit (Win32 WM_QUIT);
	// the loop then quits.
	PumpEvents() (quit bool)
}

// Loop is the main event loop, integrating platform events with idle callbacks,
// timers, and cross-goroutine dispatching.
//
// The loop runs on a single goroutine (the main goroutine). All handlers,
// idle callbacks, and timer callbacks execute on this goroutine. The only
// other goroutines are the single readEvents goroutine, which posts raw
// events to eventCh, and timer goroutines, which only enqueue callbacks.
// Fields like rawHandler and idleQueue are not synchronized and must only
// be accessed from the loop goroutine or before calling Run().
type Loop struct {
	server     platform.DisplayServer
	parser     platform.EventParser
	hasIM      bool
	dispatcher *Dispatcher
	pumper     EventPumper // non-nil on platforms needing main-thread event pumping

	// eventCh carries raw events from the reader goroutine. The reader is
	// started once, however many nested loops run, so events keep the
	// order the platform delivered them in.
	eventCh    chan *platform.RawEvent
	readerOnce sync.Once
	readerDone chan struct{} // closed when readEvents returns

	done     chan struct{}
	quitOnce sync.Once

	// Idle and main-thread callbacks are queued under mu so that posting
	// never blocks, even from a handler on the loop goroutine itself.
	// wake (cap 1) tells the loop there is queued work.
	mu          sync.Mutex
	idlePending []func()
	mainPending []func()
	wake        chan struct{}

	// idleQueue and mainQueue hold the callbacks taken from idlePending
	// and mainPending, in order; loop-only. They are fields, not locals of
	// the functions running them, so that a callback running a nested loop
	// (a modal dialog) leaves the callbacks queued behind it to that loop.
	idleQueue []func()
	mainQueue []func()

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
		readerDone: make(chan struct{}),
		done:       make(chan struct{}),
		wake:       make(chan struct{}, 1),
	}
	// Check if the server needs main-thread event pumping (e.g. macOS/Cocoa).
	if p, ok := server.(EventPumper); ok {
		l.pumper = p
	}
	return l
}

// Run starts the event loop. It blocks until Quit is called.
func (l *Loop) Run() {
	l.run(nil, nil)
}

// run is the event-loop body shared by Run, RunNested and RunNestedContext.
// extraDone and ctxDone may be nil; the loop also returns when Quit is
// called. Pump mode is enabled when the display server implements
// EventPumper (macOS/Cocoa, Windows).
func (l *Loop) run(extraDone, ctxDone <-chan struct{}) {
	l.readerOnce.Do(func() { go l.readEvents() })

	// Adaptive pump interval: 10ms when idle, down to 1ms while events
	// are flowing, so pump mode does not spin at a fixed high rate.
	const (
		minInterval = 1 * time.Millisecond
		maxInterval = 10 * time.Millisecond
	)
	var pumpTick <-chan time.Time
	var ticker *time.Ticker
	interval := maxInterval
	if l.pumper != nil {
		ticker = time.NewTicker(interval)
		defer ticker.Stop()
		pumpTick = ticker.C
	}
	eventsSinceTick := 0

	// Like Tcl_DoOneEvent, idle callbacks run only once no event is
	// waiting, so a burst of Motion or Configure events costs one redraw
	// and one relayout rather than one per event. maxIdleDefer bounds how
	// long a steady stream can hold them off.
	const maxIdleDefer = 64
	deferred := 0

	for {
		if len(l.eventCh) == 0 || deferred >= maxIdleDefer {
			l.processIdleQueue()
			deferred = 0
		} else {
			deferred++
		}
		if len(l.mainQueue) > 0 {
			// Left behind by the callback that started this nested loop.
			l.runMainQueue()
		}

		select {
		case <-l.done:
			return

		case <-extraDone:
			return

		case <-ctxDone:
			return

		case raw := <-l.eventCh:
			l.handleRaw(raw)
			eventsSinceTick++

		case <-l.wake:
			l.runMainQueue()

		case <-pumpTick:
			if l.pumper.PumpEvents() {
				l.Quit()
				return
			}
			for drained := false; !drained; {
				select {
				case raw := <-l.eventCh:
					l.handleRaw(raw)
					eventsSinceTick++
				default:
					drained = true
				}
			}
			if eventsSinceTick > 0 {
				if interval > minInterval {
					interval = max(interval/2, minInterval)
					ticker.Reset(interval)
				}
			} else if interval < maxInterval {
				interval = min(interval*2, maxInterval)
				ticker.Reset(interval)
			}
			eventsSinceTick = 0
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

// Quit stops the event loop and every nested loop running inside it.
// Safe to call more than once and from any goroutine.
func (l *Loop) Quit() {
	l.quitOnce.Do(func() { close(l.done) })
}

// Done returns a channel that is closed once Quit has been called.
func (l *Loop) Done() <-chan struct{} {
	return l.done
}

// Stop quits the loop and waits up to timeout for the reader goroutine to
// return, so the display can be closed without a read still in flight on
// it. Call it from the loop goroutine or after Run has returned.
func (l *Loop) Stop(timeout time.Duration) {
	l.Quit()
	started := true
	l.readerOnce.Do(func() { started = false })
	if !started || !l.server.WakeEventReader() {
		return
	}
	select {
	case <-l.readerDone:
	case <-time.After(timeout):
	}
}

// DoWhenIdle schedules fn to run once the loop has no events to process.
// It never blocks and is safe from any goroutine, including handlers on
// the loop goroutine. Each call runs fn once; callers that want
// coalescing keep their own pending flag (see canvas.scheduleRedraw).
func (l *Loop) DoWhenIdle(fn func()) {
	l.mu.Lock()
	l.idlePending = append(l.idlePending, fn)
	l.mu.Unlock()
	l.signal()
}

// freezeTimers drops every After with a positive delay so demo screenshots
// are deterministic; scripts/demo_wrapper.tcl applies the same rule to Tcl's
// after command.
var freezeTimers = os.Getenv("TAKIGO_FREEZE_TIMERS") == "1"

// After schedules fn to run on the loop goroutine after duration d. It
// returns a cancel function, Tcl's "after cancel": once cancel returns,
// fn will not run, and cancel reports whether it stopped fn (false if fn
// already ran or was cancelled before).
func (l *Loop) After(d time.Duration, fn func()) (cancel func() bool) {
	const pending, ran, cancelled = 0, 1, 2
	var state atomic.Int32
	stop := func() bool { return state.CompareAndSwap(pending, cancelled) }
	if freezeTimers && d > 0 {
		return stop
	}
	t := time.AfterFunc(d, func() {
		l.RunOnMain(func() {
			if state.CompareAndSwap(pending, ran) {
				fn()
			}
		})
	})
	return func() bool {
		t.Stop()
		return stop()
	}
}

// RunOnMain schedules fn to run on the event loop goroutine. It never
// blocks and is safe to call from any goroutine. Callbacks posted after
// Quit are dropped.
func (l *Loop) RunOnMain(fn func()) {
	select {
	case <-l.done:
		return
	default:
	}
	l.mu.Lock()
	l.mainPending = append(l.mainPending, fn)
	l.mu.Unlock()
	l.signal()
}

// signal wakes the loop without blocking; one pending wakeup is enough
// because the loop drains every queue when it wakes.
func (l *Loop) signal() {
	select {
	case l.wake <- struct{}{}:
	default:
	}
}

// RunNested processes events until the done channel is closed.
// This implements nested event loops needed for modal dialogs (like Tcl's vwait).
// It must be called from within a handler running on the main goroutine.
func (l *Loop) RunNested(done <-chan struct{}) {
	l.run(done, nil)
}

// RunNestedContext processes events until the context is cancelled or the done channel is closed.
// This is the context-aware version of RunNested for cancellation support.
// It must be called from within a handler running on the main goroutine.
func (l *Loop) RunNestedContext(ctx context.Context, done <-chan struct{}) {
	l.run(done, ctx.Done())
}

// UpdateIdleTasks runs idle callbacks until none are pending, like Tcl's
// "update idletasks", so pending layout and redraws happen now. Idle
// callbacks that keep rescheduling themselves are cut off after 100
// rounds. Loop goroutine only (or before Run).
func (l *Loop) UpdateIdleTasks() {
	for range 100 {
		l.mu.Lock()
		n := len(l.idlePending)
		l.mu.Unlock()
		if n == 0 && len(l.idleQueue) == 0 {
			return
		}
		l.processIdleQueue()
	}
}

// runMainQueue runs the callbacks posted with RunOnMain or After.
func (l *Loop) runMainQueue() {
	l.mu.Lock()
	l.mainQueue = append(l.mainQueue, l.mainPending...)
	clear(l.mainPending)
	l.mainPending = l.mainPending[:0]
	l.mu.Unlock()
	if runQueue(&l.mainQueue) {
		l.server.Flush()
	}
}

// processIdleQueue runs all pending idle callbacks. Callbacks scheduled
// while it runs wait for the next round, as in Tcl_DoOneEvent.
func (l *Loop) processIdleQueue() {
	l.mu.Lock()
	l.idleQueue = append(l.idleQueue, l.idlePending...)
	clear(l.idlePending)
	l.idlePending = l.idlePending[:0]
	l.mu.Unlock()
	if runQueue(&l.idleQueue) {
		l.server.Flush()
	}
}

// runQueue runs the callbacks that are in *q now, taking them off one at a
// time so that a nested loop started by one of them sees, and runs, the
// rest (TclServiceIdle does the same with its idle generation). It reports
// whether there was anything to run.
func runQueue(q *[]func()) bool {
	n := len(*q)
	for range n {
		if len(*q) == 0 {
			break // a nested loop ran the rest
		}
		fn := (*q)[0]
		(*q)[0] = nil
		*q = (*q)[1:]
		fn()
	}
	if len(*q) == 0 {
		*q = nil
	}
	return n > 0
}

// readEvents runs in its own goroutine for the lifetime of the loop,
// blocking on NextEvent and posting raw events to eventCh.
func (l *Loop) readEvents() {
	defer close(l.readerDone)
	for {
		raw := l.server.NextEvent()
		select {
		case <-l.done:
			return
		default:
		}
		select {
		case l.eventCh <- raw:
		case <-l.done:
			return
		}
	}
}
