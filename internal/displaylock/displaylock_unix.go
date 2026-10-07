//go:build unix

// Package displaylock serializes display tests across test binaries.
// go test runs the packages' binaries in parallel against one X screen,
// where their windows overlap, so a test reading pixels back sees another
// binary's window. It imports nothing from takigo, so every test helper can
// use it.
package displaylock

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

var (
	mu      sync.Mutex
	file    *os.File
	holders int
)

// Require skips t when no display is available, after pointing DISPLAY at
// the private Xvfb when one can be started (UseVirtualDisplay), and holds
// the display lock for t. Every display test starts with it, or with a
// helper that calls it.
func Require(t testing.TB) {
	t.Helper()
	UseVirtualDisplay()
	if os.Getenv("DISPLAY") == "" && os.Getenv("WAYLAND_DISPLAY") == "" {
		t.Skip("no display available (set DISPLAY or use xvfb-run)")
	}
	Acquire(t)
}

// Acquire holds the display's lock file until t and its cleanups finish.
// A process holds the lock once however many tests in it have acquired it:
// flock on a second descriptor would wait for the first.
func Acquire(t testing.TB) {
	t.Helper()
	UseVirtualDisplay()
	if virtual {
		return // this binary has a display of its own: nothing to share
	}
	mu.Lock()
	defer mu.Unlock()
	if holders == 0 {
		name := strings.NewReplacer("/", "_", ":", "_").Replace(os.Getenv("DISPLAY") + os.Getenv("WAYLAND_DISPLAY"))
		f, err := os.OpenFile(filepath.Join(os.TempDir(), "takigo-display-"+name+".lock"), os.O_CREATE|os.O_RDWR, 0o666)
		if err != nil {
			t.Logf("displaylock: %v; running unlocked", err)
			return
		}
		for {
			err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
			if err != syscall.EINTR {
				break
			}
		}
		if err != nil {
			f.Close()
			t.Logf("displaylock: %v; running unlocked", err)
			return
		}
		file = f
	}
	holders++
	t.Cleanup(release)
}

func release() {
	mu.Lock()
	defer mu.Unlock()
	if holders--; holders == 0 {
		file.Close() // closing the descriptor drops the lock
		file = nil
	}
}
