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

	parser := server.EventParser()
	// FontOpener is backend-specific; use type assertion to access it.
	var fontOpener font.FontOpener
	if core, ok := server.(interface{ FontOpener(int) font.FontOpener }); ok {
		fontOpener = core.FontOpener(server.DefaultScreen())
	}

	return server, parser, fontOpener, nil
}
