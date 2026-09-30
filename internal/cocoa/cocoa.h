// cocoa.h — C interface for Cocoa/AppKit/CoreGraphics used by cgo.
// All Objective-C is confined to cocoa.m; Go calls only C functions declared here.

#ifndef TAKIGO_COCOA_H
#define TAKIGO_COCOA_H

#include <stdint.h>
#include <stdbool.h>

// Opaque handle types matching platform layer expectations.
typedef uintptr_t CocoaWindowID;
typedef uintptr_t CocoaDrawableID;  // NSView* or pixmap CGContext
typedef uintptr_t CocoaPixmapID;
typedef uintptr_t CocoaGCID;
typedef uintptr_t CocoaCursorID;

// ---- Application lifecycle ----

// CocoaInit initializes NSApplication on the main thread.
// Must be called from the main goroutine (runtime.LockOSThread).
void CocoaInit(void);

// CocoaRun runs the Cocoa event loop. Blocks until CocoaStop is called.
void CocoaRun(void);

// CocoaStop stops the Cocoa event loop.
void CocoaStop(void);

// CocoaFlush flushes pending drawing operations.
void CocoaFlush(void);

// ---- Screen info ----
int CocoaScreenWidth(void);
int CocoaScreenHeight(void);
int CocoaScreenWidthMM(void);
int CocoaScreenHeightMM(void);
int CocoaScreenDepth(void);
double CocoaScreenBackingScale(void);

// ---- Window management ----

CocoaWindowID CocoaCreateWindow(CocoaWindowID parent, int x, int y,
                                 unsigned int width, unsigned int height,
                                 unsigned int borderWidth, uint64_t bgPixel,
                                 int64_t eventMask, bool overrideRedirect);

CocoaWindowID CocoaCreateSimpleWindow(CocoaWindowID parent, int x, int y,
                                       unsigned int width, unsigned int height,
                                       unsigned int borderWidth,
                                       uint64_t border, uint64_t background);

void CocoaDestroyWindow(CocoaWindowID w);
void CocoaMapWindow(CocoaWindowID w);
void CocoaMapRaised(CocoaWindowID w);
void CocoaUnmapWindow(CocoaWindowID w);
void CocoaRaiseWindow(CocoaWindowID w);
void CocoaLowerWindow(CocoaWindowID w);
void CocoaMoveWindow(CocoaWindowID w, int x, int y);
void CocoaResizeWindow(CocoaWindowID w, unsigned int width, unsigned int height);
void CocoaMoveResizeWindow(CocoaWindowID w, int x, int y,
                            unsigned int width, unsigned int height);
void CocoaSelectInput(CocoaWindowID w, int64_t eventMask);
void CocoaStoreName(CocoaWindowID w, const char *name);
void CocoaTranslateCoordinates(CocoaWindowID src, CocoaWindowID dst,
                                int srcX, int srcY, int *dstX, int *dstY);
void CocoaSetWindowBackground(CocoaWindowID w, uint64_t pixel);
void CocoaClearWindow(CocoaWindowID w);
void CocoaClearArea(CocoaWindowID w, int x, int y,
                     unsigned int width, unsigned int height, bool exposures);
void CocoaSetInputFocus(CocoaWindowID w);

// ---- Graphics context emulation ----
// On macOS we emulate X11 GCs as lightweight structs holding drawing state.

CocoaGCID CocoaCreateGC(uint64_t fg, uint64_t bg, int lineWidth, int function);
void CocoaFreeGC(CocoaGCID gc);
void CocoaSetForeground(CocoaGCID gc, uint64_t pixel);
void CocoaSetBackground(CocoaGCID gc, uint64_t pixel);
void CocoaSetLineAttributes(CocoaGCID gc, unsigned int lineWidth,
                             int lineStyle, int capStyle, int joinStyle);
void CocoaSetFillStyle(CocoaGCID gc, int fillStyle);
void CocoaSetDashes(CocoaGCID gc, int dashOffset, const unsigned char *dashList, int n);

// ---- Drawing primitives ----
// All drawing targets a CocoaDrawableID (NSView or off-screen CGContext).

void CocoaFillRectangle(CocoaDrawableID d, CocoaGCID gc,
                         int x, int y, unsigned int w, unsigned int h);
void CocoaDrawRectangle(CocoaDrawableID d, CocoaGCID gc,
                         int x, int y, unsigned int w, unsigned int h);
void CocoaDrawLine(CocoaDrawableID d, CocoaGCID gc,
                    int x1, int y1, int x2, int y2);
void CocoaDrawLines(CocoaDrawableID d, CocoaGCID gc,
                     const int16_t *points, int npoints, int mode);
void CocoaFillPolygon(CocoaDrawableID d, CocoaGCID gc,
                       const int16_t *points, int npoints, int shape, int mode);
void CocoaFillArc(CocoaDrawableID d, CocoaGCID gc,
                   int x, int y, unsigned int w, unsigned int h,
                   int angle1, int angle2);
void CocoaDrawArc(CocoaDrawableID d, CocoaGCID gc,
                   int x, int y, unsigned int w, unsigned int h,
                   int angle1, int angle2);
void CocoaCopyArea(CocoaDrawableID src, CocoaDrawableID dst, CocoaGCID gc,
                    int srcX, int srcY, unsigned int w, unsigned int h,
                    int dstX, int dstY);
void CocoaPutImageRGBA(CocoaDrawableID d, CocoaGCID gc, int depth,
                        const unsigned char *rgbaData, int stride,
                        int imgW, int imgH,
                        int srcX, int srcY, int dstX, int dstY,
                        int w, int h, uint64_t bgPixel);

