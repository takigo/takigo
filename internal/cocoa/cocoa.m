// cocoa.m — Objective-C implementation of the Cocoa backend for takigo.
// This file implements all functions declared in cocoa.h using AppKit + CoreGraphics.

#import <Cocoa/Cocoa.h>
#import <CoreText/CoreText.h>
#import <QuartzCore/QuartzCore.h>
#import <objc/runtime.h>
#include "cocoa.h"
#include <stdlib.h>
#include <string.h>
#include <pthread.h>

// ============================================================================
// Forward declarations
// ============================================================================

@class TKWindow;
@class TKApplication;

// TKContentView needs to be declared early because helper functions reference it.
@interface TKContentView : NSView <NSTextInputClient>
@property (nonatomic) CocoaWindowID windowID;
@property (nonatomic) uint64_t bgPixel;
@property (nonatomic) int64_t eventMask;
@property (nonatomic) BOOL isTopLevel;
@property (nonatomic) BOOL isMapped;
@property (nonatomic) CGContextRef backingContext;
@property (nonatomic) void *backingData;
@property (nonatomic) unsigned int backingWidth;
@property (nonatomic) unsigned int backingHeight;
- (instancetype)initWithFrame:(NSRect)frame windowID:(CocoaWindowID)wid;
- (void)updateBackingStore;
@end

// ============================================================================
// Global state
// ============================================================================

static TKApplication *tkApp = nil;
static NSMutableDictionary<NSNumber *, TKContentView *> *windowRegistry = nil;
static uintptr_t nextWindowID = 1;
static uintptr_t nextGCID = 1;
static uintptr_t nextPixmapID = 1;
static uintptr_t nextAtomID = 100; // start above predefined atoms
static CocoaWindowID rootWindowID = 0; // virtual root window ID

// Atom emulation: bidirectional name<->ID mapping.
static NSMutableDictionary<NSString *, NSNumber *> *atomByName = nil;
static NSMutableDictionary<NSNumber *, NSString *> *atomByID = nil;

// Event queue: ring buffer of CocoaRawEvents posted from Cocoa callbacks.
#define EVENT_QUEUE_SIZE 1024
static CocoaRawEvent eventQueue[EVENT_QUEUE_SIZE];
static volatile int eventQueueHead = 0;
static volatile int eventQueueTail = 0;
static pthread_mutex_t eventQueueMutex = PTHREAD_MUTEX_INITIALIZER;
static pthread_cond_t eventQueueCond = PTHREAD_COND_INITIALIZER;

// Execute a block on the main thread. If already on the main thread, run directly.
// This avoids deadlocks since dispatch_async to the main queue requires
// the main run loop to be serviced, which doesn't happen in Go's event loop.
static inline void runOnMain(void (^block)(void)) {
    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }
}

static void postEvent(CocoaRawEvent *ev) {
    pthread_mutex_lock(&eventQueueMutex);
    int next = (eventQueueTail + 1) % EVENT_QUEUE_SIZE;
    if (next != eventQueueHead) {
        eventQueue[eventQueueTail] = *ev;
        eventQueueTail = next;
    }
    pthread_cond_signal(&eventQueueCond);
    pthread_mutex_unlock(&eventQueueMutex);
}

// ============================================================================
// GC emulation: lightweight drawing state
// ============================================================================

typedef struct {
    uint64_t foreground;
    uint64_t background;
    int lineWidth;
    int lineStyle;
    int capStyle;
    int joinStyle;
    int fillStyle;
    int function;
    int dashOffset;
    unsigned char dashList[32];
    int dashCount;
} CocoaGCState;

static NSMutableDictionary<NSNumber *, NSValue *> *gcRegistry = nil;

static CocoaGCState *lookupGC(CocoaGCID gcid) {
    NSValue *v = gcRegistry[@(gcid)];
    if (!v) return NULL;
    return (CocoaGCState *)[v pointerValue];
}

// ============================================================================
// Pixmap emulation: off-screen CGBitmapContext
// ============================================================================

typedef struct {
    CGContextRef ctx;
    unsigned int width;
    unsigned int height;
    unsigned int depth;
    void *data; // pixel buffer
} CocoaPixmap;

static NSMutableDictionary<NSNumber *, NSValue *> *pixmapRegistry = nil;

static CocoaPixmap *lookupPixmap(CocoaPixmapID pmid) {
    NSValue *v = pixmapRegistry[@(pmid)];
    if (!v) return NULL;
    return (CocoaPixmap *)[v pointerValue];
}

// ============================================================================
// Helper: pixel value → NSColor / CGColor
// ============================================================================

static void pixelToRGB(uint64_t pixel, CGFloat *r, CGFloat *g, CGFloat *b) {
    // Pixel format: 0x00RRGGBB
    *r = ((pixel >> 16) & 0xFF) / 255.0;
    *g = ((pixel >> 8) & 0xFF) / 255.0;
    *b = (pixel & 0xFF) / 255.0;
}

static NSColor *pixelToNSColor(uint64_t pixel) {
    CGFloat r, g, b;
    pixelToRGB(pixel, &r, &g, &b);
    return [NSColor colorWithCalibratedRed:r green:g blue:b alpha:1.0];
}

// ============================================================================
// Helper: get CGContext for a drawable (window view or pixmap)
// ============================================================================

static CGContextRef getDrawableContext(CocoaDrawableID d) {
    // Check pixmap registry first
    CocoaPixmap *pm = lookupPixmap((CocoaPixmapID)d);
    if (pm) return pm->ctx;

    // Otherwise it's a window — get the view's backing context
    TKContentView *view = windowRegistry[@(d)];
    if (view && view.backingContext) {
        return view.backingContext;
    }
    return NULL;
}

static void getDrawableSize(CocoaDrawableID d, unsigned int *w, unsigned int *h) {
    CocoaPixmap *pm = lookupPixmap((CocoaPixmapID)d);
    if (pm) {
        *w = pm->width;
        *h = pm->height;
        return;
    }
    TKContentView *view = windowRegistry[@(d)];
    if (view) {
        NSRect bounds = [view bounds];
        *w = (unsigned int)bounds.size.width;
        *h = (unsigned int)bounds.size.height;
        return;
    }
    *w = 0;
    *h = 0;
}

// ============================================================================
// Helper: apply GC state to a CGContext
// ============================================================================

static void applyGC(CGContextRef ctx, CocoaGCState *gc) {
    if (!ctx || !gc) return;
    CGFloat r, g, b;
    pixelToRGB(gc->foreground, &r, &g, &b);
    CGContextSetRGBFillColor(ctx, r, g, b, 1.0);
    CGContextSetRGBStrokeColor(ctx, r, g, b, 1.0);
    CGContextSetLineWidth(ctx, gc->lineWidth > 0 ? gc->lineWidth : 1.0);

    // Cap style
    CGLineCap cap = kCGLineCapButt;
    switch (gc->capStyle) {
        case 2: cap = kCGLineCapRound; break;
        case 3: cap = kCGLineCapSquare; break;
    }
    CGContextSetLineCap(ctx, cap);

    // Join style
    CGLineJoin join = kCGLineJoinMiter;
    switch (gc->joinStyle) {
        case 1: join = kCGLineJoinRound; break;
        case 2: join = kCGLineJoinBevel; break;
    }
    CGContextSetLineJoin(ctx, join);

    // Dash pattern
    if (gc->dashCount > 0) {
        CGFloat dashes[32];
        for (int i = 0; i < gc->dashCount; i++) {
            dashes[i] = (CGFloat)gc->dashList[i];
        }
        CGContextSetLineDash(ctx, gc->dashOffset, dashes, gc->dashCount);
    } else {
        CGContextSetLineDash(ctx, 0, NULL, 0);
    }
}

// ============================================================================
// TKContentView — NSView subclass for Tk content rendering
// ============================================================================

@implementation TKContentView

- (instancetype)initWithFrame:(NSRect)frame windowID:(CocoaWindowID)wid {
    self = [super initWithFrame:frame];
    if (self) {
        _windowID = wid;
        _bgPixel = 0x00D9D9D9; // Tk default background (light gray)
        _eventMask = 0;
        _isTopLevel = NO;
        _isMapped = NO;
        _backingContext = NULL;
        _backingData = NULL;
        _backingWidth = 0;
        _backingHeight = 0;
        self.wantsLayer = YES;
        self.layerContentsRedrawPolicy = NSViewLayerContentsRedrawOnSetNeedsDisplay;
        // Disable autoresizing — Tk manages all geometry explicitly.
        self.autoresizesSubviews = NO;
        self.autoresizingMask = 0;
        [self updateBackingStore];
    }
    return self;
}

