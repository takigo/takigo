//go:build linux || freebsd || openbsd || netbsd

#include <stdio.h>
#include "xerror.h"

static unsigned long takigo_x_errors;

// The trap. Xlib calls the error handler on whichever thread reads the
// error from the connection (the event reader's, usually), so the fields
// are published with atomics: serial and hit before dpy.
static Display *trap_dpy;
static unsigned long trap_serial;
static int trap_hit;

unsigned long takigo_x_error_count(void)
{
	return __atomic_load_n(&takigo_x_errors, __ATOMIC_RELAXED);
}

// takigo_x_error reports an X protocol error and carries on. Xlib's default
// handler exits the process; Tk's ErrorProc (tkError.c) ignores errors on
// windows it is destroying, which takigo cannot yet tell apart.
//
// It is the only handler the process ever installs: a second Xlib error
// handler swapped in and out around a request would race with other
// displays' threads doing the same.
static int takigo_x_error(Display *dpy, XErrorEvent *ev)
{
	if (dpy == __atomic_load_n(&trap_dpy, __ATOMIC_SEQ_CST) &&
	    ev->serial >= __atomic_load_n(&trap_serial, __ATOMIC_SEQ_CST)) {
		__atomic_store_n(&trap_hit, 1, __ATOMIC_SEQ_CST);
		return 0;
	}
	__atomic_add_fetch(&takigo_x_errors, 1, __ATOMIC_RELAXED);
	char text[128];
	XGetErrorText(dpy, ev->error_code, text, sizeof text);
	fprintf(stderr, "takigo: X error: %s (request %d.%d, resource 0x%lx)\n",
		text, ev->request_code, ev->minor_code, ev->resourceid);
	return 0;
}

void takigo_init_xlib(void)
{
	XInitThreads();
	XSetErrorHandler(takigo_x_error);
}

void takigo_trap_errors(Display *dpy)
{
	XSync(dpy, False); // errors of earlier requests are reported as usual
	__atomic_store_n(&trap_serial, NextRequest(dpy), __ATOMIC_SEQ_CST);
	__atomic_store_n(&trap_hit, 0, __ATOMIC_SEQ_CST);
	__atomic_store_n(&trap_dpy, dpy, __ATOMIC_SEQ_CST);
}

int takigo_untrap_errors(Display *dpy)
{
	XSync(dpy, False);
	__atomic_store_n(&trap_dpy, NULL, __ATOMIC_SEQ_CST);
	return __atomic_load_n(&trap_hit, __ATOMIC_SEQ_CST);
}
