package event

import (
	"testing"

	"github.com/msorc/takigo/platform"
)

func BenchmarkDispatcher_Dispatch(b *testing.B) {
	d := NewDispatcher()
	w := platform.WindowID(1)

	// Register some handlers
	for i := 0; i < 10; i++ {
		d.Bind(w, Mask(i+1), func(ev *Event) {})
		d.BindGlobal(Mask(i+1), func(ev *Event) {})
	}

	ev := &Event{
		Type:  ConfigureType,
		Window: w,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Dispatch(ev)
	}
}

func BenchmarkDispatcher_Bind(b *testing.B) {
	d := NewDispatcher()
	w := platform.WindowID(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Bind(w, Mask(i%10), func(ev *Event) {})
	}
}

func BenchmarkDispatcher_UnbindID(b *testing.B) {
	d := NewDispatcher()
	w := platform.WindowID(1)

	// Pre-register handlers
	ids := make([]BindingID, 1000)
	for i := 0; i < 1000; i++ {
		ids[i] = d.Bind(w, Mask(i%10), func(ev *Event) {})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.UnbindID(ids[i%1000])
	}
}