- (void)dealloc {
    if (_backingContext) {
        CGContextRelease(_backingContext);
        _backingContext = NULL;
    }
    if (_backingData) {
        free(_backingData);
        _backingData = NULL;
    }
    [super dealloc];
}

- (void)updateBackingStore {
    NSRect bounds = [self bounds];
    unsigned int w = (unsigned int)bounds.size.width;
    unsigned int h = (unsigned int)bounds.size.height;
    if (w == 0) w = 1;
    if (h == 0) h = 1;
    if (w == _backingWidth && h == _backingHeight && _backingContext != NULL) return;

    if (_backingContext) {
        CGContextRelease(_backingContext);
        _backingContext = NULL;
    }
    if (_backingData) {
        free(_backingData);
        _backingData = NULL;
    }

    _backingWidth = w;
    _backingHeight = h;
    size_t bytesPerRow = w * 4;
    _backingData = calloc(h, bytesPerRow);
    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    _backingContext = CGBitmapContextCreate(_backingData, w, h, 8, bytesPerRow,
                                            cs, kCGImageAlphaPremultipliedFirst | kCGBitmapByteOrder32Host);
    CGColorSpaceRelease(cs);

    if (_backingContext) {
        // Flip coordinate system to match Tk (origin at top-left)
        CGContextTranslateCTM(_backingContext, 0, h);
        CGContextScaleCTM(_backingContext, 1.0, -1.0);

        // Fill with background color
        CGFloat r, g, b;
        pixelToRGB(_bgPixel, &r, &g, &b);
        CGContextSetRGBFillColor(_backingContext, r, g, b, 1.0);
        CGContextFillRect(_backingContext, CGRectMake(0, 0, w, h));
    }
}

- (BOOL)isFlipped {
    return YES; // origin at top-left, matching Tk
}

- (BOOL)acceptsFirstResponder {
    return YES;
}

- (BOOL)acceptsFirstMouse:(NSEvent *)event {
    return YES;
}

- (void)drawRect:(NSRect)dirtyRect {
    if (!_backingContext) return;
    CGImageRef img = CGBitmapContextCreateImage(_backingContext);
    if (!img) return;
    CGContextRef ctx = [[NSGraphicsContext currentContext] CGContext];
    if (ctx) {
        NSRect bounds = [self bounds];
        CGFloat h = bounds.size.height;
        // The backing context was drawn with a Y-flip CTM (translate+scale) so
        // Tk's top-left origin maps to CG's bottom-left. The resulting CGImage
        // has Tk-top at CG-top. Since isFlipped=YES, the view context also
        // has a Y-flip. CGContextDrawImage always maps image-bottom to rect
        // origin, so in our flipped view it would flip the image again.
        // Undo the view's flip locally so the image appears correctly.
        CGContextSaveGState(ctx);
        CGContextTranslateCTM(ctx, 0, h);
        CGContextScaleCTM(ctx, 1.0, -1.0);
        CGContextDrawImage(ctx, CGRectMake(0, 0, bounds.size.width, h), img);
        CGContextRestoreGState(ctx);
    }
    CGImageRelease(img);
}

- (void)setFrameSize:(NSSize)newSize {
    // Guard against no-op resizes to prevent configure event cascades.
    NSSize oldSize = self.frame.size;
    if (NSEqualSizes(oldSize, newSize)) return;

    [super setFrameSize:newSize];
    [self updateBackingStore];

    if (_isMapped) {
        CocoaRawEvent ev = {0};
        ev.type = COCOA_EVENT_CONFIGURE;
        ev.window = _windowID;
        ev.x = (int)self.frame.origin.x;
        ev.y = (int)self.frame.origin.y;
        ev.width = (int)newSize.width;
        ev.height = (int)newSize.height;
        ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
        postEvent(&ev);
    }
}

// ---- Mouse events ----

- (unsigned int)modifierFlags:(NSEvent *)event {
    NSEventModifierFlags flags = [event modifierFlags];
    unsigned int state = 0;
    if (flags & NSEventModifierFlagShift)   state |= (1 << 0); // ShiftMask
    if (flags & NSEventModifierFlagControl) state |= (1 << 2); // ControlMask
    if (flags & NSEventModifierFlagOption)  state |= (1 << 3); // Mod1Mask (Alt/Option)
    if (flags & NSEventModifierFlagCommand) state |= (1 << 4); // Mod2Mask (Command)
    if (flags & NSEventModifierFlagCapsLock) state |= (1 << 1); // LockMask
    return state;
}

- (void)postMouseEvent:(NSEvent *)event type:(int)type button:(unsigned int)btn {
    NSPoint loc = [self convertPoint:[event locationInWindow] fromView:nil];
    NSPoint screen = [[self window] convertPointToScreen:[event locationInWindow]];

    CocoaRawEvent ev = {0};
    ev.type = type;
    ev.window = _windowID;
    ev.x = (int)loc.x;
    ev.y = (int)loc.y;
    ev.rootX = (int)screen.x;
    ev.rootY = (int)(CocoaScreenHeight() - screen.y); // flip Y for root coords
    ev.state = [self modifierFlags:event];
    ev.button = btn;
    ev.time = (uint64_t)([event timestamp] * 1000);
    postEvent(&ev);
}

- (void)mouseDown:(NSEvent *)event    { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_PRESS button:1]; }
- (void)mouseUp:(NSEvent *)event      { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_RELEASE button:1]; }
- (void)rightMouseDown:(NSEvent *)event { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_PRESS button:3]; }
- (void)rightMouseUp:(NSEvent *)event   { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_RELEASE button:3]; }
- (void)otherMouseDown:(NSEvent *)event  { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_PRESS button:2]; }
- (void)otherMouseUp:(NSEvent *)event    { [self postMouseEvent:event type:COCOA_EVENT_BUTTON_RELEASE button:2]; }

- (void)mouseMoved:(NSEvent *)event    { [self postMouseEvent:event type:COCOA_EVENT_MOTION button:0]; }
- (void)mouseDragged:(NSEvent *)event  { [self postMouseEvent:event type:COCOA_EVENT_MOTION button:0]; }
- (void)rightMouseDragged:(NSEvent *)event { [self postMouseEvent:event type:COCOA_EVENT_MOTION button:0]; }
- (void)otherMouseDragged:(NSEvent *)event { [self postMouseEvent:event type:COCOA_EVENT_MOTION button:0]; }

- (void)scrollWheel:(NSEvent *)event {
    // Map scroll wheel to button 4/5 (up/down) matching X11 convention
    if ([event deltaY] > 0) {
        [self postMouseEvent:event type:COCOA_EVENT_BUTTON_PRESS button:4];
        [self postMouseEvent:event type:COCOA_EVENT_BUTTON_RELEASE button:4];
    } else if ([event deltaY] < 0) {
        [self postMouseEvent:event type:COCOA_EVENT_BUTTON_PRESS button:5];
        [self postMouseEvent:event type:COCOA_EVENT_BUTTON_RELEASE button:5];
    }
}

- (void)mouseEntered:(NSEvent *)event  { [self postMouseEvent:event type:COCOA_EVENT_ENTER button:0]; }
- (void)mouseExited:(NSEvent *)event   { [self postMouseEvent:event type:COCOA_EVENT_LEAVE button:0]; }

- (void)updateTrackingAreas {
    [super updateTrackingAreas];
    for (NSTrackingArea *area in [self trackingAreas]) {
        [self removeTrackingArea:area];
    }
    NSTrackingArea *ta = [[NSTrackingArea alloc]
        initWithRect:[self bounds]
             options:(NSTrackingMouseEnteredAndExited | NSTrackingMouseMoved |
                      NSTrackingActiveAlways | NSTrackingInVisibleRect)
               owner:self
            userInfo:nil];
    [self addTrackingArea:ta];
}

// ---- Key events ----