// ---- Pixmap management ----

CocoaPixmapID CocoaCreatePixmap(unsigned int width, unsigned int height,
                                 unsigned int depth);
void CocoaFreePixmap(CocoaPixmapID pm);
CocoaPixmapID CocoaCreateBitmapFromData(const unsigned char *bits,
                                         unsigned int width, unsigned int height);

// ---- Cursor management ----

CocoaCursorID CocoaCreateCursor(unsigned int shape);
void CocoaDefineCursor(CocoaWindowID w, CocoaCursorID cursor);
void CocoaSetCursorShape(CocoaWindowID w, unsigned int shape);
void CocoaUndefineCursor(CocoaWindowID w);
void CocoaFreeCursor(CocoaCursorID cursor);

// ---- Grab management ----

int CocoaGrabPointer(CocoaWindowID w);
void CocoaUngrabPointer(void);
int CocoaGrabKeyboard(CocoaWindowID w);
void CocoaUngrabKeyboard(void);

// ---- Selection / clipboard ----

void CocoaSetClipboardText(const char *text);
char *CocoaGetClipboardText(void);  // caller must free()
void CocoaFreeString(char *s);

// ---- Event system ----

// Event types matching platform constants.
#define COCOA_EVENT_KEY_PRESS         2
#define COCOA_EVENT_KEY_RELEASE       3
#define COCOA_EVENT_BUTTON_PRESS      4
#define COCOA_EVENT_BUTTON_RELEASE    5
#define COCOA_EVENT_MOTION            6
#define COCOA_EVENT_ENTER             7
#define COCOA_EVENT_LEAVE             8
#define COCOA_EVENT_FOCUS_IN          9
#define COCOA_EVENT_FOCUS_OUT         10
#define COCOA_EVENT_EXPOSE            12
#define COCOA_EVENT_DESTROY           17
#define COCOA_EVENT_UNMAP             18
#define COCOA_EVENT_MAP               19
#define COCOA_EVENT_CONFIGURE         22
#define COCOA_EVENT_PROPERTY          28
#define COCOA_EVENT_CLIENT_MESSAGE    33
#define COCOA_EVENT_VIRTUAL           0x100 // str holds the name

// CocoaRawEvent is the C-level event structure passed to Go.
typedef struct {
    int type;
    CocoaWindowID window;

    // Key event fields
    unsigned int state;      // modifier mask
    unsigned int keycode;
    uint64_t keysym;
    char str[64];

    // Button/motion fields
    int x, y;
    int rootX, rootY;
    unsigned int button;

    // Configure fields
    int width, height;

    // Expose fields
    int exposeX, exposeY;
    int exposeWidth, exposeHeight;
    int exposeCount;

    // Client message fields
    uint64_t messageType;
    int64_t messageData[5];

    // Focus fields
    int focusMode;
    int focusDetail;

    // Timestamp (milliseconds since boot)
    uint64_t time;
} CocoaRawEvent;

// CocoaNextEvent blocks until the next event is available and fills ev.
// Returns 1 if an event was returned, 0 if the app is stopping.
int CocoaNextEvent(CocoaRawEvent *ev);

// CocoaPending returns the number of events in the queue.
int CocoaPending(void);

// CocoaWakeEventReader posts an event of no type, which the event loop
// ignores, to return a CocoaNextEvent blocked on another thread.
void CocoaWakeEventReader(void);

// CocoaIsMainThread reports whether the caller is on the main thread.
int CocoaIsMainThread(void);

// CocoaStartEventPump starts processing NSEvents on the main thread,
// routing them to the internal event queue. Call from the main goroutine.
// This does not block — it runs a timer that pumps events.
void CocoaStartEventPump(void);

// CocoaPumpEvents processes all pending NSEvents synchronously.
// Must be called from the main thread. This is the main-thread event
// pump that feeds events into the internal queue.
void CocoaPumpEvents(void);

// ---- Window properties (emulated atom system) ----

uint64_t CocoaInternAtom(const char *name, bool onlyIfExists);
char *CocoaGetAtomName(uint64_t atom);

// ---- Font support (Core Text) ----

typedef uintptr_t CocoaFontID;

CocoaFontID CocoaOpenFont(const char *family, double size, int weight, int slant);
void CocoaCloseFont(CocoaFontID font);
int CocoaFontAscent(CocoaFontID font);
int CocoaFontDescent(CocoaFontID font);
int CocoaFontMaxWidth(CocoaFontID font);
bool CocoaFontIsFixed(CocoaFontID font);
int CocoaMeasureString(CocoaFontID font, const char *s, int len);
void CocoaDrawString(CocoaDrawableID d, CocoaFontID font,
                      int x, int y, const char *s, int len,
                      uint64_t pixel, uint16_t r, uint16_t g, uint16_t b);

// ---- WM helpers ----

void CocoaSetWMProtocols(CocoaWindowID w);
void CocoaSetWMHints(CocoaWindowID w, bool input, int initialState);
void CocoaSetWMNormalHints(CocoaWindowID w, int minW, int minH,
                            int maxW, int maxH);
void CocoaSetClassHint(CocoaWindowID w, const char *name, const char *cls);
void CocoaSetTransientFor(CocoaWindowID w, CocoaWindowID parent);
void CocoaIconifyWindow(CocoaWindowID w);
void CocoaWithdrawWindow(CocoaWindowID w);
void CocoaSetIconName(CocoaWindowID w, const char *name);

#endif // TAKIGO_COCOA_H
