//go:build darwin

package takigo

import (
	"errors"

	"github.com/takigo/takigo/font"
	clib "github.com/takigo/takigo/internal/cocoa"
	"github.com/takigo/takigo/platform"
	cocoaplatform "github.com/takigo/takigo/platform/cocoa"
)

// platformInit creates the platform-specific display server, event parser,
// and font opener. On macOS, this uses Cocoa/AppKit/CoreGraphics.
func platformInit(displayName string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	server, fontOpener, err := cocoaplatform.NewDisplayServer(displayName)
	if err != nil {
		return nil, nil, nil, err
	}

	parser := server.EventParser()
	return server, parser, fontOpener, nil
}

// checkEventLoopThread reports whether the event loop may run on the
// calling goroutine. AppKit may be used from the main thread only, and
// the main thread is the one the App was created on.
func checkEventLoopThread() error {
	if !clib.IsMainThread() {
		return errors.New("takigo: on macOS the event loop must run on the main goroutine (see THREADING.md)")
	}
	return nil
}