- (void)postKeyEvent:(NSEvent *)event type:(int)type {
    CocoaRawEvent ev = {0};
    ev.type = type;
    ev.window = _windowID;
    ev.state = [self modifierFlags:event];
    ev.keycode = [event keyCode];
    ev.time = (uint64_t)([event timestamp] * 1000);

    // Get characters
    NSString *chars = [event characters];
    if (chars && [chars length] > 0) {
        unichar ch = [chars characterAtIndex:0];
        ev.keysym = ch;
        const char *utf8 = [chars UTF8String];
        if (utf8) {
            strncpy(ev.str, utf8, sizeof(ev.str) - 1);
        }
    } else {
        // Use charactersIgnoringModifiers for keysym
        NSString *plain = [event charactersIgnoringModifiers];
        if (plain && [plain length] > 0) {
            ev.keysym = [plain characterAtIndex:0];
        }
    }

    // Map special keys to X11 keysyms
    switch ([event keyCode]) {
        case 36: ev.keysym = 0xff0d; break; // Return
        case 48: ev.keysym = 0xff09; break; // Tab
        case 51: ev.keysym = 0xff08; break; // Backspace/Delete
        case 53: ev.keysym = 0xff1b; break; // Escape
        case 117: ev.keysym = 0xffff; break; // Forward Delete
        case 115: ev.keysym = 0xff50; break; // Home
        case 119: ev.keysym = 0xff57; break; // End
        case 116: ev.keysym = 0xff55; break; // PageUp
        case 121: ev.keysym = 0xff56; break; // PageDown
        case 123: ev.keysym = 0xff51; break; // Left arrow
        case 124: ev.keysym = 0xff53; break; // Right arrow
        case 125: ev.keysym = 0xff54; break; // Down arrow
        case 126: ev.keysym = 0xff52; break; // Up arrow
        case 114: ev.keysym = 0xff63; break; // Insert (Help key)
    }

    postEvent(&ev);
}

- (void)keyDown:(NSEvent *)event {
    [self postKeyEvent:event type:COCOA_EVENT_KEY_PRESS];
    // Also feed to input method for composition support
    [self interpretKeyEvents:@[event]];
}

- (void)keyUp:(NSEvent *)event {
    [self postKeyEvent:event type:COCOA_EVENT_KEY_RELEASE];
}

- (void)flagsChanged:(NSEvent *)event {
    // Post as key event with modifier state change
    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_KEY_PRESS;
    ev.window = _windowID;
    ev.state = [self modifierFlags:event];
    ev.keycode = [event keyCode];
    ev.time = (uint64_t)([event timestamp] * 1000);
    postEvent(&ev);
}

// ---- NSTextInputClient protocol (required for IME support) ----

- (void)insertText:(id)string replacementRange:(NSRange)replacementRange {
    // Already handled via keyDown → postKeyEvent
}

- (void)doCommandBySelector:(SEL)selector {
    // Ignore — we handle key events directly
}

- (void)setMarkedText:(id)string selectedRange:(NSRange)selectedRange
     replacementRange:(NSRange)replacementRange {
    // TODO: IME composition display
}

- (void)unmarkText {
}

- (NSRange)selectedRange {
    return NSMakeRange(NSNotFound, 0);
}

- (NSRange)markedRange {
    return NSMakeRange(NSNotFound, 0);
}

- (BOOL)hasMarkedText {
    return NO;
}

- (NSAttributedString *)attributedSubstringForProposedRange:(NSRange)range
                                                actualRange:(NSRangePointer)actualRange {
    return nil;
}

- (NSArray<NSAttributedStringKey> *)validAttributesForMarkedText {
    return @[];
}

- (NSRect)firstRectForCharacterRange:(NSRange)range actualRange:(NSRangePointer)actualRange {
    return NSZeroRect;
}

- (NSUInteger)characterIndexForPoint:(NSPoint)point {
    return NSNotFound;
}

@end

// ============================================================================
// TKWindow — NSWindow subclass
// ============================================================================

@interface TKWindow : NSWindow
@property (nonatomic) CocoaWindowID windowID;
@end

@implementation TKWindow

- (BOOL)canBecomeKeyWindow {
    return YES;
}

- (BOOL)canBecomeMainWindow {
    return YES;
}

@end

// ============================================================================
// TKApplication — NSApplication subclass
// ============================================================================

@interface TKApplication : NSApplication <NSApplicationDelegate>
@end

@implementation TKApplication

- (void)applicationDidFinishLaunching:(NSNotification *)notification {
    // Activate the app so it appears in front
    [self activateIgnoringOtherApps:YES];
}

- (BOOL)applicationShouldTerminateAfterLastWindowClosed:(NSApplication *)sender {
    return NO; // Let Go code decide when to quit
}

- (void)applicationDidBecomeActive:(NSNotification *)notification {
    // Post focus events for key window
}

@end

// ============================================================================
// TKWindowDelegate — handles window lifecycle events
// ============================================================================

@interface TKWindowDelegate : NSObject <NSWindowDelegate>
@property (nonatomic) CocoaWindowID windowID;
@end

@implementation TKWindowDelegate

- (void)windowDidBecomeKey:(NSNotification *)notification {
    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_FOCUS_IN;
    ev.window = _windowID;
    ev.focusMode = 0; // NotifyNormal
    ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
    postEvent(&ev);
}

- (void)windowDidResignKey:(NSNotification *)notification {
    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_FOCUS_OUT;
    ev.window = _windowID;
    ev.focusMode = 0;
    ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
    postEvent(&ev);
}

- (void)windowDidResize:(NSNotification *)notification {
    // Handled by TKContentView setFrameSize
}

- (void)windowDidMove:(NSNotification *)notification {
    TKWindow *window = (TKWindow *)[notification object];
    TKContentView *view = windowRegistry[@(_windowID)];
    // Only post configure events for mapped windows. The window may receive
    // move notifications during initial creation (before layout completes),
    // and posting those would inject stale sizes into the event queue.
    if (!view || !view.isMapped) return;

    NSRect frame = [window frame];
    NSRect screen = [[window screen] frame];
    // Report content dimensions, not frame dimensions.
    // Frame includes title bar; Tk expects content-area size.
    NSRect content = [window contentRectForFrameRect:frame];

    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_CONFIGURE;
    ev.window = _windowID;
    ev.x = (int)content.origin.x;
    ev.y = (int)(screen.size.height - content.origin.y - content.size.height);
    ev.width = (int)content.size.width;
    ev.height = (int)content.size.height;
    ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
    postEvent(&ev);
}

- (BOOL)windowShouldClose:(NSWindow *)sender {
    // Post WM_DELETE_WINDOW equivalent as client message
    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_CLIENT_MESSAGE;
    ev.window = _windowID;
    // Use atom ID for WM_DELETE_WINDOW (will be set up during init)
    ev.messageType = CocoaInternAtom("WM_DELETE_WINDOW", false);
    ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
    postEvent(&ev);
    return NO; // Don't close; let Go code handle it
}

- (void)windowWillClose:(NSNotification *)notification {
    CocoaRawEvent ev = {0};
    ev.type = COCOA_EVENT_DESTROY;
    ev.window = _windowID;
    ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
    postEvent(&ev);
}

@end

// ============================================================================
// Application lifecycle
// ============================================================================

void CocoaInit(void) {
    @autoreleasepool {
        tkApp = (TKApplication *)[TKApplication sharedApplication];
        tkApp.delegate = tkApp;

        windowRegistry = [NSMutableDictionary new];
        gcRegistry = [NSMutableDictionary new];
        pixmapRegistry = [NSMutableDictionary new];
        atomByName = [NSMutableDictionary new];
        atomByID = [NSMutableDictionary new];

        // Set activation policy to regular app (shows in dock)
        [NSApp setActivationPolicy:NSApplicationActivationPolicyRegular];

        // Create default menu bar
        NSMenu *menuBar = [NSMenu new];
        NSMenuItem *appMenuItem = [NSMenuItem new];
        [menuBar addItem:appMenuItem];
        [NSApp setMainMenu:menuBar];

        NSMenu *appMenu = [NSMenu new];
        [appMenu addItemWithTitle:@"Quit"
                           action:@selector(terminate:)
                    keyEquivalent:@"q"];
        [appMenuItem setSubmenu:appMenu];

        // Pre-register standard atoms
        CocoaInternAtom("WM_DELETE_WINDOW", false);
        CocoaInternAtom("WM_PROTOCOLS", false);
        CocoaInternAtom("WM_NAME", false);
        CocoaInternAtom("STRING", false);
        CocoaInternAtom("WM_NORMAL_HINTS", false);
        CocoaInternAtom("PRIMARY", false);
        CocoaInternAtom("SECONDARY", false);
        CocoaInternAtom("ATOM", false);
        CocoaInternAtom("CARDINAL", false);
        CocoaInternAtom("WINDOW", false);

        // Finish launching so the app is properly activated.
        // We don't call [NSApp run] because Go manages its own event loop;
        // instead finishLaunching sets up the app state and
        // activateIgnoringOtherApps brings us to the foreground.
        [NSApp finishLaunching];
        [NSApp activateIgnoringOtherApps:YES];
    }
}

void CocoaRun(void) {
    @autoreleasepool {
        [NSApp run];
    }
}

