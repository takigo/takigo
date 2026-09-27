package event

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/msorc/takigo/platform"
)

// fakeServer is a DisplayServer whose NextEvent reads from a channel.
// Methods the loop does not call are left to the nil embedded interface.
type fakeServer struct {
	platform.DisplayServer
	events      chan *platform.RawEvent
	inflight    atomic.Int32
	maxInflight atomic.Int32
}

func newFakeServer() *fakeServer {
	return &fakeServer{events: make(chan *platform.RawEvent, 4096)}
}

func (f *fakeServer) HasIM() bool                         { return false }
func (f *fakeServer) FilterEvent(*platform.RawEvent) bool { return false }
func (f *fakeServer) Flush()                              {}

func (f *fakeServer) NextEvent() *platform.RawEvent {
	n := f.inflight.Add(1)
	for {
		m := f.maxInflight.Load()
		if n <= m || f.maxInflight.CompareAndSwap(m, n) {
			break
		}
	}
	ev := <-f.events
	f.inflight.Add(-1)
	return ev
}

// mapEvent returns a MapNotify raw event; its window ID carries seq so
// tests can check delivery order without an EventParser.
func mapEvent(seq int) *platform.RawEvent {
	return &platform.RawEvent{EventType: platform.MapNotifyEvent, EventWindow: platform.WindowID(seq)}
}

// runLoop runs l.Run on its own goroutine and returns a function that
// waits for it to finish, failing the test after a timeout.
func runLoop(t *testing.T, l *Loop) func() {
	t.Helper()
	finished := make(chan struct{})
	go func() {
		l.Run()
		close(finished)
	}()
	return func() {
		t.Helper()
		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			l.Quit()
			t.Fatal("event loop did not finish (deadlock?)")
		}
	}
}

func TestLoopNestedKeepsOrderAndOneReader(t *testing.T) {
	const n = 1000
	srv := newFakeServer()
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
	for i := range n {
		srv.events <- mapEvent(i)
	}
	runLoop(t, l)()

	if len(got) != n {
		t.Fatalf("dispatched %d events, want %d", len(got), n)
	}
	for i, seq := range got {
		if seq != i {
			t.Fatalf("event %d delivered as #%d: events reordered", seq, i)
		}
	}
	if m := srv.maxInflight.Load(); m != 1 {
		t.Errorf("max concurrent NextEvent callers = %d, want 1", m)
	}
}

func TestLoopQueuesDoNotBlockOnLoopGoroutine(t *testing.T) {
	const n = 1000
	l := NewLoop(newFakeServer(), nil, NewDispatcher())

	var idleRuns, mainRuns int
	check := func() {
		if idleRuns == n && mainRuns == n {
			l.Quit()
		}
	}
	// Posting from the loop goroutine used to deadlock once the
	// buffered channels (256 idle, 64 main) filled up.
	l.DoWhenIdle(func() {
		for range n {
			l.DoWhenIdle(func() { idleRuns++; check() })
			l.RunOnMain(func() { mainRuns++; check() })
		}
	})
	runLoop(t, l)()

	if idleRuns != n || mainRuns != n {
		t.Fatalf("idle ran %d, main ran %d; want %d each", idleRuns, mainRuns, n)
	}
}

func TestLoopDoWhenIdleBeforeRun(t *testing.T) {
	const n = 1000
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	runs := 0
	for range n {
		l.DoWhenIdle(func() { runs++ })
	}
	l.DoWhenIdle(l.Quit)
	runLoop(t, l)()
	if runs != n {
		t.Fatalf("idle callbacks ran %d times, want %d", runs, n)
	}
}

func TestLoopAfterRunsOnLoop(t *testing.T) {
	if freezeTimers {
		t.Skip("TAKIGO_FREEZE_TIMERS drops positive-delay timers")
	}
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	fired := false
	l.After(time.Millisecond, func() {
		fired = true
		l.Quit()
	})
	runLoop(t, l)()
	if !fired {
		t.Fatal("After callback did not run")
	}
}

func TestLoopQuitConcurrent(t *testing.T) {
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	wait := runLoop(t, l)
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(l.Quit)
	}
	wg.Wait()
	wait()
	l.Quit()
	l.RunOnMain(func() { t.Error("RunOnMain callback ran after Quit") })
}

func TestLoopRunNestedContext(t *testing.T) {
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	var before, after int
	l.DoWhenIdle(func() {
		// Warm up so the reader goroutine already exists.
		warm := make(chan struct{})
		l.DoWhenIdle(func() { close(warm) })
		l.RunNested(warm)

		before = runtime.NumGoroutine()
		for range 20 {
			done := make(chan struct{})
			l.DoWhenIdle(func() { close(done) })
			l.RunNestedContext(context.Background(), done)
		}
		ctx, cancel := context.WithCancel(context.Background())
		l.DoWhenIdle(cancel)
		l.RunNestedContext(ctx, nil)
		after = runtime.NumGoroutine()
		l.Quit()
	})
	runLoop(t, l)()
	if after > before {
		t.Errorf("goroutines grew from %d to %d across nested loops", before, after)
	}
}
