package platform

import (
	"sync"
	"testing"
	"time"
)

func TestEventQueueNeverDropsAndKeepsOrder(t *testing.T) {
	const producers, perProducer = 4, 5000
	q := NewEventQueue()
	var wg sync.WaitGroup
	for p := range producers {
		wg.Go(func() {
			for i := range perProducer {
				q.Push(&RawEvent{EventType: p, EventWindow: WindowID(i)})
			}
		})
	}
	next := make([]WindowID, producers)
	for range producers * perProducer {
		ev := q.Pop()
		if ev.EventWindow != next[ev.EventType] {
			t.Fatalf("producer %d: got event %d, want %d", ev.EventType, ev.EventWindow, next[ev.EventType])
		}
		next[ev.EventType]++
	}
	wg.Wait()
	if n := q.Len(); n != 0 {
		t.Errorf("Len() = %d after draining, want 0", n)
	}
}

func TestEventQueuePushFromConsumerGoroutine(t *testing.T) {
	// A window procedure posts on the loop goroutine; far more events than
	// the old 256-slot channel held must neither block nor be dropped.
	q := NewEventQueue()
	for i := range 10000 {
		q.Push(&RawEvent{EventWindow: WindowID(i)})
	}
	for i := range 10000 {
		if ev := q.Pop(); ev.EventWindow != WindowID(i) {
			t.Fatalf("event %d popped as %d", i, ev.EventWindow)
		}
	}
}

func TestEventQueuePopBlocksUntilPush(t *testing.T) {
	q := NewEventQueue()
	got := make(chan *RawEvent)
	go func() { got <- q.Pop() }()
	select {
	case <-got:
		t.Fatal("Pop returned from an empty queue")
	case <-time.After(20 * time.Millisecond):
	}
	want := &RawEvent{EventType: 7}
	q.Push(want)
	select {
	case ev := <-got:
		if ev != want {
			t.Fatalf("Pop returned %v, want %v", ev, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Pop did not wake after Push")
	}
}