void CocoaStop(void) {
    runOnMain(^{
        [NSApp stop:nil];
        // Post a dummy event to wake the run loop so it processes the stop
        NSEvent *event = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                            location:NSMakePoint(0, 0)
                                       modifierFlags:0
                                           timestamp:0
                                        windowNumber:0
                                             context:nil
                                             subtype:0
                                               data1:0
                                               data2:0];
        [NSApp postEvent:event atStart:YES];
    });
}

void CocoaFlush(void) {
    // Trigger display updates on all dirty views
    runOnMain(^{
        for (NSNumber *key in windowRegistry) {
            TKContentView *view = windowRegistry[key];
            if (view.needsDisplay) {
                [view displayIfNeeded];
            }
        }
    });
}

// ============================================================================
// Screen info
// ============================================================================

int CocoaScreenWidth(void) {
    NSScreen *screen = [NSScreen mainScreen];
    return (int)[screen frame].size.width;
}

int CocoaScreenHeight(void) {
    NSScreen *screen = [NSScreen mainScreen];
    return (int)[screen frame].size.height;
}

int CocoaScreenWidthMM(void) {
    // NSScreen doesn't directly provide mm — compute from DPI
    NSScreen *screen = [NSScreen mainScreen];
    CGDirectDisplayID displayID = [[[screen deviceDescription] objectForKey:@"NSScreenNumber"] unsignedIntValue];
    CGSize sizeInMM = CGDisplayScreenSize(displayID);
    return (int)sizeInMM.width;
}

int CocoaScreenHeightMM(void) {
    NSScreen *screen = [NSScreen mainScreen];
    CGDirectDisplayID displayID = [[[screen deviceDescription] objectForKey:@"NSScreenNumber"] unsignedIntValue];
    CGSize sizeInMM = CGDisplayScreenSize(displayID);
    return (int)sizeInMM.height;
}

int CocoaScreenDepth(void) {
    return 24; // macOS uses 24-bit color (32-bit with alpha)
}

double CocoaScreenBackingScale(void) {
    NSScreen *screen = [NSScreen mainScreen];
    return [screen backingScaleFactor];
}

// ============================================================================
// Window management
// ============================================================================

CocoaWindowID CocoaCreateWindow(CocoaWindowID parent, int x, int y,
                                 unsigned int width, unsigned int height,
                                 unsigned int borderWidth, uint64_t bgPixel,
                                 int64_t eventMask, bool overrideRedirect) {
    __block CocoaWindowID result = 0;

    void (^block)(void) = ^{
        @autoreleasepool {
            CocoaWindowID wid = nextWindowID++;

            unsigned int w = width > 0 ? width : 1;
            unsigned int h = height > 0 ? height : 1;

            // Determine if this is a top-level window.
            // In Tk, the root window is like X11's screen root. Children of the
            // root are toplevel windows and need their own NSWindow on macOS.
            bool isTopLevel = (parent == 0 || parent == rootWindowID);

            if (parent == 0 && rootWindowID == 0) {
                // This is the virtual root window — don't create a real NSWindow.
                // Just register a placeholder view.
                TKContentView *view = [[TKContentView alloc]
                    initWithFrame:NSMakeRect(0, 0, 1, 1) windowID:wid];
                view.bgPixel = bgPixel;
                view.eventMask = eventMask;
                view.isTopLevel = NO;
                view.isMapped = YES; // root is always "mapped"
                windowRegistry[@(wid)] = view;
                rootWindowID = wid;
                result = wid;
                return;
            }

            TKContentView *view = [[TKContentView alloc]
                initWithFrame:NSMakeRect(0, 0, w, h) windowID:wid];
            view.bgPixel = bgPixel;
            view.eventMask = eventMask;

            if (isTopLevel) {
                // Top-level window — create NSWindow
                NSWindowStyleMask style;
                if (overrideRedirect) {
                    style = NSWindowStyleMaskBorderless;
                } else {
                    style = NSWindowStyleMaskTitled | NSWindowStyleMaskClosable |
                            NSWindowStyleMaskMiniaturizable | NSWindowStyleMaskResizable;
                }

                // Convert Y coordinate: Tk uses top-left origin, Cocoa uses bottom-left
                int screenH = CocoaScreenHeight();
                int cocoaY = screenH - y - (int)h;

                TKWindow *window = [[TKWindow alloc]
                    initWithContentRect:NSMakeRect(x, cocoaY, w, h)
                              styleMask:style
                                backing:NSBackingStoreBuffered
                                  defer:YES];
                window.windowID = wid;
                [window setContentView:view];
                [window setAcceptsMouseMovedEvents:YES];

                TKWindowDelegate *delegate = [TKWindowDelegate new];
                delegate.windowID = wid;
                [window setDelegate:delegate];
                // Keep delegate alive (retained by window)
                objc_setAssociatedObject(window, "delegate", delegate, OBJC_ASSOCIATION_RETAIN_NONATOMIC);

                view.isTopLevel = YES;
            } else {
                // Child view — add as subview of parent
                TKContentView *parentView = windowRegistry[@(parent)];
                if (parentView) {
                    view.frame = NSMakeRect(x, y, w, h);
                    [parentView addSubview:view];
                }
            }

            windowRegistry[@(wid)] = view;
            result = wid;
        }
    };

    if ([NSThread isMainThread]) {
        block();
    } else {
        dispatch_sync(dispatch_get_main_queue(), block);
    }

    return result;
}

CocoaWindowID CocoaCreateSimpleWindow(CocoaWindowID parent, int x, int y,
                                       unsigned int width, unsigned int height,
                                       unsigned int borderWidth,
                                       uint64_t border, uint64_t background) {
    return CocoaCreateWindow(parent, x, y, width, height, borderWidth,
                              background, 0, false);
}

void CocoaDestroyWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;

            if (view.isTopLevel) {
                [[view window] close];
            } else {
                [view removeFromSuperview];
            }
            [windowRegistry removeObjectForKey:@(w)];
        }
    });
}

void CocoaMapWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            view.isMapped = YES;

            if (view.isTopLevel) {
                [[view window] makeKeyAndOrderFront:nil];
            } else {
                view.hidden = NO;
            }

            // Post configure event with current size so that geometry
            // managers (grid) re-layout now that the view is mapped.
            // The view may have been resized while unmapped, and those
            // configure events were suppressed.
            NSRect bounds = [view bounds];
            uint64_t now = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);

            CocoaRawEvent cfgEv = {0};
            cfgEv.type = COCOA_EVENT_CONFIGURE;
            cfgEv.window = w;
            cfgEv.x = (int)view.frame.origin.x;
            cfgEv.y = (int)view.frame.origin.y;
            cfgEv.width = (int)bounds.size.width;
            cfgEv.height = (int)bounds.size.height;
            cfgEv.time = now;
            postEvent(&cfgEv);

            // Post map event
            CocoaRawEvent ev = {0};
            ev.type = COCOA_EVENT_MAP;
            ev.window = w;
            ev.time = now;
            postEvent(&ev);

            // Post initial expose event
            CocoaRawEvent exposeEv = {0};
            exposeEv.type = COCOA_EVENT_EXPOSE;
            exposeEv.window = w;
            exposeEv.exposeWidth = (int)bounds.size.width;
            exposeEv.exposeHeight = (int)bounds.size.height;
            exposeEv.time = now;
            postEvent(&exposeEv);
        }
    });
}

void CocoaMapRaised(CocoaWindowID w) {
    CocoaMapWindow(w);
    CocoaRaiseWindow(w);
}

void CocoaUnmapWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            view.isMapped = NO;

            if (view.isTopLevel) {
                [[view window] orderOut:nil];
            } else {
                view.hidden = YES;
            }

            CocoaRawEvent ev = {0};
            ev.type = COCOA_EVENT_UNMAP;
            ev.window = w;
            ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
            postEvent(&ev);
        }
    });
}

void CocoaRaiseWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            if (view.isTopLevel) {
                [[view window] makeKeyAndOrderFront:nil];
            } else {
                NSView *superview = [view superview];
                if (superview) {
                    [view removeFromSuperview];
                    [superview addSubview:view];
                }
            }
        }
    });
}

void CocoaLowerWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            if (view.isTopLevel) {
                [[view window] orderBack:nil];
            }
        }
    });
}

void CocoaMoveWindow(CocoaWindowID w, int x, int y) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            if (view.isTopLevel) {
                NSRect frame = [[view window] frame];
                int screenH = CocoaScreenHeight();
                int cocoaY = screenH - y - (int)frame.size.height;
                [[view window] setFrameOrigin:NSMakePoint(x, cocoaY)];
            } else {
                NSRect frame = view.frame;
                frame.origin = NSMakePoint(x, y);
                view.frame = frame;
            }
        }
    });
}

