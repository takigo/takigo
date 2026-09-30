package event

import "testing"

// Throughput of the posted-callback queues, run on the loop goroutine.
func BenchmarkRunMainQueue1000(b *testing.B) {
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	n := 0
	fn := func() { n++ }
	for b.Loop() {
		for range 1000 {
			l.RunOnMain(fn)
		}
		l.runMainQueue()
	}
}

func BenchmarkIdleQueue1000(b *testing.B) {
	l := NewLoop(newFakeServer(), nil, NewDispatcher())
	n := 0
	fn := func() { n++ }
	for b.Loop() {
		for range 1000 {
			l.DoWhenIdle(fn)
		}
		l.processIdleQueue()
	}
}
