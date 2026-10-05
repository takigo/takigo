//go:build (linux || freebsd || openbsd || netbsd) && cgo

package takigo

import (
	"github.com/takigo/takigo/font"
	"github.com/takigo/takigo/platform"
	x11platform "github.com/takigo/takigo/platform/x11"
)

// platformInit creates the platform-specific display server, event parser,
// and font opener. Each platform backend provides its own implementation.
func platformInit(displayName string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	server, fontOpener, err := x11platform.NewDisplayServer(displayName)
	if err != nil {
		return nil, nil, nil, err
	}

	parser := server.EventParser()
	return server, parser, fontOpener, nil
}
