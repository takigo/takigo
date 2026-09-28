package platform

import "sync"

// EventQueue is an unbounded FIFO of raw events for backends whose native
// callbacks produce events (a Win32 window procedure, Cocoa handlers).
// Push never blocks or drops: those callbacks run on the event-loop
// goroutine, so waiting for room would deadlock, and a dropped
// ButtonRelease or KeyRelease would leave a grab or key stuck. Pop blocks
// until an event is available.
type EventQueue struct {
	mu     sync.Mutex
	events []*RawEvent
	ready  chan struct{} // cap 1: at least one Push since the last Pop wait
}

// NewEventQueue returns an empty queue.
func NewEventQueue() *EventQueue {
	return &EventQueue{ready: make(chan struct{}, 1)}
}

// Push appends ev. It is safe from any goroutine.
func (q *EventQueue) Push(ev *RawEvent) {
	q.mu.Lock()
	q.events = append(q.events, ev)
	q.mu.Unlock()
	select {
	case q.ready <- struct{}{}:
	default:
	}
}

// Pop removes and returns the oldest event, blocking while the queue is
// empty.
func (q *EventQueue) Pop() *RawEvent {
	for {
		q.mu.Lock()
		if len(q.events) > 0 {
			ev := q.events[0]
			q.events[0] = nil
			q.events = q.events[1:]
			if len(q.events) == 0 {
				q.events = nil
			}
			q.mu.Unlock()
			return ev
		}
		q.mu.Unlock()
		<-q.ready
	}
}

// Len returns the number of queued events.
func (q *EventQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.events)
}
