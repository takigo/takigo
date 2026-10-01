//go:build unix

package displaylock

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	virtualOnce sync.Once
	virtual     bool // DISPLAY is this binary's private Xvfb
)

// UseVirtualDisplay points DISPLAY at an Xvfb server private to the running
// test binary, started on first use and stopped with the binary, so display
// tests neither open windows on the developer's screen nor meet a window
// manager and the user's applications. Call it before reading DISPLAY.
//
// TAKIGO_TEST_DISPLAY=real keeps the DISPLAY of the environment, for
// looking at what the tests draw. When Xvfb is not installed the
// environment's display is used as it is.
func UseVirtualDisplay() {
	virtualOnce.Do(func() {
		if os.Getenv("TAKIGO_TEST_DISPLAY") == "real" {
			return
		}
		path, err := exec.LookPath("Xvfb")
		if err != nil {
			return
		}
		r, w, err := os.Pipe()
		if err != nil {
			return
		}
		defer r.Close()

		started := make(chan bool, 1)
		go func() {
			// The server dies with this thread (Pdeathsig), so the goroutine
			// must keep its thread for as long as the server runs.
			runtime.LockOSThread()
			cmd := exec.Command(path, "-screen", "0", "1280x1024x24", "-noreset", "-nolisten", "tcp", "-displayfd", "3")
			cmd.ExtraFiles = []*os.File{w}
			dieWithParent(cmd)
			err := cmd.Start()
			w.Close()
			started <- err == nil
			if err == nil {
				_ = cmd.Wait()
			}
		}()
		if !<-started {
			return
		}
		line := make(chan string, 1)
		go func() {
			s, _ := bufio.NewReader(r).ReadString('\n')
			line <- s
		}()
		select {
		case s := <-line:
			if n := strings.TrimSpace(s); n != "" {
				_ = os.Setenv("DISPLAY", ":"+n)
				_ = os.Unsetenv("WAYLAND_DISPLAY")
				virtual = true
			}
		case <-time.After(10 * time.Second):
		}
	})
}
