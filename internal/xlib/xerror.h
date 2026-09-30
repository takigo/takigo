// X protocol error handling shared by the cgo files of this package.
#ifndef TAKIGO_XERROR_H
#define TAKIGO_XERROR_H

#include <X11/Xlib.h>

// takigo_init_xlib runs XInitThreads and installs the error handler. Call
// it once, before the first Xlib call.
void takigo_init_xlib(void);

// takigo_x_error_count returns how many X errors have been reported.
unsigned long takigo_x_error_count(void);

// takigo_trap_errors makes the errors of dpy's requests from now on set a
// flag instead of being reported, like a Tk_CreateErrorHandler with no
// handler procedure; takigo_untrap_errors ends that and returns whether
// there was one. The trap is process-wide: the caller holds trapMu.
void takigo_trap_errors(Display *dpy);
int takigo_untrap_errors(Display *dpy);

#endif
