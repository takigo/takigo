package ttk_test

import (
	"testing"
	"time"

	"github.com/takigo/takigo/internal/testutil"
	"github.com/takigo/takigo/ttk"
	_ "github.com/takigo/takigo/ttk/defaulttheme"
)

func TestProgressbarRestartRunsOneTimerChain(t *testing.T) {
	app := testutil.NewTestApp(t)
	p := ttk.NewProgressbar(app, "p", ttk.ProgressbarMaximum(1000))

	const interval = 20 * time.Millisecond
	p.Start(interval)
	p.Stop()
	p.Start(interval) // the first chain must not keep ticking
	app.After(110*time.Millisecond, app.Quit)
	app.MainLoop()

	// Two Start steps plus at most one per interval from a single chain;
	// a leaked second chain roughly doubles this.
	if max := 2 + int(110*time.Millisecond/interval) + 1; p.Value > float64(max) {
		t.Errorf("value = %v after restart, want <= %d (stale timer chain still running)", p.Value, max)
	}
	if p.Value < 2 {
		t.Errorf("value = %v, want the animation to have stepped", p.Value)
	}
}
