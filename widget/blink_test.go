package widget

import (
	"context"
	"testing"
	"time"
)

type timerSched struct {
	delays []time.Duration
	fns    []func()
}

func (s *timerSched) DoWhenIdle(fn func()) { fn() }

func (s *timerSched) RunNestedLoop(<-chan struct{})                         {}
func (s *timerSched) RunNestedLoopContext(context.Context, <-chan struct{}) {}
func (s *timerSched) Quit()                                                 {}

func (s *timerSched) After(d time.Duration, fn func()) func() bool {
	s.delays = append(s.delays, d)
	s.fns = append(s.fns, fn)
	return func() bool { return true }
}

func TestCursorBlinkerOnOffTimes(t *testing.T) {
	s := &timerSched{}
	n := 0
	b := NewCursorBlinker(s, func() { n++ })
	b.Start()
	s.fns[0]()
	s.fns[1]()
	if s.delays[0] != InsertOnTime || s.delays[1] != InsertOffTime || s.delays[2] != InsertOnTime {
		t.Errorf("delays = %v", s.delays)
	}
	if n != 2 {
		t.Errorf("%d toggles, want 2", n)
	}
}

// A timer of a stopped run must not re-arm when the blinker restarts.
func TestCursorBlinkerStaleTimer(t *testing.T) {
	s := &timerSched{}
	n := 0
	b := NewCursorBlinker(s, func() { n++ })
	b.Start()
	stale := s.fns[0]
	b.Stop()
	b.Start()
	stale()
	if n != 0 || len(s.fns) != 2 {
		t.Errorf("stale timer ran: toggles %d, timers %d", n, len(s.fns))
	}
}