void CocoaResizeWindow(CocoaWindowID w, unsigned int width, unsigned int height) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            if (view.isTopLevel) {
                NSRect frame = [[view window] frame];
                NSRect content = [[view window] contentRectForFrameRect:frame];
                content.size = NSMakeSize(width, height);
                NSRect newFrame = [[view window] frameRectForContentRect:content];
                // Keep top-left corner fixed
                newFrame.origin.y = frame.origin.y + frame.size.height - newFrame.size.height;
                [[view window] setFrame:newFrame display:YES];
            } else {
                NSRect frame = view.frame;
                frame.size = NSMakeSize(width, height);
                view.frame = frame;
            }
        }
    });
}

void CocoaMoveResizeWindow(CocoaWindowID w, int x, int y,
                            unsigned int width, unsigned int height) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view) return;
            if (view.isTopLevel) {
                int screenH = CocoaScreenHeight();
                int cocoaY = screenH - y - (int)height;
                NSRect content = NSMakeRect(x, cocoaY, width, height);
                NSRect frame = [[view window] frameRectForContentRect:content];
                [[view window] setFrame:frame display:YES];
            } else {
                view.frame = NSMakeRect(x, y, width, height);
            }
        }
    });
}

void CocoaSelectInput(CocoaWindowID w, int64_t eventMask) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view) {
            view.eventMask = eventMask;
        }
    });
}

void CocoaStoreName(CocoaWindowID w, const char *name) {
    NSString *title = [NSString stringWithUTF8String:name];
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view && view.isTopLevel) {
            [[view window] setTitle:title];
        }
    });
}

void CocoaTranslateCoordinates(CocoaWindowID src, CocoaWindowID dst,
                                int srcX, int srcY, int *dstX, int *dstY) {
    __block int rx = srcX;
    __block int ry = srcY;
    runOnMain(^{
        @autoreleasepool {
            TKContentView *srcView = windowRegistry[@(src)];
            TKContentView *dstView = windowRegistry[@(dst)];
            if (!srcView || !dstView) return;

            NSPoint p = NSMakePoint(srcX, srcY);
            // Convert through window coordinates
            NSPoint screenP = [srcView convertPoint:p toView:nil];
            if ([srcView window]) {
                screenP = [[srcView window] convertPointToScreen:screenP];
            }
            if ([dstView window]) {
                screenP = [[dstView window] convertPointFromScreen:screenP];
            }
            NSPoint localP = [dstView convertPoint:screenP fromView:nil];
            rx = (int)localP.x;
            ry = (int)localP.y;
        }
    });
    *dstX = rx;
    *dstY = ry;
}

void CocoaSetWindowBackground(CocoaWindowID w, uint64_t pixel) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view) {
            view.bgPixel = pixel;
            if (view.isTopLevel) {
                [[view window] setBackgroundColor:pixelToNSColor(pixel)];
            }
        }
    });
}

void CocoaClearWindow(CocoaWindowID w) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view || !view.backingContext) return;

            NSRect bounds = [view bounds];
            CGFloat r, g, b;
            pixelToRGB(view.bgPixel, &r, &g, &b);
            CGContextSetRGBFillColor(view.backingContext, r, g, b, 1.0);
            CGContextFillRect(view.backingContext, CGRectMake(0, 0,
                              bounds.size.width, bounds.size.height));
            [view setNeedsDisplay:YES];
        }
    });
}

void CocoaClearArea(CocoaWindowID w, int x, int y,
                     unsigned int width, unsigned int height, bool exposures) {
    runOnMain(^{
        @autoreleasepool {
            TKContentView *view = windowRegistry[@(w)];
            if (!view || !view.backingContext) return;

            CGFloat r, g, b;
            pixelToRGB(view.bgPixel, &r, &g, &b);
            CGContextSetRGBFillColor(view.backingContext, r, g, b, 1.0);
            CGContextFillRect(view.backingContext, CGRectMake(x, y, width, height));
            [view setNeedsDisplayInRect:NSMakeRect(x, y, width, height)];

            if (exposures) {
                CocoaRawEvent ev = {0};
                ev.type = COCOA_EVENT_EXPOSE;
                ev.window = (CocoaWindowID)w;
                ev.exposeX = x;
                ev.exposeY = y;
                ev.exposeWidth = width;
                ev.exposeHeight = height;
                ev.time = (uint64_t)([NSProcessInfo processInfo].systemUptime * 1000);
                postEvent(&ev);
            }
        }
    });
}

void CocoaSetInputFocus(CocoaWindowID w) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view && [view window]) {
            [[view window] makeFirstResponder:view];
        }
    });
}

// ============================================================================
// Graphics context emulation
// ============================================================================

CocoaGCID CocoaCreateGC(uint64_t fg, uint64_t bg, int lineWidth, int function) {
    CocoaGCState *gc = (CocoaGCState *)calloc(1, sizeof(CocoaGCState));
    gc->foreground = fg;
    gc->background = bg;
    gc->lineWidth = lineWidth;
    gc->function = function;
    gc->capStyle = 1; // CapButt
    gc->joinStyle = 0; // JoinMiter

    CocoaGCID gcid = nextGCID++;
    gcRegistry[@(gcid)] = [NSValue valueWithPointer:gc];
    return gcid;
}

void CocoaFreeGC(CocoaGCID gcid) {
    NSValue *v = gcRegistry[@(gcid)];
    if (v) {
        free([v pointerValue]);
        [gcRegistry removeObjectForKey:@(gcid)];
    }
}

void CocoaSetForeground(CocoaGCID gcid, uint64_t pixel) {
    CocoaGCState *gc = lookupGC(gcid);
    if (gc) gc->foreground = pixel;
}

void CocoaSetBackground(CocoaGCID gcid, uint64_t pixel) {
    CocoaGCState *gc = lookupGC(gcid);
    if (gc) gc->background = pixel;
}

void CocoaSetLineAttributes(CocoaGCID gcid, unsigned int lineWidth,
                             int lineStyle, int capStyle, int joinStyle) {
    CocoaGCState *gc = lookupGC(gcid);
    if (!gc) return;
    gc->lineWidth = lineWidth;
    gc->lineStyle = lineStyle;
    gc->capStyle = capStyle;
    gc->joinStyle = joinStyle;
}

void CocoaSetFillStyle(CocoaGCID gcid, int fillStyle) {
    CocoaGCState *gc = lookupGC(gcid);
    if (gc) gc->fillStyle = fillStyle;
}

void CocoaSetDashes(CocoaGCID gcid, int dashOffset, const unsigned char *dashList, int n) {
    CocoaGCState *gc = lookupGC(gcid);
    if (!gc) return;
    gc->dashOffset = dashOffset;
    gc->dashCount = n < 32 ? n : 32;
    memcpy(gc->dashList, dashList, gc->dashCount);
}

// ============================================================================
// Drawing primitives
// ============================================================================

void CocoaFillRectangle(CocoaDrawableID d, CocoaGCID gcid,
                         int x, int y, unsigned int w, unsigned int h) {
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);
    CGContextFillRect(ctx, CGRectMake(x, y, w, h));
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(x, y, w, h)];
}

void CocoaDrawRectangle(CocoaDrawableID d, CocoaGCID gcid,
                         int x, int y, unsigned int w, unsigned int h) {
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);
    CGContextStrokeRect(ctx, CGRectMake(x + 0.5, y + 0.5, w, h));
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(x, y, w + 1, h + 1)];
}

void CocoaDrawLine(CocoaDrawableID d, CocoaGCID gcid,
                    int x1, int y1, int x2, int y2) {
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);
    CGContextMoveToPoint(ctx, x1 + 0.5, y1 + 0.5);
    CGContextAddLineToPoint(ctx, x2 + 0.5, y2 + 0.5);
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) {
        int minX = x1 < x2 ? x1 : x2;
        int minY = y1 < y2 ? y1 : y2;
        int maxX = x1 > x2 ? x1 : x2;
        int maxY = y1 > y2 ? y1 : y2;
        [view setNeedsDisplayInRect:NSMakeRect(minX, minY, maxX - minX + 2, maxY - minY + 2)];
    }
}

void CocoaDrawLines(CocoaDrawableID d, CocoaGCID gcid,
                     const int16_t *points, int npoints, int mode) {
    if (npoints < 2) return;
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);

    int cx = points[0], cy = points[1];
    CGContextMoveToPoint(ctx, cx + 0.5, cy + 0.5);
    for (int i = 1; i < npoints; i++) {
        if (mode == 1) { // CoordModePrevious
            cx += points[i * 2];
            cy += points[i * 2 + 1];
        } else {
            cx = points[i * 2];
            cy = points[i * 2 + 1];
        }
        CGContextAddLineToPoint(ctx, cx + 0.5, cy + 0.5);
    }
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplay:YES];
}

