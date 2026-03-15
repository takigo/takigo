//go:build linux || freebsd || openbsd || netbsd

package takigo

import (
	"github.com/msorc/takigo/font"
	"github.com/msorc/takigo/platform"
	x11platform "github.com/msorc/takigo/platform/x11"
)

// platformInit creates the platform-specific display server, event parser,
// and font opener. Each platform backend provides its own implementation.
func platformInit(displayName string) (platform.DisplayServer, platform.EventParser, font.FontOpener, error) {
	server, err := x11platform.NewDisplayServer(displayName)
	if err != nil {
		return nil, nil, nil, err
	}

	x11platform.InitPredefinedAtoms()

	parser := server.EventParser()
	fontOpener := server.FontOpener(server.DefaultScreen())

	return server, parser, fontOpener, nil
}
