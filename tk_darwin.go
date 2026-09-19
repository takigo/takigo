//go:build darwin

package takigo

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	cocoaplatform "github.com/msorc/takigo/platform/cocoa"
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