void CocoaFillPolygon(CocoaDrawableID d, CocoaGCID gcid,
                       const int16_t *points, int npoints, int shape, int mode) {
    if (npoints < 3) return;
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);

    int cx = points[0], cy = points[1];
    CGContextMoveToPoint(ctx, cx, cy);
    for (int i = 1; i < npoints; i++) {
        if (mode == 1) {
            cx += points[i * 2];
            cy += points[i * 2 + 1];
        } else {
            cx = points[i * 2];
            cy = points[i * 2 + 1];
        }
        CGContextAddLineToPoint(ctx, cx, cy);
    }
    CGContextClosePath(ctx);
    CGContextFillPath(ctx);
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplay:YES];
}

void CocoaFillArc(CocoaDrawableID d, CocoaGCID gcid,
                   int x, int y, unsigned int w, unsigned int h,
                   int angle1, int angle2) {
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);

    // X11 angles are in 64ths of a degree; convert to radians
    // X11: counterclockwise from 3 o'clock; CG: counterclockwise from 3 o'clock (but Y is flipped)
    double startAngle = -(angle1 / 64.0) * M_PI / 180.0;
    double endAngle = -((angle1 + angle2) / 64.0) * M_PI / 180.0;
    int clockwise = (angle2 > 0) ? 1 : 0;

    CGFloat cx = x + w / 2.0;
    CGFloat cy = y + h / 2.0;
    CGFloat rx = w / 2.0;
    CGFloat ry = h / 2.0;

    CGContextTranslateCTM(ctx, cx, cy);
    CGContextScaleCTM(ctx, 1.0, ry / rx);
    CGContextMoveToPoint(ctx, 0, 0);
    CGContextAddArc(ctx, 0, 0, rx, startAngle, endAngle, clockwise);
    CGContextClosePath(ctx);
    CGContextFillPath(ctx);
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(x, y, w, h)];
}

void CocoaDrawArc(CocoaDrawableID d, CocoaGCID gcid,
                   int x, int y, unsigned int w, unsigned int h,
                   int angle1, int angle2) {
    CGContextRef ctx = getDrawableContext(d);
    CocoaGCState *gc = lookupGC(gcid);
    if (!ctx || !gc) return;

    CGContextSaveGState(ctx);
    applyGC(ctx, gc);

    double startAngle = -(angle1 / 64.0) * M_PI / 180.0;
    double endAngle = -((angle1 + angle2) / 64.0) * M_PI / 180.0;
    int clockwise = (angle2 > 0) ? 1 : 0;

    CGFloat cx = x + w / 2.0;
    CGFloat cy = y + h / 2.0;
    CGFloat rx = w / 2.0;
    CGFloat ry = h / 2.0;

    CGContextTranslateCTM(ctx, cx, cy);
    CGContextScaleCTM(ctx, 1.0, ry / rx);
    CGContextAddArc(ctx, 0, 0, rx, startAngle, endAngle, clockwise);
    CGContextStrokePath(ctx);
    CGContextRestoreGState(ctx);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(x, y, w + 1, h + 1)];
}

void CocoaCopyArea(CocoaDrawableID src, CocoaDrawableID dst, CocoaGCID gcid,
                    int srcX, int srcY, unsigned int w, unsigned int h,
                    int dstX, int dstY) {
    CGContextRef srcCtx = getDrawableContext(src);
    CGContextRef dstCtx = getDrawableContext(dst);
    if (!srcCtx || !dstCtx) return;

    // Create an image from the source region
    CGImageRef fullImg = CGBitmapContextCreateImage(srcCtx);
    if (!fullImg) return;

    unsigned int srcH = (unsigned int)CGBitmapContextGetHeight(srcCtx);
    // Note: the backing context has a flipped transform, but CGBitmapContextCreateImage
    // returns the raw bitmap. We need to account for the flip.
    CGRect cropRect = CGRectMake(srcX, srcH - srcY - h, w, h);
    CGImageRef croppedImg = CGImageCreateWithImageInRect(fullImg, cropRect);
    CGImageRelease(fullImg);
    if (!croppedImg) return;

    // Draw into destination. The dst context is also flipped, so we need to
    // save/restore and temporarily undo the flip for image drawing.
    CGContextSaveGState(dstCtx);
    // Undo the flip transform for image drawing
    unsigned int dstH = (unsigned int)CGBitmapContextGetHeight(dstCtx);
    CGContextTranslateCTM(dstCtx, 0, dstH);
    CGContextScaleCTM(dstCtx, 1.0, -1.0);
    // Now draw in unflipped coords
    CGContextDrawImage(dstCtx, CGRectMake(dstX, dstH - dstY - h, w, h), croppedImg);
    CGContextRestoreGState(dstCtx);

    CGImageRelease(croppedImg);

    TKContentView *view = windowRegistry[@(dst)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(dstX, dstY, w, h)];
}

void CocoaPutImageRGBA(CocoaDrawableID d, CocoaGCID gcid, int depth,
                        const unsigned char *rgbaData, int stride,
                        int imgW, int imgH,
                        int srcX, int srcY, int dstX, int dstY,
                        int w, int h, uint64_t bgPixel) {
    CGContextRef ctx = getDrawableContext(d);
    if (!ctx) return;

    // Create CGImage from RGBA data
    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    CGDataProviderRef provider = CGDataProviderCreateWithData(NULL, rgbaData,
                                                              imgH * stride, NULL);
    CGImageRef img = CGImageCreate(imgW, imgH, 8, 32, stride, cs,
                                    kCGImageAlphaPremultipliedLast | kCGBitmapByteOrderDefault,
                                    provider, NULL, false, kCGRenderingIntentDefault);
    CGDataProviderRelease(provider);
    CGColorSpaceRelease(cs);
    if (!img) return;

    // Crop to source region if needed
    CGImageRef srcImg = img;
    if (srcX != 0 || srcY != 0 || w != imgW || h != imgH) {
        srcImg = CGImageCreateWithImageInRect(img, CGRectMake(srcX, srcY, w, h));
        CGImageRelease(img);
        if (!srcImg) return;
    }

    // Draw into context (accounting for flip)
    CGContextSaveGState(ctx);
    unsigned int ctxH = (unsigned int)CGBitmapContextGetHeight(ctx);
    CGContextTranslateCTM(ctx, 0, ctxH);
    CGContextScaleCTM(ctx, 1.0, -1.0);
    CGContextDrawImage(ctx, CGRectMake(dstX, ctxH - dstY - h, w, h), srcImg);
    CGContextRestoreGState(ctx);

    CGImageRelease(srcImg);

    TKContentView *view = windowRegistry[@(d)];
    if (view) [view setNeedsDisplayInRect:NSMakeRect(dstX, dstY, w, h)];
}

// ============================================================================
// Pixmap management
// ============================================================================

CocoaPixmapID CocoaCreatePixmap(unsigned int width, unsigned int height,
                                 unsigned int depth) {
    if (width == 0) width = 1;
    if (height == 0) height = 1;

    CocoaPixmap *pm = (CocoaPixmap *)calloc(1, sizeof(CocoaPixmap));
    pm->width = width;
    pm->height = height;
    pm->depth = depth;

    size_t bytesPerRow = width * 4;
    pm->data = calloc(height, bytesPerRow);

    CGColorSpaceRef cs = CGColorSpaceCreateDeviceRGB();
    pm->ctx = CGBitmapContextCreate(pm->data, width, height, 8, bytesPerRow,
                                     cs, kCGImageAlphaPremultipliedFirst | kCGBitmapByteOrder32Host);
    CGColorSpaceRelease(cs);

    if (pm->ctx) {
        // Flip coordinates to match Tk (origin at top-left)
        CGContextTranslateCTM(pm->ctx, 0, height);
        CGContextScaleCTM(pm->ctx, 1.0, -1.0);
    }

    CocoaPixmapID pmid = nextPixmapID++;
    pixmapRegistry[@(pmid)] = [NSValue valueWithPointer:pm];
    return pmid;
}

void CocoaFreePixmap(CocoaPixmapID pmid) {
    NSValue *v = pixmapRegistry[@(pmid)];
    if (!v) return;
    CocoaPixmap *pm = (CocoaPixmap *)[v pointerValue];
    if (pm->ctx) CGContextRelease(pm->ctx);
    if (pm->data) free(pm->data);
    free(pm);
    [pixmapRegistry removeObjectForKey:@(pmid)];
}

