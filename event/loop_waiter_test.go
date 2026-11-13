package event

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/takigo/takigo/platform"
)

// waitingServer is an EventWaiter like the X11 backend: events sit in a
// queue only the loop may read, and WaitForEvents reports the connection
// readable. queue adds an event without making it readable, like an event
// a round trip read off the connection.
type waitingServer struct {
	platform.DisplayServer
	mu       sync.Mutex
	events   []*platform.RawEvent
	readable chan struct{}
	stop     chan struct{}
	stopOnce sync.Once
	reading  atomic.Int32 // goroutines inside Pending or NextEvent
	overlaps atomic.Int32
}

func newWaitingServer() *waitingServer {
	return &waitingServer{readable: make(chan struct{}, 1), stop: make(chan struct{})}
}

func (s *waitingServer) HasIM() bool                         { return false }
func (s *waitingServer) FilterEvent(*platform.RawEvent) bool { return false }
func (s *waitingServer) Flush()                              {}

func (s *waitingServer) WakeEventReader() bool {
	s.stopOnce.Do(func() { close(s.stop) })
	return true
}

func (s *waitingServer) WaitForEvents() bool {
	select {
	case <-s.readable:
		return true
	case <-s.stop:
		return false
	}
}

func (s *waitingServer) enter() func() {
	if s.reading.Add(1) != 1 {
		s.overlaps.Add(1)
	}
	return func() { s.reading.Add(-1) }
}

func (s *waitingServer) Pending() int {
	defer s.enter()()
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.events)
}

func (s *waitingServer) NextEvent() *platform.RawEvent {
	defer s.enter()()
	s.mu.Lock()
	defer s.mu.Unlock()
	ev := s.events[0]
	s.events = s.events[1:]
	return ev
}

func (s *waitingServer) queue(seq int) {
	s.mu.Lock()
	s.events = append(s.events, mapEvent(seq))
	s.mu.Unlock()
}

func (s *waitingServer) send(seq int) {
	s.queue(seq)
	select {
	case s.readable <- struct{}{}:
	default:
	}
}

func TestLoopWaiterDeliversInOrder(t *testing.T) {
	const n = 1000
	srv := newWaitingServer()
	d := NewDispatcher()
	l := NewLoop(srv, nil, d)

	var got []int
	var nestDone chan struct{}
	d.BindGlobal(StructureNotifyMask, func(ev *Event) {
		seq := int(ev.Window)
		got = append(got, seq)
		switch {
		case seq == n-1:
			l.Quit()
		case seq%100 == 10:
			nestDone = make(chan struct{})
			l.RunNested(nestDone)
		case seq%100 == 60:
			close(nestDone)
		}
	})
	wait := runLoop(t, l)
	for i := range n {
		srv.send(i)
		if i%50 == 0 {
			time.Sleep(time.Millisecond) // let the loop block and re-arm
		}
	}
	wait()

	if len(got) != n {
		t.Fatalf("dispatched %d events, want %d", len(got), n)
	}
	for i, seq := range got {
		if seq != i {
			t.Fatalf("event %d delivered as #%d: events reordered", seq, i)
		}
	}
	if o := srv.overlaps.Load(); o != 0 {
		t.Errorf("Pending/NextEvent ran on two goroutines at once %d times", o)
	}
}

// An event read off the connection by a callback's round trip leaves
// nothing to wake the reader: the loop must find it before it blocks.
func TestLoopWaiterSeesEventsQueuedByCallbacks(t *testing.T) {
	srv := newWaitingServer()
	d := NewDispatcher()
	l := NewLoop(srv, nil, d)

	var got []int
	d.BindGlobal(StructureNotifyMask, func(ev *Event) {
		got = append(got, int(ev.Window))
		if len(got) == 3 {
			l.Quit()
		}
	})
	l.After(10*time.Millisecond, func() { srv.queue(0) })
	l.After(20*time.Millisecond, func() { l.DoWhenIdle(func() { srv.queue(1) }) })
	l.After(30*time.Millisecond, func() { l.RunOnMain(func() { srv.queue(2) }) })
	runLoop(t, l)()
	if len(got) != 3 {
		t.Fatalf("dispatched %v, want 3 events", got)
	}
}

func TestLoopWaiterStopWaitsForReader(t *testing.T) {
	srv := newWaitingServer()
	l := NewLoop(srv, nil, NewDispatcher())
	stopped := false
	l.After(10*time.Millisecond, func() {
		l.Stop(5 * time.Second)
		select {
		case <-l.readerDone:
			stopped = true
		default:
		}
	})
	runLoop(t, l)()
	if !stopped {
		t.Fatal("Stop returned while the reader goroutine was still running")
	}
}
