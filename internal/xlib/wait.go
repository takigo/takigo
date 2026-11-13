//go:build linux || freebsd || openbsd || netbsd

package xlib

/*
#include <errno.h>
#include <poll.h>

// takigo_wait_readable returns 1 when xfd is readable (or broken, which the
// next Xlib call reports) and 0 when wakefd is, or on an error.
static int takigo_wait_readable(int xfd, int wakefd)
{
	struct pollfd p[2] = {{xfd, POLLIN, 0}, {wakefd, POLLIN, 0}};
	for (;;) {
		if (poll(p, 2, -1) < 0) {
			if (errno == EINTR) continue;
			return 0;
		}
		if (p[1].revents) return 0;
		if (p[0].revents) return 1;
	}
}
*/
import "C"

// WaitReadable blocks until the connection has data to read and returns
// true, or until StopWaiting was called and returns false. It makes no Xlib
// call, so it may block on another goroutine than the one using the
// display: a thread blocked in XNextEvent makes libX11 abort when the other
// one reads an X error during a round trip (the assertion on event_waiter
// in xcb_io.c's poll_for_event).
func (d *Display) WaitReadable() bool {
	return C.takigo_wait_readable(d.fd, d.wakeFd) == 1
}

// StopWaiting makes this and every later WaitReadable return false. Safe
// from any goroutine and after Close.
func (d *Display) StopWaiting() {
	_, _ = d.wakeW.Write([]byte{0})
}