CocoaPixmapID CocoaCreateBitmapFromData(const unsigned char *bits,
                                         unsigned int width, unsigned int height) {
    // XBM format: each byte is LSB-first, rows padded to 8 bits
    CocoaPixmapID pmid = CocoaCreatePixmap(width, height, 1);
    CocoaPixmap *pm = lookupPixmap(pmid);
    if (!pm || !pm->ctx) return pmid;

    // Convert XBM bits to the bitmap context
    unsigned int bytesPerRow = (width + 7) / 8;
    for (unsigned int y = 0; y < height; y++) {
        for (unsigned int x = 0; x < width; x++) {
            unsigned int byteIdx = y * bytesPerRow + x / 8;
            unsigned int bitIdx = x % 8;
            bool set = (bits[byteIdx] >> bitIdx) & 1;
            if (set) {
                // Set pixel to black (foreground)
                CGContextSetRGBFillColor(pm->ctx, 0, 0, 0, 1);
                CGContextFillRect(pm->ctx, CGRectMake(x, y, 1, 1));
            }
        }
    }

    return pmid;
}

// ============================================================================
// Cursor management
// ============================================================================

// Map abstract cursor shapes to NSCursor
static NSCursor *cursorForShape(unsigned int shape) {
    switch (shape) {
        case 0:  return [NSCursor arrowCursor];         // Arrow
        case 1:  return [NSCursor crosshairCursor];     // Crosshair
        case 2:  return [NSCursor openHandCursor];      // Fleur
        case 3:  return [NSCursor pointingHandCursor];  // Hand1
        case 4:  return [NSCursor pointingHandCursor];  // Hand2
        case 5:  return [NSCursor arrowCursor];         // LeftPtr
        case 6:  return [NSCursor crosshairCursor];     // Plus
        case 7:  return [NSCursor arrowCursor];         // QuestionArrow
        case 8:  return [NSCursor resizeLeftRightCursor]; // SBHDoubleArrow
        case 9:  return [NSCursor resizeUpDownCursor];  // SBVDoubleArrow
        case 10: return [NSCursor arrowCursor];         // SizingAngle
        case 11: return [NSCursor arrowCursor];         // TopLeftArrow
        case 12: return [NSCursor operationNotAllowedCursor]; // Watch (closest)
        case 13: return [NSCursor IBeamCursor];         // XTerm
        default: return [NSCursor arrowCursor];
    }
}

CocoaCursorID CocoaCreateCursor(unsigned int shape) {
    NSCursor *cursor = cursorForShape(shape);
    [cursor retain];
    return (CocoaCursorID)(void *)cursor;
}

void CocoaDefineCursor(CocoaWindowID w, CocoaCursorID cursor) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view) {
            NSCursor *c = (NSCursor *)(void *)cursor;
            [view addCursorRect:[view bounds] cursor:c];
            [c set];
        }
    });
}

void CocoaSetCursorShape(CocoaWindowID w, unsigned int shape) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view) {
            NSCursor *c = cursorForShape(shape);
            [view addCursorRect:[view bounds] cursor:c];
            [c set];
        }
    });
}

void CocoaUndefineCursor(CocoaWindowID w) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view) {
            [view discardCursorRects];
            [[NSCursor arrowCursor] set];
        }
    });
}

void CocoaFreeCursor(CocoaCursorID cursor) {
    if (cursor) {
        NSCursor *c = (NSCursor *)(void *)cursor;
        [c release];
    }
}

// ============================================================================
// Grab management
// ============================================================================

int CocoaGrabPointer(CocoaWindowID w) {
    // macOS doesn't have X11-style grabs. We can approximate with
    // mouse capture, but for now return success.
    return 0; // GrabSuccess
}

void CocoaUngrabPointer(void) {
    // No-op on macOS
}

int CocoaGrabKeyboard(CocoaWindowID w) {
    return 0; // GrabSuccess
}

void CocoaUngrabKeyboard(void) {
    // No-op on macOS
}

// ============================================================================
// Selection / clipboard
// ============================================================================

void CocoaSetClipboardText(const char *text) {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        [pb clearContents];
        [pb setString:[NSString stringWithUTF8String:text] forType:NSPasteboardTypeString];
    }
}

char *CocoaGetClipboardText(void) {
    @autoreleasepool {
        NSPasteboard *pb = [NSPasteboard generalPasteboard];
        NSString *str = [pb stringForType:NSPasteboardTypeString];
        if (!str) return NULL;
        const char *utf8 = [str UTF8String];
        return strdup(utf8);
    }
}

void CocoaFreeString(char *s) {
    free(s);
}

// ============================================================================
// Event system
// ============================================================================

static volatile bool eventPumpRunning = false;

// CocoaStartEventPump processes NSEvents on the main thread and routes
// them through Cocoa's standard event dispatch (which triggers our
// NSView/NSWindow callbacks that post to the internal event queue).
void CocoaStartEventPump(void) {
    eventPumpRunning = true;

    // Use a dispatch source / timer to continuously pump events.
    // This runs on the main dispatch queue (main thread).
    dispatch_source_t timer = dispatch_source_create(DISPATCH_SOURCE_TYPE_TIMER,
                                                       0, 0, dispatch_get_main_queue());
    // Fire every 1ms — fast enough for responsive UI
    dispatch_source_set_timer(timer, DISPATCH_TIME_NOW, 1 * NSEC_PER_MSEC, 0);
    dispatch_source_set_event_handler(timer, ^{
        if (!eventPumpRunning) {
            dispatch_source_cancel(timer);
            return;
        }
        @autoreleasepool {
            NSEvent *event;
            while ((event = [NSApp nextEventMatchingMask:NSEventMaskAny
                                              untilDate:nil
                                                 inMode:NSDefaultRunLoopMode
                                                dequeue:YES])) {
                [NSApp sendEvent:event];
            }
        }
    });
    dispatch_resume(timer);
}

void CocoaPumpEvents(void) {
    @autoreleasepool {
        // Process all pending NSEvents. This dispatches them through
        // Cocoa's normal event handling (NSApp sendEvent:), which triggers
        // NSView/NSWindow delegate callbacks that post to our event queue.
        NSEvent *event;
        while ((event = [NSApp nextEventMatchingMask:NSEventMaskAny
                                          untilDate:nil
                                             inMode:NSDefaultRunLoopMode
                                            dequeue:YES])) {
            [NSApp sendEvent:event];
            [NSApp updateWindows];
        }
    }
}

int CocoaNextEvent(CocoaRawEvent *ev) {
    pthread_mutex_lock(&eventQueueMutex);
    while (eventQueueHead == eventQueueTail) {
        // Wait with a timeout so we can check periodically.
        // On macOS, events come from the Cocoa event pump on the main thread.
        struct timespec ts;
        clock_gettime(CLOCK_REALTIME, &ts);
        ts.tv_nsec += 10 * 1000000; // 10ms timeout
        if (ts.tv_nsec >= 1000000000) {
            ts.tv_sec += 1;
            ts.tv_nsec -= 1000000000;
        }
        pthread_cond_timedwait(&eventQueueCond, &eventQueueMutex, &ts);
    }
    if (eventQueueHead == eventQueueTail) {
        // Spurious wakeup or timeout — no event available
        pthread_mutex_unlock(&eventQueueMutex);
        memset(ev, 0, sizeof(*ev));
        return 0;
    }
    *ev = eventQueue[eventQueueHead];
    eventQueueHead = (eventQueueHead + 1) % EVENT_QUEUE_SIZE;
    pthread_mutex_unlock(&eventQueueMutex);
    return 1;
}

int CocoaPending(void) {
    pthread_mutex_lock(&eventQueueMutex);
    int count = (eventQueueTail - eventQueueHead + EVENT_QUEUE_SIZE) % EVENT_QUEUE_SIZE;
    pthread_mutex_unlock(&eventQueueMutex);
    return count;
}

// ============================================================================
// Atom emulation
// ============================================================================

uint64_t CocoaInternAtom(const char *name, bool onlyIfExists) {
    @autoreleasepool {
        NSString *key = [NSString stringWithUTF8String:name];
        NSNumber *existing = atomByName[key];
        if (existing) return [existing unsignedLongLongValue];

        if (onlyIfExists) return 0;

        uint64_t aid = nextAtomID++;
        atomByName[key] = @(aid);
        atomByID[@(aid)] = key;
        return aid;
    }
}

const char *CocoaGetAtomName(uint64_t atom) {
    @autoreleasepool {
        NSString *name = atomByID[@(atom)];
        if (!name) return "";
        return [name UTF8String]; // valid until pool is drained
    }
}

// ============================================================================
// Font support (Core Text)
// ============================================================================

typedef struct {
    CTFontRef font;
    int ascent;
    int descent;
    int maxWidth;
    bool fixed;
} CocoaFont;

static NSMutableDictionary<NSNumber *, NSValue *> *fontRegistry = nil;
static uintptr_t nextFontID = 1;

CocoaFontID CocoaOpenFont(const char *family, double size, int weight, int slant) {
    @autoreleasepool {
        if (!fontRegistry) fontRegistry = [NSMutableDictionary new];

        NSString *familyName = [NSString stringWithUTF8String:family];
        if ([familyName isEqualToString:@"sans-serif"]) {
            familyName = @"Helvetica Neue";
        } else if ([familyName isEqualToString:@"monospace"]) {
            familyName = @"Menlo";
        } else if ([familyName isEqualToString:@"serif"]) {
            familyName = @"Times New Roman";
        }

        if (size <= 0) size = 12.0;

        // Create base font
        CTFontRef baseFont = CTFontCreateWithName((CFStringRef)familyName, size, NULL);
        if (!baseFont) {
            baseFont = CTFontCreateWithName(CFSTR("Helvetica Neue"), size, NULL);
        }
        if (!baseFont) return 0;

        // Apply traits (bold, italic)
        CTFontSymbolicTraits traits = 0;
        if (weight == 1) traits |= kCTFontBoldTrait;
        if (slant == 1 || slant == 2) traits |= kCTFontItalicTrait;

        CTFontRef styledFont = baseFont;
        if (traits != 0) {
            CTFontRef withTraits = CTFontCreateCopyWithSymbolicTraits(baseFont, size, NULL,
                                                                      traits, traits);
            if (withTraits) {
                CFRelease(baseFont);
                styledFont = withTraits;
            }
        }

        CocoaFont *cf = (CocoaFont *)calloc(1, sizeof(CocoaFont));
        cf->font = styledFont;
        cf->ascent = (int)ceil(CTFontGetAscent(styledFont));
        cf->descent = (int)ceil(CTFontGetDescent(styledFont));

        // Max advance width
        CGGlyph glyph;
        UniChar ch = 'M';
        CTFontGetGlyphsForCharacters(styledFont, &ch, &glyph, 1);
        CGSize advance;
        CTFontGetAdvancesForGlyphs(styledFont, kCTFontOrientationHorizontal, &glyph, &advance, 1);
        cf->maxWidth = (int)ceil(advance.width);

        // Check if monospace
        CTFontSymbolicTraits actualTraits = CTFontGetSymbolicTraits(styledFont);
        cf->fixed = (actualTraits & kCTFontMonoSpaceTrait) != 0;

        CocoaFontID fid = nextFontID++;
        fontRegistry[@(fid)] = [NSValue valueWithPointer:cf];
        return fid;
    }
}

void CocoaCloseFont(CocoaFontID fid) {
    NSValue *v = fontRegistry[@(fid)];
    if (!v) return;
    CocoaFont *cf = (CocoaFont *)[v pointerValue];
    if (cf->font) CFRelease(cf->font);
    free(cf);
    [fontRegistry removeObjectForKey:@(fid)];
}

static CocoaFont *lookupFont(CocoaFontID fid) {
    NSValue *v = fontRegistry[@(fid)];
    if (!v) return NULL;
    return (CocoaFont *)[v pointerValue];
}

int CocoaFontAscent(CocoaFontID fid) {
    CocoaFont *cf = lookupFont(fid);
    return cf ? cf->ascent : 12;
}

int CocoaFontDescent(CocoaFontID fid) {
    CocoaFont *cf = lookupFont(fid);
    return cf ? cf->descent : 3;
}

int CocoaFontMaxWidth(CocoaFontID fid) {
    CocoaFont *cf = lookupFont(fid);
    return cf ? cf->maxWidth : 8;
}

bool CocoaFontIsFixed(CocoaFontID fid) {
    CocoaFont *cf = lookupFont(fid);
    return cf ? cf->fixed : false;
}

int CocoaMeasureString(CocoaFontID fid, const char *s, int len) {
    @autoreleasepool {
        CocoaFont *cf = lookupFont(fid);
        if (!cf || !cf->font) return 0;

        NSString *str = [[NSString alloc] initWithBytes:s length:len encoding:NSUTF8StringEncoding];
        if (!str) return 0;

        NSDictionary *attrs = @{
            (id)kCTFontAttributeName: (id)cf->font
        };
        NSAttributedString *attrStr = [[NSAttributedString alloc] initWithString:str
                                                                      attributes:attrs];
        CTLineRef line = CTLineCreateWithAttributedString((CFAttributedStringRef)attrStr);
        if (!line) return 0;

        double width = CTLineGetTypographicBounds(line, NULL, NULL, NULL);
        CFRelease(line);
        return (int)ceil(width);
    }
}

void CocoaDrawString(CocoaDrawableID d, CocoaFontID fid,
                      int x, int y, const char *s, int len,
                      uint64_t pixel, uint16_t r, uint16_t g, uint16_t b) {
    @autoreleasepool {
        CGContextRef ctx = getDrawableContext(d);
        CocoaFont *cf = lookupFont(fid);
        if (!ctx || !cf || !cf->font) return;

        NSString *str = [[NSString alloc] initWithBytes:s length:len encoding:NSUTF8StringEncoding];
        if (!str) return;

        CGFloat cr = r / 65535.0;
        CGFloat cg = g / 65535.0;
        CGFloat cb = b / 65535.0;
        CGColorRef color = CGColorCreateGenericRGB(cr, cg, cb, 1.0);

        NSDictionary *attrs = @{
            (id)kCTFontAttributeName: (id)cf->font,
            (id)kCTForegroundColorAttributeName: (id)color
        };
        NSAttributedString *attrStr = [[NSAttributedString alloc] initWithString:str
                                                                      attributes:attrs];
        CTLineRef line = CTLineCreateWithAttributedString((CFAttributedStringRef)attrStr);
        CGColorRelease(color);
        if (!line) return;

        // Draw text. The context is flipped (origin top-left), but Core Text
        // expects bottom-left origin. We need to temporarily unflip.
        CGContextSaveGState(ctx);
        unsigned int ctxH = (unsigned int)CGBitmapContextGetHeight(ctx);
        CGContextTranslateCTM(ctx, 0, ctxH);
        CGContextScaleCTM(ctx, 1.0, -1.0);
        // y is baseline in Tk coords (top-down), convert to CG coords (bottom-up)
        CGFloat cgY = ctxH - y;
        CGContextSetTextPosition(ctx, x, cgY);
        CTLineDraw(line, ctx);
        CGContextRestoreGState(ctx);

        CFRelease(line);

        TKContentView *view = windowRegistry[@(d)];
        if (view) [view setNeedsDisplay:YES];
    }
}

// ============================================================================
// WM helpers
// ============================================================================

void CocoaSetWMProtocols(CocoaWindowID w) {
    // On macOS, WM protocols are handled natively (window close, etc.)
    // No-op — our windowShouldClose delegate handles WM_DELETE_WINDOW
}

void CocoaSetWMHints(CocoaWindowID w, bool input, int initialState) {
    // No direct equivalent on macOS
}

void CocoaSetWMNormalHints(CocoaWindowID w, int minW, int minH,
                            int maxW, int maxH) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (!view || !view.isTopLevel) return;
        NSWindow *window = [view window];
        if (minW > 0 && minH > 0) {
            [window setContentMinSize:NSMakeSize(minW, minH)];
        }
        if (maxW > 0 && maxH > 0) {
            [window setContentMaxSize:NSMakeSize(maxW, maxH)];
        }
    });
}

void CocoaSetClassHint(CocoaWindowID w, const char *name, const char *cls) {
    // No direct equivalent on macOS — informational only
}

void CocoaSetTransientFor(CocoaWindowID w, CocoaWindowID parent) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        TKContentView *parentView = windowRegistry[@(parent)];
        if (!view || !parentView) return;
        NSWindow *childWin = [view window];
        NSWindow *parentWin = [parentView window];
        if (childWin && parentWin) {
            [parentWin addChildWindow:childWin ordered:NSWindowAbove];
        }
    });
}

void CocoaIconifyWindow(CocoaWindowID w) {
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view && view.isTopLevel) {
            [[view window] miniaturize:nil];
        }
    });
}

void CocoaWithdrawWindow(CocoaWindowID w) {
    CocoaUnmapWindow(w);
}

void CocoaSetIconName(CocoaWindowID w, const char *name) {
    NSString *iconName = [NSString stringWithUTF8String:name];
    runOnMain(^{
        TKContentView *view = windowRegistry[@(w)];
        if (view && view.isTopLevel) {
            [[view window] setMiniwindowTitle:iconName];
        }
    });
}
